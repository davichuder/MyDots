## Exploration: phase-4-shell-scripts

### Current State

Phase 4 is an audit/completion change, not a blank implementation. The seven installer scripts exist under `assets/scripts/`, and focused deterministic POSIX execution tests cover the three download wrappers, Caveman, Ghostty, and Nerd Font behavior. Phase 3 explicitly verified the Ghostty Ubuntu wrapper, Caveman download-failure propagation, Nerd Font failure paths, and the supported runtime boundary of macOS plus Ubuntu/WSL2; it did not claim ShellCheck success because ShellCheck was unavailable.

The scripts are invoked through the embedded filesystem and `runner.Script`, which executes them with `/bin/sh`. The current scripts use `set -eu`, temporary files with cleanup traps, and guarded downloads for most network wrappers. The main exception is `docker-linux.sh`, which is only `set -e`, downloads mutable code to `/tmp/get-docker.sh`, always runs `sudo usermod -aG docker "$USER"`, does not check Docker/group state, and does not clean up. `omz-install.sh` does not itself set the required `RUNZSH=no CHSH=no` variables, although its module supplies them through the runner environment. `sdkman-install.sh` and the Homebrew/OMZ/Caveman wrappers execute downloaded installers and need their unattended, idempotence, and failure contracts audited rather than assumed.

`docs/tasks.md` still marks T-073–T-076, T-078, T-079, and T-080 incomplete, despite implementations and focused tests already existing. T-078 is stale: `docs/specs.md` still describes a GitHub release tarball, while the current script downloads and executes the documented community `ghostty-ubuntu` installer from Ghostty's binary-install guidance. No `openspec/config.yaml` exists, so no repository-local OpenSpec rules were available to refine this exploration. No files exist under `.github/workflows/`; therefore T-080 has missing implementation, not an existing CI step.

### Current-State Matrix

| Task | Current implementation | Existing evidence | Phase 4 gap / status |
|---|---|---|---|
| T-073 Homebrew | `assets/scripts/homebrew-install.sh` downloads the official installer to a temp file, runs it with `NONINTERACTIVE=1`, and cleans up. | `internal/installer/runner/runner_test.go` covers success, download failure, and no false completion. | Audit idempotence/preflight and shell contract; likely documentation/task status update. Existing behavior is substantially implemented. |
| T-074 Oh My Zsh | `assets/scripts/omz-install.sh` downloads and runs the installer; module passes `RUNZSH=no CHSH=no`. | Same runner tests cover success and download failure; module tests cover environment handoff. | Make the script contract self-contained or explicitly document runner-provided variables; add deterministic checks for installer failure and no prompts. |
| T-075 SDKMAN | `assets/scripts/sdkman-install.sh` downloads and runs the official installer with `bash`. | Runner tests cover success and download failure; runtime tests cover script dispatch. | Audit unattended behavior, failure propagation, idempotence, and dependency/error messaging. |
| T-076 Docker Linux | `assets/scripts/docker-linux.sh` downloads and runs `get.docker.com`, then appends the user to the Docker group. | Module tests cover dispatch and script exit propagation; no direct shipped-script contract test was found. | Significant gap: no idempotence/group check, fixed temp path, no cleanup, no required-tool validation, mutable elevated remote execution, and weak diagnostics. |
| T-077 Nerd Font | `assets/scripts/font-linux.sh` validates `FONT_NAME`, checks `unzip`, downloads to temp storage, extracts into the user font directory, refreshes cache, and cleans up. | Extensive deterministic tests cover success, download/extraction/cache failures, missing unzip, and unsafe names. | Audit idempotence and archive/content edge cases; implementation is already substantially complete and must not be recreated. |
| T-078 Ghostty Linux | `assets/scripts/ghostty-linux.sh` downloads the `ghostty-ubuntu` community installer to a temp file, runs it with `bash`, propagates failures, and cleans up. | Deterministic tests cover URL/temporary-file reuse, download failure, missing tools, installer failure, and cleanup; Phase 3 verified the wrapper. | Update stale task/spec contract from nonexistent official tarball to documented community installer; retain the no-`curl | bash` design. |
| T-079 Caveman | `assets/scripts/caveman-install.sh` downloads to a temp file, then runs it with `--only openclaw`, with cleanup and guarded execution. | Deterministic tests cover download failure, installer invocation/argument, success, and no false completion. | Audit idempotence and installer failure diagnostics; implementation is substantially complete. |
| T-080 ShellCheck CI | No workflow exists under `.github/workflows/`; the repository has no provisioned CI ShellCheck gate. | `docs/tasks.md` and `docs/tests.md` describe intended CI commands, but they are documentation only; Phase 3 records ShellCheck unavailable locally. | Later apply work must create a new workflow at a proposed path (for example, `.github/workflows/ci.yml`), provision ShellCheck explicitly, and make warnings or missing tooling fail the job. |

