# Application Entrypoint Specification

## Purpose

Define only T-098–T-100 CLI routing and privileged startup.

## Requirements

### Requirement: Deterministic Route Contract

The application MUST route in this order: `--version`; Windows guide; unsupported-platform error; then Darwin/Linux mode. Unattended MUST load saved configuration, default MUST save defaults, and no mode flag MUST start TUI. Privileged modes MUST start the pipeline without TUI.

#### Scenario: Version route
- GIVEN `--version` and any platform
- WHEN routing runs
- THEN stdout MUST contain `buildVersion`, code MUST be 0, and nothing else runs

#### Scenario: Platform routes
- GIVEN no version flag and an injected platform
- WHEN routing runs
- THEN Windows MUST run only the guide with code 0
- AND unsupported platforms MUST use stderr and code 1 without guide, TUI, or pipeline

#### Scenario: Supported mode routes
- GIVEN Darwin or Linux
- WHEN unattended, default, or no-flag input runs
- THEN routes MUST respectively load then pipeline, save defaults then pipeline, or run only TUI

### Requirement: Exact FR-22 Failure

Missing `mydots-config.json` under `--unattended` MUST return 1, print exactly `Error: no config found. Run 'mydots' to configure first.` to stderr, and start no TUI, elevation, keepalive, or pipeline.

#### Scenario: Saved configuration is absent
- GIVEN an absent injected configuration path
- WHEN unattended runs
- THEN the exact FR-22 stderr and code 1 MUST occur without side effects

### Requirement: Isolated Injectable Verification

Tests MUST inject arguments, platform, stdout/stderr, config path/load/save/default, guide, TUI, elevation, keepalive, context/lifecycle, and pipeline. They MUST use temporary storage and MUST NOT use real HOME, network, root, TUI, processes, timers, terminal, or host state.

#### Scenario: Routes are tested in isolation
- GIVEN recording fakes and temporary storage
- WHEN routes and failures run
- THEN outputs, codes, calls, omissions, and order MUST be deterministic

### Requirement: Privileged Startup Ordering and Lifecycle

Unattended and default MUST order elevation → keepalive → pipeline. Failure MUST prevent later operations. Keepalive ownership MUST end with pipeline execution; other routes MUST invoke none of them.

#### Scenario: Privileged route succeeds
- GIVEN valid unattended or default input
- WHEN startup runs
- THEN order MUST be elevation → keepalive → pipeline, followed by keepalive shutdown

#### Scenario: Privileged startup fails
- GIVEN elevation or another pre-pipeline dependency fails
- WHEN startup runs
- THEN code MUST be 1 and later operations MUST NOT start

### Requirement: Shared Strict-TDD Evidence

T-098 MUST be the shared RED for T-099 GREEN and route-level T-100. T-100 MUST have a separate RED→GREEN only for genuinely uncovered ordering/lifecycle. Coherent units MUST be verified before one post-Phase-6-and-7 PR; a reported unit-only exception MAY cover the 400 modified/800 new-file-predominant budgets.

#### Scenario: TDD units are reviewed
- GIVEN T-098 fails before T-099 and T-100 coverage is known
- WHEN evidence is grouped
- THEN T-099 MUST pass it and T-100 MUST add RED only for a genuine gap
