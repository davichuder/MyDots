package modules

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
var zshrcWriteFile = os.WriteFile

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

// IsInstalled checks whether any known framework config directory exists.
// Scans all known nvim-<name> directories under ~/.config/. If any exists,
// the module is considered installed.
func (m NeovimFrameworkModule) IsInstalled(_ platform.Platform) bool {
	home, err := nvimHomeDir()
	if err != nil {
		return false
	}
	for _, dir := range frameworkDirs {
		if _, err := os.Stat(filepath.Join(home, ".config", dir)); err == nil {
			return true
		}
	}
	return false
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
	alias := fmt.Sprintf("alias %s='NVIM_APPNAME=%s nvim'", string(fw), dirName)

	data, err := zshrcReadFile(zshrc)
	if err != nil {
		// .zshrc might not exist yet — treat as empty.
		data = []byte{}
	}

	if bytes.Contains(data, []byte(alias)) {
		return nil // already present, skip
	}

	// Append alias with a newline separator.
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}
	data = append(data, []byte(alias+"\n")...)

	return zshrcWriteFile(zshrc, data, 0644)
}

// AuditInfo returns the framework name (e.g. "lazyvim").
func (m NeovimFrameworkModule) AuditInfo() string {
	return strings.Join(frameworkDirs, ", ")
}
