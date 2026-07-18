package installer

import (
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/platform"
)

// BrewModule is a simple module that installs a Homebrew formula.
// All ~33 simple brew tools are declared as data, not code — see catalogue.go.
type BrewModule struct {
	id           ModuleID
	name         string
	formula      string   // brew install <formula>
	checkCommand string   // which <checkCommand> to verify install
	deps         []ModuleID
}

func (m BrewModule) ID() ModuleID                     { return m.id }
func (m BrewModule) Name() string                     { return m.name }
func (m BrewModule) Criticality() Criticality          { return NonCritical }
func (m BrewModule) Dependencies() []ModuleID          { return m.deps }

func (m BrewModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists(m.checkCommand)
}

func (m BrewModule) Install(ctx InstallContext) error {
	return runner.Brew(ctx.Cancel, ctx.Log, "install", m.formula)
}

func (m BrewModule) AuditInfo() string {
	return runner.CaptureOutput(m.checkCommand, "--version")
}
