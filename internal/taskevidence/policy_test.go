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
