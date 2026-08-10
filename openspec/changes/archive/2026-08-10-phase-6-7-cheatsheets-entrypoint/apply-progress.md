# Apply Progress: Phase 6–7 Cheatsheets and Entrypoint

## Status

Unit 1 / T-097 is complete, independently verified, Judgment Day Round 2 approved, and committed as `ea642fb`. Unit 2 / T-098–T-100 is independently verified PASS and Judgment Day Round 2 approved; task markers are complete. The final corrective evidence proves joined failure shutdown, exact `--default` save arguments before pipeline execution, and WSL2 propagation through both entry routes. Unit 3 is resolved: shared T-098 evidence fully covers T-100 ordering, failure suppression, and joined lifecycle behavior, so no redundant work unit is required.

## Strict TDD Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | Triangulate | Refactor |
|---|---|---|---|---|---|---|---|
| 1.1–1.2 | `assets_test.go` | Unit | N/A (new) | Final corrected bytes with cheatsheets withheld: `rtk go test . -run 'TestEmbeddedCheatsheet(InventoryIsExactAndReadable|sHaveUsefulSCR07Content)$|TestCorrectedCheatsheetContentMatchesInstalledTools' -count=1` → 51 failures, including missing `assets/cheatsheets`. | Same command restored → 51 pass. | Exact sorted inventory, 48 section subtests, and eight corrected-reference regressions. | Content contracts reject prior incorrect helpers, repositories, targets, and commands. |
| 1.3 | `internal/tui/screens/reference_slug_test.go` | Unit | `rtk go test ./internal/tui/screens` baseline: 85 pass. | Final test with `reference_slug.go` withheld → build failure: `undefined: referenceSlug`. | Restored unchanged test → 1 pass. | 48 mappings, unknown-ID rejection, and install-screen lookup. | One readable slug per module. |
| T-098 shared RED | `entrypoint_test.go` | Unit | `rtk go test . -count=1`: 51 pass; `rtk go test ./internal/sudo -count=1`: 8 pass. | Before production code: `rtk go test . -run 'TestRun(Routes|WaitsForKeepaliveShutdown)$' -count=1` → package build failure with `undefined: entrypointDependencies`, `buildVersion`, and `run`. | Same focused command after GREEN → 9 pass. | Version, Windows, unsupported OS, exact FR-22/no-side-effects, unattended/default/no-TUI, no-flag TUI, UTC timestamp, and a gate proving `run` cannot return before shutdown joins. | Removed unsafe test parallelism; all test factory swaps are mutex-serialized and cleanup-restored. |
| T-099 GREEN | `entrypoint_test.go`, `internal/sudo/sudo_test.go` | Unit | Same baselines. | Shared T-098 RED above. | Focused main+sudo: 68 pass; changed-package race: 68 pass. | Valid unattended and default use different config paths; missing config and unsupported platform prove omissions; blocked-stop test proves joined cleanup. | `StartKeepalive` now returns an idempotent synchronous handle; old timer test synchronizes on the validation call and stops its handle. |
| T-098–T-100 corrective RED | `entrypoint_test.go` | Unit | Before new tests: focused entrypoint/TUI/sudo: 96 pass. | `rtk go test . -run 'TestProduction(TUICompositionRoutesInstallAndBackupWithInjectedPaths|PipelinePropagatesPlatformAndCriticalFailure|DefaultsAndAssetsAreInjected)$|TestRunFailurePathsPreventLaterCallsAndJoinKeepalive' -count=1` → build failure: undefined `productionTUIApp` and `runProductionPipelineWith`. | Same focused command after GREEN: 24 pass. | Darwin + Linux Native composition/platform cases; install + backup model routes; critical pipeline error; validation/elevation/start-error/nil-handle/save/pipeline/TUI failures. | Added only composition and progress-draining seams; retained Phase-5 adapters and top-level `run` factory seam. |
| T-100 failure-join corrective RED | `entrypoint_test.go` | Unit | Existing success-only shutdown gate: `TestRunWaitsForKeepaliveShutdown`. | Final `TestRunPipelineFailureWaitsForKeepaliveShutdown` bytes unchanged; temporarily withheld only `defer keepalive.Stop()` in `runPrivileged`, then `rtk go test . -run '^TestRunPipelineFailureWaitsForKeepaliveShutdown$' -count=1` → 0 pass, 1 fail: `run returned 1 before failure shutdown started`. | Restored the one-line defer, then `rtk go test . -run '^(TestRunPipelineFailureWaitsForKeepaliveShutdown|TestRunDefaultSavesExactInjectedPathAndConfig)$' -count=1` → 2 pass. | Blocking fake sends `stopStarted`, holds `stopGate`, and proves `run` remains blocked until the gate releases on the pipeline-error path. | No production behavior changed beyond restoring the existing defer; the RED is assertion failure, not compilation or timeout. |
| T-098 default boundary corrective RED | `entrypoint_test.go` | Unit | Existing default-route test asserted only the `save` call. | Final `TestRunDefaultSavesExactInjectedPathAndConfig` bytes unchanged; temporarily withheld only the injected config path (`saveConfig("", configuration)`), then `rtk go test . -run '^TestRunDefaultSavesExactInjectedPathAndConfig$' -count=1` → 0 pass, 1 fail: wrong saved path and wrong path observed by pipeline. | Restored `saveConfig(dependencies.configPath(), configuration)`; same two-test GREEN command → 2 pass. | Captures the exact injected path and `config.DefaultConfig()` value at Save and again at the pipeline boundary. | The counterfactual changed only the route argument; production default behavior was restored unchanged. |
| JD WSL2 propagation | `entrypoint_test.go` | Direct route / production composition | Injected detector returns Linux + WSL2 + `amd64`; TUI and privileged pipeline both assert the complete value. | With the detector seam present but unused by `run`: `rtk go test . -run '^TestRunPropagatesDetectedWSL2PlatformToTUIAndPipeline$' -count=1` → 0 pass, 3 fail; TUI and pipeline received Linux/Native with empty Arch. | `run` invokes the injected detector and production wires `platform.Detect`; same command → 1 pass. | The test runs TUI and `--unattended` routes with no host reads and checks OS, variant, and architecture. | Windows and unsupported routes remain explicitly guarded before detection. |

