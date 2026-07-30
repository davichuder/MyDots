package modules

import (
	"os"
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ZshModule implements the Zsh (M-02) installation module.
// Installs Zsh via Homebrew, adds it to /etc/shells on Linux, and changes the shell.
// Zsh is non-critical: if it fails, the install pipeline continues.
type ZshModule struct{}

// Compile-time interface check.
var _ types.Module = ZshModule{}

// Zsh is the exported package-level instance used by the catalogue.
var Zsh types.Module = ZshModule{}

// Injectable file read for /etc/shells mocking (like readProcVersion in detect.go).
var readShellsFile = os.ReadFile

// ID returns the stable module identifier M-02.
func (m ZshModule) ID() types.ModuleID { return types.ModZsh }

// Name returns a human-readable module name.
func (m ZshModule) Name() string { return "Zsh" }

// Criticality returns NonCritical — Zsh failure should not stop the pipeline.
func (m ZshModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency (Zsh is installed via brew).
func (m ZshModule) Dependencies() []types.ModuleID { return []types.ModuleID{types.ModHomebrew} }

// IsInstalled checks whether zsh is reachable on PATH.
func (m ZshModule) IsInstalled(_ platform.Platform) bool {
	return runner.CommandExists("zsh")
}

// Install installs Zsh via Homebrew, adds it to /etc/shells on Linux Native,
// and changes the default shell for the current user.
func (m ZshModule) Install(ctx types.InstallContext) error {
	// Step 1: Install zsh via Homebrew.
	if err := runner.BrewAt(ctx.Cancel, ctx.Log, brewPath(ctx), "install", "zsh"); err != nil {
		return err
	}

	// Step 2: Get the Homebrew prefix to construct the zsh binary path.
	prefix, err := runner.CaptureOutputErrorAt(brewPath(ctx), "--prefix")
	if err != nil {
		return err
	}
	zshPath := prefix + "/bin/zsh"

	// Step 3: On Linux Native (not Darwin, not WSL2), ensure zsh is in /etc/shells.
	if ctx.Platform.OS == platform.Linux && ctx.Platform.Variant == platform.Native {
		data, err := readShellsFile("/etc/shells")
		if err != nil {
			return err
		}

		found := false
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == zshPath {
				found = true
				break
			}
		}

		if !found {
			// Positional parameters keep the shell path and target path out of the
			// shell program, avoiding expansion or quoting hazards under sudo.
			if err := runner.Run(ctx.Cancel, ctx.Log, "sudo", "sh", "-c", `printf '%s\n' "$1" >> "$2"`, "sh", zshPath, "/etc/shells"); err != nil {
				return err
			}
		}
	}

	// Step 4: Change the default shell to zsh for the current user.
	if err := runner.Run(ctx.Cancel, ctx.Log, "chsh", "-s", zshPath, os.Getenv("USER")); err != nil {
		return err
	}

	return nil
}

// AuditInfo returns the trimmed output of "zsh --version".
func (m ZshModule) AuditInfo() string {
	return runner.CaptureOutput("zsh", "--version")
}
