# Review Ledger: Phase 5 TUI

## Judgment Day — Design Round 1

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-001 | judgment-day | `design.md:29,50-58,69` | BLOCKER | verified | Both blind judges verified that the Phase-5-owned `SessionRunner` checks the sole session context before every module and prevents all later starts/writes after cancellation. |
| JD-002 | judgment-day | `design.md:29,51-61,69` | CRITICAL | verified | Both blind judges verified the typed elevation/handoff/clock seams, `sudo -v`, raw-terminal restoration, joined 45-second keepalive, cancellation cleanup, and deterministic fakes. |
| JD-A-002 | judgment-day | `design.md:15-16,72` | CRITICAL | info | Suspect: Judge A found the current teatest package incompatible with Bubble Tea v2; Judge B only confirmed the missing `testing.TB` argument as a warning. |
| JD-B-001 | judgment-day | `design.md:15,29` | CRITICAL | info | Suspect: Judge B found Huh v2 requires a local Bubble Tea v2 adapter; Judge A recorded the same concern only as a warning. |
| JD-B-003 | judgment-day | `design.md:60` | CRITICAL | info | Suspect: Judge B found buffered runner output prevents live logs; Judge A classified the same evidence as a warning. |
| JD-B-004 | judgment-day | `design.md:13-15,60` | CRITICAL | info | Suspect: only Judge B found WSL2 skip/failure status semantics inconsistent with the proposed mapping. |
| JD-B-005 | judgment-day | `design.md:27-29,53-54,62` | CRITICAL | info | Suspect: only Judge B found no backup-before-config-save ordering despite FR-08. |
| JD-B-006 | judgment-day | `design.md:45-54,62` | CRITICAL | info | Suspect: Judge B found managed paths lack file/directory and missing-path semantics; Judge A classified the missing-path portion as a warning. |
| JD-A-004 | judgment-day | `design.md:60` | WARNING | info | Real: existing command output is buffered until command completion, limiting perceived live progress. |
| JD-A-005 | judgment-day | `design.md:14,55-60` | WARNING | info | Real: disabled rows have no complete module metadata source in the design. |
| JD-A-006 | judgment-day | `design.md:61-62` | WARNING | info | Real: missing managed targets and partial backup failures need explicit semantics. |
| JD-A-007 | judgment-day | `design.md:15` | WARNING | info | Real: Huh v2 requires a wrapper converting its string view into a Bubble Tea v2 view. |
| JD-B-008 | judgment-day | `design.md:72` | WARNING | info | Real: the documented teatest call omits the required `testing.TB` argument. |
| JD-B-009 | judgment-day | `design.md:45-58` | WARNING | info | Real: several local contracts remain underspecified, including module criticality needed by result classification. |

**Confirmed:** 2 · **Suspect:** 6 · **Contradictions:** 0 · **Info:** 6

**State:** `JUDGMENT: APPROVED` — both confirmed findings were fixed and independently verified in scoped blind re-judgment.

## Judgment Day — Apply T-081/T-082 Round 1

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-A-001 | reliability | `internal/tui/components/log_pane_test.go:26-35` | WARNING | info | Current behavior is correct, but test input and expectation both derive from `LogPaneMaxLines`, so the test does not independently lock the exact value 200. |
| JD-A-002 | resilience | `internal/tui/components/log_pane.go:6-8` | WARNING | info | Exported mutable `Lines` could let future consumers bypass `Append`; no current consumer or user-impacting defect exists. |

**Confirmed:** 0 · **Suspect:** 0 · **Contradictions:** 0 · **Info:** 2

**State:** `JUDGMENT: APPROVED` — no BLOCKER/CRITICAL finding; both focused and full suites passed.

## Judgment Day — Apply T-085/T-086 Round 1

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-003 | reliability | `internal/tui/screens/main_menu.go:1` | WARNING | info | Both judges confirmed focused/full/race tests pass but focused lint reports ST1000 because the new package has no package comment. |
| JD-004 | judgment-day | `openspec/changes/phase-5-tui/apply-progress.md:56-57` | WARNING | info | Both judges found cumulative file-change evidence still says task markers were updated only through T-084 although T-085/T-086 are checked. |

