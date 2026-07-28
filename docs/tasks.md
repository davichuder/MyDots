# MyDots — Tasks

**Version:** 1.0.0-draft  
**Status:** Design — pending implementation  
**Author:** David  
**Date:** May 2026  
**Repo:** <https://github.com/davichuder/MyDots>

---

## How to read this file

### Task structure

Every feature follows the TDD RED-GREEN cycle. Within each group, tests come before implementation:

```text
[RED]   Write the failing test — it must fail before any implementation exists
[GREEN] Implement the minimum code to make the test pass
        Refactor if needed before moving to the next task
```

Unit tests use standard `go test` with **triangulation** (multiple inputs), **black-box** (behavior not internals), **path coverage** (all branches), and **real edge cases**.

Acceptance tests use **godog** (BDD / Gherkin) and map directly to the `SC-*` scenarios in `specs.md` Section 9.

TUI tests use **teatest** (`github.com/charmbracelet/x/exp/teatest`) with **golden files** for visual regression.

### Commit cadence

Commit after each RED-GREEN pair — not after a phase. Each `[RED] + [GREEN]` block is one atomic commit. Example: `test(platform): detect WSL2 from /proc/version` then `feat(platform): implement Detect()`.

### Acceptance scenario references

`SC-01` through `SC-15` point to `specs.md` Section 9. `FR-*` and `NFR-*` point to `specs.md` Sections 1 and 2.

---

## Phase 0 — Project Setup

- [x] **T-001** Initialize `go.mod` with path `github.com/davichuder/MyDots` and `go 1.26.3`
- [x] **T-002** Add dependencies:
  - `charm.land/bubbletea/v2` — TUI framework (v2.0.6+, stable)
  - `charm.land/bubbles/v2` — UI components
  - `charm.land/lipgloss/v2` — styling
  - `charm.land/huh` — form inputs for Config Menu
  - `github.com/charmbracelet/x/exp/teatest` — TUI testing
  - `github.com/charmbracelet/x/exp/golden` — snapshot golden files
  - `github.com/cucumber/godog` — BDD acceptance tests
- [x] **T-003** Create the full directory structure defined in `design.md` Section 2, including `assets/cheatsheets/`, `assets/scripts/`, and all `internal/` packages
- [x] **T-004** Create `assets.go` with `//go:embed assets` directive
- [x] **T-005** Create `assets/wsl2-guide.md` with WSL2 installation instructions (installation steps, links, prerequisites)
- [x] **T-006** Add `Makefile` with targets: `build`, `test`, `test-integration`, `test-bdd`, `lint`, `shellcheck`, `update-golden`
- [x] **T-007** Configure `.golangci.yml` enforcing: no untyped string enums, `errcheck`, `govet`, `staticcheck`
- [x] **T-008** Add `.gitattributes` marking `testdata/*.golden` as binary to prevent git from modifying line endings
- [x] **T-009** Add `.github/workflows/ci.yml`: `go test ./...`, `golangci-lint`, `shellcheck assets/scripts/` on every PR
- [x] **T-010** Create `github.com/davichuder/homebrew-tap` repo with base formula stub for `mydots`
- [x] **T-011** Create `github.com/davichuder/dotfiles` repo initialized for chezmoi

---

## Phase 1 — Core Infrastructure

### 1.1 Platform detection

**[RED]**

- [x] **T-012** Write failing tests for `Detect()`.
  - `Detect()` must accept an injectable `goos string` parameter so tests can simulate any OS without cross-compilation: `Detect(goos string) (Platform, error)`. `main.go` calls `Detect(runtime.GOOS)`.
  - Triangulation: `goos="darwin"`, `goos="linux"` + native `/proc/version`, `goos="linux"` + WSL2 `/proc/version`, `goos="windows"`, `goos="plan9"`
  - Edge cases: `/proc/version` unreadable (permission denied), file missing, file empty, file contains `Microsoft` in uppercase, file contains multiple lines with "microsoft" on line 2
  - Path coverage: all branches of the OS switch, both variants of the Linux branch
  - Black-box: assert on `Platform` fields and error types, not on internal string parsing

**[GREEN]**

- [x] **T-013** Implement `Platform`, `OS`, `Variant` typed string constants, `ErrWindowsDetected`, `ErrUnsupportedOS`, and `Detect(goos string) (Platform, error)` (design.md §3.1, §6)

---

### 1.2 Config system

**[RED]**

