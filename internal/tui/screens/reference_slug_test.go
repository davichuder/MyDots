package screens

import (
	"testing"

	"github.com/davichuder/MyDots/internal/installer"
)

func TestReferenceSlugMapsEveryModuleID(t *testing.T) {
	cases := map[installer.ModuleID]string{
		installer.ModHomebrew: "homebrew", installer.ModZsh: "zsh", installer.ModOhMyZsh: "oh-my-zsh", installer.ModZellij: "zellij",
		installer.ModGit: "git", installer.ModGitCredOAuth: "git-credential-oauth", installer.ModLazygit: "lazygit", installer.ModFnm: "fnm",
		installer.ModNode: "node", installer.ModUv: "uv", installer.ModPython: "python", installer.ModGo: "go", installer.ModCppToolchain: "c-cpp-toolchain",
		installer.ModSdkman: "sdkman", installer.ModJava: "java", installer.ModPhp: "php", installer.ModNeovim: "neovim", installer.ModNeovimPersonal: "neovim-personal",
		installer.ModNeovimFramework: "neovim-framework", installer.ModAtuin: "atuin", installer.ModZoxide: "zoxide", installer.ModBat: "bat", installer.ModEza: "eza",
		installer.ModFd: "fd", installer.ModRipgrep: "ripgrep", installer.ModFzf: "fzf", installer.ModSd: "sd", installer.ModJq: "jq", installer.ModYq: "yq",
		installer.ModTldr: "tldr", installer.ModDelta: "delta", installer.ModBottom: "bottom", installer.ModThefuck: "thefuck", installer.ModCarapace: "carapace",
		installer.ModGlow: "glow", installer.ModGh: "gh", installer.ModClipboard: "clipboard", installer.ModDocker: "docker", installer.ModLazydocker: "lazydocker",
		installer.ModOpencode: "opencode", installer.ModRtk: "rtk", installer.ModCaveman: "caveman", installer.ModGentleAi: "gentle-ai", installer.ModMcpConfig: "mcp-servers",
		installer.ModChezmoi: "chezmoi", installer.ModTheme: "theme", installer.ModNerdFont: "nerd-font", installer.ModGhostty: "ghostty",
	}
	if len(cases) != 48 {
		t.Fatalf("mapping case count = %d, want 48", len(cases))
	}
	for moduleID, want := range cases {
		if got := referenceSlug(moduleID); got != want {
			t.Errorf("referenceSlug(%q) = %q, want %q", moduleID, got, want)
		}
	}
	if got := referenceSlug("M-99"); got != "" {
		t.Errorf("referenceSlug(unknown) = %q, want empty", got)
	}
}
