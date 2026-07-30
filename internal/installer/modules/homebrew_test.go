package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// mockExecutor implements runner.Executor for testing.
type mockExecutor struct {
	executeFunc  func(ctx context.Context, name string, args ...string) ([]byte, error)
	lookPathFunc func(name string) (string, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, name, args...)
	}
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func (m *mockExecutor) LookPath(name string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(name)
	}
	return exec.LookPath(name)
}

// defaultExecutor restores the runner to OS-level command execution.
type defaultExecutor struct{}

func (defaultExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func (defaultExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// withMockExecutor sets a mock executor for the test and restores to OS-level on cleanup.
func withMockExecutor(t *testing.T, e runner.Executor) {
	t.Helper()
	runner.SetExecutor(e)
	t.Cleanup(func() { runner.SetExecutor(defaultExecutor{}) })
}

// skipIfWindows skips the test on Windows because runner.Script uses /bin/sh.
func skipIfWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: /bin/sh not available")
	}
}

// --- ID ---

func TestHomebrew_ID(t *testing.T) {
	h := HomebrewModule{}
	if got := h.ID(); got != types.ModHomebrew {
		t.Errorf("ID() = %v, want %v", got, types.ModHomebrew)
	}
}

// --- Name ---

func TestHomebrew_Name(t *testing.T) {
	h := HomebrewModule{}
	if got := h.Name(); got != "Homebrew" {
		t.Errorf("Name() = %q, want %q", got, "Homebrew")
	}
}

// --- Criticality ---

func TestHomebrew_Criticality(t *testing.T) {
	h := HomebrewModule{}
	if got := h.Criticality(); got != types.Critical {
		t.Errorf("Criticality() = %v, want %v", got, types.Critical)
	}
}

// --- Dependencies ---

func TestHomebrew_Dependencies(t *testing.T) {
	h := HomebrewModule{}
	if deps := h.Dependencies(); deps != nil {
		t.Errorf("Dependencies() = %v, want nil", deps)
	}
}

// --- IsInstalled ---

func TestHomebrew_IsInstalled(t *testing.T) {
	t.Run("brew found on PATH", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		h := HomebrewModule{}
		p := platform.Platform{OS: platform.Darwin}
		if !h.IsInstalled(p) {
			t.Error("expected true when brew is on PATH")
		}
	})

	t.Run("brew not found on PATH", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		h := HomebrewModule{}
		p := platform.Platform{OS: platform.Darwin}
		if h.IsInstalled(p) {
			t.Error("expected false when brew is not on PATH")
		}
	})
}

// --- Install ---

func TestHomebrew_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("runs script on darwin", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) { return "/opt/homebrew/bin/" + name, nil },
		})
		mapFS := fstest.MapFS{
			"assets/scripts/homebrew-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 0"),
				Mode: 0555,
			},
		}
		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Assets:   mapFS,
			Platform: platform.Platform{OS: platform.Darwin},
		}
		h := HomebrewModule{}
		err := h.Install(ctx)
		if err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("runs script on linux", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) { return "/home/linuxbrew/.linuxbrew/bin/" + name, nil },
		})
		mapFS := fstest.MapFS{
			"assets/scripts/homebrew-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 0"),
				Mode: 0555,
			},
		}
		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Assets:   mapFS,
			Platform: platform.Platform{OS: platform.Linux},
		}
		h := HomebrewModule{}
		err := h.Install(ctx)
		if err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("script error returns CurlScriptError with What/Why/Fix", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"assets/scripts/homebrew-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 42"),
				Mode: 0555,
			},
		}
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: mapFS,
		}
		h := HomebrewModule{}
		err := h.Install(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cerr myerr.CurlScriptError
		if !errors.As(err, &cerr) {
			t.Fatalf("expected CurlScriptError, got: %T", err)
		}
		if cerr.ExitCode != 42 {
			t.Errorf("expected ExitCode 42, got: %d", cerr.ExitCode)
		}
		if cerr.What() == "" {
			t.Error("What() should not be empty")
		}
		if cerr.Why() == "" {
			t.Error("Why() should not be empty")
		}
		if cerr.Fix() == "" {
			t.Error("Fix() should not be empty")
		}
	})
}

func TestHomebrew_InstallRefreshesBrewForDependentCommands(t *testing.T) {
	originalScript := runHomebrewScript
	runHomebrewScript = func(context.Context, io.Writer, fs.FS, string, map[string]string) error { return nil }
	t.Cleanup(func() { runHomebrewScript = originalScript })

	var command string
	var args []string
	withMockExecutor(t, &mockExecutor{
		lookPathFunc: func(name string) (string, error) {
			if name == "/opt/homebrew/bin/brew" {
				return name, nil
			}
			return "", errors.New("not found")
		},
		executeFunc: func(_ context.Context, name string, gotArgs ...string) ([]byte, error) {
			command, args = name, gotArgs
			return nil, nil
		},
	})

	brewPath := ""
	session := runner.WithBrewPath(context.Background(), &brewPath)
	ctx := types.InstallContext{
		Cancel:   session,
		Log:      &bytes.Buffer{},
		Platform: platform.Platform{OS: platform.Darwin},
		BrewPath: &brewPath,
	}
	if err := (HomebrewModule{}).Install(ctx); err != nil {
		t.Fatalf("Install() = %v", err)
	}
	if err := runner.Brew(ctx.Cancel, ctx.Log, "install", "zoxide"); err != nil {
		t.Fatalf("dependent Brew() = %v", err)
	}
	if command != "/opt/homebrew/bin/brew" || !slices.Equal(args, []string{"install", "zoxide"}) {
		t.Errorf("dependent brew command = %q %v, want %q %v", command, args, "/opt/homebrew/bin/brew", []string{"install", "zoxide"})
	}
}

func TestHomebrew_InstallReturnsErrorWhenBrewCannotBeDiscovered(t *testing.T) {
	originalScript := runHomebrewScript
	runHomebrewScript = func(context.Context, io.Writer, fs.FS, string, map[string]string) error { return nil }
	t.Cleanup(func() { runHomebrewScript = originalScript })

	withMockExecutor(t, &mockExecutor{
		lookPathFunc: func(string) (string, error) { return "", errors.New("not found") },
	})

	err := (HomebrewModule{}).Install(types.InstallContext{
		Cancel:   context.Background(),
		Log:      &bytes.Buffer{},
		Platform: platform.Platform{OS: platform.Linux},
		BrewPath: new(string),
	})
	if err == nil || !strings.Contains(err.Error(), "discover Homebrew") {
		t.Fatalf("Install() = %v, want discovery failure", err)
	}
}

// --- AuditInfo ---

func TestHomebrew_AuditInfo(t *testing.T) {
	t.Run("success returns trimmed brew version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("Homebrew 4.4.0\n"), nil
			},
		})
		h := HomebrewModule{}
		got := h.AuditInfo()
		if got != "Homebrew 4.4.0" {
			t.Errorf("AuditInfo() = %q, want %q", got, "Homebrew 4.4.0")
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		h := HomebrewModule{}
		got := h.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