- [x] **T-014** Write failing tests for all config operations:
  - `DefaultConfig()`: assert every field matches specs.md §6.1 defaults exactly
  - `Load()`: valid JSON, malformed JSON, missing file, file with unknown fields (forward compatibility), file with valid JSON but invalid enum value
  - `Save()`: creates parent directories if missing, overwrites existing file atomically, returns error on permission denied
  - `Validate()`: valid config passes; each invalid enum value for `Font`, `Theme`, `NvimConfig`, `NvimFramework` fails independently; `chezmoi.repo_url` as HTTP fails, as empty string fails, as valid HTTPS passes
  - `DefaultConfigPath()`: correct path constructed from `os.UserHomeDir()` + `.config/mydots/mydots-config.json`
  - Golden file: `Save()` output of `DefaultConfig()` matches `testdata/default-config.golden`

**[GREEN]**

- [x] **T-015** Implement all typed string constants, `Config` struct, and all config functions (design.md §7)

---

### 1.3 Audit system

**[RED]**

- [x] **T-016** Write failing tests for audit operations.
  - Critical requirement: `Append()` must be safe for concurrent calls. The read-modify-write cycle (`loadOrEmpty → append → writeFile → Rename`) is NOT thread-safe with only `os.Rename`. A `sync.Mutex` at package level is required.
  - `Append()` first call: creates file with correct JSON structure and exactly one entry
  - `Append()` subsequent calls: accumulates entries in order, never overwrites existing ones
  - `Append()` with missing parent directory: creates the directory before writing
  - Concurrent `Append()` calls via `sync.WaitGroup` launching N goroutines: final file has exactly N entries and is valid parseable JSON
  - `entry.Error` is omitted from JSON (no `"error": ""`) when the error field is empty
  - Timestamp field is UTC ISO 8601: `2026-05-23T10:30:00Z`
  - Golden files: single-entry audit JSON, multi-entry audit JSON, entry with non-empty error field

**[GREEN]**

- [x] **T-017** Implement `Entry` struct, `InstallStatus` constants, `loadOrEmpty()`, and `Append()` with `sync.Mutex` + temp file + `os.Rename` (design.md §8, NFR-07)

---

### 1.4 Backup system

**[RED]**

- [x] **T-018** Write failing tests for backup operations:
  - `BackupFile()`: file is copied to `~/.mydots-backups/<timestamp>/<relative-from-HOME>`, relative path is preserved correctly
  - `BackupFile()` with symlink: the symlink target content is copied (not a dangling link)
  - `BackupDir()`: all files in the directory are backed up with correct paths; subdirectories are preserved
  - `ListBackups()`: returns sorted list newest-first; returns empty slice (not error) when no backups exist
  - `DeleteBackup()`: removes the correct directory; returns error for non-existent timestamp
  - Edge cases: source file outside `$HOME` (returns error), backup dir already exists (merges, does not fail), source file is empty (backup succeeds)

**[GREEN]**

- [x] **T-019** Implement `BackupFile()`, `BackupDir()`, `ListBackups()`, `DeleteBackup()` (design.md §9)

---

### 1.5 Sudo keepalive

**[RED]**

- [x] **T-020** Write failing tests for sudo keepalive:
  - `StartKeepalive()` goroutine stops cleanly when context is cancelled — verify no goroutine leak using a done channel or timeout assertion
  - `StartKeepalive()` goroutine output goes to `io.Discard` — mock executor verifies no bytes written to caller's writer
  - `RequestElevation()` connects process stdin/stdout/stderr to the terminal (mock executor verifies these are set, not nil)
  - Keepalive interval is 45 seconds — mock executor records timestamps of `sudo -v` calls

**[GREEN]**

- [x] **T-021** Implement `RequestElevation()` and `StartKeepalive(ctx context.Context)` (design.md §10)

---

### 1.6 Error types

**[RED]**

- [x] **T-022** Write failing tests for error interface and implementations:
  - Each error type (`BrewInstallError`, `AptInstallError`, `CurlScriptError`, `ConfigWriteError`) implements `MyDotsError`
  - `What()`, `Why()`, `Fix()` return non-empty strings for all types
  - `.Error()` string is non-empty for all types
  - `GenericError` wrapping a known stderr pattern (e.g., "not found in PATH") returns a non-empty, actionable `Fix()`
  - `GenericError` wrapping an unknown error returns the fallback message defined in design.md §12
  - Triangulation: at least 3 different error types with different formula/command/path values

**[GREEN]**

- [x] **T-023** Implement `MyDotsError` interface, `BrewInstallError`, `AptInstallError`, `CurlScriptError`, `ConfigWriteError`, and `GenericError` (design.md §12, NFR-16)

---

## Phase 2 — Install Pipeline

### 2.1 Module interface and types

**[RED]**

