package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// CavemanModule tests — T-069 (RED)
// ---------------------------------------------------------------------------

func TestCaveman_ID(t *testing.T) {
	m := CavemanModule{}
	if got := m.ID(); got != types.ModCaveman {
		t.Errorf("ID() = %v, want %v", got, types.ModCaveman)
	}
}

func TestCaveman_Name(t *testing.T) {
	m := CavemanModule{}
	if got := m.Name(); got != "Caveman" {
		t.Errorf("Name() = %q, want %q", got, "Caveman")
	}
}

func TestCaveman_Criticality(t *testing.T) {
	m := CavemanModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestCaveman_Dependencies(t *testing.T) {
	m := CavemanModule{}
	if deps := m.Dependencies(); deps != nil {
		t.Errorf("Dependencies() = %v, want nil", deps)
	}
}

// --- IsInstalled ---

func TestCaveman_IsInstalled(t *testing.T) {
	t.Run("caveman on PATH returns true", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/local/bin/" + name, nil
			},
		})
		m := CavemanModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when caveman is on PATH")
		}
	})

	t.Run("caveman not on PATH returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		m := CavemanModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when caveman is not on PATH")
		}
	})
}

// --- Install ---

func TestCaveman_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("runs caveman-install.sh script", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"assets/scripts/caveman-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 0"),
				Mode: 0555,
			},
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: mapFS,
		}
		m := CavemanModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
	})

	t.Run("script failure returns error", func(t *testing.T) {
		skipIfWindows(t)
		mapFS := fstest.MapFS{
			"assets/scripts/caveman-install.sh": &fstest.MapFile{
				Data: []byte("#!/bin/sh\nexit 42"),
				Mode: 0555,
			},
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: mapFS,
		}
		m := CavemanModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing script returns error", func(t *testing.T) {
		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Assets: fstest.MapFS{},
		}
		m := CavemanModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error for missing script, got nil")
		}
	})
}

// --- AuditInfo ---

func TestCaveman_AuditInfo(t *testing.T) {
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("caveman 0.3.0\n"), nil
		},
	})
	m := CavemanModule{}
	got := m.AuditInfo()
	if got != "caveman 0.3.0" {
		t.Errorf("AuditInfo() = %q, want %q", got, "caveman 0.3.0")
	}
}