## Unit 1 Implemented Files

- `assets_test.go`, `internal/tui/screens/reference_slug.go`, `internal/tui/screens/reference_slug_test.go`, `internal/tui/screens/install_screen.go`, and `internal/tui/screens/install_screen_test.go`.
- The exact 48 `assets/cheatsheets/*.md` files required by the delta spec.

## Unit 2 Implemented Files

- `main.go` — thin `os.Exit(run(os.Args, runtime.GOOS))` wrapper and ldflags-addressable `buildVersion`.
- `entrypoint.go` — flag routing, serialized dependency factory, production TUI install/backup composition, platform propagation, UTC timestamp, critical-progress propagation, and privileged lifecycle ownership.
- `entrypoint_test.go` — deterministic direct top-level route, production-composition, failure-path, critical-progress, and joined-shutdown tests using temp paths and injected dependencies only.
- `internal/sudo/sudo.go` and `internal/sudo/sudo_test.go` — joinable, idempotent `Keepalive.Stop` and deterministic lifecycle cleanup.

## Verification

| Command | Result |
|---|---|
| Unit 1 focused / full / changed race | PASS: 137 / 821 / 137 tests at its verification point |
| Final narrow RED, withheld shutdown: `rtk go test . -run '^TestRunPipelineFailureWaitsForKeepaliveShutdown$' -count=1` | Intended FAIL — 0 pass, 1 fail; `run returned 1 before failure shutdown started` |
| Final narrow RED, withheld path: `rtk go test . -run '^TestRunDefaultSavesExactInjectedPathAndConfig$' -count=1` | Intended FAIL — 0 pass, 1 fail; wrong saved path and pipeline observation |
| Final narrow GREEN: `rtk go test . -run '^(TestRunPipelineFailureWaitsForKeepaliveShutdown|TestRunDefaultSavesExactInjectedPathAndConfig)$' -count=1` | PASS — 2 tests / 1 package |
| Final focused: `rtk go test . ./internal/tui ./internal/sudo -count=1` | PASS — 98 tests / 3 packages |
| Final changed-package race: `rtk go test -race . ./internal/sudo -count=1` | PASS — 85 tests / 2 changed packages |
| `rtk go test ./...` | PASS — 847 tests / 16 packages |
| `rtk go vet ./...` | PASS |
| `rtk golangci-lint run` | PASS — no issues |
| `rtk go build ./...` | PASS |
| `rtk git diff --check` plus untracked `entrypoint.go` / `entrypoint_test.go` checks | PASS |
| JD WSL2 final narrow / focused / race / full / vet / lint / build | PASS: 1 / 78 / 86 / 848 tests; vet, `golangci-lint`, and build clean |

## T-100 Coverage Determination

No redundant standalone T-100 work unit was created. The shared corrective route suite proves validation failure, elevation failure, keepalive-start error and nil handle all prevent later calls; default save failure; pipeline failure invokes shutdown and cannot return until the blocking Stop/join gate is released; and TUI failure returns 1. The behavior-level RED for withheld shutdown proves the failure-path assertion genuinely detects lifecycle ownership. Independent verification PASS and Judgment Day Round 2 approval authorize T-100 as complete.

## Scope and Review Budget

- Delivery: approved Unit-2-only `size:exception`.
- Unit 2 corrective worktree: **781 additions, 10 deletions, 791 changed lines** across five implementation/test files plus the committed `docs/tasks.md` completion markers. This remains within the authorized Unit-2-only `size:exception` and below the 800-line stop limit.
- Task markers are complete. Authorized Unit 2 work-unit commit: `36fde190a1757c1fd291f5584e65695bba8d56a1` (`feat(cli): add entrypoint execution modes`). This local SDD directory remains untracked and unstaged.
- Phase 3 archive changes, Phase 5 archive changes, and `openspec/specs/interactive-tui/` were not modified or staged.

## Release Gate Reconciliation

- **3.1 complete:** Final verification reports T-097 unit `ea642fb` at **687 changed lines** with **51 new files**, within its approved 800-line new-file-predominant `size:exception`; T-098–T-100 unit `36fde19` is **797 changed lines** with **2 new files**, within its approved 800-line `size:exception`. The whole range includes Phase 3 and Phase 5 archive commits, which are distinct reviewed archive units and are not implementation-unit budget evidence.
- **3.2 complete:** No PR existed before full Phase 6–7 verification. The user explicitly authorizes **final verify → archive → PR**. This local artifact reconciliation performs no staging, commit, push, merge, or PR creation.
