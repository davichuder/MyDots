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
// PhpModule tests — T-071 (RED)
// ---------------------------------------------------------------------------

func TestPhpModule_ID(t *testing.T) {
	m := PhpModule{}
	if got := m.ID(); got != types.ModPhp {
		t.Errorf("ID() = %v, want %v", got, types.ModPhp)
	}
}

func TestPhpModule_Name(t *testing.T) {
	m := PhpModule{}
	if got := m.Name(); got != "PHP" {
		t.Errorf("Name() = %q, want %q", got, "PHP")
	}
}

func TestPhpModule_Criticality(t *testing.T) {
	m := PhpModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestPhpModule_Dependencies(t *testing.T) {
	m := PhpModule{}
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

func TestPhpModule_IsInstalled(t *testing.T) {
	t.Run("php on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := PhpModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when php is on PATH")
		}
	})

	t.Run("php not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := PhpModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when php is not on PATH")
		}
	})
}

// --- Install ---

func TestPhpModule_Install(t *testing.T) {
	t.Run("brew install php", func(t *testing.T) {
		origBrew := phpRunBrew
		t.Cleanup(func() { phpRunBrew = origBrew })

		var brewArgs []string
		phpRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			brewArgs = args
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := PhpModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if len(brewArgs) < 2 || brewArgs[0] != "install" || brewArgs[1] != "php" {
			t.Errorf("brew args = %v, want ['install', 'php']", brewArgs)
		}
	})

	t.Run("brew failure returns error", func(t *testing.T) {
		origBrew := phpRunBrew
		t.Cleanup(func() { phpRunBrew = origBrew })

		phpRunBrew = func(_ context.Context, _ io.Writer, args ...string) error {
			return errors.New("brew failed")
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := PhpModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// --- AuditInfo ---

func TestPhpModule_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("PHP 8.4.0\n"), nil
		},
	})
	m := PhpModule{}
	got := m.AuditInfo()
	if got != "PHP 8.4.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "PHP 8.4.0")
	}
}
