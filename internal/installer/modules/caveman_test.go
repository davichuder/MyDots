package modules

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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
	tests := []struct {
		name         string
		skill        string
		skillDir     bool
		missingSkill bool
		soul         string
		installed    bool
	}{
		{name: "valid complete state", skill: "skill", soul: "<!-- caveman-begin -->\n<!-- caveman-end -->\n", installed: true},
		{name: "duplicate begin markers", skill: "skill", soul: "<!-- caveman-begin -->\n<!-- caveman-begin -->\n<!-- caveman-end -->\n"},
		{name: "duplicate end markers", skill: "skill", soul: "<!-- caveman-begin -->\n<!-- caveman-end -->\n<!-- caveman-end -->\n"},
		{name: "orphan begin marker", skill: "skill", soul: "<!-- caveman-begin -->\n"},
		{name: "orphan end marker", skill: "skill", soul: "<!-- caveman-end -->\n"},
		{name: "reversed markers", skill: "skill", soul: "<!-- caveman-end -->\n<!-- caveman-begin -->\n"},
		{name: "missing skill", missingSkill: true, soul: "<!-- caveman-begin -->\n<!-- caveman-end -->\n"},
		{name: "empty skill", skill: "", soul: "<!-- caveman-begin -->\n<!-- caveman-end -->\n"},
		{name: "skill path is directory", skillDir: true, soul: "<!-- caveman-begin -->\n<!-- caveman-end -->\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := t.TempDir()
			t.Setenv("OPENCLAW_WORKSPACE", workspace)
			skillPath := filepath.Join(workspace, "skills", "caveman", "SKILL.md")
			if !tt.missingSkill {
				if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
					t.Fatalf("MkdirAll skill directory: %v", err)
				}
				if tt.skillDir {
					if err := os.Mkdir(skillPath, 0755); err != nil {
						t.Fatalf("Mkdir skill path: %v", err)
					}
				} else if err := os.WriteFile(skillPath, []byte(tt.skill), 0600); err != nil {
					t.Fatalf("WriteFile skill: %v", err)
				}
			}
			if err := os.WriteFile(filepath.Join(workspace, "SOUL.md"), []byte(tt.soul), 0600); err != nil {
				t.Fatalf("WriteFile SOUL.md: %v", err)
			}

			if got := (CavemanModule{}).IsInstalled(platform.Platform{}); got != tt.installed {
				t.Errorf("IsInstalled() = %t, want %t", got, tt.installed)
			}
		})
	}
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
	if got != "OpenClaw workspace skill" {
		t.Errorf("AuditInfo() = %q, want %q", got, "OpenClaw workspace skill")
	}
}
