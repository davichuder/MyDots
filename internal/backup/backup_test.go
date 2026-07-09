package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupFile(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcContent := "test backup content"
	srcPath := filepath.Join(tempHome, "testfile.txt")
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-05-23T10-30-00Z"
	err := BackupFile(srcPath, sessionTS)
	if err != nil {
		t.Fatalf("BackupFile returned error: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	destPath := filepath.Join(root, sessionTS, "testfile.txt")

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("backup file should exist at %q, got: %v", destPath, err)
	}
	if string(data) != srcContent {
		t.Errorf("expected backup content %q, got: %q", srcContent, string(data))
	}
}

func TestBackupFilePreservesRelativePath(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	subDir := filepath.Join(tempHome, ".config", "nvim")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(subDir, "init.lua")
	if err := os.WriteFile(srcPath, []byte("vim.opt.tabstop = 2"), 0644); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-06-01T12-00-00Z"
	err := BackupFile(srcPath, sessionTS)
	if err != nil {
		t.Fatalf("BackupFile returned error: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	destPath := filepath.Join(root, sessionTS, ".config", "nvim", "init.lua")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("backup should preserve relative path, expected file at %q", destPath)
	}
}

func TestBackupDir(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcDir := filepath.Join(tempHome, ".config", "mydots")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "config.json"), []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "data.txt"), []byte("nested data"), 0644); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-06-15T08-30-00Z"
	err := BackupDir(srcDir, sessionTS)
	if err != nil {
		t.Fatalf("BackupDir returned error: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)

	destFile := filepath.Join(root, sessionTS, ".config", "mydots", "config.json")
	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Errorf("backup should include config.json at %q, got: %v", destFile, err)
	} else if string(data) != `{"key": "value"}` {
		t.Errorf("config.json content mismatch: got %q", string(data))
	}

	destSubFile := filepath.Join(root, sessionTS, ".config", "mydots", "sub", "data.txt")
	data, err = os.ReadFile(destSubFile)
	if err != nil {
		t.Errorf("backup should include sub/data.txt at %q, got: %v", destSubFile, err)
	} else if string(data) != "nested data" {
		t.Errorf("sub/data.txt content mismatch: got %q", string(data))
	}
}

func TestListBackupsNewestFirst(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	root := filepath.Join(tempHome, backupDirName)
	timestamps := []string{
		"2026-05-01T10-00-00Z",
		"2026-06-15T08-30-00Z",
		"2026-04-10T12-00-00Z",
	}
	for _, ts := range timestamps {
		dir := filepath.Join(root, ts)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".backup"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups returned error: %v", err)
	}
	if len(backups) != 3 {
		t.Fatalf("expected 3 backups, got: %d", len(backups))
	}
	if backups[0].Timestamp != "2026-06-15T08-30-00Z" {
		t.Errorf("expected newest first: 2026-06-15T08-30-00Z, got: %q", backups[0].Timestamp)
	}
	if backups[1].Timestamp != "2026-05-01T10-00-00Z" {
		t.Errorf("expected middle: 2026-05-01T10-00-00Z, got: %q", backups[1].Timestamp)
	}
	if backups[2].Timestamp != "2026-04-10T12-00-00Z" {
		t.Errorf("expected oldest last: 2026-04-10T12-00-00Z, got: %q", backups[2].Timestamp)
	}
}

func TestListBackupsEmpty(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups should not return error when backups dir is missing, got: %v", err)
	}
	if backups == nil {
		t.Error("ListBackups should return empty slice (not nil) when no backups exist")
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got: %d", len(backups))
	}
}

func TestDeleteBackup(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	root := filepath.Join(tempHome, backupDirName)
	ts := "2026-07-01T09-00-00Z"
	dir := filepath.Join(root, ts)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("backup dir should exist before deletion")
	}

	err := DeleteBackup(ts)
	if err != nil {
		t.Fatalf("DeleteBackup returned error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("backup dir should be gone after deletion")
	}
}

func TestDeleteBackupNonExistent(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	err := DeleteBackup("nonexistent-timestamp")
	if err == nil {
		t.Fatal("DeleteBackup should return error for non-existent timestamp")
	}
	if !strings.Contains(err.Error(), "nonexistent-timestamp") {
		t.Errorf("error should mention the timestamp, got: %v", err)
	}
}

func TestBackupFileSourceOutsideHome(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	outsidePath := filepath.Join(tempHome, "..", "outside.txt")
	outsidePath, err := filepath.Abs(outsidePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outsidePath, []byte("outside content"), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(outsidePath) }()

	err = BackupFile(outsidePath, "2026-07-01T00-00-00Z")
	if err == nil {
		t.Error("BackupFile should return error for path outside $HOME")
	}
}

