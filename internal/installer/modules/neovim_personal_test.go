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
// NeovimPersonalModule tests — T-051 (RED)
// ---------------------------------------------------------------------------

func TestNeovimPersonal_ID(t *testing.T) {
	m := NeovimPersonalModule{}
	if got := m.ID(); got != types.ModNeovimPersonal {
		t.Errorf("ID() = %v, want %v", got, types.ModNeovimPersonal)
	}
}

func TestNeovimPersonal_Name(t *testing.T) {
	m := NeovimPersonalModule{}
	if got := m.Name(); got != "Neovim Personal" {
		t.Errorf("Name() = %q, want %q", got, "Neovim Personal")
	}
}

func TestNeovimPersonal_Criticality(t *testing.T) {
	m := NeovimPersonalModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestNeovimPersonal_Dependencies(t *testing.T) {
	m := NeovimPersonalModule{}
	got := m.Dependencies()
	// Must include both M-17 (Neovim base) and M-45 (chezmoi).
	// Order is not guaranteed but both must be present.
	if len(got) != 2 {
		t.Fatalf("Dependencies() = %v (len=%d), want 2 deps", got, len(got))
	}
	seen := map[types.ModuleID]bool{}
	for _, d := range got {
		seen[d] = true
	}
	if !seen[types.ModNeovim] {
		t.Error("Dependencies() should include ModNeovim (M-17)")
	}
	if !seen[types.ModChezmoi] {
		t.Error("Dependencies() should include ModChezmoi (M-45)")
	}
}

// --- IsInstalled ---

func TestNeovimPersonal_IsInstalled(t *testing.T) {
	t.Run("chezmoi status returns empty — installed", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte(""), nil
			},
		})
		m := NeovimPersonalModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when chezmoi status output is empty")
		}
	})

	t.Run("chezmoi status returns diff — not installed", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte(" M .config/nvim/init.lua\n"), nil
			},
		})
		m := NeovimPersonalModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when chezmoi status shows diff")
		}
	})

	t.Run("chezmoi status errors — not installed", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("chezmoi not initialized")
			},
		})
		m := NeovimPersonalModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when chezmoi status errors")
		}
	})
}

// --- Install ---

func TestNeovimPersonal_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("backs up nvim dir then runs chezmoi apply", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create nvim config dir so backup is triggered
		nvimDir := filepath.Join(tmpDir, ".config", "nvim")
		if err := os.MkdirAll(nvimDir, 0755); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "chezmoi" {
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		var backupCalled bool
		var backupPath string
		origBackup := nvimBackupDir
		nvimBackupDir = func(dir, ts string) error {
			backupCalled = true
			backupPath = dir
			return nil
		}
		t.Cleanup(func() { nvimBackupDir = origBackup })

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := NeovimPersonalModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if !backupCalled {
			t.Error("expected backup to be called")
		}
		if backupPath != nvimDir {
			t.Errorf("backup path = %q, want %q", backupPath, nvimDir)
		}
	})

	t.Run("backup fails aborts install", func(t *testing.T) {
		tmpDir := t.TempDir()
		nvimDir := filepath.Join(tmpDir, ".config", "nvim")
		if err := os.MkdirAll(nvimDir, 0755); err != nil {
			t.Fatal(err)
		}

		withMockExecutor(t, &mockExecutor{})

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
		m := NeovimPersonalModule{}
		if err := m.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})
}

// --- AuditInfo ---

func TestNeovimPersonal_AuditInfo(t *testing.T) {
	m := NeovimPersonalModule{}
	if got := m.AuditInfo(); got != "personal" {
		t.Errorf("AuditInfo() = %q, want %q", got, "personal")
	}
}
