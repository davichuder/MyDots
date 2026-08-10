# Tasks: Phase 5 Interactive TUI

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 1,200–1,700 additions + deletions |
| 800-line budget risk | High |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Eight T-xxx-pair work units |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

Single PR exceeds the 800-line budget; obtain maintainer `size:exception` approval before apply.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| T-081/082 | Components | single PR | Tests and behavior together; conventional pair commit with rationale. |
| T-083/084 → T-095/096 | Remaining pair units | single PR | Continue only after `size:exception`; each pair stays independently reviewable. |

## Phase 1: Components

- [x] **T-081 [RED]** Add `internal/tui/components/log_pane_test.go`: direct tests for empty logs and exact 200-line eviction.
- [x] **T-082 [GREEN]** Create `log_pane.go` with `LogPaneMaxLines=200`; pass T-081 without host I/O.
- [x] **T-083 [RED]** Add `progress_row_test.go` direct status mapping tests and fixed 80x24/no-color row goldens.
- [x] **T-084 [GREEN]** Create `progress_row.go`; render names and pending/running/installed/failed/skipped/disabled indicators.

## Phase 2: Menu and Configuration

- [x] **T-085 [RED]** Add `screens/main_menu_test.go`: five choices, quit, missing-config inline error, unloadable-config assistant route, and golden.
- [x] **T-086 [GREEN]** Create `main_menu.go` using injected `ConfigStore`; route valid Install only.
- [x] **T-087 [RED]** Add `config_menu_test.go`: prove `charm.land/huh/v2` Bubble Tea v2 adapter lifecycle, six defaults, five fonts, HTTPS rejection, temp-dir persistence, goldens.
- [x] **T-088 [GREEN]** Modify `go.mod`/`go.sum` for Huh v2; create `config_menu.go` that delegates the adapter and saves valid config before menu return.

## Phase 3: Installation and Results

- [x] **T-089 [RED]** Add `install_screen_test.go` and `install_adapter_test.go`: direct events, `?`, cancellation, one-event requeue, SessionRunner no-later-start/write, elevation release/restore, `sudo -v`, joined 45-second fake-clock keepalive.
- [x] **T-090 [GREEN]** Create `install_screen.go` and `install_adapter.go`; own one cancellable session, typed seams, raw-terminal handoff, deterministic runner/elevation/clock fakes.
- [x] **T-091 [RED]** Add `result_screen_test.go` fixed goldens for success, warnings, critical failure, cancellation progress, and WSL2 note.
- [x] **T-092 [GREEN]** Create `result_screen.go`; enforce cancellation > critical > warnings > success precedence.

## Phase 4: Root and Backups (Phase-5 TUI only)

- [x] **T-093 [RED]** Add `app_test.go`/`preflight_test.go`: first prove teatest+BTea v2 feasibility, then teatest only Darwin menu and Windows guide/key exit; direct reference `not available` tests.
- [x] **T-094 [GREEN]** Create `app.go`, `preflight.go`, `reference_menu.go`, and dependencies; use injected platform/assets and never start Windows installation.
- [x] **T-095 [RED]** Add `backup_menu_test.go`: temp-dir managed paths, prior-session list, confirmed selected-only deletion, and empty-state golden.
- [x] **T-096 [GREEN]** Create `backup_menu.go`; create all pre-resolved paths and confirm deletion by timestamp.

## Phase 5: Verification and Commit Discipline

- [x] Run each RED test demonstrably failing, then narrow and full `go test ./...`; goldens update only via `-update` then rerun normally.
- [x] Commit each pair with tests and behavior in one conventional T-xxx-pair commit whose body states the decision rationale; no Phase 4/6/7 edits, network, root, real HOME, or host state.
