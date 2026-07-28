package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// GentleAiModule tests — T-069 (RED)
// ---------------------------------------------------------------------------

func TestGentleAi_ID(t *testing.T) {
	m := GentleAiModule{}
	if got := m.ID(); got != types.ModGentleAi {
		t.Errorf("ID() = %v, want %v", got, types.ModGentleAi)
	}
}

func TestGentleAi_Name(t *testing.T) {
	m := GentleAiModule{}
	if got := m.Name(); got != "Gentle AI" {
		t.Errorf("Name() = %q, want %q", got, "Gentle AI")
	}
}

func TestGentleAi_Criticality(t *testing.T) {
	m := GentleAiModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestGentleAi_Dependencies(t *testing.T) {
	m := GentleAiModule{}
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

func TestGentleAi_IsInstalled(t *testing.T) {
	t.Run("gentle-ai on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := GentleAiModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when gentle-ai is on PATH")
		}
	})

	t.Run("gentle-ai not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := GentleAiModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when gentle-ai is not on PATH")
		}
	})
}

// --- Install ---

func TestGentleAi_Install(t *testing.T) {
	t.Run("install sequence: brew tap, brew install", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := GentleAiModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("tap failure returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("Error: tap failed\n"), errors.New("tap failed")
			},
		})

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		m := GentleAiModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error when tap fails, got nil")
		}
	})
}

// --- AuditInfo ---

func TestGentleAi_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("gentle-ai 0.1.0\n"), nil
		},
	})
	m := GentleAiModule{}
	got := m.AuditInfo()
	if got != "gentle-ai 0.1.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "gentle-ai 0.1.0")
	}
}
