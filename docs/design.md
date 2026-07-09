# MyDots — Design

**Version:** 1.0.0-draft  
**Status:** Design — pending implementation  
**Author:** David  
**Date:** May 2026  
**Repo:** <https://github.com/davichuder/MyDots>

---

## 1. Overview

MyDots is a single Go binary with three moving parts:

- **TUI** (Bubble Tea v2) — collects user decisions and shows install progress
- **Install pipeline** — runs modules in a fixed order, captures output, updates audit
- **Embedded assets** — shell scripts, cheatsheets, and the WSL2 guide baked into the binary at compile time

The binary has no runtime dependencies beyond Go's standard library and the declared `go.mod` modules. Everything it needs is either embedded or fetched on demand (brew, apt, curl) during the install phase.

---

## 2. Project Structure

```text
mydots/
├── main.go                        # entry point: flag parsing, preflight, TUI start
├── assets.go                      # //go:embed assets
├── go.mod
├── go.sum
│
├── assets/
│   ├── wsl2-guide.md
│   ├── cheatsheets/               # one .md file per tool, named after the tool
│   └── scripts/
│       ├── homebrew-install.sh
│       ├── omz-install.sh
│       ├── sdkman-install.sh
│       ├── docker-linux.sh
│       ├── font-linux.sh
│       ├── ghostty-linux.sh
│       └── caveman-install.sh
│
└── internal/
    ├── platform/
    │   ├── detect.go              # OS/variant detection, Platform struct
    │   └── detect_test.go
    │
    ├── config/
    │   ├── config.go              # read, write, validate mydots-config.json
    │   └── config_test.go
    │
    ├── audit/
    │   ├── audit.go               # append entry atomically to ~/.mydots-audit.json
    │   └── audit_test.go
    │
    ├── backup/
    │   ├── backup.go              # copy files preserving relative path under $HOME
    │   └── backup_test.go
    │
    ├── sudo/
    │   ├── sudo.go                # one-time prompt + keepalive goroutine
    │   └── sudo_test.go
    │
    ├── runtime/
    │   ├── manager.go             # RuntimeManager interface
    │   ├── fnm.go                 # Node via fnm
    │   ├── uv.go                  # Python via uv
    │   └── sdkman.go              # Java via sdkman
    │
    ├── installer/
    │   ├── context.go             # InstallContext — bundles shared state for the pipeline
    │   ├── module.go              # Module interface, ModuleID, Criticality, InstallStatus
    │   ├── brew_module.go         # GenericBrewModule — covers all simple brew installs
    │   ├── catalogue.go           # ordered slice of all 48 modules
    │   ├── planner.go             # filters catalogue based on config
    │   ├── executor.go            # runs the plan, feeds progress to TUI via channel
    │   ├── runner/
    │   │   ├── runner.go          # exec.CommandContext wrapper with io.Writer logging
    │   │   └── runner_test.go
    │   └── modules/               # one file per COMPLEX module only
    │       ├── homebrew.go
    │       ├── zsh.go
    │       ├── neovim.go
    │       ├── neovim_personal.go
    │       ├── neovim_framework.go
    │       ├── clipboard.go
    │       ├── docker.go
    │       ├── theme.go
    │       ├── nerd_font.go
    │       ├── ghostty.go
    │       ├── chezmoi.go
    │       ├── mcp_config.go
    │       ├── rtk.go
    │       ├── caveman.go
    │       └── gentle_ai.go
    │
    └── tui/
        ├── app.go                 # root Bubble Tea model, screen routing
        ├── screens/
        │   ├── preflight.go
        │   ├── main_menu.go
        │   ├── config_menu.go
        │   ├── install_screen.go
        │   ├── backup_menu.go
        │   ├── reference_menu.go
        │   └── result_screen.go
        └── components/
            ├── progress_row.go
            └── log_pane.go        # bounded ring buffer, max 200 lines
```

---

## 3. Core Types and Interfaces

### 3.1 — Platform

```go
// internal/platform/detect.go

type OS string

const (
    Darwin OS = "darwin"
    Linux  OS = "linux"
)

type Variant string

const (
    Native Variant = "native"
    WSL2   Variant = "wsl2"
)

type Platform struct {
    OS      OS
    Variant Variant // only meaningful when OS == Linux; Darwin is always Native
    Arch    string  // "amd64" or "arm64"
}
```

Detection logic in `Detect() (Platform, error)`:

