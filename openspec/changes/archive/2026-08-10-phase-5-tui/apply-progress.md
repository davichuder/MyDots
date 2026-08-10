# Apply Progress: Phase 5 Interactive TUI

## Delivery Context

- Mode: Strict TDD
- Delivery strategy: exception-ok
- Maintainer approval: `size:exception`
- Current work unit: T-091/T-092 — independently verified and finalized.
- PR/commit operations: none authorized or performed.
- Task markers: T-087/T-088 checked in both task trackers after independent verification returned `safe_to_mark_complete: true`.
- Status: T-081 through T-092 verified; T-091/T-092 are complete. No commit was performed.

## Completed Tasks

- [x] T-081 [RED] Add direct `LogPane` tests for empty logs and 200-line eviction.
- [x] T-082 [GREEN] Add the bounded `LogPane` implementation.
- [x] T-083 [RED] Add direct `ProgressRow` status-mapping tests and deterministic row goldens.
- [x] T-084 [GREEN] Add deterministic `ProgressRow` rendering.
- [x] T-085 [RED] Add direct main-menu tests and deterministic golden rendering.
- [x] T-086 [GREEN] Implement the main menu with an injected `ConfigStore`.
- [x] T-087 [RED] Add direct config-menu tests for the Huh v2 adapter lifecycle, defaults, HTTPS validation, persistence, and deterministic goldens. **Independently verified.**
- [x] T-088 [GREEN] Implement the Huh v2 configuration menu and persist valid configuration before routing to the main menu. **Independently verified.**
- [x] T-089 [RED] Add direct installation-screen and adapter tests for events, reference toggling, cancellation, requeueing, session lifecycle, elevation, sudo validation, and fake-clock keepalive. **Independently verified.**
- [x] T-090 [GREEN] Implement the cancellable installation screen and adapter with typed seams, raw-terminal handoff, and deterministic runner/elevation/clock fakes. **Independently verified.**
- [x] T-091 [RED] Add fixed result-screen goldens for success, warnings, critical failure, cancellation progress, and the WSL2 font note. **Independently verified.**
- [x] T-092 [GREEN] Implement the result screen with cancellation > critical > warnings > success precedence and main-menu key routing. **Independently verified.**

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-081 | `internal/tui/components/log_pane_test.go` | Unit | N/A (new files) | ✅ `rtk go test ./internal/tui/components -run '^TestLogPaneAppend$' -count=1` failed to build because `LogPane`, `LogPaneMaxLines`, and `NewLogPane` were undefined. | ➡️ Deferred to T-082. | ✅ Empty input and 201-line eviction. | ➖ None needed. |
| T-082 | `internal/tui/components/log_pane_test.go` | Unit | N/A (new files) | ✅ Uses T-081 RED. | ✅ Narrow test passed: 3 tests. | ✅ Empty and capacity paths. | ➖ None needed. |
| T-083 | `internal/tui/components/progress_row_test.go` | Unit + golden | N/A (new files) | ✅ `NewProgressRow` was undefined. | ➡️ Deferred to T-084. | ✅ Six status paths and goldens. | ➖ None needed. |
| T-084 | `internal/tui/components/progress_row_test.go` | Unit + golden | N/A (new files) | ✅ Uses T-083 RED. | ✅ Narrow test passed: 14 tests. | ✅ Direct mapping and goldens. | ➖ None needed. |
| T-085 | `internal/tui/screens/main_menu_test.go` | Unit + golden | N/A (new files) | ✅ `rtk go test ./internal/tui/screens -run '^TestMainMenu' -count=1` failed to build because main-menu symbols were undefined. | ➡️ Deferred to T-086. | ✅ Five choices, config branches, quit, and a fixed 80x24 golden. | ➖ None needed. |
| T-086 | `internal/tui/screens/main_menu_test.go` | Unit + golden | N/A (new files) | ✅ Uses T-085 RED. | ✅ Narrow test and normal golden rerun passed: 8 tests. | ✅ Missing, unloadable, invalid, valid, static route, and quit branches. | ➖ None needed. |
| T-087 | `internal/tui/screens/config_menu_test.go` | Unit + golden | ✅ `rtk go test ./internal/tui/screens -count=1` passed: 12 tests before modifications. | ✅ `rtk go test ./internal/tui/screens -run '^TestConfigMenu' -count=1` failed to build: `no required module provides package charm.land/huh/v2`; no T-088 dependency or implementation existed. | ➡️ Deferred to T-088; final independent verification confirmed this genuine RED against HEAD `93ad1d1`. | ✅ Huh v2 form lifecycle, six defaults, five font options, HTTP rejection without completion, six-step temp-directory persistence, and six fixed 80x24/no-color goldens. | ➖ None needed; command-message consumption is extracted only in the test harness to model Huh's adapter lifecycle deterministically. |
| T-088 | `internal/tui/screens/config_menu_test.go` | Unit + golden | ✅ T-087 safety net. | ✅ Uses T-087 RED. | ✅ Initial narrow suite passed 11 tests and full suite 694 tests; final persistence-regression GREEN passed 3 tests, focused config suite passed 14 tests, and `go test ./...` passed 697 tests in 15 packages. | ✅ Six distinct form groups, valid and invalid URL paths, persisted file path, every fixed-size screen step, and Java-only plus Java+PHP persistence. | ✅ ANSI stripping makes the adapter output deterministic/no-color while preserving delegated Huh lifecycle; stable form-owned bindings fixed copied-config and language persistence. |

## Test Results

- Safety net: `rtk go test ./internal/tui/screens -count=1` — passed, 12 tests in 1 package before T-087/T-088 modifications.
- RED: `rtk go test ./internal/tui/screens -run '^TestConfigMenu' -count=1` — failed as expected (0 passed, 1 package failed): `no required module provides package charm.land/huh/v2`.
- Golden update: `rtk go test ./internal/tui/screens -run '^TestConfigMenuStepGoldens$' -count=1 -update` — passed, 7 tests in 1 package; created six fixtures.
- Narrow GREEN rerun: `rtk go test ./internal/tui/screens -run '^TestConfigMenu' -count=1` — passed, 11 tests in 1 package.
- Full suite: `rtk go test ./...` — passed, 694 tests in 15 packages.
- Host I/O: none; persistence uses `t.TempDir()` and every form is driven directly with Bubble Tea v2 messages.

## Judgment Day Round 1 Fix Evidence

