# Phase 3 Apply Progress

**Mode**: Strict TDD  
**Artifact store**: hybrid  
**Delivery**: four reviewable work-unit slices  
**Verified remediation commit**: `3a56717` (`test(installer): isolate config backup seams`)
**Status**: **VERIFIED BY FINAL SCOPED JUDGES — R3-001-CI remediation is locally and WSL-validated; GitHub PR #5 CI rerun is pending after push**

## Completed Work

- [x] Work unit 1 — Shell and runner boundaries: `R1-001`, `R1-002`, `R4-003`, applicable `R3-004` coverage.
- [x] Work unit 2 — Runtime and catalogue boundaries: `R2-001`, `R3-001`.
- [x] Work unit 3 — Config safety and idempotence: `R3-002`, `R1-003`, `R3-003`.
- [x] Work unit 4 — Linux platform correctness and font audit: `R4-001`, `R4-002`, final `R3-004` coverage.
- [x] Final verification closure: `V3-001`, `V3-002`, `R3-007`.
- [x] C-001 — Ghostty Ubuntu wrapper verified.
- [x] C-002 — Ghostty Darwin cask test verified.
- [x] C-003 — Docker Darwin cask test verified.
- [x] C-004 — Fresh Homebrew executable handoff verified across changed brew-dependent paths.
- [x] R3-001-CI — Isolated MCP, Neovim framework, and Theme backup test seams from Linux `/tmp` homes; Theme now asserts backup-before-write ordering.

## R3-001-CI Remediation Verification

- Root cause: tests redirected module homes to `t.TempDir()` while their real backup dependency still resolved `/home/runner`; backup containment correctly rejected `/tmp` paths in Linux CI.
- Fix: the existing injectable backup seams use scoped no-op fakes where backup storage is not under test. The Theme ordering test uses a recording fake and proves `backup` precedes `write`.
- Production behavior is unchanged: `mcpBackupFile`, `zshrcBackupFile`, and `themeBackupFile` still default to `backup.BackupFile`; backup containment remains intact.
- Final scoped judges: independent Judge A and Judge B found no open R3-001-CI findings.
- Local and WSL evidence: the exact MCP reconciliation test, focused MCP/Neovim/Theme suites, modules package, focused race/shuffle checks, `rtk go test ./...`, `go vet ./...`, and `git diff --check` pass. The WSL run used a Go 1.26.3 linux/amd64 binary.
- Full-suite evidence: `rtk go test ./...` reports 559 passed events, 0 failures, 20 intentional Windows-host skips, and 12 packages.
- Remote follow-up: GitHub PR #5 CI rerun is pending after the remediation commit is pushed; this is remote confirmation, not an open local or scoped-review finding.

## TDD Cycle Evidence

| Work unit | RED | GREEN | REFACTOR | Final evidence |
|---|---|---|---|---|
| Shell and runner boundaries | Failure and command-boundary tests written first | Focused suite passed | Shared deterministic test helpers retained | Verified |
| Runtime and catalogue boundaries | Independent runtime and SDKMAN tests written first | Contract-focused suites passed | Obsolete combined tests removed | Verified |
| Config safety and idempotence | Stale-state and backup tests written first | Focused configuration suite passed | Shared atomic writer retained | Verified |
| Linux platform and font audit | Failure-path and parser tests written first | Focused script/module suites passed | Test-only PATH helper extracted | Verified |
| Golden and EOL closure | Golden assertions and fixtures added first | Focused Theme/MCP suite passed | `gofmt`; no production refactor needed | Verified |
| C-001 through C-004 remediation | Exact cask, wrapper, and handoff tests written first | Full suite passed | Session-scoped brew path avoids global `PATH` mutation | Verified |
| R3-001-CI remediation | GitHub Actions failure identified the real backup invocation against a `/tmp` test home | Focused MCP/Neovim/Theme suites, module package, focused race suite, WSL Linux binary, and full suite pass | Test-only fakes restore globals with `t.Cleanup`; Theme records backup-before-write | Verified by final scoped judges; PR #5 CI rerun pending after push |

## Final Verification

- `rtk go test ./...`: exit 0; 559 passed events, 0 failures, 20 intentional Windows-host skips, 12 packages.
- `go vet ./...`: exit 0.
- `git diff --check main...dev` and `origin/main...dev`: exit 0.
- Full 4R/refutation and scoped Judgment Day are complete; C-001 through C-004 are verified.
- R3-001-CI final scoped Judge A and Judge B reviews: no open findings; production backup containment remains unchanged.
- Supported scope: macOS and Ubuntu/WSL2; native Windows host is unsupported.

## Informational Warnings

- ShellCheck is unavailable in this environment.
- `rtk` reports no installed pre-commit hook.
- `docs/tasks.md` retains stale golden-fixture metadata; `.gitattributes` is authoritative.
- Repository-wide lint retains 11 unrelated pre-existing `errcheck` findings; changed-lines lint reports none.

**Next recommended phase**: push `3a56717`, observe the GitHub PR #5 CI rerun, then archive/merge if it passes.
