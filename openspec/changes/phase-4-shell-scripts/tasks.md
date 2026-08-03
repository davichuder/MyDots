# Tasks: Phase 4 Shell Scripts

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 450–700 (under 800) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | T-073–075 → T-076 → T-077–079 → T-080/final verification |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | T-073–075 download wrappers | PR 1 | Independent, tests/docs/checkboxes together. |
| 2 | T-076 Docker hardening | PR 2 | Separate elevated-risk deliverable; depends on harness from PR 1. |
| 3 | T-077–079 verification/docs | PR 3 | Preserve valid scripts; Ghostty contract correction. |
| 4 | T-080 CI and phase verification | PR 4 | Gate after scripts; final evidence is separate. |

## Phase 1: Audit-First Wrapper Units

- [x] 1.1 **T-073** Audit `assets/scripts/homebrew-install.sh`; create RED evidence only for an unmet contract, minimally GREEN it, deterministically verify cleanup/failure/retry in `internal/installer/runner/runner_test.go`, ShellCheck, then update only T-073 in `docs/tasks.md` in one atomic Conventional Commit with WHY.
- [x] 1.2 **T-074** Audit `assets/scripts/omz-install.sh` for script-owned `RUNZSH=no CHSH=no`; RED only for a gap, minimally GREEN, verify no prompts/failure/cleanup with fakes, ShellCheck, then update only T-074 with tests and docs in its atomic WHY commit.
- [ ] 1.3 **T-075** Audit `assets/scripts/sdkman-install.sh`; RED only for a gap, minimally GREEN, verify POSIX execution, dependencies, cleanup, and propagated failure without network/root, ShellCheck, then update only T-075 in its atomic WHY commit.

## Phase 2: Docker Safety Unit

- [ ] 2.1 **T-076** Add deterministic RED cases, then minimally harden `assets/scripts/docker-linux.sh`: prerequisites, fresh temp cleanup, Docker skip/retry, `id -nG` failure stop, exact `docker` token probe, and propagated `usermod`; focused fake-tool tests, ShellCheck, then only T-076 and its behavior/docs in one atomic WHY commit.

## Phase 3: Preserve and Correct Units

- [ ] 3.1 **T-077** Audit `assets/scripts/font-linux.sh`; preserve it if its font validation, `unzip -o`, cache, cleanup, and failure contracts pass; otherwise RED/GREEN minimally. Run isolated focused tests and ShellCheck, then only T-077 in its atomic WHY commit.
- [ ] 3.2 **T-078** Verify-only audit `assets/scripts/ghostty-linux.sh`; preserve its temporary community `ghostty-ubuntu` flow, no `curl | bash`, cleanup, and failure propagation. Correct stale Ghostty contract in `docs/specs.md` and T-078 wording in `docs/tasks.md`; test, ShellCheck, atomic WHY commit.
- [ ] 3.3 **T-079** Audit `assets/scripts/caveman-install.sh` and `--only openclaw`; RED only for a gap, minimally GREEN, verify dependencies/cleanup/failure with fakes, ShellCheck, then only T-079 in its atomic WHY commit.

## Phase 4: CI and Whole-Phase Verification

- [ ] 4.1 **T-080** RED-audit the absent gate, create `.github/workflows/shellcheck.yml` with relevant `push`/`pull_request` paths, pinned SHA-verified ShellCheck v0.10.0 provisioning, and exact fail-closed command; validate workflow contract, ShellCheck, then only T-080 in its atomic WHY commit.
- [ ] 4.2 After all focused units, separately verify every T-073–T-080 scenario deterministically, rerun `shellcheck --shell=sh assets/scripts/*.sh`, and confirm each checkbox has task-specific evidence; do not mark incomplete tasks or open a PR.
