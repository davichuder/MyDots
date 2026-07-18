package installer

import (
	"testing"
)

// expectedOrder defines the canonical execution order from specs.md §5.
var expectedOrder = []ModuleID{
	ModHomebrew,       //  0 — Exec #1  — CRITICAL
	ModGhostty,        //  1 — Exec #2
	ModNerdFont,       //  2 — Exec #3
	ModGit,            //  3 — Exec #4
	ModGitCredOAuth,   //  4 — Exec #5
	ModChezmoi,        //  5 — Exec #6
	ModZsh,            //  6 — Exec #7
	ModOhMyZsh,        //  7 — Exec #8
	ModZellij,         //  8 — Exec #9
	ModFnm,            //  9 — Exec #10
	ModNode,           // 10 — Exec #11
	ModUv,             // 11 — Exec #12
	ModPython,         // 12 — Exec #13
	ModGo,             // 13 — Exec #14
	ModCppToolchain,   // 14 — Exec #15
	ModSdkman,         // 15 — Exec #16
	ModJava,           // 16 — Exec #17
	ModPhp,            // 17 — Exec #18
	ModNeovim,         // 18 — Exec #19
	ModNeovimPersonal, // 19 — Exec #20
	ModNeovimFramework,// 20 — Exec #21
	ModTheme,          // 21 — Exec #22
	ModAtuin,          // 22 — Exec #23
	ModZoxide,         // 23 — Exec #24
	ModBat,            // 24 — Exec #25
	ModEza,            // 25 — Exec #26
	ModFd,             // 26 — Exec #27
	ModRipgrep,        // 27 — Exec #28
	ModFzf,            // 28 — Exec #29
	ModSd,             // 29 — Exec #30
	ModJq,             // 30 — Exec #31
	ModYq,             // 31 — Exec #32
	ModTldr,           // 32 — Exec #33
	ModDelta,          // 33 — Exec #34
	ModBottom,         // 34 — Exec #35
	ModThefuck,        // 35 — Exec #36
	ModCarapace,       // 36 — Exec #37
	ModGlow,           // 37 — Exec #38
	ModGh,             // 38 — Exec #39
	ModClipboard,      // 39 — Exec #40
	ModDocker,         // 40 — Exec #41
	ModLazydocker,     // 41 — Exec #42
	ModLazygit,        // 42 — Exec #43
	ModOpencode,       // 43 — Exec #44
	ModMcpConfig,      // 44 — Exec #45
	ModRtk,            // 45 — Exec #46
	ModCaveman,        // 46 — Exec #47
	ModGentleAi,       // 47 — Exec #48
}

func TestAllModules_Count(t *testing.T) {
	modules := allModules()
	if len(modules) != 48 {
		t.Errorf("allModules() returned %d modules, want 48", len(modules))
	}
}

func TestAllModules_NoDuplicateIDs(t *testing.T) {
	modules := allModules()
	seen := make(map[ModuleID]bool, len(modules))
	for _, m := range modules {
		id := m.ID()
		if seen[id] {
			t.Errorf("duplicate ModuleID %q found", id)
		}
		seen[id] = true
	}
	if len(seen) != len(modules) {
		t.Errorf("expected %d unique IDs, got %d", len(modules), len(seen))
	}
}

func TestAllModules_FirstIsCritical(t *testing.T) {
	modules := allModules()
	if len(modules) == 0 {
		t.Fatal("allModules() returned empty slice")
	}
	first := modules[0]
	if first.ID() != ModHomebrew {
		t.Errorf("first module ID = %q, want %q", first.ID(), ModHomebrew)
	}
	if first.Criticality() != Critical {
		t.Errorf("first module Criticality = %q, want %q", first.Criticality(), Critical)
	}
}

func TestAllModules_AllNonCriticalExceptFirst(t *testing.T) {
	modules := allModules()
	if len(modules) < 2 {
		t.Fatal("allModules() must have at least 2 modules for this test")
	}
	for i, m := range modules[1:] {
		if m.Criticality() != NonCritical {
			t.Errorf("modules[%d] (%s) Criticality = %q, want %q", i+1, m.ID(), m.Criticality(), NonCritical)
		}
	}
}

func TestAllModules_Order(t *testing.T) {
	modules := allModules()
	if len(modules) != len(expectedOrder) {
		t.Fatalf("allModules() returned %d modules, expectedOrder has %d", len(modules), len(expectedOrder))
	}
	for i, want := range expectedOrder {
		got := modules[i].ID()
		if got != want {
			t.Errorf("allModules()[%d] = %q, want %q", i, got, want)
		}
	}
}

func TestAllModules_ChezmoiBeforeNvimPersonal(t *testing.T) {
	modules := allModules()
	// M-45 chezmoi at exec position 6 → index 5
	if modules[5].ID() != ModChezmoi {
		t.Errorf("modules[5] = %q, want M-45 (chezmoi)", modules[5].ID())
	}
	// M-18 NeovimPersonal at exec position 20 → index 19
	if modules[19].ID() != ModNeovimPersonal {
		t.Errorf("modules[19] = %q, want M-18 (NeovimPersonal)", modules[19].ID())
	}
}

func TestAllModules_GhosttyBeforeTheme(t *testing.T) {
	modules := allModules()
	// M-48 Ghostty at exec position 2 → index 1
	if modules[1].ID() != ModGhostty {
		t.Errorf("modules[1] = %q, want M-48 (Ghostty)", modules[1].ID())
	}
	// M-46 Theme at exec position 22 → index 21
	if modules[21].ID() != ModTheme {
		t.Errorf("modules[21] = %q, want M-46 (Theme)", modules[21].ID())
	}
}

func TestAllModules_DependenciesExist(t *testing.T) {
	modules := allModules()
	idSet := make(map[ModuleID]bool, len(modules))
	for _, m := range modules {
		idSet[m.ID()] = true
	}
	for _, m := range modules {
		for _, dep := range m.Dependencies() {
			if !idSet[dep] {
				t.Errorf("module %q depends on %q which does not exist in allModules()", m.ID(), dep)
			}
		}
	}
}
