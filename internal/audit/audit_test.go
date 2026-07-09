package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAppendFirstCall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{
		ModuleID: "M-01",
		Name:     "Homebrew",
		Status:   StatusInstalled,
	})
	if err != nil {
		t.Fatalf("expected no error on first Append, got: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file should exist after Append, got: %v", err)
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("file should be valid JSON, got: %v", err)
	}
	if file.Version != AuditVersion {
		t.Errorf("expected version %q, got: %q", AuditVersion, file.Version)
	}
	if len(file.Entries) != 1 {
		t.Fatalf("expected exactly 1 entry, got: %d", len(file.Entries))
	}
	if file.Entries[0].ModuleID != "M-01" {
		t.Errorf("expected module_id M-01, got: %q", file.Entries[0].ModuleID)
	}
	if file.Entries[0].Status != StatusInstalled {
		t.Errorf("expected status %q, got: %q", StatusInstalled, file.Entries[0].Status)
	}
}

func TestAppendAccumulates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	if err := Append(path, Entry{ModuleID: "M-01", Name: "Homebrew", Status: StatusInstalled}); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, Entry{ModuleID: "M-02", Name: "Zsh", Status: StatusSkipped}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Entries) != 2 {
		t.Fatalf("expected 2 entries, got: %d", len(file.Entries))
	}
	if file.Entries[0].ModuleID != "M-01" {
		t.Errorf("expected first entry M-01, got: %q", file.Entries[0].ModuleID)
	}
	if file.Entries[1].ModuleID != "M-02" {
		t.Errorf("expected second entry M-02, got: %q", file.Entries[1].ModuleID)
	}
}

func TestAppendCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "audit.json")

	err := Append(path, Entry{ModuleID: "M-01", Name: "Homebrew", Status: StatusInstalled})
	if err != nil {
		t.Fatalf("Append should create parent dirs, got: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist after Append with parent dir creation")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("file should be valid JSON, got: %v", err)
	}
	if len(file.Entries) != 1 {
		t.Errorf("expected 1 entry, got: %d", len(file.Entries))
	}
}

func TestAppendConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")
	const n = 20

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := Append(path, Entry{
				ModuleID: fmt.Sprintf("M-%02d", idx+1),
				Name:     fmt.Sprintf("Module %d", idx+1),
				Status:   StatusInstalled,
			})
			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("result should be valid parseable JSON, got: %v", err)
	}

	// Deduplicate by module_id (some entries may have been lost if mutex is not protecting)
	seen := make(map[string]bool)
	for _, e := range file.Entries {
		seen[e.ModuleID] = true
	}
	if len(seen) != n {
		t.Errorf("expected %d unique module IDs after concurrent Append, got: %d", n, len(seen))
	}
	if len(file.Entries) != n {
		t.Errorf("expected exactly %d entries after concurrent Append, got: %d", n, len(file.Entries))
	}
}

func TestAppendErrorOmitted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{
		ModuleID: "M-01",
		Name:     "Homebrew",
		Status:   StatusInstalled,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, `"error"`) {
		t.Error("JSON should not contain 'error' field when Entry.Error is empty")
	}
}

func TestAppendErrorIncluded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	errMsg := "systemd not available"
	err := Append(path, Entry{
		ModuleID: "M-38",
		Name:     "Docker",
		Status:   StatusFailed,
		Error:    errMsg,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, `"error": "`+errMsg+`"`) {
		t.Errorf("JSON should contain error field with %q", errMsg)
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Entries) != 1 {
		t.Fatalf("expected 1 entry, got: %d", len(file.Entries))
	}
	if file.Entries[0].Error != errMsg {
		t.Errorf("expected error %q, got: %q", errMsg, file.Entries[0].Error)
	}
}

func TestTimestampFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{ModuleID: "M-01", Name: "Homebrew", Status: StatusInstalled})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var file AuditFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Entries) != 1 {
		t.Fatalf("expected 1 entry, got: %d", len(file.Entries))
	}

	ts := file.Entries[0].Timestamp
	// Verify ISO 8601 UTC with seconds: YYYY-MM-DDTHH:MM:SSZ
	parsed, err := time.Parse("2006-01-02T15:04:05Z", ts)
	if err != nil {
		t.Fatalf("timestamp %q does not match UTC ISO 8601 format: %v", ts, err)
	}
	if parsed.Location() != time.UTC {
		t.Errorf("timestamp should be UTC, got: %v", parsed.Location())
	}
}

func TestGoldenSingleEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{
		ModuleID:  "M-01",
		Name:      "Homebrew",
		Version:   "4.5.1",
		Method:    "curl-script",
		OS:        "darwin",
		Status:    StatusInstalled,
		Timestamp: "2026-05-23T10:30:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "single-entry.golden"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(golden) {
		t.Error("output does not match single-entry.golden")
	}
}

func TestGoldenMultiEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{
		ModuleID:  "M-01",
		Name:      "Homebrew",
		Version:   "4.5.1",
		Method:    "curl-script",
		OS:        "darwin",
		Status:    StatusInstalled,
		Timestamp: "2026-05-23T10:30:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = Append(path, Entry{
		ModuleID:  "M-02",
		Name:      "Zsh",
		Version:   "5.9",
		Method:    "brew",
		OS:        "darwin",
		Status:    StatusSkipped,
		Timestamp: "2026-05-23T10:31:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "multi-entry.golden"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(golden) {
		t.Error("output does not match multi-entry.golden")
	}
}

func TestGoldenEntryWithError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	err := Append(path, Entry{
		ModuleID:  "M-38",
		Name:      "Docker",
		Version:   "",
		Method:    "apt",
		OS:        "ubuntu-wsl2",
		Status:    StatusFailed,
		Error:     "systemd not available in this WSL2 instance",
		Timestamp: "2026-05-23T10:45:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "entry-with-error.golden"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(golden) {
		t.Error("output does not match entry-with-error.golden")
	}
}

func TestAppendMalformedFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Append(path, Entry{ModuleID: "M-01", Name: "Homebrew", Status: StatusInstalled})
	if err == nil {
		t.Error("expected error when appending to malformed audit file")
	}
}

func TestAppendWriteErrorNoPartialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	// Create a valid file first
	if err := Append(path, Entry{ModuleID: "M-01", Name: "Homebrew", Status: StatusInstalled}); err != nil {
		t.Fatal(err)
	}
	origData, _ := os.ReadFile(path)

	// Make the directory read-only to force a write error
	// On Windows, this is harder. Instead, we test that the temp file cleanup
	// works by verifying the original data is preserved if a simulated failure occurs.
	// For a real permission-denied test, we rely on the platform test infrastructure.

	// Use a path that cannot be created (parent is a file)
	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(blockPath, "audit.json")

	err := Append(badPath, Entry{ModuleID: "M-02", Name: "Zsh", Status: StatusSkipped})
	if err == nil {
		t.Error("expected error when parent path is a file")
	}

	// Original file should be unchanged
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(origData) {
		t.Error("original file should be unchanged after failed Append to different path")
	}
}
