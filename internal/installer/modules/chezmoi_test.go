package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// ChezmoiModule tests — T-067 (RED)
// ---------------------------------------------------------------------------

func TestChezmoi_ID(t *testing.T) {
	m := ChezmoiModule{}
	if got := m.ID(); got != types.ModChezmoi {
		t.Errorf("ID() = %v, want %v", got, types.ModChezmoi)
	}
}

func TestChezmoi_Name(t *testing.T) {
	m := ChezmoiModule{}
	if got := m.Name(); got != "Chezmoi" {
		t.Errorf("Name() = %q, want %q", got, "Chezmoi")
	}
}

func TestChezmoi_Criticality(t *testing.T) {
	m := ChezmoiModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestChezmoi_Dependencies(t *testing.T) {
	m := ChezmoiModule{}
	deps := m.Dependencies()
	expected := []types.ModuleID{types.ModHomebrew, types.ModGit}
	if len(deps) != len(expected) {
		t.Fatalf("Dependencies() = %v, want %v", deps, expected)
	}
	for i := range expected {
		if deps[i] != expected[i] {
			t.Errorf("Dependencies()[%d] = %v, want %v", i, deps[i], expected[i])
		}
	}
}

// --- IsInstalled ---

func TestChezmoi_IsInstalled(t *testing.T) {
	t.Run("binary on PATH and chezmoi dir with git repo returns true", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		// Create ~/.local/share/chezmoi with a .git dir (simulating initialized repo)
		chezmoiDir := filepath.Join(tmpDir, ".local", "share", "chezmoi")
		if err := os.MkdirAll(filepath.Join(chezmoiDir, ".git"), 0755); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "chezmoi" {
					return "/usr/bin/chezmoi", nil
				}
				return "", errors.New("not found")
			},
		})

		m := ChezmoiModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when binary + chezmoi dir with .git exist")
		}
	})

	t.Run("binary not on PATH returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		// Create chezmoi dir but no binary
		chezmoiDir := filepath.Join(tmpDir, ".local", "share", "chezmoi")
		if err := os.MkdirAll(filepath.Join(chezmoiDir, ".git"), 0755); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})

		m := ChezmoiModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when chezmoi binary not on PATH")
		}
	})

	t.Run("chezmoi dir missing returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/chezmoi", nil
			},
		})

		m := ChezmoiModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when chezmoi dir does not exist")
		}
	})

	t.Run("chezmoi dir without .git returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		// Create chezmoi dir but WITHOUT .git subdirectory
		chezmoiDir := filepath.Join(tmpDir, ".local", "share", "chezmoi")
		if err := os.MkdirAll(chezmoiDir, 0755); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/chezmoi", nil
			},
		})

		m := ChezmoiModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when chezmoi dir has no .git")
		}
	})
}

// --- Install ---

func TestChezmoi_Install(t *testing.T) {
	t.Run("install sequence: brew install, chezmoi init, chezmoi apply", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		origBrew := chezmoiRunBrew
		origRun := chezmoiRun
		t.Cleanup(func() {
			chezmoiRunBrew = origBrew
			chezmoiRun = origRun
		})

		var commands []string
		chezmoiRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			commands = append(commands, "brew "+args[0])
			return nil
		}
		chezmoiRun = func(_ context.Context, _ io.Writer, name string, args ...string) error {
			commands = append(commands, name+" "+args[0])
			return nil
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config: config.Config{
				Chezmoi: config.ChezmoiOptions{
					RepoURL: "https://github.com/davichuder/dotfiles",
				},
			},
		}
		m := ChezmoiModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if len(commands) < 3 {
			t.Fatalf("expected 3 commands, got %d: %v", len(commands), commands)
		}

		expected := []string{"brew install", "chezmoi init", "chezmoi apply"}
		for i, exp := range expected {
			if commands[i] != exp {
				t.Errorf("command[%d] = %q, want %q", i, commands[i], exp)
			}
		}
	})

	t.Run("repo_url is read from ctx.Config.Chezmoi.RepoURL", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		origBrew := chezmoiRunBrew
		origRun := chezmoiRun
		t.Cleanup(func() {
			chezmoiRunBrew = origBrew
			chezmoiRun = origRun
		})

		var initArgs []string
		chezmoiRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			return nil
		}
		chezmoiRun = func(_ context.Context, _ io.Writer, name string, args ...string) error {
			if name == "chezmoi" && len(args) > 0 && args[0] == "init" {
				initArgs = args
			}
			return nil
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config: config.Config{
				Chezmoi: config.ChezmoiOptions{
					RepoURL: "https://github.com/user/dotfiles",
				},
			},
		}
		m := ChezmoiModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if len(initArgs) < 2 || initArgs[1] != "https://github.com/user/dotfiles" {
			t.Errorf("chezmoi init args = %v, want repo URL at [1]", initArgs)
		}
	})

	t.Run("chezmoi init failure skips apply", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := chezmoiHomeDir
		chezmoiHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { chezmoiHomeDir = origHome })

		origBrew := chezmoiRunBrew
		origRun := chezmoiRun
		t.Cleanup(func() {
			chezmoiRunBrew = origBrew
			chezmoiRun = origRun
		})

		var applyCalled bool
		chezmoiRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			return nil
		}
		chezmoiRun = func(_ context.Context, _ io.Writer, name string, args ...string) error {
			if name == "chezmoi" && len(args) > 0 && args[0] == "init" {
				return errors.New("init failed: repo unreachable")
			}
			if name == "chezmoi" && len(args) > 0 && args[0] == "apply" {
				applyCalled = true
			}
			return nil
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config: config.Config{
				Chezmoi: config.ChezmoiOptions{
					RepoURL: "https://github.com/davichuder/dotfiles",
				},
			},
		}
		m := ChezmoiModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error when chezmoi init fails, got nil")
		}
		if applyCalled {
			t.Error("chezmoi apply should not be called when init fails")
		}
	})
}

// --- AuditInfo ---

func TestChezmoi_AuditInfo(t *testing.T) {
	t.Run("returns chezmoi version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("chezmoi 2.60.0\n"), nil
			},
		})
		m := ChezmoiModule{}
		got := m.AuditInfo()
		if got != "chezmoi 2.60.0" {
			t.Errorf("AuditInfo() = %q, want %q", got, "chezmoi 2.60.0")
		}
	})
}
