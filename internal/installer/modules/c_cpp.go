// Package modules contains complex module implementations that require
// OS-specific logic, multi-step installs, or config writes.
// Simple modules are declared as data in the catalogue package.
package modules

import (
	"fmt"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	myerr "github.com/davichuder/MyDots/internal/errors"
	"github.com/davichuder/MyDots/internal/platform"
)

// CppToolchainModule implements the C++ Toolchain (M-13) installation module.
// On Darwin it installs gcc, cmake, and llvm via Homebrew.
// On Linux (both Native and WSL2) it installs build-essential, gdb, and cmake via apt.
// C++ Toolchain is non-critical: if it fails, the install pipeline continues.
type CppToolchainModule struct{}

// Compile-time interface check.
var _ types.Module = CppToolchainModule{}

// CppToolchain is the exported package-level instance used by the catalogue.
var CppToolchain types.Module = CppToolchainModule{}

// ID returns the stable module identifier M-13.
func (m CppToolchainModule) ID() types.ModuleID { return types.ModCppToolchain }

// Name returns a human-readable module name.
func (m CppToolchainModule) Name() string { return "C++ Toolchain" }

// Criticality returns NonCritical — C++ Toolchain failure should not stop the pipeline.
func (m CppToolchainModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency (some tools are installed via brew on Darwin).
func (m CppToolchainModule) Dependencies() []types.ModuleID { return []types.ModuleID{types.ModHomebrew} }

// IsInstalled checks whether the C/C++ toolchain is available on PATH.
// On Darwin it checks for clangd; on Linux (Native and WSL2) it checks for gcc.
func (m CppToolchainModule) IsInstalled(p platform.Platform) bool {
	if p.OS == platform.Darwin {
		return runner.CommandExists("clangd")
	}
	return runner.CommandExists("gcc")
}

// Install installs the C/C++ toolchain.
// On Darwin: installs gcc, cmake, and llvm via Homebrew.
// On Linux (both Native and WSL2): installs build-essential, gdb, and cmake via apt.
func (m CppToolchainModule) Install(ctx types.InstallContext) error {
	if ctx.Platform.OS == platform.Darwin {
		return runner.Brew(ctx.Cancel, ctx.Log, "install", "gcc", "cmake", "llvm")
	}

	// Linux (both Native and WSL2) — install via apt.
	err := runner.Run(ctx.Cancel, ctx.Log, "sudo", "apt", "install", "-y", "build-essential", "gdb", "cmake")
	if err != nil {
		exitCode := 1
		_, _ = fmt.Sscanf(err.Error(), "exit status %d", &exitCode)
		return myerr.AptInstallError{
			Command:  "sudo apt install -y build-essential gdb cmake",
			ExitCode: exitCode,
			Stderr:   err.Error(),
		}
	}
	return nil
}

// AuditInfo returns the installed toolchain version.
// On Darwin it tries to get the clangd version via brew --prefix llvm.
// On Linux (or if brew prefix fails) it falls back to gcc --version.
func (m CppToolchainModule) AuditInfo() string {
	prefix := runner.CaptureOutput("brew", "--prefix", "llvm")
	if prefix != "" {
		return runner.CaptureOutput(prefix+"/bin/clangd", "--version")
	}
	return runner.CaptureOutput("gcc", "--version")
}
