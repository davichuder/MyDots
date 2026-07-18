package installer

import (
	"testing"
)

// --- T-030: Catalogue tests ---

func TestAllModules_Returns48Modules(t *testing.T) {
	modules := allModules()
	if len(modules) != 48 {
		t.Errorf("allModules() returned %d modules, want 48", len(modules))
	}
}

func TestAllModules_NoDuplicateIDs(t *testing.T) {
	modules := allModules()
	seen := make(map[ModuleID]bool)
	for _, m := range modules {
		id := m.ID()
		if seen[id] {
			t.Errorf("duplicate module ID: %s", id)
		}
		seen[id] = true
	}
}

func TestAllModules_FirstIsHomebrewCritical(t *testing.T) {
	modules := allModules()
	if len(modules) == 0 {
		t.Fatal("allModules() returned empty slice")
	}

	first := modules[0]
	if first.ID() != ModHomebrew {
		t.Errorf("first module ID = %s, want %s", first.ID(), ModHomebrew)
	}
	if first.Criticality() != Critical {
		t.Errorf("first module Criticality = %v, want Critical", first.Criticality())
	}
}

func TestAllModules_Remaining47AreNonCritical(t *testing.T) {
	modules := allModules()
	for i := 1; i < len(modules); i++ {
		if modules[i].Criticality() != NonCritical {
			t.Errorf("module at index %d (ID=%s) Criticality = %v, want NonCritical",
				i, modules[i].ID(), modules[i].Criticality())
		}
	}
}

func TestAllModules_Order(t *testing.T) {
	modules := allModules()

	// Chezmoi (M-45) at exec position 6 (index 5)
	if modules[5].ID() != ModChezmoi {
		t.Errorf("index 5: got %s, want %s (ModChezmoi, M-45)", modules[5].ID(), ModChezmoi)
	}

	// NeovimPersonal (M-18) at exec position 20 (index 19)
	if modules[19].ID() != ModNeovimPersonal {
		t.Errorf("index 19: got %s, want %s (ModNeovimPersonal, M-18)", modules[19].ID(), ModNeovimPersonal)
	}

	// Ghostty (M-48) at exec position 2 (index 1)
	if modules[1].ID() != ModGhostty {
		t.Errorf("index 1: got %s, want %s (ModGhostty, M-48)", modules[1].ID(), ModGhostty)
	}

	// Theme (M-46) at exec position 22 (index 21)
	if modules[21].ID() != ModTheme {
		t.Errorf("index 21: got %s, want %s (ModTheme, M-46)", modules[21].ID(), ModTheme)
	}
}

func TestAllModules_DependenciesExist(t *testing.T) {
	modules := allModules()
	ids := make(map[ModuleID]bool)
	for _, m := range modules {
		ids[m.ID()] = true
	}

	for _, m := range modules {
		for _, dep := range m.Dependencies() {
			if !ids[dep] {
				t.Errorf("module %s depends on %s, which is not present in allModules()", m.ID(), dep)
			}
		}
	}
}

func TestAllModules_ListIsComplete(t *testing.T) {
	modules := allModules()
	ids := make(map[ModuleID]bool)
	for _, m := range modules {
		ids[m.ID()] = true
	}

	// Verify all known ModuleID constants are present
	expected := []ModuleID{
		ModHomebrew, ModZsh, ModOhMyZsh, ModZellij,
		ModGit, ModGitCredOAuth, ModLazygit, ModFnm,
		ModNode, ModUv, ModPython, ModGo,
		ModCppToolchain, ModSdkman, ModJava, ModPhp,
		ModNeovim, ModNeovimPersonal, ModNeovimFramework,
		ModAtuin, ModZoxide, ModBat, ModEza,
		ModFd, ModRipgrep, ModFzf, ModSd,
		ModJq, ModYq, ModTldr, ModDelta,
		ModBottom, ModThefuck, ModCarapace, ModGlow,
		ModGh, ModClipboard, ModDocker, ModLazydocker,
		ModOpencode, ModRtk, ModCaveman, ModGentleAi,
		ModMcpConfig, ModChezmoi, ModTheme, ModNerdFont,
		ModGhostty,
	}

	for _, id := range expected {
		if !ids[id] {
			t.Errorf("allModules() is missing module %s", id)
		}
	}
}