- [x] **T-024** Write failing tests for module types and constants:
  - All `ModuleID` constants are non-empty strings with no duplicates across the full set
  - `Criticality` and `InstallStatus` string values match specs.md §6.2 exactly
  - `InstallContext` fields are all accessible and a zero-value struct does not panic on field access

**[GREEN]**

- [x] **T-025** Implement `ModuleID` constants (one per module M-01 through M-48), `Criticality`, `InstallStatus`, `Module` interface, and `InstallContext` struct (design.md §3.2, §3.5)

---

### 2.2 Runner

**[RED]**

- [x] **T-026** Write failing tests for runner functions using mock executor (`SetExecutor`):
  - `CommandExists()`: returns true when binary is on PATH, false when not found
  - `CaptureOutput()`: returns stdout trimmed on success, returns empty string on non-zero exit
  - `Run(ctx, logw, name, args...)`: every stdout line is written to `logw` via `bufio.Scanner`; stderr lines are also written to `logw`; non-zero exit returns error; context cancellation returns error
  - `Run()` edge cases: command not found, binary exits immediately with non-zero code, binary produces no output, binary produces a single partial line without newline
  - `Brew()`: passes `brew` as the command with correct remaining args
  - `BrewCask(p, caskName)`: returns `ErrNotDarwin` when `p.OS == Linux`; calls correct brew args on Darwin
  - `Script()`: temp file is created with `0700` permissions; env vars are merged (new values take precedence); temp file is removed after execution regardless of success or error
  - Path coverage: success path, non-zero exit path, context-cancelled path, missing command path, temp file creation failure

**[GREEN]**

- [x] **T-027** Implement all runner functions with `exec.CommandContext`, `bufio.Scanner` for line-by-line output, and `SetExecutor()` for mock injection (design.md §5.4)

---

### 2.3 GenericBrewModule

**[RED]**

- [x] **T-028** Write failing tests for `BrewModule`:
  - `IsInstalled()` calls `CommandExists` with the configured `checkCommand`
  - `Install()` calls `runner.Brew` with the configured `formula`
  - `AuditInfo()` calls `CaptureOutput` with `checkCommand --version`
  - `Dependencies()` returns the configured deps slice
  - `Criticality()` always returns `NonCritical`
  - Triangulation: three different `BrewModule` instances with different formulas/checkCommands (e.g., zoxide, atuin, bat)

**[GREEN]**

- [x] **T-029** Implement `BrewModule` struct and all interface methods (design.md §3.3)

---

### 2.4 Catalogue

**[RED]**

- [x] **T-030** Write failing tests for the module catalogue:
  - `allModules()` returns exactly 48 modules
  - No two modules share the same `ID()`
  - The first module is `M-01` (Homebrew) and its `Criticality()` returns `Critical`
  - All 47 remaining modules return `NonCritical`
  - The order matches specs.md §5 exactly: M-45 (chezmoi) appears at exec position 6, before M-18 (personal nvim) at position 20
  - M-48 (Ghostty) appears at exec position 2, before M-46 (Theme) at position 22
  - All `ModuleID` values in any module's `Dependencies()` refer to IDs that exist in `allModules()`

**[GREEN]**

- [x] **T-031** Implement `allModules()` returning all 48 modules in the canonical order defined in specs.md §5 (design.md §5.1). Note: this is a statically ordered list, not a runtime topological sort. `Dependencies()` is used only by the executor for failure propagation.

---

### 2.5 Planner

**[RED]**

- [x] **T-032** Write failing tests for `BuildPlan()`:
  - All defaults: plan has 43 modules (48 minus 5 optional)
  - java=true: M-14 and M-15 are included in the plan in correct order
  - php=true: M-16 is included
  - nvim=personal: M-18 is included; M-45 still precedes M-18 in the result
  - framework=lazyvim: M-19 is included
  - All optional enabled: all 48 modules in plan
  - Edge cases: unknown `nvim.config` value treated as base; unknown `framework` treated as none

**[GREEN]**

- [x] **T-033** Implement `BuildPlan()` and `isDisabled()` (design.md §5.2)

---

### 2.6 Executor

**[RED]**

- [x] **T-034** Write failing tests for executor behavior:
  - All modules succeed: channel receives `StatusInstalled` for each in plan order, then channel closes
  - Non-critical module fails: subsequent modules still run; channel stays open; `failedIDs` records the failure
  - Critical module (M-01) fails: channel closes immediately after failure event; no subsequent module events are sent
  - Module already installed (`IsInstalled` returns true): channel receives `StatusSkipped`
  - Direct dependency failed: dependent module receives `StatusSkippedDependencyFailed`
  - Transitive dependency failed: skipped modules propagate to `failedIDs` so their dependents are also skipped
  - Channel is always closed — even on critical failure; no goroutine leak possible
  - Triangulation: test with 1 module, 3 modules in chain, 5 modules with branching dependencies

