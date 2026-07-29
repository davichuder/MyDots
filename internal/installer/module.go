// Package installer defines the module interface, types, and constants
// used by the install pipeline. Every module in the catalogue implements
// the Module interface.
//
// Core types (Module, ModuleID, Criticality, InstallContext) are defined in
// the types sub-package and re-exported here via type aliases so that
// existing code referencing installer.Module, installer.ModuleID, etc.
// continues to work without import changes.
package installer

import "github.com/davichuder/MyDots/internal/installer/types"

// ModuleID uniquely identifies a module in the catalogue.
type ModuleID = types.ModuleID

const (
	ModHomebrew        ModuleID = types.ModHomebrew
	ModZsh             ModuleID = types.ModZsh
	ModOhMyZsh         ModuleID = types.ModOhMyZsh
	ModZellij          ModuleID = types.ModZellij
	ModGit             ModuleID = types.ModGit
	ModGitCredOAuth    ModuleID = types.ModGitCredOAuth
	ModLazygit         ModuleID = types.ModLazygit
	ModFnm             ModuleID = types.ModFnm
	ModNode            ModuleID = types.ModNode
	ModUv              ModuleID = types.ModUv
	ModPython          ModuleID = types.ModPython
	ModGo              ModuleID = types.ModGo
	ModCppToolchain    ModuleID = types.ModCppToolchain
	ModSdkman          ModuleID = types.ModSdkman
	ModJava            ModuleID = types.ModJava
	ModPhp             ModuleID = types.ModPhp
	ModNeovim          ModuleID = types.ModNeovim
	ModNeovimPersonal  ModuleID = types.ModNeovimPersonal
	ModNeovimFramework ModuleID = types.ModNeovimFramework
	ModAtuin           ModuleID = types.ModAtuin
	ModZoxide          ModuleID = types.ModZoxide
	ModBat             ModuleID = types.ModBat
	ModEza             ModuleID = types.ModEza
	ModFd              ModuleID = types.ModFd
	ModRipgrep         ModuleID = types.ModRipgrep
	ModFzf             ModuleID = types.ModFzf
	ModSd              ModuleID = types.ModSd
	ModJq              ModuleID = types.ModJq
	ModYq              ModuleID = types.ModYq
	ModTldr            ModuleID = types.ModTldr
	ModDelta           ModuleID = types.ModDelta
	ModBottom          ModuleID = types.ModBottom
	ModThefuck         ModuleID = types.ModThefuck
	ModCarapace        ModuleID = types.ModCarapace
	ModGlow            ModuleID = types.ModGlow
	ModGh              ModuleID = types.ModGh
	ModClipboard       ModuleID = types.ModClipboard
	ModDocker          ModuleID = types.ModDocker
	ModLazydocker      ModuleID = types.ModLazydocker
	ModOpencode        ModuleID = types.ModOpencode
	ModRtk             ModuleID = types.ModRtk
	ModCaveman         ModuleID = types.ModCaveman
	ModGentleAi        ModuleID = types.ModGentleAi
	ModMcpConfig       ModuleID = types.ModMcpConfig
	ModChezmoi         ModuleID = types.ModChezmoi
	ModTheme           ModuleID = types.ModTheme
	ModNerdFont        ModuleID = types.ModNerdFont
	ModGhostty         ModuleID = types.ModGhostty
)

// Criticality indicates whether a module failure should stop the entire
// installation (Critical) or be logged and continue (NonCritical).
type Criticality = types.Criticality

const (
	Critical    Criticality = types.Critical
	NonCritical Criticality = types.NonCritical
)

// InstallStatus records the outcome of a module's installation attempt.
type InstallStatus = types.InstallStatus

const (
	StatusInstalled               InstallStatus = types.StatusInstalled
	StatusSkipped                 InstallStatus = types.StatusSkipped
	StatusSkippedDisabled         InstallStatus = types.StatusSkippedDisabled
	StatusSkippedDependencyFailed InstallStatus = types.StatusSkippedDependencyFailed
	StatusSkippedNoWayland        InstallStatus = types.StatusSkippedNoWayland
	StatusFailed                  InstallStatus = types.StatusFailed
)

// Module is the interface every installable module must implement.
type Module = types.Module

// ConfiguredStateModule is the optional configuration-aware idempotence contract.
type ConfiguredStateModule = types.ConfiguredStateModule

// InstallContext bundles shared state for an entire install session.
type InstallContext = types.InstallContext
