package taskevidence

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrIncompleteEvidence  = errors.New("task evidence is incomplete")
	ErrTaskNotFound        = errors.New("markdown task checkbox not found")
	ErrTaskAlreadyComplete = errors.New("markdown task is already complete")
	ErrTaskAmbiguous       = errors.New("markdown task checkbox is ambiguous")
)

// Evidence records the focused completion evidence required for a tracked task.
type Evidence struct {
	AuditPassed               bool
	RemediationRequired       bool
	RemediationPassed         bool
	FocusedVerificationPassed bool
}

// CanMarkComplete reports whether the required focused evidence permits a task checkbox to be checked.
func CanMarkComplete(evidence Evidence) bool {
	return evidence.AuditPassed &&
		(!evidence.RemediationRequired || evidence.RemediationPassed) &&
		evidence.FocusedVerificationPassed
}

// MarkResult describes whether a Markdown task checkbox was updated.
type MarkResult struct {
	Marked bool
}

// MarkCompleteInMarkdown checks focused evidence before changing exactly one Markdown task checkbox.
func MarkCompleteInMarkdown(path, taskID string, evidence Evidence) (MarkResult, error) {
	if !CanMarkComplete(evidence) {
		return MarkResult{}, fmt.Errorf("%w; %s remains unchecked", ErrIncompleteEvidence, taskID)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return MarkResult{}, err
	}

	lines := strings.Split(string(contents), "\n")
	match := -1
	state := byte(0)
	for i, line := range lines {
		checkboxState, found := markdownTaskCheckboxState(line, taskID)
		if !found {
			continue
		}
		if match >= 0 {
			return MarkResult{}, fmt.Errorf("%w: %s", ErrTaskAmbiguous, taskID)
		}
		match, state = i, checkboxState
	}

	if match < 0 {
		return MarkResult{}, fmt.Errorf("%w: %s", ErrTaskNotFound, taskID)
	}
	if state == 'x' || state == 'X' {
		return MarkResult{}, fmt.Errorf("%w: %s", ErrTaskAlreadyComplete, taskID)
	}

	checkbox := strings.Index(lines[match], "- [")
	lines[match] = lines[match][:checkbox+3] + "x" + lines[match][checkbox+4:]
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return MarkResult{}, err
	}
	return MarkResult{Marked: true}, nil
}

func markdownTaskCheckboxState(line, taskID string) (byte, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	prefix := "- ["
	label := "] **" + taskID + "**"
	if !strings.HasPrefix(trimmed, prefix) || len(trimmed) <= len(prefix) || trimmed[len(prefix)+1:] == "" {
		return 0, false
	}
	state := trimmed[len(prefix)]
	if state != ' ' && state != 'x' && state != 'X' {
		return 0, false
	}
	if !strings.HasPrefix(trimmed[len(prefix)+1:], label) {
		return 0, false
	}
	return state, true
}
