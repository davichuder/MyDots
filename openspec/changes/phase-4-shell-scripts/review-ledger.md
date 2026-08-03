# Review Ledger: Phase 4 Shell Scripts

## Judgment Day — Design Round 1

**State:** APPROVED — both judges verified all Round 1 fixes.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-A-001 | judgment-day | `design.md:54-58` | CRITICAL | verified | Adds `push` and `pull_request` path triggers for shipped scripts and the workflow itself, making ShellCheck an automatic relevant-change gate. | Both judges verified the correction. |
| JD-A-002 | judgment-day | `design.md:48` | CRITICAL | verified | Requires a successful `id -nG "$USER"` probe and exact-token membership before skipping mutation; probe failure stops before `usermod`. | Both judges verified the correction. |
| JD-A-003 | judgment-day | `design.md:72-83` | CRITICAL | verified | Adds a separate final all-T-073–T-080 verification after focused verification and earned checkbox updates. | Both judges verified the correction. |

Judge B returned an empty ledger initially. The user authorized all three corrections, and both judges verified every fix in the scoped re-judgment.

## Judgment Day — T-073 Apply Round 1

**State:** APPROVED — no converged BLOCKER or CRITICAL findings.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T073-001 | judgment-day | `assets/scripts/homebrew-install.sh:11-29` | WARNING | info | Both judges noted that `/bin/bash` is not preflighted explicitly. Supported macOS and Ubuntu/WSL2 runtimes provide it; invocation failure remains fail-closed with cleanup and no false success. | Real/theoretical assessment differed; non-blocking. |
| JD-T073-002 | judgment-day | `internal/installer/runner/runner_test.go:560-689` | WARNING | info | Both judges noted that download-failure cleanup is mechanically covered by the EXIT trap but not asserted through an isolated temporary-path check. | Converged coverage warning; non-blocking. |
| JD-T073-003 | judgment-day | `internal/installer/runner/runner_test.go:582-689` | WARNING | info | Judge B requested direct repeated-invocation evidence; Judge A did not. The approved contract assigns normal skip to the Go probe and direct repeat behavior to upstream Homebrew. | Suspect; non-blocking. |
| JD-T073-004 | judgment-day | `internal/installer/runner/runner_test.go:994-1008` | SUGGESTION | info | Judge B observed helper duplication; behavior and isolation remain correct. | Suspect; non-blocking. |

Warnings and suggestions are informational under the severity floor and do not enter a fix/re-judge loop.
