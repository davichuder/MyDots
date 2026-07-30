package modules

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// DockerModule tests — T-057 (RED)
// ---------------------------------------------------------------------------

func TestDocker_ID(t *testing.T) {
	m := DockerModule{}
	if got := m.ID(); got != types.ModDocker {
		t.Errorf("ID() = %v, want %v", got, types.ModDocker)
	}
}

func TestDocker_Name(t *testing.T) {
	m := DockerModule{}
	if got := m.Name(); got != "Docker" {
		t.Errorf("Name() = %q, want %q", got, "Docker")
	}
}

func TestDocker_Criticality(t *testing.T) {
	m := DockerModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestDocker_Dependencies(t *testing.T) {
	m := DockerModule{}
	deps := m.Dependencies()
	if len(deps) != 1 || deps[0] != types.ModHomebrew {
		t.Errorf("Dependencies() = %v, want [%v]", deps, types.ModHomebrew)
	}
}

// --- IsInstalled ---

func TestDocker_IsInstalled(t *testing.T) {
	t.Run("docker on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/docker", nil
			},
		})
		m := DockerModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when docker is on PATH")
		}
	})

	t.Run("docker not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := DockerModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when docker is not on PATH")
		}
	})
}

// --- Install ---

func TestDocker_Install(t *testing.T) {
	successScript := fstest.MapFS{
		"assets/scripts/docker-linux.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\nexit 0"),
			Mode: 0555,
		},
	}
	failScript := fstest.MapFS{
		"assets/scripts/docker-linux.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\nexit 42"),
			Mode: 0555,
		},
	}

	t.Run("darwin installs via brew cask", func(t *testing.T) {
		brewPath := "/opt/homebrew/bin/brew"
		var brewCommand string
		var brewArgs []string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == brewPath {
					brewCommand = name
					brewArgs = args
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
			BrewPath: &brewPath,
		}
		m := DockerModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		want := []string{"install", "--cask", "docker-desktop"}
		if brewCommand != brewPath || !slices.Equal(brewArgs, want) {
			t.Errorf("brew command = %q %v, want %q %v", brewCommand, brewArgs, brewPath, want)
		}
	})

	t.Run("ubuntu native runs docker-linux.sh script", func(t *testing.T) {
		skipIfWindows(t)
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
			Assets:   successScript,
		}
		m := DockerModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("wsl2 with systemd runs docker-linux.sh script", func(t *testing.T) {
		skipIfWindows(t)
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "systemctl" {
					return []byte("running"), nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
			Assets:   successScript,
		}
		m := DockerModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("wsl2 without systemd logs warning and returns error", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "systemctl" {
					return nil, errors.New("exit status 1")
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		var logBuf bytes.Buffer
		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &logBuf,
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
		}
		m := DockerModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error when systemd is not running")
		}
		logged := logBuf.String()
		if !strings.Contains(logged, "systemd") {
			t.Errorf("expected warning about systemd, got log: %q", logged)
		}
	})

	t.Run("apt repo failure returns AptInstallError", func(t *testing.T) {
		skipIfWindows(t)
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:   context.Background(),
			Log:      &bytes.Buffer{},
			Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
			Assets:   failScript,
		}
		m := DockerModule{}
		err := m.Install(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var aerr myerr.AptInstallError
		if !errors.As(err, &aerr) {
			t.Fatalf("expected AptInstallError, got: %T", err)
		}
	})
}

// --- AuditInfo ---

func TestDocker_AuditInfo(t *testing.T) {
	t.Run("returns docker version", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("Docker version 24.0.7\n"), nil
			},
		})
		m := DockerModule{}
		got := m.AuditInfo()
		if got != "Docker version 24.0.7" {
			t.Errorf("AuditInfo() = %q, want %q", got, "Docker version 24.0.7")
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		m := DockerModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}
