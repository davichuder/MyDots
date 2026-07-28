package modules

import (
	"regexp"
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

var python312VersionLine = regexp.MustCompile(`^\s*(?:cpython-)?3\.12\.(?:0|[1-9]\d*)(?:[-+][0-9A-Za-z_.-]+)?(?:\s+\S.*)?\s*$`)

type UvModule struct{ manager runtime.RuntimeManager }

var _ types.Module = UvModule{}

var Uv types.Module = NewUvModule(runtime.UvManager{})

func NewUvModule(manager runtime.RuntimeManager) UvModule { return UvModule{manager: manager} }
func (m UvModule) ID() types.ModuleID                     { return types.ModUv }
func (m UvModule) Name() string                           { return "Uv" }
func (m UvModule) Criticality() types.Criticality         { return types.NonCritical }
func (m UvModule) Dependencies() []types.ModuleID         { return []types.ModuleID{types.ModHomebrew} }
func (m UvModule) IsInstalled(_ platform.Platform) bool   { return m.manager.IsInstalled() }
func (m UvModule) Install(ctx types.InstallContext) error {
	return m.manager.Install(ctx.Cancel, ctx.Log, ctx.Assets)
}
func (m UvModule) AuditInfo() string { return m.manager.AuditInfo() }

type PythonModule struct{ manager runtime.RuntimeManager }

var _ types.Module = PythonModule{}

var Python types.Module = NewPythonModule(runtime.UvManager{})

func NewPythonModule(manager runtime.RuntimeManager) PythonModule {
	return PythonModule{manager: manager}
}
func (m PythonModule) ID() types.ModuleID             { return types.ModPython }
func (m PythonModule) Name() string                   { return "Python 3.12" }
func (m PythonModule) Criticality() types.Criticality { return types.NonCritical }
func (m PythonModule) Dependencies() []types.ModuleID { return []types.ModuleID{types.ModUv} }
func (m PythonModule) IsInstalled(_ platform.Platform) bool {
	for _, line := range strings.Split(runner.CaptureOutput("uv", "python", "list", "--only-installed"), "\n") {
		if python312VersionLine.MatchString(line) {
			return true
		}
	}
	return false
}
func (m PythonModule) Install(ctx types.InstallContext) error {
	return m.manager.InstallRuntime(ctx.Cancel, "3.12", ctx.Log)
}
func (m PythonModule) AuditInfo() string {
	return runner.CaptureOutput("uv", "run", "python3.12", "--version")
}
