# MyDots — Proposal

> *"A development environment is a thinking tool. If it fights you, you think less."*

**Version:** 1.0.0-draft  
**Status:** Design — pending implementation  
**Author:** David  
**Date:** May 2026  
**Repo:** <https://github.com/davichuder/MyDots>

---

## 1. Why We're Doing This

Every time you set up a new machine — a work laptop, a WSL2 instance, a VPS — you lose hours.
Hours hunting down the right version of a tool, hours reconstructing shell aliases from memory,
hours figuring out why nvim looks broken on this machine but not the other.

The goal is to make that problem disappear permanently:
**one command, any machine, identical result, every time.**

This is not a dotfiles manager. This is a full development environment installer
with a first-class TUI, opinionated defaults, and zero tolerance for manual steps.

---

## 2. What MyDots Is

A Go binary (`mydots`) that runs a terminal UI built on Bubble Tea v2.
It installs, configures, and syncs a complete development environment on any supported platform.

```bash
brew tap davichuder/homebrew-tap
brew install mydots
mydots
```

Or, without the tap:

```bash
git clone https://github.com/davichuder/MyDots
go run main.go
```

### Supported platforms (v1.0.0)

| Platform     | Status      | Package manager        |
| ------------ | ----------- | ---------------------- |
| macOS Darwin | ✅ v1.0.0   | Homebrew               |
| Ubuntu/WSL2  | ✅ v1.0.0   | Homebrew + apt         |
| Arch Linux   | 🔵 v1.1.0   | Homebrew + pacman      |
| Debian       | 🔵 v1.1.0   | Homebrew + apt         |
| Fedora       | 🔵 v1.1.0   | Homebrew + dnf         |
| Windows      | ❌ Never    | Shows WSL2 setup guide |

> The OS is detected at runtime on every launch. It is never saved to config.

---

## 3. Non-Negotiable Principles

| Principle | Description |
| --------- | ----------- |
| **Idempotence** | Running `mydots` N times produces the same result. Already-installed tools are skipped. |
| **Fail-fast selective** | Critical modules stop everything on failure. Non-critical modules log and continue. |
| **Audit trail** | Every installed tool is recorded with version, method, and timestamp in `~/.mydots-audit.json`. |
| **Decisions first** | All user choices are collected in the Config menu before any installation begins. |
| **Brew-first** | Homebrew is the default installer. OS-native managers (apt, pacman, dnf) are used only when brew cannot or should not handle a tool. Each tool defines its own install strategy per OS with brew as the fallback. |
| **Backup-before-overwrite** | Any existing config is backed up before replacement. |
| **Zero Windows installs** | Nothing installs on Windows. A rendered markdown guide explains WSL2 setup. |
| **Human-readable code** | Every Go file is readable by a trainee or a 3-year-old AI model. No clever one-liners. Clarity beats brevity. |

---

## 4. What Gets Installed

### Always installed (no user choice)

| Category | Tools |
| -------- | ----- |
| **Package manager** | Homebrew |
| **Shell** | Zsh + Oh My Zsh |
| **Multiplexer** | Zellij |
| **Shell tools** | Atuin, zoxide, lazygit, gh, wl-clipboard (Linux) / pbcopy (Darwin) |
| **AI tools** | opencode, RTK, gentle-ai, caveman |
| **Editor** | Neovim (base config) |
| **Version control** | Git, git-credential-oauth |
| **Dotfiles sync** | chezmoi → `github.com/davichuder/dotfiles` |
| **Containers** | Docker Desktop (Darwin, brew cask) / Docker Engine (Linux, official apt repo) |
| **MCP servers** | supabase, angular, primeng, context7, postman, playwright — configured via `opencode.json` |
| **Languages (mandatory)** | Node 24 (fnm), Python 3.12, Go 1.26, C/C++ (build-essential + gdb + cmake) |

### User-selectable

| Choice | Options | Default |
| ------ | ------- | ------- |
| **Optional languages** | Java 25 (sdkman), PHP (brew) — select any, all, or none | None |
| **Nvim config** | Base (default) / Personal (overwrites base, sourced from dotfiles) | Base |
| **Nvim framework** | LazyVim, LunarVim, AstroNvim, NvChad, Launch.nvim / None — installed separately, invoked via alias | None |
| **Theme** | tokyo-night, catppuccin-mocha, gruvbox-dark, dracula, kanagawa | tokyo-night |
| **Nerd Font** | JetBrainsMono, FiraCode, CascadiaCode, Hack, Iosevka | JetBrainsMono |

---

## 5. TUI Structure

```text
[Preflight: OS detection]
      │
      ├── Windows ──→ WSL2 guide screen (rendered .md) → exit
      │
      └── Linux/Darwin ──→ Main Menu
                                │
                    ┌───────────┼───────────────┬──────────────┐
                    │           │               │              │
              Config Menu  Install Menu   Backup Menu   Reference Menu
              (write       (read config,  create /      per-tool
               config)      install)      delete        cheatsheets
                    │           │               │              │
              steps:       progress        result        (also: ? key
              font,        per module      screen        during install)
              theme,       success/fail    → Main Menu
              languages?,  screen
              nvim config, → Main Menu
              nvim
              framework?,
              chezmoi
              repo url
                    │
              config saved
              → Main Menu

              Quit
```

> Each tool's cheatsheet is an embedded `.md` file with: official links, key shortcuts, and usage examples.

---

## 6. Architecture Overview

The binary is a hybrid:

- **Simple installs** (`brew install nvim`, `fnm install 24`): executed directly via `exec.Command` in Go.
- **Complex installs** (multi-step, conditional, clone-based): orchestrated via embedded shell scripts.

The rule: if it's more than one command with conditional logic, loops, or path manipulation — it goes in a `.sh` file embedded in the binary via `embed.FS`.

Every install strategy is defined per-tool, per-OS. Brew is the fallback when no OS-specific strategy exists.

---

## 7. Out of Scope

- Installing anything on the Windows host
- Kubernetes
- Multiple install profiles (minimal / full / server)
- WSL2 installation automation

---

## 8. Repositories

| Repo | URL | Status |
| ---- | --- | ------ |
| Main installer | `github.com/davichuder/MyDots` | ✅ Created (empty) |
| Homebrew tap | `github.com/davichuder/homebrew-tap` | 🔴 To create |
| Dotfiles (chezmoi) | `github.com/davichuder/dotfiles` | 🔴 To create |

---

## 9. Roadmap

| Version | Scope | Status |
| ------- | ----- | ------ |
| **v1.0.0** | Full installer: all modules, TUI, config, audit, backup, chezmoi | 🔴 In design |
| **v1.1.0** | Multi-distro support: Arch, Debian, Fedora | 🔴 Backlog |
| **v1.2.0** | Public Homebrew tap distribution | 🔴 Backlog |
