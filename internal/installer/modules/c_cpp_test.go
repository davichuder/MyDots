package modules

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/platform"
)

// --- ID ---

func TestCppToolchain_ID(t *testing.T) {
	m := CppToolchainModule{}
	if got := m.ID(); got != types.ModCppToolchain {
		t.Errorf("ID() = %v, want %v", got, types.ModCppToolchain)
	}
}

// --- Name ---

func TestCppToolchain_Name(t *testing.T) {
	m := CppToolchainModule{}
	if got := m.Name(); got != "C++ Toolchain" {
		t.Errorf("Name() = %q, want %q", got, "C++ Toolchain")
	}
}

// --- Criticality ---

func TestCppToolchain_Criticality(t *testing.T) {
	m := CppToolchainModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

// --- Dependencies ---

func TestCppToolchain_Dependencies(t *testing.T) {
	m := CppToolchainModule{}
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

func TestCppToolchain_IsInstalled(t *testing.T) {
	t.Run("darwin checks clangd", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "clangd" {
					return "/usr/bin/clangd", nil
				}
				return "", errors.New("not found: " + name)
			},
		})
		m := CppToolchainModule{}
		p := platform.Platform{OS: platform.Darwin}
		if !m.IsInstalled(p) {
			t.Error("expected true when clangd is on PATH")
		}
	})

	t.Run("darwin clangd not found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := CppToolchainModule{}
		p := platform.Platform{OS: platform.Darwin}
		if m.IsInstalled(p) {
			t.Error("expected false when clangd is not on PATH")
		}
	})

	t.Run("ubuntu checks gcc", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "gcc" {
					return "/usr/bin/gcc", nil
				}
				return "", errors.New("not found: " + name)
			},
		})
		m := CppToolchainModule{}
		p := platform.Platform{OS: platform.Linux, Variant: platform.Native}
		if !m.IsInstalled(p) {
			t.Error("expected true when gcc is on PATH")
		}
	})

	t.Run("ubuntu gcc not found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := CppToolchainModule{}
		p := platform.Platform{OS: platform.Linux, Variant: platform.Native}
		if m.IsInstalled(p) {
			t.Error("expected false when gcc is not on PATH")
		}
	})

	t.Run("wsl2 also checks gcc", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				if name == "gcc" {
					return "/usr/bin/gcc", nil
				}
				return "", errors.New("not found: " + name)
			},
		})
		m := CppToolchainModule{}
		p := platform.Platform{OS: platform.Linux, Variant: platform.WSL2}
		if !m.IsInstalled(p) {
			t.Error("expected true when gcc is on PATH (WSL2)")
		}
	})
}

// --- Install ---

func TestCppToolchain_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("darwin installs via brew", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" && len(args) == 4 && args[0] == "install" && args[1] == "gcc" && args[2] == "cmake" && args[3] == "llvm" {
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin},
		}
		m := CppToolchainModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("darwin brew fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error: brew failed\n"), errors.New("exit status 1")
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin},
		}
		m := CppToolchainModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("ubuntu installs via apt", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "sudo" && len(args) == 6 && args[0] == "apt" && args[1] == "install" && args[2] == "-y" && args[3] == "build-essential" && args[4] == "gdb" && args[5] == "cmake" {
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := CppToolchainModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("ubuntu apt fails returns AptInstallError", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("E: unable to fetch packages\n"), errors.New("exit status 100")
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := CppToolchainModule{}
		err := m.Install(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var aptErr myerr.AptInstallError
		if !errors.As(err, &aptErr) {
			t.Fatalf("expected AptInstallError, got: %T", err)
		}
		if aptErr.ExitCode != 100 {
			t.Errorf("ExitCode = %d, want %d", aptErr.ExitCode, 100)
		}
		if !strings.Contains(aptErr.Command, "apt install") {
			t.Errorf("Command = %q, should contain 'apt install'", aptErr.Command)
		}
		if aptErr.What() == "" {
			t.Error("What() should not be empty")
		}
		if aptErr.Why() == "" {
			t.Error("Why() should not be empty")
		}
		if aptErr.Fix() == "" {
			t.Error("Fix() should not be empty")
		}
	})

	t.Run("wsl2 also uses apt (same as native linux)", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "sudo" && len(args) == 6 && args[0] == "apt" && args[1] == "install" && args[2] == "-y" && args[3] == "build-essential" && args[4] == "gdb" && args[5] == "cmake" {
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
		}
		m := CppToolchainModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})
}

// --- AuditInfo ---

func TestCppToolchain_AuditInfo(t *testing.T) {
	t.Run("darwin returns clangd version via brew prefix", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" && len(args) == 2 && args[0] == "--prefix" && args[1] == "llvm" {
					return []byte("/opt/homebrew/opt/llvm\n"), nil
				}
				if name == "/opt/homebrew/opt/llvm/bin/clangd" && len(args) == 1 && args[0] == "--version" {
					return []byte("clangd version 18.1.8\n"), nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})
		m := CppToolchainModule{}
		got := m.AuditInfo()
		want := "clangd version 18.1.8"
		if got != want {
			t.Errorf("AuditInfo() = %q, want %q", got, want)
		}
	})

	t.Run("darwin brew prefix fails returns empty", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("brew --prefix failed")
			},
		})
		m := CppToolchainModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})

	t.Run("ubuntu returns gcc version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "gcc" && len(args) == 1 && args[0] == "--version" {
					return []byte("gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0\n"), nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})
		m := CppToolchainModule{}
		got := m.AuditInfo()
		want := "gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0"
		if got != want {
			t.Errorf("AuditInfo() = %q, want %q", got, want)
		}
	})

	t.Run("ubuntu gcc fails returns empty", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := CppToolchainModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
