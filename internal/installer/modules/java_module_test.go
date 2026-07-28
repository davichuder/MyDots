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
// JavaModule tests — T-071 (RED)
// ---------------------------------------------------------------------------

func TestJavaModule_ID(t *testing.T) {
	m := NewJavaModule(&mockRuntimeManager{})
	if got := m.ID(); got != types.ModJava {
		t.Errorf("ID() = %v, want %v", got, types.ModJava)
	}
}

func TestJavaModule_Name(t *testing.T) {
	m := NewJavaModule(&mockRuntimeManager{})
	if got := m.Name(); got != "Java" {
		t.Errorf("Name() = %q, want %q", got, "Java")
	}
}

func TestJavaModule_Criticality(t *testing.T) {
	m := NewJavaModule(&mockRuntimeManager{})
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestJavaModule_Dependencies(t *testing.T) {
	m := NewJavaModule(&mockRuntimeManager{})
	deps := m.Dependencies()
	expected := []types.ModuleID{types.ModSdkman}
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

func TestJavaModule_IsInstalled(t *testing.T) {
	t.Run("sdk list java shows installed version returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte(" 25.*installed 25-open /user/.sdkman/candidates/java/25-open\n"), nil
			},
		})
		m := NewJavaModule(&mockRuntimeManager{})
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when sdk list java shows installed version")
		}
	})

	t.Run("sdk list java shows no installed version returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte(" 21.*installed 21-open /user/.sdkman/candidates/java/21-open\n"), nil
			},
		})
		m := NewJavaModule(&mockRuntimeManager{})
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when only java 21 is installed")
		}
	})

	t.Run("sdk command fails returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("sdk not found")
			},
		})
		m := NewJavaModule(&mockRuntimeManager{})
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when sdk command fails")
		}
	})
}

// --- Install ---

func TestJavaModule_Install(t *testing.T) {
	t.Run("delegates to manager.InstallRuntime with 25-open", func(t *testing.T) {
		var calls []string
		mgr := &mockRuntimeManager{
			installRuntimeFunc: func(_ context.Context, version string, _ io.Writer) error {
				calls = append(calls, "InstallRuntime:"+version)
				return nil
			},
		}
		mod := NewJavaModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		if err := mod.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(calls) != 1 {
			t.Fatalf("expected 1 manager call, got %d: %v", len(calls), calls)
		}
		if calls[0] != "InstallRuntime:25-open" {
			t.Errorf("call[0] = %q, want %q", calls[0], "InstallRuntime:25-open")
		}
	})

	t.Run("manager error propagates", func(t *testing.T) {
		wantErr := errors.New("sdk install failed")
		mgr := &mockRuntimeManager{
			installRuntimeFunc: func(_ context.Context, _ string, _ io.Writer) error {
				return wantErr
			},
		}
		mod := NewJavaModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})
}

// --- AuditInfo ---

func TestJavaModule_AuditInfo(t *testing.T) {
	mgr := &mockRuntimeManager{
		auditInfoFunc: func() string { return "openjdk 25 2025-09-16" },
	}
	m := NewJavaModule(mgr)
	if got := m.AuditInfo(); got != "openjdk 25 2025-09-16" {
		t.Errorf("AuditInfo() = %q, want %q", got, "openjdk 25 2025-09-16")
	}
}