**[GREEN]**

- [x] **T-035** Implement `Run()` top-level executor and `runOne()` helper (design.md §5.3)
- [x] **T-036** Define `ProgressEvent` struct with `ModuleID`, `Status`, `LogLine`, `Err`

---

## Phase 3 — Complex Modules

Each module task: tests cover idempotence (already installed → skip), install per OS, failure path, and any WSL2-specific logic.

### 3.1 Critical module

**[RED]**

- [x] **T-037** Write failing tests for `homebrew.go` (M-01):
  - `IsInstalled()` true when `which brew` succeeds; false otherwise
  - `Install()` on Darwin: calls homebrew curl-script (mock)
  - `Install()` on Linux: same script (mock)
  - `Criticality()` returns `Critical`
  - Edge case: script returns non-zero exit → error wraps as `CurlScriptError` with `What/Why/Fix`

**[GREEN]**

- [x] **T-038** Implement `homebrew.go`

---

### 3.2 Shell modules

**[RED]**

- [x] **T-039** Write failing tests for `zsh.go` (M-02):
  - `IsInstalled()` checks `which zsh`
  - `Install()` on Ubuntu: brew install, then appends brew zsh path to `/etc/shells` only if not already present, then `chsh -s $(which zsh)`
  - `Install()` on Darwin: brew install + `chsh` only; no `/etc/shells` modification
  - Re-run idempotence: if zsh path already in `/etc/shells`, it is not appended again
  - Edge case: `chsh` fails (permission) → error returned, shell change aborted

**[GREEN]**

- [x] **T-040** Implement `zsh.go`

**[RED]**

- [x] **T-041** Write failing tests for `oh_my_zsh.go` (M-03):
  - `IsInstalled()` checks `~/.oh-my-zsh` directory exists
  - `Install()` runs omz-install.sh script with `RUNZSH=no CHSH=no` env vars set
  - Edge case: `~/.oh-my-zsh` exists but is empty → still treated as installed (directory check only)

**[GREEN]**

- [x] **T-042** Implement `oh_my_zsh.go`

---

### 3.3 Toolchain

**[RED]**

- [x] **T-043** Write failing tests for `c_cpp.go` (M-13):
  - `IsInstalled()` on Darwin: checks `which clangd`
  - `IsInstalled()` on Ubuntu: checks `which gcc`
  - `Install()` on Darwin: calls `brew install gcc cmake llvm` — no `xcode-select` call (ADR-009)
  - `Install()` on Ubuntu: calls `sudo apt install -y build-essential gdb cmake`
  - Edge case: apt install exits non-zero → `AptInstallError` with exit code and stderr

**[GREEN]**

- [x] **T-044** Implement `c_cpp.go`

---

### 3.4 Runtime managers

**[RED]**

- [x] **T-045** Write failing tests for `FnmManager`, `UvManager`, `SdkmanManager`:
  - Each satisfies `RuntimeManager` interface
  - `Install()` calls correct brew formula or curl-script (via mock runner)
  - `InstallRuntime(version)` calls the manager's install command with the given version string
  - `IsInstalled()` checks the correct binary exists
  - Triangulation: two different version strings for each manager

**[GREEN]**

- [x] **T-046** Implement `internal/runtime/fnm.go`, `uv.go`, `sdkman.go` (design.md §3.4)

**[RED]**

- [x] **T-047** Write failing tests for modules using runtime managers (`fnm_node.go` for M-08+M-09, `uv_python.go` for M-10+M-11):
  - Module constructor accepts a `RuntimeManager` — no import of concrete manager type
  - `Install()` delegates install and runtime installation to the injected manager
  - Mock manager returning error on `Install()` → module returns error
  - Mock manager returning error on `InstallRuntime()` → module returns error

**[GREEN]**

- [x] **T-048** Implement `fnm_node.go` and `uv_python.go`

---

### 3.5 Neovim suite

**[RED]**

- [x] **T-049** Write failing tests for `neovim.go` (M-17):
  - `IsInstalled()` requires BOTH `which nvim` AND `~/.config/nvim/init.lua` to exist
  - Binary exists but `init.lua` missing: `Install()` applies config only, does not reinstall binary
  - `Install()` backs up `~/.config/nvim/` before writing config if the directory exists
  - Edge case: backup fails → install aborted, no partial state

**[GREEN]**

- [x] **T-050** Implement `neovim.go`

**[RED]**

