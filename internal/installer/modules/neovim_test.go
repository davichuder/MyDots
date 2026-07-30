package modules

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// NeovimModule tests — T-049 (RED)
// ---------------------------------------------------------------------------

func TestNeovim_ID(t *testing.T) {
	m := NeovimModule{}
	if got := m.ID(); got != types.ModNeovim {
		t.Errorf("ID() = %v, want %v", got, types.ModNeovim)
	}
}

func TestNeovim_Name(t *testing.T) {
	m := NeovimModule{}
	if got := m.Name(); got != "Neovim" {
		t.Errorf("Name() = %q, want %q", got, "Neovim")
	}
}

func TestNeovim_Criticality(t *testing.T) {
	m := NeovimModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestNeovim_Dependencies(t *testing.T) {
	m := NeovimModule{}
	want := []types.ModuleID{types.ModHomebrew}
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

func TestNeovim_IsInstalled(t *testing.T) {
	t.Run("nvim on PATH and init.lua exists returns true", func(t *testing.T) {
		tmpDir := t.TempDir()
		initLua := filepath.Join(tmpDir, ".config", "nvim", "init.lua")
		if err := os.MkdirAll(filepath.Dir(initLua), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(initLua, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		m := NeovimModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when nvim is on PATH and init.lua exists")
		}
	})

	t.Run("nvim not on PATH returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		initLua := filepath.Join(tmpDir, ".config", "nvim", "init.lua")
		if err := os.MkdirAll(filepath.Dir(initLua), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(initLua, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		m := NeovimModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when nvim is not on PATH")
		}
	})

	t.Run("nvim on PATH but init.lua missing returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		m := NeovimModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when init.lua does not exist")
		}
	})
}

// --- Install ---

func TestNeovim_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("binary missing, no config dir — installs brew, writes init.lua", func(t *testing.T) {
		tmpDir := t.TempDir()

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" {
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		// Track backup call
		var backupCalled bool
		origBackup := nvimBackupDir
		nvimBackupDir = func(dir, ts string) error {
			backupCalled = true
			return nil
		}
		t.Cleanup(func() { nvimBackupDir = origBackup })

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := NeovimModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify init.lua was written
		initLua := filepath.Join(tmpDir, ".config", "nvim", "init.lua")
		if _, err := os.Stat(initLua); err != nil {
			t.Errorf("init.lua was not written: %v", err)
		}

		// Backup should not be called when dir doesn't exist
		if backupCalled {
			t.Error("backup should not be called when nvim config dir does not exist")
		}
	})

	t.Run("binary exists, config dir exists — skips brew, backs up, writes config", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create existing config dir with some content
		existingFile := filepath.Join(tmpDir, ".config", "nvim", "old.lua")
		if err := os.MkdirAll(filepath.Dir(existingFile), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(existingFile, []byte("-- old config"), 0644); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		var backupDir string
		origBackup := nvimBackupDir
		nvimBackupDir = func(dir, ts string) error {
			backupDir = dir
			return nil
		}
		t.Cleanup(func() { nvimBackupDir = origBackup })

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := NeovimModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify init.lua was written
		initLua := filepath.Join(tmpDir, ".config", "nvim", "init.lua")
		if _, err := os.Stat(initLua); err != nil {
			t.Errorf("init.lua was not written: %v", err)
		}

		// Backup should have been called with the nvim config dir
		wantDir := filepath.Join(tmpDir, ".config", "nvim")
		if backupDir != wantDir {
			t.Errorf("backup dir = %q, want %q", backupDir, wantDir)
		}
	})

	t.Run("backup fails aborts install", func(t *testing.T) {
		tmpDir := t.TempDir()
		existingFile := filepath.Join(tmpDir, ".config", "nvim", "old.lua")
		if err := os.MkdirAll(filepath.Dir(existingFile), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(existingFile, []byte("-- old config"), 0644); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		wantErr := errors.New("backup failed")
		origBackup := nvimBackupDir
		nvimBackupDir = func(dir, ts string) error {
			return wantErr
		}
		t.Cleanup(func() { nvimBackupDir = origBackup })

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := NeovimModule{}
		if err := m.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}

		// Verify init.lua was NOT written (install aborted)
		initLua := filepath.Join(tmpDir, ".config", "nvim", "init.lua")
		if _, err := os.Stat(initLua); err == nil {
			t.Error("init.lua was written despite backup failure")
		}
	})
}

// --- AuditInfo ---

func TestNeovim_AuditInfo(t *testing.T) {
	t.Run("returns nvim version string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("NVIM v0.10.0\n"), nil
			},
		})
		m := NeovimModule{}
		got := m.AuditInfo()
		want := "NVIM v0.10.0"
		if got != want {
			t.Errorf("AuditInfo() = %q, want %q", got, want)
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := NeovimModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
