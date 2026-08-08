package screens

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ManagedBackup identifies a backup session created from pre-resolved paths.
type ManagedBackup struct {
	Timestamp string
	Path      string
}

// ManagedPath identifies a pre-resolved source and its canonical backup destination.
type ManagedPath struct {
	Source              string
	RelativeDestination string
}

// BackupStore isolates backup filesystem operations from the menu model.
type BackupStore interface {
	Create() (ManagedBackup, error)
	List() ([]ManagedBackup, error)
	Delete(timestamp string) error
}

// FileBackupStore stores managed-path copies beneath an injected root.
// Paths are deliberately supplied by the caller; this type never resolves HOME.
type FileBackupStore struct {
	root      string
	paths     []ManagedPath
	timestamp func() string
	fileOps   fileOps
}

func NewFileBackupStore(root string, paths []ManagedPath, timestamp func() string) FileBackupStore {
	return newFileBackupStoreWithFileOps(root, paths, timestamp, osBackupFileOps{})
}

type fileOps interface {
	MkdirAll(path string, perm os.FileMode) error
	Mkdir(path string, perm os.FileMode) error
	Rename(oldpath, newpath string) error
	RemoveAll(path string) error
	ReadDir(name string) ([]os.DirEntry, error)
	Stat(name string) (os.FileInfo, error)
}

type osBackupFileOps struct{}

func (osBackupFileOps) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osBackupFileOps) Mkdir(path string, perm os.FileMode) error    { return os.Mkdir(path, perm) }
func (osBackupFileOps) Rename(oldpath, newpath string) error         { return os.Rename(oldpath, newpath) }
func (osBackupFileOps) RemoveAll(path string) error                  { return os.RemoveAll(path) }
func (osBackupFileOps) ReadDir(name string) ([]os.DirEntry, error)   { return os.ReadDir(name) }
func (osBackupFileOps) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }

func newFileBackupStoreWithFileOps(root string, paths []ManagedPath, timestamp func() string, fileOps fileOps) FileBackupStore {
	return FileBackupStore{root: root, paths: append([]ManagedPath(nil), paths...), timestamp: timestamp, fileOps: fileOps}
}

func (store FileBackupStore) Create() (ManagedBackup, error) {
	if store.timestamp == nil {
		return ManagedBackup{}, fmt.Errorf("backup timestamp is required")
	}
	timestamp := store.timestamp()
	if err := validBackupTimestamp(timestamp); err != nil {
		return ManagedBackup{}, err
	}
	sessionRoot := filepath.Join(store.root, timestamp)
	stagingRoot := filepath.Join(store.root, "."+timestamp+".staging")
	destinations := make([]string, len(store.paths))
	for index, path := range store.paths {
		destination, err := backupDestination(stagingRoot, path.RelativeDestination)
		if err != nil {
			return ManagedBackup{}, err
		}
		destinations[index] = destination
	}
	if err := store.fileOps.MkdirAll(store.root, 0o755); err != nil {
		return ManagedBackup{}, err
	}
	if err := store.fileOps.Mkdir(stagingRoot, 0o755); err != nil {
		return ManagedBackup{}, err
	}
	for index, path := range store.paths {
		if err := copyManagedPath(path.Source, destinations[index]); err != nil {
			if removeErr := store.fileOps.RemoveAll(stagingRoot); removeErr != nil {
				return ManagedBackup{}, fmt.Errorf("copy managed path: %w; remove partial backup: %v", err, removeErr)
			}
			return ManagedBackup{}, err
		}
	}
	if err := os.WriteFile(filepath.Join(stagingRoot, ".complete"), nil, 0o644); err != nil {
		_ = store.fileOps.RemoveAll(stagingRoot)
		return ManagedBackup{}, err
	}
	if err := store.fileOps.Rename(stagingRoot, sessionRoot); err != nil {
		if removeErr := store.fileOps.RemoveAll(stagingRoot); removeErr != nil {
			return ManagedBackup{}, fmt.Errorf("publish backup: %w; remove staging backup: %v", err, removeErr)
		}
		return ManagedBackup{}, err
	}
	return ManagedBackup{Timestamp: timestamp, Path: sessionRoot}, nil
}

func (store FileBackupStore) List() ([]ManagedBackup, error) {
	entries, err := store.fileOps.ReadDir(store.root)
	if err != nil {
		if os.IsNotExist(err) {
			return []ManagedBackup{}, nil
		}
		return nil, err
	}
	backups := make([]ManagedBackup, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && validBackupTimestamp(entry.Name()) == nil {
			backups = append(backups, ManagedBackup{Timestamp: entry.Name(), Path: filepath.Join(store.root, entry.Name())})
		}
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].Timestamp > backups[j].Timestamp })
	return backups, nil
}

func (store FileBackupStore) Delete(timestamp string) error {
	if err := validBackupTimestamp(timestamp); err != nil {
		return err
	}
	path := filepath.Join(store.root, timestamp)
	info, err := store.fileOps.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("backup %q is not a directory", timestamp)
	}
	tombstone := filepath.Join(store.root, "."+timestamp+".deleting")
	if err := store.fileOps.Rename(path, tombstone); err != nil {
		return fmt.Errorf("quarantine backup %q: %w", timestamp, err)
	}
	if err := store.fileOps.RemoveAll(tombstone); err != nil {
		return fmt.Errorf("backup %q quarantined for cleanup: %w", timestamp, err)
	}
	return nil
}

