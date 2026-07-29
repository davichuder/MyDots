package modules

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// ThemeModule tests — T-061 (RED)
// ---------------------------------------------------------------------------

func TestTheme_ID(t *testing.T) {
	m := ThemeModule{}
	if got := m.ID(); got != types.ModTheme {
		t.Errorf("ID() = %v, want %v", got, types.ModTheme)
	}
}

func TestTheme_Name(t *testing.T) {
	m := ThemeModule{}
	if got := m.Name(); got != "Theme" {
		t.Errorf("Name() = %q, want %q", got, "Theme")
	}
}

func TestTheme_Criticality(t *testing.T) {
	m := ThemeModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestTheme_Dependencies(t *testing.T) {
	m := ThemeModule{}
	deps := m.Dependencies()
	expected := []types.ModuleID{types.ModNeovim, types.ModOhMyZsh, types.ModZellij, types.ModGhostty}
	if len(deps) != len(expected) {
		t.Fatalf("Dependencies() = %v (len=%d), want %v (len=%d)", deps, len(deps), expected, len(expected))
	}
	for i := range expected {
		if deps[i] != expected[i] {
			t.Errorf("Dependencies()[%d] = %v, want %v", i, deps[i], expected[i])
		}
	}
}

// --- IsInstalled ---

func TestTheme_IsInstalled(t *testing.T) {
	t.Run("all 4 files contain theme markers returns true", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		// Write all 4 config files with theme markers
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"),
			"-- MYDOTS_THEME_START\nvim.cmd.colorscheme(\"tokyo-night\")\n-- MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"),
			"# MYDOTS_THEME_START\nZSH_THEME=\"tokyo-night\"\n# MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"),
			"// MYDOTS_THEME_START\ntheme \"tokyo-night\"\n// MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"),
			"# MYDOTS_THEME_START\ntheme = tokyo-night\n# MYDOTS_THEME_END\n")

		m := ThemeModule{}
		if !m.IsInstalledForConfig(platform.Platform{}, config.Config{Theme: config.ThemeTokyoNight}) {
			t.Error("expected true when all 4 files contain theme markers")
		}
	})

	t.Run("missing one file returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		// Write only 3 files — missing ghostty
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"), "-- MYDOTS_THEME_START\n-- MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"), "# MYDOTS_THEME_START\n# MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"), "// MYDOTS_THEME_START\n// MYDOTS_THEME_END\n")
		// ghostty file NOT created

		m := ThemeModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when one file is missing")
		}
	})

	t.Run("files exist but no theme markers returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		// Write files but without markers
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"), "-- other content\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"), "export EDITOR=nvim\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"), "theme \"default\"\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"), "theme = default\n")

		m := ThemeModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when theme markers not present")
		}
	})
}

// --- Install ---

