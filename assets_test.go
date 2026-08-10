package main

import (
	"io/fs"
	"net/url"
	"sort"
	"strings"
	"testing"
)

var expectedCheatsheetFiles = []string{
	"atuin.md", "bat.md", "bottom.md", "c-cpp-toolchain.md", "carapace.md", "caveman.md", "chezmoi.md", "clipboard.md",
	"delta.md", "docker.md", "eza.md", "fd.md", "fnm.md", "fzf.md", "gentle-ai.md", "gh.md", "ghostty.md", "git-credential-oauth.md",
	"git.md", "glow.md", "go.md", "homebrew.md", "java.md", "jq.md", "lazydocker.md", "lazygit.md", "mcp-servers.md", "neovim-framework.md",
	"neovim-personal.md", "neovim.md", "nerd-font.md", "node.md", "oh-my-zsh.md", "opencode.md", "php.md", "python.md", "ripgrep.md",
	"rtk.md", "sd.md", "sdkman.md", "thefuck.md", "theme.md", "tldr.md", "uv.md", "yq.md", "zellij.md", "zoxide.md", "zsh.md",
}

func TestEmbeddedCheatsheetInventoryIsExactAndReadable(t *testing.T) {
	entries, err := fs.ReadDir(assets, "assets/cheatsheets")
	if err != nil {
		t.Fatalf("ReadDir(assets/cheatsheets): %v", err)
	}

	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := "assets/cheatsheets/" + entry.Name()
		if _, err := fs.ReadFile(assets, path); err != nil {
			t.Errorf("ReadFile(%s): %v", path, err)
		}
		got = append(got, entry.Name())
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(expectedCheatsheetFiles, ",") {
		t.Errorf("embedded cheatsheet filenames = %v, want %v", got, expectedCheatsheetFiles)
	}
}

func TestEmbeddedCheatsheetsHaveUsefulSCR07Content(t *testing.T) {
	for _, name := range expectedCheatsheetFiles {
		t.Run(name, func(t *testing.T) {
			content, err := fs.ReadFile(assets, "assets/cheatsheets/"+name)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", name, err)
			}
			sections := scr07Sections(t, string(content))
			if !hasHTTPLink(sections["## Links"]) {
				t.Errorf("Links = %q, want a concrete HTTP(S) URL", sections["## Links"])
			}
			if !hasConcreteBullet(sections["## Key Shortcuts"]) {
				t.Errorf("Key Shortcuts = %q, want a concrete action/key bullet", sections["## Key Shortcuts"])
			}
			if !hasCommandExample(sections["## Usage Examples"]) {
				t.Errorf("Usage Examples = %q, want a command or code example", sections["## Usage Examples"])
			}
			for heading, section := range sections {
				if containsPlaceholder(section) {
					t.Errorf("%s = %q, must not contain placeholders or filler", heading, section)
				}
			}
		})
	}
}

func TestCorrectedCheatsheetContentMatchesInstalledTools(t *testing.T) {
	required := map[string][]string{
		"git-credential-oauth.md": {"https://github.com/hickford/git-credential-oauth", "git credential-oauth configure"},
		"caveman.md":              {"https://github.com/JuliusBrussee/caveman", "--only openclaw"},
		"gentle-ai.md":            {"https://github.com/Gentleman-Programming/gentle-ai"},
		"clipboard.md":            {"pbcopy", "pbpaste", "wl-copy", "wl-paste", "xclip"},
		"neovim-personal.md":      {"~/.config/nvim", "nvim"},
		"neovim-framework.md":     {"lazyvim", "lunarvim", "astronvim", "nvchad", "launchnvim"},
		"theme.md":                {"Neovim", "Zsh", "Zellij", "Ghostty"},
		"fd.md":                   {"fd '\\.go$' internal"},
	}
	for name, snippets := range required {
		content, err := fs.ReadFile(assets, "assets/cheatsheets/"+name)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", name, err)
		}
		for _, snippet := range snippets {
			if !strings.Contains(string(content), snippet) {
				t.Errorf("%s must contain %q", name, snippet)
			}
		}
	}
	for name, forbidden := range map[string]string{
		"caveman.md":         "caveman init",
		"neovim-personal.md": "nvim --clean",
		"theme.md":           "bat --",
	} {
		content, err := fs.ReadFile(assets, "assets/cheatsheets/"+name)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", name, err)
		}
		if strings.Contains(string(content), forbidden) {
			t.Errorf("%s must not contain %q", name, forbidden)
		}
	}
}

func scr07Sections(t *testing.T, content string) map[string]string {
	t.Helper()
	headings := []string{"## Links", "## Key Shortcuts", "## Usage Examples"}
	positions := make([]int, len(headings))
	for index, heading := range headings {
		positions[index] = strings.Index(content, heading)
		if positions[index] < 0 {
			t.Fatalf("missing %q", heading)
		}
		if index > 0 && positions[index] < positions[index-1] {
			t.Fatalf("headings are not ordered: %v", headings)
		}
	}
	sections := make(map[string]string, len(headings))
	for index, heading := range headings {
		end := len(content)
		if index+1 < len(headings) {
			end = positions[index+1]
		}
		sections[heading] = strings.TrimSpace(content[positions[index]+len(heading) : end])
	}
	return sections
}

func hasHTTPLink(section string) bool {
	for _, field := range strings.Fields(section) {
		parsed, err := url.Parse(strings.Trim(field, "()[]<>.,"))
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
			return true
		}
	}
	return false
}

func hasConcreteBullet(section string) bool {
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")) && len(strings.TrimSpace(line[2:])) > 3 {
			return true
		}
	}
	return false
}

func hasCommandExample(section string) bool {
	return strings.Contains(section, "`") || strings.Contains(section, "$ ")
}

func containsPlaceholder(section string) bool {
	lower := strings.ToLower(section)
	return strings.TrimSpace(section) == "" || strings.Contains(lower, "todo") || strings.Contains(lower, "tbd") || strings.Contains(lower, "lorem ipsum") || strings.Contains(lower, "placeholder")
}