- Task markers were checked only after independent verification and Judgment Day Round 2 approval.
- Safety net: `rtk go test ./internal/tui/screens` — passed, 23 tests before the regression test.
- RED: `rtk go test ./internal/tui/screens -run '^TestConfigMenuPersistsChangedChoicesThroughHuhLifecycle$' -count=1 -v` — failed as expected: the persisted config retained defaults because Huh accessors targeted copied model state, and languages had no bound value.
- GREEN: the same focused command passed 3 tests after binding Huh fields to stable form-owned values and converting selected language IDs before persistence.
- Triangulation: real Bubble Tea/Huh key lifecycle proves both Java-only and Java+PHP selections, plus non-default font, theme, Neovim config/framework, and HTTPS URL persistence.
- Focused suite: `rtk go test ./internal/tui/screens -run '^TestConfigMenu' -count=1 -v` — passed, 14 tests.
- Full suite: `rtk go test ./...` — passed, 697 tests in 15 packages.
- Host I/O: none; every persisted path is under `t.TempDir()`.

## Files Changed

| File | Action | Purpose |
|---|---|---|
| `go.mod`, `go.sum` | Modified | Adopt `charm.land/huh/v2` v2.0.3 and its ordered dependency. |
| `internal/tui/screens/config_menu_test.go` | Created | Direct Huh v2 lifecycle, defaults, validation, persistence, and golden coverage. |
| `internal/tui/screens/config_menu.go` | Created | Six-step Huh v2 adapter that validates, saves, and routes only valid configuration. |
| `internal/tui/screens/main_menu.go` | Modified | Add `ScreenMain` navigation target for configuration completion. |
| `internal/tui/screens/testdata/config_menu/*.golden` | Created | Six fixed 80x24, no-color adapter render fixtures. |
| `openspec/changes/phase-5-tui/apply-progress.md` | Modified | Merge cumulative TDD evidence and final independent-verification record. |

## Deviations and Issues

- Deviations: None — implementation matches the approved design and specification.
- Issues: The Huh v2 adapter emits internal chained command messages (`nextFieldMsg` then `nextGroupMsg`); the direct test helper consumes them to prove real lifecycle state rather than bypassing the adapter.

## Remaining Tasks

- [ ] T-089 through T-096.

## Final Independent Verification

- Verdict: **PASS WITH WARNINGS**; `safe_to_mark_complete: true`.
- Judgment Day Round 2: **APPROVED**. The canonical merged copied-config and unbound-language persistence findings were both fixed and verified.
- Focused config suite: 14 passed; persistence regression: 3 passed, including `-race`; six golden fixtures passed twice without hash drift.
- Final full suite: `go test ./...` — **697 passed in 15 packages**.
- Non-blocking warning: `go mod tidy -diff` proposes a project-wide manifest rewrite. Dependency resolution and `go mod verify` passed; module-hygiene cleanup remains separate.
- Final task state: T-081 through T-088 are complete; T-087/T-088 are verified / ready for commit. No commit was performed.

## Workload Boundary

- `size:exception` was explicitly approved for the Phase 5 pair work units.
- This batch starts with the T-087 Huh v2 dependency RED and ends with six-step adapter persistence, deterministic goldens, and full-suite verification.
- Rollback boundary: remove the Huh v2 dependency, `config_menu.go`, its tests and fixtures, and `ScreenMain`; no Phase 4, 6, or 7 code changed.

## T-089/T-090 — Independent Verification Complete

- Task markers: checked in both task trackers after independent verification returned `safe_to_mark_complete: true`.
- Scope: only the installation screen and adapter pair; no Phase 4, 6, or 7 code changed.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-089 | `internal/tui/screens/{install_screen,install_adapter}_test.go` | Unit | N/A (new files); existing screen safety net passed: 26 tests. | ✅ `rtk go test ./internal/tui/screens -run '^(TestInstallScreen|TestModuleSessionRunner|TestSudoElevation)' -count=1 -v` failed to build because `InstallEvent`, `InstallRequest`, `Keepalive`, `NewModuleSessionRunner`, and `Ticker` were undefined. | ➡️ Deferred to T-090. | ✅ Direct screen events, reference toggle, cancellation/no-requeue, cancellation between modules, and sudo success/failure/keepalive paths. | ➖ None needed. |
| T-090 | `internal/tui/screens/{install_screen,install_adapter}_test.go` | Unit | ✅ T-089 RED; existing screen package passed 26 tests before new files. | ✅ Uses T-089 RED; `q` cancellation and `sudo -v` command validation each failed first, then passed after minimal additions. | ✅ Focused 9 tests passed; full `go test ./...` passed 706 tests in 15 packages. | ✅ Success completion and cancellation paths; reference toggles both directions; terminal success/failure; explicit `sudo -v`; one 45-second fake tick and joined shutdown. | ✅ `finishSession` separates normal completion from cancellation so successful results are not labelled cancelled. |

### Test Results

- Safety net: `rtk go test ./internal/tui/screens -count=1` — passed, 26 tests in 1 package before adding the pair.
- RED: `rtk go test ./internal/tui/screens -run '^(TestInstallScreen|TestModuleSessionRunner|TestSudoElevation)' -count=1 -v` — failed as expected at build time because the install screen and adapter production symbols were undefined.
- GREEN: the same focused command — passed, 6 tests in 1 package.
- Triangulation RED/GREEN: `rtk go test ./internal/tui/screens -run '^TestInstallScreenCancellationStopsSessionAndDoesNotRequeue$' -count=1 -v` failed for `q` cancellation, then passed after adding the same cancellation lifecycle as `esc`.
- Triangulation RED/GREEN: `rtk go test ./internal/tui/screens -run '^TestCommandSudoValidatorRunsSudoValidationOnly$' -count=1 -v` failed because `NewCommandSudoValidator` was undefined, then passed after adding the narrow `sudo -v` command seam.
- Final focused/refactor command: `rtk go test ./internal/tui/screens -run '^(TestInstallScreen|TestModuleSessionRunner|TestSudoElevation|TestCommandSudoValidator)' -count=1 -v` — passed, 9 tests in 1 package.
- Full suite: `rtk go test ./...` — passed, 706 tests in 15 packages.
- Golden updates: none required; this pair introduces no rendered golden fixtures.
- Host I/O: none; all terminal, sudo, clock, runner, and module seams are deterministic fakes. No HOME, network, root, real sudo, clock sleep, or host state was used.

### Files Changed

