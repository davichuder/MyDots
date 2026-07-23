package installer

import (
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

// BuildPlan filters the full module catalogue based on the user's config.
// Disabled modules are excluded from the plan. The remaining modules keep
// their canonical execution order from allModules().
//
// The Platform parameter is reserved for future platform-specific disabling
// (e.g. WSL2-only modules). Currently unused.
func BuildPlan(cfg config.Config, _ platform.Platform) []Module {
	all := allModules()
	plan := make([]Module, 0, len(all))
	for _, mod := range all {
		if isDisabled(mod, cfg) {
			continue
		}
		plan = append(plan, mod)
	}
	return plan
}

// isDisabled reports whether a module should be excluded from the install
// plan based on the current configuration. Unknown or invalid config values
// are treated conservatively (disabled) to avoid installing unwanted tools.
func isDisabled(mod Module, cfg config.Config) bool {
	switch mod.ID() {
	case ModSdkman, ModJava:
		return !cfg.Languages.Java
	case ModPhp:
		return !cfg.Languages.PHP
	case ModNeovimPersonal:
		return cfg.Nvim.Config != config.NvimConfigPersonal
	case ModNeovimFramework:
		// Disabled if no framework is selected or the value is unknown.
		switch cfg.Nvim.Framework {
		case config.NvimFrameworkLazyVim, config.NvimFrameworkLunarVim,
			config.NvimFrameworkAstroNvim, config.NvimFrameworkNvChad,
			config.NvimFrameworkLaunchVim:
			return false
		default:
			return true // none, empty, or unknown → disabled
		}
	}
	return false
}
