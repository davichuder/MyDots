package screens

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestBackupMenuCreatesAllPreResolvedPathsAndListsPriorSessions(t *testing.T) {
	managedRoot := t.TempDir()
	configFile := filepath.Join(managedRoot, ".config", "mydots", "settings.toml")
	configDir := filepath.Join(managedRoot, ".config", "nvim")
	if err := os.MkdirAll(filepath.Dir(configFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte("theme = dark"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "init.lua"), []byte("vim.o.number = true"), 0o644); err != nil {
		t.Fatal(err)
	}
	duplicateFile := filepath.Join(managedRoot, ".local", "share", "settings.toml")
	if err := os.MkdirAll(filepath.Dir(duplicateFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicateFile, []byte("cache = true"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupRoot := t.TempDir()
	priorTimestamp := "2026-08-07T09-00-00Z"
	if err := os.MkdirAll(filepath.Join(backupRoot, priorTimestamp), 0o755); err != nil {
		t.Fatal(err)
	}
	markBackupComplete(t, filepath.Join(backupRoot, priorTimestamp))
	store := NewFileBackupStore(backupRoot, []ManagedPath{
		{Source: configFile, RelativeDestination: filepath.Join(".config", "mydots", "settings.toml")},
		{Source: configDir, RelativeDestination: filepath.Join(".config", "nvim")},
		{Source: duplicateFile, RelativeDestination: filepath.Join(".local", "share", "settings.toml")},
	}, func() string { return "2026-08-08T10-00-00Z" })
	menu := NewBackupMenu(store)
	menu = applyBackupMessage(t, menu, commandMessage(t, menu.Init()))

	updated, command := menu.Update(keyPress('c', "c"))
	menu = backupMenu(t, updated)
	if command != nil {
		t.Fatal("create request must wait for confirmation")
	}
	updated, command = menu.Update(keyPress('y', "y"))
	menu = backupMenu(t, updated)
	menu = applyBackupMessage(t, menu, commandMessage(t, command))

	view := menu.View().Content
	for _, timestamp := range []string{"2026-08-08T10-00-00Z", priorTimestamp} {
		if !strings.Contains(view, timestamp) {
			t.Errorf("View() = %q, want backup timestamp %q", view, timestamp)
		}
	}
	for _, path := range []string{
		filepath.Join(backupRoot, "2026-08-08T10-00-00Z", ".config", "mydots", "settings.toml"),
		filepath.Join(backupRoot, "2026-08-08T10-00-00Z", ".config", "nvim", "init.lua"),
		filepath.Join(backupRoot, "2026-08-08T10-00-00Z", ".local", "share", "settings.toml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("created backup path %q: %v", path, err)
		}
	}
	for path, want := range map[string]string{
		filepath.Join(backupRoot, "2026-08-08T10-00-00Z", ".config", "mydots", "settings.toml"): "theme = dark",
		filepath.Join(backupRoot, "2026-08-08T10-00-00Z", ".local", "share", "settings.toml"):   "cache = true",
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read backup path %q: %v", path, err)
		} else if string(got) != want {
			t.Errorf("backup path %q = %q, want %q", path, got, want)
		}
	}
}

func TestBackupMenuConfirmedDeletionRemovesOnlySelectedTimestamp(t *testing.T) {
	backupRoot := t.TempDir()
	selected := "2026-08-08T10-00-00Z"
	remaining := "2026-08-07T09-00-00Z"
	for _, timestamp := range []string{selected, remaining} {
		if err := os.MkdirAll(filepath.Join(backupRoot, timestamp), 0o755); err != nil {
			t.Fatal(err)
		}
		markBackupComplete(t, filepath.Join(backupRoot, timestamp))
	}
	menu := NewBackupMenu(NewFileBackupStore(backupRoot, nil, func() string { return "unused" }))
	menu = applyBackupMessage(t, menu, commandMessage(t, menu.Init()))

	updated, command := menu.Update(keyPress('d', "d"))
	menu = backupMenu(t, updated)
	if command != nil {
		t.Fatal("delete request must wait for confirmation")
	}
	if !strings.Contains(menu.View().Content, "Delete backup "+selected+"? (y/n)") {
		t.Fatalf("View() = %q, want confirmation for selected timestamp", menu.View().Content)
	}

	updated, command = menu.Update(keyPress('y', "y"))
	menu = backupMenu(t, updated)
	menu = applyBackupMessage(t, menu, commandMessage(t, command))

	if _, err := os.Stat(filepath.Join(backupRoot, selected)); !os.IsNotExist(err) {
		t.Errorf("selected backup stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(backupRoot, remaining)); err != nil {
		t.Errorf("unselected backup stat error = %v, want it preserved", err)
	}
}

func TestBackupMenuCreateRequiresConfirmationAndShowsSuccess(t *testing.T) {
	backupRoot := t.TempDir()
	source := filepath.Join(t.TempDir(), "settings.toml")
	if err := os.WriteFile(source, []byte("safe"), 0o644); err != nil {
		t.Fatal(err)
	}
	timestamp := "2026-08-09T10-00-00Z"
	menu := NewBackupMenu(NewFileBackupStore(backupRoot, []ManagedPath{{Source: source, RelativeDestination: "settings.toml"}}, func() string { return timestamp }))
	menu = applyBackupMessage(t, menu, commandMessage(t, menu.Init()))

	updated, command := menu.Update(keyPress('c', "c"))
	menu = backupMenu(t, updated)
	if command != nil {
		t.Fatal("create request must wait for confirmation")
	}
	if !strings.Contains(menu.View().Content, "Create backup now? (y/n)") {
		t.Fatalf("View() = %q, want create confirmation", menu.View().Content)
	}
	if _, err := os.Stat(filepath.Join(backupRoot, timestamp)); !os.IsNotExist(err) {
		t.Fatalf("backup exists before confirmation: %v", err)
	}

	updated, command = menu.Update(keyPress('y', "y"))
	menu = backupMenu(t, updated)
	menu = applyBackupMessage(t, menu, commandMessage(t, command))
	if !strings.Contains(menu.View().Content, "Backup created: "+timestamp) {
		t.Errorf("View() = %q, want visible creation success", menu.View().Content)
	}
}

func TestBackupMenuEmptyStateGolden(t *testing.T) {
	menu := NewBackupMenu(NewFileBackupStore(t.TempDir(), nil, func() string { return "unused" }))
	menu = applyBackupMessage(t, menu, commandMessage(t, menu.Init()))
	menu = applyBackupMessage(t, menu, tea.WindowSizeMsg{Width: 80, Height: 24})

	assertBackupGolden(t, "empty", menu.View().Content)
}

func TestFileBackupStoreRejectsUnsafeSessionAndDestinationPaths(t *testing.T) {
	backupRoot := t.TempDir()
	source := filepath.Join(t.TempDir(), "settings.toml")
	if err := os.WriteFile(source, []byte("safe"), 0o644); err != nil {
		t.Fatal(err)
	}
	absDestination := filepath.Join(t.TempDir(), "escaped.toml")

	for _, test := range []struct {
		name        string
		timestamp   string
		destination string
		outsidePath string
	}{
		{name: "current session", timestamp: ".", destination: "settings.toml", outsidePath: filepath.Join(backupRoot, "settings.toml")},
		{name: "parent session", timestamp: "..", destination: "settings.toml", outsidePath: filepath.Join(filepath.Dir(backupRoot), "settings.toml")},
		{name: "absolute destination", timestamp: "2026-08-08T10-00-00Z", destination: absDestination, outsidePath: absDestination},
		{name: "parent destination", timestamp: "2026-08-08T10-00-00Z", destination: filepath.Join("..", "escaped.toml"), outsidePath: filepath.Join(backupRoot, "escaped.toml")},
		{name: "nested parent destination", timestamp: "2026-08-08T10-00-00Z", destination: "safe/../escaped.toml", outsidePath: filepath.Join(backupRoot, "escaped.toml")},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := NewFileBackupStore(backupRoot, []ManagedPath{{Source: source, RelativeDestination: test.destination}}, func() string { return test.timestamp })
			if _, err := store.Create(); err == nil {
				t.Fatal("Create() error = nil, want unsafe path rejection")
			}
			backups, err := store.List()
			if err != nil {
				t.Fatal(err)
			}
			if len(backups) != 0 {
				t.Errorf("List() = %#v, want no backup session", backups)
			}
			if _, err := os.Stat(test.outsidePath); !os.IsNotExist(err) {
				t.Errorf("escaped path stat error = %v, want not exist", err)
			}
		})
	}
}

func TestFileBackupStoreRollsBackPartialSessionWhenCopyFails(t *testing.T) {
	backupRoot := t.TempDir()
	firstSource := filepath.Join(t.TempDir(), "first.toml")
	if err := os.WriteFile(firstSource, []byte("copied first"), 0o644); err != nil {
		t.Fatal(err)
	}
	timestamp := "2026-08-08T10-00-00Z"
	store := NewFileBackupStore(backupRoot, []ManagedPath{
		{Source: firstSource, RelativeDestination: filepath.Join("config", "first.toml")},
		{Source: filepath.Join(t.TempDir(), "missing.toml"), RelativeDestination: filepath.Join("config", "missing.toml")},
	}, func() string { return timestamp })

	if _, err := store.Create(); err == nil {
		t.Fatal("Create() error = nil, want deterministic second-copy failure")
	}
	if _, err := os.Stat(filepath.Join(backupRoot, timestamp)); !os.IsNotExist(err) {
		t.Errorf("partial session stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(backupRoot, "."+timestamp+".staging")); !os.IsNotExist(err) {
		t.Errorf("staging session stat error = %v, want cleanup", err)
	}
	backups, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Errorf("List() = %#v, want no partial session", backups)
	}
}

func TestFileBackupStoreListIgnoresStagingAndIncompleteSessions(t *testing.T) {
	backupRoot := t.TempDir()
	complete := "2026-08-08T10-00-00Z"
	incomplete := "2026-08-07T10-00-00Z"
	staging := ".2026-08-09T10-00-00Z.staging"
	for _, name := range []string{complete, incomplete, staging} {
		if err := os.MkdirAll(filepath.Join(backupRoot, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	markBackupComplete(t, filepath.Join(backupRoot, complete))

	backups, err := NewFileBackupStore(backupRoot, nil, func() string { return "unused" }).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 || backups[0].Timestamp != complete {
		t.Errorf("List() = %#v, want only completed backup %q", backups, complete)
	}
}

func TestFileBackupStoreDeleteQuarantinesBeforeRemoval(t *testing.T) {
	backupRoot := t.TempDir()
	timestamp := "2026-08-08T10-00-00Z"
	original := filepath.Join(backupRoot, timestamp)
	if err := os.MkdirAll(original, 0o755); err != nil {
		t.Fatal(err)
	}
	markBackupComplete(t, original)

	removeErr := errors.New("remove failed")
	store := newFileBackupStoreWithFileOps(backupRoot, nil, func() string { return "unused" }, failingBackupFileOps{fileOps: osBackupFileOps{}, removeAllErr: removeErr})
	err := store.Delete(timestamp)
	if !errors.Is(err, removeErr) {
		t.Fatalf("Delete() error = %v, want removal failure", err)
	}
	if _, err := os.Stat(original); !os.IsNotExist(err) {
		t.Errorf("original session stat error = %v, want quarantined", err)
	}
	if _, err := os.Stat(filepath.Join(backupRoot, "."+timestamp+".deleting")); err != nil {
		t.Errorf("quarantine stat error = %v, want retained for safe cleanup", err)
	}
	backups, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Errorf("List() = %#v, want quarantined session hidden", backups)
	}
}

func TestFileBackupStoreDeleteRenameFailureLeavesOriginalIntact(t *testing.T) {
	backupRoot := t.TempDir()
	timestamp := "2026-08-08T10-00-00Z"
	original := filepath.Join(backupRoot, timestamp)
	if err := os.MkdirAll(original, 0o755); err != nil {
		t.Fatal(err)
	}
	markBackupComplete(t, original)

	renameErr := errors.New("rename failed")
	store := newFileBackupStoreWithFileOps(backupRoot, nil, func() string { return "unused" }, failingBackupFileOps{fileOps: osBackupFileOps{}, renameErr: renameErr})
	if err := store.Delete(timestamp); !errors.Is(err, renameErr) {
		t.Fatalf("Delete() error = %v, want rename failure", err)
	}
	if _, err := os.Stat(original); err != nil {
		t.Errorf("original session stat error = %v, want intact", err)
	}
}

type failingBackupFileOps struct {
	fileOps
	renameErr    error
	removeAllErr error
}

func (ops failingBackupFileOps) Rename(oldpath, newpath string) error {
	if ops.renameErr != nil {
		return ops.renameErr
	}
	return ops.fileOps.Rename(oldpath, newpath)
}

func (ops failingBackupFileOps) RemoveAll(path string) error {
	if ops.removeAllErr != nil {
		return ops.removeAllErr
	}
	return ops.fileOps.RemoveAll(path)
}

func markBackupComplete(t *testing.T, sessionPath string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(sessionPath, ".complete"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertBackupGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "backup_menu", name+".golden")
	if *updateMainMenuGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create golden directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("View() at 80x24 = %q, want %q", got, want)
	}
}

func applyBackupMessage(t *testing.T, menu BackupMenu, message tea.Msg) BackupMenu {
	t.Helper()
	updated, _ := menu.Update(message)
	return backupMenu(t, updated)
}

func backupMenu(t *testing.T, model tea.Model) BackupMenu {
	t.Helper()
	menu, ok := model.(BackupMenu)
	if !ok {
		t.Fatalf("Update() model = %T, want BackupMenu", model)
	}
	return menu
}