| File | Action | Purpose |
|---|---|---|
| `internal/tui/screens/install_screen_test.go` | Created | Direct event, reference-toggle, cancellation, and one-event requeue coverage. |
| `internal/tui/screens/install_adapter_test.go` | Created | Deterministic runner cancellation and sudo elevation/clock lifecycle coverage. |
| `internal/tui/screens/install_screen.go` | Created | One-session Bubble Tea installation model with typed events and cancellation ownership. |
| `internal/tui/screens/install_adapter.go` | Created | Typed runner, terminal/elevation, sudo validation, and joined fake-clock seams. |
| `openspec/changes/phase-5-tui/apply-progress.md` | Modified | Merged cumulative TDD evidence, leaving task markers pending verification. |

### Current State

- T-081 through T-088 are complete and verified.
- T-089/T-090 are implemented, independently verified, and complete.
- Rollback boundary: remove the two install source files and their tests; no Phase 4, 6, or 7 code changed.

## Judgment Day Fix Round 1 Evidence — T-089/T-090

- Task markers remain deliberately unchecked in both task trackers; no commit, push, or PR operation was performed.
- RED: `rtk go test ./internal/tui/screens/... -run 'TestInstallScreenInitAndInitialSizeStartOneSession|TestInstallScreenCancellationCancelsInFlightElevation|TestModuleSessionRunnerPreservesInstallerRunSemantics' -count=1` failed as expected: Init returned a second startup command, elevation used an uncancellable context, and the local runner missed dependency, configured-state, and shared-session semantics.
- GREEN: `rtk go test ./internal/tui/screens/... -run 'TestInstallScreenInitAndInitialSizeStartOneSession|TestInstallScreenCancellationCancelsInFlightElevation|TestModuleSessionRunnerPreservesInstallerRunSemantics|TestModuleSessionRunnerCancellationPreventsLaterStartAndWrite' -count=1` — passed, 4 tests.
- Triangulation: existing Escape and `q` cancellation paths plus new in-flight acquisition cancellation; failed dependency, context-configured skip, and session BrewPath paths; and Init plus initial size prove the single startup owner.
- Focused suite: `rtk go test ./internal/tui/screens/... -count=1` — passed, 38 tests in 1 package.
- Executor suite: `rtk go test ./internal/installer/...` — passed, 522 tests in 4 packages.
- Full suite: `rtk go test ./...` — passed, 709 tests in 15 packages.
- Host I/O: none; all new coverage uses direct Bubble Tea messages and deterministic fake elevation, runner, and module seams.

## Critical Verify-Report Remediation — T-089/T-090

- Source: Engram verify-report #418 five CRITICAL findings only.
- Task markers: deliberately remain unchecked pending an independent verify rerun. No commit, push, PR, proposal, design, or T-091+ file was changed.
- Safety net: `rtk go test ./internal/tui/screens -count=1` — passed, 38 tests before this remediation.
- RED (result/reference/indicators): `rtk go test ./internal/tui/screens -run '^(TestInstallScreenConsumesOneEventThenRequeues|TestInstallScreenQuestionTogglesCurrentModuleReference|TestInstallScreenQuestionReportsUnavailableReference|TestInstallScreenDirectInstalledAndFailedEventsRenderMatchingIndicators)$' -count=1 -v` — failed to build because `ScreenResult` and `InstallResult` were undefined.
- GREEN: the same focused screen command — passed, 4 tests; final focused suite below covers the closed-channel triangulation.
- RED (clock/streams): `rtk go test ./internal/tui/screens -run '^(TestSudoElevationReleasesRestoresAndJoinsKeepalive|TestCommandSudoValidatorRunsSudoValidationOnly)$' -count=1 -v` — failed to build because `CommandStreams`, `Command`, and `ValidateQuiet` were undefined and the validator constructor lacked stream injection.
- GREEN: `rtk go test ./internal/tui/screens -run '^(TestInstallScreen|TestModuleSessionRunner|TestSudoElevation|TestCommandSudoValidator)' -count=1 -v` — passed, 15 tests.
- Triangulation: completion and closed-channel paths both emit `ChangeScreenMsg{Screen: ScreenResult, Payload: InstallResult}`; injected `M-05` cheatsheet renders while no running module/assets renders `Reference not available`; direct matching `installed` and `failed` events render `✅` and `❌`; the fake clock records an exact `45*time.Second`; initial validation receives injected interactive stdin/stdout/stderr and keepalive uses discarded stdout/stderr; the fake ticker publishes `time.Time{}` and waits through channels only.
- REFACTOR/race: the first focused race command exposed a test-fake channel publication race; pre-create the acknowledgement channel and retain joined shutdown. `rtk go test -race ./internal/tui/screens -run '^(TestInstallScreen|TestModuleSessionRunner|TestSudoElevation|TestCommandSudoValidator)' -count=1 -v` — passed, 15 tests.
- Relevant installer race: `rtk go test -race ./internal/installer/... -count=1` — passed, 448 tests in 4 packages.
- Full suite: `rtk go test ./... -count=1` — passed, 712 tests in 15 packages.
- Quality: `rtk go vet ./internal/tui/screens ./internal/installer/...` and `rtk git diff --check 7adfcc4` — passed.
- Host I/O: none; all commands, terminal streams, assets, clocks, runners, and modules remain injected fakes.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-089 | `internal/tui/screens/install_screen_test.go` | Direct model unit | ✅ 38 screen tests | ✅ Missing result contract compilation RED | ✅ 15 focused tests | ✅ completion/closed channel, embedded/unavailable reference, installed/failed rows | ✅ Deterministic result payload and reference lookup |
| T-090 | `internal/tui/screens/install_adapter_test.go` | Seam unit | ✅ 38 screen tests | ✅ Missing command/stream contract compilation RED | ✅ 15 focused tests | ✅ exact 45s plus interactive/quiet stream policies | ✅ Channel-only fake synchronization; focused race green |

### Design-Deviation Reconciliation

- The existing accepted installer-parity remediation calls canonical `installer.Run` and adds its cancellation-before-next-module guard in `internal/installer/executor.go`. This differs from the approved design text that describes a local module loop and no Phase-4 changes.
- This remediation does not expand that deviation: it only consumes the accepted runner behavior. `ScreenResult` and `InstallResult` are a minimal navigation/payload contract for the later T-091/T-092 result screen; no result screen was implemented.
- Proposal and design remain intentionally unchanged at the user's direction; the discrepancy is recorded here for the next formal artifact reconciliation.

## T-089/T-090 Record Finalization

- Independent verification verdict: **PASS WITH WARNINGS**; `safe_to_mark_complete: true`.
- Final full suite: `rtk go test ./... -count=1` — **712 passed in 15 packages**.
- Final JSON evidence: `go test -json ./... -count=1` — **712 passed, 0 failed, 20 existing platform/integration skips; 13 package passes, 0 package failures**.
- Canonical Judgment Day findings were confirmed as verified; all five critical verifier findings were remediated and verified.
- Remaining warning: stale design wording about the accepted installer-parity deviation is informational only; proposal and design intentionally remain unchanged.
- Task markers: T-089/T-090 are checked in `docs/tasks.md` and this OpenSpec task tracker. No code, test, commit, push, or PR operation was performed during finalization.

