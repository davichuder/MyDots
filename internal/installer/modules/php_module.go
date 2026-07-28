package modules

import (
	"fmt"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// Injectable dependencies for testing
// ---------------------------------------------------------------------------

var phpRunBrew = runner.Brew

// ---------------------------------------------------------------------------
// PhpModule (M-16)
// ---------------------------------------------------------------------------

// PhpModule implements PHP installation via Homebrew (M-16).
type PhpModule struct{}

// Compile-time interface check.
var _ types.Module = PhpModule{}

// Php is the exported package-level instance used by the catalogue.
var Php types.Module = PhpModule{}

// ID returns the stable module identifier M-16.
func (m PhpModule) ID() types.ModuleID { return types.ModPhp }

// Name returns a human-readable module name.
func (m PhpModule) Name() string { return "PHP" }

// Criticality returns NonCritical — PHP failure does not stop the pipeline.
func (m PhpModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a requirement.
func (m PhpModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether php is on PATH.
func (m PhpModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("php")
}

// Install installs PHP via Homebrew: `brew install php`.
func (m PhpModule) Install(ctx types.InstallContext) error {
	if err := phpRunBrew(ctx.Cancel, ctx.Log, "install", "php"); err != nil {
		return fmt.Errorf("brew install php: %w", err)
	}
	return nil
}

// AuditInfo returns the installed PHP version.
func (m PhpModule) AuditInfo() string {
	return runner.CaptureOutput("php", "--version")
}