1. `runtime.GOOS == "windows"` → return `ErrWindowsDetected`
2. `runtime.GOOS == "darwin"` → `Platform{OS: Darwin, Variant: Native}`
3. `runtime.GOOS == "linux"` → read `/proc/version`; if it contains `"microsoft"` (case-insensitive) → `Platform{OS: Linux, Variant: WSL2}`; otherwise → `Platform{OS: Linux, Variant: Native}`
4. Any other `GOOS` → `ErrUnsupportedOS`

Module install strategies switch on `p.OS` for the primary split, then on `p.Variant` for WSL2-specific behaviour within Linux.

---

### 3.2 — Module

```go
// internal/installer/module.go

type ModuleID string

const (
    ModHomebrew        ModuleID = "M-01"
    ModGhostty         ModuleID = "M-48"
    ModNerdFont        ModuleID = "M-47"
    // ... one constant per module
)

type Criticality string

const (
    Critical    Criticality = "critical"
    NonCritical Criticality = "non-critical"
)

type InstallStatus string

const (
    StatusInstalled               InstallStatus = "installed"
    StatusSkipped                 InstallStatus = "skipped"
    StatusSkippedDisabled         InstallStatus = "skipped-disabled"
    StatusSkippedDependencyFailed InstallStatus = "skipped-dependency-failed"
    StatusSkippedNoWayland        InstallStatus = "skipped-no-wayland"
    StatusFailed                  InstallStatus = "failed"
)

type Module interface {
    ID()           ModuleID
    Name()         string
    Criticality()  Criticality
    Dependencies() []ModuleID

    IsInstalled(p platform.Platform) bool
    Install(ctx InstallContext) error
    AuditInfo() string // version string, "config", "personal", etc.
}
```

`AuditInfo()` replaces `RecordedVersion()`. The name reflects what is actually recorded: a version for tools that expose one (`zoxide --version`), a descriptive string for config-only modules (`"config"`, `"personal"`, font name).

---

### 3.3 — GenericBrewModule

Simple brew modules are declared as data, not code. No individual file per tool.

```go
// internal/installer/brew_module.go

type BrewModule struct {
    id           ModuleID
    name         string
    formula      string   // brew install <formula>
    checkCommand string   // which <checkCommand> to verify install
    deps         []ModuleID
}

func (m BrewModule) ID()           ModuleID    { return m.id }
func (m BrewModule) Name()         string      { return m.name }
func (m BrewModule) Criticality()  Criticality { return NonCritical }
func (m BrewModule) Dependencies() []ModuleID  { return m.deps }

func (m BrewModule) IsInstalled(_ platform.Platform) bool {
    return runner.CommandExists(m.checkCommand)
}

func (m BrewModule) Install(ctx InstallContext) error {
    return runner.Brew(ctx.Log, "install", m.formula)
}

func (m BrewModule) AuditInfo() string {
    return runner.CaptureOutput(m.checkCommand, "--version")
}
```

All ~33 simple brew modules are declared in `catalogue.go`:

```go
var (
    ModZoxide  = BrewModule{id: ModZoxideID,  name: "zoxide",  formula: "zoxide",  checkCommand: "zoxide",  deps: []ModuleID{ModHomebrewID}}
    ModAtuin   = BrewModule{id: ModAtuinID,   name: "atuin",   formula: "atuin",   checkCommand: "atuin",   deps: []ModuleID{ModHomebrewID}}
    ModBat     = BrewModule{id: ModBatID,     name: "bat",     formula: "bat",     checkCommand: "bat",     deps: []ModuleID{ModHomebrewID}}
    // ... etc
)
```

Complex modules with OS-specific logic, multi-step installs, or config writes get their own file under `internal/installer/modules/`.

---

### 3.4 — RuntimeManager

Implements ADR-011. Adding a new runtime manager requires only a new struct.

```go
// internal/runtime/manager.go

type RuntimeManager interface {
    Install(logw io.Writer) error
    IsInstalled() bool
    InstallRuntime(version string, logw io.Writer) error
    AuditInfo() string
}
```

Example:

```go
// internal/runtime/fnm.go

type FnmManager struct{}

func (f FnmManager) Install(logw io.Writer) error {
    return runner.Brew(logw, "install", "fnm")
}
func (f FnmManager) IsInstalled() bool { return runner.CommandExists("fnm") }
func (f FnmManager) InstallRuntime(v string, logw io.Writer) error {
    return runner.Run(logw, "fnm", "install", v)
}
func (f FnmManager) AuditInfo() string { return runner.CaptureOutput("fnm", "--version") }
```

