# Apply Progress: Phase 4 Shell Scripts

## Delivery

- Mode: single PR with maintainer-approved `size:exception`.
- Work units completed: T-073, T-074, and T-075; this batch: T-075 only.
- Boundary: SDKMAN wrapper unattended, idempotence, dependency, cleanup, and failure-propagation contracts; no T-076 onward or Phase 3 artifacts.

## Task Status

- [x] 1.1 **T-073** — implementation, deterministic Go evidence, and mandatory ShellCheck pass.
- [x] 1.2 **T-074** — implementation, deterministic Go evidence, and mandatory ShellCheck pass.
- [x] 1.3 **T-075** — implementation, deterministic Go evidence, and mandatory ShellCheck pass.
- [ ] 2.1 **T-076**
- [ ] 3.1 **T-077**
- [ ] 3.2 **T-078**
- [ ] 3.3 **T-079**
- [ ] 4.1 **T-080**
- [ ] 4.2 Final whole-phase verification

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| T-073 | `internal/installer/runner/runner_test.go` | Deterministic `/bin/sh` contract | PASS — `TestShippedInstallerScriptsRunUnderPOSIXSh` and `TestShippedInstallerScriptsReturnDownloadFailure`: 8/8 cases | PASS — missing-`curl` test failed before production change because `mktemp` ran and no explicit dependency error; missing-`mktemp` test also failed before its minimal change | PASS — focused Homebrew tests: 4/4 | PASS — isolated success and upstream-failure cases prove cleanup, propagation, and no false success; Go module owns normal `brew` skip and upstream owns direct repeat | PASS — extracted only `runPOSIXScriptWithEnv` test helper; no unrelated harness refactor |
| T-074 | `internal/installer/runner/runner_test.go` | Deterministic `/bin/sh` contract | PASS — `TestShippedInstallerScriptsRunUnderPOSIXSh` and `TestShippedInstallerScriptsReturnDownloadFailure`: 8/8 cases before production edits | PASS — four new Oh My Zsh tests failed before the production change: missing script-owned unattended environment, no installed-directory skip, and no early dependency validation | PASS — focused Oh My Zsh tests: 7/7 | PASS — success, installed-directory skip, missing `curl`/`mktemp`/`sh`, and installer-failure cases cover distinct contract paths | PASS — shared shipped-script tests set `HOME` to `t.TempDir()`, and both blind judges verified the fix in Round 2 |
| T-075 | `internal/installer/runner/runner_test.go` | Deterministic `/bin/sh` contract | PASS — shared shipped-installer safety net: 8/8 cases | PASS — seven new SDKMAN cases failed before production edits: no installed-state skip, no explicit `curl`/`mktemp`/`bash` validation, and no `SDKMAN_DIR` invocation; later Judgment Day RED test rejected fake Bash 3 before download | PASS — focused SDKMAN tests: 7/7, then 11/11 after the authorized Bash-version fix | PASS — official URL/invocation, installed-state skip, three missing-dependency cases, deterministic Bash 3/invalid rejection and Bash 4 acceptance, cleanup, propagated installer failure, and no false success | PASS — cross-platform path assertion was relaxed to verify the observable SDKMAN directory contract; version tests fake `bash --version` and do not inspect the host shell |

## Verification

