package modules

import (
	"fmt"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/platform"
)

// DockerModule implements the Docker (M-38) installation module.
// On Darwin it uses brew cask. On Linux it runs the embedded docker-linux.sh script.
// On WSL2 it checks systemd first; if not running, the module is skipped.
type DockerModule struct{}

// Compile-time interface check.
var _ types.Module = DockerModule{}

// Docker is the exported package-level instance used by the catalogue.
var Docker types.Module = DockerModule{}

// ID returns the stable module identifier M-38.
func (m DockerModule) ID() types.ModuleID { return types.ModDocker }

// Name returns a human-readable module name.
func (m DockerModule) Name() string { return "Docker" }

// Criticality returns NonCritical — Docker failure does not stop the pipeline.
func (m DockerModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency.
func (m DockerModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether docker is on PATH.
func (m DockerModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("docker")
}

// Install installs Docker based on platform:
//   - Darwin: brew install --cask docker-desktop
//   - Ubuntu native: run docker-linux.sh script
//   - WSL2 + systemd running: run docker-linux.sh script
//   - WSL2 + systemd not running: log warning, return error
func (m DockerModule) Install(ctx types.InstallContext) error {
	// Darwin: brew cask.
	if ctx.Platform.OS == platform.Darwin {
		return runner.BrewCask(ctx.Cancel, ctx.Log, ctx.Platform, "docker-desktop")
	}

	// Linux — check systemd on WSL2 first.
	if ctx.Platform.Variant == platform.WSL2 {
		if err := runner.Run(ctx.Cancel, ctx.Log, "systemctl", "is-system-running"); err != nil {
			fmt.Fprintln(ctx.Log, "[WARN] systemd is not running on WSL2 — Docker cannot be installed.")
			fmt.Fprintln(ctx.Log, "[WARN] Enable systemd in your WSL2 distro, or use Docker Desktop for Windows.")
			return fmt.Errorf("systemd not running on WSL2 — Docker skipped")
		}
	}

	// Run the embedded Docker installation script.
	err := runner.Script(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/docker-linux.sh", nil)
	if err != nil {
		exitCode := 1
		_, _ = fmt.Sscanf(err.Error(), "exit status %d", &exitCode)
		return myerr.AptInstallError{
			Command:  "assets/scripts/docker-linux.sh",
			ExitCode: exitCode,
			Stderr:   err.Error(),
		}
	}
	return nil
}

// AuditInfo returns the trimmed output of "docker --version".
func (m DockerModule) AuditInfo() string {
	return runner.CaptureOutput("docker", "--version")
}
