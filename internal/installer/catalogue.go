package installer

// allModules returns all 48 modules in the canonical execution order
// defined in specs.md §5. This is a statically ordered list — not a
// runtime topological sort. Dependencies() is used by the executor
// for failure propagation only, not for ordering.
func allModules() []Module {
	return []Module{
		// Exec #1  — M-01 Homebrew (CRITICAL)
		Homebrew,
		// Exec #2  — M-48 Ghostty
		Ghostty,
		// Exec #3  — M-47 Nerd Font
		NerdFont,
		// Exec #4  — M-05 Git
		Git,
		// Exec #5  — M-06 git-credential-oauth
		GitCredOAuth,
		// Exec #6  — M-45 chezmoi
		Chezmoi,
		// Exec #7  — M-02 Zsh
		Zsh,
		// Exec #8  — M-03 Oh My Zsh
		OhMyZsh,
		// Exec #9  — M-04 Zellij (simple brew)
		BrewModule{id: ModZellij, name: "Zellij", formula: "zellij", checkCommand: "zellij", deps: []ModuleID{ModHomebrew}},
		// Exec #10 — M-08 fnm
		Fnm,
		// Exec #11 — M-09 Node 24
		Node,
		// Exec #12 — M-10 uv
		Uv,
		// Exec #13 — M-11 Python 3.12
		Python,
		// Exec #14 — M-12 Go (simple brew)
		BrewModule{id: ModGo, name: "Go", formula: "go", checkCommand: "go", deps: []ModuleID{ModHomebrew}},
		// Exec #15 — M-13 C/C++ toolchain
		CppToolchain,
		// Exec #16 — M-14 sdkman (optional)
		Sdkman,
		// Exec #17 — M-15 Java 25 (optional)
		Java,
		// Exec #18 — M-16 PHP (optional)
		Php,
		// Exec #19 — M-17 Neovim
		Neovim,
		// Exec #20 — M-18 Nvim personal config (optional)
		NeovimPersonal,
		// Exec #21 — M-19 Nvim framework (optional)
		NeovimFramework,
		// Exec #22 — M-46 Theme
		Theme,
		// Exec #23 — M-20 Atuin (simple brew)
		BrewModule{id: ModAtuin, name: "Atuin", formula: "atuin", checkCommand: "atuin", deps: []ModuleID{ModHomebrew}},
		// Exec #24 — M-21 zoxide (simple brew)
		BrewModule{id: ModZoxide, name: "zoxide", formula: "zoxide", checkCommand: "zoxide", deps: []ModuleID{ModHomebrew}},
		// Exec #25 — M-22 bat (simple brew)
		BrewModule{id: ModBat, name: "bat", formula: "bat", checkCommand: "bat", deps: []ModuleID{ModHomebrew}},
		// Exec #26 — M-23 eza (simple brew)
		BrewModule{id: ModEza, name: "eza", formula: "eza", checkCommand: "eza", deps: []ModuleID{ModHomebrew}},
		// Exec #27 — M-24 fd (simple brew)
		BrewModule{id: ModFd, name: "fd", formula: "fd", checkCommand: "fd", deps: []ModuleID{ModHomebrew}},
		// Exec #28 — M-25 ripgrep (simple brew)
		BrewModule{id: ModRipgrep, name: "ripgrep", formula: "ripgrep", checkCommand: "rg", deps: []ModuleID{ModHomebrew}},
		// Exec #29 — M-26 fzf (simple brew)
		BrewModule{id: ModFzf, name: "fzf", formula: "fzf", checkCommand: "fzf", deps: []ModuleID{ModHomebrew}},
		// Exec #30 — M-27 sd (simple brew)
		BrewModule{id: ModSd, name: "sd", formula: "sd", checkCommand: "sd", deps: []ModuleID{ModHomebrew}},
		// Exec #31 — M-28 jq (simple brew)
		BrewModule{id: ModJq, name: "jq", formula: "jq", checkCommand: "jq", deps: []ModuleID{ModHomebrew}},
		// Exec #32 — M-29 yq (simple brew)
		BrewModule{id: ModYq, name: "yq", formula: "yq", checkCommand: "yq", deps: []ModuleID{ModHomebrew}},
		// Exec #33 — M-30 tldr (simple brew)
		BrewModule{id: ModTldr, name: "tldr", formula: "tldr", checkCommand: "tldr", deps: []ModuleID{ModHomebrew}},
		// Exec #34 — M-31 delta (simple brew)
		BrewModule{id: ModDelta, name: "delta", formula: "git-delta", checkCommand: "delta", deps: []ModuleID{ModGit}},
		// Exec #35 — M-32 bottom (simple brew)
		BrewModule{id: ModBottom, name: "bottom", formula: "bottom", checkCommand: "btm", deps: []ModuleID{ModHomebrew}},
		// Exec #36 — M-33 thefuck (simple brew)
		BrewModule{id: ModThefuck, name: "thefuck", formula: "thefuck", checkCommand: "thefuck", deps: []ModuleID{ModPython}},
		// Exec #37 — M-34 carapace (simple brew)
		BrewModule{id: ModCarapace, name: "carapace", formula: "carapace", checkCommand: "carapace", deps: []ModuleID{ModHomebrew}},
		// Exec #38 — M-35 glow (simple brew)
		BrewModule{id: ModGlow, name: "glow", formula: "glow", checkCommand: "glow", deps: []ModuleID{ModHomebrew}},
		// Exec #39 — M-36 gh (simple brew)
		BrewModule{id: ModGh, name: "gh", formula: "gh", checkCommand: "gh", deps: []ModuleID{ModHomebrew}},
		// Exec #40 — M-37 Clipboard
		Clipboard,
		// Exec #41 — M-38 Docker
		Docker,
		// Exec #42 — M-39 lazydocker (simple brew)
		BrewModule{id: ModLazydocker, name: "lazydocker", formula: "lazydocker", checkCommand: "lazydocker", deps: []ModuleID{ModDocker}},
		// Exec #43 — M-07 lazygit (simple brew)
		BrewModule{id: ModLazygit, name: "lazygit", formula: "lazygit", checkCommand: "lazygit", deps: []ModuleID{ModGit}},
		// Exec #44 — M-40 opencode (simple brew-tap)
		BrewModule{id: ModOpencode, name: "Opencode", formula: "opencode/cli/opencode", checkCommand: "opencode", deps: []ModuleID{ModHomebrew}},
		// Exec #45 — M-44 MCP config
		McpConfig,
		// Exec #46 — M-41 RTK
		Rtk,
		// Exec #47 — M-42 caveman
		Caveman,
		// Exec #48 — M-43 gentle-ai
		GentleAi,
	}
}
