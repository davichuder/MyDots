package modules

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// --- ID ---

func TestZsh_ID(t *testing.T) {
	m := ZshModule{}
	if got := m.ID(); got != types.ModZsh {
		t.Errorf("ID() = %v, want %v", got, types.ModZsh)
	}
}

// --- Name ---

func TestZsh_Name(t *testing.T) {
	m := ZshModule{}
	if got := m.Name(); got != "Zsh" {
		t.Errorf("Name() = %q, want %q", got, "Zsh")
	}
}

// --- Criticality ---

func TestZsh_Criticality(t *testing.T) {
	m := ZshModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

// --- Dependencies ---

func TestZsh_Dependencies(t *testing.T) {
	m := ZshModule{}
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

func TestZsh_IsInstalled(t *testing.T) {
	t.Run("zsh found on PATH", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})
		m := ZshModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when zsh is on PATH")
		}
	})

	t.Run("zsh not found", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := ZshModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when zsh is not on PATH")
		}
	})
}

// --- Install ---

func TestZsh_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("darwin installs zsh and runs chsh", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				switch {
				case name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "zsh":
					return nil, nil
				case name == "brew" && len(args) == 1 && args[0] == "--prefix":
					return []byte("/opt/homebrew"), nil
				case name == "chsh":
					return nil, nil
				default:
					return nil, errors.New("unexpected command: " + name)
				}
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin},
		}
		m := ZshModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
	})

	t.Run("ubuntu appends to /etc/shells and runs chsh", func(t *testing.T) {
		var sudoArgs []string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				switch {
				case name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "zsh":
					return nil, nil
				case name == "brew" && len(args) == 1 && args[0] == "--prefix":
					return []byte("/home/linuxbrew/.linuxbrew"), nil
				case name == "sudo":
					sudoArgs = append([]string(nil), args...)
					return nil, nil
				case name == "chsh":
					return nil, nil
				default:
					return nil, errors.New("unexpected command: " + name)
				}
			},
		})

		origRead := readShellsFile
		t.Cleanup(func() {
			readShellsFile = origRead
		})

		readShellsFile = func(path string) ([]byte, error) {
			return []byte("/bin/zsh\n"), nil
		}

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := ZshModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
		if !reflect.DeepEqual(sudoArgs, []string{"sh", "-c", `printf '%s\n' "$1" >> "$2"`, "sh", "/home/linuxbrew/.linuxbrew/bin/zsh", "/etc/shells"}) {
			t.Errorf("sudo args = %#v, want safe elevated shell registration", sudoArgs)
		}
	})

	t.Run("ubuntu idempotent — brew path already in /etc/shells", func(t *testing.T) {
		var sudoCalled bool
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				switch {
				case name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "zsh":
					return nil, nil
				case name == "brew" && len(args) == 1 && args[0] == "--prefix":
					return []byte("/home/linuxbrew/.linuxbrew"), nil
				case name == "sudo":
					sudoCalled = true
					return nil, nil
				case name == "chsh":
					return nil, nil
				default:
					return nil, errors.New("unexpected command: " + name)
				}
			},
		})

		origRead := readShellsFile
		t.Cleanup(func() {
			readShellsFile = origRead
		})

		readShellsFile = func(path string) ([]byte, error) {
			return []byte("/bin/zsh\n/home/linuxbrew/.linuxbrew/bin/zsh\n"), nil
		}

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := ZshModule{}
		if err := m.Install(ctx); err != nil {
			t.Errorf("Install() = %v, want nil", err)
		}
		if sudoCalled {
			t.Error("shell registration should not be called when brew path already in /etc/shells")
		}
	})

	t.Run("ubuntu chsh fails returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				switch {
				case name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "zsh":
					return nil, nil
				case name == "brew" && len(args) == 1 && args[0] == "--prefix":
					return []byte("/home/linuxbrew/.linuxbrew"), nil
				case name == "sudo":
					return nil, nil
				case name == "chsh":
					return nil, errors.New("permission denied")
				default:
					return nil, errors.New("unexpected command: " + name)
				}
			},
		})

		origRead := readShellsFile
		t.Cleanup(func() {
			readShellsFile = origRead
		})

		readShellsFile = func(path string) ([]byte, error) {
			return []byte("/bin/zsh\n"), nil
		}

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
		}
		m := ZshModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// --- AuditInfo ---

func TestZsh_AuditInfo(t *testing.T) {
	t.Run("success returns zsh version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("zsh 5.9 (x86_64-apple-darwin23.0)\n"), nil
			},
		})
		m := ZshModule{}
		got := m.AuditInfo()
		want := "zsh 5.9 (x86_64-apple-darwin23.0)"
		if got != want {
			t.Errorf("AuditInfo() = %q, want %q", got, want)
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := ZshModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
