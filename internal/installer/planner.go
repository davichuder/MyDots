package installer

import (
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

// isDisabled returns true when a module should be excluded from the
// install plan based on the user's configuration.
func isDisabled(mod Module, cfg config.Config) bool {
	switch mod.ID() {
	case ModSdkman, ModJava:
		return !cfg.Languages.Java
	case ModPhp:
		return !cfg.Languages.PHP
	case ModNeovimPersonal:
		return cfg.Nvim.Config != config.NvimConfigPersonal
	case ModNeovimFramework:
		// Disabled for none and any unrecognized value.
		switch cfg.Nvim.Framework {
		case config.NvimFrameworkLazyVim, config.NvimFrameworkLunarVim, config.NvimFrameworkAstroNvim, config.NvimFrameworkNvChad, config.NvimFrameworkLaunchVim:
			return false
		default:
			return true
		}
	}
	return false
}

// BuildPlan returns the ordered list of modules to install based on the
// user's configuration. Modules are returned in the same order as
// allModules(), with disabled modules removed.
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