func TestBackupDirAlreadyExists(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcDir := filepath.Join(tempHome, "mydir")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-07-01T00-00-00Z"
	if err := BackupDir(srcDir, sessionTS); err != nil {
		t.Fatalf("first BackupDir should succeed, got: %v", err)
	}

	err := BackupDir(srcDir, sessionTS)
	if err != nil {
		t.Fatalf("second BackupDir (merge) should not error, got: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	destPath := filepath.Join(root, sessionTS, "mydir", "file.txt")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("backup file should exist after merge, got: %v", err)
	}
	if string(data) != "v1" {
		t.Errorf("expected content 'v1', got: %q", string(data))
	}
}

func TestBackupFileEmptyFile(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcPath := filepath.Join(tempHome, "empty.txt")
	if err := os.WriteFile(srcPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-07-01T12-00-00Z"
	err := BackupFile(srcPath, sessionTS)
	if err != nil {
		t.Fatalf("BackupFile on empty file returned error: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	destPath := filepath.Join(root, sessionTS, "empty.txt")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("backup of empty file should exist, got: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected 0 bytes for empty file backup, got: %d", len(data))
	}
}

func TestBackupFileSymlink(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	targetContent := "symlink target content"
	targetPath := filepath.Join(tempHome, "target.txt")
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(tempHome, "link.txt")
	err := os.Symlink(targetPath, linkPath)
	if err != nil {
		t.Skipf("skipping symlink test: %v (may require admin on Windows)", err)
	}

	sessionTS := "2026-07-01T15-00-00Z"
	err = BackupFile(linkPath, sessionTS)
	if err != nil {
		t.Fatalf("BackupFile on symlink returned error: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	destPath := filepath.Join(root, sessionTS, "link.txt")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("backup of symlink should exist at %q, got: %v", destPath, err)
	}
	if string(data) != targetContent {
		t.Errorf("expected symlink target content %q, got: %q", targetContent, string(data))
	}
}

func TestBackupFileNonExistentFile(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	nonExistent := filepath.Join(tempHome, "does-not-exist.txt")
	err := BackupFile(nonExistent, "2026-07-01T00-00-00Z")
	if err == nil {
		t.Error("BackupFile should return error for non-existent source file")
	}
}

func TestRelativeOutOfHome(t *testing.T) {
	tests := []struct {
		name  string
		setup func(home string) string
	}{
		{
			name: "parent directory",
			setup: func(home string) string {
				return filepath.Join(home, "..", "outside")
			},
		},
		{
			name: "absolutely outside",
			setup: func(_ string) string {
				return filepath.Join(t.TempDir(), "other")
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			orig := osUserHomeDir
			osUserHomeDir = func() (string, error) { return home, nil }
			defer func() { osUserHomeDir = orig }()

			path := tc.setup(home)
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
			testFile := filepath.Join(path, "test.txt")
			if err := os.WriteFile(testFile, []byte("data"), 0644); err != nil {
				t.Fatal(err)
			}

			err := BackupFile(testFile, "ts")
			if err == nil {
				t.Error("expected error for path outside home")
			}
		})
	}
}

func TestListBackupsOnlyDirectories(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	root := filepath.Join(tempHome, backupDirName)
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}

	validTS := "2026-08-01T00-00-00Z"
	if err := os.MkdirAll(filepath.Join(root, validTS), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "not-a-backup.txt"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups returned error: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected 1 backup dir, got: %d (backups: %v)", len(backups), backups)
	}
	if backups[0].Timestamp != validTS {
		t.Errorf("expected timestamp %q, got: %q", validTS, backups[0].Timestamp)
	}
}

func TestDeleteBackupPreservesOtherBackups(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	root := filepath.Join(tempHome, backupDirName)
	tsA := "2026-07-01T00-00-00Z"
	tsB := "2026-07-02T00-00-00Z"

	for _, ts := range []string{tsA, tsB} {
		dir := filepath.Join(root, ts)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	if err := DeleteBackup(tsA); err != nil {
		t.Fatalf("DeleteBackup returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, tsA)); !os.IsNotExist(err) {
		t.Error("tsA should be deleted")
	}
	if _, err := os.Stat(filepath.Join(root, tsB)); os.IsNotExist(err) {
		t.Error("tsB should still exist after deleting tsA")
	}

	backups, err := ListBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup after deleting tsA, got: %d", len(backups))
	}
	if len(backups) > 0 && backups[0].Timestamp != tsB {
		t.Errorf("expected remaining backup %q, got: %q", tsB, backups[0].Timestamp)
	}
}

func TestBackupFileDestDirCreation(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcPath := filepath.Join(tempHome, "deep.txt")
	if err := os.WriteFile(srcPath, []byte("deep file"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatal(err)
	}

	sessionTS := "2026-09-01T10-00-00Z"
	err := BackupFile(srcPath, sessionTS)
	if err != nil {
		t.Fatalf("BackupFile should create dest dir, got: %v", err)
	}

	root := filepath.Join(tempHome, backupDirName)
	dest := filepath.Join(root, sessionTS, "deep.txt")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("backup file should exist at %q, got: %v", dest, err)
	}
	if string(data) != "deep file" {
		t.Errorf("expected 'deep file', got: %q", string(data))
	}
}

func TestBackupFileDefaultPermissions(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcPath := filepath.Join(tempHome, "perms.txt")
	if err := os.WriteFile(srcPath, []byte("check perms"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := BackupFile(srcPath, "ts"); err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(tempHome, backupDirName)
	dest := filepath.Join(root, "ts", "perms.txt")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("backup should be readable, got: %v", err)
	}
	if string(data) != "check perms" {
		t.Errorf("expected 'check perms', got: %q", string(data))
	}
}

func TestBackupTimestampWithDashes(t *testing.T) {
	tempHome := t.TempDir()
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tempHome, nil }
	defer func() { osUserHomeDir = orig }()

	srcPath := filepath.Join(tempHome, "config.toml")
	if err := os.WriteFile(srcPath, []byte("theme = dark"), 0644); err != nil {
		t.Fatal(err)
	}

	ts := "2026-05-23T10-30-00Z"
	if err := BackupFile(srcPath, ts); err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(tempHome, backupDirName)
	dest := filepath.Join(root, ts, "config.toml")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("backup should use timestamp with dashes, expected %q", dest)
	}
}
