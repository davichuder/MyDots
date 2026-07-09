package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != ConfigVersion {
		t.Errorf("expected version %q, got: %q", ConfigVersion, cfg.Version)
	}
	if cfg.Font != FontJetBrainsMono {
		t.Errorf("expected font %q, got: %q", FontJetBrainsMono, cfg.Font)
	}
	if cfg.Theme != ThemeTokyoNight {
		t.Errorf("expected theme %q, got: %q", ThemeTokyoNight, cfg.Theme)
	}
	if cfg.Languages.Java {
		t.Errorf("expected languages.java=false, got: true")
	}
	if cfg.Languages.PHP {
		t.Errorf("expected languages.php=false, got: true")
	}
	if cfg.Nvim.Config != NvimConfigBase {
		t.Errorf("expected nvim.config %q, got: %q", NvimConfigBase, cfg.Nvim.Config)
	}
	if cfg.Nvim.Framework != NvimFrameworkNone {
		t.Errorf("expected nvim.framework %q, got: %q", NvimFrameworkNone, cfg.Nvim.Framework)
	}
	if cfg.Chezmoi.RepoURL != DefaultChezmoiRepoURL {
		t.Errorf("expected chezmoi.repo_url %q, got: %q", DefaultChezmoiRepoURL, cfg.Chezmoi.RepoURL)
	}
}

func writeTestConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const testRepoURL string = "https://github.com/user/dotfiles"

func TestLoadValidJSON(t *testing.T) {
	path := writeTestConfig(t, t.TempDir(), `{
		"version": "1.0.0",
		"font": "JetBrainsMono",
		"theme": "tokyo-night",
		"languages": {"java": true, "php": false},
		"nvim": {"config": "base", "framework": "lazyvim"},
		"chezmoi": {"repo_url": "`+testRepoURL+`"}
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Font != FontJetBrainsMono {
		t.Errorf("expected font JetBrainsMono, got: %q", cfg.Font)
	}
	if cfg.Theme != ThemeTokyoNight {
		t.Errorf("expected theme tokyo-night, got: %q", cfg.Theme)
	}
	if !cfg.Languages.Java {
		t.Errorf("expected java=true, got: false")
	}
	if cfg.Nvim.Framework != NvimFrameworkLazyVim {
		t.Errorf("expected framework lazyvim, got: %q", cfg.Nvim.Framework)
	}
	if cfg.Chezmoi.RepoURL != testRepoURL {
		t.Errorf("expected repo_url %q, got: %q", testRepoURL, cfg.Chezmoi.RepoURL)
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	path := writeTestConfig(t, t.TempDir(), "{invalid json}")

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadUnknownFields(t *testing.T) {
	path := writeTestConfig(t, t.TempDir(), `{
		"version": "1.0.0",
		"font": "JetBrainsMono",
		"theme": "tokyo-night",
		"unknown_field": "should be ignored",
		"languages": {"java": false, "php": false, "extra": true},
		"nvim": {"config": "base", "framework": "none"},
		"chezmoi": {"repo_url": "`+testRepoURL+`"}
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error for unknown fields, got: %v", err)
	}
	if cfg.Font != FontJetBrainsMono {
		t.Errorf("expected font JetBrainsMono after Load, got: %q", cfg.Font)
	}
}

func TestLoadInvalidEnumValue(t *testing.T) {
	path := writeTestConfig(t, t.TempDir(), `{
		"version": "1.0.0",
		"font": "NonExistentFont",
		"theme": "tokyo-night",
		"languages": {"java": false, "php": false},
		"nvim": {"config": "base", "framework": "none"},
		"chezmoi": {"repo_url": "`+testRepoURL+`"}
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed with valid JSON even if enum is invalid, got: %v", err)
	}
	if err := Validate(cfg); err == nil {
		t.Error("Validate should return error for invalid font")
	}
}

func TestSaveCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "config.json")

	cfg := DefaultConfig()
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save should create parent dirs, got: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist after Save")
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("should load saved file, got: %v", err)
	}
	if loaded.Font != cfg.Font {
		t.Errorf("expected font %q after reload, got: %q", cfg.Font, loaded.Font)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := Save(path, DefaultConfig()); err != nil {
		t.Fatal(err)
	}

	cfg2 := DefaultConfig()
	cfg2.Font = FontFiraCode
	cfg2.Theme = ThemeCatppuccin
	if err := Save(path, cfg2); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Font != FontFiraCode {
		t.Errorf("expected FontFiraCode after overwrite, got: %q", loaded.Font)
	}
	if loaded.Theme != ThemeCatppuccin {
		t.Errorf("expected ThemeCatppuccin after overwrite, got: %q", loaded.Theme)
	}
}

func TestSaveWriteError(t *testing.T) {
	dir := t.TempDir()
	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(blockPath, "config.json")
	err := Save(path, DefaultConfig())
	if err == nil {
		t.Error("expected error when parent path is a file, blocking directory creation")
	}
}

func TestValidateValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := Validate(cfg); err != nil {
		t.Errorf("expected no error for valid config, got: %v", err)
	}
}

func TestValidateInvalidFont(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Font = "NonExistent"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid font")
	}
}

func TestValidateInvalidTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "unknown-theme"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid theme")
	}
}

func TestValidateInvalidNvimConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Nvim.Config = "custom"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid nvim config")
	}
}

func TestValidateInvalidNvimFramework(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Nvim.Framework = "unknown-framework"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid nvim framework")
	}
}

func TestValidateChezmoiRepoHTTP(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Chezmoi.RepoURL = "http://github.com/user/dotfiles"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for HTTP chezmoi repo URL")
	}
}

func TestValidateChezmoiRepoEmpty(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Chezmoi.RepoURL = ""
	if err := Validate(cfg); err == nil {
		t.Error("expected error for empty chezmoi repo URL")
	}
}

func TestValidateChezmoiRepoHTTPSPasses(t *testing.T) {
	cfg := DefaultConfig()
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected no error for default HTTPS URL, got: %v", err)
	}

	cfg.Chezmoi.RepoURL = "https://gitlab.com/user/dotfiles"
	if err := Validate(cfg); err != nil {
		t.Errorf("expected no error for different HTTPS URL, got: %v", err)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if !strings.HasSuffix(path, filepath.Join(".config", "mydots", "mydots-config.json")) {
		t.Errorf("unexpected config path suffix, got: %q", path)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, home) {
		t.Errorf("expected path to start with home dir %q, got: %q", home, path)
	}
}

func TestDefaultConfigPathWithMockHome(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) {
		return "/fake/home", nil
	}
	defer func() { osUserHomeDir = orig }()

	got := DefaultConfigPath()
	expected := filepath.Join("/fake/home", ".config", "mydots", "mydots-config.json")
	if got != expected {
		t.Errorf("expected %q, got: %q", expected, got)
	}
}

func TestSaveGoldenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := Save(path, DefaultConfig()); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("should read saved file: %v", err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "default-config.golden"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(golden) {
		t.Error("saved config does not match golden file")
	}
}