The Node, Python, and Java modules receive a `RuntimeManager` instance via their constructor. They import only the interface, never the concrete type.

---

### 3.5 — InstallContext

Bundles shared state for the entire install session. Passed to every `Install()` call instead of threading individual dependencies through every function signature.

```go
// internal/installer/context.go

type InstallContext struct {
    Platform         platform.Platform
    Config           config.Config
    SessionTimestamp string        // shared across all backup operations
    Log              io.Writer     // receives stdout/stderr from runner for TUI
    Cancel           context.Context
}
```

`Log` is an `io.Writer` that the TUI install screen provides. The runner writes every output line to it. The TUI reads from it and appends to the `LogPane`. `Cancel` carries the context used by `exec.CommandContext` so pressing `q` kills the running subprocess.

---

## 4. TUI Architecture

Bubble Tea v2 uses the Elm architecture. Import paths for v2:

```go
import (
    tea      "charm.land/bubbletea/v2"
    "charm.land/bubbles/v2/progress"
    "charm.land/bubbles/v2/spinner"
    "charm.land/lipgloss/v2"
)
```

In v2, `View()` returns `tea.View`, key events are `tea.KeyPressMsg`, and the program is started with `tea.NewProgram(model)`.

### 4.1 — Root Model and Screen Routing

```go
// internal/tui/app.go

type Screen string

const (
    ScreenPreflight  Screen = "preflight"
    ScreenMainMenu   Screen = "main-menu"
    ScreenConfigMenu Screen = "config-menu"
    ScreenInstall    Screen = "install"
    ScreenBackup     Screen = "backup"
    ScreenReference  Screen = "reference"
    ScreenResult     Screen = "result"
)

type App struct {
    currentScreen Screen
    platform      platform.Platform
    config        config.Config

    mainMenu      MainMenuModel
    configMenu    ConfigMenuModel
    installScreen InstallScreenModel
    backupMenu    BackupMenuModel
    reference     ReferenceMenuModel
    result        ResultModel
}
```

Sub-models signal screen transitions by returning a `ChangeScreenMsg` from `Update`. The root `App.Update` intercepts it and sets `currentScreen`.

### 4.2 — Install Screen

```go
type InstallScreenModel struct {
    rows       []ProgressRow         // one per module in the plan
    logPane    LogPane               // bounded ring buffer, max 200 lines
    progressCh chan installer.ProgressEvent
    cancelFn   context.CancelFunc
    cheatsheet struct {
        visible bool
        content string
    }
}
```

The executor runs in a goroutine. It sends `ProgressEvent` structs to `progressCh`. `waitForProgress` bridges the channel to the Bubble Tea event loop:

```go
type ProgressEvent struct {
    ModuleID ModuleID
    Status   InstallStatus
    LogLine  string
    Err      error
}

func waitForProgress(ch chan installer.ProgressEvent) tea.Cmd {
    return func() tea.Msg {
        return <-ch
    }
}
```

When `Update` receives a `ProgressEvent`, it updates the matching row and re-queues `waitForProgress`. When the channel closes, the model transitions to the result screen.

Key events handled:

```go
func (m InstallScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "?":
            // toggle cheatsheet for current module
        case "ctrl+c", "q":
            m.cancelFn() // cancels context → kills running subprocess
            return m, tea.Quit
        }
    case installer.ProgressEvent:
        // update row, append to log pane
        return m, waitForProgress(m.progressCh)
    }
    return m, nil
}
```

### 4.3 — Config Menu

Built with `huh?` (Charm). Each step is a `huh.Form` with one `huh.Group`. Steps are rendered sequentially. The final step writes `mydots-config.json` and sends a `ChangeScreenMsg` to the root model.

---

## 5. Install Pipeline

### 5.1 — Catalogue and Fixed Order

The install order is not computed at runtime. It is defined statically in `catalogue.go`, matching the canonical execution order in `specs.md` Section 5. `Dependencies()` on each module is used by the executor for failure propagation only — not for ordering.

