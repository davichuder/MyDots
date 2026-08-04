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

## Judgment Day — T-074 Apply Round 1

**State:** FIX APPLIED — pending independent re-judgment; one real warning converged across both blind judges; no CRITICAL findings.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T074-001 | judgment-day | `internal/installer/runner/runner_test.go:533-580` | WARNING | info | The shared shipped-script success and download-failure tests inherit the developer's real `HOME`. With the new Oh My Zsh skip path, a machine that already has `~/.oh-my-zsh` bypasses the stubs and makes the suite host-state dependent. | Both judges independently confirmed the deterministic-isolation regression. |
| JD-T074-002 | judgment-day | `assets/scripts/omz-install.sh:40` | WARNING | info | Judge A found that `RUNZSH=no CHSH=no` may still permit an upstream `.zshrc` overwrite prompt in a direct TTY invocation unless the official unattended option or equivalent confirmation setting is supplied. | Suspect; Judge B found no unattended-contract defect. |
| JD-T074-003 | judgment-day | `assets/scripts/omz-install.sh:11-14` | WARNING | info | Judge A treated any existing `.oh-my-zsh` directory as a concrete partial-install false-success risk; Judge B classified the same edge as theoretical and consistent with the approved module-owned skip boundary. | Suspect; assessment did not converge. |

Only the deterministic `HOME` isolation warning converged. The user authorized its scoped fix: both shared shipped-script tests now set `HOME` to `t.TempDir()`. The focused shared tests (8 passed), full Go suite (12 packages passed), formatting/diff checks, and exact WSL ShellCheck gate passed; independent re-judgment remains judge-owned. Suspect findings JD-T074-002 and JD-T074-003 remain unaddressed.

## Judgment Day — T-074 Apply Round 2

**State:** APPROVED — both blind judges verified the authorized deterministic-isolation fix.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T074-001 | judgment-day | `internal/installer/runner/runner_test.go:533-580` | WARNING | verified | Both shared shipped-script tests now set `HOME` to independent `t.TempDir()` values before invoking scripts, preventing a real `~/.oh-my-zsh` from bypassing stubs. | Both judges reproduced focused and full-suite success and approved the resolution. |

JD-T074-002 and JD-T074-003 remain non-convergent informational first-pass signals outside the authorized fix scope. T-074 reached terminal `JUDGMENT: APPROVED` with zero confirmed CRITICAL or real WARNING findings remaining.

## Judgment Day — T-075 Apply Round 1

**State:** ESCALATED — one judge reported a supported-macOS Bash-version defect; the second judge reported no real warnings.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T075-001 | judgment-day | `assets/scripts/sdkman-install.sh:26-29,40` | WARNING | info | The official SDKMAN installer requires Bash 4+, while stock macOS 15 provides Bash 3.2. The wrapper checks only that `bash` exists, so a normal supported macOS install can pass preflight and fail upstream. | Suspect; Judge A reproduced the platform mismatch, while Judge B reported no real warning and did not evaluate Bash major version. |

The finding is concrete but non-convergent. Per Judgment Day rules it requires explicit triage before any fix; no automatic remediation was applied.

## Judgment Day — T-075 Apply Round 2

**State:** APPROVED — both blind judges verified the Bash 4+ compatibility gate.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T075-001 | judgment-day | `assets/scripts/sdkman-install.sh` | WARNING | verified | The wrapper resolves `bash --version` before `mktemp` or `curl`, accepts parsed major version 4 or newer, and fails Bash 3.x or invalid output with an actionable Bash 4+ error. Deterministic tests fake Bash 3, invalid, and Bash 4 outputs without reading the host shell version. | Both judges reproduced early rejection, Bash 4+ acceptance, deterministic tests, and full regression success. |

Only JD-T075-001 was addressed. No T-076+, Phase 3, or unrelated findings were changed. T-075 reached terminal `JUDGMENT: APPROVED` with zero confirmed CRITICAL or real WARNING findings remaining.

## Judgment Day — T-076 Apply Round 1

**State:** ESCALATED — both blind judges confirmed a retry false-success defect.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T076-001 | judgment-day | `assets/scripts/docker-linux.sh:13-16` | CRITICAL | open | `command -v docker` skips the entire script, including group reconciliation. If installation succeeds but `id` or `usermod` fails, the retry exits successfully because Docker now exists while the user remains outside the `docker` group. It also cannot reconcile a pre-existing Docker installation. | Both judges independently confirmed the normal-use retry/pre-existing-install defect; severity differed between CRITICAL and real WARNING. |
| JD-T076-002 | judgment-day | `assets/scripts/docker-linux.sh:18` | WARNING | info | `getent` is validated but unused. Supported Debian/Ubuntu targets normally provide it, so impact is limited to stripped environments. | Judge B only; theoretical INFO. |
| JD-T076-003 | judgment-day | `assets/scripts/docker-linux.sh:27` | WARNING | info | Cleanup uses `rm` without explicit validation. Supported targets provide it as an essential utility. | Judge B only; theoretical INFO. |

Only JD-T076-001 enters the fix loop. Explicit user authorization is required before scoped remediation and blind re-judgment.

## Judgment Day — T-076 Apply Round 2

**State:** APPROVED — both blind judges verified the retry/group-reconciliation fix.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T076-001 | judgment-day | `assets/scripts/docker-linux.sh:13-46` | CRITICAL | verified | The Docker-present branch skips only APT/package work, then proceeds to the fail-closed `id -nG "$USER"` exact-token probe and conditional `sudo usermod -aG docker "$USER"`. Deterministic fake-only tests cover retry/pre-existing non-membership, existing membership, probe failure, and usermod failure without network, root, or host mutation. | Both judges reproduced focused/full verification and confirmed no false-success path remains. |

JD-T076-002 and JD-T076-003 remain informational and were not changed. T-076 reached terminal `JUDGMENT: APPROVED` with zero confirmed CRITICAL or real WARNING findings remaining.

## Judgment Day — T-077 Apply Round 1

**State:** APPROVED — both blind judges returned empty findings ledgers after Judge B's model was corrected.

| id | lens | location | severity | status | evidence | convergence |
|---|---|---|---|---|---|---|
| JD-T077-001 | judgment-day | T-077 work unit | WARNING | info | Judge A reproduced 11 focused font tests, 600 full-suite tests, clean formatting/ShellCheck, current release assets, and concluded the pre-existing checkbox is now earned. | Historical infrastructure note: the original Judge B model returned malformed empty output; after the user changed the model, replacement Judge B independently approved with zero findings. |

Both judges verified the dependency preflight, safe name handling, official asset URL, cleanup, repeat behavior, failure propagation, deterministic tests, and tracker consistency. T-077 reached terminal `JUDGMENT: APPROVED` with zero confirmed findings.
