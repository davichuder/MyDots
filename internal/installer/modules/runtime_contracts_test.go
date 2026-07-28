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

func TestRuntimeContractMetadata(t *testing.T) {
	for _, tt := range []struct {
		name string
		mod  types.Module
		id   types.ModuleID
		want string
		dep  types.ModuleID
	}{
		{"fnm", NewFnmModule(&mockRuntimeManager{}), types.ModFnm, "Fnm", types.ModHomebrew},
		{"node", NewNodeModule(&mockRuntimeManager{}), types.ModNode, "Node 24", types.ModFnm},
		{"uv", NewUvModule(&mockRuntimeManager{}), types.ModUv, "Uv", types.ModHomebrew},
		{"python", NewPythonModule(&mockRuntimeManager{}), types.ModPython, "Python 3.12", types.ModUv},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mod.ID() != tt.id || tt.mod.Name() != tt.want || tt.mod.Dependencies()[0] != tt.dep {
				t.Fatalf("metadata = %s, %q, %v", tt.mod.ID(), tt.mod.Name(), tt.mod.Dependencies())
			}
		})
	}
}

func TestRuntimeContractsRespectPartialInstalledStates(t *testing.T) {
	for _, tt := range []struct {
		name            string
		mod             types.Module
		command, output string
		want            bool
	}{
		{"node missing despite fnm", NewNodeModule(&mockRuntimeManager{}), "fnm", "v22.0.0", false},
		{"python missing despite uv", NewPythonModule(&mockRuntimeManager{}), "uv", "cpython-3.11.9", false},
		{"fnm present", NewFnmModule(&mockRuntimeManager{isInstalledFunc: func() bool { return true }}), "", "", true},
		{"uv present", NewUvModule(&mockRuntimeManager{isInstalledFunc: func() bool { return true }}), "", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			withMockExecutor(t, &mockExecutor{executeFunc: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				if name != tt.command {
					t.Fatalf("command = %q, want %q", name, tt.command)
				}
				return []byte(tt.output), nil
			}})
			if got := tt.mod.IsInstalled(platform.Platform{}); got != tt.want {
				t.Fatalf("IsInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuntimeContractInstallDelegation(t *testing.T) {
	for _, tt := range []struct {
		name string
		make func(*mockRuntimeManager) types.Module
		want []string
	}{
		{"fnm", func(m *mockRuntimeManager) types.Module { return NewFnmModule(m) }, []string{"install"}},
		{"uv", func(m *mockRuntimeManager) types.Module { return NewUvModule(m) }, []string{"install"}},
		{"python", func(m *mockRuntimeManager) types.Module { return NewPythonModule(m) }, []string{"runtime:3.12"}},
		{"node", func(m *mockRuntimeManager) types.Module { return NewNodeModule(m) }, []string{"fnm install 24", "fnm default 24"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			m := &mockRuntimeManager{
				installFunc: func(context.Context, io.Writer, fs.FS) error { calls = append(calls, "install"); return nil },
				installRuntimeFunc: func(_ context.Context, v string, _ io.Writer) error {
					calls = append(calls, "runtime:"+v)
					if v == "24" {
						calls[len(calls)-1] = "fnm install 24"
					}
					return nil
				},
				setDefaultFunc: func(_ context.Context, v string, _ io.Writer) error {
					calls = append(calls, "fnm default "+v)
					return nil
				},
			}
			if err := tt.make(m).Install(types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}}); err != nil {
				t.Fatal(err)
			}
			if len(calls) != len(tt.want) {
				t.Fatalf("calls = %v, want %v", calls, tt.want)
			}
			for i := range calls {
				if calls[i] != tt.want[i] {
					t.Fatalf("calls = %v, want %v", calls, tt.want)
				}
			}
		})
	}

	t.Run("Node installation failure is propagated without setting a default", func(t *testing.T) {
		installErr, defaultCalled := errors.New("fnm install failed"), false
		m := &mockRuntimeManager{installRuntimeFunc: func(context.Context, string, io.Writer) error { return installErr }, setDefaultFunc: func(context.Context, string, io.Writer) error { defaultCalled = true; return nil }}
		if err := NewNodeModule(m).Install(types.InstallContext{Cancel: context.Background(), Log: io.Discard}); !errors.Is(err, installErr) {
			t.Fatalf("Install() = %v, want %v", err, installErr)
		}
		if defaultCalled {
			t.Fatal("SetDefaultRuntime() ran after InstallRuntime() failed")
		}
	})
}

func TestRuntimeVersionChecksRequireExactSemanticVersions(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mod    types.Module
		output string
		want   bool
	}{
		{"Node 24", NewNodeModule(&mockRuntimeManager{}), "* v24.0.0 default", true},
		{"Node 24 prerelease", NewNodeModule(&mockRuntimeManager{}), "v24.1.0-rc.1", true},
		{"Node rejects 240", NewNodeModule(&mockRuntimeManager{}), "v240.0.0", false},
		{"Node rejects malformed", NewNodeModule(&mockRuntimeManager{}), "v24.0", false},
		{"Node rejects unrelated text", NewNodeModule(&mockRuntimeManager{}), "available: v24.0.0", false},
		{"Python 3.12", NewPythonModule(&mockRuntimeManager{}), "cpython-3.12.8-windows-x86_64-none", true},
		{"Python rejects 13.12", NewPythonModule(&mockRuntimeManager{}), "cpython-13.12.0-linux-x86_64", false},
		{"Python rejects 3.120", NewPythonModule(&mockRuntimeManager{}), "cpython-3.120.0-linux-x86_64", false},
		{"Python rejects unrelated substring", NewPythonModule(&mockRuntimeManager{}), "available version: 3.12.0", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			withMockExecutor(t, &mockExecutor{executeFunc: func(context.Context, string, ...string) ([]byte, error) { return []byte(tt.output), nil }})
			if got := tt.mod.IsInstalled(platform.Platform{}); got != tt.want {
				t.Fatalf("IsInstalled() = %v, want %v for %q", got, tt.want, tt.output)
			}
		})
	}
}