- [x] **T-051** Write failing tests for `neovim_personal.go` (M-18):
  - If M-45 (chezmoi) is in `failedIDs` → returns `StatusSkippedDependencyFailed` without calling chezmoi
  - Backs up `~/.config/nvim/` before `chezmoi apply`
  - `IsInstalled()` runs `chezmoi status` and checks for no diff output

**[GREEN]**

- [x] **T-052** Implement `neovim_personal.go` (SC-09, SC-14)

**[RED]**

- [x] **T-053** Write failing tests for `neovim_framework.go` (M-19):
  - `IsInstalled()` checks `~/.config/nvim-<framework>` directory exists
  - `Install()` clones the correct repo URL for each of the 5 frameworks (triangulation)
  - Alias is appended to `~/.zshrc` in correct format: `alias <name>='NVIM_APPNAME=nvim-<name> nvim'`
  - Re-run: if alias already exists in `~/.zshrc`, it is not duplicated
  - Edge case: git clone fails → no partial `~/.config/nvim-<name>` directory left behind

**[GREEN]**

- [x] **T-054** Implement `neovim_framework.go` (SC-10)

---

### 3.6 WSL2-aware modules

**[RED]**

- [x] **T-055** Write failing tests for `clipboard.go` (M-37) — all 5 paths must be covered:
  - Darwin: `IsInstalled()` always true; `Install()` is a no-op
  - Ubuntu native + `$WAYLAND_DISPLAY` set: installs `wl-clipboard` via brew
  - Ubuntu native + `$WAYLAND_DISPLAY` not set: installs `xclip` via brew (X11 fallback, NFR-12)
  - WSL2 + `$WAYLAND_DISPLAY` set: installs `wl-clipboard` via brew
  - WSL2 + `$WAYLAND_DISPLAY` not set: logs warning, returns `StatusSkippedNoWayland`

**[GREEN]**

- [x] **T-056** Implement `clipboard.go` (NFR-12)

**[RED]**

- [x] **T-057** Write failing tests for `docker.go` (M-38):
  - Darwin: calls `brew install --cask docker-desktop`
  - Ubuntu native: runs `docker-linux.sh` script (mock)
  - WSL2 + systemd active (`systemctl is-system-running` returns 0): runs `docker-linux.sh` script
  - WSL2 + systemd not active: logs warning, module is skipped with `StatusFailed` (not a hard stop)
  - Edge case: apt repository setup fails → `AptInstallError` returned

**[GREEN]**

- [x] **T-058** Implement `docker.go` (ADR-006, NFR-13)

**[RED]**

- [x] **T-059** Write failing tests for `ghostty.go` (M-48):
  - Darwin: `brew install --cask ghostty`
  - Ubuntu native: runs `ghostty-linux.sh` script (mock)
  - WSL2 + `$WAYLAND_DISPLAY` set: runs `ghostty-linux.sh` (WSLg provides display)
  - WSL2 + `$WAYLAND_DISPLAY` not set: returns `StatusSkippedNoWayland` with warning

**[GREEN]**

- [x] **T-060** Implement `ghostty.go`

---

### 3.7 Config-writing modules

**[RED]**

- [x] **T-061** Write failing tests for `theme.go` (M-46):
  - Theme written to all 4 config files (nvim, zsh, zellij, ghostty)
  - Uses `# MYDOTS_THEME_START / END` markers — content between markers is replaced, not appended
  - First run: markers and theme content are inserted
  - Re-run with same theme: markers found, content unchanged, no duplication
  - Re-run with different theme: old theme replaced by new one between markers
  - Backup created for each file before writing
  - Edge case: one config file missing → that file is skipped, others are written successfully
  - Golden files: each config file after first theme application

**[GREEN]**

- [x] **T-062** Implement `theme.go` (SC-15)

**[RED]**

- [x] **T-063** Write failing tests for `nerd_font.go` (M-47):
  - Darwin: `brew install --cask font-<name>-nerd-font` — correct cask name per font (triangulation: all 5 fonts)
  - Ubuntu: runs `font-linux.sh` with `FONT_NAME` env var set to the correct zip filename
  - Post-install on both platforms: Ghostty config updated with `font-family = <Name> Nerd Font` using idempotent marker
  - Re-run: Ghostty config marker found and updated, not duplicated

**[GREEN]**

- [x] **T-064** Implement `nerd_font.go` (ADR-010)

**[RED]**

