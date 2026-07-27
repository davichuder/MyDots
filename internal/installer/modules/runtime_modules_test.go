package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

// mockRuntimeManager implements runtime.RuntimeManager for testing.
type mockRuntimeManager struct {
	installFunc        func(ctx context.Context, logw io.Writer, assets fs.FS) error
	isInstalledFunc    func() bool
	installRuntimeFunc func(ctx context.Context, version string, logw io.Writer) error
	auditInfoFunc      func() string
}

func (m *mockRuntimeManager) Install(ctx context.Context, logw io.Writer, assets fs.FS) error {
	if m.installFunc != nil {
		return m.installFunc(ctx, logw, assets)
	}
	return nil
}

func (m *mockRuntimeManager) IsInstalled() bool {
	if m.isInstalledFunc != nil {
		return m.isInstalledFunc()
	}
	return false
}

func (m *mockRuntimeManager) InstallRuntime(ctx context.Context, version string, logw io.Writer) error {
	if m.installRuntimeFunc != nil {
		return m.installRuntimeFunc(ctx, version, logw)
	}
	return nil
}

func (m *mockRuntimeManager) AuditInfo() string {
	if m.auditInfoFunc != nil {
		return m.auditInfoFunc()
	}
	return ""
}

// compile-time check: mockRuntimeManager implements runtime.RuntimeManager
var _ runtime.RuntimeManager = (*mockRuntimeManager)(nil)

// ---------------------------------------------------------------------------
// FnmNodeModule tests
// ---------------------------------------------------------------------------

func TestFnmNodeModule_ID(t *testing.T) {
	m := NewFnmNodeModule(&mockRuntimeManager{})
	if got := m.ID(); got != types.ModFnm {
		t.Errorf("ID() = %v, want %v", got, types.ModFnm)
	}
}

func TestFnmNodeModule_Name(t *testing.T) {
	m := NewFnmNodeModule(&mockRuntimeManager{})
	if got := m.Name(); got != "Fnm" {
		t.Errorf("Name() = %q, want %q", got, "Fnm")
	}
}

func TestFnmNodeModule_Criticality(t *testing.T) {
	m := NewFnmNodeModule(&mockRuntimeManager{})
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestFnmNodeModule_Dependencies(t *testing.T) {
	m := NewFnmNodeModule(&mockRuntimeManager{})
	want := []types.ModuleID{types.ModHomebrew}
	got := m.Dependencies()
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("Dependencies() = %v, want %v", got, want)
	}
}

func TestFnmNodeModule_IsInstalled(t *testing.T) {
	t.Run("delegates to manager — installed", func(t *testing.T) {
		m := NewFnmNodeModule(&mockRuntimeManager{
			isInstalledFunc: func() bool { return true },
		})
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when manager.IsInstalled() returns true")
		}
	})

	t.Run("delegates to manager — not installed", func(t *testing.T) {
		m := NewFnmNodeModule(&mockRuntimeManager{
			isInstalledFunc: func() bool { return false },
		})
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when manager.IsInstalled() returns false")
		}
	})
}

func TestFnmNodeModule_Install(t *testing.T) {
	t.Run("calls Install then InstallRuntime with version 24", func(t *testing.T) {
		var calls []string
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				calls = append(calls, "Install")
				return nil
			},
			installRuntimeFunc: func(_ context.Context, version string, _ io.Writer) error {
				calls = append(calls, "InstallRuntime:"+version)
				return nil
			},
		}
		mod := NewFnmNodeModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(calls) != 2 {
			t.Fatalf("expected 2 manager calls, got %d: %v", len(calls), calls)
		}
		if calls[0] != "Install" {
			t.Errorf("call[0] = %q, want %q", calls[0], "Install")
		}
		if calls[1] != "InstallRuntime:24" {
			t.Errorf("call[1] = %q, want %q", calls[1], "InstallRuntime:24")
		}
	})

	t.Run("manager Install error propagates", func(t *testing.T) {
		wantErr := errors.New("brew install failed")
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				return wantErr
			},
		}
		mod := NewFnmNodeModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})

	t.Run("manager InstallRuntime error propagates", func(t *testing.T) {
		wantErr := errors.New("fnm install failed")
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				return nil
			},
			installRuntimeFunc: func(_ context.Context, _ string, _ io.Writer) error {
				return wantErr
			},
		}
		mod := NewFnmNodeModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})
}

