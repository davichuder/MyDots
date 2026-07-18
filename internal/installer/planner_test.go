package installer

import (
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

// --- T-032: BuildPlan tests ---

func TestBuildPlan_Defaults(t *testing.T) {
	cfg := config.DefaultConfig()
	plan := BuildPlan(cfg, platform.Platform{})

	// Defaults: Java=false, PHP=false, nvim=base, framework=none
	// Disabled: M-14 Sdkman, M-15 Java, M-16 PHP, M-18 NeovimPersonal, M-19 NeovimFramework
	// Expected: 48 - 5 = 43
	if len(plan) != 43 {
		t.Errorf("BuildPlan with defaults: got %d modules, want 43", len(plan))
	}

	// Verify no disabled modules leaked into the plan
	disabled := map[ModuleID]bool{
		ModSdkman: true, ModJava: true, ModPhp: true,
		ModNeovimPersonal: true, ModNeovimFramework: true,
	}
	for _, m := range plan {
		if disabled[m.ID()] {
			t.Errorf("disabled module %s found in plan with defaults", m.ID())
		}
	}

	// Verify all non-disabled modules are present
	all := allModules()
	for _, m := range all {
		if disabled[m.ID()] {
			continue
		}
		found := false
		for _, p := range plan {
			if p.ID() == m.ID() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("non-disabled module %s missing from plan", m.ID())
		}
	}
}

func TestBuildPlan_JavaEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Languages.Java = true
	plan := BuildPlan(cfg, platform.Platform{})

	// M-14 (Sdkman) and M-15 (Java) should be in the plan
	found := map[ModuleID]bool{}
	for _, m := range plan {
		found[m.ID()] = true
	}
	if !found[ModSdkman] {
		t.Error("M-14 (Sdkman) should be in plan when Java=true")
	}
	if !found[ModJava] {
		t.Error("M-15 (Java) should be in plan when Java=true")
	}

	// M-16, M-18, M-19 should still be disabled
	if found[ModPhp] {
		t.Error("M-16 (PHP) should not be in plan when PHP=false")
	}
	if found[ModNeovimPersonal] {
		t.Error("M-18 (NeovimPersonal) should not be in plan with nvim=base")
	}
	if found[ModNeovimFramework] {
		t.Error("M-19 (NeovimFramework) should not be in plan with framework=none")
	}
}

func TestBuildPlan_PHPEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Languages.PHP = true
	plan := BuildPlan(cfg, platform.Platform{})

	found := false
	for _, m := range plan {
		if m.ID() == ModPhp {
			found = true
			break
		}
	}
	if !found {
		t.Error("M-16 (PHP) should be in plan when PHP=true")
	}
}

func TestBuildPlan_NvimPersonal(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Nvim.Config = config.NvimConfigPersonal
	plan := BuildPlan(cfg, platform.Platform{})

	// Find indices
	m18idx := -1
	m45idx := -1
	for i, m := range plan {
		switch m.ID() {
		case ModNeovimPersonal:
			m18idx = i
		case ModChezmoi:
			m45idx = i
		}
	}

	if m18idx == -1 {
		t.Error("M-18 (NeovimPersonal) should be in plan when nvim=personal")
	}
	if m45idx == -1 {
		t.Error("M-45 (Chezmoi) should be in plan")
	}
	if m45idx >= 0 && m18idx >= 0 && m45idx >= m18idx {
		t.Errorf("M-45 (Chezmoi) at index %d should precede M-18 at index %d", m45idx, m18idx)
	}
}

func TestBuildPlan_FrameworkLazyVim(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Nvim.Framework = config.NvimFrameworkLazyVim
	plan := BuildPlan(cfg, platform.Platform{})

	found := false
	for _, m := range plan {
		if m.ID() == ModNeovimFramework {
			found = true
			break
		}
	}
	if !found {
		t.Error("M-19 (NeovimFramework) should be in plan when framework=lazyvim")
	}
}

func TestBuildPlan_AllOptionalEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Languages.Java = true
	cfg.Languages.PHP = true
	cfg.Nvim.Config = config.NvimConfigPersonal
	cfg.Nvim.Framework = config.NvimFrameworkLazyVim
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 48 {
		t.Errorf("BuildPlan with all optional enabled: got %d modules, want 48", len(plan))
	}

	optionalIDs := []ModuleID{ModSdkman, ModJava, ModPhp, ModNeovimPersonal, ModNeovimFramework}
	for _, id := range optionalIDs {
		found := false
		for _, m := range plan {
			if m.ID() == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("optional module %s missing from plan when all enabled", id)
		}
	}
}

func TestBuildPlan_EdgeCases(t *testing.T) {
	t.Run("unknown nvim config treated as base", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Nvim.Config = "invalid"
		plan := BuildPlan(cfg, platform.Platform{})

		for _, m := range plan {
			if m.ID() == ModNeovimPersonal {
				t.Error("M-18 (NeovimPersonal) should be disabled when nvim.config is unknown")
			}
		}
	})

	t.Run("unknown framework treated as none", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Nvim.Framework = "invalid"
		plan := BuildPlan(cfg, platform.Platform{})

		for _, m := range plan {
			if m.ID() == ModNeovimFramework {
				t.Error("M-19 (NeovimFramework) should be disabled when framework is unknown")
			}
		}
	})
}

func TestBuildPlan_PreservesOrder(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Languages.Java = true
	cfg.Languages.PHP = true
	cfg.Nvim.Config = config.NvimConfigPersonal
	cfg.Nvim.Framework = config.NvimFrameworkLazyVim
	plan := BuildPlan(cfg, platform.Platform{})

	all := allModules()
	if len(plan) != len(all) {
		t.Fatalf("expected same length as allModules(): got %d, want %d", len(plan), len(all))
	}

	for i := range plan {
		if plan[i].ID() != all[i].ID() {
			t.Errorf("plan order mismatch at index %d: got %s, want %s",
				i, plan[i].ID(), all[i].ID())
		}
	}
}
