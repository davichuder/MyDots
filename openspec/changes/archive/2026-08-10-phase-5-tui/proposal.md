# Proposal: Phase 5 Interactive TUI

## Intent

Deliver the interactive Bubble Tea v2 experience for the existing installer so supported users can configure, run, monitor, understand, and manage MyDots safely. Invalid saved configuration must be visible and lead to configuration—not silent defaults; cancellation must be distinct from failure.

## Scope

### In Scope
- T-081–T-084: bounded 200-line log pane and status progress rows with golden renders.
- T-085–T-088: five-item main menu and six-step validated, persisted configuration wizard.
- T-089–T-092: installation progress bridge, cheatsheet toggle, cancellation outcome, and success/warning/critical result views.
- T-093–T-094: root routing, deterministic preflight, embedded WSL2 guide, and reference menu; existing embedded content renders, absent Phase 6 content says `not available`.
- T-095–T-096: create, list (including prior sessions), confirm deletion, and empty state for managed backups.

### Out of Scope
- Broad Phase 4 pipeline/status-contract changes; Phase 6 cheatsheet creation; Phase 7 CLI/entry-point flags.
- Network, root, real HOME, or host-dependent tests.

### Approved Phase 5 Integration Seam

`internal/installer/executor.go` is a narrowly approved integration seam, not Phase 4 scope expansion. `Run` checks cancellation before each next module and remains the canonical executor used by the Phase 5 `SessionRunner`; this preserves dependency, configured-state, shared-session, and cancellation semantics. Progress carries each module's human name alongside its typed ID so rendering never replaces typed routing/status identity.

## Capabilities

### New Capabilities
- `interactive-tui`: Routed configuration, installation, results, reference, preflight, and backup workflows over existing installer services.

### Modified Capabilities
- None.

## Approach

Add layered `internal/tui` screen models and reusable components. Inject config, backup, platform, embedded-asset, executor, and cancellation seams; map the canonical executor's literal `running` status and module names in the renderer while retaining typed IDs. Work strictly RED then GREEN for each T-081–T-096 pair: direct `Update`/`View` tests for units, fixed-dimension golden files for renders, and the approved native Bubble Tea v2 program harness for full-program T-093 behavior. Future commits are conventional, one per T-xxx pair, with decision rationale in the body.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/tui/**` | New | Root, components, seven screen flows, unit/integration/golden tests |
| `assets.go`, `assets/wsl2-guide.md` | Consumed | Injected embedded guide/reference source |
| `internal/config/**`, `internal/backup/**` | Consumed | Injected boundaries |
| `internal/installer/executor.go` | Minimal integration seam | Canonical cancellation guard and module-name progress metadata used by the Phase 5 adapter |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| v2/TUI render drift | Med | v2 APIs, direct model tests, fixed golden dimensions |
| Host-state leakage | Med | Temp dirs and injected seams |
| Backup path coordination gap | Med | Define TUI-owned managed-path/session input |

## Rollback Plan

Revert the Phase 5 T-xxx commits, including the approved executor integration seam; no broad installer pipeline, persisted-data migration, or Phase 6/7 behavior is changed.

## Dependencies

- Existing Bubble Tea v2, Huh, teatest, golden dependencies; `assets/wsl2-guide.md`; installer/config/backup APIs.
- Strict TDD gate: `go test ./...`.

## Success Criteria

- [ ] T-081–T-096 pass RED-to-GREEN evidence, direct-model tests, deterministic goldens, and `go test ./...`.
- [ ] Supported flows route correctly; Windows shows the guide and exits; config errors enter the assistant; cancellation is separately reported.
- [ ] Backup menu sees prior managed backups; unavailable future reference content explicitly reports `not available`.
