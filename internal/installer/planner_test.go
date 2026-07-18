package installer

import (
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

func buildCfg(java, php bool, nvimCfg config.NvimConfig, framework config.NvimFramework) config.Config {
	cfg := config.DefaultConfig()
	cfg.Languages.Java = java
	cfg.Languages.PHP = php
	cfg.Nvim.Config = nvimCfg
	cfg.Nvim.Framework = framework
	return cfg
}

func TestBuildPlan_Defaults(t *testing.T) {
	cfg := config.DefaultConfig() // java=false, php=false, nvim=base, framework=none
	plan := BuildPlan(cfg, platform.Platform{})

	// 48 - 5 = 43
	if len(plan) != 43 {
		t.Errorf("expected 43 modules, got %d", len(plan))
	}

	// First is always M-01 Homebrew
	if plan[0].ID() != ModHomebrew {
		t.Errorf("first module should be Homebrew (M-01), got %s", plan[0].ID())
	}

	// Order matches allModules() with skips
	all := allModules()
	planIdx := 0
	for _, mod := range all {
		if isDisabled(mod, cfg) {
			continue
		}
		if plan[planIdx].ID() != mod.ID() {
			t.Errorf("position %d: expected %s, got %s", planIdx, mod.ID(), plan[planIdx].ID())
		}
		planIdx++
	}
}

func TestBuildPlan_JavaEnabled(t *testing.T) {
	cfg := buildCfg(true, false, config.NvimConfigBase, config.NvimFrameworkNone)
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 45 {
		t.Errorf("expected 45 modules, got %d", len(plan))
	}

	hasSdkman := false
	hasJava := false
	sdkmanBeforeJava := false
	for _, mod := range plan {
		if mod.ID() == ModSdkman {
			hasSdkman = true
			if !hasJava {
				sdkmanBeforeJava = true
			}
		}
		if mod.ID() == ModJava {
			hasJava = true
		}
	}
	if !hasSdkman {
		t.Error("expected Sdkman (M-14) to be included")
	}
	if !hasJava {
		t.Error("expected Java (M-15) to be included")
	}
	if !sdkmanBeforeJava {
		t.Error("Sdkman should appear before Java in the plan")
	}
}

func TestBuildPlan_PhpEnabled(t *testing.T) {
	cfg := buildCfg(false, true, config.NvimConfigBase, config.NvimFrameworkNone)
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 44 {
		t.Errorf("expected 44 modules, got %d", len(plan))
	}

	hasPhp := false
	for _, mod := range plan {
		if mod.ID() == ModPhp {
			hasPhp = true
			break
		}
	}
	if !hasPhp {
		t.Error("expected PHP (M-16) to be included")
	}
}

func TestBuildPlan_NvimPersonal(t *testing.T) {
	cfg := buildCfg(false, false, config.NvimConfigPersonal, config.NvimFrameworkNone)
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 44 {
		t.Errorf("expected 44 modules, got %d", len(plan))
	}

	hasNvimPersonal := false
	chezmoiBeforePersonal := false
	for _, mod := range plan {
		if mod.ID() == ModChezmoi {
			if !hasNvimPersonal {
				chezmoiBeforePersonal = true
			}
		}
		if mod.ID() == ModNeovimPersonal {
			hasNvimPersonal = true
		}
	}
	if !hasNvimPersonal {
		t.Error("expected NvimPersonal (M-18) to be included")
	}
	if !chezmoiBeforePersonal {
		t.Error("Chezmoi (M-45) should appear before NvimPersonal (M-18) in the plan")
	}
}

func TestBuildPlan_FrameworkLazyvim(t *testing.T) {
	cfg := buildCfg(false, false, config.NvimConfigBase, config.NvimFrameworkLazyVim)
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 44 {
		t.Errorf("expected 44 modules, got %d", len(plan))
	}

	hasFramework := false
	for _, mod := range plan {
		if mod.ID() == ModNeovimFramework {
			hasFramework = true
			break
		}
	}
	if !hasFramework {
		t.Error("expected NvimFramework (M-19) to be included")
	}
}

func TestBuildPlan_AllOptionalEnabled(t *testing.T) {
	cfg := buildCfg(true, true, config.NvimConfigPersonal, config.NvimFrameworkLazyVim)
	plan := BuildPlan(cfg, platform.Platform{})

	if len(plan) != 48 {
		t.Errorf("expected 48 modules, got %d", len(plan))
	}

	// Order matches allModules() exactly
	all := allModules()
	for i, mod := range all {
		if plan[i].ID() != mod.ID() {
			t.Errorf("position %d: expected %s, got %s", i, mod.ID(), plan[i].ID())
		}
	}
}

func TestBuildPlan_UnknownNvimConfig(t *testing.T) {
	cfg := buildCfg(false, false, "unknown", config.NvimFrameworkNone)
	plan := BuildPlan(cfg, platform.Platform{})

	// Unknown nvim config value treated as base: 43 modules
	if len(plan) != 43 {
		t.Errorf("expected 43 modules (unknown treated as base), got %d", len(plan))
	}

	// Should NOT include M-18 (NeovimPersonal)
	for _, mod := range plan {
		if mod.ID() == ModNeovimPersonal {
			t.Error("NvimPersonal should not be in plan when nvim config is 'unknown'")
		}
	}
}

func TestBuildPlan_UnknownFramework(t *testing.T) {
	cfg := buildCfg(false, false, config.NvimConfigBase, "unknown")
	plan := BuildPlan(cfg, platform.Platform{})

	// Unknown framework treated as none: 43 modules
	if len(plan) != 43 {
		t.Errorf("expected 43 modules (unknown treated as none), got %d", len(plan))
	}

	// Should NOT include M-19 (NvimFramework)
	for _, mod := range plan {
		if mod.ID() == ModNeovimFramework {
			t.Error("NvimFramework should not be in plan when framework is 'unknown'")
		}
	}
}
