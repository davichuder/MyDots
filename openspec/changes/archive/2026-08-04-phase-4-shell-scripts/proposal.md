# Proposal: Phase 4 Shell Scripts

## Intent

Complete T-073–T-080 through evidence-based audits, preserving Phase 3 work and correcting unsafe, stale, or unverified shell-script behavior for macOS and Ubuntu/WSL2 users.

## Scope

### In Scope
- Audit each installer for POSIX `/bin/sh`, dependencies, idempotence, cleanup, error propagation, and no false success; remediate only proven gaps.
- Harden Docker's temporary download, cleanup, prerequisite, and Docker-group behavior with deterministic strict-TDD coverage.
- Preserve audited Nerd Font and Ghostty behavior; correct Ghostty documentation to the documented community `ghostty-ubuntu` installer, never an official Ubuntu tarball.
- Make `shellcheck --shell=sh` a provisioned, mandatory CI gate; mark each task in `docs/tasks.md` only after its audit, remediation (if needed), and focused verification pass.

### Out of Scope
- Reopening or modifying Phase 3 artifacts; recreating scripts solely because they already exist.
- Native Windows-host support, real network/root/system mutation in tests, or replacing upstream installers with local package logic.
- PR creation, merge, or task-tracker edits during this proposal phase.

## Capabilities

### New Capabilities
- `shell-script-installation`: Audited, deterministic, POSIX installer-wrapper contracts for T-073–T-080, including a mandatory ShellCheck gate.

### Modified Capabilities
- None — `openspec/specs/` does not exist.

## Approach

Use atomic RED-GREEN work units. Test scripts under `/bin/sh` with temporary homes, paths, fake tools, and command logs—never live installers. Audit before changing; retain Phase 3's temporary-download, no-`curl | bash`, failure-propagating Ghostty strategy. Provision ShellCheck in CI and fail on its absence, warnings, or non-zero status. Keep `docs/tasks.md` completion incremental and evidence-backed.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `assets/scripts/` | Modified | Audit all seven wrappers; remediate proven gaps. |
| `internal/installer/runner/runner_test.go` | Modified | Deterministic shipped-script contracts. |
| `docs/specs.md` | Modified | Correct Ghostty Ubuntu contract. |
| `docs/tests.md` | Modified | Define deterministic shell-test and ShellCheck evidence. |
| `docs/tasks.md` | Modified later | Incremental evidence-backed T-073–T-080 tracking. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Elevated mutable Docker installer | High | Fake-tool tests; cleanup and explicit failure checks. |
| Upstream installer drift | Med | Assert wrapper contracts, not upstream execution. |
| Missing ShellCheck | Med | Explicit CI provisioning and failing gate. |

## Rollback Plan

Revert the atomic work unit that regresses a wrapper, tests, CI gate, or documentation; retain independently verified Phase 3 behavior and do not alter task status without passing evidence.

## Dependencies

- CI environment able to install and run ShellCheck.
- Ghostty binary-install guidance and community `ghostty-ubuntu` installer.

## Success Criteria

- [ ] Each T-073–T-080 audit has focused deterministic evidence; only proven gaps are changed.
- [ ] Tests cover failure propagation, idempotence boundary, cleanup, and no false success without network, root, or system mutation.
- [ ] CI runs `shellcheck --shell=sh` for every shell script and fails when unavailable or non-zero.
- [ ] Ghostty remains the temporary-download community-installer flow; task tracking is updated only after passing task-specific verification.
