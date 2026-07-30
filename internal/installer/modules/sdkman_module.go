package modules

import (
	"os"
	"path/filepath"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

// ---------------------------------------------------------------------------
// Injectable dependencies for testing
// ---------------------------------------------------------------------------

// sdkmanCheckInitFile checks whether ~/.sdkman/bin/sdkman-init.sh exists.
// Injectable for testing.
var sdkmanCheckInitFile = func() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	initFile := filepath.Join(home, ".sdkman", "bin", "sdkman-init.sh")
	info, err := os.Stat(initFile)
	return err == nil && !info.IsDir()
}

// ---------------------------------------------------------------------------
// SdkmanModule (M-14)
// ---------------------------------------------------------------------------

// SdkmanModule implements the SDKMAN! installation module (M-14).
// It delegates to a runtime.RuntimeManager for the install and has its own
// IsInstalled check for ~/.sdkman/bin/sdkman-init.sh.
type SdkmanModule struct {
	manager runtime.RuntimeManager
}

// Compile-time interface check.
var _ types.Module = SdkmanModule{}

// Sdkman is the exported package-level instance used by the catalogue.
var Sdkman types.Module = NewSdkmanModule(&runtime.SdkmanManager{})

// NewSdkmanModule creates a new SdkmanModule with the given manager.
// Used by tests to inject a mock RuntimeManager.
func NewSdkmanModule(manager runtime.RuntimeManager) SdkmanModule {
	return SdkmanModule{manager: manager}
}

// ID returns the stable module identifier M-14.
func (m SdkmanModule) ID() types.ModuleID { return types.ModSdkman }

// Name returns a human-readable module name.
func (m SdkmanModule) Name() string { return "Sdkman" }

// Criticality returns NonCritical — sdkman failure does not stop the pipeline.
func (m SdkmanModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a requirement.
func (m SdkmanModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether ~/.sdkman/bin/sdkman-init.sh exists.
func (m SdkmanModule) IsInstalled(_ platform.Platform) bool {
	return sdkmanCheckInitFile()
}

// Install delegates to the runtime manager's Install (sdkman-install.sh script).
func (m SdkmanModule) Install(ctx types.InstallContext) error {
	return m.manager.Install(ctx.Cancel, ctx.Log, ctx.Assets)
}

// AuditInfo delegates to the runtime manager's version reporting.
func (m SdkmanModule) AuditInfo() string {
	return m.manager.AuditInfo()
}
