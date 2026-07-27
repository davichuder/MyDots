// Package modules contains complex module implementations that require
// OS-specific logic, multi-step installs, or config writes.
// Simple modules are declared as data in the catalogue package.
package modules

import (
	"fmt"
	"os"
	"os/user"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/platform"
)

// OhMyZshModule implements the Oh My Zsh (M-03) installation module.
// It installs Oh My Zsh via the official curl script embedded in assets/.
// Oh My Zsh is non-critical: if it fails, the install pipeline continues.
type OhMyZshModule struct{}

// Ensure OhMyZshModule satisfies the Module interface at compile time.
var _ types.Module = OhMyZshModule{}

// OhMyZsh is the exported package-level instance used by the catalogue.
var OhMyZsh types.Module = OhMyZshModule{}

// checkOhMyZshDir is injectable for testing. Checks whether ~/.oh-my-zsh
// exists on the filesystem using os.UserHomeDir via os/user.Current().
var checkOhMyZshDir = func() bool {
	usr, err := user.Current()
	if err != nil {
		return false
	}
	_, err = os.Stat(usr.HomeDir + "/.oh-my-zsh")
	return err == nil
}

// ID returns the stable module identifier M-03.
func (m OhMyZshModule) ID() types.ModuleID { return types.ModOhMyZsh }

// Name returns a human-readable module name.
func (m OhMyZshModule) Name() string { return "Oh My Zsh" }

// Criticality returns NonCritical — Oh My Zsh failure should not stop the pipeline.
func (m OhMyZshModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Zsh as a dependency (Oh My Zsh requires Zsh).
func (m OhMyZshModule) Dependencies() []types.ModuleID { return []types.ModuleID{types.ModZsh} }

// IsInstalled checks whether ~/.oh-my-zsh exists on the filesystem.
func (m OhMyZshModule) IsInstalled(_ platform.Platform) bool {
	return checkOhMyZshDir()
}

// Install runs the embedded OMZ install script via runner.Script.
// ctx.Assets provides the embedded filesystem (set by the main package).
// RUNZSH=no and CHSH=no prevent the script from changing the shell interactively.
// Returns a CurlScriptError if the script exits with a non-zero code.
func (m OhMyZshModule) Install(ctx types.InstallContext) error {
	err := runner.Script(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/omz-install.sh", map[string]string{
		"RUNZSH": "no",
		"CHSH":   "no",
	})
	if err != nil {
		return wrapOMZScriptError(err)
	}
	return nil
}

// AuditInfo returns the trimmed output of the Oh My Zsh version string.
func (m OhMyZshModule) AuditInfo() string {
	return runner.CaptureOutput("zsh", "-c", "source ~/.oh-my-zsh/oh-my-zsh.sh && echo $ZSH_VERSION")
}

// wrapOMZScriptError converts a runner.Script error into a CurlScriptError
// with What/Why/Fix context for TUI rendering.
func wrapOMZScriptError(err error) error {
	// Try to extract exit code from the error message.
	exitCode := 1
	_, _ = fmt.Sscanf(err.Error(), "exit status %d", &exitCode)

	return myerr.CurlScriptError{
		URL:      "assets/scripts/omz-install.sh",
		ExitCode: exitCode,
		Stderr:   err.Error(),
	}
}
