package modules

import (
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

// FnmNodeModule combines fnm installation (M-08) and Node 24 installation (M-09)
// into a single module. It delegates to a runtime.RuntimeManager so the
// orchestration logic is independent of the concrete manager implementation.
type FnmNodeModule struct {
	manager runtime.RuntimeManager
}

// Compile-time interface check.
var _ types.Module = FnmNodeModule{}

// FnmNode is the exported package-level instance used by the catalogue.
var FnmNode types.Module = FnmNodeModule{manager: runtime.FnmManager{}}

// NewFnmNodeModule creates a new FnmNodeModule with the given manager.
// Used by tests to inject a mock RuntimeManager.
func NewFnmNodeModule(manager runtime.RuntimeManager) FnmNodeModule {
	return FnmNodeModule{manager: manager}
}

// ID returns the stable module identifier M-08 (fnm).
func (m FnmNodeModule) ID() types.ModuleID { return types.ModFnm }

// Name returns a human-readable module name.
func (m FnmNodeModule) Name() string { return "Fnm" }

// Criticality returns NonCritical — fnm/Node failure does not stop the pipeline.
func (m FnmNodeModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency (fnm is installed via brew).
func (m FnmNodeModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled delegates to the runtime manager's idempotence check.
func (m FnmNodeModule) IsInstalled(_ platform.Platform) bool {
	return m.manager.IsInstalled()
}

// Install installs fnm via Homebrew and then installs Node 24.
func (m FnmNodeModule) Install(ctx types.InstallContext) error {
	if err := m.manager.Install(ctx.Cancel, ctx.Log, ctx.Assets); err != nil {
		return err
	}
	return m.manager.InstallRuntime(ctx.Cancel, "24", ctx.Log)
}

// AuditInfo delegates to the runtime manager's version reporting.
func (m FnmNodeModule) AuditInfo() string {
	return m.manager.AuditInfo()
}
