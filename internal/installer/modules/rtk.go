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

var rtkRunBrew = runner.Brew
var rtkRun = runner.Run

// ---------------------------------------------------------------------------
// RtkModule (M-41)
// ---------------------------------------------------------------------------

// RtkModule implements the rtk (React Toolkit) installation module (M-41).
// It installs rtk via Homebrew and runs `rtk init` to set up the config.
type RtkModule struct{}

// Compile-time interface check.
var _ types.Module = RtkModule{}

// Rtk is the exported package-level instance used by the catalogue.
var Rtk types.Module = RtkModule{}

// ID returns the stable module identifier M-41.
func (m RtkModule) ID() types.ModuleID { return types.ModRtk }

// Name returns a human-readable module name.
func (m RtkModule) Name() string { return "Rtk" }

// Criticality returns NonCritical — rtk failure does not stop the pipeline.
func (m RtkModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a requirement.
func (m RtkModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether rtk is on PATH.
func (m RtkModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("rtk")
}

// Install installs rtk via Homebrew and runs rtk init:
//
//	1. brew install rtk
//	2. rtk init
//
// If brew install fails, rtk init is not called.
func (m RtkModule) Install(ctx types.InstallContext) error {
	if err := rtkRunBrew(ctx.Cancel, ctx.Log, "install", "rtk"); err != nil {
		return fmt.Errorf("brew install rtk: %w", err)
	}
	return rtkRun(ctx.Cancel, ctx.Log, "rtk", "init")
}

// AuditInfo returns the installed rtk version.
func (m RtkModule) AuditInfo() string {
	return runner.CaptureOutput("rtk", "--version")
}
