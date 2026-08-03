# Apply Progress: Phase 4 Shell Scripts

## Delivery

- Mode: single PR with maintainer-approved `size:exception`.
- Work unit: T-073 only.
- Boundary: Homebrew wrapper dependency validation and deterministic contract tests; no T-074 onward or Phase 3 artifacts.

## Task Status

- [x] 1.1 **T-073** — implementation, deterministic Go evidence, and mandatory ShellCheck pass.
- [ ] 1.2 **T-074**
- [ ] 1.3 **T-075**
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

## Verification

- PASS: `rtk go test ./internal/installer/runner -run '^TestHomebrewInstallScript' -count=1` — 4 passed.
- PASS: `rtk go test ./...` — 569 passed in 12 packages.
- PASS: `rtk git diff --check`.
- PASS: In disposable WSL2 Ubuntu 24.04 distro `Test`, installed apt package `shellcheck` version `0.9.0-1`; `shellcheck --version` reported `version: 0.9.0`.
- PASS: From the mounted repository via `wsl.exe -d Test --cd "D:\Descargas\proyectos futuros\MyDots-1" -- bash -lc "shellcheck --shell=sh assets/scripts/*.sh"` — exact gate command returned zero with no findings.

## Completed Changes

- `assets/scripts/homebrew-install.sh`: fail early with explicit errors if `curl` or `mktemp` is unavailable.
- `internal/installer/runner/runner_test.go`: isolate HOME/TMPDIR/PATH and fake command behavior for dependency, success/cleanup, and upstream-failure contracts.

## Completion

T-073 is complete. Only its OpenSpec and repository task markers were updated; T-074 onward and Phase 3 remain untouched. The work unit is intentionally uncommitted for pre-commit review.