```go
// internal/installer/catalogue.go

func allModules() []Module {
    return []Module{
        modules.Homebrew,      // M-01 — CRITICAL
        modules.Ghostty,       // M-48
        modules.NerdFont,      // M-47
        modules.Git,           // M-05
        modules.GitCredOAuth,  // M-06
        modules.Chezmoi,       // M-45
        modules.Zsh,           // M-02
        modules.OhMyZsh,       // M-03
        ModZellij,             // M-04 — simple brew
        modules.Fnm,           // M-08
        modules.Node,          // M-09
        modules.Uv,            // M-10
        modules.Python,        // M-11
        ModGo,                 // M-12 — simple brew
        modules.CppToolchain,  // M-13
        modules.Sdkman,        // M-14 — optional
        modules.Java,          // M-15 — optional
        modules.Php,           // M-16 — optional
        modules.Neovim,        // M-17
        modules.NeovimPersonal,// M-18 — optional
        modules.NeovimFramework,// M-19 — optional
        modules.Theme,         // M-46
        ModAtuin,              // M-20 — simple brew
        ModZoxide,             // M-21 — simple brew
        ModBat,                // M-22 — simple brew
        ModEza,                // M-23 — simple brew
        ModFd,                 // M-24 — simple brew
        ModRipgrep,            // M-25 — simple brew
        ModFzf,                // M-26 — simple brew
        ModSd,                 // M-27 — simple brew
        ModJq,                 // M-28 — simple brew
        ModYq,                 // M-29 — simple brew
        ModTldr,               // M-30 — simple brew
        ModDelta,              // M-31 — simple brew
        ModBottom,             // M-32 — simple brew
        ModThefuck,            // M-33 — simple brew
        ModCarapace,           // M-34 — simple brew
        ModGlow,               // M-35 — simple brew
        ModGh,                 // M-36 — simple brew
        modules.Clipboard,     // M-37
        modules.Docker,        // M-38
        ModLazydocker,         // M-39 — simple brew
        ModLazygit,            // M-07 — simple brew
        ModOpencode,           // M-40 — simple brew-tap
        modules.McpConfig,     // M-44
        modules.Rtk,           // M-41
        modules.Caveman,       // M-42
        modules.GentleAi,      // M-43
    }
}
```

### 5.2 — Planner

```go
// internal/installer/planner.go

func BuildPlan(cfg config.Config, p platform.Platform) []Module {
    all := allModules()
    plan := make([]Module, 0, len(all))
    for _, mod := range all {
        if isDisabled(mod, cfg) {
            continue
        }
        plan = append(plan, mod)
    }
    return plan
}

func isDisabled(mod Module, cfg config.Config) bool {
    switch mod.ID() {
    case ModSdkmanID, ModJavaID:
        return !cfg.Languages.Java
    case ModPhpID:
        return !cfg.Languages.PHP
    case ModNeovimPersonalID:
        return cfg.Nvim.Config != NvimConfigPersonal
    case ModNeovimFrameworkID:
        return cfg.Nvim.Framework == NvimFrameworkNone
    }
    return false
}
```

Disabled modules get a `StatusSkippedDisabled` audit entry written at plan time, before execution starts.

### 5.3 — Executor

```go
// internal/installer/executor.go

func Run(plan []Module, ctx InstallContext, ch chan ProgressEvent) {
    defer close(ch)

    failedIDs := map[ModuleID]bool{}

    for _, mod := range plan {
        if err := runOne(mod, ctx, ch, failedIDs); err != nil {
            // only Critical modules return a non-nil error from runOne
            return
        }
    }
}

func runOne(mod Module, ctx InstallContext, ch chan ProgressEvent, failedIDs map[ModuleID]bool) error {
    // dependency check
    for _, dep := range mod.Dependencies() {
        if failedIDs[dep] {
            ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusSkippedDependencyFailed}
            audit.Append(audit.Entry{ModuleID: string(mod.ID()), Status: string(StatusSkippedDependencyFailed)})
            return nil
        }
    }

    // idempotence check
    if mod.IsInstalled(ctx.Platform) {
        ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusSkipped}
        audit.Append(audit.Entry{ModuleID: string(mod.ID()), Status: string(StatusSkipped), Version: mod.AuditInfo()})
        return nil
    }

    // install
    ch <- ProgressEvent{ModuleID: mod.ID(), Status: "running"}
    if err := mod.Install(ctx); err != nil {
        failedIDs[mod.ID()] = true
        ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusFailed, Err: err}
        audit.Append(audit.Entry{ModuleID: string(mod.ID()), Status: string(StatusFailed), Error: err.Error()})
        if mod.Criticality() == Critical {
            return err // signals executor to stop
        }
        return nil
    }

    ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusInstalled}
    audit.Append(audit.Entry{ModuleID: string(mod.ID()), Status: string(StatusInstalled), Version: mod.AuditInfo()})
    return nil
}
```