### Affected Areas

- `assets/scripts/homebrew-install.sh` — Homebrew download/execute/idempotence contract.
- `assets/scripts/omz-install.sh` — unattended environment and downloaded-installer failure behavior.
- `assets/scripts/sdkman-install.sh` — official installer, Bash dependency, unattended/idempotent behavior.
- `assets/scripts/docker-linux.sh` — Linux installation, privilege boundary, cleanup, group idempotence, and false-success risks.
- `assets/scripts/font-linux.sh` — already hardened Nerd Font implementation to audit without regression.
- `assets/scripts/ghostty-linux.sh` — documented Ubuntu community installer path and cleanup/failure behavior.
- `assets/scripts/caveman-install.sh` — downloaded wrapper execution and `--only openclaw` contract.
- `internal/installer/runner/runner_test.go` — existing deterministic shipped-script tests and the natural location for additional shell contract coverage.
- `internal/installer/modules/{homebrew,oh_my_zsh,docker,ghostty,nerd_font,caveman}*.go` — environment, platform gating, dispatch, and error wrapping that scripts depend on.
- `.github/workflows/ci.yml` — proposed/new workflow path for T-080; no workflow currently exists under `.github/workflows/`.
- `docs/tasks.md` — stale Phase 4 completion markers and the obsolete T-078 wording.
- `docs/specs.md` — M-48 Ubuntu strategy contradicts the documented/current Ghostty implementation.
- `docs/tests.md` — claims ShellCheck in CI but does not define deterministic shell-script test boundaries.
- `openspec/changes/Phase-3/{verify-report.md,review-ledger.md}` — existing historical evidence only; Phase 3 must not be reopened or reverified.

### Approaches

1. **Audit and minimally complete existing scripts** — preserve Phase 3 implementations, harden only proven gaps (especially Docker and script-owned unattended contracts), add deterministic no-network/no-root/no-mutation tests, make ShellCheck provisioning explicit, and correct Phase 4 documentation.
   - Pros: smallest review surface, respects Phase 3 closure, preserves verified Ghostty/Font/Caveman behavior, directly addresses false-success and idempotence risks.
   - Cons: some installer behavior remains dependent on upstream scripts; requires careful seams for shell tests and CI tool availability.
   - Effort: Medium

2. **Replace all wrappers with locally managed installers** — avoid downloaded upstream scripts by reproducing package/repository setup and install logic in this repository.
   - Pros: more locally controlled behavior and easier static inspection.
   - Cons: high maintenance/security burden, likely diverges from official installers, unnecessarily reopens hardened Phase 3 work, and is incompatible with the corrected Ghostty community-installer contract.
   - Effort: High

### Proposed Audit and Completion Approach

Use strict TDD in atomic work units, without reopening Phase 3:

