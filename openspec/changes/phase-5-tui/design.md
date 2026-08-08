# Design: Phase 5 Interactive TUI

## Technical Approach

Implement T-081–T-096 as Bubble Tea v2 models in `internal/tui`. `App` routes typed outcomes; commands own I/O; updates are deterministic. Inject config, backup, platform, assets, installation, elevation, terminal, clock, and render seams. `main.go`, Phase 4 contracts, Phase 6 assets, and Phase 7 behavior remain unchanged.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Layered root/screens/components | More constructors | `App` routes; screens own state and emit typed navigation messages. |
| Phase-5 session adapter | Local orchestration | `SessionRunner` owns the module loop and cancellation boundary without modifying Phase 4. |
| One-event command | Serial consumption | `waitForProgress` receives once and is re-issued after `Update`. |
| Local status mapping | Mapping layer | Render literal `running`; collapse skip variants; retain pending/disabled locally. |
| Huh v2 form | Dependency migration | Use `charm.land/huh/v2` and delegate its form lifecycle. |
| Narrow teatest seam | Limited integral coverage | `NewApp(Dependencies)` covers Darwin startup and Windows guide/exit only. |

## Data Flow

```text
Preflight -> Main -> Config / Install / Backup / Reference / Quit
WindowsGuide -> keypress -> Quit
Install -> terminal handoff + elevation -> SessionRunner -> rows/log/result
ConfigStore / BackupService / fs.FS -> typed messages -> screen
```

`App` owns the active model and factories. Install validates saved config: missing remains inline; malformed routes to Config with the original error and explicit reset defaults. Windows shows embedded `assets/wsl2-guide.md`, exits on input, and creates no install session. Unsupported platforms retain an app error for Phase 7 handling. `screens` owns `Screen` and `ChangeScreenMsg{Screen, Payload}`; commands return typed messages and never mutate model state. Reference records its previous screen.

`startInstall` creates one Phase-5 session context and is its sole cancellation owner. Cancellation marks the session, invokes its cancel function, stops elevation, and schedules no further runner step. The production `SessionRunner` builds the existing `installer.BuildPlan`, then locally iterates exported `installer.Module` values. Before inspecting or invoking each module it checks `ctx.Err()`; it passes that same context to `Install`. Therefore cancellation between non-critical modules prevents every later module from starting or mutating the host, while the currently running subprocess receives cancellation. The adapter does not call `installer.Run` and does not change Phase 4. It emits typed progress, log, and one completion event.

## File Changes

| File | Action | Description |
|---|---|---|
| `go.mod`, `go.sum` | Modify | Adopt Huh v2. |
| `internal/tui/{app,dependencies,install_adapter}.go` | Create | Routing, session/elevation adapter, contracts, render policy. |
| `internal/tui/components/{log_pane,progress_row}.go` | Create | Bounded logs and status rendering. |
| `internal/tui/screens/{preflight,main_menu,config_menu,install_screen,result_screen,reference_menu,backup_menu}.go` | Create | Seven flows. |
| Matching tests and goldens | Create | RED tests and deterministic renders. |
| `main.go`, `assets/**`, `internal/{installer,backup,config}/**` | Unchanged | Phase boundary. |

## Interfaces / Contracts

```go
type Dependencies struct {
    Config ConfigStore; Backup BackupService; Assets fs.FS
    Runner SessionRunner; Elevation Elevation; Terminal TerminalHandoff
    Clock Clock; Render RenderPolicy
}
type SessionRunner interface { Start(context.Context, InstallRequest) <-chan InstallEvent }
type Elevation interface { Acquire(context.Context) (Keepalive, error) }
type Keepalive interface { Stop() } // cancels and joins
type TerminalHandoff interface { Release() error; Restore() error }
type Clock interface { Every(time.Duration) Ticker }
type InstallEvent struct { Kind EventKind; Progress installer.ProgressEvent; Err error }
type elevationFailedMsg struct{ Err error }
type progressMsg struct{ Event installer.ProgressEvent }
type installDoneMsg struct{ Cancelled bool; Err error }
```

The production elevation adapter releases Bubble Tea's raw terminal, runs `sudo -v` on inherited terminal streams, and defers restore on success, failure, cancellation, and completion. It starts the runner only after successful restore. It uses the injected clock for one 45-second keepalive; each tick runs `sudo -v` with discarded output. `Stop` cancels and joins that lifecycle before `installDoneMsg`. Cancellation during handoff suppresses runner start; cancellation during execution binds the runner context and stops keepalive. Result precedence is cancellation, critical failure, warnings, success.

`ManagedPaths` stays authoritative and pre-resolved; nested targets deduplicate. `t.TempDir()` paths and `List()` support prior sessions; deletion requires selected timestamp and confirmation. Injected `fs.FS` renders existing references and returns `not available` for absent Phase 6 content.

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | Routing, components, elevation, cancellation, outcomes, backup/reference | RED before GREEN; direct `Update`/`View`, `fstest.MapFS`, and `t.TempDir()`. Fake runner records starts/writes; fake elevation, terminal, and clock script acquire, release/restore, ticks, stop/join, and cancellation. Assert cancellation before/between modules starts no later module and makes no later writes. |
| Golden | Rows and screens | Fixed 80x24, no color, fixed spinner, explicit `WindowSizeMsg`; update then rerun normally. |
| Integral | Darwin menu; Windows guide/exit | `teatest` only. |

Run narrow package tests, then `go test ./...`. No real commands, HOME, platform, network, root, timers, or mutable global styles.

## Migration / Rollout

No migration; Huh v2 only.

## Open Questions

None.
