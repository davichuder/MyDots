package modules

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// NerdFontModule tests — T-063 (RED)
// ---------------------------------------------------------------------------

func TestNerdFont_ID(t *testing.T) {
	m := NerdFontModule{}
	if got := m.ID(); got != types.ModNerdFont {
		t.Errorf("ID() = %v, want %v", got, types.ModNerdFont)
	}
}

func TestNerdFont_Name(t *testing.T) {
	m := NerdFontModule{}
	if got := m.Name(); got != "Nerd Font" {
		t.Errorf("Name() = %q, want %q", got, "Nerd Font")
	}
}

func TestNerdFont_Criticality(t *testing.T) {
	m := NerdFontModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestNerdFont_Dependencies(t *testing.T) {
	m := NerdFontModule{}
	deps := m.Dependencies()
	if len(deps) != 0 {
		t.Errorf("Dependencies() = %v, want empty", deps)
	}
}

// --- IsInstalled ---

func TestNerdFont_IsInstalled(t *testing.T) {
	t.Run("darwin font cask listed returns true", func(t *testing.T) {
		home := t.TempDir()
		mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), ghosttyFontBlock(config.DefaultConfig().Font))
		originalHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { nerdFontHomeDir = originalHome })

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" && len(args) >= 2 && args[0] == "list" && args[1] == "--cask" {
					return []byte("font-jetbrains-mono-nerd-font\nfont-fira-code-nerd-font\n"), nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})
		m := NerdFontModule{}
		if !m.IsInstalled(platform.Platform{OS: platform.Darwin}) {
			t.Error("expected true when brew cask list shows nerd fonts")
		}
	})

	t.Run("darwin no font casks returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte(""), nil
			},
		})
		m := NerdFontModule{}
		if m.IsInstalled(platform.Platform{OS: platform.Darwin}) {
			t.Error("expected false when no nerd font casks installed")
		}
	})

	t.Run("linux fc-list shows Nerd returns true", func(t *testing.T) {
		home := t.TempDir()
		mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), ghosttyFontBlock(config.DefaultConfig().Font))
		originalHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { nerdFontHomeDir = originalHome })

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "fc-list" {
					return []byte("JetBrainsMono Nerd Font:style=Regular\n"), nil
				}
				return nil, errors.New("unexpected: " + name)
			},
		})
		m := NerdFontModule{}
		if !m.IsInstalled(platform.Platform{OS: platform.Linux}) {
			t.Error("expected true when fc-list shows Nerd Font")
		}
	})

	t.Run("linux fc-list no Nerd returns false", func(t *testing.T) {
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("Noto Sans:style=Regular\n"), nil
			},
		})
		m := NerdFontModule{}
		if m.IsInstalled(platform.Platform{OS: platform.Linux}) {
			t.Error("expected false when fc-list has no Nerd Font")
		}
	})
}

// --- Install ---