No `goto`. `runOne` extracts the per-module logic into a named function that returns early via `return nil` or `return err`.

### 5.4 — Runner

```go
// internal/installer/runner/runner.go

// Run executes a command and writes each output line to logw.
// Uses exec.CommandContext so the process is killed if ctx is cancelled.
func Run(ctx context.Context, logw io.Writer, name string, args ...string) error

// Brew is a shorthand for `brew install/uninstall/tap ...`
func Brew(ctx context.Context, logw io.Writer, args ...string) error

// BrewCask installs a cask. Returns ErrNotDarwin if called on Linux.
func BrewCask(ctx context.Context, logw io.Writer, p platform.Platform, caskName string) error

// BrewTap taps a Homebrew tap and installs the formula.
func BrewTap(ctx context.Context, logw io.Writer, tap, formula string) error

// Script extracts an embedded shell script to a temp file and executes it.
func Script(ctx context.Context, logw io.Writer, fs embed.FS, path string, env map[string]string) error

// CommandExists checks if a binary is reachable on PATH.
func CommandExists(name string) bool

// CaptureOutput runs a command and returns its stdout as a string.
func CaptureOutput(name string, args ...string) string
```

`Run` and all wrappers use `exec.CommandContext(ctx, ...)`. When the context is cancelled (user presses `q`), the OS sends SIGKILL to the running subprocess. Every output line is written to `logw` line by line via `bufio.Scanner` on the command's combined stdout+stderr pipe.

`Script` extracts the embedded file to `os.CreateTemp`, sets `0700`, runs it with `/bin/sh`, then defers `os.Remove`. The `env` map is merged with `os.Environ()` with new values appended last so they take precedence.

---

## 6. Platform Detection

```go
// internal/platform/detect.go

func Detect() (Platform, error) {
    switch runtime.GOOS {
    case "windows":
        return Platform{}, ErrWindowsDetected

    case "darwin":
        return Platform{OS: Darwin, Variant: Native, Arch: runtime.GOARCH}, nil

    case "linux":
        data, err := os.ReadFile("/proc/version")
        variant := Native
        if err == nil && strings.Contains(strings.ToLower(string(data)), "microsoft") {
            variant = WSL2
        }
        return Platform{OS: Linux, Variant: variant, Arch: runtime.GOARCH}, nil

    default:
        return Platform{}, ErrUnsupportedOS{GOOS: runtime.GOOS}
    }
}
```

---

## 7. Config System

```go
// internal/config/config.go

type Config struct {
    Version   string          `json:"version"`
    Font      FontChoice      `json:"font"`
    Theme     ThemeChoice     `json:"theme"`
    Languages LanguageOptions `json:"languages"`
    Nvim      NvimOptions     `json:"nvim"`
    Chezmoi   ChezmoiOptions  `json:"chezmoi"`
}

type FontChoice    string
type ThemeChoice   string
type NvimConfig    string
type NvimFramework string

const (
    FontJetBrainsMono FontChoice = "JetBrainsMono"
    FontFiraCode      FontChoice = "FiraCode"
    FontCascadiaCode  FontChoice = "CascadiaCode"
    FontHack          FontChoice = "Hack"
    FontIosevka       FontChoice = "Iosevka"
)

const (
    ThemeTokyoNight ThemeChoice = "tokyo-night"
    ThemeCatppuccin ThemeChoice = "catppuccin-mocha"
    ThemeGruvbox    ThemeChoice = "gruvbox-dark"
    ThemeDracula    ThemeChoice = "dracula"
    ThemeKanagawa   ThemeChoice = "kanagawa"
)

const (
    NvimConfigBase     NvimConfig = "base"
    NvimConfigPersonal NvimConfig = "personal"
)

const (
    NvimFrameworkLazyVim   NvimFramework = "lazyvim"
    NvimFrameworkLunarVim  NvimFramework = "lunarvim"
    NvimFrameworkAstroNvim NvimFramework = "astronvim"
    NvimFrameworkNvChad    NvimFramework = "nvchad"
    NvimFrameworkLaunchVim NvimFramework = "launchnvim"
    NvimFrameworkNone      NvimFramework = "none"
)

func DefaultConfig() Config
func Load(path string) (Config, error)
func Save(path string, cfg Config) error
func Validate(cfg Config) error
func DefaultConfigPath() string
```