**Confirmed:** 0 · **Suspect:** 0 · **Contradictions:** 0 · **Info:** 2

**State:** `JUDGMENT: APPROVED` — no BLOCKER/CRITICAL finding; focused, full, race, build, and vet checks passed.

## Pre-commit Reliability — T-085/T-086

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| R3-001 | reliability | `internal/tui/screens/main_menu_test.go:54-72` | BLOCKER | verified | After two bounded fix rounds, Config, Backup, and Reference are exercised through direct Bubble Tea v2 `Update()` key messages and assert their exact resulting screens. |

**State:** verified — focused route test and full suite passed after the final fix.

## Judgment Day — Apply T-083/T-084 Round 1

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-A-001 | judgment-day | `internal/tui/components/progress_row_test.go:38-64`; `design.md:70`; `tasks.md:33,59` | WARNING | info | Real: the golden harness does not define the documented `-update` workflow, and its 80x24 constants appear only in the failure message rather than constraining rendering. ProgressRow behavior and deterministic fixtures otherwise pass. |

**Confirmed:** 0 · **Suspect:** 0 · **Contradictions:** 0 · **Info:** 1

**State:** `JUDGMENT: APPROVED` — no BLOCKER/CRITICAL finding; focused tests passed repeatedly and the full suite passed.

## Judgment Day — Apply T-087/T-088 Round 2

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-A/B-001 | persistence | `internal/tui/screens/config_menu.go` | CRITICAL | verified | **Copied-config persistence**: Huh field bindings now target stable pointer-owned form state, so edits survive value-returning Bubble Tea model updates. The real lifecycle regression reloads and verifies non-default persisted values. |
| JD-A/B-002 | persistence | `internal/tui/screens/config_menu.go` | CRITICAL | verified | **Unbound-language persistence**: selected language IDs bind to a slice and are converted to `config.LanguageOptions` before validation and save. Real lifecycle regressions verify Java-only and Java+PHP persistence. |
| V-001 | module hygiene | `go.mod`, `go.sum` | INFO | info | `go mod tidy -diff` proposes a project-wide manifest rewrite. This is non-blocking: `go mod verify`, dependency resolution, focused/race checks, and the 697-test full suite passed. Keep cleanup separate from this task-marker update. |

**Confirmed:** 2 · **Verified:** 2 · **Info:** 1

**State:** `JUDGMENT: APPROVED` — Round 2 verified both canonical CRITICAL persistence findings as fixed. Independent verification returned `safe_to_mark_complete: true`; T-087/T-088 may be marked complete. No commit was performed.

## Judgment Day — Apply T-089/T-090 Finalization

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-T089-001 | lifecycle | `internal/tui/screens/install_screen.go` | CRITICAL | verified | Canonical Judgment Day finding verified: one install session starts from the initial window-size path; cancellation reaches in-flight elevation and prevents later work. |
| JD-T089-002 | parity | `internal/tui/screens/install_adapter.go`, `internal/installer/executor.go` | CRITICAL | verified | Canonical Judgment Day finding verified: `ModuleSessionRunner` delegates to canonical installer behavior, including cancellation-before-next-module protection. |
| V-T089-001 | result contract | `internal/tui/screens/install_screen.go` | CRITICAL | remediated | Verifier critical remediated and verified: completion and closed-channel paths emit the `ScreenResult` payload contract. |
| V-T089-002 | reference rendering | `internal/tui/screens/install_screen.go` | CRITICAL | remediated | Verifier critical remediated and verified: the active module's injected reference renders, and unavailable reference states are explicit. |
| V-T089-003 | status rendering | `internal/tui/screens/install_screen.go` | CRITICAL | remediated | Verifier critical remediated and verified: direct installed and failed events render matching indicators. |
| V-T090-001 | sudo streams | `internal/tui/screens/install_adapter.go` | CRITICAL | remediated | Verifier critical remediated and verified: initial `sudo -v` uses interactive streams; 45-second keepalive is quiet. |
| V-T090-002 | deterministic clock | `internal/tui/screens/install_adapter_test.go` | CRITICAL | remediated | Verifier critical remediated and verified: fake-clock synchronization is channel-only and race-safe. |
| W-T089-001 | design wording | `design.md`, `apply-progress.md` | WARNING | info | Stale design wording describes a local module loop/no Phase-4 edits, while the accepted parity remediation delegates to canonical `installer.Run`; reconciliation remains intentionally deferred. |

