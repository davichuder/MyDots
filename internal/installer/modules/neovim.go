package modules

import (
	"os"
	"path/filepath"

	"github.com/davichuder/MyDots/internal/backup"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// NeovimModule implements the Neovim (M-17) installation module.
// It installs Neovim via Homebrew and writes a base init.lua config.
// If the nvim binary already exists but init.lua is missing, it skips
// the brew install and only applies the config.
type NeovimModule struct{}

// Compile-time interface check.
var _ types.Module = NeovimModule{}

// Neovim is the exported package-level instance used by the catalogue.
var Neovim types.Module = NeovimModule{}

// Injectable file I/O for testing.
var nvimHomeDir = os.UserHomeDir
var nvimBackupDir = backup.BackupDir

// baseConfig is the default Neovim init.lua content written to ~/.config/nvim/init.lua.
// It provides sensible defaults (line numbers, relative numbers, tabs, leader key).
const baseConfig string = `-- MyDots base Neovim config
-- Installed by mydots-installer (M-17)

local opt = vim.opt

opt.number = true
opt.relativenumber = true
opt.tabstop = 2
opt.shiftwidth = 2
opt.expandtab = true
opt.mouse = "a"
opt.ignorecase = true
opt.smartcase = true
opt.termguicolors = true
opt.updatetime = 250
`

// ID returns the stable module identifier M-17.
func (m NeovimModule) ID() types.ModuleID { return types.ModNeovim }

// Name returns a human-readable module name.
func (m NeovimModule) Name() string { return "Neovim" }

// Criticality returns NonCritical — Neovim failure does not stop the pipeline.
func (m NeovimModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency (Neovim is installed via brew).
func (m NeovimModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether nvim is on PATH AND init.lua exists.
// Both conditions must be met for the module to be considered installed.
func (m NeovimModule) IsInstalled(_ platform.Platform) bool {
	if !runner.CommandExists("nvim") {
		return false
	}

	home, err := nvimHomeDir()
	if err != nil {
		return false
	}

	initLua := filepath.Join(home, ".config", "nvim", "init.lua")
	_, err = os.Stat(initLua)
	return err == nil
}

// Install installs Neovim (if missing), backs up existing config, and
// writes the base init.lua config.
func (m NeovimModule) Install(ctx types.InstallContext) error {
	// Step 1: Install nvim binary if not present.
	if !runner.CommandExists("nvim") {
		if err := runner.BrewAt(ctx.Cancel, ctx.Log, brewPath(ctx), "install", "neovim"); err != nil {
			return err
		}
	}

	home, err := nvimHomeDir()
	if err != nil {
		return err
	}

	nvimDir := filepath.Join(home, ".config", "nvim")
	initLua := filepath.Join(nvimDir, "init.lua")

	// Step 2: Back up existing config directory if it exists.
	if _, statErr := os.Stat(nvimDir); statErr == nil {
		if err := nvimBackupDir(nvimDir, ctx.SessionTimestamp); err != nil {
			return err
		}
	}

	// Step 3: Create config directory and write base init.lua.
	if err := os.MkdirAll(nvimDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(initLua, []byte(baseConfig), 0644)
}

// AuditInfo returns the trimmed output of "nvim --version".
func (m NeovimModule) AuditInfo() string {
	return runner.CaptureOutput("nvim", "--version")
}
