package installer

import (
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/platform"
)

// BrewModule is a Module that installs a single Homebrew formula.
//
// Simple brew modules are declared as data in catalogue.go — no individual
// file per tool. Each BrewModule value carries its identity, the formula
// to install, the command to verify installation, and its dependencies.
type BrewModule struct {
	id           ModuleID
	name         string
	formula      string // brew install <formula>
	checkCommand string // which binary to check for IsInstalled / AuditInfo
	deps         []ModuleID
}

// ID returns the stable module identifier (e.g. "M-20" for zoxide).
func (m BrewModule) ID() ModuleID { return m.id }

// Name returns the human-readable module name.
func (m BrewModule) Name() string { return m.name }

// Criticality always returns NonCritical — a single brew failure does not
// stop the entire install pipeline.
func (m BrewModule) Criticality() Criticality { return NonCritical }

// Dependencies returns the ModuleIDs this module depends on (typically [ModHomebrew]).
func (m BrewModule) Dependencies() []ModuleID { return m.deps }

// IsInstalled checks whether the configured checkCommand is reachable on PATH
// via runner.CommandExists.
func (m BrewModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists(m.checkCommand)
}

// Install runs "brew install <formula>" via runner.Brew, streaming output
// to the install context's Log writer.
func (m BrewModule) Install(ctx InstallContext) error {
	return runner.BrewAt(ctx.Cancel, ctx.Log, brewPath(ctx), "install", m.formula)
}

func brewPath(ctx InstallContext) string {
	if ctx.BrewPath == nil {
		return ""
	}
	return *ctx.BrewPath
}

// AuditInfo returns the trimmed output of "<checkCommand> --version".
// Returns an empty string if the command fails or is not found.
func (m BrewModule) AuditInfo() string {
	return runner.CaptureOutput(m.checkCommand, "--version")
}
