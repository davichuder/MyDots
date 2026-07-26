// Package runner provides shell command execution for the install pipeline.
//
// The runner wraps exec.CommandContext with line-by-line output streaming,
// mock injection via SetExecutor, and convenience wrappers for Homebrew
// and embedded shell scripts.
package runner

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/davichuder/MyDots/internal/platform"
)

// ErrNotDarwin is returned by BrewCask when called on a non-Darwin platform.
var ErrNotDarwin = errors.New("brew cask is only available on macOS")

// Executor abstracts shell command execution for testing.
// Tests supply a mock via SetExecutor to record calls and return
// configured output without running real commands.
type Executor interface {
	// Execute runs a command with combined stdout+stderr and returns
	// the raw output. The error is non-nil when the exit code is non-zero
	// or the command could not be started.
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPath finds an executable in the system PATH.
	LookPath(name string) (string, error)
}

// osExecutor is the production executor that delegates to os/exec.
type osExecutor struct{}

func (osExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

func (osExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// currentExecutor is the globally active executor. Replaced by SetExecutor in tests.
var currentExecutor Executor = osExecutor{}

// createTempFile is an injectable os.CreateTemp for testing Script's temp file behaviour.
var createTempFile = func(pattern string) (*os.File, error) {
	return os.CreateTemp("", pattern)
}

// removeFile is an injectable os.Remove for testing Script's cleanup.
var removeFile = os.Remove

// SetExecutor replaces the current executor. Call in test setup to inject a mock.
func SetExecutor(e Executor) {
	currentExecutor = e
}

// Run executes a command and writes each output line to logw.
// Uses the active executor so it is fully mockable in tests.
func Run(ctx context.Context, logw io.Writer, name string, args ...string) error {
	out, err := currentExecutor.Execute(ctx, name, args...)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		_, _ = fmt.Fprintln(logw, scanner.Text())
	}
	return err
}

// Brew is a shorthand for Run(ctx, logw, "brew", args...).
func Brew(ctx context.Context, logw io.Writer, args ...string) error {
	return Run(ctx, logw, "brew", args...)
}

// BrewCask installs a Homebrew cask. Returns ErrNotDarwin when the
// platform is not Darwin.
func BrewCask(ctx context.Context, logw io.Writer, p platform.Platform, caskName string) error {
	if p.OS != platform.Darwin {
		return ErrNotDarwin
	}
	return Run(ctx, logw, "brew", "--cask", caskName)
}

// BrewTap taps a Homebrew tap and installs the formula in two steps:
//   brew tap <tap>
//   brew install <formula>
func BrewTap(ctx context.Context, logw io.Writer, tap, formula string) error {
	if err := Run(ctx, logw, "brew", "tap", tap); err != nil {
		return err
	}
	return Run(ctx, logw, "brew", "install", formula)
}

// Script extracts an embedded shell script to a temp file and executes it
// with /bin/sh. The temp file is created with 0700 permissions and removed
// after execution (deferred cleanup). Environment variables from env are
// merged into the process environment; new values take precedence over
// existing ones.
func Script(ctx context.Context, logw io.Writer, fSys fs.FS, path string, env map[string]string) error {
	data, err := fs.ReadFile(fSys, path)
	if err != nil {
		return fmt.Errorf("read embedded script %s: %w", path, err)
	}

	tmp, err := createTempFile("mydots-*.sh")
	if err != nil {
		return fmt.Errorf("create temp script: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = removeFile(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp script: %w", err)
	}

	if err := tmp.Chmod(0700); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp script: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp script: %w", err)
	}

	// Merge environment: append new values so they take precedence.
	mergedEnv := os.Environ()
	for k, v := range env {
		mergedEnv = append(mergedEnv, k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", tmpPath)
	cmd.Env = mergedEnv

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		_, _ = fmt.Fprintln(logw, scanner.Text())
	}

	return runErr
}

// CommandExists checks whether a binary is reachable on PATH.
func CommandExists(name string) bool {
	_, err := currentExecutor.LookPath(name)
	return err == nil
}

// CaptureOutput runs a command and returns its trimmed stdout as a string.
// Returns an empty string if the command fails (non-zero exit).
func CaptureOutput(name string, args ...string) string {
	out, err := currentExecutor.Execute(context.Background(), name, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
