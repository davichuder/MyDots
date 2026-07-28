package modules

import (
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// GentleAiModule (M-43)
// ---------------------------------------------------------------------------

// GentleAiModule implements the Gentle AI installation module (M-43).
// It taps the Gentle Programming Homebrew tap and installs gentle-ai.
type GentleAiModule struct{}

// Compile-time interface check.
var _ types.Module = GentleAiModule{}

// GentleAi is the exported package-level instance used by the catalogue.
var GentleAi types.Module = GentleAiModule{}

// ID returns the stable module identifier M-43.
func (m GentleAiModule) ID() types.ModuleID { return types.ModGentleAi }

// Name returns a human-readable module name.
func (m GentleAiModule) Name() string { return "Gentle AI" }

// Criticality returns NonCritical — gentle-ai failure does not stop the pipeline.
func (m GentleAiModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a requirement.
func (m GentleAiModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether gentle-ai is on PATH.
func (m GentleAiModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("gentle-ai")
}

// Install taps the Homebrew tap and installs gentle-ai:
//
//	1. brew tap Gentleman-Programming/homebrew-tap
//	2. brew install gentle-ai
//
// Tap failure returns an error before the install attempt (handled by runner.BrewTap).
func (m GentleAiModule) Install(ctx types.InstallContext) error {
	return runner.BrewTap(ctx.Cancel, ctx.Log, "Gentleman-Programming/homebrew-tap", "gentle-ai")
}

// AuditInfo returns the installed gentle-ai version.
func (m GentleAiModule) AuditInfo() string {
	return runner.CaptureOutput("gentle-ai", "--version")
}
