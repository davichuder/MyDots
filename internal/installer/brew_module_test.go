package installer

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/platform"
)

// mockExecutor implements runner.Executor for testing BrewModule.
// Tests supply custom Execute and LookPath functions to verify
// that BrewModule delegates to the runner correctly.
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
	return "", nil
}

// setMockRunner replaces the runner executor with a mock for the duration of the test.
func setMockRunner(t *testing.T, e runner.Executor) {
	t.Helper()
	runner.SetExecutor(e)
}

// --- Identity tests ---

func TestBrewModule_ID(t *testing.T) {
	m := BrewModule{id: ModZoxide}
	if m.ID() != ModZoxide {
		t.Errorf("ID() = %q, want %q", m.ID(), ModZoxide)
	}
}

func TestBrewModule_Name(t *testing.T) {
	m := BrewModule{name: "zoxide"}
	if m.Name() != "zoxide" {
		t.Errorf("Name() = %q, want %q", m.Name(), "zoxide")
	}
}

func TestBrewModule_Criticality(t *testing.T) {
	m := BrewModule{}
	if m.Criticality() != NonCritical {
		t.Errorf("Criticality() = %v, want NonCritical", m.Criticality())
	}
}

func TestBrewModule_Dependencies(t *testing.T) {
	deps := []ModuleID{ModHomebrew}
	m := BrewModule{deps: deps}
	got := m.Dependencies()
	if len(got) != 1 || got[0] != ModHomebrew {
		t.Errorf("Dependencies() = %v, want [ModHomebrew]", got)
	}
}

// --- IsInstalled ---

func TestBrewModule_IsInstalled(t *testing.T) {
	t.Run("calls CommandExists with checkCommand", func(t *testing.T) {
		var capturedName string
		setMockRunner(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				capturedName = name
				return "/usr/bin/" + name, nil
			},
		})

		m := BrewModule{checkCommand: "zoxide"}
		got := m.IsInstalled(platform.Platform{})

		if capturedName != "zoxide" {
			t.Errorf("CommandExists called with %q, want %q", capturedName, "zoxide")
		}
		if !got {
			t.Error("IsInstalled() = false, want true")
		}
	})

	t.Run("returns false when command not found", func(t *testing.T) {
		setMockRunner(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})

		m := BrewModule{checkCommand: "nonexistent"}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("IsInstalled() = true, want false")
		}
	})
}

// --- Install ---

func TestBrewModule_Install(t *testing.T) {
	t.Run("calls Brew with formula", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		setMockRunner(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return nil, nil
			},
		})

		var buf bytes.Buffer
		ctx := InstallContext{Log: &buf, Cancel: context.Background()}
		m := BrewModule{formula: "zoxide"}

		err := m.Install(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if capturedName != "brew" {
			t.Errorf("brew command = %q, want %q", capturedName, "brew")
		}
		if len(capturedArgs) != 2 || capturedArgs[0] != "install" || capturedArgs[1] != "zoxide" {
			t.Errorf("brew args = %v, want [install zoxide]", capturedArgs)
		}
	})

	t.Run("returns error when brew fails", func(t *testing.T) {
		setMockRunner(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error"), errors.New("exit status 1")
			},
		})

		var buf bytes.Buffer
		ctx := InstallContext{Log: &buf, Cancel: context.Background()}
		m := BrewModule{formula: "zoxide"}

		err := m.Install(ctx)
		if err == nil {
			t.Error("Install() expected error, got nil")
		}
	})
}

// --- AuditInfo ---

func TestBrewModule_AuditInfo(t *testing.T) {
	t.Run("calls CaptureOutput with checkCommand --version", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		setMockRunner(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return []byte("2.0.0\n"), nil
			},
		})

		m := BrewModule{checkCommand: "zoxide"}
		info := m.AuditInfo()

		if capturedName != "zoxide" {
			t.Errorf("CaptureOutput called with %q, want %q", capturedName, "zoxide")
		}
		if len(capturedArgs) != 1 || capturedArgs[0] != "--version" {
			t.Errorf("CaptureOutput args = %v, want [--version]", capturedArgs)
		}
		if info != "2.0.0" {
			t.Errorf("AuditInfo() = %q, want %q", info, "2.0.0")
		}
	})

	t.Run("returns empty string when command fails", func(t *testing.T) {
		setMockRunner(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("not found")
			},
		})

		m := BrewModule{checkCommand: "nonexistent"}
		if info := m.AuditInfo(); info != "" {
			t.Errorf("AuditInfo() = %q, want empty string", info)
		}
	})
}

// --- Triangulation ---

func TestBrewModule_Triangulation(t *testing.T) {
	modules := []struct {
		mod    BrewModule
		wantID ModuleID
	}{
		{
			mod:    BrewModule{id: ModZoxide, name: "zoxide", formula: "zoxide", checkCommand: "zoxide", deps: []ModuleID{ModHomebrew}},
			wantID: ModZoxide,
		},
		{
			mod:    BrewModule{id: ModAtuin, name: "atuin", formula: "atuin", checkCommand: "atuin", deps: []ModuleID{ModHomebrew}},
			wantID: ModAtuin,
		},
		{
			mod:    BrewModule{id: ModBat, name: "bat", formula: "bat", checkCommand: "bat", deps: []ModuleID{ModHomebrew}},
			wantID: ModBat,
		},
	}

	for _, tc := range modules {
		t.Run(tc.mod.name, func(t *testing.T) {
			// Identity
			if tc.mod.ID() != tc.wantID {
				t.Errorf("ID() = %q, want %q", tc.mod.ID(), tc.wantID)
			}
			if tc.mod.Name() != tc.mod.name {
				t.Errorf("Name() = %q, want %q", tc.mod.Name(), tc.mod.name)
			}

			// Criticality — always NonCritical
			if tc.mod.Criticality() != NonCritical {
				t.Errorf("Criticality() = %v, want NonCritical", tc.mod.Criticality())
			}

			// Dependencies
			deps := tc.mod.Dependencies()
			if len(deps) != 1 || deps[0] != ModHomebrew {
				t.Errorf("Dependencies() = %v, want [ModHomebrew]", deps)
			}

			// IsInstalled delegates to CommandExists
			setMockRunner(t, &mockExecutor{
				lookPathFunc: func(name string) (string, error) {
					if name == tc.mod.checkCommand {
						return "/usr/bin/" + name, nil
					}
					return "", errors.New("not found")
				},
			})
			if !tc.mod.IsInstalled(platform.Platform{}) {
				t.Errorf("IsInstalled(%q) = false, want true", tc.mod.checkCommand)
			}

			// Install delegates to Brew
			setMockRunner(t, &mockExecutor{
				executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
					return nil, nil
				},
			})
			var buf bytes.Buffer
			ctx := InstallContext{Log: &buf, Cancel: context.Background()}
			if err := tc.mod.Install(ctx); err != nil {
				t.Errorf("Install(%q) unexpected error: %v", tc.mod.formula, err)
			}

			// AuditInfo delegates to CaptureOutput
			setMockRunner(t, &mockExecutor{
				executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
					return []byte("1.0.0\n"), nil
				},
			})
			if info := tc.mod.AuditInfo(); info == "" {
				t.Error("AuditInfo() = empty string, want non-empty")
			}
		})
	}
}