1. Establish deterministic shell harness coverage using temporary homes, fake `curl`/`bash`/`sudo`/`apt`/`usermod`/`getent` commands, command logs, and injected PATHs. Never use network, root, or real-system mutation.
2. Audit each wrapper for POSIX `/bin/sh` syntax, `shellcheck --shell=sh`, explicit dependency failures, cleanup on every exit path, installer failure propagation, and absence of success output after failure.
3. Harden Docker specifically: safe temporary download and cleanup, explicit prerequisites, controlled apt installer failure propagation, and a group-membership check before `usermod`; preserve platform gating in the Go module.
4. Make unattended variables part of the OMZ script contract (or test/document the runner contract explicitly), and verify SDKMAN/Homebrew/Caveman idempotence boundaries without executing their real installers.
5. Verify the existing Nerd Font and Ghostty contracts with current tests; do not alter them merely to satisfy stale task prose.
6. Create/provision a real mandatory CI ShellCheck gate in a later apply phase. The workflow path is proposed, not existing; install or otherwise guarantee ShellCheck availability, retain the exact `shellcheck --shell=sh` invocation, and ensure any warning/non-zero exit fails the job.
7. Update Phase 4 task/spec/test documentation, correcting Ghostty to the documented `ghostty-ubuntu` temporary-download installer and recording Windows host as unsupported. Mark each `docs/tasks.md` item incrementally only after that task's audit/remediation and focused verification succeed. Use atomic Conventional Commit work units with a blank line and mandatory WHY body; no attribution.

### Deterministic Test Strategy

- Run scripts under `/bin/sh` with temporary `HOME`, `TMPDIR`, and PATH-only fake tools.
- Fake network downloads by making `curl` write controlled installer/archive content or fail; assert installers are not invoked after download failure.
- Fake privileged/package commands and record arguments; never invoke `sudo`, `apt`, `usermod`, `fc-cache`, or system directories for real.
- Cover success, dependency absence, upstream installer failure, cleanup, idempotent rerun, and no false-success output for every script; retain focused existing coverage for Ghostty and Nerd Font.
- Validate Docker group membership using deterministic fake `id`/`getent` behavior and assert `usermod` is skipped when already present.
- Run `shellcheck --shell=sh` against every `assets/scripts/*.sh` in CI and treat missing ShellCheck or any warning as failure; local inability to run it is evidence to solve, not a pass.

### Recommendation

Proceed with Approach 1. The repository already contains meaningful Phase 3 hardening and deterministic tests; the Phase 4 change should close the remaining Docker, unattended-contract, CI provisioning, idempotence, and documentation gaps while preserving verified behavior. The Ghostty contract must be corrected to the documented community `ghostty-ubuntu` installer, not reverted to a nonexistent official Ubuntu tarball URL.

### Risks

- Docker's mutable remote installer executes with elevation and requires especially strict failure/cleanup tests and clear user diagnostics.
- Upstream Homebrew, OMZ, SDKMAN, and Caveman installers can change independently; tests must validate wrapper behavior without pinning or executing network content.
- No `.github/workflows/` files exist, so T-080 requires new CI implementation; the proposed workflow path and provisioning method must be reviewed in proposal/design.
- Idempotence is partly owned by Go module `IsInstalled` checks and partly by scripts; the proposal must define that boundary to avoid duplicate installs or misleading success.
- Existing `docs/specs.md` and `docs/tasks.md` contain stale contracts/status; updating them must not modify Phase 3 artifacts.

### Ready for Proposal

Yes. The proposal should scope Phase 4 as an audit/minimal-completion change: Docker hardening, deterministic contract tests for all scripts, ShellCheck as a provisioned mandatory gate, and documentation/status correction. It should explicitly exclude Phase 3 re-verification, production implementation outside the agreed shell-script scope, native Windows-host support, and any real network/root/system mutation in tests.

**Status**: success
**Executive Summary**: Audited T-073 through T-080 against the current scripts, Go dispatch/tests, documentation, and Phase 3 evidence. Most scripts are already implemented and hardened; Docker, self-contained unattended contracts, a missing CI ShellCheck gate, and stale documentation remain the principal Phase 4 gaps.
**Artifacts**: `openspec/changes/phase-4-shell-scripts/exploration.md` | Engram `sdd/phase-4-shell-scripts/explore`
**Next Recommended**: `sdd-propose`
**Risks**: Docker's elevated mutable remote installer; upstream script drift; new CI workflow/provisioning required for T-080; split idempotence ownership; stale Ghostty contract.
**Skill Resolution**: `paths-injected` — loaded `sdd-explore`, `sdd-phase-common`, `openspec-convention`, and `engram-convention` from the exact requested paths.
