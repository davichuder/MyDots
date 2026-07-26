// Package types defines the core types and interfaces for the install pipeline:
// ModuleID, Criticality, InstallStatus, the Module interface, and InstallContext.
//
// This package exists to break the import cycle between internal/installer and
// internal/installer/modules. Both packages import types rather than each other.
package types

import (
	"context"
	"io"
	"io/fs"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

// ModuleID uniquely identifies a module in the catalogue.
// Each value is a stable identifier following the M-XX pattern.
type ModuleID string

const (
	ModHomebrew        ModuleID = "M-01"
	ModZsh             ModuleID = "M-02"
	ModOhMyZsh         ModuleID = "M-03"
	ModZellij          ModuleID = "M-04"
	ModGit             ModuleID = "M-05"
	ModGitCredOAuth    ModuleID = "M-06"
	ModLazygit         ModuleID = "M-07"
	ModFnm             ModuleID = "M-08"
	ModNode            ModuleID = "M-09"
	ModUv              ModuleID = "M-10"
	ModPython          ModuleID = "M-11"
	ModGo              ModuleID = "M-12"
	ModCppToolchain    ModuleID = "M-13"
	ModSdkman          ModuleID = "M-14"
	ModJava            ModuleID = "M-15"
	ModPhp             ModuleID = "M-16"
	ModNeovim          ModuleID = "M-17"
	ModNeovimPersonal  ModuleID = "M-18"
	ModNeovimFramework ModuleID = "M-19"
	ModAtuin           ModuleID = "M-20"
	ModZoxide          ModuleID = "M-21"
	ModBat             ModuleID = "M-22"
	ModEza             ModuleID = "M-23"
	ModFd              ModuleID = "M-24"
	ModRipgrep         ModuleID = "M-25"
	ModFzf             ModuleID = "M-26"
	ModSd              ModuleID = "M-27"
	ModJq              ModuleID = "M-28"
	ModYq              ModuleID = "M-29"
	ModTldr            ModuleID = "M-30"
	ModDelta           ModuleID = "M-31"
	ModBottom          ModuleID = "M-32"
	ModThefuck         ModuleID = "M-33"
	ModCarapace        ModuleID = "M-34"
	ModGlow            ModuleID = "M-35"
	ModGh              ModuleID = "M-36"
	ModClipboard       ModuleID = "M-37"
	ModDocker          ModuleID = "M-38"
	ModLazydocker      ModuleID = "M-39"
	ModOpencode        ModuleID = "M-40"
	ModRtk             ModuleID = "M-41"
	ModCaveman         ModuleID = "M-42"
	ModGentleAi        ModuleID = "M-43"
	ModMcpConfig       ModuleID = "M-44"
	ModChezmoi         ModuleID = "M-45"
	ModTheme           ModuleID = "M-46"
	ModNerdFont        ModuleID = "M-47"
	ModGhostty         ModuleID = "M-48"
)

// Criticality indicates whether a module failure should stop the entire
// installation (Critical) or be logged and continue (NonCritical).
type Criticality string

const (
	Critical    Criticality = "critical"
	NonCritical Criticality = "non-critical"
)

// InstallStatus records the outcome of a module's installation attempt.
type InstallStatus string

const (
	StatusInstalled               InstallStatus = "installed"
	StatusSkipped                 InstallStatus = "skipped"
	StatusSkippedDisabled         InstallStatus = "skipped-disabled"
	StatusSkippedDependencyFailed InstallStatus = "skipped-dependency-failed"
	StatusSkippedNoWayland        InstallStatus = "skipped-no-wayland"
	StatusFailed                  InstallStatus = "failed"
)

// Module is the interface every installable module must implement.
//
// Each module provides its identity, criticality, dependency graph,
// an idempotence check (IsInstalled), the install operation itself,
// and a string describing what version or value was installed.
type Module interface {
	// ID returns the stable module identifier (e.g. "M-01").
	ID() ModuleID

	// Name returns a human-readable module name (e.g. "Homebrew").
	Name() string

	// Criticality indicates whether a failure here stops the pipeline.
	Criticality() Criticality

	// Dependencies returns the ModuleIDs this module requires.
	Dependencies() []ModuleID

	// IsInstalled checks whether the tool is already present on the system.
	IsInstalled(p platform.Platform) bool

	// Install performs the module installation.
	Install(ctx InstallContext) error

	// AuditInfo returns a string representing the installed version,
	// configuration name, or other identifying information for the audit log.
	AuditInfo() string
}

// InstallContext bundles shared state for an entire install session.
// It is passed to every Module.Install() call so that all modules
// share the same platform, config, timestamp, log writer, cancellation
// context, and embedded asset filesystem.
type InstallContext struct {
	// Platform is the detected OS and variant for this session.
	Platform platform.Platform

	// Config is the user's configuration loaded from mydots-config.json.
	Config config.Config

	// SessionTimestamp is generated once at session start and shared
	// across all backup operations (colons replaced with dashes).
	SessionTimestamp string

	// Log receives stdout/stderr output from the runner for TUI display.
	Log io.Writer

	// Cancel carries the context used by exec.CommandContext.
	// Cancelling this context kills the running subprocess.
	Cancel context.Context

	// Assets provides access to embedded files (shell scripts, cheatsheets, etc.)
	// used by complex modules during installation.
	// In production this is set to the main package's embed.FS; in tests it can be
	// replaced with a fstest.MapFS.
	Assets fs.FS
}
