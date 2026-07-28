package modules

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// Injectable dependencies for testing
// ---------------------------------------------------------------------------

var chezmoiHomeDir = os.UserHomeDir
var chezmoiRunBrew = runner.Brew
var chezmoiRun = runner.Run

// ---------------------------------------------------------------------------
// ChezmoiModule (M-45)
// ---------------------------------------------------------------------------

// ChezmoiModule implements the chezmoi dotfile manager (M-45) for MyDots.
// It installs chezmoi via Homebrew, initializes it with the user's dotfiles
// repo URL, and applies the configuration.
type ChezmoiModule struct{}

// Compile-time interface check.
var _ types.Module = ChezmoiModule{}

// Chezmoi is the exported package-level instance used by the catalogue.
var Chezmoi types.Module = ChezmoiModule{}

// ID returns the stable module identifier M-45.
func (m ChezmoiModule) ID() types.ModuleID { return types.ModChezmoi }

// Name returns a human-readable module name.
func (m ChezmoiModule) Name() string { return "Chezmoi" }

// Criticality returns NonCritical — chezmoi failure does not stop the pipeline.
func (m ChezmoiModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew and Git as requirements.
func (m ChezmoiModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew, types.ModGit}
}

// IsInstalled checks whether chezmoi is installed and initialized:
//   - `chezmoi` binary is on PATH
//   - ~/.local/share/chezmoi/ exists and contains a .git directory
func (m ChezmoiModule) IsInstalled(_ platform.Platform) bool {
	if !runner.CommandExists("chezmoi") {
		return false
	}

	home, err := chezmoiHomeDir()
	if err != nil {
		return false
	}

	chezmoiDir := filepath.Join(home, ".local", "share", "chezmoi")
	gitDir := filepath.Join(chezmoiDir, ".git")

	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// Install installs chezmoi and applies the user's dotfiles:
//   1. brew install chezmoi
//   2. chezmoi init <repo_url>
//   3. chezmoi apply
//
// If chezmoi init fails, chezmoi apply is not called.
func (m ChezmoiModule) Install(ctx types.InstallContext) error {
	// Step 1: brew install chezmoi.
	if err := chezmoiRunBrew(ctx.Cancel, logWriter(ctx.Log), "install", "chezmoi"); err != nil {
		return fmt.Errorf("brew install chezmoi: %w", err)
	}

	// Step 2: chezmoi init <repo_url>.
	repoURL := ctx.Config.Chezmoi.RepoURL
	if err := chezmoiRun(ctx.Cancel, logWriter(ctx.Log), "chezmoi", "init", repoURL); err != nil {
		return fmt.Errorf("chezmoi init %s: %w", repoURL, err)
	}

	// Step 3: chezmoi apply.
	if err := chezmoiRun(ctx.Cancel, logWriter(ctx.Log), "chezmoi", "apply"); err != nil {
		return fmt.Errorf("chezmoi apply: %w", err)
	}

	return nil
}

// AuditInfo returns the installed chezmoi version.
func (m ChezmoiModule) AuditInfo() string {
	return runner.CaptureOutput("chezmoi", "--version")
}

// logWriter adapts io.Writer to all the places we need it.
func logWriter(w io.Writer) io.Writer {
	if w == nil {
		// Return a discard writer to avoid nil panics.
		return io.Discard
	}
	return w
}