## Pre-commit R3-001 Remediation — T-089/T-090

- Scope: only the verified post-cancellation event-drain defect in the uncommitted T-089/T-090 slice. No task markers were changed.
- Safety net: `rtk go test ./internal/tui/screens/...` — passed before the regression test was added.
- RED: `rtk go test ./internal/tui/screens -run '^TestInstallScreenDrainsPostCancellationEventsUntilCompletion$' -count=1 -v` — failed as expected because a post-cancellation non-terminal event returned no requeue command, leaving `InstallDoneEvent` unread.
- GREEN: the same focused command — passed after non-terminal events always requeue `waitForProgress`, including after cancellation.
- Triangulation: retained `q`/Escape cancellation behavior and `TestModuleSessionRunnerCancellationPreventsLaterStartAndWrite`; the new direct-model scenario proves a non-terminal event is drained before the terminal completion transitions to `ScreenResult` without another runner start.
- Refactor: none needed; removed only the cancellation-specific early return that abandoned the existing event stream.
- Focused suite: `rtk go test ./internal/tui/screens -count=1` — passed, 42 tests in 1 package.
- Full suite: `rtk go test ./... -count=1` — passed, 713 tests in 15 packages.
- Formatting and diff check: `gofmt -d internal/tui/screens/install_screen.go internal/tui/screens/install_screen_test.go` and `git diff --check` — passed.

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-089/T-090 | `internal/tui/screens/install_screen_test.go` | Direct model unit | ✅ screen package passed | ✅ post-cancel progress left completion unread | ✅ focused regression passed | ✅ cancellation keys, no-later-start/write, and terminal result transition | ➖ Minimal early-return removal only |

## T-091/T-092 — Independent Verification Complete

- Task markers: checked in both task trackers after independent verification returned `safe_to_mark_complete: true`.
- Scope: only the result-screen pair; no Phase 4, 6, or 7 production file changed.
- Judgment Day Round 2: **APPROVED** — both canonical result-screen CRITICAL findings are verified.
- Final full suite: `go test ./... -count=1` — **727 tests in 15 packages passed**.
- No code, test, commit, push, or PR operation was performed during record finalization.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-091 | `internal/tui/screens/result_screen_test.go` | Direct model + golden | ✅ Existing TUI screen suite | ✅ Result-screen tests failed before production symbols existed | ➡️ Deferred to T-092 | ✅ Success, warnings, critical failure, cancellation progress, WSL2 note, and five deterministic fixtures | ➖ No additional refactor needed. |
| T-092 | `internal/tui/screens/result_screen_test.go`, `internal/tui/screens/install_adapter_test.go` | Direct model + seam | ✅ T-091 RED | ✅ Focused and race suites passed | ✅ Production critical propagation and every-key return to `ScreenMain` regressions failed before their fixes, then passed | ✅ Retained typed `InstallResult` error propagation and unconditional key routing. |

### Verification Evidence

- Focused behavior and focused race: **14 tests passed** each.
- Relevant race: **504 tests in 5 packages passed**.
- Golden repeatability: five fixtures passed twice with stable hashes.
- Build, `go vet`, scoped `gofmt -d`, and `git diff --check 5964e98` passed.
- Verdict: **PASS**; `safe_to_mark_complete: true`; **727 passed, 0 failed**.

## T-093/T-094 — Blocked at Required Teatest Feasibility Gate

- Scope attempted: T-093 RED only, before any T-094 production implementation.
- Safety net: `rtk go test ./internal/tui/... -count=1` — passed, 73 tests in 2 packages before adding the feasibility test.
- RED: added `internal/tui/app_test.go:TestTeatestSupportsBubbleTeaV2Models`, which runs the existing Bubble Tea v2 `screens.MainMenu` through `teatest.NewTestModel` at a fixed 80x24 size. `rtk go test ./internal/tui -run '^TestTeatestSupportsBubbleTeaV2Models$' -count=1 -v` failed during dependency compilation: `github.com/charmbracelet/x/cellbuf@v0.0.13` expects the old `github.com/charmbracelet/x/ansi` API, but minimal version selection resolves `x/ansi@v0.11.7` for Bubble Tea v2.
- Compatibility investigation: the configured `github.com/charmbracelet/x/exp/teatest@v0.0.0-20260705004817-2cc9a8fe1146` imports `github.com/charmbracelet/bubbletea@v1.3.5`, while Phase 5 uses `charm.land/bubbletea/v2@v2.0.8`. Their `tea.Model` contracts differ (`View() string` versus `View() tea.View`), so resolving the `cellbuf` version alone cannot make teatest exercise a v2 model.
- GREEN: blocked. No `app.go`, `preflight.go`, `reference_menu.go`, dependency changes, goldens, or T-095+ work were created. The new test deliberately remains as the required feasibility proof until an approved Bubble Tea v2-compatible integral-test harness is selected.
- Task markers: deliberately unchanged (`[ ]`) in `docs/tasks.md` and this OpenSpec tracker; independent verification has not occurred.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-093 | `internal/tui/app_test.go` | Integral feasibility | ✅ 73 TUI tests | ❌ Blocked: teatest v1 dependency/API is incompatible with Bubble Tea v2 before the test can run. | ⛔ Blocked pending approved v2-compatible harness. | ⛔ Darwin/Windows scenarios cannot be added until feasibility is real. | ➖ No production code written. |
| T-094 | N/A | N/A | N/A | ⛔ Depends on T-093 feasibility. | ⛔ Not started. | ⛔ Not started. | ⛔ Not started. |

### Current Boundary

- Completed and independently verified: T-081 through T-092.
- Pending and unchecked: T-093 through T-096.
- Delivery: scoped pair-pure `size:exception`; no commit, push, or PR.
- Host I/O: none. The blocked test uses only an existing in-memory model and fixed terminal dimensions.

## T-093/T-094 — Native Bubble Tea v2 Harness Implementation (Pending Independent Verification)

