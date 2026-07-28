package modules

import (
	"context"
	"io"
	"regexp"
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

var node24VersionLine = regexp.MustCompile(`^\s*(?:\*\s*)?v24\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)(?:-(?:0|[1-9]\d*|[0-9A-Za-z-]+)(?:\.(?:0|[1-9]\d*|[0-9A-Za-z-]+))*)?(?:\+(?:[0-9A-Za-z-]+)(?:\.[0-9A-Za-z-]+)*)?(?:\s+\S.*)?\s*$`)

type FnmModule struct{ manager runtime.RuntimeManager }

var _ types.Module = FnmModule{}

var Fnm types.Module = NewFnmModule(runtime.FnmManager{})

func NewFnmModule(manager runtime.RuntimeManager) FnmModule { return FnmModule{manager: manager} }
func (m FnmModule) ID() types.ModuleID                      { return types.ModFnm }
func (m FnmModule) Name() string                            { return "Fnm" }
func (m FnmModule) Criticality() types.Criticality          { return types.NonCritical }
func (m FnmModule) Dependencies() []types.ModuleID          { return []types.ModuleID{types.ModHomebrew} }
func (m FnmModule) IsInstalled(_ platform.Platform) bool    { return m.manager.IsInstalled() }
func (m FnmModule) Install(ctx types.InstallContext) error {
	return m.manager.Install(ctx.Cancel, ctx.Log, ctx.Assets)
}
func (m FnmModule) AuditInfo() string { return m.manager.AuditInfo() }

type nodeRuntimeManager interface {
	runtime.RuntimeManager
	SetDefaultRuntime(context.Context, string, io.Writer) error
}

type NodeModule struct{ manager nodeRuntimeManager }

var _ types.Module = NodeModule{}

var Node types.Module = NewNodeModule(runtime.FnmManager{})

func NewNodeModule(manager nodeRuntimeManager) NodeModule { return NodeModule{manager: manager} }
func (m NodeModule) ID() types.ModuleID                   { return types.ModNode }
func (m NodeModule) Name() string                         { return "Node 24" }
func (m NodeModule) Criticality() types.Criticality       { return types.NonCritical }
func (m NodeModule) Dependencies() []types.ModuleID       { return []types.ModuleID{types.ModFnm} }
func (m NodeModule) IsInstalled(_ platform.Platform) bool {
	for _, line := range strings.Split(runner.CaptureOutput("fnm", "list"), "\n") {
		if node24VersionLine.MatchString(line) {
			return true
		}
	}
	return false
}
func (m NodeModule) Install(ctx types.InstallContext) error {
	if err := m.manager.InstallRuntime(ctx.Cancel, "24", ctx.Log); err != nil {
		return err
	}
	return m.manager.SetDefaultRuntime(ctx.Cancel, "24", ctx.Log)
}
func (m NodeModule) AuditInfo() string { return runner.CaptureOutput("node", "--version") }
