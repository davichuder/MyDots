# Tasks: Phase 6–7 Cheatsheets and Entrypoint

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | T-097: 600–780; T-098/T-099: 420–560; conditional T-100: 80–150 |
| 400-line / 800-new-file risk | High / Medium (T-097 only) |
| Chained PRs recommended | Yes; user-requested delivery remains one PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | T-097 embedded references | Final single PR | First apply scope; report a scope-exclusive size exception only if over budget. |
| 2 | T-098 RED + T-099 GREEN | Final single PR | Depends on Unit 1; one coherent commit. |
| 3 | Conditional T-100 gap | Final single PR | Only if Unit 2 evidence leaves a real ordering/lifecycle gap. |

## Phase 6: T-097 Embedded Cheatsheets

- [x] 1.1 **RED:** Create `assets_test.go` plus `internal/tui/screens/reference_slug_test.go`: read embedded `assets`, require the exact 48 sorted readable files, and require every `ModuleID` to resolve to its slug; reject missing, extra, or unreadable paths.
- [x] 1.2 Extend the unchanged RED to slice ordered `## Links`, `## Key Shortcuts`, and `## Usage Examples`; require concrete HTTP(S) links, action/key bullets, commands/examples, and reject TODO/TBD/filler. Capture genuine inventory/content/mapping failure output before documents or mapping exist.
- [x] 1.3 **GREEN:** Add the exact 48 concise `assets/cheatsheets/*.md` files and `internal/tui/screens/reference_slug.go` plus `reference_slug_test.go` mapping every 48 `ModuleID` values to those slugs.
- [x] 1.4 Run focused `go test . ./internal/tui/screens`, then `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run`, and `go build ./...`; no golden is expected—if added, use the repository `-update` flow then rerun without it. Independently verify and Judgment Day before markers.
- [x] 1.5 Only after independent PASS, mark T-097 in `docs/tasks.md`, then hybrid progress/review evidence and `openspec/.../tasks.md`; pre-commit review inventory/content/slug coverage and diff size. Commit tests, docs, and code together with a conventional-commit body explaining RED→GREEN rationale; no push/PR without authorization.

## Phase 7: Entrypoint Routes and Privileged Lifecycle

- [x] 2.1 **T-098 shared RED:** Add `entrypoint_test.go` direct top-level `run(args, goos)` table tests for version, Windows guide, unsupported stderr/1, exact FR-22, unattended/default pipeline without TUI, and no-flag TUI. Capture deterministic failing output.
- [x] 2.2 Inject `t.TempDir`, IO, config/path/load/save/default/validate, assets, TUI, pipeline, elevation, joinable keepalive, context, and fixed clock; fakes must prove calls, omissions, order, and joined shutdown with no HOME/network/root/process/timer/TUI/terminal use.
- [x] 2.3 **T-099 GREEN:** Modify `main.go`; add `entrypoint.go`; modify `internal/sudo/{sudo.go,sudo_test.go}` only as needed for `run` factory seams, `buildVersion`, routes, UTC timestamp, elevation → keepalive → pipeline, and Stop/join ownership.
- [x] 2.4 Repeat focused/full/race/vet/lint/build verification, independent verification, and Judgment Day; only then mark T-098/T-099 in `docs/tasks.md` and hybrid `openspec/.../tasks.md`/review evidence. Make a conventional RED→GREEN work-unit commit with body rationale; review lifecycle, seams, output, and size first.
- [x] 2.5 **T-100:** Resolved by shared T-098 evidence: deterministic route tests prove elevation → keepalive → pipeline order, later-call suppression on failures, and joined shutdown including the pipeline-failure path. No redundant Unit 3 was needed.

## Release Gate

- [x] 3.1 After all Phase 6 then Phase 7 units independently pass, report each unit’s additions+deletions and new-file count; require explicit `size:exception` if the single PR exceeds 400 lines (or Unit 1 exceeds 800 new-file-predominant lines). Reconciled from final verification: T-097 `ea642fb` is 687 changed lines with 51 new files, within the approved 800-line new-file-predominant exception; T-098–T-100 `36fde19` is 797 changed lines with 2 new files, within the approved 800-line exception. Archive commits are distinct reviewed archive units.
- [x] 3.2 Create no PR before all units are verified; never stage, push, merge, or open the authorized single PR without explicit authorization. Reconciled from final verification: no PR existed before full Phase 6–7 verification; the user now explicitly authorizes final verify → archive → PR. This reconciliation does not stage, commit, push, merge, or create a PR.