- Approved deviation: retain the genuine `x/exp/teatest` feasibility RED above as evidence, but replace the v1-only dependency/harness with Bubble Tea v2's native `tea.NewProgram` integration harness. The harness injects input, output, context, an 80x24 window, and `WithoutSignals`; no incompatible teatest, Bubble Tea v1, or cellbuf dependency remains.
- Task markers: deliberately remain unchecked (`[ ]`) in both task trackers until independent verification. T-095/T-096 were not edited.
- Windows safety: injected `InstallFactory` is never constructed for the Windows program path; the app selects `Preflight`, renders the embedded guide, and exits on the injected keypress.
- Reference safety: injected assets render available embedded Markdown; absent or unreadable Phase 6 references explicitly render `Reference not available` and do not create content.
- Host I/O: none. Tests use `fstest.MapFS`, in-memory streams, a bounded context, injected platform/assets, and no real HOME, network, root, timer, command, or installation.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-093 [RED] | `internal/tui/app_test.go`, `internal/tui/screens/preflight_test.go`, `internal/tui/screens/reference_menu_test.go` | Native-v2 program integration + direct-model unit | ✅ Prior blocked feasibility safety net: `rtk go test ./internal/tui/... -count=1` — 73 tests / 2 packages. | ✅ Preserved genuine feasibility RED: `TestTeatestSupportsBubbleTeaV2Models` could not compile because v1-only teatest resolves `x/cellbuf@v0.0.13` against Bubble Tea v2's `x/ansi@v0.11.7`; the model contracts are `View() string` (v1) vs `View() tea.View` (v2). Native-v2 behavior RED: `rtk go test ./internal/tui/... -run '^(TestProgram|TestReferenceMenu|TestPreflight)' -count=1 -v` produced 4 passes / 2 failures: `Dependencies.InstallFactory` was undefined and embedded reference content rendered `Reference not available`. | ➡️ Deferred to T-094. | ✅ Darwin main menu; Windows guide/key exit with zero install-factory calls; embedded and unavailable reference paths; embedded and missing guide paths. | ✅ Kept the program helper as an injected, deterministic boundary; production stays host-independent. |
| T-094 [GREEN] | Same files | Native-v2 program integration + direct-model unit | ✅ T-093 RED. | ✅ Uses T-093 RED. | ✅ Focused GREEN: `rtk go test ./internal/tui/... -run '^(TestProgram|TestReferenceMenu|TestPreflight)' -count=1 -v` — 7 passed / 3 packages. Race-relevant GREEN: `rtk go test -race ./internal/tui/... -run '^(TestProgram|TestReferenceMenu|TestPreflight)' -count=1 -v` — 7 passed / 3 packages. Full GREEN: `rtk go test ./... -count=1` — 734 passed / 16 packages. | ✅ Both platforms, guide asset present/missing, reference asset present/missing, and no Windows install construction. | ➖ No further refactor needed after extracting the injected install factory boundary. |

### Files Changed

| File | Action | Purpose |
|---|---|---|
| `go.mod`, `go.sum` | Modified | Remove incompatible Bubble Tea v1/teatest dependency graph and retain the Bubble Tea v2 build graph. |
| `internal/tui/app.go` | Created | Route root outcomes while constructing installation only through an injected factory after an explicit install route. |
| `internal/tui/dependencies.go` | Created | Inject platform, assets, config boundary, config path, and deferred installation factory. |
| `internal/tui/app_test.go` | Created | Native Bubble Tea v2 full-program Darwin and Windows behavior tests. |
| `internal/tui/screens/preflight.go` | Created | Render the injected WSL2 guide and quit on key input. |
| `internal/tui/screens/preflight_test.go` | Created | Direct embedded/missing guide behavior tests. |
| `internal/tui/screens/reference_menu.go` | Created | Render injected embedded Markdown or explicit unavailable feedback. |
| `internal/tui/screens/reference_menu_test.go` | Created | Direct embedded/unavailable reference and return navigation tests. |

### Deviations and Issues

- Approved deviation from the original T-093 wording/design: `x/exp/teatest` was proven Bubble Tea v1-only and cannot validly execute a Bubble Tea v2 model. Native Bubble Tea v2 `Program` integration is used instead.
- No other design deviation. Backup routing/implementation remains T-095/T-096 scope.

### Current Boundary

- Completed and independently verified: T-081 through T-092.
- Implemented but deliberately unchecked pending independent verification: T-093/T-094.
- Untouched and pending: T-095/T-096.
- Delivery: scoped pair-pure `size:exception`; no commit, push, or PR.

## T-093/T-094 — Verify-Report #418 Critical Remediation

- Scope: only the seven CRITICAL findings from verification report #418. Warnings were not changed unless inseparable from the corrected root flow.
- Task markers: deliberately remain unchecked (`[ ]`) in both task trackers pending a fresh independent verification. T-095+ remains untouched. No commit, push, or PR was performed.
- Retained approved deviation: the native Bubble Tea v2 `Program` harness remains in place; the Bubble Tea v1-only teatest feasibility RED remains valid and is not reopened.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-093 | `internal/tui/app_test.go` | Native-v2 full-program integration + direct root model + golden | ✅ `rtk go test ./internal/tui/... -count=1` — 81 tests / 3 packages before remediation. | ✅ Unsupported typed platform contract initially failed to compile; host `TEA_TRACE` regression created the host trace file when isolation was removed; golden fixture was initially absent. | ✅ Focused `TestProgram`/`TestApp` suite — 10 tests passed; explicit `-update` golden pass then normal golden rerun passed. | ✅ Darwin/Linux/Windows/unsupported startup; embedded Windows guide and no factory start; routed 80x24 install factory/start; direct ConfigIssue; host trace isolation. | ✅ Removed ambiguous `Windows` dependency; centralized typed platform selection and retained/replayed size state. |
| T-094 | `internal/tui/screens/reference_menu_test.go`, `internal/tui/screens/preflight_test.go` | Direct model unit + embedded-fs integration | ✅ T-093 safety net. | ✅ Multiple-reference test failed: list rendered the first cheat sheet and selection was unavailable. | ✅ Focused reference/preflight suite passed. | ✅ Available/missing guide; unavailable assets; two Markdown files, ordered list, and selected tool content. | ✅ Reference model now holds deterministic tool list/cursor/selection with unchanged unavailable state. |

### Verification Evidence

- Focused: `rtk go test ./internal/tui/... -run '^(TestProgram|TestApp|TestPreflight|TestReferenceMenu)' -count=1 -v` — **17 passed / 3 packages**.
- Focused race: same command with `-race` — **17 passed / 3 packages**.
- Relevant race: `rtk go test -race ./internal/tui/... -count=1` — **90 passed / 3 packages**.
- Full: `rtk go test ./... -count=1` — **744 passed / 16 packages**.
- Required golden workflow: `rtk go test ./internal/tui -run '^TestProgramMainMenuGolden$' -count=1 -update -v`, then the same command without `-update` — both passed; fixture `internal/tui/testdata/app/main_menu_program.golden` added.
- Quality: `go build ./...`, `go vet ./...`, `go mod verify`, `go mod tidy -diff`, scoped `gofmt -d`, and `git diff --check d58b781` — passed.

