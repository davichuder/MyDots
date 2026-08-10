## Exploration: phase-6-7-cheatsheets-entrypoint

### Current State
The scoped work is currently unimplemented. `assets/cheatsheets/` contains only `.gitkeep`, while `assets.go` embeds the whole `assets` tree through `embed.FS`; the repository therefore has an existing compile-time asset seam but no 48 tool documents to validate. `main.go` is an empty wrapper. The TUI root already exposes injected platform, asset, configuration, and screen-factory dependencies, and the installer runner exposes a context-based command executor seam. Configuration supports `DefaultConfig`, `Load`, `Save`, validation, and an injectable home-directory function, but `DefaultConfigPath` still resolves the process home by default. `sudo.RequestElevation` and `sudo.StartKeepalive` are package functions; the former binds process stdio and the latter launches a goroutine with a real ticker.

`docs/tasks.md` defines T-097 as 48 embedded Markdown files with the SCR-07 headings, then T-098 as failing entry-point tests and T-099/T-100 as the implementation. `docs/design.md` §13 routes `--version`, Windows, unsupported OS, `--unattended`, `--default`, and the default TUI path. FR-22 requires the exact missing-config message. There is no existing OpenSpec config at the repository root; the established change convention is a named directory containing phase artifacts, and the Phase 5 directory must remain untouched.

### Affected Areas
- `assets/cheatsheets/*.md` — 48 new embedded tool references required by T-097; each needs the three SCR-07 headings and deterministic, useful content.
- `assets.go` — existing `embed.FS` is the repository/compile-time validation seam; no production change should be presumed until a RED test proves one is needed.
- `main.go` — currently empty; becomes the thin `os.Args`/`runtime.GOOS` wrapper and entry-point implementation boundary for T-098–T-100.
- `internal/tui/app.go`, `internal/tui/dependencies.go` — existing injected TUI boundary for the no-flag route and Windows guide behavior.
- `internal/config/config.go` — config load/save/default path behavior used by unattended/default routes; tests should pass `t.TempDir()` paths rather than alter HOME.
- `internal/sudo/sudo.go` — elevation/keepalive side effects required before unattended/default pipeline startup; calls need injectable function boundaries (or an entry-point dependency bundle) so tests never invoke sudo or real host state.
- `internal/installer/runner/runner.go` and installer pipeline constructors — likely pipeline seam for proving startup without network, root, or external commands; use injected executor and context.
- `docs/tasks.md` — only T-097 through T-100 are in scope; task markers must remain unchanged until later verified implementation.
- `docs/specs.md`, `docs/design.md` — normative SCR-07, FR-22, SC-12/SC-13, and §13 behavior to preserve.

### Approaches
1. **Repository validation, then grouped entry-point seams** — first create a RED validation test that enumerates the required 48 embedded files, checks each is readable through the embedded FS, and verifies all three headings; then add the 48 documents as GREEN. Separately write one T-098 routing/behavior RED suite, with focused subtests for version, Windows, unsupported OS, unattended config presence/absence, default persistence, and TUI selection. Implement T-099 as the shared routing/runtime GREEN, then implement T-100 as a separate RED/GREEN only if elevation ordering is not already fully specified by the T-098 startup assertions.
   - Pros: catches omissions before embedding content; treats T-098 as the shared behavioral safety net for T-099 and most of T-100; keeps commits behavior-oriented and deterministic.
   - Cons: the asset validation test is a temporary or retained repository contract and the entry-point dependency bundle needs careful design.
   - Effort: High

2. **Content first, then one monolithic entry-point pair** — add all cheatsheets without a validation RED, and make T-098 test every route including elevation before implementing T-099/T-100 together.
   - Pros: fewer apparent units and less test scaffolding.
   - Cons: a typo, missing file, or wrong embed path can survive until later reference-menu use; combining routing, pipeline construction, sudo ordering, and TUI startup makes failures ambiguous and risks exceeding the 400-line modification budget.
   - Effort: High

### Recommendation
Use Approach 1. T-097 **does need a repository/embedding validation RED before creating the 48 files**: the current directory is empty except for `.gitkeep`, and `embed.FS` only proves files exist at compile time, not that the complete task inventory, filenames, headings, and readable content are present. The validation should enumerate the explicit task list and read through `assets` (or a narrow exported/test package seam), without network or host assumptions.

Treat T-098 as the shared RED for T-099 plus the route-level portion of T-100, not as four independent task-number pairs. Its tests should use injected stdout/stderr, config path/load/save functions, TUI runner, WSL2 renderer, pipeline starter, elevation function, keepalive function, context, and platform/argument inputs. `t.TempDir()` should supply config storage. T-099 is the first entry-point GREEN: implement routing and thin wrappers while preserving exact FR-22 output and exit codes. Then use a focused T-100 RED for ordering and cancellation/keepalive ownership if the initial T-098 suite cannot prove that elevation and keepalive occur before pipeline start; implement T-100 as the second entry-point work unit. This avoids claiming T-100 is independently testable when its behavior is a prerequisite/order constraint of the shared startup flow.

Proposed work units:
1. **T-097 RED → GREEN: embedded cheatsheet inventory** — RED enumerates all 48 paths and SCR-07 headings; GREEN adds the 48 Markdown files and keeps the validation green.
2. **T-098 RED → T-099 GREEN: entry-point routing and injectable runtime** — RED covers all route outcomes and side-effect selection; GREEN adds `run`, the thin `main`, guide/TUI/unattended/default handlers and injected seams.
3. **T-100 RED → GREEN: privileged startup ordering** — RED proves elevation and keepalive precede pipeline start and are not invoked on TUI/guide/error routes; GREEN wires the ordered startup lifecycle and cancellation ownership. If unit 2's RED already proves this completely, retain the evidence in unit 2 and do not manufacture a redundant pair; report T-100 as covered by the shared RED/GREEN.

### Risks
- The 48-name inventory includes hyphenated and compound IDs; filename drift would break reference-menu lookup even if embedding succeeds.
- A test that uses the package-global `assets` directly may prove only the current build, so the inventory must assert exact paths and headings rather than merely count files.
- Existing package-global sudo hooks and `runner` executor state can leak between parallel tests; restore hooks or encapsulate dependencies in the entry-point seam.
- `DefaultConfigPath` and pipeline code can touch HOME, network, root, or host tools unless callers inject paths and command execution.
- Bubble Tea `tea.Program` startup must be replaced by an injected runner in tests; do not spawn a real terminal or process.
- The combined change has 48 new files, so the 800-line new-file budget applies only when new files predominate. The entry-point modifications/tests should target roughly 250–400 changed lines; if the cheatsheet prose averages 8–12 lines each, the whole change is likely roughly 700–1,050 changed lines and should be reported as a size exception or split into unit-only review slices while preserving the requested single-PR delivery.

### Ready for Proposal
Yes. The proposal should preserve the T-097–T-100-only scope, explicitly require the T-097 embedding inventory RED, model T-098 as the shared entry-point RED, and make T-100 conditional on uncovered ordering behavior rather than forcing task-number pairing. The first recommended implementation unit is T-097 RED → GREEN, followed by the T-098 RED → T-099 GREEN entry-point unit.
