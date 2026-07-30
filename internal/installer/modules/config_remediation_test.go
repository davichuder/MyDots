package modules

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

func TestConfiguredStateRejectsStaleSelections(t *testing.T) {
	t.Run("theme requires the currently selected theme in every managed file", func(t *testing.T) {
		home := t.TempDir()
		originalHome := themeHomeDir
		themeHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { themeHomeDir = originalHome })

		for _, target := range themeTargets {
			mustWriteFile(t, filepath.Join(home, target.path), target.startMarker+"\n"+target.themeLine("tokyo-night")+"\n"+target.endMarker+"\n")
		}

		module := ThemeModule{}
		if !module.IsInstalledForConfig(platform.Platform{}, config.Config{Theme: config.ThemeTokyoNight}) {
			t.Fatal("expected the configured theme to be installed")
		}
		if module.IsInstalledForConfig(platform.Platform{}, config.Config{Theme: config.ThemeCatppuccin}) {
			t.Fatal("stale theme markers must not satisfy a changed configuration")
		}
		if err := os.Remove(filepath.Join(home, ".config", "ghostty", "config")); err != nil {
			t.Fatal(err)
		}
		if !module.IsInstalledForConfig(platform.Platform{}, config.Config{Theme: config.ThemeTokyoNight}) {
			t.Fatal("a missing optional target skipped by Install must satisfy idempotence")
		}
	})

	t.Run("nerd font requires the selected font and matching Ghostty block", func(t *testing.T) {
		home := t.TempDir()
		mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), ghosttyFontBlock(config.FontFiraCode))
		originalHome := nerdFontHomeDir
		nerdFontHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { nerdFontHomeDir = originalHome })

		withMockExecutor(t, &mockExecutor{executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			if name != "brew" || len(args) != 2 || args[0] != "list" || args[1] != "--cask" {
				return nil, errors.New("unexpected command")
			}
			return []byte("font-fira-code-nerd-font\n"), nil
		}})

		module := NerdFontModule{}
		if !module.IsInstalledForConfig(platform.Platform{OS: platform.Darwin}, config.Config{Font: config.FontFiraCode}) {
			t.Fatal("expected the selected font cask to satisfy idempotence")
		}
		if module.IsInstalledForConfig(platform.Platform{OS: platform.Darwin}, config.Config{Font: config.FontJetBrainsMono}) {
			t.Fatal("a different nerd font must not satisfy idempotence")
		}
		mustWriteFile(t, filepath.Join(home, ".config", "ghostty", "config"), "font-family = FiraCode Nerd Font\n")
		if module.IsInstalledForConfig(platform.Platform{OS: platform.Darwin}, config.Config{Font: config.FontFiraCode}) {
			t.Fatal("an installed font with a stale Ghostty block must retry configuration")
		}
		if err := os.Remove(filepath.Join(home, ".config", "ghostty", "config")); err != nil {
			t.Fatal(err)
		}
		if !module.IsInstalledForConfig(platform.Platform{OS: platform.Darwin}, config.Config{Font: config.FontFiraCode}) {
			t.Fatal("a missing optional Ghostty config skipped by Install must satisfy idempotence")
		}
	})

	t.Run("neovim framework requires the selected directory and alias", func(t *testing.T) {
		home := t.TempDir()
		mustWriteFile(t, filepath.Join(home, ".config", "nvim-lazyvim", ".git", "HEAD"), "ref: refs/heads/main\n")
		originalHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return home, nil }
		originalRead := zshrcReadFile
		zshrcReadFile = os.ReadFile
		t.Cleanup(func() { nvimHomeDir, zshrcReadFile = originalHome, originalRead })
		mustWriteFile(t, filepath.Join(home, ".zshrc"), frameworkAlias(config.NvimFrameworkLazyVim)+"\n")

		module := NeovimFrameworkModule{}
		if !module.IsInstalledForConfig(platform.Platform{}, config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}) {
			t.Fatal("expected the selected framework clone to satisfy idempotence")
		}
		if module.IsInstalledForConfig(platform.Platform{}, config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkAstroNvim}}) {
			t.Fatal("a different framework directory must not satisfy idempotence")
		}
		mustWriteFile(t, filepath.Join(home, ".zshrc"), "export EDITOR=nvim\n")
		if module.IsInstalledForConfig(platform.Platform{}, config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}) {
			t.Fatal("a cloned framework with a missing alias must retry configuration")
		}
		mustWriteFile(t, filepath.Join(home, ".zshrc"), frameworkAlias(config.NvimFrameworkLazyVim)+"\n"+"alias lazyvim='NVIM_APPNAME=nvim-astronvim nvim'\n")
		if module.IsInstalledForConfig(platform.Platform{}, config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}) {
			t.Fatal("a later conflicting alias must override an earlier desired alias")
		}
	})

	t.Run("mcp requires exact managed server definitions", func(t *testing.T) {
		home := t.TempDir()
		originalHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { mcpHomeDir = originalHome })

		writeTestJSON(t, home, map[string]interface{}{"mcp": map[string]interface{}{
			"supabase":   map[string]interface{}{"type": "local", "command": []string{"stale"}, "enabled": true, "env": map[string]string{"STALE": "1"}},
			"angular":    map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@angular/mcp@latest"}, "enabled": true},
			"primeng":    map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "primeng-mcp@latest"}, "enabled": true},
			"postman":    map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@postman/mcp-server-local@latest"}, "enabled": true},
			"context7":   map[string]interface{}{"type": "remote", "url": "https://mcp.context7.com/mcp", "enabled": true},
			"playwright": map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@playwright/mcp@latest"}, "enabled": true},
		}})

		if (McpConfigModule{}).IsInstalled(platform.Platform{}) {
			t.Fatal("a stale managed MCP value must not satisfy idempotence")
		}
	})
}