### Remediated Behaviors

1. `Dependencies.Platform` is the sole typed startup input: Darwin/Linux open Main Menu, Windows opens Preflight, and other values retain `UnsupportedPlatformError` and a clear rendered error.
2. `App` retains every `tea.WindowSizeMsg` and replays it to newly routed models, allowing the deferred installation screen to start after route construction.
3. The native program has a deterministic main-menu render fixture under fixed 80x24/no-color input/output conditions.
4. The native harness fixes Program environment (`TERM=dumb`, neutral `TEA_TRACE`), temporarily neutralizes Bubble Tea v2's direct host trace lookup, uses `WithoutSignals`, and uses `WithoutSignalHandler`; the regression proves host `TEA_TRACE` creates no file.
5. The Windows full-program test uses `fstest.MapFS` and asserts injected embedded WSL2 guide content while proving no installation factory invocation.
6. Reference Menu discovers all embedded Markdown tools, lists them before selection, renders the selected cheatsheet, and preserves the explicit unavailable state.
7. `App` routes direct `screens.ConfigIssue` messages from the Config Menu to a freshly rendered assistant carrying the original error, without loading, saving, or retrying configuration.

### Current Boundary

- Completed and independently verified: T-081 through T-092.
- Remediated but intentionally unchecked: T-093/T-094, pending independent verification.
- Untouched and pending: T-095/T-096.
- Rollback boundary: the root/preflight/reference files, their scoped tests, and the single program golden; no Phase 4, 6, or 7 production behavior changed.

## T-093/T-094 Record Finalization

- Independent verification verdict: **PASS WITH WARNINGS**; `safe_to_mark_complete: true`.
- Task markers: T-093/T-094 are checked in `docs/tasks.md` and this OpenSpec task tracker after the clean independent rerun.
- Final full suite: `go test -json ./... -count=1` — **744 passed, 0 failed, 20 existing skips, 14 package passes**.
- Approved deviation retained: Bubble Tea v1-only `x/exp/teatest` remains infeasible for Bubble Tea v2 models; the genuine feasibility RED is retained and the native Bubble Tea v2 `Program` harness is the approved replacement.
- Canonical ConfigIssue routing finding is fixed and verified: the root preserves the issue, the assistant renders the original error, and no load/save/retry loop occurs.
- Seven verifier remediations were verified: typed platform routing; retained/replayed window size; deterministic 80x24 program golden; trace/signal/environment isolation; embedded Windows guide with zero install construction; multi-reference listing/selection; and ConfigIssue propagation.
- Non-blocking warnings remain informational: `app.go` statement coverage is 79.5%, and an invalid install payload or nil `InstallFactory` retains/re-initializes the active model rather than rendering a dependency error.
- No code, test, commit, push, or PR operation was performed during record finalization. T-095/T-096 remain pending.

## T-095/T-096 — Strict TDD Implementation Pending Independent Verification

- Baseline: `b95935db024b0811f40b70fb6f03175b21b60464`.
- Delivery: pair-pure, maintainer-approved `size:exception`; no commit, push, or PR.
- Task markers: deliberately remain unchecked in this OpenSpec tracker and the Engram task record until independent verification.
- Scope: `BackupMenu`, its injected/pre-resolved filesystem store, one deterministic empty-state golden, and the minimal injected root route only.
- Host isolation: every test uses `t.TempDir()` roots and managed source paths. No HOME lookup, network, root, platform state, or host filesystem path is used.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-095 [RED] | `internal/tui/screens/backup_menu_test.go`, `internal/tui/app_test.go` | Direct Bubble Tea model + temp-dir filesystem + golden | ✅ `rtk go test ./internal/tui/... -count=1` — 90 passed / 3 packages before existing-file edits. | ✅ `rtk go test ./internal/tui/screens -run 'TestBackupMenu' -count=1` failed to build because `NewBackupMenu`, `NewFileBackupStore`, and model helpers were undefined. | ➡️ Deferred to T-096. | ✅ Prior-session listing, all pre-resolved file/directory paths, selected timestamp confirmation/deletion, empty state, and injected root route. | ➖ No additional refactor needed. |
| T-096 [GREEN] | Same | Direct Bubble Tea model + temp-dir filesystem + golden | ✅ T-095 RED and TUI safety net. | ✅ Uses T-095 RED. | ✅ `rtk go test ./internal/tui/screens -run 'TestBackupMenu' -count=1 -update` — 3 passed; normal rerun — 3 passed. | ✅ Create handles file and directory paths; list retains a prior session; `d` requires `y` and removes only the selected timestamp; deterministic 80x24 empty golden. | ✅ `BackupStore` keeps filesystem/path resolution injected and `FileBackupStore` rejects non-basename timestamps before removal. |

### Verification Evidence (Implementation Only)

- Golden workflow: `rtk go test ./internal/tui/screens -run 'TestBackupMenu' -count=1 -update` — 3 passed; then the same command without `-update` — 3 passed.
- TUI suite: `rtk go test ./internal/tui/... -count=1` — 94 passed / 3 packages.
- Relevant race suite: `rtk go test -race ./internal/tui/... -count=1` — 94 passed / 3 packages.
- Full suite: `rtk go test ./... -count=1` — 748 passed / 16 packages.
- Formatting/diff: `gofmt -d` and `git diff --check` passed.

### Files Changed

| File | Action | Purpose |
|---|---|---|
| `internal/tui/screens/backup_menu_test.go` | Created | Temp-dir create/list/delete/empty-golden acceptance coverage. |
| `internal/tui/screens/backup_menu.go` | Created | Injected managed-path backup store and Bubble Tea menu. |
| `internal/tui/screens/testdata/backup_menu/empty.golden` | Created | Deterministic empty-state fixture. |
| `internal/tui/{app,dependencies}.go` | Modified | Injected backup-factory route only. |
| `internal/tui/app_test.go` | Modified | Root route regression coverage. |

### Current Boundary

- T-081 through T-094: completed and independently verified.
- T-095/T-096: implemented with GREEN evidence; intentionally unchecked pending independent verification.
- Rollback boundary: remove the backup menu, its test/fixture, and the injected backup-factory route; no Phase 4, 6, or 7 behavior changed.

## Judgment Day Fix Round 1 — T-095/T-096