- [x] **T-065** Write failing tests for `mcp_config.go` (M-44):
  - Writes all 6 MCP entries when `opencode.json` does not exist
  - Merges entries into existing `opencode.json` — existing unmanaged keys (e.g., `"theme"`) are preserved
  - Creates `~/.config/opencode/` directory if missing
  - `IsInstalled()` checks all 6 keys exist in the `mcp` object
  - Edge case: existing file has malformed JSON → returns error, does not overwrite the file
  - Golden file: resulting `opencode.json` after merge with pre-existing unmanaged keys

**[GREEN]**

- [x] **T-066** Implement `mcp_config.go`

**[RED]**

- [x] **T-067** Write failing tests for `chezmoi.go` (M-45):
  - `IsInstalled()` requires binary AND `~/.local/share/chezmoi` directory containing a git repo
  - `Install()` sequence: brew install → `chezmoi init <repo_url>` → `chezmoi apply`
  - `repo_url` is read from `ctx.Config.Chezmoi.RepoURL`
  - Edge case: `chezmoi init` fails (repo unreachable) → error returned; `chezmoi apply` is never called

**[GREEN]**

- [x] **T-068** Implement `chezmoi.go`

---

### 3.8 AI tools

**[RED]**

- [x] **T-069** Write failing tests for `rtk.go` (M-41), `caveman.go` (M-42), `gentle_ai.go` (M-43):
  - RTK: `Install()` calls `brew install rtk` then `rtk init` in sequence via shell-script; if first step fails, second step is not called
  - caveman: `Install()` runs `caveman-install.sh` script (mock)
  - gentle-ai: `Install()` calls `brew tap Gentleman-Programming/homebrew-tap` then `brew install gentle-ai`; tap failure returns error before install attempt

**[GREEN]**

- [x] **T-070** Implement `rtk.go`, `caveman.go`, `gentle_ai.go`

---

### 3.9 Optional language modules

**[RED]**

- [ ] **T-071** Write failing tests for `sdkman.go` (M-14), `java.go` (M-15), `php.go` (M-16):
  - sdkman: runs `sdkman-install.sh` curl-script; `IsInstalled()` checks `~/.sdkman` directory exists and contains a `bin/sdkman-init.sh` file
  - java: calls `SdkmanManager.InstallRuntime("25-open")`; `IsInstalled()` checks `sdk list java` output contains `25.*installed`
  - php: simple brew install; `IsInstalled()` checks `which php`
  - All three: only enqueued when corresponding config flag is true

**[GREEN]**

- [ ] **T-072** Implement `sdkman.go`, `java.go`, `php.go`

---

## Phase 4 — Shell Scripts

All scripts must pass `shellcheck --shell=sh` before any commit.

- [ ] **T-073** Write and verify `assets/scripts/homebrew-install.sh` — idempotent, unattended, exits non-zero on failure
- [ ] **T-074** Write and verify `assets/scripts/omz-install.sh` — `RUNZSH=no CHSH=no`, no interactive prompts
- [ ] **T-075** Write and verify `assets/scripts/sdkman-install.sh` — official installer, unattended
- [ ] **T-076** Write and verify `assets/scripts/docker-linux.sh` — apt repo setup + install + `usermod -aG docker $USER`; idempotent (checks if user already in group)
- [ ] **T-077** Write and verify `assets/scripts/font-linux.sh` — reads `$FONT_NAME` env var, downloads zip from nerdfonts GitHub releases, extracts to `~/.local/share/fonts/`, runs `fc-cache -fv`. Note: `unzip` must be available; script must check and fail clearly if missing.
- [ ] **T-078** Write and verify `assets/scripts/ghostty-linux.sh` — downloads release tarball, extracts, moves binary to `~/.local/bin/ghostty`
- [ ] **T-079** Write and verify `assets/scripts/caveman-install.sh` — wrapper for the caveman curl-script with `--only openclaw`
- [ ] **T-080** Add `shellcheck` CI step that fails the build if any script has warnings

---

## Phase 5 — TUI

TUI tests use **teatest** for full-program behavior and **golden files** for visual regression. Unit tests drive model state directly without teatest for faster feedback.

### 5.1 Components

**[RED]**

- [ ] **T-081** Write failing tests for `LogPane`:
  - Appending a line increases content by exactly one line
  - After 200 lines, oldest lines are dropped (ring buffer); content length stays at 200
  - Appending empty string is a no-op (content unchanged)
  - Note: 200 is defined as the constant `LogPaneMaxLines` — change the constant to change the limit

**[GREEN]**

- [ ] **T-082** Implement `LogPane` with `LogPaneMaxLines = 200` constant

**[RED]**

- [ ] **T-083** Write failing tests for `ProgressRow`:
  - Each `InstallStatus` value renders the correct icon (`✅`, `❌`, `⏭`, `—`, spinner for running)
  - Module name is visible in the rendered output
  - Golden file: rendered row for each distinct status value

