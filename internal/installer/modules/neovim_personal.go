package modules

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// NeovimPersonalModule implements the Neovim personal config (M-18) installation module.
// It backs up ~/.config/nvim/ and runs `chezmoi apply` to deploy the user's
// personal Neovim configuration from their dotfiles repository.
//
// Dependencies: M-17 (Neovim base) and M-45 (chezmoi). If chezmoi fails, the
// executor propagates the failure through failedIDs and skips this module.
type NeovimPersonalModule struct{}

// Compile-time interface check.
var _ types.Module = NeovimPersonalModule{}

// NeovimPersonal is the exported package-level instance used by the catalogue.
var NeovimPersonal types.Module = NeovimPersonalModule{}

// ID returns the stable module identifier M-18.
func (m NeovimPersonalModule) ID() types.ModuleID { return types.ModNeovimPersonal }

// Name returns a human-readable module name.
func (m NeovimPersonalModule) Name() string { return "Neovim Personal" }

// Criticality returns NonCritical — personal config failure does not stop the pipeline.
func (m NeovimPersonalModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns M-17 (Neovim base) and M-45 (chezmoi).
// The executor uses these for failure propagation: if chezmoi is in failedIDs,
// this module is automatically skipped with StatusSkippedDependencyFailed.
func (m NeovimPersonalModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModNeovim, types.ModChezmoi}
}

// IsInstalled checks whether chezmoi status shows no pending changes.
// An empty output means the personal config is already applied.
// Uses runner.Run instead of CaptureOutput so we can distinguish
// "empty output" (no diffs) from "command error" (chezmoi not initialized).
func (m NeovimPersonalModule) IsInstalled(_ platform.Platform) bool {
	var buf bytes.Buffer
	err := runner.Run(context.Background(), &buf, "chezmoi", "status")
	if err != nil {
		return false
	}
	return strings.TrimSpace(buf.String()) == ""
}

// Install backs up ~/.config/nvim/ then runs chezmoi apply.
func (m NeovimPersonalModule) Install(ctx types.InstallContext) error {
	home, err := nvimHomeDir()
	if err != nil {
		return err
	}

	nvimDir := filepath.Join(home, ".config", "nvim")

	// Step 1: Back up existing config directory if it exists.
	if _, statErr := os.Stat(nvimDir); statErr == nil {
		if err := nvimBackupDir(nvimDir, ctx.SessionTimestamp); err != nil {
			return err
		}
	}

	// Step 2: Deploy personal config via chezmoi.
	return runner.Run(ctx.Cancel, ctx.Log, "chezmoi", "apply")
}

// AuditInfo returns "personal" to indicate personal config mode.
func (m NeovimPersonalModule) AuditInfo() string {
	return "personal"
}
