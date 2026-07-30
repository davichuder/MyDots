package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// RtkModule tests — T-069 (RED)
// ---------------------------------------------------------------------------

func TestRtk_ID(t *testing.T) {
	m := RtkModule{}
	if got := m.ID(); got != types.ModRtk {
		t.Errorf("ID() = %v, want %v", got, types.ModRtk)
	}
}

func TestRtk_Name(t *testing.T) {
	m := RtkModule{}
	if got := m.Name(); got != "Rtk" {
		t.Errorf("Name() = %q, want %q", got, "Rtk")
	}
}

func TestRtk_Criticality(t *testing.T) {
	m := RtkModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestRtk_Dependencies(t *testing.T) {
	m := RtkModule{}
	deps := m.Dependencies()
	expected := []types.ModuleID{types.ModHomebrew}
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

func TestRtk_IsInstalled(t *testing.T) {
	t.Run("rtk on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := RtkModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when rtk is on PATH")
		}
	})

	t.Run("rtk not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := RtkModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when rtk is not on PATH")
		}
	})
}

// --- Install ---

func TestRtk_Install(t *testing.T) {
	t.Run("install sequence: brew install rtk, rtk init", func(t *testing.T) {
		origBrew := rtkRunBrew
		origRun := rtkRun
		t.Cleanup(func() {
			rtkRunBrew = origBrew
			rtkRun = origRun
		})

		var commands []string
		rtkRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			commands = append(commands, "brew "+args[0])
			return nil
		}
		rtkRun = func(_ context.Context, _ io.Writer, name string, args ...string) error {
			commands = append(commands, name+" "+args[0])
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := RtkModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if len(commands) < 2 {
			t.Fatalf("expected 2 commands, got %d: %v", len(commands), commands)
		}
		expected := []string{"brew install", "rtk init"}
		for i, exp := range expected {
			if commands[i] != exp {
				t.Errorf("command[%d] = %q, want %q", i, commands[i], exp)
			}
		}
	})

	t.Run("brew install failure skips rtk init", func(t *testing.T) {
		origBrew := rtkRunBrew
		origRun := rtkRun
		t.Cleanup(func() {
			rtkRunBrew = origBrew
			rtkRun = origRun
		})

		var initCalled bool
		rtkRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			return errors.New("brew failed")
		}
		rtkRun = func(_ context.Context, _ io.Writer, name string, args ...string) error {
			if name == "rtk" {
				initCalled = true
			}
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := RtkModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error when brew install fails, got nil")
		}
		if initCalled {
			t.Error("rtk init should not be called when brew install fails")
		}
	})
}

// --- AuditInfo ---

func TestRtk_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("rtk 1.0.0\n"), nil
		},
	})
	m := RtkModule{}
	got := m.AuditInfo()
	if got != "rtk 1.0.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "rtk 1.0.0")
	}
}
