# Design: Phase 4 Shell Scripts

## Technical Approach

Treat T-073–T-080 as audit/completion work: establish deterministic evidence first, preserve Phase 3 behavior, and correct only proven gaps. Go owns normal idempotence; wrappers remain safe directly. Native Windows stays rejected by `internal/platform/detect.go`. Phase 3 and PR state remain untouched.

## Architecture Decisions

| Option | Tradeoff | Decision and rationale |
|---|---|---|
| Replace upstream installers | More control; large maintenance/security surface | Reject; audit documented wrappers. |
| Test live installers | Realistic but mutates hosts | Reject; use `/bin/sh` with fakes. |
| Put every repeat check in shell | Self-contained; duplicates module state logic | Go `IsInstalled` owns normal skip; shell owns safe direct retry and partial-state protection. |
| Rely on runner ShellCheck | Tooling may be absent | Add pinned, fail-closed CI because no workflow exists. |

## Data Flow

    executor IsInstalled/platform gate
                │ install needed
                v
    runner.Script ──/bin/sh──> wrapper ──> temporary upstream payload
                                      │
                          fake PATH in tests / real tools in production

## Audit and Work Classification

| Task | Likely class | Required audit contract |
|---|---|---|
| T-073 Homebrew | Focused hardening candidate | Validate `curl`, `mktemp`, `/bin/bash`; propagate failures; clean temp; prevent false success. Module `brew` probe owns skip; upstream owns direct repeat. |
| T-074 OMZ | Focused hardening candidate | Validate `curl`, `mktemp`, `sh`; invoke with script-owned `RUNZSH=no CHSH=no`; preserve runner env as defense; clean and prevent false success. Module directory probe owns skip. |
| T-075 SDKMAN | Focused hardening candidate | Validate `curl`, `mktemp`, `bash`; verify failures and cleanup. Module init-file probe owns skip; upstream owns direct repeat. |
| T-076 Docker | **Confirmed focused hardening** | See Docker contract below. |
| T-077 Nerd Font | Verify-only unless evidence fails | Preserve safe `FONT_NAME`, temp archive, extraction/cache propagation, cleanup. Module configured `fc-list`/cask probe owns skip; `unzip -o` owns safe direct repeat. |
| T-078 Ghostty | **Documentation-contract correction + verify-only script** | Preserve community source, required-tool checks, temp cleanup, installer propagation, and no false success. Module binary probe owns skip. |
| T-079 Caveman | Focused hardening candidate | Validate `curl`, `mktemp`, `sh`; preserve `--only openclaw`; propagate failures, clean, prevent false success. Module binary probe owns skip. |
| T-080 | **Confirmed new CI implementation** | Pinned ShellCheck must provision and run fail-closed. |

Candidates change only after RED evidence.

## Interfaces / Contracts

### Deterministic shell harness

Extend `internal/installer/runner/runner_test.go` with one table-driven helper. Each case gets temporary `HOME`, `TMPDIR`, command log, and controlled `PATH`; fake `curl`, `bash`/`sh`, `sudo`, `id`, `getent`, `usermod`, `unzip`, and `fc-cache` return controlled results. Assert order, arguments, status, cleanup, and no false success. Never touch network, root, or host paths.

### Docker

Use `mktemp "${TMPDIR:-/tmp}/mydots-docker.XXXXXX"` with cleanup. Validate `USER`, `curl`, `mktemp`, `sudo`, `sh`, `id`, `getent`, and `usermod` before mutation. If `docker` exists, skip; otherwise download/run the temp script, propagating failure. `id -nG "$USER"` must succeed; an exact space-delimited `docker` token skips `usermod`. Probe failure exits before `usermod`; only confirmed non-membership runs and propagates `sudo usermod -aG docker "$USER"`. Retries use fresh files and never report false success.

### Ghostty and CI

Ghostty remains a temporary download of community `mkasberg/ghostty-ubuntu` executed by `bash`; never invent an official Ubuntu tarball or use `curl | bash`. Correct stale `docs/specs.md`, `docs/tasks.md`, and Ubuntu acceptance wording.

Create `.github/workflows/shellcheck.yml` on `ubuntu-24.04`; trigger on `push` and `pull_request` changes to `assets/scripts/**` or the workflow. Provision official ShellCheck `v0.10.0` Linux x86_64 and verify SHA-256 `6c881ab0698e4e6ea235245f22832860544f17ba386442fe7e9d629f8cbedf87`; provisioning failure fails the job.

```sh
shellcheck --shell=sh assets/scripts/*.sh
```

## File Changes

| Path | Action | Certainty |
|---|---|---|
| `internal/installer/runner/runner_test.go` | Modify harness/contracts | Confirmed |
| `assets/scripts/docker-linux.sh` | Harden lifecycle and retry | Confirmed |
| `assets/scripts/{homebrew-install,omz-install,sdkman-install,caveman-install}.sh` | Minimal remediation only after RED evidence | Candidate |
| `assets/scripts/{font-linux,ghostty-linux}.sh` | Preserve unless audit fails | Verify-only |
| `.github/workflows/shellcheck.yml` | Create pinned gate | Confirmed |
| `docs/specs.md`, `docs/tests.md` | Correct Ghostty/testing contracts | Confirmed |
| `docs/tasks.md` | Update one checkbox in its own work unit | Confirmed |

## Testing Strategy

| Layer | Approach |
|---|---|
| Contract | Table-driven `/bin/sh` success, missing dependency, action failure, cleanup, retry, and no-false-success cases. |
| Module/platform | Retain Go precheck, WSL2 systemd/Wayland gating, and Windows rejection tests. |
| Static | Run `shellcheck --shell=sh assets/scripts/*.sh` before every future commit and in CI. |
| Final phase | After all focused verification and checkbox updates, separately verify all T-073–T-080 requirements. |

## Migration / Rollout

No migration. One strict-TDD unit per task: RED audit, minimal GREEN remediation, focused verification, then only that task's `docs/tasks.md` update. Preserve T-077's earned check. Each atomic commit stays within 800 lines and uses a Conventional Commit subject, blank line, and WHY body without AI attribution. No Phase 3 reopening or PR actions.

## Open Questions

None.
