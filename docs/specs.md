# MyDots — Specs

**Version:** 1.0.0-draft  
**Status:** Design — pending implementation  
**Author:** David  
**Date:** May 2026  
**Repo:** <https://github.com/davichuder/MyDots>

---

## 1. Functional Requirements

| ID | Requirement |
| -- | ----------- |
| **FR-01** | The binary MUST detect the host OS on every launch before rendering any UI. |
| **FR-02** | On Windows, the binary MUST render a WSL2 setup guide (embedded `.md`) and exit. It MUST NOT attempt any installation. |
| **FR-03** | On Linux or Darwin, the binary MUST open the main TUI menu. |
| **FR-04** | The Config Menu MUST collect all user choices and write `mydots-config.json` before any module is installed. |
| **FR-05** | The Install Menu MUST read `mydots-config.json`. If the file does not exist, it MUST show an inline error and stay on the Main Menu. |
| **FR-06** | Every module MUST check if it is already installed before attempting installation. If installed, it MUST skip and log the skip to the audit file. |
| **FR-07** | Critical modules MUST stop the entire installation on failure. Non-critical modules MUST log the failure and continue. |
| **FR-08** | Before overwriting any existing config file, the binary MUST create a backup in `~/.mydots-backups/<timestamp>/`. |
| **FR-09** | Every installed tool MUST be recorded in `~/.mydots-audit.json` with: module ID, name, version, install method, OS, status, and timestamp. |
| **FR-10** | The binary MUST request sudo elevation once at the start of the Install Menu using `sudo -v`. A background goroutine MUST renew the session every 45 seconds. The password MUST NOT be stored or logged anywhere. |
| **FR-11** | The Reference Menu MUST display a cheatsheet for each tool, rendered from an embedded `.md` file. |
| **FR-12** | During installation, pressing `?` MUST show the cheatsheet of the module currently being installed. |
| **FR-13** | The Backup Menu MUST allow creating a new backup or deleting an existing one. |
| **FR-14** | Theme selection MUST apply to: Neovim, Zsh (Oh My Zsh), Zellij, and Ghostty. |
| **FR-15** | Font selection MUST install the selected Nerd Font to the system font directory. Ghostty config MUST be updated to use it. |
| **FR-16** | Neovim frameworks MUST be installed to `~/.config/nvim-<framework-name>`. Each MUST have a shell alias so it can be launched by name instead of `nvim`. |
| **FR-17** | Selecting "Personal" Neovim config MUST overwrite the base config. chezmoi MUST be installed before this module runs. |
| **FR-18** | MCP servers MUST be configured by writing their entries into `~/.config/opencode/opencode.json` under the `mcp` key. No MCP binary is installed by this module. |
| **FR-19** | The binary MUST support two flags: `--unattended` (skips TUI, uses saved config — exits with error if config does not exist) and `--default` (generates config with all defaults, then installs without TUI). |
| **FR-20** | `go run .` MUST be a valid execution path for development purposes. |
| **FR-21** | On WSL2, the binary MUST detect the environment via `/proc/version` and apply WSL2-specific install strategies where they differ from native Linux. |
| **FR-22** | If `--unattended` is used and `mydots-config.json` does not exist, the binary MUST exit with code 1 and print: `"Error: no config found. Run 'mydots' to configure first."` |

---

## 2. Non-Functional Requirements

| ID | Requirement |
| -- | ----------- |
| **NFR-01** | **Idempotence.** Running the installer N times on the same machine MUST produce the same end state. |
| **NFR-02** | **Human-readable code.** Every Go file, function, and variable MUST be understandable by a junior developer without comments as a crutch. No one-liners that sacrifice clarity. |
| **NFR-03** | **Supported platforms.** v1.0.0 targets macOS Darwin and Ubuntu (native and WSL2). Any other OS MUST fail gracefully with a clear message. |
| **NFR-04** | **Go version.** The module MUST target `go 1.26.3` in `go.mod`. |
| **NFR-05** | **Binary name.** The compiled binary MUST be named `mydots`. Version is accessible via `mydots --version`. |
| **NFR-06** | **Files outside `$HOME`.** All files written by `mydots` MUST live under the user's home directory or standard system directories. Acknowledged exceptions: Docker on Linux writes to `/etc/apt/sources.list.d/` and `/etc/group` — this is unavoidable and documented in ADR-006. |
| **NFR-07** | **Audit file integrity.** Writes to `~/.mydots-audit.json` MUST be atomic: write to a temp file, then `os.Rename`. |
| **NFR-08** | **Sudo keepalive.** The sudo session MUST be renewed every 45 seconds during installation via `sudo -v` in a background goroutine. stdout/stderr of that goroutine MUST be redirected away from the TTY to prevent TUI corruption. |
| **NFR-09** | **Embedded assets.** All shell scripts, markdown cheatsheets, and the WSL2 guide MUST be embedded in the binary via `embed.FS`. |
| **NFR-10** | **Shell scripts.** All embedded `.sh` scripts MUST be POSIX-compliant and pass `shellcheck` with no errors. This applies only to scripts authored inside the MyDots repo, not to external install scripts fetched at runtime. |
| **NFR-11** | **Brew-first.** Homebrew is the default installer for every tool. OS-native managers (apt) are used only when explicitly defined in the tool's install strategy. `--cask` is Darwin-only and MUST NOT be used on Linux. |
| **NFR-12** | **WSL2 caveat — wl-clipboard.** The binary MUST check for an active Wayland session (`$WAYLAND_DISPLAY`) on WSL2. If not present, it MUST warn the user and skip the clipboard module. On Ubuntu native, if `$WAYLAND_DISPLAY` is not set, the binary MUST fall back to `xclip` (X11 clipboard). |
| **NFR-13** | **WSL2 caveat — Docker.** Docker Engine on WSL2 requires systemd. The binary MUST check `systemctl is-system-running` before attempting Docker install on WSL2. If systemd is not running, it MUST warn and skip. |
| **NFR-14** | **WSL2 caveat — Fonts.** Fonts installed to `~/.local/share/fonts/` are available to Linux GUI apps via WSLg but NOT to Windows-side terminal emulators. This limitation MUST be noted in the install result screen on WSL2. |
| **NFR-15** | **Runtime manager extensibility.** The install logic for each runtime manager (fnm, uv, sdkman) MUST follow the strategy pattern in Go. Adding a new runtime manager (e.g., swapping fnm for bun) MUST only require adding a new strategy struct, without modifying orchestration logic. |
| **NFR-16** | **Error messages as first-class feature.** Every error shown to the user MUST answer three questions: (1) What happened — concrete description of the problem. (2) Why it happened — context about the cause. (3) How to fix it — a suggested next step. A good error message is a conversation, not a shout. `Error: invalid input` is not acceptable. `Error: brew install failed because Homebrew is not in PATH. Run 'eval $(/opt/homebrew/bin/brew shellenv)' to fix it.` is. |
| **NFR-17** | **Developer Experience is User Experience.** The installer itself is a product used by developers. DX covers: (1) Easy onboarding — `brew install mydots && mydots` must work in under 2 minutes. (2) Clear errors — see NFR-16. (3) Useful documentation — cheatsheets and README answer real questions. (4) Fast feedback — the TUI must show progress in real time, not freeze silently. (5) Predictable behavior — running the same command twice must produce the same result (see NFR-01). |
| **NFR-18** | **No legacy code.** Code without tests is legacy code, regardless of when it was written or what technology it uses. Every module installer, config writer, and utility function MUST have unit tests. Acceptance scenarios in Section 9 define the integration test surface. Code that cannot be tested MUST be refactored until it can. |
| **NFR-19** | **Convention over configuration.** The installer ships with sensible defaults for every choice. The user is only asked to decide what genuinely differs from the common case (font, theme, optional languages, nvim config). Everything else installs with zero input. Fewer decisions equals more productivity. |
| **NFR-20** | **Design for the real case.** Avoid over-engineering for hypothetical edge cases that will never occur. Plan for probable changes, not possible ones. The right question is not "what if someone wants 10 fonts?" but "does the current user need more than one font right now?". When a genuine edge case is identified (WSL2 systemd, Wayland detection), handle it. When it is hypothetical, document it as out of scope and move on. |

