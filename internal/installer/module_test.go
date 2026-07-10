package installer

import (
	"testing"
)

// allModuleIDs returns every ModuleID constant for verification tests.
func allModuleIDs() []ModuleID {
	return []ModuleID{
		ModHomebrew,                // M-01
		ModZsh,                     // M-02
		ModOhMyZsh,                 // M-03
		ModZellij,                  // M-04
		ModGit,                     // M-05
		ModGitCredOAuth,            // M-06
		ModLazygit,                 // M-07
		ModFnm,                     // M-08
		ModNode,                    // M-09
		ModUv,                      // M-10
		ModPython,                  // M-11
		ModGo,                      // M-12
		ModCppToolchain,            // M-13
		ModSdkman,                  // M-14
		ModJava,                    // M-15
		ModPhp,                     // M-16
		ModNeovim,                  // M-17
		ModNeovimPersonal,          // M-18
		ModNeovimFramework,         // M-19
		ModAtuin,                   // M-20
		ModZoxide,                  // M-21
		ModBat,                     // M-22
		ModEza,                     // M-23
		ModFd,                      // M-24
		ModRipgrep,                 // M-25
		ModFzf,                     // M-26
		ModSd,                      // M-27
		ModJq,                      // M-28
		ModYq,                      // M-29
		ModTldr,                    // M-30
		ModDelta,                   // M-31
		ModBottom,                  // M-32
		ModThefuck,                 // M-33
		ModCarapace,                // M-34
		ModGlow,                    // M-35
		ModGh,                      // M-36
		ModClipboard,               // M-37
		ModDocker,                  // M-38
		ModLazydocker,              // M-39
		ModOpencode,                // M-40
		ModRtk,                     // M-41
		ModCaveman,                 // M-42
		ModGentleAi,                // M-43
		ModMcpConfig,               // M-44
		ModChezmoi,                 // M-45
		ModTheme,                   // M-46
		ModNerdFont,                // M-47
		ModGhostty,                 // M-48
	}
}

func TestModuleIDConstants_NonEmpty(t *testing.T) {
	ids := allModuleIDs()
	for _, id := range ids {
		if string(id) == "" {
			t.Errorf("ModuleID constant is empty string")
		}
	}
}

func TestModuleIDConstants_NoDuplicates(t *testing.T) {
	ids := allModuleIDs()
	seen := make(map[ModuleID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ModuleID: %q", id)
		}
		seen[id] = true
	}
	if len(seen) != len(ids) {
		t.Errorf("expected %d unique ModuleIDs, got: %d", len(ids), len(seen))
	}
}

func TestModuleIDConstants_ExactCount(t *testing.T) {
	ids := allModuleIDs()
	if len(ids) != 48 {
		t.Errorf("expected exactly 48 ModuleID constants, got: %d", len(ids))
	}
}

func TestCriticalityConstants_Values(t *testing.T) {
	if Critical != "critical" {
		t.Errorf("Critical = %q, want %q", Critical, "critical")
	}
	if NonCritical != "non-critical" {
		t.Errorf("NonCritical = %q, want %q", NonCritical, "non-critical")
	}
}

func TestInstallStatusConstants_Values(t *testing.T) {
	tests := []struct {
		got  InstallStatus
		want string
	}{
		{StatusInstalled, "installed"},
		{StatusSkipped, "skipped"},
		{StatusSkippedDisabled, "skipped-disabled"},
		{StatusSkippedDependencyFailed, "skipped-dependency-failed"},
		{StatusSkippedNoWayland, "skipped-no-wayland"},
		{StatusFailed, "failed"},
	}
	for _, tt := range tests {
		if string(tt.got) != tt.want {
			t.Errorf("InstallStatus constant = %q, want %q", tt.got, tt.want)
		}
	}
}