func validBackupTimestamp(timestamp string) error {
	if timestamp == "" || timestamp == "." || timestamp == ".." || filepath.Base(timestamp) != timestamp || strings.ContainsAny(timestamp, `/\\`) {
		return fmt.Errorf("invalid backup timestamp %q", timestamp)
	}
	return nil
}

func backupDestination(sessionRoot, relativeDestination string) (string, error) {
	if relativeDestination == "" || filepath.IsAbs(relativeDestination) || filepath.VolumeName(relativeDestination) != "" {
		return "", fmt.Errorf("invalid backup destination %q", relativeDestination)
	}
	for _, component := range strings.FieldsFunc(relativeDestination, func(character rune) bool {
		return character == '/' || character == '\\'
	}) {
		if component == "." || component == ".." {
			return "", fmt.Errorf("invalid backup destination %q", relativeDestination)
		}
	}
	destination := filepath.Join(sessionRoot, relativeDestination)
	relative, err := filepath.Rel(sessionRoot, destination)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("invalid backup destination %q", relativeDestination)
	}
	return destination, nil
}

func copyManagedPath(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copyBackupFile(source, destination, info.Mode())
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyBackupFile(path, target, info.Mode())
	})
}

func copyBackupFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// BackupMenu creates, lists, and deletes only caller-managed backup sessions.
type BackupMenu struct {
	store            BackupStore
	backups          []ManagedBackup
	cursor           int
	confirming       string
	confirmingCreate bool
	err              error
	notice           string
	width            int
	height           int
}

func NewBackupMenu(store BackupStore) BackupMenu {
	return BackupMenu{store: store}
}

func (menu BackupMenu) Init() tea.Cmd { return menu.load() }

func (menu BackupMenu) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		menu.width, menu.height = message.Width, message.Height
	case backupsLoadedMsg:
		menu.backups, menu.err = message.backups, message.err
		menu.notice = message.notice
		if menu.cursor >= len(menu.backups) {
			menu.cursor = max(len(menu.backups)-1, 0)
		}
	case tea.KeyMsg:
		return menu.key(message)
	}
	return menu, nil
}

func (menu BackupMenu) key(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	if menu.confirmingCreate {
		switch message.Key().Text {
		case "y":
			menu.confirmingCreate = false
			return menu, menu.create()
		case "n":
			menu.confirmingCreate = false
		}
		return menu, nil
	}
	if menu.confirming != "" {
		switch message.Key().Text {
		case "y":
			timestamp := menu.confirming
			menu.confirming = ""
			return menu, menu.remove(timestamp)
		case "n":
			menu.confirming = ""
		}
		return menu, nil
	}
	switch message.Key().Code {
	case tea.KeyUp:
		if menu.cursor > 0 {
			menu.cursor--
		}
	case tea.KeyDown:
		if menu.cursor < len(menu.backups)-1 {
			menu.cursor++
		}
	case tea.KeyEscape:
		return menu, changeScreen(ScreenMain, nil)
	}
	switch message.Key().Text {
	case "c":
		menu.notice = ""
		menu.confirmingCreate = true
	case "d":
		if len(menu.backups) > 0 {
			menu.confirming = menu.backups[menu.cursor].Timestamp
		}
	case "q":
		return menu, changeScreen(ScreenMain, nil)
	}
	return menu, nil
}

func (menu BackupMenu) View() tea.View {
	var content strings.Builder
	content.WriteString("Backups\n\n")
	if menu.err != nil {
		content.WriteString("Error: " + menu.err.Error() + "\n")
	} else {
		if menu.notice != "" {
			content.WriteString(menu.notice + "\n")
		}
		if len(menu.backups) == 0 {
			content.WriteString("No managed backups yet.\n")
		} else {
			for index, backup := range menu.backups {
				prefix := "  "
				if index == menu.cursor {
					prefix = "> "
				}
				content.WriteString(prefix + backup.Timestamp + "\n")
			}
		}
	}
	if menu.confirmingCreate {
		content.WriteString("\nCreate backup now? (y/n)\n")
	} else if menu.confirming != "" {
		content.WriteString("\nDelete backup " + menu.confirming + "? (y/n)\n")
	} else {
		content.WriteString("\nc create • d delete • q/esc return\n")
	}
	return tea.NewView(content.String())
}

type backupsLoadedMsg struct {
	backups []ManagedBackup
	err     error
	notice  string
}

func (menu BackupMenu) load() tea.Cmd {
	return func() tea.Msg {
		if menu.store == nil {
			return backupsLoadedMsg{err: fmt.Errorf("backup store is unavailable")}
		}
		backups, err := menu.store.List()
		return backupsLoadedMsg{backups: backups, err: err}
	}
}

func (menu BackupMenu) create() tea.Cmd {
	return func() tea.Msg {
		if menu.store == nil {
			return backupsLoadedMsg{err: fmt.Errorf("backup store is unavailable")}
		}
		backup, err := menu.store.Create()
		if err != nil {
			return backupsLoadedMsg{err: err}
		}
		backups, err := menu.store.List()
		return backupsLoadedMsg{backups: backups, err: err, notice: "Backup created: " + backup.Timestamp}
	}
}

func (menu BackupMenu) remove(timestamp string) tea.Cmd {
	return func() tea.Msg {
		if menu.store == nil {
			return backupsLoadedMsg{err: fmt.Errorf("backup store is unavailable")}
		}
		if err := menu.store.Delete(timestamp); err != nil {
			return backupsLoadedMsg{err: err}
		}
		backups, err := menu.store.List()
		return backupsLoadedMsg{backups: backups, err: err}
	}
}