---

## 3. Architecture Decision Records

### ADR-001 — TUI Framework: Bubble Tea v2

**Status:** Accepted  
**Context:** The installer needs a rich interactive TUI with multi-step wizards, progress views, and markdown rendering. Bash-based tools (gum, dialog) are limited and not composable.  
**Decision:** Bubble Tea v2 + lipgloss + huh? (Charm ecosystem).  
**Alternatives considered:**

| Option | Rejected reason |
| ------ | -------------- |
| gum (bash) | Not composable, hard to test, no native Go integration |
| Bubble Tea v1 | v2 has better multi-model architecture and is the active branch |
| survey (Go) | Limited to form inputs, no layout control |

**Consequences:** Dependency on a pre-stable major version. Accepted because v2 is actively maintained and this is a personal tool, not an enterprise product.

---

### ADR-002 — Primary Package Manager: Homebrew

**Status:** Accepted  
**Context:** The installer targets both macOS and Linux. A unified tool reduces the install strategy surface.  
**Decision:** Homebrew is the default for all tools on both platforms. OS-native managers (apt) are used only when brew cannot handle the install (e.g., Docker Engine, system-level C/C++ deps on Ubuntu).  
**Alternatives considered:**

| Option | Rejected reason |
| ------ | -------------- |
| apt-only | Darwin incompatible |
| nix | Too steep a learning curve for contributors |
| mise | Good for runtimes only, not general tools |

**Consequences:** `--cask` is macOS-only. Linux font and Docker installs require OS-native strategies. This is handled per-module in the install strategy matrix.

---

### ADR-003 — Hybrid Architecture: Go exec.Command + Embedded Shell Scripts

**Status:** Accepted  
**Context:** Some installs are a single command. Others require conditional logic, loops, or path manipulation.  
**Decision:** Single-command installs use `exec.Command` directly in Go. Multi-step or conditional installs use `.sh` scripts embedded via `embed.FS`.  
**Rule:** If an install requires more than one command with any conditional logic, loops, or path manipulation → it goes in a shell script.  
**Consequences:** Shell scripts must be POSIX-compliant and pass `shellcheck`. They receive runtime parameters via environment variables set by Go before execution.

---

### ADR-004 — Python Runtime Manager: uv

**Status:** Accepted  
**Context:** Python version management has historically been fragmented (pyenv, conda, system python, brew python).  
**Decision:** `uv` is installed via brew and used to install Python 3.12 via `uv python install 3.12`.  
**Alternatives considered:**

| Option | Rejected reason |
| ------ | -------------- |
| pyenv | Slower, requires shell init hooks, being superseded by uv |
| brew install python@3.12 | Works but doesn't give a version manager for future use |
| conda | Overkill for a dev environment installer |

**Consequences:** uv becomes the Python runtime manager. The install logic follows the strategy pattern (NFR-15), making it trivial to swap uv for another manager in the future.

---

### ADR-005 — Java Runtime Manager: sdkman

**Status:** Accepted  
**Context:** Java version management requires switching between versions across projects.  
**Decision:** sdkman installed via official curl script. Java 25 installed via `sdk install java 25-open`.  
**Alternatives considered:**

| Option | Rejected reason |
| ------ | -------------- |
| brew install java | No multi-version support |
| jabba | Unmaintained |
| mise | sdkman is the JVM ecosystem standard |

**Consequences:** sdkman adds `~/.sdkman` and shell init hooks to `.zshrc`. Only installed if user selects Java in Config Menu.

---

### ADR-006 — Docker Installation per OS

**Status:** Accepted  
**Context:** Docker Engine cannot be installed via brew on macOS (brew only provides the CLI client). On Linux, Docker requires its own apt repository and modifies system files outside `$HOME`.  
**Decision:**

- **Darwin:** `brew install --cask docker-desktop` (Personal plan, free).
- **Ubuntu native / WSL2:** Official Docker apt repository + `docker-ce` + `docker-ce-cli` + `containerd.io`. User added to `docker` group. Writes to `/etc/apt/sources.list.d/docker.list` and `/etc/group` — acknowledged NFR-06 exception.

**Consequences:** On Darwin, Docker Desktop requires a manual first launch to accept terms. On WSL2, Docker requires systemd (checked per NFR-13).

---

### ADR-007 — Neovim Frameworks via Git Clone

**Status:** Accepted  
**Context:** Official install scripts for some frameworks write directly to `~/.config/nvim`, overwriting the base config.  
**Decision:** All frameworks are cloned to `~/.config/nvim-<framework-name>` using `NVIM_APPNAME` isolation. Each gets a shell alias. The main `nvim` command always opens the base or personal config.  
**Consequences:** Some framework features that assume `~/.config/nvim` as home may behave differently. Accepted trade-off for co-existence.

---

### ADR-008 — Sudo Strategy: Single Prompt + Keepalive

**Status:** Accepted  
**Context:** Bubble Tea renders the TUI in raw terminal mode. A sudo password prompt appearing mid-render corrupts the display.  
**Decision:** Sudo is requested once at the start of the Install Menu via `sudo -v`. A background goroutine runs `sudo -v` every 45 seconds to keep the session alive, with stdout/stderr redirected to `/dev/null` to prevent TUI corruption. Session expires naturally after install completes.  
**Consequences:** If the user cancels the sudo prompt, the install is aborted and control returns to the Main Menu. Modules that do not require sudo (the majority) are not affected by this.

---

### ADR-009 — C/C++ Toolchain on Darwin: brew over xcode-select

**Status:** Accepted  
**Context:** `xcode-select --install` opens a native macOS GUI dialog requiring user interaction, which is incompatible with TUI automation.  
**Decision:** On Darwin, the C/C++ toolchain is installed entirely via Homebrew: `brew install gcc cmake llvm`. This provides GCC, CMake, and LLVM (which includes `clangd` for nvim LSP support).  
**Alternatives considered:**

| Option | Rejected reason |
| ------ | -------------- |
| `xcode-select --install` | Opens GUI dialog, incompatible with TUI |
| Download `.pkg` silently | Requires additional tooling and Apple auth |

