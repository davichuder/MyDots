package modules

import (
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

// UvPythonModule combines uv installation (M-10) and Python 3.12 installation
// (M-11) into a single module. It delegates to a runtime.RuntimeManager so the
// orchestration logic is independent of the concrete manager implementation.
type UvPythonModule struct {
	manager runtime.RuntimeManager
}

// Compile-time interface check.
var _ types.Module = UvPythonModule{}

// UvPython is the exported package-level instance used by the catalogue.
var UvPython types.Module = UvPythonModule{manager: runtime.UvManager{}}

// NewUvPythonModule creates a new UvPythonModule with the given manager.
// Used by tests to inject a mock RuntimeManager.
func NewUvPythonModule(manager runtime.RuntimeManager) UvPythonModule {
	return UvPythonModule{manager: manager}
}

// ID returns the stable module identifier M-10 (uv).
func (m UvPythonModule) ID() types.ModuleID { return types.ModUv }

// Name returns a human-readable module name.
func (m UvPythonModule) Name() string { return "Uv" }

// Criticality returns NonCritical — uv/Python failure does not stop the pipeline.
func (m UvPythonModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency (uv is installed via brew).
func (m UvPythonModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled delegates to the runtime manager's idempotence check.
func (m UvPythonModule) IsInstalled(_ platform.Platform) bool {
	return m.manager.IsInstalled()
}

// Install installs uv via Homebrew and then installs Python 3.12.
func (m UvPythonModule) Install(ctx types.InstallContext) error {
	if err := m.manager.Install(ctx.Cancel, ctx.Log, ctx.Assets); err != nil {
		return err
	}
	return m.manager.InstallRuntime(ctx.Cancel, "3.12", ctx.Log)
}

// AuditInfo delegates to the runtime manager's version reporting.
func (m UvPythonModule) AuditInfo() string {
	return m.manager.AuditInfo()
}
