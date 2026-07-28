package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"runtime"
	"strings"
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var gotName string
		var gotArgs []string
		var gotCtx context.Context
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(commandCtx context.Context, name string, args ...string) ([]byte, error) {
				gotCtx = commandCtx
				gotName = name
				gotArgs = append([]string(nil), args...)
				return []byte("Installing: java 25-open\n"), nil
			},
		})
		m := SdkmanManager{}
		var log bytes.Buffer
		err := m.InstallRuntime(ctx, "25-open", &log)
		if err != nil {
			t.Errorf("InstallRuntime(25-open) = %v, want nil", err)
		}
		if gotCtx != ctx {
			t.Error("InstallRuntime() did not pass its context to the command runner")
		}
		if gotName != "sh" {
			t.Errorf("command = %q, want sh", gotName)
		}
		if len(gotArgs) != 4 || gotArgs[0] != "-c" || gotArgs[2] != "sh" || gotArgs[3] != "25-open" {
			t.Fatalf("command args = %#v, want sh -c <script> sh 25-open", gotArgs)
		}
		if !strings.Contains(gotArgs[1], `. "$HOME/.sdkman/bin/sdkman-init.sh"`) {
			t.Errorf("shell command = %q, want SDKMAN init script to be sourced", gotArgs[1])
		}
		if !strings.Contains(gotArgs[1], `sdk install java "$1"`) {
			t.Errorf("shell command = %q, want sdk install command", gotArgs[1])
		}
		if got, want := log.String(), "Installing: java 25-open\n"; got != want {
			t.Errorf("log = %q, want %q", got, want)
		}
	})

	t.Run("sdk install fails returns error", func(t *testing.T) {
		wantErr := errors.New("sdk install failed")
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name != "sh" {
					t.Errorf("command = %q, want sh", name)
				}
				return []byte("sdk install failed\n"), wantErr
			},
		})
		m := SdkmanManager{}
		var log bytes.Buffer
		err := m.InstallRuntime(context.Background(), "25-open", &log)
		if !errors.Is(err, wantErr) {
			t.Errorf("InstallRuntime() = %v, want %v", err, wantErr)
		}
		if got, want := log.String(), "sdk install failed\n"; got != want {
			t.Errorf("log = %q, want %q", got, want)
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