**[GREEN]**

- [ ] **T-084** Implement `ProgressRow`

---

### 5.2 Screens

**[RED]**

- [ ] **T-085** Write failing tests for `main_menu.go` using direct model state manipulation:
  - All 5 menu items are present in `View()` output
  - Selecting "Quit" returns `tea.Quit` command
  - Selecting "Install" when config file is missing: model stays on main menu, error text is visible in `View()`
  - Selecting "Install" when config exists: `Update` returns `ChangeScreenMsg{Screen: ScreenInstall}`
  - Golden file: main menu rendered output

**[GREEN]**

- [ ] **T-086** Implement `main_menu.go` (SC-08)

**[RED]**

- [ ] **T-087** Write failing tests for `config_menu.go` using direct model manipulation:
  - Step 1 presents exactly 5 font options
  - Each step pre-selects the value from `DefaultConfig()`
  - Completing all 6 steps writes `mydots-config.json` to the correct path
  - Step 6 validates the chezmoi URL: HTTP scheme shows error message and does not advance to next step
  - Golden file: each step rendered output

**[GREEN]**

- [ ] **T-088** Implement `config_menu.go` (FR-04)

**[RED]**

- [ ] **T-089** Write failing tests for `install_screen.go` using direct model manipulation:
  - Pressing `?` sets `cheatsheetVisible = true`
  - Pressing `?` again sets `cheatsheetVisible = false`
  - Pressing `q` calls `cancelFn()`
  - `ProgressEvent` with `StatusInstalled` updates the matching row icon to `✅`
  - `ProgressEvent` with `StatusFailed` updates the row to `❌`
  - `ProgressEvent` with `LogLine` appends to `LogPane`
  - Closed channel triggers `ChangeScreenMsg` to result screen
  - `waitForProgress` re-queues itself on each event until channel closes

**[GREEN]**

- [ ] **T-090** Implement `install_screen.go` and `waitForProgress()` (FR-10, FR-12)

**[RED]**

- [ ] **T-091** Write failing tests for `result_screen.go`:
  - All modules succeeded: success message with count visible
  - Some non-critical failures: warning count and failed module names visible
  - Critical failure: specific module name and failure message visible
  - WSL2 platform: font note always visible regardless of other outcomes
  - Golden files: all 4 result variants

**[GREEN]**

- [ ] **T-092** Implement `result_screen.go` (SCR-08, NFR-14)

**[RED]**

- [ ] **T-093** Write teatest integration tests for full TUI flow:
  - Program starts on Darwin: `WaitFor` until main menu text is visible in output
  - Program starts with Windows GOOS: WSL2 guide text is visible; program exits after keypress
  - Golden file: main menu full render output from teatest

**[GREEN]**

- [ ] **T-094** Implement `app.go` root model, `preflight.go`, and `reference_menu.go` (FR-11)

**[RED]**

- [ ] **T-095** Write failing tests for `backup_menu.go`:
  - "Create backup now" triggers backup of all managed config paths
  - "Delete a backup" lists existing backups and deletes the selected one
  - Empty backup list: appropriate empty-state message in `View()`

**[GREEN]**

- [ ] **T-096** Implement `backup_menu.go` (FR-13)

---

## Phase 6 — Cheatsheets

- [ ] **T-097** Write cheatsheets in `assets/cheatsheets/` using the structure from specs.md SCR-07: `## Links`, `## Key Shortcuts`, `## Usage Examples`. One `.md` file per tool, named after the tool (e.g., `ghostty.md`, `homebrew.md`).

All 48 tools: homebrew, zsh, oh-my-zsh, zellij, git, git-credential-oauth, lazygit, fnm, node, uv, python, go, c-cpp-toolchain, sdkman, java, php, neovim, neovim-personal, neovim-framework, atuin, zoxide, bat, eza, fd, ripgrep, fzf, sd, jq, yq, tldr, delta, bottom, thefuck, carapace, glow, gh, clipboard, docker, lazydocker, opencode, rtk, caveman, gentle-ai, mcp-servers, chezmoi, theme, nerd-font, ghostty.

---

## Phase 7 — Entry Point and Flags

**[RED]**

- [ ] **T-098** Write failing tests for entry point logic.
  - Extract all flag and routing logic to `run(args []string, goos string) int` — `main()` calls this and passes `os.Args` and `runtime.GOOS`. Tests call `run()` directly without spawning a process or calling `os.Exit`.
  - `--version`: returns exit code 0 and `buildVersion` appears in output
  - Windows GOOS: `renderWSL2Guide()` is called; return value is 0
  - Unknown OS: error message on stderr; return value is 1
  - `--unattended` with missing config: stderr contains the exact message from FR-22; return value is 1
  - `--unattended` with valid config: pipeline starts, TUI not started
  - `--default`: `DefaultConfig()` is written to disk; pipeline starts; TUI not started
  - Default (no flags): TUI program is started

