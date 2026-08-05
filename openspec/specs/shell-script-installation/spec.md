# Shell Script Installation Specification

## Purpose

Specify T-073–T-080 without reopening Phase 3.

## Requirements

### Requirement: Audit-First Completion

Each task MUST audit first, preserve correct behavior, and remediate only proven gaps. Phase 3 MUST NOT be reopened.

#### Scenario: Existing behavior is correct
- GIVEN focused evidence satisfies this specification
- WHEN audit finishes
- THEN the behavior remains unchanged and is not recreated

#### Scenario: Audit proves a gap
- GIVEN focused evidence proves a requirement unmet
- WHEN remediation occurs
- THEN changes are limited to that gap

### Requirement: Runtime and Shell Boundary

Installers MUST support applicable macOS and Ubuntu/WSL2, execute compatibly with POSIX `/bin/sh`, and reject native Windows hosts.

#### Scenario: Supported runtime
- GIVEN a supported macOS or Ubuntu/WSL2 runtime
- WHEN `/bin/sh` executes a shipped installer
- THEN the wrapper uses only POSIX-compatible behavior

#### Scenario: Unsupported host
- GIVEN a native Windows host
- WHEN installation eligibility is evaluated
- THEN installation does not begin and unsupported status is explicit

### Requirement: Installer Outcome Contracts

T-073 Homebrew, T-074 Oh My Zsh, T-075 SDKMAN, T-076 Docker Linux, T-077 Nerd Font, T-078 Ghostty Linux, and T-079 Caveman MUST each validate dependencies, propagate failures, clean owned temporary resources, and never report false success. Each MUST define whether repeats skip satisfied local work or safely defer upstream. Oh My Zsh MUST be unattended; Docker MUST avoid redundant install or group mutation; Nerd Font MUST validate the requested font before mutation.

#### Scenario: Successful installation
- GIVEN dependencies exist and installation is needed
- WHEN every required action completes
- THEN the installer reports success and removes temporary resources

#### Scenario: Dependency or action fails
- GIVEN a missing dependency or failed action
- WHEN any listed installer runs
- THEN it fails, cleans temporary resources, and stops without success

#### Scenario: Repeated invocation
- GIVEN satisfied state or upstream-owned repeat handling
- WHEN installation repeats
- THEN the declared boundary prevents destructive duplicate local mutation

### Requirement: Ghostty Ubuntu Source Contract

Ghostty on Ubuntu MUST use community `ghostty-ubuntu`, never a nonexistent official Ubuntu tarball. It MUST download temporarily and MUST NOT use `curl | bash`.

#### Scenario: Community installer succeeds
- GIVEN eligible Ubuntu and available dependencies
- WHEN `ghostty-ubuntu` succeeds
- THEN its temporary download is removed and success MAY be reported

#### Scenario: Community installer fails
- GIVEN `ghostty-ubuntu` download or execution fails
- WHEN the wrapper handles failure
- THEN it propagates failure, cleans up, and never uses an official-tarball contract

### Requirement: Deterministic Verification Isolation

Verification MUST cover success, failure, cleanup, dependencies, idempotence, and false-success prevention without network, root, or system mutation.

#### Scenario: Isolated verification
- GIVEN deterministic substitutes and isolated state
- WHEN installer contracts are verified
- THEN no live network, privilege, or host mutation occurs

### Requirement: Mandatory ShellCheck Gate

Because `.github/workflows/` is absent, a new mandatory CI gate MUST guarantee ShellCheck and execute `shellcheck --shell=sh` over every shipped script. It MUST fail when unavailable or non-zero.

#### Scenario: Gate passes
- GIVEN ShellCheck is available and all shipped scripts pass `--shell=sh`
- WHEN CI runs the mandatory gate
- THEN the gate passes

#### Scenario: Gate cannot validate all scripts
- GIVEN ShellCheck is unavailable or one script returns non-zero
- WHEN CI runs the mandatory gate
- THEN the gate fails without a false pass

### Requirement: Incremental Task Evidence

Each T-073–T-080 checkbox in `docs/tasks.md` MUST wait for its focused audit, required remediation, and focused verification pass. Final whole-phase verification MUST remain separate.

#### Scenario: Task earns completion
- GIVEN one task passes audit, needed remediation, and focused verification
- WHEN tracking is updated
- THEN only that task MAY be marked complete

#### Scenario: Evidence is incomplete
- GIVEN a script exists but a focused step has not passed
- WHEN tracking is evaluated
- THEN its checkbox remains incomplete regardless of other task status
