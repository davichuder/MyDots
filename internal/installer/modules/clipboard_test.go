package modules

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// ClipboardModule tests — T-055 (RED)
// ---------------------------------------------------------------------------

func TestClipboard_ID(t *testing.T) {
	m := ClipboardModule{}
	if got := m.ID(); got != types.ModClipboard {
		t.Errorf("ID() = %v, want %v", got, types.ModClipboard)
	}
}

func TestClipboard_Name(t *testing.T) {
	m := ClipboardModule{}
	if got := m.Name(); got != "Clipboard" {
		t.Errorf("Name() = %q, want %q", got, "Clipboard")
	}
}

func TestClipboard_Criticality(t *testing.T) {
	m := ClipboardModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestClipboard_Dependencies(t *testing.T) {
	m := ClipboardModule{}
	deps := m.Dependencies()
	if len(deps) != 1 || deps[0] != types.ModHomebrew {
		t.Errorf("Dependencies() = %v, want [%v]", deps, types.ModHomebrew)
	}
}

// --- IsInstalled ---

func TestClipboard_IsInstalled(t *testing.T) {
	t.Run("darwin always returns true", func(t *testing.T) {
		m := ClipboardModule{}
		if !m.IsInstalled(platform.Platform{OS: platform.Darwin}) {
			t.Error("expected true on Darwin")
		}
	})

	t.Run("linux with wl-copy returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "wl-copy" {
					return "/usr/bin/wl-copy", nil
				}
				return "", errors.New("not found")
			},
		})

		m := ClipboardModule{}
		if !m.IsInstalled(platform.Platform{OS: platform.Linux}) {
			t.Error("expected true when wl-copy is on PATH")
		}
	})

	t.Run("linux with xclip returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "xclip" {
					return "/usr/bin/xclip", nil
				}
				return "", errors.New("not found")
			},
		})

		m := ClipboardModule{}
		if !m.IsInstalled(platform.Platform{OS: platform.Linux}) {
			t.Error("expected true when xclip is on PATH")
		}
	})

	t.Run("linux with no clipboard tool returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})

		m := ClipboardModule{}
		if m.IsInstalled(platform.Platform{OS: platform.Linux}) {
			t.Error("expected false when neither wl-copy nor xclip is on PATH")
		}
	})
}

// --- Install ---

func TestClipboard_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("darwin is a no-op", func(t *testing.T) {
		var executed bool
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				executed = true
				return nil, nil
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		}
		m := ClipboardModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if executed {
			t.Error("Install should not execute any commands on Darwin")
		}
	})

	t.Run("ubuntu wayland installs wl-clipboard via brew", func(t *testing.T) {
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

		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return "wayland-0"
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := ClipboardModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(brewArgs) < 2 || brewArgs[0] != "install" || brewArgs[1] != "wl-clipboard" {
			t.Errorf("expected brew install wl-clipboard, got brew %v", brewArgs)
		}
	})

	t.Run("ubuntu no wayland installs xclip via brew", func(t *testing.T) {
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

		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return ""
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := ClipboardModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(brewArgs) < 2 || brewArgs[0] != "install" || brewArgs[1] != "xclip" {
			t.Errorf("expected brew install xclip, got brew %v", brewArgs)
		}
	})

	t.Run("wsl2 with wayland installs wl-clipboard via brew", func(t *testing.T) {
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

		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return "wayland-0"
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
		}
		m := ClipboardModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(brewArgs) < 2 || brewArgs[0] != "install" || brewArgs[1] != "wl-clipboard" {
			t.Errorf("expected brew install wl-clipboard, got brew %v", brewArgs)
		}
	})

	t.Run("wsl2 no wayland logs warning and returns nil", func(t *testing.T) {
		var executed bool
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				executed = true
				return nil, nil
			},
		})

		origLookup := lookupEnv
		lookupEnv = func(key string) string {
			if key == "WAYLAND_DISPLAY" {
				return ""
			}
			return origLookup(key)
		}
		t.Cleanup(func() { lookupEnv = origLookup })

		var logBuf bytes.Buffer
		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &logBuf,
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
		}
		m := ClipboardModule{}
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

func TestClipboard_AuditInfo(t *testing.T) {
	t.Run("darwin returns built-in", func(t *testing.T) {
		m := ClipboardModule{}
		got := m.AuditInfo()
		if got != "built-in" {
			t.Errorf("AuditInfo() = %q, want %q", got, "built-in")
		}
	})

	t.Run("returns wl-copy version when on PATH", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("wl-copy version 2.2.0\n"), nil
			},
		})
		m := ClipboardModule{}
		got := m.AuditInfo()
		if got != "wl-copy version 2.2.0" {
			t.Errorf("AuditInfo() = %q, want %q", got, "wl-copy version 2.2.0")
		}
	})

	t.Run("returns xclip version when wl-copy not available", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "wl-copy" {
					return []byte(""), errors.New("not found")
				}
				return []byte("xclip version 0.13\n"), nil
			},
		})
		m := ClipboardModule{}
		got := m.AuditInfo()
		if got != "xclip version 0.13" {
			t.Errorf("AuditInfo() = %q, want %q", got, "xclip version 0.13")
		}
	})
}
