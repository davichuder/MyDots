package modules

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davichuder/MyDots/internal/backup"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// frameworkRepos maps each Neovim framework name to its git clone URL.
var frameworkRepos = map[config.NvimFramework]string{
	config.NvimFrameworkLazyVim:   "https://github.com/LazyVim/starter",
	config.NvimFrameworkLunarVim:  "https://github.com/LunarVim/Launch.nvim",
	config.NvimFrameworkAstroNvim: "https://github.com/AstroNvim/template",
	config.NvimFrameworkNvChad:    "https://github.com/NvChad/starter",
	config.NvimFrameworkLaunchVim: "https://github.com/LaunchVim/launch.nvim",
}

// frameworkDirs lists all known framework config directories for IsInstalled.
var frameworkDirs = []string{
	"nvim-lazyvim",
	"nvim-lunarvim",
	"nvim-astronvim",
	"nvim-nvchad",
	"nvim-launchnvim",
}

// Injectable file I/O for .zshrc in testing.
var zshrcReadFile = os.ReadFile
var zshrcWriteFile = atomicWriteFile
var zshrcBackupFile = backup.BackupFile

// NeovimFrameworkModule implements the Neovim framework (M-19) installation module.
// It clones the selected framework's starter repo to ~/.config/nvim-<name>
// and appends a shell alias to ~/.zshrc for launching it via NVIM_APPNAME.
type NeovimFrameworkModule struct{}

// Compile-time interface check.
var _ types.Module = NeovimFrameworkModule{}

// NeovimFramework is the exported package-level instance used by the catalogue.
var NeovimFramework types.Module = NeovimFrameworkModule{}

// ID returns the stable module identifier M-19.
func (m NeovimFrameworkModule) ID() types.ModuleID { return types.ModNeovimFramework }

// Name returns a human-readable module name.
func (m NeovimFrameworkModule) Name() string { return "Neovim Framework" }

// Criticality returns NonCritical — framework failure does not stop the pipeline.
func (m NeovimFrameworkModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns M-17 (Neovim base must exist first).
func (m NeovimFrameworkModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModNeovim}
}

// IsInstalled checks the default selected framework for direct callers. The
// executor uses IsInstalledForConfig with the active session configuration.
func (m NeovimFrameworkModule) IsInstalled(p platform.Platform) bool {
	return m.IsInstalledForConfig(p, config.DefaultConfig())
}

// IsInstalledForConfig checks the selected framework directory and its managed
// shell alias.
func (m NeovimFrameworkModule) IsInstalledForConfig(_ platform.Platform, cfg config.Config) bool {
	home, err := nvimHomeDir()
	if err != nil {
		return false
	}
	if _, ok := frameworkRepos[cfg.Nvim.Framework]; !ok {
		return false
	}
	if _, err = os.Stat(filepath.Join(home, ".config", "nvim-"+string(cfg.Nvim.Framework))); err != nil {
		return false
	}
	data, err := zshrcReadFile(filepath.Join(home, ".zshrc"))
	return err == nil && hasFrameworkAlias(data, frameworkAlias(cfg.Nvim.Framework))
}

// Install clones the selected framework repo and appends its shell alias.
func (m NeovimFrameworkModule) Install(ctx types.InstallContext) error {
	fw := ctx.Config.Nvim.Framework
	repoURL, ok := frameworkRepos[fw]
	if !ok {
		return fmt.Errorf("unknown neovim framework: %q", fw)
	}

	home, err := nvimHomeDir()
	if err != nil {
		return err
	}

	dirName := "nvim-" + string(fw)
	targetDir := filepath.Join(home, ".config", dirName)
	zshrc := filepath.Join(home, ".zshrc")

	// Step 1: Clone if target dir doesn't exist yet.
	if _, statErr := os.Stat(targetDir); os.IsNotExist(statErr) {
		// Ensure parent .config dir exists.
		if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
			return err
		}

		if err := runner.Run(ctx.Cancel, ctx.Log, "git", "clone", repoURL, targetDir); err != nil {
			// Clean up partial directory left by failed clone.
			os.RemoveAll(targetDir)
			return err
		}
	}

	// Step 2: Append alias to .zshrc if not already present.
	alias := frameworkAlias(fw)

	data, err := zshrcReadFile(zshrc)
	exists := err == nil
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read .zshrc: %w", err)
		}
		data = []byte{}
	}

	if hasFrameworkAlias(data, alias) {
		return nil // already present, skip
	}

	// Append alias with a newline separator.
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}
	data = append(data, []byte(alias+"\n")...)

	if exists {
		if err := zshrcBackupFile(zshrc, ctx.SessionTimestamp); err != nil {
			return fmt.Errorf("backup .zshrc: %w", err)
		}
	}
	return zshrcWriteFile(zshrc, data, 0644)
}

func frameworkAlias(fw config.NvimFramework) string {
	return fmt.Sprintf("alias %s='NVIM_APPNAME=nvim-%s nvim'", string(fw), string(fw))
}

// hasFrameworkAlias checks the effective alias: zsh uses the last definition
// for a name, so an earlier desired alias cannot mask a later conflicting one.
func hasFrameworkAlias(data []byte, alias string) bool {
	name, _, ok := strings.Cut(alias, "=")
	if !ok {
		return false
	}
	var effective string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, name+"=") {
			effective = line
		}
	}
	return effective == alias
}

// AuditInfo returns the framework name (e.g. "lazyvim").
func (m NeovimFrameworkModule) AuditInfo() string {
	return strings.Join(frameworkDirs, ", ")
}
