// Package audit provides concurrent-safe append-only audit logging for install sessions.
// Each Entry records a module's install status, version, and timestamp.
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const AuditVersion = "1.0.0"

type InstallStatus string

const (
	StatusInstalled               InstallStatus = "installed"
	StatusSkipped                 InstallStatus = "skipped"
	StatusSkippedDisabled         InstallStatus = "skipped-disabled"
	StatusSkippedDependencyFailed InstallStatus = "skipped-dependency-failed"
	StatusSkippedNoWayland        InstallStatus = "skipped-no-wayland"
	StatusFailed                  InstallStatus = "failed"
)

type AuditFile struct {
	Version string  `json:"version"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	ModuleID  string        `json:"module_id"`
	Name      string        `json:"name"`
	Version   string        `json:"version"`
	Method    string        `json:"method"`
	OS        string        `json:"os"`
	Status    InstallStatus `json:"status"`
	Error     string        `json:"error,omitempty"`
	Timestamp string        `json:"timestamp"`
}

var mu sync.Mutex

const tmpPattern = "mydots-audit-*.tmp"

func loadOrEmpty(path string) (AuditFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return AuditFile{Version: AuditVersion}, nil
		}
		return AuditFile{}, err
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		return AuditFile{}, err
	}
	return file, nil
}

func Append(path string, entry Entry) error {
	mu.Lock()
	defer mu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	file, err := loadOrEmpty(path)
	if err != nil {
		return err
	}

	file.Entries = append(file.Entries, entry)

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		_ = os.Remove(tmp.Name())
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	return nil
}
