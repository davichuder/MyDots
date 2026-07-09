package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Backup struct {
	Timestamp string
	Path      string
}

const backupDirName string = ".mydots-backups"

var osUserHomeDir = os.UserHomeDir

func backupsRoot() (string, error) {
	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, backupDirName), nil
}

func relativeToHome(absPath string) (string, error) {
	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(home, absPath)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q is not within home directory %q", absPath, home)
	}

	return rel, nil
}

// BackupFile copies a single file to ~/.mydots-backups/<sessionTimestamp>/<relative-from-HOME>.
func BackupFile(filePath string, sessionTimestamp string) error {
	root, err := backupsRoot()
	if err != nil {
		return err
	}

	rel, err := relativeToHome(filePath)
	if err != nil {
		return err
	}

	dest := filepath.Join(root, sessionTimestamp, rel)
	destDir := filepath.Dir(dest)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, data, 0644)
}

// BackupDir walks a directory recursively and calls BackupFile for each file.
func BackupDir(dirPath string, sessionTimestamp string) error {
	root, err := backupsRoot()
	if err != nil {
		return err
	}

	rel, err := relativeToHome(dirPath)
	if err != nil {
		return err
	}

	destDir := filepath.Join(root, sessionTimestamp, rel)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := BackupDir(fullPath, sessionTimestamp); err != nil {
				return err
			}
		} else {
			if err := BackupFile(fullPath, sessionTimestamp); err != nil {
				return err
			}
		}
	}

	return nil
}

// ListBackups returns all backup directories sorted by timestamp descending (newest first).
// Returns an empty slice (not nil) when the backups directory does not exist.
func ListBackups() ([]Backup, error) {
	root, err := backupsRoot()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []Backup{}, nil
		}
		return nil, err
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, Backup{
				Timestamp: entry.Name(),
				Path:      filepath.Join(root, entry.Name()),
			})
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})

	return backups, nil
}

// DeleteBackup removes a backup directory identified by its timestamp string.
// Returns an error if the timestamp does not exist.
func DeleteBackup(timestamp string) error {
	root, err := backupsRoot()
	if err != nil {
		return err
	}

	dir := filepath.Join(root, timestamp)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup %q does not exist", timestamp)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%q is not a backup directory", timestamp)
	}

	return os.RemoveAll(dir)
}
