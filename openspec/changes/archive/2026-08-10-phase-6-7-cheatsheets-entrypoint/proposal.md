# Proposal: Phase 6–7 Cheatsheets and Entrypoint

## Intent

Complete only T-097–T-100: ship discoverable embedded references for all 48 tools and a deterministic CLI entrypoint for supported routes, unattended execution, and defaults.

## Scope

### In Scope
- T-097 RED→GREEN: validate the exact 48 embedded cheatsheet paths, readable content, and SCR-07 headings before adding documents.
- T-098 RED→T-099 GREEN: test and implement flag/platform/config routing through an injectable entrypoint seam.
- T-100: prove privileged startup ordering with a separate RED→GREEN only when T-098 does not already prove it.

### Out of Scope
- T-101+ BDD acceptance work, distribution, and unrelated TUI changes.
- Changing task markers, Phase 5 artifacts, or live host/network/root behavior.

## Capabilities

### New Capabilities
- `embedded-cheatsheets`: complete embedded reference inventory with SCR-07 content contract.
- `application-entrypoint`: deterministic CLI routing for version, platform, TUI, unattended, and default-config flows.

### Modified Capabilities
None.

## Approach

Start with T-097 RED→GREEN. The RED enumerates every required file and reads it through the embedded FS, asserting `## Links`, `## Key Shortcuts`, and `## Usage Examples`; GREEN adds the 48 Markdown files. Use one T-098 route suite as the shared RED for T-099 and route-level T-100. Inject IO, config/path, TUI, pipeline, elevation, keepalive, args, and platform dependencies; use `t.TempDir()` only. Add a focused T-100 pair solely for uncovered elevation/keepalive-before-pipeline ordering or lifecycle ownership. Commit each verified RED→GREEN unit; create one PR only after both phases are verified.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `assets/cheatsheets/*.md` | New | 48 embedded tool references. |
| `assets.go`, `assets_test.go` | Modified/New | Embedded inventory validation seam. |
| `main.go`, `main_test.go` | Modified/New | Thin wrapper, routing, and deterministic tests. |
| `internal/{config,sudo,tui,installer}/` | Modified | Injectable entrypoint dependencies only if needed. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Filename/content drift | Med | Exact embedded-FS inventory RED. |
| Host side effects in tests | Med | Temp paths and injected functions; no real HOME/network/root/TUI. |
| Oversized review | High | Coherent RED→GREEN units; report unit-only size exception if needed. |

## Rollback Plan

Revert the affected RED→GREEN work unit. Remove added cheatsheets with its validation test, or restore the previous entrypoint seam without touching Phase 5.

## Dependencies

- Existing `embed.FS`, config, TUI, pipeline, and sudo boundaries.

## Success Criteria

- [ ] All 48 cheatsheets are embedded, readable, and contain the three SCR-07 headings.
- [ ] Every T-098 route has deterministic coverage with FR-22 output and exit codes preserved.
- [ ] Privileged routes acquire elevation and start keepalive before pipeline startup, without redundant T-100 tests.
