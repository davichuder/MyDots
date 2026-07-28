package modules

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// GhosttyModule tests — T-059 (RED)
// ---------------------------------------------------------------------------

func TestGhostty_ID(t *testing.T) {
	m := GhosttyModule{}
	if got := m.ID(); got != types.ModGhostty {
		t.Errorf("ID() = %v, want %v", got, types.ModGhostty)
	}
}

func TestGhostty_Name(t *testing.T) {
	m := GhosttyModule{}
	if got := m.Name(); got != "Ghostty" {
		t.Errorf("Name() = %q, want %q", got, "Ghostty")
	}
}

func TestGhostty_Criticality(t *testing.T) {
	m := GhosttyModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestGhostty_Dependencies(t *testing.T) {
	m := GhosttyModule{}
	deps := m.Dependencies()
	if len(deps) != 1 || deps[0] != types.ModHomebrew {
		t.Errorf("Dependencies() = %v, want [%v]", deps, types.ModHomebrew)
	}
}

// --- IsInstalled ---

func TestGhostty_IsInstalled(t *testing.T) {
	t.Run("ghostty on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/ghostty", nil
			},
		})
		m := GhosttyModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when ghostty is on PATH")
		}
	})

	t.Run("ghostty not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := GhosttyModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when ghostty is not on PATH")
		}
	})
}

// --- Install ---

func TestGhostty_Install(t *testing.T) {
	skipIfWindows(t)

	successScript := fstest.MapFS{
		"assets/scripts/ghostty-linux.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\nexit 0"),
			Mode: 0555,
		},
	}

	t.Run("darwin installs via brew cask", func(t *testing.T) {
		var brewArgs []string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" {
					brewArgs = args
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		}
		m := GhosttyModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(brewArgs) < 2 || brewArgs[0] != "--cask" || brewArgs[1] != "ghostty" {
			t.Errorf("expected brew --cask ghostty, got brew %v", brewArgs)
		}
	})

	t.Run("ubuntu native runs ghostty-linux.sh script", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
			Assets:   successScript,
		}
		m := GhosttyModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("wsl2 with wayland runs ghostty-linux.sh script", func(t *testing.T) {
		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return "wayland-0"
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
			Assets:   successScript,
		}
		m := GhosttyModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("wsl2 no wayland logs warning and returns nil", func(t *testing.T) {
		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return ""
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		var executed bool
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				executed = true
				return nil, nil
			},
		})

		var logBuf bytes.Buffer
		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &logBuf,
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
		}
		m := GhosttyModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if executed {
			t.Error("Install should not execute commands when WSL2 has no Wayland")
		}
		logged := logBuf.String()
		if !strings.Contains(logged, "Wayland") && !strings.Contains(logged, "wayland") {
			t.Errorf("expected warning about Wayland, got log: %q", logged)
		}
	})
}

// --- AuditInfo ---

func TestGhostty_AuditInfo(t *testing.T) {
	t.Run("returns ghostty version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("ghostty 1.0.0\n"), nil
			},
		})
		m := GhosttyModule{}
		got := m.AuditInfo()
		if got != "ghostty 1.0.0" {
			t.Errorf("AuditInfo() = %q, want %q", got, "ghostty 1.0.0")
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := GhosttyModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
