package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ConfigVersion = "1.0.0"

const DefaultChezmoiRepoURL = "https://github.com/davichuder/dotfiles"

type FontChoice string

const (
	FontJetBrainsMono FontChoice = "JetBrainsMono"
	FontFiraCode      FontChoice = "FiraCode"
	FontCascadiaCode  FontChoice = "CascadiaCode"
	FontHack          FontChoice = "Hack"
	FontIosevka       FontChoice = "Iosevka"
)

type ThemeChoice string

const (
	ThemeTokyoNight  ThemeChoice = "tokyo-night"
	ThemeCatppuccin  ThemeChoice = "catppuccin-mocha"
	ThemeGruvbox     ThemeChoice = "gruvbox-dark"
	ThemeDracula     ThemeChoice = "dracula"
	ThemeKanagawa    ThemeChoice = "kanagawa"
)

type NvimConfig string

const (
	NvimConfigBase     NvimConfig = "base"
	NvimConfigPersonal NvimConfig = "personal"
)

type NvimFramework string

const (
	NvimFrameworkLazyVim   NvimFramework = "lazyvim"
	NvimFrameworkLunarVim  NvimFramework = "lunarvim"
	NvimFrameworkAstroNvim NvimFramework = "astronvim"
	NvimFrameworkNvChad    NvimFramework = "nvchad"
	NvimFrameworkLaunchVim NvimFramework = "launchnvim"
	NvimFrameworkNone      NvimFramework = "none"
)

type LanguageOptions struct {
	Java bool `json:"java"`
	PHP  bool `json:"php"`
}

type NvimOptions struct {
	Config    NvimConfig    `json:"config"`
	Framework NvimFramework `json:"framework"`
}

type ChezmoiOptions struct {
	RepoURL string `json:"repo_url"`
}

type Config struct {
	Version   string          `json:"version"`
	Font      FontChoice      `json:"font"`
	Theme     ThemeChoice     `json:"theme"`
	Languages LanguageOptions `json:"languages"`
	Nvim      NvimOptions     `json:"nvim"`
	Chezmoi   ChezmoiOptions  `json:"chezmoi"`
}

func DefaultConfig() Config {
	return Config{
		Version: ConfigVersion,
		Font:    FontJetBrainsMono,
		Theme:   ThemeTokyoNight,
		Languages: LanguageOptions{
			Java: false,
			PHP:  false,
		},
		Nvim: NvimOptions{
			Config:    NvimConfigBase,
			Framework: NvimFrameworkNone,
		},
		Chezmoi: ChezmoiOptions{
			RepoURL: DefaultChezmoiRepoURL,
		},
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

const tmpFilePattern string = "mydots-config-*.tmp"

func Save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, tmpFilePattern)
	if err != nil {
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	return nil
}

const configDirName string = ".config"

const appDirName string = "mydots"

const configFileName string = "mydots-config.json"

var osUserHomeDir = os.UserHomeDir

func DefaultConfigPath() string {
	home, err := osUserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, configDirName, appDirName, configFileName)
}

func Validate(cfg Config) error {
	switch cfg.Font {
	case FontJetBrainsMono, FontFiraCode, FontCascadiaCode, FontHack, FontIosevka:
	default:
		return fmt.Errorf("invalid font: %q", cfg.Font)
	}

	switch cfg.Theme {
	case ThemeTokyoNight, ThemeCatppuccin, ThemeGruvbox, ThemeDracula, ThemeKanagawa:
	default:
		return fmt.Errorf("invalid theme: %q", cfg.Theme)
	}

	switch cfg.Nvim.Config {
	case NvimConfigBase, NvimConfigPersonal:
	default:
		return fmt.Errorf("invalid nvim config: %q", cfg.Nvim.Config)
	}

	switch cfg.Nvim.Framework {
	case NvimFrameworkLazyVim, NvimFrameworkLunarVim, NvimFrameworkAstroNvim, NvimFrameworkNvChad, NvimFrameworkLaunchVim, NvimFrameworkNone:
	default:
		return fmt.Errorf("invalid nvim framework: %q", cfg.Nvim.Framework)
	}

	if cfg.Chezmoi.RepoURL == "" {
		return fmt.Errorf("chezmoi repo_url cannot be empty")
	}
	const httpsScheme string = "https://"
	if !strings.HasPrefix(cfg.Chezmoi.RepoURL, httpsScheme) {
		return fmt.Errorf("chezmoi repo_url must be an HTTPS URL")
	}

	return nil
}
