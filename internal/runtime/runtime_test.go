package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/davichuder/MyDots/internal/installer/runner"
)

// mockExecutor implements runner.Executor for testing.
type mockExecutor struct {
	executeFunc  func(ctx context.Context, name string, args ...string) ([]byte, error)
	lookPathFunc func(name string) (string, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, name, args...)
	}
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func (m *mockExecutor) LookPath(name string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(name)
	}
	return exec.LookPath(name)
}

// defaultExecutor restores the runner to OS-level command execution.
type defaultExecutor struct{}

func (defaultExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func (defaultExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// withMockExecutor sets a mock executor for the test and restores to OS-level on cleanup.
func withMockExecutor(t *testing.T, e runner.Executor) {
	t.Helper()
	runner.SetExecutor(e)
	t.Cleanup(func() { runner.SetExecutor(defaultExecutor{}) })
}

// skipIfWindows skips the test on Windows because runner.Script uses /bin/sh.
func skipIfWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: /bin/sh not available")
	}
}

// --- FnmManager --- //

func TestFnmManager_IsInstalled(t *testing.T) {
	t.Run("fnm found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := FnmManager{}
		if !m.IsInstalled() {
			t.Error("expected true when fnm is on PATH")
		}
	})

	t.Run("fnm not found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := FnmManager{}
		if m.IsInstalled() {
			t.Error("expected false when fnm is not on PATH")
		}
	})
}

func TestFnmManager_Install(t *testing.T) {
	t.Run("install runs brew install fnm", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := FnmManager{}
		err := m.Install(context.Background(), io.Discard, nil)
		if err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("brew fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("brew failed")
			},
		})
		m := FnmManager{}
		err := m.Install(context.Background(), io.Discard, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFnmManager_InstallRuntime(t *testing.T) {
	t.Run("version 24", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := FnmManager{}
		err := m.InstallRuntime(context.Background(), "24", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(24) = %v, want nil", err)
		}
	})

	t.Run("version 22", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := FnmManager{}
		err := m.InstallRuntime(context.Background(), "22", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(22) = %v, want nil", err)
		}
	})

	t.Run("fnm fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("fnm install failed")
			},
		})
		m := FnmManager{}
		err := m.InstallRuntime(context.Background(), "24", io.Discard)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFnmManager_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("fnm 1.37.0\n"), nil
		},
	})
	m := FnmManager{}
	got := m.AuditInfo()
	if got != "fnm 1.37.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "fnm 1.37.0")
	}
}

// --- UvManager --- //

func TestUvManager_IsInstalled(t *testing.T) {
	t.Run("uv found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := UvManager{}
		if !m.IsInstalled() {
			t.Error("expected true when uv is on PATH")
		}
	})

	t.Run("uv not found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := UvManager{}
		if m.IsInstalled() {
			t.Error("expected false when uv is not on PATH")
		}
	})
}

func TestUvManager_Install(t *testing.T) {
	t.Run("install runs brew install uv", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := UvManager{}
		err := m.Install(context.Background(), io.Discard, nil)
		if err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("brew fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("brew failed")
			},
		})
		m := UvManager{}
		err := m.Install(context.Background(), io.Discard, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUvManager_InstallRuntime(t *testing.T) {
	t.Run("version 3.12", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := UvManager{}
		err := m.InstallRuntime(context.Background(), "3.12", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(3.12) = %v, want nil", err)
		}
	})

	t.Run("version 3.11", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := UvManager{}
		err := m.InstallRuntime(context.Background(), "3.11", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(3.11) = %v, want nil", err)
		}
	})

	t.Run("uv fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("uv install failed")
			},
		})
		m := UvManager{}
		err := m.InstallRuntime(context.Background(), "3.12", io.Discard)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUvManager_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("uv 0.6.14\n"), nil
		},
	})
	m := UvManager{}
	got := m.AuditInfo()
	if got != "uv 0.6.14" {
		t.Errorf("AuditInfo() = %q, want %q", got, "uv 0.6.14")
	}
}

// --- SdkmanManager --- //

func TestSdkmanManager_IsInstalled(t *testing.T) {
	t.Run("sdkman dir exists", func(t *testing.T) {
		saved := checkSdkmanDir
		checkSdkmanDir = func() bool { return true }
		t.Cleanup(func() { checkSdkmanDir = saved })

		m := SdkmanManager{}
		if !m.IsInstalled() {
			t.Error("expected true when sdkman dir exists")
		}
	})

	t.Run("sdkman dir not found", func(t *testing.T) {
		saved := checkSdkmanDir
		checkSdkmanDir = func() bool { return false }
		t.Cleanup(func() { checkSdkmanDir = saved })

		m := SdkmanManager{}
		if m.IsInstalled() {
			t.Error("expected false when sdkman dir does not exist")
		}
	})
}

func TestSdkmanManager_Install(t *testing.T) {
	skipIfWindows(t)

	saved := checkSdkmanDir
	checkSdkmanDir = func() bool { return false }
	t.Cleanup(func() { checkSdkmanDir = saved })

	mapFS := fstest.MapFS{
		"assets/scripts/sdkman-install.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\nexit 0"),
			Mode: 0555,
		},
	}

	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return nil, nil
		},
	})

	m := SdkmanManager{}
	err := m.Install(context.Background(), io.Discard, mapFS)
	if err != nil {
		t.Errorf("Install() = %v, want nil", err)
	}
}

func TestSdkmanManager_InstallRuntime(t *testing.T) {
	t.Run("version 25-open", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := SdkmanManager{}
		err := m.InstallRuntime(context.Background(), "25-open", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(25-open) = %v, want nil", err)
		}
	})

	t.Run("version 21-open", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})
		m := SdkmanManager{}
		err := m.InstallRuntime(context.Background(), "21-open", io.Discard)
		if err != nil {
			t.Errorf("InstallRuntime(21-open) = %v, want nil", err)
		}
	})

	t.Run("sdk install fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("sdk install failed")
			},
		})
		m := SdkmanManager{}
		err := m.InstallRuntime(context.Background(), "25-open", io.Discard)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestSdkmanManager_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("sdk 5.0.0\n"), nil
		},
	})
	m := SdkmanManager{}
	got := m.AuditInfo()
	if got != "sdk 5.0.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "sdk 5.0.0")
	}
}