func TestNerdFont_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("darwin installs via brew cask with correct name", func(t *testing.T) {
		var caskName string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "brew" && len(args) >= 2 && args[0] == "--cask" {
					caskName = args[1]
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			Platform:         platform.Platform{OS: platform.Darwin},
			Config:           config.Config{Font: config.FontJetBrainsMono},
			SessionTimestamp: "test-ts",
		}
		m := NerdFontModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if caskName != "font-jetbrains-mono-nerd-font" {
			t.Errorf("cask name = %q, want %q", caskName, "font-jetbrains-mono-nerd-font")
		}
	})

	t.Run("triangulation — all 5 fonts map to correct cask on darwin", func(t *testing.T) {
		cases := []struct {
			font     config.FontChoice
			wantCask string
		}{
			{config.FontJetBrainsMono, "font-jetbrains-mono-nerd-font"},
			{config.FontFiraCode, "font-fira-code-nerd-font"},
			{config.FontCascadiaCode, "font-caskaydia-cove-nerd-font"},
			{config.FontHack, "font-hack-nerd-font"},
			{config.FontIosevka, "font-iosevka-nerd-font"},
		}
		for _, tc := range cases {
			t.Run(string(tc.font), func(t *testing.T) {
				var caskName string
				withMockExecutor(t, &mockExecutor{
					executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
						if name == "brew" && len(args) >= 2 && args[0] == "--cask" {
							caskName = args[1]
							return nil, nil
						}
						return nil, nil
					},
				})
				ctx := types.InstallContext{
					Cancel:           context.Background(),
					Log:              &bytes.Buffer{},
					Platform:         platform.Platform{OS: platform.Darwin},
					Config:           config.Config{Font: tc.font},
					SessionTimestamp: "test-ts",
				}
				m := NerdFontModule{}
				if err := m.Install(ctx); err != nil {
					t.Fatalf("Install(%s) = %v, want nil", tc.font, err)
				}
				if caskName != tc.wantCask {
					t.Errorf("cask = %q, want %q", caskName, tc.wantCask)
				}
			})
		}
	})

	t.Run("linux runs font-linux.sh with correct FONT_NAME", func(t *testing.T) {
		var envVars map[string]string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("unexpected command: " + name)
			},
		})

		envVars = make(map[string]string)
		origScript := nerdFontRunScript
		t.Cleanup(func() { nerdFontRunScript = origScript })
		nerdFontRunScript = func(_ context.Context, _ io.Writer, _ fs.FS, _ string, env map[string]string) error {
			for k, v := range env {
				envVars[k] = v
			}
			return nil
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			Platform:         platform.Platform{OS: platform.Linux},
			Config:           config.Config{Font: config.FontJetBrainsMono},
			SessionTimestamp: "test-ts",
		}
		m := NerdFontModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if envVars["FONT_NAME"] != "JetBrainsMono.zip" {
			t.Errorf("FONT_NAME env = %v, want %q", envVars, "JetBrainsMono.zip")
		}
	})

	t.Run("post-install updates ghostty config with font-family marker", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nerdFontHomeDir = origHome })

		origWrite := nerdFontWriteFile
		origBkp := nerdFontBackupFile
		t.Cleanup(func() {
			nerdFontWriteFile = origWrite
			nerdFontBackupFile = origBkp
		})
		nerdFontBackupFile = func(path, ts string) error { return nil }

		// Pre-create ghostty config so it gets modified
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"), "theme = tokyo-night\n")

		// Mock the runner
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})

		// Mock the script runner
		origScript := nerdFontRunScript
		t.Cleanup(func() { nerdFontRunScript = origScript })
		nerdFontRunScript = func(_ context.Context, _ io.Writer, _ fs.FS, _ string, _ map[string]string) error {
			return nil
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			Platform:         platform.Platform{OS: platform.Linux},
			Config:           config.Config{Font: config.FontJetBrainsMono},
			SessionTimestamp: "test-ts",
		}
		m := NerdFontModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify ghostty config has font-family marker
		ghosttyConfig := filepath.Join(tmpDir, ".config", "ghostty", "config")
		data, err := os.ReadFile(ghosttyConfig)
		if err != nil {
			t.Fatalf("read ghostty config: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "MYDOTS_FONT_START") {
			t.Errorf("ghostty config missing MYDOTS_FONT_START\ncontent: %q", content)
		}
		if !strings.Contains(content, "MYDOTS_FONT_END") {
			t.Errorf("ghostty config missing MYDOTS_FONT_END\ncontent: %q", content)
		}
		if !strings.Contains(content, "JetBrainsMono Nerd Font") {
			t.Errorf("ghostty config missing font-family\ncontent: %q", content)
		}
	})

	t.Run("re-run with same font does not duplicate ghostty entry", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nerdFontHomeDir = origHome })

		origWrite := nerdFontWriteFile
		origBkp := nerdFontBackupFile
		t.Cleanup(func() {
			nerdFontWriteFile = origWrite
			nerdFontBackupFile = origBkp
		})
		nerdFontBackupFile = func(path, ts string) error { return nil }

		// Pre-create ghostty config with font marker already set
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"),
			"# MYDOTS_FONT_START\nfont-family = JetBrainsMono Nerd Font\n# MYDOTS_FONT_END\n")

		// Mock runner
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return nil, nil
			},
		})

		origScript := nerdFontRunScript
		t.Cleanup(func() { nerdFontRunScript = origScript })
		nerdFontRunScript = func(_ context.Context, _ io.Writer, _ fs.FS, _ string, _ map[string]string) error {
			return nil
		}

		// Track writes to ghostty config
		var ghosttyWriteCount int
		nerdFontWriteFile = func(path string, data []byte, perm os.FileMode) error {
			if strings.Contains(path, "ghostty") {
				ghosttyWriteCount++
			}
			return origWrite(path, data, perm)
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			Platform:         platform.Platform{OS: platform.Linux},
			Config:           config.Config{Font: config.FontJetBrainsMono},
			SessionTimestamp: "test-ts",
		}
		m := NerdFontModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if ghosttyWriteCount != 0 {
			t.Errorf("expected 0 ghostty writes on re-run with same font, got %d", ghosttyWriteCount)
		}
	})
}

