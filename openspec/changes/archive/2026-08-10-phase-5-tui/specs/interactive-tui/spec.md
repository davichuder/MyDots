# Interactive TUI Specification

## Purpose

Define interactive TUI behavior for T-081–T-096.

## Requirements

### Requirement: Bounded feedback

The TUI MUST ignore empty logs and retain the newest 200 lines. Rows MUST show module names and map `pending`, `running`, `installed`, `failed`, `skipped`, and `disabled` to `—`, spinner, `✅`, `❌`, `⏭`, and `—`.

#### Scenario: Full log
- GIVEN 200 lines
- WHEN one arrives
- THEN the oldest MUST drop and the new line remain

#### Scenario: Status render
- GIVEN a supported status
- WHEN its row renders
- THEN its name and indicator MUST appear

### Requirement: Main menu and configuration

The menu MUST contain Config, Install, Backup, Reference, and Quit. Quit MUST exit successfully. Config MUST collect six validated choices: five-option font, theme, optional languages, Neovim config, Neovim framework, and HTTPS chezmoi URL. Defaults MUST be preselected; valid completion MUST persist before installation.

#### Scenario: Complete configuration
- GIVEN six valid choices
- WHEN submitted
- THEN they MUST persist before returning to the menu

#### Scenario: Missing configuration
- GIVEN no saved config
- WHEN Install is selected
- THEN an inline error MUST appear without leaving the menu

#### Scenario: Unloadable configuration
- GIVEN invalid or unloadable saved config
- WHEN Install is selected
- THEN an error MUST appear and assistance MUST open without silent defaults

### Requirement: Installation and outcomes

Install MUST update matching rows, append logs, toggle available current-module reference with `?`, consume through completion, then show results. Results MUST distinguish success, warnings, critical failure, and cancellation; identify relevant counts or modules; and show the WSL2 font note on WSL2. Cancellation MUST NOT appear as failure and SHOULD explain known progress.

#### Scenario: Event completion
- GIVEN installation events
- WHEN status, log, and completion arrive
- THEN feedback MUST update before results open

#### Scenario: Warnings
- GIVEN non-critical failures
- WHEN results render
- THEN warning count and failed modules MUST appear

#### Scenario: Cancellation
- GIVEN partial progress
- WHEN cancellation occurs
- THEN cancellation and known progress MUST appear, not failure

### Requirement: Root, preflight, and references

The root MUST route screen outcomes. Preflight MUST send Linux/Darwin to the menu, show the embedded WSL2 guide on Windows then exit after input without installing, and error on unsupported platforms. Reference MUST render embedded content; absent Phase 6 content MUST report `not available` without creating Phase 6 artifacts.

#### Scenario: Darwin startup
- GIVEN Darwin
- WHEN the full app starts
- THEN the menu MUST appear

#### Scenario: Windows startup
- GIVEN Windows
- WHEN the app starts
- THEN the WSL2 guide MUST appear and input MUST exit without installing

#### Scenario: Missing reference
- GIVEN missing embedded content
- WHEN selected
- THEN `not available` MUST appear

### Requirement: Managed backups

Backup MUST include all managed paths, list every detected MyDots-managed backup including prior sessions, confirm deletion, delete only the selection, and show an empty state.

#### Scenario: Create
- GIVEN managed paths
- WHEN Create is selected
- THEN all MUST be backed up

#### Scenario: Prior-session deletion
- GIVEN prior-session backups
- WHEN one deletion is confirmed
- THEN only that backup MUST be deleted

#### Scenario: Empty state
- GIVEN no managed backups
- WHEN Backup opens
- THEN an empty state MUST appear

### Requirement: Verification and boundaries

Every behavior MUST have a failing RED test before GREEN. Unit behavior MUST use direct model manipulation; the approved native Bubble Tea v2 program harness MUST cover integral full-program behavior; visual regression MUST use fixed deterministic goldens. Tests MUST NOT depend on network, root, real HOME, or host state. Phase 5 MUST NOT broadly change Phase 4, create Phase 6 content, or define Phase 7 CLI/entry-point flags. The sole approved Phase 5/Phase 4 integration seam is `internal/installer/executor.go`: canonical `Run` MUST check cancellation before each next module and carry `Module.Name()` with the typed `ModuleID` in progress events. The Phase 5 adapter MUST delegate to canonical `Run`, render the human name, and retain typed ID routing/status. Future commits MUST be conventional, T-xxx-aligned, and explain decisions.

#### Scenario: Isolated proof
- GIVEN Phase 5 behavior
- WHEN tested in isolation
- THEN it MUST be proven without host dependencies