**[GREEN]**

- [ ] **T-099** Implement `run(args []string, goos string) int`, `main()` as a thin wrapper, `renderWSL2Guide()`, `runTUI()`, `runUnattended()`, `runWithDefaults()` (design.md §13, FR-22, SC-12, SC-13)
- [ ] **T-100** In `runUnattended()` and `runWithDefaults()`: call `sudo.RequestElevation()` and `sudo.StartKeepalive(ctx)` before pipeline starts

---

## Phase 8 — BDD Acceptance Tests

Each `SC-*` from `specs.md` Section 9 maps to a Gherkin `.feature` file under `internal/test/features/`. Step definitions are in `internal/test/steps/`.

### Feature files

- [ ] **T-101** Write `features/fresh_install_macos.feature` (SC-01)

```gherkin
Feature: Fresh install on macOS
  Scenario: User completes default config and runs Install Menu
    Given a Darwin machine with no previous mydots run
    When the user runs "mydots" and completes Config with all defaults
    And runs the Install Menu
    Then the audit log has one entry per module
    And mydots-config.json contains the default font "JetBrainsMono"
    And the JetBrainsMono font is installed under ~/Library/Fonts/
    And Ghostty config contains the selected theme
```

- [ ] **T-102** Write `features/fresh_install_ubuntu.feature` (SC-02)
- [ ] **T-103** Write `features/fresh_install_wsl2.feature` (SC-03) — clipboard and Docker skip paths
- [ ] **T-104** Write `features/idempotent_rerun.feature` (SC-04)
- [ ] **T-105** Write `features/noncritical_failure.feature` (SC-05)
- [ ] **T-106** Write `features/critical_failure.feature` (SC-06)
- [ ] **T-107** Write `features/windows_detection.feature` (SC-07)
- [ ] **T-108** Write `features/install_without_config.feature` (SC-08)
- [ ] **T-109** Write `features/personal_nvim.feature` (SC-09)
- [ ] **T-110** Write `features/nvim_framework.feature` (SC-10)
- [ ] **T-111** Write `features/sudo_keepalive.feature` (SC-11)
- [ ] **T-112** Write `features/default_flag.feature` (SC-12)
- [ ] **T-113** Write `features/unattended_no_config.feature` (SC-13)
- [ ] **T-114** Write `features/chezmoi_failure.feature` (SC-14)
- [ ] **T-115** Write `features/theme_idempotent.feature` (SC-15)

### Step definitions

- [ ] **T-116** Implement godog step definitions for all platform Given steps — inject platform via `Detect(goos)` with a mock GOOS string and a temp home directory
- [ ] **T-117** Implement step definitions for all install pipeline When/Then steps (mock runner, temp home dir, real audit/backup/config writers)
- [ ] **T-118** Implement step definitions for audit log assertions (parse `~/.mydots-audit.json` and assert on fields)
- [ ] **T-119** Implement step definitions for file system assertions (file exists, file contains substring, dir exists)
- [ ] **T-120** Add `test-bdd` Makefile target: `godog ./internal/test/features/...`

---

## Phase 9 — Distribution

- [ ] **T-121** Write Homebrew formula in `github.com/davichuder/homebrew-tap` pointing to GitHub release
- [ ] **T-122** Add `goreleaser` config building for Darwin (amd64 + arm64) and Linux (amd64) on tag push
- [ ] **T-123** Add GitHub Actions release workflow triggered on `v*` tags: build → test → release
- [ ] **T-124** End-to-end test on Darwin: `brew tap davichuder/homebrew-tap && brew install mydots && mydots --version`
- [ ] **T-125** End-to-end test on Ubuntu: same install + version check
- [ ] **T-126** Write `README.md`: one-line install, quick start, supported platforms, link to specs

---

## Acceptance Criteria

A task is done when:

1. Code compiles with no warnings (`go build ./...`)
2. For `[RED]` tasks: test file exists and all new tests **fail** before implementation exists
3. For `[GREEN]` tasks: all tests in the package **pass** (`go test ./...`)
4. `golangci-lint run` reports no issues
5. Any shell script passes `shellcheck --shell=sh`
6. Behavior matches the corresponding FR/NFR in `specs.md`

The pipeline is considered complete when all BDD scenarios in Phase 8 pass against a real (non-mock) install on each supported platform.