- PASS: `rtk go test ./internal/installer/runner -run '^TestHomebrewInstallScript' -count=1` — 4 passed.
- PASS: `rtk go test ./...` — 569 passed in 12 packages.
- PASS: `rtk git diff --check`.
- PASS: In disposable WSL2 Ubuntu 24.04 distro `Test`, installed apt package `shellcheck` version `0.9.0-1`; `shellcheck --version` reported `version: 0.9.0`.
- PASS: From the mounted repository via `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — exact gate command returned zero with no findings.
- PASS: `rtk go test ./internal/installer/runner -run '^TestOhMyZshInstallScript' -count=1` — 7 passed.
- PASS: `rtk go test ./...` — 576 passed in 12 packages.
- PASS: From the mounted repository via `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — exact gate command returned zero with no findings (WSL emitted a non-fatal systemd user-session warning).
- PASS: `gofmt -d internal/installer/runner/runner_test.go` and `git diff --check` — no formatting or whitespace errors.
- PASS: `rtk go test ./internal/installer/runner -run '^(TestShippedInstallerScriptsRunUnderPOSIXSh|TestShippedInstallerScriptsReturnDownloadFailure)$' -count=1` — 8 passed.
- PASS: `go test ./...` — 12 packages passed.
- PASS: `gofmt -d internal/installer/runner/runner_test.go` and `git diff --check` — no Go-format output or whitespace errors.
- PASS: `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — zero findings; WSL emitted a non-fatal systemd user-session warning.
- PASS: RED `rtk go test ./internal/installer/runner -run '^TestSdkmanInstallScriptRequiresBashFourOrNewerBeforeDownloading$' -count=1` — 0 passed, 3 failed before the Bash-version production change.
- PASS: GREEN `rtk go test ./internal/installer/runner -run '^TestSdkmanInstallScript' -count=1` — 11 passed after the Bash-version production change.
- PASS: `rtk go test ./internal/installer/runner -run '^(TestShippedInstallerScriptsRunUnderPOSIXSh|TestShippedInstallerScriptsReturnDownloadFailure)$' -count=1` — 8 passed after the Bash-version production change.
- PASS: `rtk go test ./...` — 587 passed in 12 packages after the Bash-version production change.
- PASS: `gofmt -d internal/installer/runner/runner_test.go` and `rtk git diff --check` — no output or whitespace errors after the Bash-version production change.
- PASS: `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — zero findings after the Bash-version production change; WSL emitted a non-fatal systemd user-session warning.
- PASS: Judgment Day Round 2 — both blind judges verified the deterministic `HOME` isolation fix; terminal `JUDGMENT: APPROVED`.
- PASS: `rtk go test ./internal/installer/runner -run '^(TestShippedInstallerScriptsRunUnderPOSIXSh|TestShippedInstallerScriptsReturnDownloadFailure)$' -count=1` — 8 passed before T-075 edits.
- PASS: RED `rtk go test ./internal/installer/runner -run '^TestSdkmanInstallScript' -count=1` — 0 passed, 7 failed before T-075 production edits.
- PASS: GREEN `rtk go test ./internal/installer/runner -run '^TestSdkmanInstallScript' -count=1` — 7 passed.
- PASS: `gofmt -d internal/installer/runner/runner_test.go` and `rtk git diff --check` — no output or whitespace errors.
- PASS: `rtk go test ./...` — 583 passed in 12 packages.
- PASS: `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — zero findings; WSL emitted a non-fatal systemd user-session warning.

## Completed Changes

- `assets/scripts/homebrew-install.sh`: fail early with explicit errors if `curl` or `mktemp` is unavailable.
- `internal/installer/runner/runner_test.go`: isolate HOME/TMPDIR/PATH and fake command behavior for dependency, success/cleanup, and upstream-failure contracts.
- `assets/scripts/omz-install.sh`: skip an existing `$HOME/.oh-my-zsh`, validate `curl`, `mktemp`, and `sh` before mutation, and invoke the official installer with script-owned `RUNZSH=no CHSH=no`.
- `internal/installer/runner/runner_test.go`: add isolated Oh My Zsh success, skip, dependency, cleanup, failure-propagation, and no-false-success contract tests.
- `internal/installer/runner/runner_test.go`: isolate `HOME` with `t.TempDir()` in both shared shipped-script tests so the Oh My Zsh installed-state skip cannot depend on the developer machine.
- `docs/tasks.md` and `openspec/changes/phase-4-shell-scripts/tasks.md`: mark only T-074 complete.
- `assets/scripts/sdkman-install.sh`: skip a completed local SDKMAN install, validate `curl`, `mktemp`, and `bash` before mutation, download from the official `https://get.sdkman.io` endpoint, and invoke the installer with `SDKMAN_DIR`.
- `internal/installer/runner/runner_test.go`: add isolated SDKMAN official-invocation, installed-state skip, dependency, cleanup, failure-propagation, and no-false-success contracts.
- `assets/scripts/sdkman-install.sh`: validate the resolved `bash --version` major version before downloading; reject Bash 3.x or unparseable output with an actionable Bash 4+ error.
- `internal/installer/runner/runner_test.go`: use fake `bash --version` output to prove Bash 3 and invalid versions cause no download, while Bash 4 proceeds.
- `docs/tasks.md` and `openspec/changes/phase-4-shell-scripts/tasks.md`: mark only T-075 complete.

## Completion

T-073, T-074, and T-075 are complete. T-076 onward and Phase 3 remain untouched. This work unit is intentionally uncommitted; no PR action was taken.
