package installer

import (
	"github.com/davichuder/MyDots/internal/installer/modules"
	"github.com/davichuder/MyDots/internal/platform"
)

// stubModule is a temporary placeholder for complex module implementations
// (e.g. Homebrew, Zsh, Docker) that have their own files under modules/.
// Each stub is replaced with the real implementation when its phase is reached.
type stubModule struct {
	id       ModuleID
	name     string
	critical bool
}

func (s stubModule) ID() ModuleID                 { return s.id }
func (s stubModule) Name() string                  { return s.name }
func (s stubModule) Dependencies() []ModuleID      { return nil }
func (s stubModule) IsInstalled(platform.Platform) bool { return false }
func (s stubModule) Install(InstallContext) error  { return nil }
func (s stubModule) AuditInfo() string             { return "" }

func (s stubModule) Criticality() Criticality {
	if s.critical {
		return Critical
	}
	return NonCritical
}

// stub creates a non-critical stubModule.
func stub(id ModuleID, name string) Module {
	return stubModule{id: id, name: name}
}

// criticalStub creates a critical stubModule (only for Homebrew M-01).
func criticalStub(id ModuleID, name string) Module {
	return stubModule{id: id, name: name, critical: true}
}

// allModules returns all 48 modules in canonical execution order as defined
// in specs.md §5 and design.md §5.1. This is a statically ordered list — not
// a runtime topological sort. Dependencies() is used only by the executor
// for failure propagation, never for ordering.
//
// Complex modules (stubModule) are temporary placeholders replaced by real
// implementations from modules/ as each phase completes. Simple brew modules
// use BrewModule directly.
func allModules() []Module {
	return []Module{
		modules.Homebrew,                                      // M-01 — CRITICAL
		stub(ModGhostty, "Ghostty"),                           // M-48
		stub(ModNerdFont, "Nerd Font"),                        // M-47
		stub(ModGit, "Git"),                                   // M-05
		stub(ModGitCredOAuth, "Git Credential OAuth"),         // M-06
		stub(ModChezmoi, "Chezmoi"),                           // M-45
		stub(ModZsh, "Zsh"),                                   // M-02
		stub(ModOhMyZsh, "Oh My Zsh"),                         // M-03
		BrewModule{id: ModZellij, name: "Zellij", formula: "zellij", checkCommand: "zellij", deps: []ModuleID{ModHomebrew}},           // M-04
		stub(ModFnm, "Fnm"),                                   // M-08
		stub(ModNode, "Node"),                                 // M-09
		stub(ModUv, "Uv"),                                     // M-10
		stub(ModPython, "Python"),                             // M-11
		BrewModule{id: ModGo, name: "Go", formula: "go", checkCommand: "go", deps: []ModuleID{ModHomebrew}},                          // M-12
		stub(ModCppToolchain, "C++ Toolchain"),                // M-13
		stub(ModSdkman, "Sdkman"),                             // M-14
		stub(ModJava, "Java"),                                 // M-15
		stub(ModPhp, "PHP"),                                   // M-16
		stub(ModNeovim, "Neovim"),                             // M-17
		stub(ModNeovimPersonal, "Neovim Personal"),            // M-18
		stub(ModNeovimFramework, "Neovim Framework"),          // M-19
		stub(ModTheme, "Theme"),                               // M-46
		BrewModule{id: ModAtuin, name: "atuin", formula: "atuin", checkCommand: "atuin", deps: []ModuleID{ModHomebrew}},             // M-20
		BrewModule{id: ModZoxide, name: "zoxide", formula: "zoxide", checkCommand: "zoxide", deps: []ModuleID{ModHomebrew}},          // M-21
		BrewModule{id: ModBat, name: "bat", formula: "bat", checkCommand: "bat", deps: []ModuleID{ModHomebrew}},                      // M-22
		BrewModule{id: ModEza, name: "eza", formula: "eza", checkCommand: "eza", deps: []ModuleID{ModHomebrew}},                      // M-23
		BrewModule{id: ModFd, name: "fd", formula: "fd", checkCommand: "fd", deps: []ModuleID{ModHomebrew}},                          // M-24
		BrewModule{id: ModRipgrep, name: "ripgrep", formula: "ripgrep", checkCommand: "ripgrep", deps: []ModuleID{ModHomebrew}},       // M-25
		BrewModule{id: ModFzf, name: "fzf", formula: "fzf", checkCommand: "fzf", deps: []ModuleID{ModHomebrew}},                      // M-26
		BrewModule{id: ModSd, name: "sd", formula: "sd", checkCommand: "sd", deps: []ModuleID{ModHomebrew}},                          // M-27
		BrewModule{id: ModJq, name: "jq", formula: "jq", checkCommand: "jq", deps: []ModuleID{ModHomebrew}},                          // M-28
		BrewModule{id: ModYq, name: "yq", formula: "yq", checkCommand: "yq", deps: []ModuleID{ModHomebrew}},                          // M-29
		BrewModule{id: ModTldr, name: "tldr", formula: "tldr", checkCommand: "tldr", deps: []ModuleID{ModHomebrew}},                   // M-30
		BrewModule{id: ModDelta, name: "delta", formula: "git-delta", checkCommand: "delta", deps: []ModuleID{ModHomebrew}},           // M-31
		BrewModule{id: ModBottom, name: "bottom", formula: "bottom", checkCommand: "btm", deps: []ModuleID{ModHomebrew}},               // M-32
		BrewModule{id: ModThefuck, name: "thefuck", formula: "thefuck", checkCommand: "thefuck", deps: []ModuleID{ModHomebrew}},        // M-33
		BrewModule{id: ModCarapace, name: "carapace", formula: "carapace", checkCommand: "carapace", deps: []ModuleID{ModHomebrew}},    // M-34
		BrewModule{id: ModGlow, name: "glow", formula: "glow", checkCommand: "glow", deps: []ModuleID{ModHomebrew}},                   // M-35
		BrewModule{id: ModGh, name: "gh", formula: "gh", checkCommand: "gh", deps: []ModuleID{ModHomebrew}},                           // M-36
		stub(ModClipboard, "Clipboard"),                           // M-37
		stub(ModDocker, "Docker"),                                 // M-38
		BrewModule{id: ModLazydocker, name: "lazydocker", formula: "lazydocker", checkCommand: "lazydocker", deps: []ModuleID{ModHomebrew}}, // M-39
		BrewModule{id: ModLazygit, name: "lazygit", formula: "lazygit", checkCommand: "lazygit", deps: []ModuleID{ModHomebrew}},       // M-07
		BrewModule{id: ModOpencode, name: "opencode", formula: "opencode", checkCommand: "opencode", deps: []ModuleID{ModHomebrew}},   // M-40
		stub(ModMcpConfig, "MCP Config"),                          // M-44
		stub(ModRtk, "Rtk"),                                       // M-41
		stub(ModCaveman, "Caveman"),                               // M-42
		stub(ModGentleAi, "Gentle AI"),                            // M-43
	}
}