**Consequences:** Uses GCC instead of Apple Clang. For development purposes (nvim LSP, cmake-based projects), this is equivalent. `lldb` is not installed; GDB is available as `lldb` alternative.

---

### ADR-010 — Nerd Fonts on Linux: Manual Install, Not brew --cask

**Status:** Accepted  
**Context:** Homebrew on Linux (Linuxbrew) does not support `--cask`. Casks are a macOS-only feature.  
**Decision:** On Darwin: `brew install --cask font-<name>-nerd-font`. On Linux: download zip from GitHub Releases (nerdfonts.com), extract to `~/.local/share/fonts/`, run `fc-cache -fv`.  
**Consequences:** Two different code paths for font install. This is unavoidable given Homebrew's architecture.

---

### ADR-011 — Runtime Manager Extensibility: Strategy Pattern

**Status:** Accepted  
**Context:** Runtime managers (fnm for Node, uv for Python, sdkman for Java) may change in the future. The user may want to swap fnm for bun, or add gradle instead of maven.  
**Decision:** Each runtime manager is implemented as a Go struct that satisfies a `RuntimeManager` interface. Adding a new manager requires only a new struct, no changes to orchestration logic.  
**Interface contract:** `Install() error`, `IsInstalled() bool`, `InstallRuntime(version string) error`, `AuditVersion() string`.  
**Consequences:** v1.0.0 ships with fnm, uv, and sdkman as the concrete implementations. Swapping or adding managers in the future is isolated to a single new file.

---

## 4. Module Catalogue

### Criticality

| Level | Behavior on failure |
| ----- | ------------------- |
| 🔴 CRITICAL | Stops the entire installation immediately. |
| 🟡 NON-CRITICAL | Logs the failure to the audit file and continues. |

---

### M-01 — Homebrew

| Field | Value |
| ----- | ----- |
| **Criticality** | 🔴 CRITICAL |
| **Idempotence check** | `which brew` returns a path |
| **Darwin strategy** | `curl-script`: official Homebrew install script |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `brew --version` |
| **Dependencies** | None |

---

### M-02 — Zsh

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which zsh` returns a path |
| **Darwin strategy** | `brew`: `brew install zsh` |
| **Ubuntu strategy** | `brew`: `brew install zsh` |
| **Post-install** | 1. Append brew zsh path to `/etc/shells` if not already present (Ubuntu only, requires sudo). 2. `chsh -s $(which zsh)` |
| **Version recorded** | `zsh --version` |
| **Dependencies** | M-01 |

---

### M-03 — Oh My Zsh

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `~/.oh-my-zsh` directory exists |
| **Darwin strategy** | `curl-script`: OMZ unattended install (`RUNZSH=no CHSH=no sh -c "$(curl ...)"`) |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | OMZ version tag |
| **Dependencies** | M-02 |

---

### M-04 — Zellij

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which zellij` returns a path |
| **Darwin strategy** | `brew`: `brew install zellij` |
| **Ubuntu strategy** | `brew`: `brew install zellij` |
| **Version recorded** | `zellij --version` |
| **Dependencies** | M-01 |

---

### M-05 — Git

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which git` returns a path |
| **Darwin strategy** | `brew`: `brew install git` |
| **Ubuntu strategy** | `brew`: `brew install git` |
| **Version recorded** | `git --version` |
| **Dependencies** | M-01 |

---

### M-06 — git-credential-oauth

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which git-credential-oauth` returns a path |
| **Darwin strategy** | `brew`: `brew install git-credential-oauth` |
| **Ubuntu strategy** | `brew`: `brew install git-credential-oauth` |
| **Version recorded** | `git-credential-oauth --version` |
| **Dependencies** | M-01, M-05 |

---

### M-07 — lazygit

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which lazygit` returns a path |
| **Darwin strategy** | `brew`: `brew install lazygit` |
| **Ubuntu strategy** | `brew`: `brew install lazygit` |
| **Version recorded** | `lazygit --version` |
| **Dependencies** | M-01, M-05 |

---

### M-08 — fnm

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which fnm` returns a path |
| **Darwin strategy** | `brew`: `brew install fnm` |
| **Ubuntu strategy** | `brew`: `brew install fnm` |
| **Version recorded** | `fnm --version` |
| **Dependencies** | M-01 |
| **Note** | Implements `RuntimeManager` interface (ADR-011). |

---

### M-09 — Node 24

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `fnm list` output contains a line starting with `v24` |
| **Darwin strategy** | `fnm`: `fnm install 24 && fnm default 24` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `node --version` after `fnm use 24` |
| **Dependencies** | M-08 |

---

### M-10 — uv

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which uv` returns a path |
| **Darwin strategy** | `brew`: `brew install uv` |
| **Ubuntu strategy** | `brew`: `brew install uv` |
| **Version recorded** | `uv --version` |
| **Dependencies** | M-01 |
| **Note** | Implements `RuntimeManager` interface (ADR-011). |

---

### M-11 — Python 3.12

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `uv python list --only-installed` output contains `3.12` |
| **Darwin strategy** | `uv`: `uv python install 3.12` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `uv run python3.12 --version` |
| **Dependencies** | M-10 |

---

### M-12 — Go 1.26

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `go version` output contains `go1.26` |
| **Darwin strategy** | `brew`: `brew install go` |
| **Ubuntu strategy** | `brew`: `brew install go` |
| **Version recorded** | `go version` |
| **Dependencies** | M-01 |

---

### M-13 — C/C++ toolchain

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | Darwin: `which clangd` returns a path. Ubuntu: `which gcc` returns a path |
| **Darwin strategy** | `brew`: `brew install gcc cmake llvm` (see ADR-009 — no GUI, no xcode-select) |
| **Ubuntu strategy** | `apt`: `sudo apt install -y build-essential gdb cmake` |
| **Version recorded** | Darwin: `$(brew --prefix llvm)/bin/clangd --version`. Ubuntu: `gcc --version` |
| **Dependencies** | M-01 |

---

### M-14 — sdkman

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `~/.sdkman` directory exists |
| **Darwin strategy** | `curl-script`: official sdkman install script |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `sdk version` |
| **Dependencies** | M-01 |
| **Note** | Only installed if user selected Java in Config Menu. Implements `RuntimeManager` interface (ADR-011). |

---

### M-15 — Java 25 (optional)

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `sdk list java` output contains `25.*installed` |
| **Darwin strategy** | `sdkman`: `sdk install java 25-open` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `java --version` |
| **Dependencies** | M-14 |
| **Note** | Only installed if user selected Java in Config Menu. |

---

### M-16 — PHP (optional)

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which php` returns a path |
| **Darwin strategy** | `brew`: `brew install php` |
| **Ubuntu strategy** | `brew`: `brew install php` |
| **Version recorded** | `php --version` |
| **Dependencies** | M-01 |
| **Note** | Only installed if user selected PHP in Config Menu. |

---

