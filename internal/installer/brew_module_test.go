package installer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/platform"
)

// mockExecutor implements runner.Executor for testing BrewModule.
type mockExecutor struct {
	executeFunc  func(ctx context.Context, name string, args ...string) ([]byte, error)
	lookPathFunc func(name string) (string, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, name, args...)
	}
	return nil, nil
}

func (m *mockExecutor) LookPath(name string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(name)
	}
	return "", errors.New("not found")
}

// setExecutor sets a mock executor for the test and restores a neutral mock on cleanup.
func setExecutor(t *testing.T, mock *mockExecutor) {
	t.Helper()
	runner.SetExecutor(mock)
	t.Cleanup(func() {
		// Reset to neutral mock to avoid leaking state between tests.
		runner.SetExecutor(&mockExecutor{})
	})
}

// --- Triangulation ---

func TestBrewModule_triangulation(t *testing.T) {
	modules := []struct {
		name string
		mod  BrewModule
	}{
		{"zoxide", BrewModule{id: "M-XX", name: "zoxide", formula: "zoxide", checkCommand: "zoxide", deps: []ModuleID{"M-01"}}},
		{"atuin", BrewModule{id: "M-XX", name: "atuin", formula: "atuin", checkCommand: "atuin", deps: []ModuleID{"M-01"}}},
		{"bat", BrewModule{id: "M-XX", name: "bat", formula: "bat", checkCommand: "bat", deps: []ModuleID{"M-01"}}},
	}

	for _, m := range modules {
		t.Run(m.name+"_Dependencies", func(t *testing.T) {
			want := []ModuleID{"M-01"}
			got := m.mod.Dependencies()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Dependencies() = %v, want %v", got, want)
			}
		})

		t.Run(m.name+"_Criticality", func(t *testing.T) {
			if m.mod.Criticality() != NonCritical {
				t.Error("expected NonCritical")
			}
		})

		t.Run(m.name+"_ID", func(t *testing.T) {
			if m.mod.ID() == "" {
				t.Error("ID() must not be empty")
			}
		})

		t.Run(m.name+"_Name", func(t *testing.T) {
			if m.mod.Name() == "" {
				t.Error("Name() must not be empty")
			}
		})
	}
}

// --- IsInstalled ---

func TestBrewModule_IsInstalled(t *testing.T) {
	t.Run("command exists returns true", func(t *testing.T) {
		var capturedCmd string
		setExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				capturedCmd = name
				return "/usr/bin/" + name, nil
			},
		})
		m := BrewModule{checkCommand: "zoxide"}

		result := m.IsInstalled(platform.Platform{})

		if !result {
			t.Error("expected true when binary exists")
		}
		if capturedCmd != "zoxide" {
			t.Errorf("expected checkCommand 'zoxide', got: %q", capturedCmd)
		}
	})

	t.Run("command missing returns false", func(t *testing.T) {
		setExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := BrewModule{checkCommand: "nonexistent"}

		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when binary missing")
		}
	})
}

// --- Install ---

func TestBrewModule_Install(t *testing.T) {
	t.Run("calls brew install with formula", func(t *testing.T) {
		var buf bytes.Buffer
		var capturedName string
		var capturedArgs []string
		setExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return nil, nil
			},
		})

		m := BrewModule{formula: "zoxide"}
		ictx := InstallContext{
			Log:    &buf,
			Cancel: context.Background(),
		}

		err := m.Install(ictx)
		if err != nil {
			t.Fatal(err)
		}
		if capturedName != "brew" {
			t.Errorf("expected command 'brew', got: %q", capturedName)
		}
		if len(capturedArgs) != 2 || capturedArgs[0] != "install" || capturedArgs[1] != "zoxide" {
			t.Errorf("expected args ['install', 'zoxide'], got: %v", capturedArgs)
		}
	})

	t.Run("returns error when brew fails", func(t *testing.T) {
		setExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error"), errors.New("exit status 1")
			},
		})

		m := BrewModule{formula: "zoxide"}
		ictx := InstallContext{
			Log:    io.Discard,
			Cancel: context.Background(),
		}

		err := m.Install(ictx)
		if err == nil {
			t.Error("expected error when brew fails")
		}
	})
}

// --- AuditInfo ---

func TestBrewModule_AuditInfo(t *testing.T) {
	t.Run("returns version output", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		setExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return []byte("1.2.3\n"), nil
			},
		})

		m := BrewModule{checkCommand: "zoxide"}

		result := m.AuditInfo()
		if result != "1.2.3" {
			t.Errorf("expected '1.2.3', got: %q", result)
		}
		if capturedName != "zoxide" {
			t.Errorf("expected command 'zoxide', got: %q", capturedName)
		}
		if len(capturedArgs) != 1 || capturedArgs[0] != "--version" {
			t.Errorf("expected args ['--version'], got: %v", capturedArgs)
		}
	})

	t.Run("command failure returns empty string", func(t *testing.T) {
		setExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("not found")
			},
		})

		m := BrewModule{checkCommand: "nonexistent"}

		if m.AuditInfo() != "" {
			t.Error("expected empty string on failure")
		}
	})
}
