package taskevidence

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type atomicWriteFile interface {
	Write([]byte) (int, error)
	Sync() error
	Close() error
	Chmod(os.FileMode) error
	Name() string
}

var markdownAtomicWriteOps = struct {
	createTemp func(dir, pattern string) (atomicWriteFile, error)
	rename     func(oldpath, newpath string) error
}{
	createTemp: func(dir, pattern string) (atomicWriteFile, error) {
		return os.CreateTemp(dir, pattern)
	},
	rename: os.Rename,
}

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
	if err := replaceMarkdownAtomically(path, []byte(strings.Join(lines, "\n"))); err != nil {
		return MarkResult{}, err
	}
	return MarkResult{Marked: true}, nil
}

func replaceMarkdownAtomically(path string, contents []byte) (err error) {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	temporary, err := markdownAtomicWriteOps.createTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	closed := false
	renamed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		if !renamed {
			_ = os.Remove(temporaryPath)
		}
	}()

	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if written, err := temporary.Write(contents); err != nil {
		return err
	} else if written != len(contents) {
		return io.ErrShortWrite
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	closed = true
	if err := markdownAtomicWriteOps.rename(temporaryPath, path); err != nil {
		return err
	}
	renamed = true
	return nil
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
