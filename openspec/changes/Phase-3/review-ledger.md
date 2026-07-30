# Phase 3 Final Review Ledger

**Verified remediation commit**: `3a56717` (`test(installer): isolate config backup seams`)
**Review tier**: Full 4R/refutation plus scoped Judgment Day
**Supported runtime**: macOS and Ubuntu/WSL2  
**Status**: **VERIFIED BY FINAL SCOPED JUDGES — R3-001-CI is locally and WSL-validated; GitHub PR #5 CI rerun is pending after push**

## Closed Findings

| ID | Final status | Final evidence |
|---|---|---|
| `R1-001` | verified | POSIX download wrappers stop before execution on failure. |
| `R1-002` | verified | `BrewCask` uses exact `brew install --cask <name>` arguments. |
| `R2-001` | verified | Runtime contracts, versions, sequencing, and dependencies are independently tested. |
| `R3-001` | verified | SDKMAN initialization and Java row matching are coherent. |
| `R3-001-CI` | verified | Final scoped Judge A and Judge B found no open findings. Linux CI test homes are outside `/home/runner`, while real backup containment correctly rejects them. The exact MCP reconciliation test and other affected tests use scoped backup fakes restored with `t.Cleanup`; Theme asserts backup-before-write; the obsolete `themeMkdirAll` seam is removed. Production containment is unchanged. Local and WSL evidence passes; GitHub PR #5 CI rerun remains pending after push. |
| `R3-002` | verified | Configured-state reconciliation and managed-entry preservation are covered. |
| `R3-003` | verified | Nerd Font audit parsing rejects incomplete and malformed managed blocks. |
| `R3-004` | verified | Missing-unzip, extraction, and font-cache failure paths pass deterministically. |
| `R3-007` | verified | Golden fixtures use LF normalization and pass clean-clone validation. |
| `R4-001` | verified | Linux Nerd Font validation, extraction, and failure propagation are safe. |
| `R4-002` | verified | Linux Zsh registration and failure short-circuiting pass. |
| `R4-003` | verified | Caveman download failure cannot execute the installer. |
| `V3-001` | verified | `git diff --check main...dev` and `origin/main...dev` exit 0. |
| `V3-002` | verified | Theme/MCP golden assertions execute and pass. |
| `C-001` | verified | Ghostty Ubuntu wrapper is verified. |
| `C-002` | verified | Ghostty cask test is verified. |
| `C-003` | verified | Docker cask test is verified. |
| `C-004` | verified | Fresh Homebrew handoff is verified across changed brew-dependent paths. |

## Informational Items

| ID | Status | Evidence |
|---|---|---|
| `R3-005` | info | WSL2 status vocabulary differences remain non-blocking. |
| `R1-004` | warning | Docker installer executes mutable remote code with elevation. |
| `R3-006` | info | Some audit strings differ from documented metadata. |
| `R3-008` | info | Coverage remains below the informational guideline for host-skipped install bodies. |
| `R4-004` | warning | ShellCheck is unavailable in this environment. |
| `R1-003-Windows` | out-of-scope | Native Windows-host replacement semantics are unsupported. |
| `V3-I01` | info | Root build entry point is deferred to Phase 7. |
| `V3-I02` | info | Full race instability is confined to out-of-phase `internal/sudo` tests. |
| `V3-I03` | info | `rtk` reports no installed pre-commit hook. |
| `V3-I04` | info | `docs/tasks.md` has stale golden-fixture metadata; `.gitattributes` is authoritative. |
| `V3-I05` | info | Repository-wide lint has 11 unrelated pre-existing `errcheck` findings; changed-lines lint reports none. |

## Final Command Evidence

- `rtk go test ./...`: exit 0; 559 passed events, 0 failures, 20 intentional Windows-host skips, 12 packages.
- `go vet ./...`: exit 0.
- `git diff --check main...dev`: exit 0.
- `git diff --check origin/main...dev`: exit 0.
- Final scoped Judge A and Judge B reviews: no open R3-001-CI findings; local and WSL Linux evidence pass.
- `R3-001-CI` remediation commit: `3a56717`; GitHub PR #5 CI rerun is pending after push.

**Current review result**: **VERIFIED BY FINAL SCOPED JUDGES — R3-001-CI has no open local or scoped-review finding. Push `3a56717` and await the GitHub PR #5 CI rerun before merge.**