func TestMcpManagedEntryReconciliationRemovesUnknownFields(t *testing.T) {
	originalHome, originalBackup := mcpHomeDir, mcpBackupFile
	mcpBackupFile = func(string, string) error { return nil }
	t.Cleanup(func() { mcpHomeDir, mcpBackupFile = originalHome, originalBackup })

	home := t.TempDir()
	mcpHomeDir = func() (string, error) { return home, nil }
	writeTestJSON(t, home, map[string]interface{}{"mcp": map[string]interface{}{
		"supabase": map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@supabase/mcp-server-supabase@latest"}, "enabled": true, "env": map[string]string{"STALE": "1"}},
		"custom":   map[string]interface{}{"headers": map[string]string{"keep": "me"}},
	}})

	if (McpConfigModule{}).IsInstalled(platform.Platform{}) {
		t.Fatal("unknown fields in a managed entry must require reconciliation")
	}
	if err := (McpConfigModule{}).Install(types.InstallContext{SessionTimestamp: "test-ts"}); err != nil {
		t.Fatalf("Install() = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if err != nil || bytes.Contains(data, []byte(`"STALE"`)) || !bytes.Contains(data, []byte(`"custom"`)) {
		t.Fatalf("managed/unmanaged reconciliation = %q, %v", data, err)
	}
}

func TestFrameworkRetriesAfterAliasWriteFailure(t *testing.T) {
	home := t.TempDir()
	mustWriteFile(t, filepath.Join(home, ".zshrc"), "export EDITOR=nvim\n")
	if err := os.MkdirAll(filepath.Join(home, ".config", "nvim-lazyvim"), 0755); err != nil {
		t.Fatal(err)
	}

	originalHome, originalRead, originalWrite, originalBackup := nvimHomeDir, zshrcReadFile, zshrcWriteFile, zshrcBackupFile
	nvimHomeDir = func() (string, error) { return home, nil }
	zshrcReadFile = os.ReadFile
	zshrcWriteFile = func(string, []byte, os.FileMode) error { return errors.New("write failed") }
	zshrcBackupFile = func(string, string) error { return nil }
	t.Cleanup(func() {
		nvimHomeDir, zshrcReadFile, zshrcWriteFile, zshrcBackupFile = originalHome, originalRead, originalWrite, originalBackup
	})

	cfg := config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}
	module := NeovimFrameworkModule{}
	if err := module.Install(types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, SessionTimestamp: "test-ts", Config: cfg}); err == nil {
		t.Fatal("expected alias write failure")
	}
	if module.IsInstalledForConfig(platform.Platform{}, cfg) {
		t.Fatal("partial clone must remain retryable after alias write failure")
	}
}

func TestAtomicWriteFileModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX permission bits")
	}
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing")
	if err := os.WriteFile(existing, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := atomicWriteFile(existing, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(existing)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("existing mode = %v, %v; want 0600", info.Mode().Perm(), err)
	}

	created := filepath.Join(dir, "created")
	if err := atomicWriteFile(created, []byte("new"), 0640); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(created)
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("new mode = %v, %v; want 0640", info.Mode().Perm(), err)
	}
}

func TestConfigRewriteSafety(t *testing.T) {
	t.Run("framework backup failure leaves zshrc unchanged and skips the write", func(t *testing.T) {
		home := t.TempDir()
		zshrc := filepath.Join(home, ".zshrc")
		original := []byte("export EDITOR=nvim\n")
		mustWriteFile(t, zshrc, string(original))
		if err := os.MkdirAll(filepath.Join(home, ".config", "nvim-lazyvim"), 0755); err != nil {
			t.Fatal(err)
		}

		originalHome, originalBackup, originalWrite := nvimHomeDir, zshrcBackupFile, zshrcWriteFile
		nvimHomeDir = func() (string, error) { return home, nil }
		zshrcBackupFile = func(string, string) error { return errors.New("backup failed") }
		zshrcWriteFile = func(string, []byte, os.FileMode) error { t.Fatal("write must not follow failed backup"); return nil }
		t.Cleanup(func() { nvimHomeDir, zshrcBackupFile, zshrcWriteFile = originalHome, originalBackup, originalWrite })

		err := (NeovimFrameworkModule{}).Install(types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, SessionTimestamp: "test-ts", Config: config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}})
		if err == nil {
			t.Fatal("expected backup failure")
		}
		got, readErr := os.ReadFile(zshrc)
		if readErr != nil || !bytes.Equal(got, original) {
			t.Fatalf("zshrc changed after backup failure: %q, %v", got, readErr)
		}
	})

	t.Run("mcp write failure preserves existing configuration after backup", func(t *testing.T) {
		home := t.TempDir()
		opencode := filepath.Join(home, ".config", "opencode", "opencode.json")
		original := []byte(`{"theme":"dark","mcp":{}}`)
		mustWriteFile(t, opencode, string(original))

		originalHome, originalBackup, originalWrite := mcpHomeDir, mcpBackupFile, mcpWriteFile
		mcpHomeDir = func() (string, error) { return home, nil }
		backupCalled := false
		mcpBackupFile = func(path, timestamp string) error {
			backupCalled = path == opencode && timestamp == "test-ts"
			return nil
		}
		mcpWriteFile = func(string, []byte, os.FileMode) error { return errors.New("write failed") }
		t.Cleanup(func() { mcpHomeDir, mcpBackupFile, mcpWriteFile = originalHome, originalBackup, originalWrite })

		err := (McpConfigModule{}).Install(types.InstallContext{SessionTimestamp: "test-ts"})
		if err == nil || !backupCalled {
			t.Fatalf("expected backup then write error, got err=%v backup=%t", err, backupCalled)
		}
		got, readErr := os.ReadFile(opencode)
		if readErr != nil || !bytes.Equal(got, original) {
			t.Fatalf("opencode.json changed after write failure: %q, %v", got, readErr)
		}
	})

	t.Run("mcp successful merge creates a backup and preserves unmanaged data", func(t *testing.T) {
		home := t.TempDir()
		opencode := filepath.Join(home, ".config", "opencode", "opencode.json")
		original := []byte(`{"theme":"dark","mcp":{"custom":{"type":"local","enabled":true}}}`)
		mustWriteFile(t, opencode, string(original))

		originalHome, originalBackup := mcpHomeDir, mcpBackupFile
		mcpHomeDir = func() (string, error) { return home, nil }
		var backup []byte
		mcpBackupFile = func(path, _ string) error { backup, _ = os.ReadFile(path); return nil }
		t.Cleanup(func() { mcpHomeDir, mcpBackupFile = originalHome, originalBackup })

		if err := (McpConfigModule{}).Install(types.InstallContext{SessionTimestamp: "test-ts"}); err != nil {
			t.Fatalf("Install() = %v", err)
		}
		if !bytes.Equal(backup, original) {
			t.Fatalf("backup = %q, want original %q", backup, original)
		}
		data, err := os.ReadFile(opencode)
		if err != nil || !bytes.Contains(data, []byte(`"theme": "dark"`)) || !bytes.Contains(data, []byte(`"custom"`)) {
			t.Fatalf("successful merge did not preserve unmanaged data: %q, %v", data, err)
		}
	})
}