- Scope: only the confirmed CRITICAL flattened-destination defect in `FileBackupStore.Create`; task markers remain unchecked. No commit, push, or PR was performed.
- RED: `rtk go test ./internal/tui/screens -run '^TestBackupMenuCreatesAllPreResolvedPathsAndListsPriorSessions$' -count=1` failed to build because `ManagedPath` was undefined.
- GREEN: the same focused command passed after each source received an injected canonical relative destination rather than using `filepath.Base`.
- Regression: two temp-dir managed files named `settings.toml` under distinct parents retain `.config/mydots/settings.toml` and `.local/share/settings.toml` with distinct contents; no real HOME is resolved.
- Existing behavior retained: the same focused suite proves prior-session listing; `TestBackupMenuConfirmedDeletionRemovesOnlySelectedTimestamp` proves confirmation and selected-only deletion.
- Race: `rtk go test -race ./internal/tui/... -count=1` — 94 passed in 3 packages.
- Full: `rtk go test ./... -count=1` — 748 passed in 16 packages.
- Quality: `gofmt` on the changed Go files and `git diff --check` passed.

## T-095/T-096 Record Finalization

- Independent verification verdict: **PASS WITH WARNINGS**; `safe_to_mark_complete: true`.
- Task markers: T-095/T-096 are checked in `docs/tasks.md` and this OpenSpec tracker after the verified result and scoped Judgment Day Round 2 approval.
- Final full suite: `go test -json ./... -count=1` — **748 passed, 0 failed, 20 existing skips, 14 package passes**.
- Canonical CRITICAL destination-collision finding is fixed and verified: each injected `ManagedPath.RelativeDestination` is preserved beneath the timestamp root, so equal-basename sources retain distinct canonical paths and contents.
- Focused behavior: 4 passed; empty-state golden passed twice with a stable SHA-256; relevant race: 94 passed in 3 packages; build, vet, scoped formatting, and diff hygiene passed.
- Non-blocking warning: `internal/tui/screens/backup_menu.go` statement coverage is 75.2%, below the informational 80% threshold; all required T-095/T-096 scenarios passed.
- No code, test, commit, push, or PR operation was performed during record finalization.

## Pre-commit R3-001/R3-002 Remediation — T-095/T-096

- Scope: only the verified backup containment and atomic rollback defects. Existing task markers were not changed.
- Safety net: `rtk go test ./internal/tui/screens -count=1` — 72 passed before adding the regressions.
- RED: `rtk go test ./internal/tui/screens -run 'TestFileBackupStore(RejectsUnsafeSessionAndDestinationPaths|RollsBackPartialSessionWhenCopyFails)' -count=1` — 0 passed, 7 failed. It accepted `.`/`..`, absolute and traversal destinations, created a list-visible session, and retained a session after a deterministic second-copy failure.
- GREEN: the focused regression command — 7 passed after validating every timestamp/destination before writes, exclusively creating the timestamp directory, and removing it when any copy fails.
- Triangulation: `.` and `..` session timestamps; absolute, direct-parent, and nested-parent destinations; plus a valid first copy followed by a deterministic missing second source. Every rejected/failing case proves `List()` is empty and no session root remains.
- REFACTOR: extracted `backupDestination` for one containment rule shared by all managed paths; no unrelated behavior changed.
- Focused: `rtk go test ./internal/tui/screens -count=1` — 72 passed.
- Race: `rtk go test -race ./internal/tui/screens -count=1` — 72 passed.
- Full: `rtk go test ./... -count=1` — 755 passed in 16 packages.
- Quality: `gofmt` on the changed backup files and `git diff --check` passed.
- Host I/O: none; all paths and sources are under `t.TempDir()`.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-095/T-096 | `internal/tui/screens/backup_menu_test.go` | Temp-dir filesystem unit | ✅ 72 screen tests | ✅ 7 deterministic unsafe-path/partial-session failures | ✅ 7 focused regressions passed | ✅ 5 unsafe inputs plus late-copy rollback | ✅ Shared containment helper; no unrelated changes |

### Workload Boundary

- Mode: `size:exception` (previous maintainer approval retained).
- Current work unit: T-095/T-096 pre-commit R3 remediation.
- Rollback boundary: revert the containment helper and the two regression tests; no Phase 4, 6, or 7 files changed.

## Incident Recovery Fix Round 1 — Post-185139d

- Scope: confirmed surviving backup create/delete durability and SCR-06 confirmation behavior only. OpenSpec tracking is retained; task markers were not changed.
- RED: `rtk go test ./internal/tui/screens -run '^(TestBackupMenuCreateRequiresConfirmationAndShowsSuccess|TestFileBackupStoreListIgnoresStagingAndIncompleteSessions|TestFileBackupStoreDeleteQuarantinesBeforeRemoval|TestFileBackupStoreDeleteRenameFailureLeavesOriginalIntact)$' -count=1 -v` failed at build time because the injected filesystem seam did not exist.
- GREEN: the same focused regression command passed (5 tests) after staging completed copies under `.<timestamp>.staging`, publishing with `Rename`, requiring `.complete` for list visibility, and quarantining deletion through `.<timestamp>.deleting` before removal.
- Triangulation: a late missing-source copy removes its staging directory; staging and incomplete timestamp directories do not list; removal failure leaves only a hidden quarantine and reports the error; rename failure leaves the original complete session intact; `c` now requires `Create backup now? (y/n)` and renders `Backup created: <timestamp>` after completion.
- Focused: `rtk go test ./internal/tui/screens -count=1` — 76 passed.
- Race: `rtk go test -race ./internal/tui/screens -count=1` — 76 passed.
- Full: `rtk go test ./... -count=1` — 759 passed in 16 packages.
- Quality: `go vet ./...`, `go build ./...`, scoped `gofmt -d`, and `git diff --check` passed.
- Preserved: containment validation, canonical same-basename destinations, selected-only deletion, empty golden, injected root route, and `t.TempDir()`-only test I/O.
- Git: no staging, commit, push, reset, checkout, PR, or history rewrite was performed.

## Final Verify #418 — Technical PR-Readiness Remediation