**Verified:** 7 · **Remediated criticals:** 5 · **Info:** 1

**Verification:** focused 15/15, focused race 15/15, installer race 448/448, and `go test -json ./... -count=1` **712 passed, 0 failed, 20 existing skips; 13 package passes, 0 package failures**.

**State:** `JUDGMENT: APPROVED` — canonical findings are verified, all five verifier CRITICAL findings are remediated and verified, and independent verification passed with `safe_to_mark_complete: true` and 712 tests. No commit was performed.

## Pre-commit Reliability — T-089/T-090

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| R3-001 | reliability | `internal/tui/screens/install_screen.go:69-77` | BLOCKER | remediated | After cancellation, a queued non-terminal event previously returned no replacement wait command, so the following `InstallDoneEvent` was never consumed and no result transition occurred. The regression proves progress is drained and terminal completion transitions to `ScreenResult`; existing runner coverage proves cancellation still prevents later module starts/writes. |

**State:** remediated — strict RED-to-GREEN regression, focused 42-test screen suite, and 713-test full suite passed. No task marker, commit, push, or PR changed.

## Judgment Day Round 2 — T-091/T-092 Result Screen

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-T091-001 | critical result propagation | `internal/tui/screens/install_adapter.go`, `internal/tui/screens/install_screen.go`, `internal/tui/screens/result_screen.go` | CRITICAL | verified | Production critical `ProgressEvent.Err` reaches `InstallResult` and renders the critical result outcome. Regression `TestModuleSessionRunnerCriticalFailureRendersCriticalResult` passed. |
| JD-T092-001 | result navigation | `internal/tui/screens/result_screen.go` | CRITICAL | verified | Every `tea.KeyMsg` emits `ChangeScreenMsg{Screen: ScreenMain}`. Regression `TestResultScreenAnyKeyReturnsToMainMenu` passed. |

**Verified:** 2 canonical CRITICAL findings.

**Verification:** focused 14/14, focused race 14/14, relevant race 504/504, and `go test ./... -count=1` **727 passed, 0 failed, 15 package passes**; five result goldens were stable across a repeated run.

**State:** `JUDGMENT: APPROVED` — both result-screen CRITICAL findings are verified. Independent verification returned `safe_to_mark_complete: true`; T-091/T-092 may be marked complete. No code, test, commit, push, or PR operation was performed during record finalization.

