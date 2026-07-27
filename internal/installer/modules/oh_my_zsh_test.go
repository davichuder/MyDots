package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"testing/fstest"

	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// --- ID ---

func TestOhMyZsh_ID(t *testing.T) {
	m := OhMyZshModule{}
	if got := m.ID(); got != types.ModOhMyZsh {
		t.Errorf("ID() = %v, want %v", got, types.ModOhMyZsh)
	}
}

// --- Name ---

func TestOhMyZsh_Name(t *testing.T) {
	m := OhMyZshModule{}
	if got := m.Name(); got != "Oh My Zsh" {
		t.Errorf("Name() = %q, want %q", got, "Oh My Zsh")
	}
}

// --- Criticality ---

func TestOhMyZsh_Criticality(t *testing.T) {
	m := OhMyZshModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

// --- Dependencies ---

func TestOhMyZsh_Dependencies(t *testing.T) {
	m := OhMyZshModule{}
	want := []types.ModuleID{types.ModZsh}
	deps := m.Dependencies()
	if len(deps) != len(want) {
		t.Fatalf("Dependencies() = %v (len=%d), want %v (len=%d)", deps, len(deps), want, len(want))
	}
	for i := range want {
		if deps[i] != want[i] {
			t.Errorf("Dependencies()[%d] = %v, want %v", i, deps[i], want[i])
		}
	}
}

// --- IsInstalled ---

func TestOhMyZsh_IsInstalled(t *testing.T) {
	t.Run("oh-my-zsh directory exists", func(t *testing.T) {
		orig := checkOhMyZshDir
		checkOhMyZshDir = func() bool { return true }
		t.Cleanup(func() { checkOhMyZshDir = orig })

		m := OhMyZshModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when ~/.oh-my-zsh exists")
		}
	})

	t.Run("oh-my-zsh directory not found", func(t *testing.T) {
		orig := checkOhMyZshDir
		checkOhMyZshDir = func() bool { return false }
		t.Cleanup(func() { checkOhMyZshDir = orig })

		m := OhMyZshModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when ~/.oh-my-zsh does not exist")
		}
	})

	t.Run("other stat error returns false", func(t *testing.T) {
		orig := checkOhMyZshDir
		checkOhMyZshDir = func() bool { return false }
		t.Cleanup(func() { checkOhMyZshDir = orig })

		m := OhMyZshModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when stat returns an error")
		}
	})
}

// --- Install ---

func TestOhMyZsh_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("runs OMZ install script with env vars", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"assets/scripts/omz-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 0"),
				Mode: 0555,
			},
		}
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: mapFS,
		}
		m := OhMyZshModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("script error returns CurlScriptError", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"assets/scripts/omz-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 42"),
				Mode: 0555,
			},
		}
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: mapFS,
		}
		m := OhMyZshModule{}
		err := m.Install(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cerr myerr.CurlScriptError
		if !errors.As(err, &cerr) {
			t.Fatalf("expected CurlScriptError, got: %T", err)
		}
	})
}

// --- AuditInfo ---

func TestOhMyZsh_AuditInfo(t *testing.T) {
	t.Run("success returns OMZ version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("oh-my-zsh 2024-04-10\n"), nil
			},
		})
		m := OhMyZshModule{}
		got := m.AuditInfo()
		if got != "oh-my-zsh 2024-04-10" {
			t.Errorf("AuditInfo() = %q, want %q", got, "oh-my-zsh 2024-04-10")
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := OhMyZshModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
