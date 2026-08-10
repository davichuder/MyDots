## Exploration: phase-5-tui

### Current State
The repository has completed the install pipeline and Phase 4 shell scripts, but the TUI layer is not implemented. `main.go` is currently an empty `main` function, while `go.mod` already contains Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh, teatest, and golden dependencies (currently recorded as indirect dependencies). There are no `internal/tui`, screen, component, or TUI test files yet.

The existing pipeline exposes the integration boundary needed by the TUI: `installer.Run` emits `ProgressEvent` values (`ModuleID`, `Status`, `LogLine`, and `Err`) on a channel and closes it when processing ends. The executor emits the literal status `"running"`, although `InstallStatus` currently defines only terminal statuses. `InstallContext.Log` is an `io.Writer` intended for streamed output. Configuration provides `DefaultConfig`, `Load`, `Save`, `Validate`, and `DefaultConfigPath`; backup provides `BackupFile`, `BackupDir`, `ListBackups`, and `DeleteBackup`, all using the process home directory abstraction.

The design documents define a `internal/tui` package with root routing, seven screens, and two reusable components. Bubble Tea v2 conventions are explicitly documented: `tea.KeyPressMsg`, `tea.View` from `View()`, `tea.Cmd`, and screen transitions through `ChangeScreenMsg`. The specifications define the five-item main menu, six-step config wizard, progress/log behavior, result variants, reference menu, and backup menu. No Phase 5 implementation or proposal exists yet.

### Affected Areas
- `internal/tui/app.go` — new root model, screen enum, routing, platform/config/session dependencies, and `ChangeScreenMsg` handling for T-093/T-094.
- `internal/tui/components/log_pane.go` — new bounded 200-line component for T-081/T-082.
- `internal/tui/components/progress_row.go` — new status-to-icon rendering component and golden coverage for T-083/T-084.
- `internal/tui/screens/main_menu.go` — new five-item menu, config existence/error handling, and quit command for T-085/T-086.
- `internal/tui/screens/config_menu.go` — new Huh-backed six-step form, defaults, validation, and config persistence for T-087/T-088.
- `internal/tui/screens/install_screen.go` — new executor-channel bridge, row updates, log handling, cheatsheet toggle, cancellation, and result transition for T-089/T-090.
- `internal/tui/screens/result_screen.go` — new success, warning, critical-failure, and WSL2-note rendering for T-091/T-092.
- `internal/tui/screens/backup_menu.go` — new managed-config backup listing/deletion/empty state for T-095/T-096.
- `internal/tui/screens/preflight.go` and `internal/tui/screens/reference_menu.go` — new preflight/guide and embedded-cheatsheet flows required by T-093/T-094.
- `main.go` — remains outside the Phase 5 implementation boundary; T-094 should provide the root TUI entry seam without implementing Phase 7 flag/routing work.
- `internal/installer/executor.go` and `internal/installer/types/types.go` — consumed as-is by the TUI; the existing running status and module-name APIs must be represented without reopening Phase 4.
- `internal/config/config.go` and `internal/backup/backup.go` — existing filesystem boundaries must be injected or wrapped in tests so no test uses real HOME or host state.
- `assets.go` and `assets/wsl2-guide.md` — existing embedded filesystem is the source for preflight/reference content; Phase 6 cheatsheet creation is out of scope.
- `docs/tasks.md`, `docs/specs.md`, `docs/design.md`, `docs/adr/ADR-001-tui-framework-bubbletea-v2.md` — normative task, screen, architecture, and framework references.

### Approaches
1. **Layered TUI package with direct model tests and one root teatest flow** — implement components and screens under `internal/tui`, keep screen models independently constructible, inject config/backup/platform/assets/executor dependencies at boundaries, and use direct `Update`/`View` tests for T-081/T-092/T-095 plus teatest only for the T-093 full-program scenarios.
   - Pros: matches the documented architecture and strict testing gates; keeps unit tests fast and deterministic; isolates filesystem, platform, and channel behavior; supports golden files at component, screen, and full-program boundaries.
   - Cons: requires explicit constructors/interfaces before UI code; Huh integration may need a test seam rather than testing terminal interaction directly.
   - Effort: High

2. **Single monolithic root model with embedded screen state** — place routing and all screen state in one `App` model and test most behavior through the root.
   - Pros: fewer constructors and simpler initial wiring.
   - Cons: couples unrelated screens, makes direct state tests harder, increases golden/test fixture scope, and conflicts with the design's sub-model routing and unit-test requirement.
   - Effort: High

### Recommendation
Use the layered TUI package with direct model tests and a narrow teatest integration layer. Execute each TDD pair strictly as RED then GREEN, in task order, with unit tests manipulating models directly. Use `teatest` only for T-093's full-program startup/exit behavior and use deterministic golden files for the listed renders. Introduce small interfaces or injected functions for config persistence, backup operations, platform/preflight input, embedded assets, and cancellation so tests use `t.TempDir`, synthetic platforms, and in-memory assets rather than real HOME, root, network, or host state. Keep Phase 5 changes confined to T-081 through T-096; do not implement the Phase 7 entry-point contract or add Phase 6 cheatsheets.

The implementation plan should explicitly resolve two existing contract seams without changing Phase 4: represent the executor's literal `"running"` state in the progress renderer without assuming it is a declared terminal constant, and define how the backup menu receives the managed-path set/session timestamp because `backup` currently exposes operations but no TUI-specific managed-path coordinator. Future commits should remain one conventional commit per RED/GREEN pair and include the decision reason in the commit body as requested.

### Risks
- Bubble Tea v2 APIs are already selected but TUI code is absent; incorrect v1/v2 imports or `View()` signatures could make the whole package fail to compile.
- The executor emits `"running"` as an untyped status while the public status constants omit it; status mapping must remain compatible with the current pipeline and avoid reopening Phase 4.
- Huh form behavior is terminal-oriented; tests must verify model state and persistence through deterministic seams rather than real terminal interaction.
- `DefaultConfigPath`, backup roots, and other existing functions resolve the host home through package-level functions; direct use in tests risks violating the no-real-HOME requirement unless dependency injection or controlled wrappers are added.
- T-093 asks for Darwin and Windows startup behavior, but Windows preflight must be simulated through injected GOOS/platform inputs; tests must not depend on the host OS or invoke a real terminal.
- Phase 6 cheatsheets are not yet in scope, so reference/install cheatsheet tests need fixture content or existing embedded assets without creating Phase 6 files.
- The current `main.go` is empty; T-094 must provide a testable root-app seam without accidentally implementing T-098/T-099/T-100.
- Golden output can vary with terminal dimensions, Unicode width, colors, and Bubble Tea renderer behavior; fixed dimensions, deterministic styles, and the repository's golden update path are required.

### Ready for Proposal
Yes — the repository, task boundaries, existing pipeline contracts, TUI architecture, and deterministic testing strategy are sufficiently understood for a Phase 5 proposal. The proposal should preserve the T-081–T-096-only boundary and call out the running-status and managed-backup-path seams before implementation begins.