### M-17 — Neovim

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which nvim` returns a path AND `~/.config/nvim/init.lua` exists |
| **Darwin strategy** | `brew`: `brew install neovim` + copy base config to `~/.config/nvim/` |
| **Ubuntu strategy** | Same as Darwin |
| **Backup required** | Yes — backup `~/.config/nvim/` if it exists before writing |
| **Version recorded** | `nvim --version` |
| **Dependencies** | M-01 |
| **Note** | Install and base config are a single atomic module. If nvim binary exists but `init.lua` does not, config is applied without reinstalling the binary. |

---

### M-18 — Neovim personal config (optional)

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `chezmoi status` shows no diff for nvim config path |
| **Darwin strategy** | `chezmoi`: `chezmoi apply` |
| **Ubuntu strategy** | Same as Darwin |
| **Backup required** | Yes — backup `~/.config/nvim/` before chezmoi apply |
| **Version recorded** | `personal` |
| **Dependencies** | M-17 (base must exist first), M-45 (chezmoi MUST run before this module) |
| **Note** | Only installed if user selected "Personal" in Config Menu. If M-45 failed, this module MUST also be skipped and logged as `skipped-dependency-failed`. |

---

### M-19 — Neovim framework (optional)

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `~/.config/nvim-<framework-name>` directory exists |
| **Darwin strategy** | `clone`: clone framework repo to `~/.config/nvim-<name>` + append alias to `~/.zshrc` if not already present |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | Framework name + `git rev-parse HEAD` |
| **Dependencies** | M-17 |
| **Note** | Only installed if user selected a framework in Config Menu. Before appending alias, check if `~/.zshrc` already contains the alias to ensure idempotence. Alias format: `alias <name>='NVIM_APPNAME=nvim-<name> nvim'` |

#### Framework clone sources

| Framework | Clone source |
| --------- | ----------- |
| LazyVim | `https://github.com/LazyVim/starter` |
| LunarVim | `https://github.com/LunarVim/Launch.nvim` |
| AstroNvim | `https://github.com/AstroNvim/template` |
| NvChad | `https://github.com/NvChad/starter` |
| Launch.nvim | `https://github.com/LaunchVim/launch.nvim` |

---

### M-20 — Atuin

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which atuin` returns a path |
| **Darwin strategy** | `brew`: `brew install atuin` |
| **Ubuntu strategy** | `brew`: `brew install atuin` |
| **Version recorded** | `atuin --version` |
| **Dependencies** | M-01 |

---

### M-21 — zoxide

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which zoxide` returns a path |
| **Darwin strategy** | `brew`: `brew install zoxide` |
| **Ubuntu strategy** | `brew`: `brew install zoxide` |
| **Version recorded** | `zoxide --version` |
| **Dependencies** | M-01 |

---

### M-22 — bat

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which bat` returns a path |
| **Darwin strategy** | `brew`: `brew install bat` |
| **Ubuntu strategy** | `brew`: `brew install bat` |
| **Version recorded** | `bat --version` |
| **Dependencies** | M-01 |

---

### M-23 — eza

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which eza` returns a path |
| **Darwin strategy** | `brew`: `brew install eza` |
| **Ubuntu strategy** | `brew`: `brew install eza` |
| **Version recorded** | `eza --version` |
| **Dependencies** | M-01 |
| **Note** | Icons require a Nerd Font (M-47). eza installs regardless; icons render correctly only after M-47 is complete and the terminal is restarted. |

---

### M-24 — fd

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which fd` returns a path |
| **Darwin strategy** | `brew`: `brew install fd` |
| **Ubuntu strategy** | `brew`: `brew install fd` |
| **Version recorded** | `fd --version` |
| **Dependencies** | M-01 |

---

### M-25 — ripgrep

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which rg` returns a path |
| **Darwin strategy** | `brew`: `brew install ripgrep` |
| **Ubuntu strategy** | `brew`: `brew install ripgrep` |
| **Version recorded** | `rg --version` |
| **Dependencies** | M-01 |

---

### M-26 — fzf

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which fzf` returns a path |
| **Darwin strategy** | `brew`: `brew install fzf` |
| **Ubuntu strategy** | `brew`: `brew install fzf` |
| **Version recorded** | `fzf --version` |
| **Dependencies** | M-01 |

---

### M-27 — sd

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which sd` returns a path |
| **Darwin strategy** | `brew`: `brew install sd` |
| **Ubuntu strategy** | `brew`: `brew install sd` |
| **Version recorded** | `sd --version` |
| **Dependencies** | M-01 |

---

### M-28 — jq

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which jq` returns a path |
| **Darwin strategy** | `brew`: `brew install jq` |
| **Ubuntu strategy** | `brew`: `brew install jq` |
| **Version recorded** | `jq --version` |
| **Dependencies** | M-01 |

---

### M-29 — yq

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which yq` returns a path |
| **Darwin strategy** | `brew`: `brew install yq` |
| **Ubuntu strategy** | `brew`: `brew install yq` |
| **Version recorded** | `yq --version` |
| **Dependencies** | M-01 |

---

### M-30 — tldr

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which tldr` returns a path |
| **Darwin strategy** | `brew`: `brew install tldr` |
| **Ubuntu strategy** | `brew`: `brew install tldr` |
| **Version recorded** | `tldr --version` |
| **Dependencies** | M-01 |

---

### M-31 — delta

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which delta` returns a path |
| **Darwin strategy** | `brew`: `brew install git-delta` |
| **Ubuntu strategy** | `brew`: `brew install git-delta` |
| **Version recorded** | `delta --version` |
| **Dependencies** | M-01, M-05 |

---

### M-32 — bottom

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which btm` returns a path |
| **Darwin strategy** | `brew`: `brew install bottom` |
| **Ubuntu strategy** | `brew`: `brew install bottom` |
| **Version recorded** | `btm --version` |
| **Dependencies** | M-01 |

---

### M-33 — thefuck

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which thefuck` returns a path |
| **Darwin strategy** | `brew`: `brew install thefuck` |
| **Ubuntu strategy** | `brew`: `brew install thefuck` |
| **Version recorded** | `thefuck --version` |
| **Dependencies** | M-01, M-11 |

---

### M-34 — carapace

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which carapace` returns a path |
| **Darwin strategy** | `brew`: `brew install carapace` |
| **Ubuntu strategy** | `brew`: `brew install carapace` |
| **Version recorded** | `carapace --version` |
| **Dependencies** | M-01 |

---

### M-35 — glow

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which glow` returns a path |
| **Darwin strategy** | `brew`: `brew install glow` |
| **Ubuntu strategy** | `brew`: `brew install glow` |
| **Version recorded** | `glow --version` |
| **Dependencies** | M-01 |

---

