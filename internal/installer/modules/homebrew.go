// Package modules contains complex module implementations that require
// OS-specific logic, multi-step installs, or config writes.
// Simple modules are declared as data in the catalogue package.
package modules

import (
	"fmt"

	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// HomebrewModule implements the Homebrew (M-01) installation module.
// It installs Homebrew via the official curl script embedded in assets/.
// Homebrew is critical: if it fails, the entire install pipeline stops.
type HomebrewModule struct{}

// Ensure HomebrewModule satisfies the Module interface at compile time.
var _ types.Module = HomebrewModule{}

// Homebrew is the exported package-level instance used by the catalogue.
var Homebrew types.Module = HomebrewModule{}

var runHomebrewScript = runner.Script

// ID returns the stable module identifier M-01.
func (m HomebrewModule) ID() types.ModuleID { return types.ModHomebrew }

// Name returns a human-readable module name.
func (m HomebrewModule) Name() string { return "Homebrew" }

// Criticality returns Critical — Homebrew is required by most other modules.
func (m HomebrewModule) Criticality() types.Criticality { return types.Critical }

// Dependencies returns nil because Homebrew has no module dependencies.
func (m HomebrewModule) Dependencies() []types.ModuleID { return nil }

// IsInstalled checks whether the brew binary is reachable on PATH.
func (m HomebrewModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("brew")
}

// Install runs the embedded Homebrew install script via runner.Script.
// ctx.Assets provides the embedded filesystem (set by the main package).
// Returns a CurlScriptError if the script exits with a non-zero code.
func (m HomebrewModule) Install(ctx types.InstallContext) error {
	err := runHomebrewScript(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/homebrew-install.sh", nil)
	if err != nil {
		return wrapScriptError(err)
	}
	brewPath, err := runner.RefreshBrew(ctx.Platform)
	if err != nil {
		return fmt.Errorf("discover Homebrew after installation: %w", err)
	}
	if ctx.BrewPath != nil {
		*ctx.BrewPath = brewPath
	}
	return nil
}

// AuditInfo returns the trimmed output of "brew --version".
func (m HomebrewModule) AuditInfo() string {
	return runner.CaptureOutput("brew", "--version")
}

// wrapScriptError converts a runner.Script error into a CurlScriptError
// with What/Why/Fix context for TUI rendering.
func wrapScriptError(err error) error {
	// Try to extract exit code from the error message.
	exitCode := 1
	_, _ = fmt.Sscanf(err.Error(), "exit status %d", &exitCode)

	return myerr.CurlScriptError{
		URL:      "assets/scripts/homebrew-install.sh",
		ExitCode: exitCode,
		Stderr:   err.Error(),
	}
}
