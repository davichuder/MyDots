package modules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// CavemanModule (M-42)
// ---------------------------------------------------------------------------

// CavemanModule implements the caveman CLI installation module (M-42).
// It runs the caveman-install.sh curl-script wrapper with --only openclaw.
type CavemanModule struct{}

// Compile-time interface check.
var _ types.Module = CavemanModule{}

// Caveman is the exported package-level instance used by the catalogue.
var Caveman types.Module = CavemanModule{}

// ID returns the stable module identifier M-42.
func (m CavemanModule) ID() types.ModuleID { return types.ModCaveman }

// Name returns a human-readable module name.
func (m CavemanModule) Name() string { return "Caveman" }

// Criticality returns NonCritical — caveman failure does not stop the pipeline.
func (m CavemanModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns nil — caveman installs via curl-script, not brew.
func (m CavemanModule) Dependencies() []types.ModuleID { return nil }

// IsInstalled checks the durable artifacts created by the OpenClaw-only flow.
func (m CavemanModule) IsInstalled(_ platform.Platform) bool {
	workspace := os.Getenv("OPENCLAW_WORKSPACE")
	if workspace == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		workspace = filepath.Join(home, ".openclaw", "workspace")
	}

	skill, err := os.Stat(filepath.Join(workspace, "skills", "caveman", "SKILL.md"))
	if err != nil || !skill.Mode().IsRegular() || skill.Size() == 0 {
		return false
	}
	soul, err := os.ReadFile(filepath.Join(workspace, "SOUL.md"))
	if err != nil {
		return false
	}

	beginCount, endCount := 0, 0
	endAfterBegin := false
	for _, line := range strings.Split(string(soul), "\n") {
		switch line {
		case "<!-- caveman-begin -->":
			beginCount++
		case "<!-- caveman-end -->":
			endCount++
			endAfterBegin = endAfterBegin || beginCount > 0
		}
	}
	return beginCount == 1 && endCount == 1 && endAfterBegin
}

// Install runs the embedded caveman-install.sh script via runner.Script.
func (m CavemanModule) Install(ctx types.InstallContext) error {
	return runner.Script(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/caveman-install.sh", nil)
}

// AuditInfo identifies the durable OpenClaw integration.
func (m CavemanModule) AuditInfo() string {
	return "OpenClaw workspace skill"
}