### M-36 — GitHub CLI (gh)

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which gh` returns a path |
| **Darwin strategy** | `brew`: `brew install gh` |
| **Ubuntu strategy** | `brew`: `brew install gh` |
| **Version recorded** | `gh --version` |
| **Dependencies** | M-01 |

---

### M-37 — Clipboard tool

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | Darwin: always skip. Linux: `which wl-copy` OR `which xclip` returns a path |
| **Darwin strategy** | `native`: `pbcopy` is built-in, no install needed |
| **Ubuntu native strategy** | If `$WAYLAND_DISPLAY` set → `brew install wl-clipboard`. Else → `brew install xclip` (X11 fallback) |
| **WSL2 strategy** | If `$WAYLAND_DISPLAY` set → `brew install wl-clipboard`. Else → log warning, skip, notify user |
| **Version recorded** | Darwin: `built-in`. Wayland: `wl-copy --version`. X11: `xclip -version`. WSL2 no-wayland: `skipped-no-wayland` |
| **Dependencies** | M-01 (Linux only) |

---

### M-38 — Docker

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `docker --version` succeeds |
| **Darwin strategy** | `brew-cask`: `brew install --cask docker-desktop` |
| **Ubuntu native strategy** | `shell-script`: add Docker official apt repo + `sudo apt install -y docker-ce docker-ce-cli containerd.io` + add user to docker group |
| **WSL2 strategy** | Same as Ubuntu native, BUT first check systemd: if not running → log warning and skip (see NFR-13) |
| **Version recorded** | `docker --version` |
| **Dependencies** | M-01 |

---

### M-39 — lazydocker

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which lazydocker` returns a path |
| **Darwin strategy** | `brew`: `brew install lazydocker` |
| **Ubuntu strategy** | `brew`: `brew install lazydocker` |
| **Version recorded** | `lazydocker --version` |
| **Dependencies** | M-01, M-38 |

---

### M-40 — opencode

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which opencode` returns a path |
| **Darwin strategy** | `brew-tap`: `brew install anomalyco/tap/opencode` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `opencode --version` |
| **Dependencies** | M-01 |

---

### M-41 — RTK

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which rtk` returns a path |
| **Darwin strategy** | `shell-script`: `brew install rtk && rtk init` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `rtk --version` |
| **Dependencies** | M-01 |
| **Reference** | <https://github.com/rtk-ai/rtk> |

---

### M-42 — caveman

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which caveman` returns a path |
| **Darwin strategy** | `curl-script`: `curl -fsSL https://raw.githubusercontent.com/JuliusBrussee/caveman/main/install.sh \| bash -s -- --only openclaw` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `caveman --version` |
| **Dependencies** | None |
| **Reference** | <https://github.com/juliusbrussee/caveman> |

---

### M-43 — gentle-ai

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which gentle-ai` returns a path |
| **Darwin strategy** | `brew-tap`: `brew tap Gentleman-Programming/homebrew-tap && brew install gentle-ai` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `gentle-ai --version` |
| **Dependencies** | M-01 |
| **Reference** | <https://github.com/Gentleman-Programming/gentle-ai> |

---

### M-44 — MCP server config

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `~/.config/opencode/opencode.json` contains all 6 MCP server entries under the `mcp` key |
| **Darwin strategy** | `config`: write or merge MCP entries into `~/.config/opencode/opencode.json` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `config` |
| **Dependencies** | M-40 |
| **Note** | If `opencode.json` already exists, entries are merged (not overwritten). Existing keys not managed by mydots are preserved. |

#### opencode.json MCP schema

The `mcp` key follows the official opencode config format.  
Context7 is `remote`; all others are `local` (npx-based).  
Exact package names must be verified from each tool's official docs before implementation.

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "supabase": {
      "type": "local",
      "command": ["npx", "-y", "@supabase/mcp-server-supabase@latest"],
      "enabled": true
    },
    "angular": {
      "type": "local",
      "command": ["npx", "-y", "@angular/mcp@latest"],
      "enabled": true
    },
    "primeng": {
      "type": "local",
      "command": ["npx", "-y", "primeng-mcp@latest"],
      "enabled": true
    },
    "postman": {
      "type": "local",
      "command": ["npx", "-y", "@postman/mcp-server-local@latest"],
      "enabled": true
    },
    "context7": {
      "type": "remote",
      "url": "https://mcp.context7.com/mcp",
      "enabled": true
    },
    "playwright": {
      "type": "local",
      "command": ["npx", "-y", "@playwright/mcp@latest"],
      "enabled": true
    }
  }
}
```

---

### M-45 — chezmoi

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which chezmoi` returns a path AND `~/.local/share/chezmoi` is initialized (directory exists and is a git repo) |
| **Darwin strategy** | `brew`: `brew install chezmoi` + `chezmoi init <repo_url>` + `chezmoi apply` |
| **Ubuntu strategy** | Same as Darwin |
| **Version recorded** | `chezmoi --version` |
| **Dependencies** | M-01, M-05 |
| **Note** | MUST run before M-18 (personal nvim config). `repo_url` is read from `mydots-config.json`. If this module fails, M-18 MUST be skipped. |

---

### M-46 — Theme

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | All four target config files contain the selected theme name |
| **Darwin strategy** | `config`: write theme config to each tool's config file |
| **Ubuntu strategy** | Same as Darwin |
| **Backup required** | Yes — backup each affected config file before writing |
| **Version recorded** | Theme name from `mydots-config.json` |
| **Dependencies** | M-17 (nvim), M-03 (zsh), M-04 (zellij), M-48 (ghostty) |

#### Theme config targets

| Tool | Config file | Mechanism |
| ---- | ----------- | --------- |
| Neovim | `~/.config/nvim/lua/plugins/colorscheme.lua` | lazy.nvim plugin + colorscheme name |
| Zsh | `~/.zshrc` | OMZ `ZSH_THEME` or p10k theme block — idempotent via `# MYDOTS_THEME_START / END` markers |
| Zellij | `~/.config/zellij/config.kdl` | `theme "<name>"` directive — idempotent via marker comments |
| Ghostty | `~/.config/ghostty/config` | `theme = <name>` directive — idempotent via marker comments |

---

### M-47 — Nerd Font

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | Darwin: `ls ~/Library/Fonts \| grep -i <font-name>` has output. Ubuntu: `fc-list \| grep -i <font-name>` has output |
| **Darwin strategy** | `brew-cask`: `brew install --cask font-<name>-nerd-font` |
| **Ubuntu strategy** | `shell-script`: download zip from GitHub Releases → extract to `~/.local/share/fonts/` → `fc-cache -fv` (see ADR-010) |
| **WSL2 note** | Font is installed to Linux font dir. Available to WSLg apps only. Windows terminals require manual font install — noted in SCR-08. |
| **Post-install** | Update Ghostty config: `font-family = <Name> Nerd Font` — idempotent via marker comments |
| **Backup required** | Yes — backup `~/.config/ghostty/config` before writing |
| **Version recorded** | Font name from `mydots-config.json` |
| **Dependencies** | None |

#### Darwin brew cask names

| Font | Cask name |
| ---- | --------- |
| JetBrainsMono | `font-jetbrains-mono-nerd-font` |
| FiraCode | `font-fira-code-nerd-font` |
| CascadiaCode | `font-caskaydia-cove-nerd-font` |
| Hack | `font-hack-nerd-font` |
| Iosevka | `font-iosevka-nerd-font` |

#### Ubuntu GitHub release zip names (nerdfonts.com)

| Font | Zip name |
| ---- | -------- |
| JetBrainsMono | `JetBrainsMono.zip` |
| FiraCode | `FiraCode.zip` |
| CascadiaCode | `CascadiaCode.zip` |
| Hack | `Hack.zip` |
| Iosevka | `Iosevka.zip` |