// --- AuditInfo ---

func TestNerdFont_AuditInfo(t *testing.T) {
	t.Run("returns empty when no font installed", func(t *testing.T) {
		m := NerdFontModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})

	t.Run("returns selected font from a normal marker block", func(t *testing.T) {
		home := t.TempDir()
		mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), ghosttyFontBlock(config.FontFiraCode))
		originalHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { nerdFontHomeDir = originalHome })

		if got := (NerdFontModule{}).AuditInfo(); got != "FiraCode Nerd Font" {
			t.Errorf("AuditInfo() = %q, want %q", got, "FiraCode Nerd Font")
		}
	})

	for _, tt := range []struct {
		name    string
		content string
	}{
		{"start marker without a block", "# MYDOTS_FONT_START\n"},
		{"invalid font directive", "# MYDOTS_FONT_START\nfont-family\n# MYDOTS_FONT_END\n"},
		{"empty font directive", "# MYDOTS_FONT_START\nfont-family = \n# MYDOTS_FONT_END\n"},
		{"empty quoted font directive", "# MYDOTS_FONT_START\nfont-family = \"\"\n# MYDOTS_FONT_END\n"},
		{"extra equals in font directive", "# MYDOTS_FONT_START\nfont-family == Hack Nerd Font\n# MYDOTS_FONT_END\n"},
		{"spaced extra separator in font directive", "# MYDOTS_FONT_START\nfont-family = = Hack Nerd Font\n# MYDOTS_FONT_END\n"},
		{"missing font-family key", "# MYDOTS_FONT_START\n= Hack Nerd Font\n# MYDOTS_FONT_END\n"},
		{"whitespace within font-family key", "# MYDOTS_FONT_START\nfont- family = Hack Nerd Font\n# MYDOTS_FONT_END\n"},
		{"duplicate font directives", "# MYDOTS_FONT_START\nfont-family = FiraCode Nerd Font\nfont-family = Hack Nerd Font\n# MYDOTS_FONT_END\n"},
		{"valid directive followed by EOF", "# MYDOTS_FONT_START\nfont-family = FiraCode Nerd Font\n"},
		{"end marker before start marker", "# MYDOTS_FONT_END\n# MYDOTS_FONT_START\nfont-family = FiraCode Nerd Font\n# MYDOTS_FONT_END\n"},
		{"conflicting complete blocks", "# MYDOTS_FONT_START\nfont-family = FiraCode Nerd Font\n# MYDOTS_FONT_END\n# MYDOTS_FONT_START\nfont-family = Hack Nerd Font\n# MYDOTS_FONT_END\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), tt.content)
			originalHome := nerdFontHomeDir
			nerdFontHomeDir = func() (string, error) { return home, nil }
			t.Cleanup(func() { nerdFontHomeDir = originalHome })

			if got := (NerdFontModule{}).AuditInfo(); got != "" {
				t.Errorf("AuditInfo() = %q, want empty", got)
			}
		})
	}

	for _, tt := range []struct {
		name    string
		content string
		want    string
	}{
		{"no whitespace around separator", "# MYDOTS_FONT_START\nfont-family=Hack Nerd Font\n# MYDOTS_FONT_END\n", "Hack Nerd Font"},
		{"tab whitespace around separator", "# MYDOTS_FONT_START\n\tfont-family\t=\tHack Nerd Font\n# MYDOTS_FONT_END\n", "Hack Nerd Font"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), tt.content)
			originalHome := nerdFontHomeDir
			nerdFontHomeDir = func() (string, error) { return home, nil }
			t.Cleanup(func() { nerdFontHomeDir = originalHome })

			if got := (NerdFontModule{}).AuditInfo(); got != tt.want {
				t.Errorf("AuditInfo() = %q, want %q", got, tt.want)
			}
		})
	}
}
