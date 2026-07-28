package modules

import (
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/runtime"
)

// ---------------------------------------------------------------------------
// JavaModule (M-15)
// ---------------------------------------------------------------------------

// JavaModule implements Java 25 installation via sdkman (M-15).
// It delegates to a runtime.RuntimeManager for InstallRuntime and checks
// `sdk list java` output for IsInstalled.
type JavaModule struct {
	manager runtime.RuntimeManager
}

// Compile-time interface check.
var _ types.Module = JavaModule{}

// Java is the exported package-level instance used by the catalogue.
var Java types.Module = NewJavaModule(&runtime.SdkmanManager{})

// NewJavaModule creates a new JavaModule with the given manager.
// Used by tests to inject a mock RuntimeManager.
func NewJavaModule(manager runtime.RuntimeManager) JavaModule {
	return JavaModule{manager: manager}
}

// ID returns the stable module identifier M-15.
func (m JavaModule) ID() types.ModuleID { return types.ModJava }

// Name returns a human-readable module name.
func (m JavaModule) Name() string { return "Java" }

// Criticality returns NonCritical — Java failure does not stop the pipeline.
func (m JavaModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Sdkman as a requirement (Java is installed via sdkman).
func (m JavaModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModSdkman}
}

// IsInstalled checks whether `sdk list java` output contains "25.*installed"
// or "25.*local", indicating Java 25 is available via sdkman.
func (m JavaModule) IsInstalled(_ platform.Platform) bool {
	out := sdkmanOutput("sdk list java")
	for _, row := range strings.Split(out, "\n") {
		if strings.Contains(row, "25.") && (strings.Contains(row, "installed") || strings.Contains(row, "local")) {
			return true
		}
	}
	return false
}

func sdkmanOutput(command string) string {
	return runner.CaptureOutput("sh", "-c", `. "$HOME/.sdkman/bin/sdkman-init.sh"`+"\n"+command)
}

// Install installs Java 25 (OpenJDK) via sdkman using InstallRuntime("25-open").
func (m JavaModule) Install(ctx types.InstallContext) error {
	return m.manager.InstallRuntime(ctx.Cancel, "25-open", ctx.Log)
}

// AuditInfo delegates to the runtime manager's version reporting.
func (m JavaModule) AuditInfo() string {
	return m.manager.AuditInfo()
}