---

### M-48 — Ghostty

| Field | Value |
| ----- | ----- |
| **Criticality** | 🟡 NON-CRITICAL |
| **Idempotence check** | `which ghostty` returns a path |
| **Darwin strategy** | `brew-cask`: `brew install --cask ghostty` |
| **Ubuntu native strategy** | `shell-script`: download the community-maintained `ghostty-ubuntu` installer to a temporary file, execute it with `bash`, then remove the file. This is not an official Ghostty release tarball and never uses `curl \| bash`. |
| **WSL2 strategy** | Same as Ubuntu native. Ghostty is not supported on Windows natively; the Linux binary runs via WSLg. Requires WSLg active (`$WAYLAND_DISPLAY` set). If WSLg not available, log warning and skip. |
| **Version recorded** | `ghostty --version` |
| **Dependencies** | M-01 |

---

## 5. Execution Order

The installer MUST execute modules in the following order. This is the canonical topological sort derived from the dependency graph. Optional modules (marked with `*`) are only enqueued if selected in `mydots-config.json`.

| Exec # | Module | Reason for position |
| ------ | ------ | ------------------- |
| 1 | M-01 Homebrew | CRITICAL — must be first, all brew modules depend on it |
| 2 | M-48 Ghostty | Required by M-46 (theme) and M-47 (font config) |
| 3 | M-47 Nerd Font | Required early so M-23 (eza) renders icons correctly |
| 4 | M-05 Git | Required by M-06, M-07, M-45 |
| 5 | M-06 git-credential-oauth | Depends on M-05 |
| 6 | M-45 chezmoi | MUST run before M-18 (personal nvim) |
| 7 | M-02 Zsh | Required by M-03 |
| 8 | M-03 Oh My Zsh | Depends on M-02 |
| 9 | M-04 Zellij | Independent, early for theme wiring |
| 10 | M-08 fnm | Required by M-09 |
| 11 | M-09 Node 24 | Depends on M-08; required by some MCP servers |
| 12 | M-10 uv | Required by M-11 |
| 13 | M-11 Python 3.12 | Depends on M-10; required by M-33 (thefuck) |
| 14 | M-12 Go 1.26 | Independent |
| 15 | M-13 C/C++ toolchain | Independent |
| 16* | M-14 sdkman | Required by M-15, only if Java selected |
| 17* | M-15 Java 25 | Depends on M-14 |
| 18* | M-16 PHP | Independent, only if PHP selected |
| 19 | M-17 Neovim | Required by M-18, M-19 |
| 20* | M-18 Nvim personal config | Depends on M-17 + M-45 (chezmoi) |
| 21* | M-19 Nvim framework | Depends on M-17 |
| 22 | M-46 Theme | Depends on M-17, M-03, M-04, M-48 |
| 23 | M-20 Atuin | Independent |
| 24 | M-21 zoxide | Independent |
| 25 | M-22 bat | Independent |
| 26 | M-23 eza | After M-47 (font icons) |
| 27 | M-24 fd | Independent |
| 28 | M-25 ripgrep | Independent |
| 29 | M-26 fzf | Independent |
| 30 | M-27 sd | Independent |
| 31 | M-28 jq | Independent |
| 32 | M-29 yq | Independent |
| 33 | M-30 tldr | Independent |
| 34 | M-31 delta | Depends on M-05 |
| 35 | M-32 bottom | Independent |
| 36 | M-33 thefuck | Depends on M-11 (Python) |
| 37 | M-34 carapace | Independent |
| 38 | M-35 glow | Independent |
| 39 | M-36 gh | Independent |
| 40 | M-37 Clipboard | Independent |
| 41 | M-38 Docker | Independent |
| 42 | M-39 lazydocker | Depends on M-38 |
| 43 | M-07 lazygit | Depends on M-05 |
| 44 | M-40 opencode | Required by M-44 |
| 45 | M-44 MCP config | Depends on M-40 |
| 46 | M-41 RTK | Independent |
| 47 | M-42 caveman | No dependencies |
| 48 | M-43 gentle-ai | Independent |

---

## 6. Data Schemas

### 6.1 — `mydots-config.json`

Default location: `~/.config/mydots/mydots-config.json`

```json
{
  "version": "1.0.0",
  "font": "JetBrainsMono",
  "theme": "tokyo-night",
  "languages": {
    "java": false,
    "php": false
  },
  "nvim": {
    "config": "base",
    "framework": "none"
  },
  "chezmoi": {
    "repo_url": "https://github.com/davichuder/dotfiles"
  }
}
```

#### Field definitions

| Field | Type | Values | Default |
| ----- | ---- | ------ | ------- |
| `version` | string | semver | `"1.0.0"` |
| `font` | string | `JetBrainsMono`, `FiraCode`, `CascadiaCode`, `Hack`, `Iosevka` | `"JetBrainsMono"` |
| `theme` | string | `tokyo-night`, `catppuccin-mocha`, `gruvbox-dark`, `dracula`, `kanagawa` | `"tokyo-night"` |
| `languages.java` | bool | `true`, `false` | `false` |
| `languages.php` | bool | `true`, `false` | `false` |
| `nvim.config` | string | `"base"`, `"personal"` | `"base"` |
| `nvim.framework` | string | `"lazyvim"`, `"lunarvim"`, `"astronvim"`, `"nvchad"`, `"launchnvim"`, `"none"` | `"none"` |
| `chezmoi.repo_url` | string | valid GitHub HTTPS URL | `"https://github.com/davichuder/dotfiles"` |

---

### 6.2 — `~/.mydots-audit.json`

Append-only semantics (entries are always added, never removed). Written atomically via temp-file + `os.Rename` after each module completes.

```json
{
  "version": "1.0.0",
  "entries": [
    {
      "module_id": "M-01",
      "name": "Homebrew",
      "version": "4.5.1",
      "method": "curl-script",
      "os": "darwin",
      "status": "installed",
      "timestamp": "2026-05-23T10:30:00Z"
    },
    {
      "module_id": "M-02",
      "name": "Zsh",
      "version": "5.9",
      "method": "brew",
      "os": "ubuntu-wsl2",
      "status": "skipped",
      "timestamp": "2026-05-23T10:31:00Z"
    },
    {
      "module_id": "M-38",
      "name": "Docker",
      "version": "",
      "method": "apt",
      "os": "ubuntu-wsl2",
      "status": "failed",
      "error": "systemd not available in this WSL2 instance",
      "timestamp": "2026-05-23T10:45:00Z"
    }
  ]
}
```

#### Status values

| Status | Meaning |
| ------ | ------- |
| `installed` | Module was not present and was installed successfully |
| `skipped` | Module was already installed, no action taken |
| `skipped-disabled` | Module was not selected in config (optional modules) |
| `skipped-dependency-failed` | A required dependency failed, so this module was not attempted |
| `skipped-no-wayland` | Module requires WSLg/Wayland display server which is not active |
| `failed` | Installation was attempted and failed |

#### OS field values