func TestTheme_Install(t *testing.T) {
	t.Run("first run writes markers to existing config files", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		// Pre-create all 4 config files (empty) so the module modifies them.
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"), "-- existing content\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"), "export EDITOR=nvim\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"), "theme \"default\"\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"), "theme = default\n")

		// Track written files
		var writtenFiles []string
		origWrite := themeWriteFile
		origBackup := themeBackupFile
		t.Cleanup(func() {
			themeWriteFile = origWrite
			themeBackupFile = origBackup
		})
		themeWriteFile = func(path string, data []byte, perm os.FileMode) error {
			writtenFiles = append(writtenFiles, path)
			return origWrite(path, data, perm)
		}
		themeBackupFile = func(path, ts string) error { return nil } // no-op backup

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config: config.Config{
				Theme: config.ThemeTokyoNight,
			},
		}
		m := ThemeModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify all 4 files were written (modified)
		if len(writtenFiles) != 4 {
			t.Fatalf("expected 4 files written, got %d: %v", len(writtenFiles), writtenFiles)
		}

		// Verify content of each file contains the correct markers
		for _, path := range writtenFiles {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read written file %s: %v", path, err)
			}
			content := string(data)
			if !strings.Contains(content, "MYDOTS_THEME_START") {
				t.Errorf("file %s missing MYDOTS_THEME_START marker\ncontent: %q", path, content)
			}
			if !strings.Contains(content, "MYDOTS_THEME_END") {
				t.Errorf("file %s missing MYDOTS_THEME_END marker\ncontent: %q", path, content)
			}
			if !strings.Contains(content, "tokyo-night") {
				t.Errorf("file %s missing theme name 'tokyo-night'\ncontent: %q", path, content)
			}
		}
	})

	t.Run("re-run with same theme does not change content", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		origWrite := themeWriteFile
		origBackup := themeBackupFile
		t.Cleanup(func() {
			themeWriteFile = origWrite
			themeBackupFile = origBackup
		})
		themeBackupFile = func(path, ts string) error { return nil }

		// Pre-create files with theme markers already set (simulating first run).
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"),
			"-- MYDOTS_THEME_START\nvim.cmd.colorscheme(\"tokyo-night\")\n-- MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"),
			"# MYDOTS_THEME_START\nZSH_THEME=\"tokyo-night\"\n# MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"),
			"// MYDOTS_THEME_START\ntheme \"tokyo-night\"\n// MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"),
			"# MYDOTS_THEME_START\ntheme = tokyo-night\n# MYDOTS_THEME_END\n")

		// Track writes on second run
		var writeCount int
		themeWriteFile = func(path string, data []byte, perm os.FileMode) error {
			writeCount++
			return origWrite(path, data, perm)
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config:           config.Config{Theme: config.ThemeTokyoNight},
		}
		m := ThemeModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify no writes happened (content unchanged)
		if writeCount != 0 {
			t.Errorf("expected 0 writes on re-run with same theme, got %d", writeCount)
		}
	})

	t.Run("re-run with different theme replaces content between markers", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		origWrite := themeWriteFile
		origBackup := themeBackupFile
		t.Cleanup(func() {
			themeWriteFile = origWrite
			themeBackupFile = origBackup
		})
		themeBackupFile = func(path, ts string) error { return nil }

		// Pre-create files with tokyo-night theme markers already set.
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"),
			"-- MYDOTS_THEME_START\nvim.cmd.colorscheme(\"tokyo-night\")\n-- MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"),
			"# MYDOTS_THEME_START\nZSH_THEME=\"tokyo-night\"\n# MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"),
			"// MYDOTS_THEME_START\ntheme \"tokyo-night\"\n// MYDOTS_THEME_END\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "ghostty", "config"),
			"# MYDOTS_THEME_START\ntheme = tokyo-night\n# MYDOTS_THEME_END\n")

		// Second run with catppuccin-mocha — replaces existing theme
		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config:           config.Config{Theme: config.ThemeCatppuccin},
		}
		m := ThemeModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify tokyo-night was replaced by catppuccin-mocha
		for _, target := range themeTargets {
			path := filepath.Join(tmpDir, target.path)
			data, _ := os.ReadFile(path)
			content := string(data)
			if strings.Contains(content, "tokyo-night") {
				t.Errorf("tokyo-night should not remain in %s after theme change\ncontent: %q", target.path, content)
			}
			if !strings.Contains(content, "catppuccin-mocha") || !strings.Contains(content, "MYDOTS_THEME_START") {
				t.Errorf("expected catppuccin-mocha with markers in %s\ncontent: %q", target.path, content)
			}
		}
	})

	t.Run("missing config file is skipped, others written", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		origWrite := themeWriteFile
		origBackup := themeBackupFile
		t.Cleanup(func() {
			themeWriteFile = origWrite
			themeBackupFile = origBackup
		})
		themeBackupFile = func(path, ts string) error { return nil }

		// Pre-create 3 files (nvim, zsh, zellij) — ghostty is left missing.
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "nvim", "lua", "plugins", "colorscheme.lua"), "-- content\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"), "export EDITOR=nvim\n")
		mustWriteFile(t, filepath.Join(tmpDir, ".config", "zellij", "config.kdl"), "theme \"default\"\n")
		// ghostty file NOT created

		// Make the ghostty read return not-found even if the path is somehow
		// constructed (defensive — the above does not create it, but ensures
		// the test does not depend on side effects).
		origRead := themeReadFile
		t.Cleanup(func() { themeReadFile = origRead })
		themeReadFile = func(path string) ([]byte, error) {
			if strings.Contains(path, "ghostty") {
				return nil, os.ErrNotExist
			}
			return origRead(path)
		}

		// Track which files are written.
		var written []string
		themeWriteFile = func(path string, data []byte, perm os.FileMode) error {
			written = append(written, path)
			return origWrite(path, data, perm)
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config:           config.Config{Theme: config.ThemeTokyoNight},
		}
		m := ThemeModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Should have written 3 files (ghostty was missing and skipped)
		if len(written) != 3 {
			t.Fatalf("expected 3 files written, got %d: %v", len(written), written)
		}
		for _, path := range written {
			if strings.Contains(path, "ghostty") {
				t.Error("ghostty file should not have been written when it was missing")
			}
		}
	})

	t.Run("backup is created for each existing file before writing", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := themeHomeDir
		themeHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { themeHomeDir = origHome })

		origWrite := themeWriteFile
		origMkdir := themeMkdirAll
		t.Cleanup(func() {
			themeWriteFile = origWrite
			themeMkdirAll = origMkdir
		})
		themeMkdirAll = func(path string, perm os.FileMode) error { return origMkdir(path, perm) }

		// Pre-create a .zshrc file so backup is needed
		zshrc := filepath.Join(tmpDir, ".zshrc")
		os.WriteFile(zshrc, []byte("export EDITOR=nvim\n"), 0644)

		var backupCalls []string
		origBkp := themeBackupFile
		themeBackupFile = func(path, ts string) error {
			backupCalls = append(backupCalls, path)
			return origBkp(path, ts)
		}
		t.Cleanup(func() { themeBackupFile = origBkp })

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
			Config:           config.Config{Theme: config.ThemeTokyoNight},
		}
		m := ThemeModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Backup should have been called for .zshrc (the only pre-existing file)
		if len(backupCalls) != 1 || !strings.Contains(backupCalls[0], ".zshrc") {
			t.Errorf("expected backup for existing .zshrc, got %v", backupCalls)
		}
	})
}

// --- AuditInfo ---

func TestTheme_AuditInfo(t *testing.T) {
	t.Run("returns empty when no theme set", func(t *testing.T) {
		m := ThemeModule{}
		got := m.AuditInfo()
		if got != "" {
			t.Errorf("AuditInfo() = %q, want empty string", got)
		}
	})
}

// ---------------------------------------------------------------------------
// Golden file tests for Theme module
// ---------------------------------------------------------------------------

func TestTheme_GoldenFiles(t *testing.T) {
	// Golden files test: write theme to a temp dir, then compare each
	// output file against the expected golden content.
	//
	// This test is marked as update-golden compatible:
	//   go test -run TestTheme_GoldenFiles -update
	// rewrites the .golden files from the current output.
	t.Skip("golden file test scaffolding — enable after golden files are committed")
}

// --- helpers ---

// mustWriteFile writes content to path, creating parent directories.
func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