func TestFnmNodeModule_AuditInfo(t *testing.T) {
	mgr := &mockRuntimeManager{
		auditInfoFunc: func() string { return "fnm 1.37.0" },
	}
	m := NewFnmNodeModule(mgr)
	if got := m.AuditInfo(); got != "fnm 1.37.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "fnm 1.37.0")
	}
}

// ---------------------------------------------------------------------------
// UvPythonModule tests
// ---------------------------------------------------------------------------

func TestUvPythonModule_ID(t *testing.T) {
	m := NewUvPythonModule(&mockRuntimeManager{})
	if got := m.ID(); got != types.ModUv {
		t.Errorf("ID() = %v, want %v", got, types.ModUv)
	}
}

func TestUvPythonModule_Name(t *testing.T) {
	m := NewUvPythonModule(&mockRuntimeManager{})
	if got := m.Name(); got != "Uv" {
		t.Errorf("Name() = %q, want %q", got, "Uv")
	}
}

func TestUvPythonModule_Criticality(t *testing.T) {
	m := NewUvPythonModule(&mockRuntimeManager{})
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestUvPythonModule_Dependencies(t *testing.T) {
	m := NewUvPythonModule(&mockRuntimeManager{})
	want := []types.ModuleID{types.ModHomebrew}
	got := m.Dependencies()
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("Dependencies() = %v, want %v", got, want)
	}
}

func TestUvPythonModule_IsInstalled(t *testing.T) {
	t.Run("delegates to manager — installed", func(t *testing.T) {
		m := NewUvPythonModule(&mockRuntimeManager{
			isInstalledFunc: func() bool { return true },
		})
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when manager.IsInstalled() returns true")
		}
	})

	t.Run("delegates to manager — not installed", func(t *testing.T) {
		m := NewUvPythonModule(&mockRuntimeManager{
			isInstalledFunc: func() bool { return false },
		})
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when manager.IsInstalled() returns false")
		}
	})
}

func TestUvPythonModule_Install(t *testing.T) {
	t.Run("calls Install then InstallRuntime with version 3.12", func(t *testing.T) {
		var calls []string
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				calls = append(calls, "Install")
				return nil
			},
			installRuntimeFunc: func(_ context.Context, version string, _ io.Writer) error {
				calls = append(calls, "InstallRuntime:"+version)
				return nil
			},
		}
		mod := NewUvPythonModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if len(calls) != 2 {
			t.Fatalf("expected 2 manager calls, got %d: %v", len(calls), calls)
		}
		if calls[0] != "Install" {
			t.Errorf("call[0] = %q, want %q", calls[0], "Install")
		}
		if calls[1] != "InstallRuntime:3.12" {
			t.Errorf("call[1] = %q, want %q", calls[1], "InstallRuntime:3.12")
		}
	})

	t.Run("manager Install error propagates", func(t *testing.T) {
		wantErr := errors.New("brew install failed")
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				return wantErr
			},
		}
		mod := NewUvPythonModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})

	t.Run("manager InstallRuntime error propagates", func(t *testing.T) {
		wantErr := errors.New("uv python install failed")
		mgr := &mockRuntimeManager{
			installFunc: func(_ context.Context, _ io.Writer, _ fs.FS) error {
				return nil
			},
			installRuntimeFunc: func(_ context.Context, _ string, _ io.Writer) error {
				return wantErr
			},
		}
		mod := NewUvPythonModule(mgr)
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		if err := mod.Install(ctx); !errors.Is(err, wantErr) {
			t.Errorf("Install() = %v, want %v", err, wantErr)
		}
	})
}

func TestUvPythonModule_AuditInfo(t *testing.T) {
	mgr := &mockRuntimeManager{
		auditInfoFunc: func() string { return "uv 0.4.0" },
	}
	m := NewUvPythonModule(mgr)
	if got := m.AuditInfo(); got != "uv 0.4.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "uv 0.4.0")
	}
}
