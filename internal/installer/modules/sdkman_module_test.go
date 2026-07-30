package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// SdkmanModule tests — T-071 (RED)
// ---------------------------------------------------------------------------

func TestSdkmanModule_ID(t *testing.T) {
	m := NewSdkmanModule(&mockRuntimeManager{})
	if got := m.ID(); got != types.ModSdkman {
		t.Errorf("ID() = %v, want %v", got, types.ModSdkman)
	}
}

func TestSdkmanModule_Name(t *testing.T) {
	m := NewSdkmanModule(&mockRuntimeManager{})
	if got := m.Name(); got != "Sdkman" {
		t.Errorf("Name() = %q, want %q", got, "Sdkman")
	}
}

func TestSdkmanModule_Criticality(t *testing.T) {
	m := NewSdkmanModule(&mockRuntimeManager{})
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestSdkmanModule_Dependencies(t *testing.T) {
	m := NewSdkmanModule(&mockRuntimeManager{})
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

func TestSdkmanModule_IsInstalled(t *testing.T) {
	t.Run("init file exists returns true", func(t *testing.T) {
		orig := sdkmanCheckInitFile
		sdkmanCheckInitFile = func() bool { return true }
		t.Cleanup(func() { sdkmanCheckInitFile = orig })

		m := NewSdkmanModule(&mockRuntimeManager{})
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when init file exists")
		}
	})

	t.Run("init file missing returns false", func(t *testing.T) {
		orig := sdkmanCheckInitFile
		sdkmanCheckInitFile = func() bool { return false }
		t.Cleanup(func() { sdkmanCheckInitFile = orig })

		m := NewSdkmanModule(&mockRuntimeManager{})
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when init file missing")
		}
	})
}

// --- Install ---

func TestSdkmanModule_Install(t *testing.T) {
	t.Run("delegates to manager.Install", func(t *testing.T) {
		var installCalled bool
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				installCalled = true
				return nil
			},
		}
		mod := NewSdkmanModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: nil,
		}
		if err := mod.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if !installCalled {
			t.Error("expected manager.Install to be called")
		}
	})

	t.Run("manager error propagates", func(t *testing.T) {
		wantErr := errors.New("script failed")
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				return wantErr
			},
		}
		mod := NewSdkmanModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: nil,
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})
}

// --- AuditInfo ---

func TestSdkmanModule_AuditInfo(t *testing.T) {
	mgr := &mockRuntimeManager{
		auditInfoFunc: func() string { return "sdk 5.0.0" },
	}
	m := NewSdkmanModule(mgr)
	if got := m.AuditInfo(); got != "sdk 5.0.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "sdk 5.0.0")
	}
}
