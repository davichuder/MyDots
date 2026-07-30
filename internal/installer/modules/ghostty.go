package modules

import (
	"fmt"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// GhosttyModule implements the Ghostty terminal (M-48) installation module.
// On Darwin it uses brew cask. On Linux it runs the embedded ghostty-linux.sh script.
// On WSL2 it requires an active Wayland session (WSLg); if unavailable the module is skipped.
type GhosttyModule struct{}

// Compile-time interface check.
var _ types.Module = GhosttyModule{}

// Ghostty is the exported package-level instance used by the catalogue.
var Ghostty types.Module = GhosttyModule{}

// ID returns the stable module identifier M-48.
func (m GhosttyModule) ID() types.ModuleID { return types.ModGhostty }

// Name returns a human-readable module name.
func (m GhosttyModule) Name() string { return "Ghostty" }

// Criticality returns NonCritical — Ghostty failure does not stop the pipeline.
func (m GhosttyModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency.
func (m GhosttyModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether ghostty is on PATH.
func (m GhosttyModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("ghostty")
}

// Install installs Ghostty based on platform:
//   - Darwin: brew install --cask ghostty
//   - Ubuntu native: run ghostty-linux.sh script
//   - WSL2 + WAYLAND_DISPLAY: run ghostty-linux.sh script
//   - WSL2 + no WAYLAND_DISPLAY: log warning, return nil (skipped)
func (m GhosttyModule) Install(ctx types.InstallContext) error {
	// Darwin: brew cask.
	if ctx.Platform.OS == platform.Darwin {
		return runner.BrewCaskAt(ctx.Cancel, ctx.Log, ctx.Platform, brewPath(ctx), "ghostty")
	}

	// Linux — check Wayland on WSL2 first.
	if ctx.Platform.Variant == platform.WSL2 {
		wayland := lookupEnv("WAYLAND_DISPLAY")
		if wayland == "" {
			if err := writeDiagnostics(ctx.Log,
				"[WARN] WSL2 without Wayland (WSLg) detected — Ghostty cannot be installed.",
				"[WARN] Ghostty requires WSLg to render on WSL2. Install WSLg and try again.",
			); err != nil {
				return fmt.Errorf("write Ghostty skip diagnostics: %w", err)
			}
			return nil
		}
	}

	// Run the embedded Ghostty installation script.
	return runner.Script(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/ghostty-linux.sh", nil)
}

func brewPath(ctx types.InstallContext) string {
	if ctx.BrewPath == nil {
		return ""
	}
	return *ctx.BrewPath
}

// AuditInfo returns the trimmed output of "ghostty --version".
func (m GhosttyModule) AuditInfo() string {
	return runner.CaptureOutput("ghostty", "--version")
}