- Scope: only the technical blockers from final verify report #418. The two final discipline checkboxes remain intentionally unchecked; no git, issue, PR, branch, or host-I/O operation was performed.
- Module-name behavior: canonical `installer.ProgressEvent` now carries `Module.Name()` with the typed `ModuleID`. `InstallScreen` renders that name while retaining the ID as its row key and the original typed status.
- Approved integration seam: proposal, delta spec, and design now explicitly approve the minimal `internal/installer/executor.go` cancellation-before-next-module guard and canonical `installer.Run` delegation. This is Phase 5 integration, not broad Phase 4 scope.
- Lint: smallest behavior-preserving fixes add one package comment per affected package, check the input-close defer, remove dead test assignments and an unused validation wrapper. `golangci-lint run` is clean.
- Coverage: the backup coverage warning remains informational and was not expanded.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| Verify #418 module-name rendering | `internal/tui/screens/install_adapter_test.go` | Production-path adapter + direct screen unit | ✅ `go test ./internal/tui/... ./internal/installer/... -count=1` — 627 passed in 7 packages before changes. | ✅ `TestModuleSessionRunnerRendersModuleNameWhileRetainingModuleID` failed to build because `ProgressEvent.ModuleName` was undefined. | ✅ Focused regression passed after canonical executor progress metadata and name-aware row rendering. | ✅ Asserts `Module.Name()` text renders, typed `ModuleID` remains unchanged, and the ID is not rendered as the human name. | ✅ Kept ID-keyed rows and only added an empty-name compatibility fallback for direct legacy events. |
| Verify #418 lint remediation | Existing affected tests | Refactor / static analysis | ✅ 151 tests passed in TUI screens and installer packages before lint cleanup. | ➖ No executable behavior change. | ✅ `golangci-lint run` — 0 issues. | ➖ Structural-only cleanup. | ✅ Minimal docs/error/dead-code cleanup; behavior suites remain green. |

### Verification Evidence

- Focused module-name and installer-semantics suite: **3 passed**.
- Relevant race: `go test -race ./internal/tui/... ./internal/installer/... -count=1` — **554 passed in 7 packages**.
- Tagged integration: `go test -tags=integration ./... -count=1` — **760 passed in 16 packages**.
- TUI/golden repeatability: two `go test ./internal/tui/... -count=1` runs — **106 passed** each; 20 golden fixtures remained stable.
- Full suite: `go test ./... -count=1` — **760 passed in 16 packages**.
- Build and vet: `go build ./...`; `go vet ./...` — passed.
- Module and hygiene: `go mod verify`; `go mod tidy -diff`; `gofmt -d` on changed Go files; `git diff --check` — passed.
- Lint: `golangci-lint run` — **0 issues**.

## Final Discipline Record Completion

- Authorization: final verification returned **PASS** with `safe_to_mark_final_discipline_complete: true`.
- Both Phase 5 final-discipline checkboxes in this OpenSpec tracker are now marked complete; no exact matching checkbox exists in `docs/tasks.md`, so that authoritative tracker was not altered.
- Final evidence retained: `go test -json ./... -count=1` — **760 passed, 0 failed, 20 skipped, 14 package passes**; relevant race **554 passed, 0 failed**; integration-tag suite **760 passed, 0 failed**; two TUI/golden runs **106 passed** each with **20 fixtures** and **0 hash drift**; `golangci-lint run` — **0 issues**.
- Discipline evidence: **16/16** implementation tasks complete; **8/8** RED/GREEN pair commits atomic; **13/13** conventional commit subjects; approved `size:exception` remains valid.
- No executable code or tests were edited, and no tests, git, issue, PR, push, or archive operation was run during this records-only finalization.
- External PR-governance blockers remain unchanged: no open `status:approved` issue and the `dev` branch requires an explicit naming-policy exception.

## Final Pre-PR Findings R4-001, R3-001, R3-002 — Local Fix Evidence

- Scope: only terminal-restore fail-closed recovery, sudo-cancellation routing, and result module-name preservation. Task checkboxes were not changed.
- RED: the focused regression initially failed because terminal recovery types/diagnostics and result-row `ModuleName` did not exist; cancellation rendered `ScreenResult`. The terminal-quit regression then failed with `screens.ChangeScreenMsg` before the explicit fail-closed exit branch, and the result-row propagation regression failed with an empty name.
- GREEN: `TerminalRestoreError` now reports one deterministic diagnostic to an injected writer (default `stderr`) and causes `tea.Quit`, never another interactive render. A `context.Canceled` elevation outcome routes directly to `ScreenMain`; genuine elevation errors remain `ScreenResult`. `InstallResultRow` carries the human name while preserving its typed `ModuleID`; failed results render the human name with an ID fallback for legacy rows.

### TDD Cycle Evidence

| Finding | Test File | RED | GREEN | REFACTOR |
|---|---|---|---|---|
| R4-001 | `internal/tui/screens/install_adapter_test.go`, `install_screen_test.go` | ✅ Missing recovery contract; then terminal failure emitted `ChangeScreenMsg` instead of `tea.Quit`. | ✅ Deterministic diagnostic and immediate non-interactive quit pass. | ➖ Narrow typed error and injected diagnostic writer. |
| R3-001 | `internal/tui/screens/install_screen_test.go`, `internal/tui/app_test.go` | ✅ Cancelled elevation routed to result. | ✅ Direct model and root-app routing reach Main Menu; genuine failure still reaches result. | ➖ Error classification only. |
| R3-002 | `internal/tui/screens/install_screen_test.go`, `result_screen_test.go` | ✅ Result row dropped the human name. | ✅ Typed ID and human name both survive; warning/critical goldens show names. | ➖ Legacy empty-name fallback retained. |

### Verification Evidence

- Focused regressions: `rtk go test ./internal/tui/screens ./internal/tui -run '^(TestSudoElevationAbortsWithPersistentDiagnosticWhenRestoreFails|TestInstallScreenCancelledElevationReturnsDirectlyToMainMenu|TestInstallScreenElevationFailureStillRoutesToResult|TestInstallScreenTerminalRestoreFailureQuitsWithoutRenderingAResult|TestInstallScreenCarriesHumanModuleNameIntoResultRows|TestResultScreenOutcomePrecedence|TestAppRoutesCancelledElevationDirectlyToMainMenu)$' -count=1 -v` — **11 passed**.
- Full and tagged integration: `rtk go test ./... -count=1`; `rtk go test -tags=integration ./... -count=1` — **766 passed in 16 packages** each.
- Goldens: relevant `Golden$` suite passed twice (**3 passed in 3 packages** each) after explicit `-update`; warning and critical-result fixtures changed only to human module names.
- Quality: build, vet, lint (0 issues), module verify/tidy diff, and scoped `gofmt -d` passed.
- Race: focused TUI race passed (95 tests). Full `go test -race ./... -count=1` remains blocked by the pre-existing Windows-host `internal/sudo` failures `TestStartKeepalive_OutputToDiscard` and `TestStartKeepalive_AlreadyCancelledContext`; no related code was changed.
- Git: no staging, commit, push, issue, PR, or branch operation was performed.