## T-093/T-094 Finalization

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-T093-001 | config-issue routing | `internal/tui/app.go`, `internal/tui/screens/config_menu.go` | CRITICAL | verified | Canonical finding fixed and verified: `App` preserves `screens.ConfigIssue`, routes it to a fresh assistant, renders the original error, and performs no load/save/retry loop. |
| V-T093-001 | typed startup | `internal/tui/dependencies.go`, `internal/tui/app.go` | CRITICAL | remediated | `Dependencies.Platform` is the sole startup selector: Darwin/Linux open Main Menu, Windows opens Preflight, unsupported values render a typed explicit error. |
| V-T093-002 | routed sizing | `internal/tui/app.go` | CRITICAL | remediated | Root retains every `tea.WindowSizeMsg` and replays it after install routing so the deferred screen starts correctly. |
| V-T093-003 | golden determinism | `internal/tui/testdata/app/main_menu_program.golden` | CRITICAL | remediated | Fixed 80x24/no-color native Program golden was updated explicitly and passed repeated normal reruns with a stable hash. |
| V-T093-004 | harness isolation | `internal/tui/app_test.go` | CRITICAL | remediated | Fixed environment, neutral `TEA_TRACE`, `WithoutSignals`, and `WithoutSignalHandler` prevent host trace I/O and signal-handler leakage. |
| V-T093-005 | Windows preflight | `internal/tui/app.go`, `internal/tui/screens/preflight.go` | CRITICAL | remediated | Injected embedded WSL2 guide renders and Windows invokes the install factory zero times. |
| V-T094-001 | reference selection | `internal/tui/screens/reference_menu.go` | CRITICAL | remediated | All embedded Markdown references are listed deterministically and the selected tool renders; unavailable remains explicit. |
| V-T094-002 | ConfigIssue propagation | `internal/tui/app.go` | CRITICAL | remediated | Direct assistant issue messages retain their payload and render without configuration lifecycle side effects. |
| W-T093-001 | coverage | `internal/tui/app.go` | WARNING | info | 79.5% statement coverage is narrowly below the informational 80% threshold; all required behavior is covered. |
| W-T093-002 | invalid install dependency | `internal/tui/app.go` | WARNING | info | Invalid install payload or nil `InstallFactory` retains/re-initializes the active model rather than rendering a dependency error; production wiring supplies the factory. |
| D-T093-001 | native-v2 harness | `apply-progress.md`, `verify-report.md` | INFO | approved | Bubble Tea v1-only `x/exp/teatest` is incompatible with Bubble Tea v2 model contracts; the genuine feasibility RED is retained and the native Bubble Tea v2 `Program` harness is approved. |

**Verified:** 1 canonical CRITICAL · **Remediated criticals:** 7 · **Info:** 3

**Verification:** focused 17/17, focused race 17/17, relevant race 90/90, and `go test -json ./... -count=1` **744 passed, 0 failed, 20 existing skips; 14 package passes**.

**State:** `JUDGMENT: APPROVED` — independent verification returned **PASS WITH WARNINGS** and `safe_to_mark_complete: true`; T-093/T-094 are marked complete. No code, test, commit, push, or PR operation was performed during record finalization.

## T-095/T-096 Finalization

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-T095-001 | backup destination integrity | `internal/tui/screens/backup_menu.go`, `backup_menu_test.go` | CRITICAL | verified | Canonical path-collision finding is fixed: caller-injected `ManagedPath.RelativeDestination` is joined below the timestamp root, and same-basename sources preserve distinct canonical destinations and contents. |
| W-T095-001 | coverage | `internal/tui/screens/backup_menu.go` | WARNING | info | Statement coverage is 75.2%, below the informational 80% threshold; all required file/directory creation, list, confirmation, selected-only deletion, empty state, and root-routing scenarios passed. |

**Verified:** 1 canonical CRITICAL · **Info:** 1

**Verification:** focused 4/4; empty-state golden passed twice with a stable hash; relevant race 94/94; `go test -json ./... -count=1` **748 passed, 0 failed, 20 existing skips; 14 package passes**; build, vet, scoped formatting, and diff hygiene passed.

**State:** `JUDGMENT: APPROVED` — **PASS WITH WARNINGS** and `safe_to_mark_complete: true`; T-095/T-096 are marked complete. No code, test, commit, push, or PR was performed during finalization.

## Pre-commit Reliability — T-095/T-096 R3

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| R3-001 | reliability | `internal/tui/screens/backup_menu.go`, `backup_menu_test.go` | CRITICAL | remediated | `Create` rejects `.`/`..` timestamps and absolute or any `.`/`..`-component destination before filesystem writes. The regression proves every rejected input leaves `List()` empty and no write escapes its intended session root. |
| R3-002 | reliability | `internal/tui/screens/backup_menu.go`, `backup_menu_test.go` | CRITICAL | remediated | `Create` exclusively creates its session then removes it on any managed-path copy failure. A deterministic late missing-source regression proves the first copy leaves neither a session directory nor a list-visible partial backup. |

**Verification:** focused 72/72, focused race 72/72, and `go test ./... -count=1` **755 passed in 16 packages**.

**State:** remediated — strict regression-first evidence recorded; no task markers changed.
