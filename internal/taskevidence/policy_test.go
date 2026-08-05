package taskevidence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCanMarkComplete(t *testing.T) {
	tests := []struct {
		name     string
		evidence Evidence
		want     bool
	}{
		{
			name: "incomplete focused verification remains unchecked",
			evidence: Evidence{
				AuditPassed:               true,
				RemediationRequired:       false,
				FocusedVerificationPassed: false,
			},
			want: false,
		},
		{
			name: "complete evidence earns a checked task",
			evidence: Evidence{
				AuditPassed:               true,
				RemediationRequired:       true,
				RemediationPassed:         true,
				FocusedVerificationPassed: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanMarkComplete(tt.evidence); got != tt.want {
				t.Errorf("CanMarkComplete(%+v) = %t, want %t", tt.evidence, got, tt.want)
			}
		})
	}
}

func TestMarkCompleteInMarkdown(t *testing.T) {
	const incompleteTracker = "# Tasks\n\n- [ ] **T-073** First task\n- [ ] **T-074** Unrelated task\n"
	const completeTracker = "# Tasks\n\n- [x] **T-073** First task\n- [ ] **T-074** Unrelated task\n"

	tests := []struct {
		name        string
		tracker     string
		taskID      string
		evidence    Evidence
		wantErr     error
		wantMarked  bool
		wantTracker string
	}{
		{
			name:        "incomplete evidence leaves target unchecked",
			tracker:     incompleteTracker,
			taskID:      "T-073",
			evidence:    Evidence{AuditPassed: true, FocusedVerificationPassed: false},
			wantErr:     ErrIncompleteEvidence,
			wantMarked:  false,
			wantTracker: incompleteTracker,
		},
		{
			name:        "complete evidence marks only the intended task",
			tracker:     incompleteTracker,
			taskID:      "T-073",
			evidence:    Evidence{AuditPassed: true, RemediationRequired: true, RemediationPassed: true, FocusedVerificationPassed: true},
			wantMarked:  true,
			wantTracker: completeTracker,
		},
		{
			name:        "missing target leaves tracker unchanged",
			tracker:     incompleteTracker,
			taskID:      "T-999",
			evidence:    Evidence{AuditPassed: true, FocusedVerificationPassed: true},
			wantErr:     ErrTaskNotFound,
			wantMarked:  false,
			wantTracker: incompleteTracker,
		},
		{
			name:        "already complete target leaves tracker unchanged",
			tracker:     completeTracker,
			taskID:      "T-073",
			evidence:    Evidence{AuditPassed: true, FocusedVerificationPassed: true},
			wantErr:     ErrTaskAlreadyComplete,
			wantMarked:  false,
			wantTracker: completeTracker,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tasks.md")
			if err := os.WriteFile(path, []byte(tt.tracker), 0o600); err != nil {
				t.Fatal(err)
			}

			result, err := MarkCompleteInMarkdown(path, tt.taskID, tt.evidence)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("MarkCompleteInMarkdown() error = %v, want %v", err, tt.wantErr)
			}
			if result.Marked != tt.wantMarked {
				t.Errorf("MarkCompleteInMarkdown() marked = %t, want %t", result.Marked, tt.wantMarked)
			}

			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(got) != tt.wantTracker {
				t.Errorf("tracker = %q, want %q", got, tt.wantTracker)
			}
		})
	}
}

func TestMarkCompleteInMarkdownAtomicWriteFailuresLeaveTrackerUnchanged(t *testing.T) {
	const tracker = "# Tasks\n\n- [ ] **T-073** First task\n"
	completeEvidence := Evidence{AuditPassed: true, FocusedVerificationPassed: true}

	tests := []struct {
		name         string
		configureOps func(t *testing.T, tempNames *[]string)
	}{
		{
			name: "write failure",
			configureOps: func(t *testing.T, tempNames *[]string) {
				markdownAtomicWriteOps.createTemp = func(dir, pattern string) (atomicWriteFile, error) {
					file, err := os.CreateTemp(dir, pattern)
					if err != nil {
						return nil, err
					}
					*tempNames = append(*tempNames, file.Name())
					return writeFailureFile{File: file}, nil
				}
			},
		},
		{
			name: "rename failure",
			configureOps: func(t *testing.T, tempNames *[]string) {
				markdownAtomicWriteOps.createTemp = func(dir, pattern string) (atomicWriteFile, error) {
					file, err := os.CreateTemp(dir, pattern)
					if err == nil {
						*tempNames = append(*tempNames, file.Name())
					}
					return file, err
				}
				markdownAtomicWriteOps.rename = func(_, _ string) error {
					return errors.New("forced rename failure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "tasks.md")
			if err := os.WriteFile(path, []byte(tracker), 0o640); err != nil {
				t.Fatal(err)
			}

			originalOps := markdownAtomicWriteOps
			t.Cleanup(func() { markdownAtomicWriteOps = originalOps })
			var tempNames []string
			tt.configureOps(t, &tempNames)

			if _, err := MarkCompleteInMarkdown(path, "T-073", completeEvidence); err == nil {
				t.Fatal("MarkCompleteInMarkdown() error = nil, want forced write failure")
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tracker {
				t.Errorf("tracker = %q, want original %q", got, tracker)
			}
			for _, tempName := range tempNames {
				if _, err := os.Stat(tempName); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("temporary file %q remains after failure: %v", tempName, err)
				}
			}
		})
	}
}

func TestMarkCompleteInMarkdownPreservesTrackerPermissions(t *testing.T) {
	const tracker = "# Tasks\n\n- [ ] **T-073** First task\n"
	path := filepath.Join(t.TempDir(), "tasks.md")
	if err := os.WriteFile(path, []byte(tracker), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	originalOps := markdownAtomicWriteOps
	t.Cleanup(func() { markdownAtomicWriteOps = originalOps })
	var chmodMode os.FileMode
	markdownAtomicWriteOps.createTemp = func(dir, pattern string) (atomicWriteFile, error) {
		file, err := os.CreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return permissionRecordingFile{File: file, chmodMode: &chmodMode}, nil
	}

	if _, err := MarkCompleteInMarkdown(path, "T-073", Evidence{AuditPassed: true, FocusedVerificationPassed: true}); err != nil {
		t.Fatal(err)
	}
	if want := info.Mode().Perm(); chmodMode != want {
		t.Errorf("temporary file permissions = %04o, want %04o", chmodMode, want)
	}
}

type writeFailureFile struct {
	*os.File
}

func (writeFailureFile) Write([]byte) (int, error) {
	return 0, errors.New("forced write failure")
}

type permissionRecordingFile struct {
	*os.File
	chmodMode *os.FileMode
}

func (file permissionRecordingFile) Chmod(mode os.FileMode) error {
	*file.chmodMode = mode
	return file.File.Chmod(mode)
}
