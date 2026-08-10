# Design: Phase 6–7 Cheatsheets and Entrypoint

## Technical Approach

Deliver only T-097–T-100 as strict-TDD work units: validate and add 48 embedded references, then implement deterministic routes behind the mandated top-level `run(args []string, goos string) int`. `main` only calls `os.Exit(run(os.Args, runtime.GOOS))`; existing config, installer, and Phase 5 TUI boundaries are composed, not redesigned.

## Architecture Decisions

| Option | Tradeoff | Decision and rationale |
|---|---|---|
| Count Markdown files | Cannot detect substitutions or filler | `assets_test.go` compares the exact sorted slug inventory and reads through `assets`. It slices bodies between the three headings: Links requires an HTTP(S) URL, Key Shortcuts requires a concrete non-placeholder action/key bullet, and Usage Examples requires a command/code example. |
| Test an internal method | Cleaner DI but violates T-098 | Tests call top-level `run`. It constructs dependencies through one package factory. A test-only helper locks a mutex, swaps the factory, invokes `run`, and restores it with `defer`; all run tests use this helper and never `t.Parallel`. Thus overrides are serialized, panic/`Goexit` safe, and cannot leak between tests. Production never mutates the factory. |
| Cancel-only keepalive | Cannot prove goroutine completion | Change `sudo.StartKeepalive` to return `(sudo.Keepalive, error)`; `Keepalive.Stop()` is idempotent, cancels, and blocks on a `done` channel after ticker stop. A pre-cancelled context/start failure returns no handle and prevents pipeline startup. Existing statement calls may discard results. |
| Broad TUI refactor | Unnecessary Phase 5 churn | `entrypoint.go` owns thin production adapters and composes existing constructors. This keeps Phase 5 artifacts untouched and limits internal changes to the keepalive contract and reference lookup. |

## Data Flow

```text
main -> run -> FlagSet(ContinueOnError, injected stderr) -> route
 version -> stdout,0       Windows -> guide,0       unsupported -> stderr,1
 no flags -> composed TUI
 unattended/default -> config -> elevation -> keepalive -> pipeline -> Stop/join -> return
```

`--version` precedes platform work. Windows rendering is a total operation (embedded guide or fallback) and always returns 0. Parse/help diagnostics use the injected writer. Missing unattended config emits only FR-22. The privileged helper owns the handle immediately after successful start and executes `defer keepalive.Stop()` before calling pipeline, covering every post-start return.

## File Changes

| File | Action | Description |
|---|---|---|
| `assets_test.go`; `assets/cheatsheets/*.md` | Create | Exact inventory and section-specific SCR-07 content. |
| `main.go`; `entrypoint.go`; `entrypoint_test.go` | Modify/Create | `buildVersion`, routes, DI, production adapters, direct-run tests. |
| `internal/sudo/{sudo.go,sudo_test.go}` | Modify | Joinable keepalive handle and timer-free lifecycle tests. |
| `internal/tui/screens/{install_screen.go,reference_slug.go,reference_slug_test.go}` | Modify/Create | Map all 48 `ModuleID` values to slug filenames. |

## Interfaces / Contracts

`entrypointDependencies` contains writers, assets, config path/load/save/default/validate, detector, guide/TUI/pipeline runners, elevation, `startKeepalive(context.Context) (keepaliveHandle, error)`, context factory, and `now() time.Time`. A start error or nil handle is fatal and skips pipeline. Production defines `var buildVersion = "dev"`; releases populate it with `go build -ldflags "-X main.buildVersion=<version>"`.

The production pipeline builds `installer.BuildPlan`, drains `installer.Run`, and creates one UTC `SessionTimestamp` from injected `now`, formatted `2006-01-02T15-04-05Z`. `runTUI` resolves home/config/backup paths once and builds `tui.NewApp`: `fileConfigStore.Load` delegates to `config.Load`; `BackupFactory` composes `screens.NewFileBackupStore`/`NewBackupMenu`; `InstallFactory` composes `NewModuleSessionRunner`/`NewInstallScreen`. An entrypoint-local `programTerminal` delegates release/restore to the bound `tea.Program`, `systemClock` adapts `time.Ticker`, and `commandExecutor` uses `exec.CommandContext`; these feed `screens.NewSudoElevation`. The program is bound before `Run`. Tests replace outer `runTUI`, never host adapters.

## Testing Strategy

| Layer | Proof | Approach |
|---|---|---|
| Unit | Exact inventory, section bodies, 48 ID→slug mappings | Table-driven embedded-FS tests. |
| Unit | Every route, exit, output, config, fixed timestamp, omissions | Top-level `run` with buffers, fakes, `fstest.MapFS`, and `t.TempDir()`. |
| Unit | elevation→keepalive→pipeline→shutdown; start failure; joined return | Recording fake; `Stop` blocks on a test gate, proving `run` cannot return before shutdown. No timers, processes, root, terminal, HOME, `/proc`, or network. |

## Work Units and Rollout

1. T-097 RED→GREEN: inventory test and documents.
2. T-098 RED→T-099 GREEN: direct-run route suite and production composition.
3. T-100 RED→GREEN only if unit 2 lacks ordering/join evidence.

Commit each coherent test+production unit. Budget: 400 modified lines / 800 new-file-predominant lines. If 48 concise references force an overage, report `size:exception` only for unit 1; do not hide entrypoint growth. No migration required.

## Open Questions

None.