`Validate` checks that all typed fields contain known constant values and that `Chezmoi.RepoURL` is a valid HTTPS URL. Because fields use typed string constants, the compiler catches invalid values at build time. `Validate` handles values read from the JSON file at runtime.

---

## 8. Audit System

```go
// internal/audit/audit.go

type Entry struct {
    ModuleID  string        `json:"module_id"`
    Name      string        `json:"name"`
    Version   string        `json:"version"`   // from Module.AuditInfo()
    Method    string        `json:"method"`
    OS        string        `json:"os"`
    Status    InstallStatus `json:"status"`
    Error     string        `json:"error,omitempty"`
    Timestamp string        `json:"timestamp"` // UTC ISO 8601: 2026-05-23T10:30:00Z
}

func Append(path string, entry Entry) error
```

`Append` reads the current file, appends the entry, and writes back atomically using a temp file + `os.Rename`.

---

## 9. Backup System

```go
// internal/backup/backup.go

// BackupFile copies a single file to ~/.mydots-backups/<sessionTimestamp>/<relative-from-HOME>.
func BackupFile(filePath string, sessionTimestamp string) error

// BackupDir walks a directory and calls BackupFile for each file.
func BackupDir(dirPath string, sessionTimestamp string) error

// ListBackups returns all backup directories sorted by timestamp descending.
func ListBackups() ([]Backup, error)

// DeleteBackup removes a backup directory by its timestamp string.
func DeleteBackup(timestamp string) error
```

The `sessionTimestamp` is generated once at the start of the install session (`2026-05-23T10-30-00Z`, colons replaced with dashes) and passed via `InstallContext`. All backup operations in the same session share the same timestamp directory.

Any module that writes a managed config file MUST call `BackupFile` or `BackupDir` before writing. This is enforced by code review convention and documented in the constraints table (Section 14).

---

## 10. Sudo Keepalive

```go
// internal/sudo/sudo.go

// RequestElevation is called BEFORE Bubble Tea enters raw mode.
// The user types their password in the normal terminal.
func RequestElevation() error {
    cmd := exec.Command("sudo", "-v")
    cmd.Stdin  = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// StartKeepalive renews the sudo session every 45 seconds in a goroutine.
// stdout/stderr are discarded to prevent TUI corruption.
// Cancelling ctx stops the goroutine cleanly.
func StartKeepalive(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(45 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                cmd := exec.Command("sudo", "-v")
                cmd.Stdout = io.Discard
                cmd.Stderr = io.Discard
                _ = cmd.Run()
            }
        }
    }()
}
```

`RequestElevation()` is called before Bubble Tea starts. Once it returns, the TUI takes over the terminal. `StartKeepalive()` runs in the background. The same `ctx` passed to `StartKeepalive` is stored in `InstallContext.Cancel` and used by `exec.CommandContext` in the runner, so cancellation propagates cleanly to both the keepalive goroutine and any running subprocess.

---

## 11. Embedded Assets

```go
// assets.go

//go:embed assets
var assets embed.FS
```

Cheatsheets are accessed by tool name:

```go
data, err := assets.ReadFile("assets/cheatsheets/ghostty.md")
data, err := assets.ReadFile("assets/wsl2-guide.md")
```

Shell scripts are extracted and executed by `runner.Script()`. The function creates a temp file, writes the script content, sets `0700`, runs it with `/bin/sh`, and defers cleanup. Environment variables are passed via the `env` map merged into `os.Environ()`.

All scripts under `assets/scripts/` must be POSIX-compliant and pass `shellcheck --shell=sh` in CI (NFR-10).

---

## 12. Error Handling

All errors produced by MyDots satisfy (NFR-16):

```go
// internal/errors/errors.go

type MyDotsError interface {
    error
    What() string // concrete description of what happened
    Why() string  // context about the cause
    Fix() string  // suggested next step for the user
}
```

Example:

```go
type BrewInstallError struct {
    Formula  string
    ExitCode int
    Stderr   string
}

func (e BrewInstallError) Error() string {
    return fmt.Sprintf("brew install %s failed (exit %d)", e.Formula, e.ExitCode)
}
func (e BrewInstallError) What() string {
    return fmt.Sprintf("brew install %s failed with exit code %d.", e.Formula, e.ExitCode)
}
func (e BrewInstallError) Why() string {
    return fmt.Sprintf("Homebrew reported: %s", e.Stderr)
}
func (e BrewInstallError) Fix() string {
    return "Run 'brew doctor' to diagnose Homebrew issues, then re-run 'mydots'."
}
```

