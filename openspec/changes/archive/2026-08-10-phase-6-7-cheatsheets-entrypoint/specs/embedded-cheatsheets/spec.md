# Embedded Cheatsheets Specification

## Purpose

Define T-097 embedded inventory and SCR-07 content.

## Requirements

### Requirement: Exact Embedded Inventory

The embedded cheatsheet directory MUST contain exactly these 48 readable filenames:

`homebrew.md`, `zsh.md`, `oh-my-zsh.md`, `zellij.md`, `git.md`, `git-credential-oauth.md`, `lazygit.md`, `fnm.md`, `node.md`, `uv.md`, `python.md`, `go.md`, `c-cpp-toolchain.md`, `sdkman.md`, `java.md`, `php.md`, `neovim.md`, `neovim-personal.md`, `neovim-framework.md`, `atuin.md`, `zoxide.md`, `bat.md`, `eza.md`, `fd.md`, `ripgrep.md`, `fzf.md`, `sd.md`, `jq.md`, `yq.md`, `tldr.md`, `delta.md`, `bottom.md`, `thefuck.md`, `carapace.md`, `glow.md`, `gh.md`, `clipboard.md`, `docker.md`, `lazydocker.md`, `opencode.md`, `rtk.md`, `caveman.md`, `gentle-ai.md`, `mcp-servers.md`, `chezmoi.md`, `theme.md`, `nerd-font.md`, `ghostty.md`.

#### Scenario: Inventory is complete
- GIVEN the compiled embedded filesystem
- WHEN `assets/cheatsheets` is inspected and every required path is read
- THEN its Markdown filename set exactly matches the 48-name inventory

#### Scenario: Inventory drifts
- GIVEN a required file is missing, extra, misnamed, or unreadable
- WHEN the inventory contract is validated
- THEN validation MUST fail and identify the offending path

### Requirement: SCR-07 Useful Content

Every cheatsheet MUST present `## Links`, `## Key Shortcuts`, and `## Usage Examples` in that order. Each section MUST contain a concrete tool-relevant link, shortcut/action or command example and MUST NOT consist of template tokens, TODO/TBD text, or filler.

#### Scenario: Cheatsheet is useful
- GIVEN any required cheatsheet
- WHEN its readable Markdown is evaluated
- THEN all three headings and concrete section content are present

#### Scenario: Placeholder content is rejected
- GIVEN a section contains only placeholders or whitespace
- WHEN content is evaluated
- THEN the cheatsheet MUST fail the SCR-07 contract

### Requirement: T-097 RED to GREEN

T-097 MUST first demonstrate a failing exact-inventory/content test, then add the documents and pass the same test.

#### Scenario: T-097 evidence
- GIVEN the 48 documents are absent
- WHEN the inventory test runs RED and the documents are subsequently added
- THEN the unchanged test MUST pass GREEN without host side effects