| Value | Meaning |
| ----- | ------- |
| `darwin` | macOS |
| `ubuntu` | Ubuntu native |
| `ubuntu-wsl2` | Ubuntu running inside WSL2 |

#### Timestamp format

All timestamps MUST be UTC ISO 8601 with seconds precision: `2026-05-23T10:30:00Z`.

---

### 6.3 — Backup Policy

| Field | Value |
| ----- | ----- |
| **Location** | `~/.mydots-backups/<timestamp>/` |
| **Timestamp format** | `2026-05-23T10-30-00Z` (colons replaced with dashes for filesystem compatibility) |
| **What is backed up** | Any file that mydots is about to overwrite, preserving its relative path from `$HOME` |
| **Retention** | No automatic deletion. User manages via Backup Menu. |
| **Example** | `~/.mydots-backups/2026-05-23T10-30-00Z/.config/nvim/init.lua` |

---

## 7. Install Strategy Matrix

| Module | Darwin | Ubuntu native | WSL2 |
| ------ | ------ | ------------- | ---- |
| M-01 Homebrew | curl-script | curl-script | curl-script |
| M-02 Zsh | brew | brew | brew |
| M-03 Oh My Zsh | curl-script | curl-script | curl-script |
| M-04 Zellij | brew | brew | brew |
| M-05 Git | brew | brew | brew |
| M-06 git-credential-oauth | brew | brew | brew |
| M-07 lazygit | brew | brew | brew |
| M-08 fnm | brew | brew | brew |
| M-09 Node 24 | fnm | fnm | fnm |
| M-10 uv | brew | brew | brew |
| M-11 Python 3.12 | uv | uv | uv |
| M-12 Go 1.26 | brew | brew | brew |
| M-13 C/C++ toolchain | brew (gcc+cmake+llvm) | apt | apt |
| M-14 sdkman | curl-script | curl-script | curl-script |
| M-15 Java 25 | sdkman | sdkman | sdkman |
| M-16 PHP | brew | brew | brew |
| M-17 Neovim | brew + config | brew + config | brew + config |
| M-18 Nvim personal config | chezmoi | chezmoi | chezmoi |
| M-19 Nvim framework | clone + alias | clone + alias | clone + alias |
| M-20 Atuin | brew | brew | brew |
| M-21 zoxide | brew | brew | brew |
| M-22 bat | brew | brew | brew |
| M-23 eza | brew | brew | brew |
| M-24 fd | brew | brew | brew |
| M-25 ripgrep | brew | brew | brew |
| M-26 fzf | brew | brew | brew |
| M-27 sd | brew | brew | brew |
| M-28 jq | brew | brew | brew |
| M-29 yq | brew | brew | brew |
| M-30 tldr | brew | brew | brew |
| M-31 delta | brew | brew | brew |
| M-32 bottom | brew | brew | brew |
| M-33 thefuck | brew | brew | brew |
| M-34 carapace | brew | brew | brew |
| M-35 glow | brew | brew | brew |
| M-36 gh | brew | brew | brew |
| M-37 Clipboard | native (skip) | wl-clipboard or xclip | wl-clipboard or skip |
| M-38 Docker | brew-cask | apt + docker repo | apt + systemd check |
| M-39 lazydocker | brew | brew | brew |
| M-40 opencode | brew-tap | brew-tap | brew-tap |
| M-41 RTK | shell-script | shell-script | shell-script |
| M-42 caveman | curl-script | curl-script | curl-script |
| M-43 gentle-ai | brew-tap | brew-tap | brew-tap |
| M-44 MCP config | config (write/merge JSON) | config (write/merge JSON) | config (write/merge JSON) |
| M-45 chezmoi | brew | brew | brew |
| M-46 Theme | config (marker-based) | config (marker-based) | config (marker-based) |
| M-47 Nerd Font | brew-cask | shell-script | shell-script |
| M-48 Ghostty | brew-cask | shell-script | shell-script (via WSLg) |

---

## 8. TUI Screen Specifications

### SCR-01 — Preflight

- Detects OS via `runtime.GOOS`.
- On Linux: reads `/proc/version` to detect WSL2 (contains `microsoft` case-insensitive).
- If Windows: go to SCR-02.
- If Linux (native or WSL2) or Darwin: go to SCR-03.
- If unknown: display error message and exit with code 1.

---

### SCR-02 — WSL2 Setup Guide (Windows only)

- Renders `assets/wsl2-guide.md` using embedded markdown renderer.
- Footer: "Press any key to exit."
- No installation occurs.

---

### SCR-03 — Main Menu

| Item | Action |
| ---- | ------ |
| Config | Go to SCR-04 |
| Install | If `mydots-config.json` missing → show inline error, stay. Else → go to SCR-05. |
| Backup | Go to SCR-06 |
| Reference | Go to SCR-07 |
| Quit | Exit with code 0 |

---

### SCR-04 — Config Menu

Multi-step wizard. No module is installed during this flow.

| Step | Input type | Question | Options | Default |
| ---- | --------- | -------- | ------- | ------- |
| 1 | Single select | Which Nerd Font? | JetBrainsMono, FiraCode, CascadiaCode, Hack, Iosevka | JetBrainsMono |
| 2 | Single select | Which theme? | tokyo-night, catppuccin-mocha, gruvbox-dark, dracula, kanagawa | tokyo-night |
| 3 | Multi-select | Optional languages? | Java 25, PHP | None |
| 4 | Single select | Neovim config? | Base, Personal | Base |
| 5 | Single select | Neovim framework? | LazyVim, LunarVim, AstroNvim, NvChad, Launch.nvim, None | None |
| 6 | Text input | Chezmoi dotfiles repo URL? | Free text | `https://github.com/davichuder/dotfiles` |
| Final | — | "Config saved. Ready to install." | — | → Main Menu |

On completion: writes `mydots-config.json` and returns to SCR-03.

---

### SCR-05 — Install Menu

| Element | Behavior |
| ------- | -------- |
| Sudo prompt | Requested once via `sudo -v` before any module starts. Cancel → abort, return to Main Menu. |
| Progress view | One row per module. States: `pending` → `running` → `✅ done` / `❌ failed` / `⏭ skipped` / `— disabled` |
| Realtime log | Scrollable pane showing stdout/stderr of the running module |
| `?` key | Opens cheatsheet of the currently running module |
| On complete | Go to SCR-08 |

---

### SCR-06 — Backup Menu

| Item | Action |
| ---- | ------ |
| Create backup now | Backs up all mydots-managed configs to `~/.mydots-backups/<timestamp>/`. Shows confirmation. |
| Delete a backup | Lists existing backups by timestamp. User selects one. Confirms before deleting. |
| Back | Return to Main Menu |

---

### SCR-07 — Reference Menu

- Lists all tools from the embedded module catalogue.
- User selects a tool: renders `assets/cheatsheets/<module-id>.md`.
- `Esc` or `q` returns to previous screen.

#### Cheatsheet `.md` structure

```markdown
# <Tool Name>

## Links
- Official site: <url>
- GitHub: <url>
- Docs: <url>

## Key Shortcuts
| Shortcut | Action |
| -------- | ------ |

## Usage Examples
\```bash
# Example 1
<command>
\```
```

---