The TUI renders `What`, `Why`, and `Fix` in the log pane and result screen. Plain `error` values from third-party libraries are wrapped in a `GenericError` that extracts a fix suggestion from known error patterns (exit codes, common stderr strings). If no pattern matches, `Fix()` returns `"Check the log above for details and re-run 'mydots'."`.

---

## 13. Flags and Entry Point

```go
// main.go

func main() {
    unattended  := flag.Bool("unattended", false, "skip TUI, use saved config")
    useDefaults := flag.Bool("default", false, "generate default config and install")
    showVersion := flag.Bool("version", false, "print version and exit")
    flag.Parse()

    if *showVersion {
        fmt.Println(buildVersion)
        os.Exit(0)
    }

    p, err := platform.Detect()
    if errors.Is(err, platform.ErrWindowsDetected) {
        renderWSL2Guide()
        os.Exit(0)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    switch {
    case *unattended:
        runUnattended(p)     // exits 1 with message if config not found (FR-22)
    case *useDefaults:
        runWithDefaults(p)   // writes DefaultConfig(), then installs without TUI
    default:
        runTUI(p)            // tea.NewProgram(NewApp(p)).Run()
    }
}
```

In `runUnattended` and `runWithDefaults`, sudo is requested via `RequestElevation()` before the pipeline starts, same as in the TUI flow.

---

## 14. Testing Strategy

Every package under `internal/` has a `_test.go` file. No package ships without tests (NFR-18).

### Unit tests

| Package | What to test |
| ------- | ----------- |
| `platform` | `Detect()` with mocked `/proc/version` for each OS and variant |
| `config` | `Load`, `Save`, `Validate` — valid inputs, invalid enum values, bad URLs |
| `audit` | `Append` — atomic write, missing directory, correct timestamp format |
| `backup` | `BackupFile` — relative path preservation, correct session timestamp dir |
| `sudo` | `StartKeepalive` — goroutine stops cleanly when context is cancelled |
| `runtime` | Each manager: `IsInstalled` and `InstallRuntime` with mock runner |
| `installer/brew_module` | `IsInstalled` (mock `CommandExists`) and `Install` (mock runner) |
| `installer/modules` | Each complex module: `IsInstalled` and `Install` with mock runner and platform variants |
| `installer/planner` | `BuildPlan` — all config combinations produce the correct ordered list |
| `installer/executor` | Critical failure stops and closes channel; non-critical continues; dependency failure skips dependents |
| `installer/runner` | `Run` writes to `io.Writer`, cancels on context done, captures exit code |

### Integration surface

The acceptance scenarios in `specs.md` Section 9 define the integration test surface. Each SC-* scenario maps to an integration test that sets up a temp home directory, runs the install pipeline with a mock runner, and asserts the expected files, audit entries, and exit codes.

Integration tests use `//go:build integration` and run separately from unit tests.

### Mock runner

Tests never call real `brew`, `apt`, or `curl`. The `runner` package exposes `SetExecutor(exec Executor)` to swap the real shell executor for a fake one that records calls and returns configured results.

---

## 15. Key Constraints

| Constraint | Where enforced |
| ---------- | ------------- |
| `brew --cask` never on Linux | `runner.BrewCask()` checks `p.OS == Darwin`, returns `ErrNotDarwin` otherwise |
| Atomic audit writes | `audit.Append()` always uses temp file + `os.Rename` |
| Sudo never stored | `RequestElevation()` has no return value; password never touches Go memory |
| Backup before overwrite | Any function writing a managed config file must call `backup.BackupFile()` first |
| POSIX shell scripts | `shellcheck --shell=sh` runs on all files under `assets/scripts/` in CI |
| Simple modules as data | 48 files is not the goal; simple brew modules are declared in `catalogue.go` as `BrewModule` values |
| Complex modules in own file | `internal/installer/modules/` only contains modules with OS-specific or multi-step logic |
| All enums are typed strings | No raw string literals for enum values anywhere in the codebase |
| Tests required | Every `.go` file (except auto-generated files and `*_test.go`) must have a corresponding `_test.go` |
| Subprocess cancellation | All runner functions use `exec.CommandContext`. Pressing `q` in TUI cancels the running process |
