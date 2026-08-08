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
}

func NewFileBackupStore(root string, paths []ManagedPath, timestamp func() string) FileBackupStore {
	return FileBackupStore{root: root, paths: append([]ManagedPath(nil), paths...), timestamp: timestamp}
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
	destinations := make([]string, len(store.paths))
	for index, path := range store.paths {
		destination, err := backupDestination(sessionRoot, path.RelativeDestination)
		if err != nil {
			return ManagedBackup{}, err
		}
		destinations[index] = destination
	}
	if err := os.MkdirAll(store.root, 0o755); err != nil {
		return ManagedBackup{}, err
	}
	if err := os.Mkdir(sessionRoot, 0o755); err != nil {
		return ManagedBackup{}, err
	}
	for index, path := range store.paths {
		if err := copyManagedPath(path.Source, destinations[index]); err != nil {
			if removeErr := os.RemoveAll(sessionRoot); removeErr != nil {
				return ManagedBackup{}, fmt.Errorf("copy managed path: %w; remove partial backup: %v", err, removeErr)
			}
			return ManagedBackup{}, err
		}
	}
	return ManagedBackup{Timestamp: timestamp, Path: sessionRoot}, nil
}

func (store FileBackupStore) List() ([]ManagedBackup, error) {
	entries, err := os.ReadDir(store.root)
	if err != nil {
		if os.IsNotExist(err) {
			return []ManagedBackup{}, nil
		}
		return nil, err
	}
	backups := make([]ManagedBackup, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
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
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("backup %q is not a directory", timestamp)
	}
	return os.RemoveAll(path)
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
	defer input.Close()
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
	store      BackupStore
	backups    []ManagedBackup
	cursor     int
	confirming string
	err        error
	width      int
	height     int
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
		if menu.cursor >= len(menu.backups) {
			menu.cursor = max(len(menu.backups)-1, 0)
		}
	case tea.KeyMsg:
		return menu.key(message)
	}
	return menu, nil
}

func (menu BackupMenu) key(message tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return menu, menu.create()
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
	} else if len(menu.backups) == 0 {
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
	if menu.confirming != "" {
		content.WriteString("\nDelete backup " + menu.confirming + "? (y/n)\n")
	} else {
		content.WriteString("\nc create • d delete • q/esc return\n")
	}
	return tea.NewView(content.String())
}

type backupsLoadedMsg struct {
	backups []ManagedBackup
	err     error
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
		if _, err := menu.store.Create(); err != nil {
			return backupsLoadedMsg{err: err}
		}
		backups, err := menu.store.List()
		return backupsLoadedMsg{backups: backups, err: err}
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
