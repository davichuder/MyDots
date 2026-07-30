package modules

import (
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// CavemanModule (M-42)
// ---------------------------------------------------------------------------

// CavemanModule implements the caveman CLI installation module (M-42).
// It runs the caveman-install.sh curl-script wrapper with --only openclaw.
type CavemanModule struct{}

// Compile-time interface check.
var _ types.Module = CavemanModule{}

// Caveman is the exported package-level instance used by the catalogue.
var Caveman types.Module = CavemanModule{}

// ID returns the stable module identifier M-42.
func (m CavemanModule) ID() types.ModuleID { return types.ModCaveman }

// Name returns a human-readable module name.
func (m CavemanModule) Name() string { return "Caveman" }

// Criticality returns NonCritical — caveman failure does not stop the pipeline.
func (m CavemanModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns nil — caveman installs via curl-script, not brew.
func (m CavemanModule) Dependencies() []types.ModuleID { return nil }

// IsInstalled checks whether caveman is on PATH.
func (m CavemanModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("caveman")
}

// Install runs the embedded caveman-install.sh script via runner.Script.
func (m CavemanModule) Install(ctx types.InstallContext) error {
	return runner.Script(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/caveman-install.sh", nil)
}

// AuditInfo returns the installed caveman version.
func (m CavemanModule) AuditInfo() string {
	return runner.CaptureOutput("caveman", "--version")
}