### SCR-08 — Install Result

| Case | Display |
| ---- | ------- |
| All succeeded | "Installation complete. X modules installed." → Main Menu |
| Non-critical failures | "Installation finished with X warnings." + list of failed modules → Main Menu |
| Critical failure | "Installation stopped. Module failed. Fix the issue and re-run." → Main Menu |
| WSL2 font note | Always shown on WSL2: "Fonts installed for Linux apps. Windows terminal requires manual font install." |

---

## 9. Acceptance Scenarios

### SC-01 — Fresh install on macOS

```gherkin
Given: macOS Darwin, no previous mydots run, defaults selected
When: user completes Config Menu with defaults, runs Install Menu
Then:
  - All 48 modules attempt (some skipped as native: pbcopy)
  - ~/.mydots-audit.json has one entry per module
  - mydots-config.json exists with default values
  - `nvim` opens with base config
  - JetBrainsMono Nerd Font installed in ~/Library/Fonts/
  - Ghostty installed via brew cask
  - Ghostty config contains selected font and theme
  - Docker Desktop installed via brew cask
  - python3.12 available via uv
  - clangd available via llvm (no xcode-select dialog opened)
```

---

### SC-02 — Fresh install on Ubuntu native

```gherkin
Given: Ubuntu native (not WSL2), no previous mydots run
When: user runs `mydots`, completes Config, runs Install
Then:
  - All modules attempt
  - Docker Engine installed via official apt repo
  - User added to docker group
  - /proc/version does NOT contain "microsoft"
  - If WAYLAND_DISPLAY set: wl-clipboard installed
  - If WAYLAND_DISPLAY not set: xclip installed (X11 fallback)
  - C/C++ toolchain installed via apt (build-essential + gdb + cmake)
  - Font installed to ~/.local/share/fonts/ and fc-cache refreshed
  - Ghostty binary installed from GitHub releases to ~/.local/bin/ghostty
```

---

### SC-03 — Fresh install on WSL2

```gherkin
Given: Ubuntu running inside WSL2
When: user runs `mydots`, completes Config, runs Install
Then:
  - /proc/version contains "microsoft" → os recorded as "ubuntu-wsl2"
  - If WAYLAND_DISPLAY not set: M-37 (clipboard) skipped with warning in audit log
  - If WAYLAND_DISPLAY not set: M-48 (Ghostty) skipped with warning — requires WSLg
  - If WAYLAND_DISPLAY set: Ghostty Linux binary installed via shell-script (WSLg handles display)
  - If systemd not active: M-38 skipped with warning in audit log
  - Font installed to ~/.local/share/fonts/
  - SCR-08 shows note: "Fonts installed for Linux apps. Windows terminal requires manual font install."
  - All other modules behave as Ubuntu native
```

---

### SC-04 — Idempotent re-run

```gherkin
Given: mydots was fully installed before
When: user runs `mydots` again and runs Install Menu
Then:
  - All modules report "skipped" in the audit log
  - No file is overwritten
  - No backup is created
  - mydots-config.json is unchanged
  - Total install time < 30 seconds
```

---

### SC-05 — Non-critical module fails

```gherkin
Given: Atuin formula temporarily unavailable in Homebrew
When: Install runs
Then:
  - M-20 (Atuin) shows ❌ in progress view
  - Audit log has M-20 with status "failed" and error message
  - All other modules continue installing normally
  - SCR-08 shows "finished with 1 warning"
```

---

### SC-06 — Critical module fails

```gherkin
Given: No internet connection
When: Install runs and Homebrew install fails
Then:
  - M-01 shows ❌
  - No subsequent module is attempted
  - SCR-08 shows critical failure message
  - Audit log has M-01 with status "failed"
```

---

### SC-07 — Windows detection

```gherkin
Given: binary run on Windows (native, not WSL2)
When: preflight runs
Then:
  - WSL2 guide markdown rendered to terminal
  - No module installed
  - Binary exits with code 0 after user presses any key
```

---

### SC-08 — Install without config

```gherkin
Given: no mydots-config.json exists
When: user opens Install Menu
Then:
  - Inline error: "No config found. Please run Config first."
  - User stays on Main Menu
  - No installation attempted
```

---

### SC-09 — Personal Neovim config selected

```gherkin
Given: user selects "Personal" in Config step 4
When: Install runs
Then:
  - M-01 → M-45 (chezmoi) runs at exec position 6
  - M-17 (Neovim + base config) runs at exec position 19
  - M-18 (personal config) runs at exec position 20 via chezmoi apply
  - Backup of original ~/.config/nvim/ at ~/.mydots-backups/<timestamp>/
  - ~/.config/nvim/ reflects personal config from dotfiles repo
```

---

### SC-10 — Neovim framework selected

```gherkin
Given: user selects "LunarVim" in Config step 5
When: Install runs
Then:
  - M-19 clones https://github.com/LunarVim/Launch.nvim to ~/.config/nvim-lunarvim/
  - ~/.zshrc contains: alias lunarvim='NVIM_APPNAME=nvim-lunarvim nvim'
  - Running `lunarvim` opens LunarVim
  - Running `nvim` opens base or personal config, not LunarVim
```

---

### SC-11 — Sudo keepalive during long install

```gherkin
Given: Install starts on Ubuntu with all 48 modules
When: sudo is needed by Docker (M-38) which runs deep into the install
Then:
  - Sudo requested once at Install Menu start via sudo -v
  - Background goroutine runs sudo -v every 45 seconds
  - Goroutine stdout/stderr redirected away from TTY — no TUI corruption
  - Sudo session remains active for Docker install
  - Session expires naturally after install completes
```

---

### SC-12 — --default flag

```gherkin
Given: user runs `mydots --default`
When: binary starts
Then:
  - mydots-config.json generated with all default values
  - Installation starts immediately without opening TUI
  - All modules install using defaults (JetBrainsMono, tokyo-night, base nvim, no framework, no optional languages)
```

---

### SC-13 — --unattended without config

```gherkin
Given: no mydots-config.json exists
When: user runs `mydots --unattended`
Then:
  - Binary exits immediately with code 1
  - Prints to stderr: "Error: no config found. Run 'mydots' to configure first."
  - No TUI opened, no installation attempted
```

---

### SC-14 — chezmoi fails, personal config is skipped

```gherkin
Given: chezmoi repo URL is unreachable
When: Install runs with "Personal" nvim config selected
Then:
  - M-45 (chezmoi) shows ❌ with error message
  - M-18 (personal nvim config) is logged as "skipped-dependency-failed"
  - M-17 (base nvim config) is still installed normally
  - Installation continues with remaining modules
  - SCR-08 shows "finished with 2 warnings"
```

---

### SC-15 — Theme applied idempotently

```gherkin
Given: mydots already installed with tokyo-night
When: user changes theme to catppuccin-mocha in Config, then runs Install again
Then:
  - M-46 (Theme) detects existing MYDOTS_THEME markers in each config file
  - Replaces content between markers with new theme
  - Does NOT duplicate theme entries
  - Backup created before each file is modified
  - All four tools (nvim, zsh, zellij, ghostty) reflect catppuccin-mocha
```
