package installer

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/modules"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/platform"
)

// testModule is a configurable Module for executor tests.
// It lets tests control IsInstalled, Install error, Criticality,
// dependencies, and AuditInfo without real implementations.
type testModule struct {
	id          ModuleID
	name        string
	criticality Criticality
	deps        []ModuleID
	installed   bool
	installErr  error
	auditInfo   string
}

func (m *testModule) ID() ModuleID                       { return m.id }
func (m *testModule) Name() string                       { return m.name }
func (m *testModule) Criticality() Criticality           { return m.criticality }
func (m *testModule) Dependencies() []ModuleID           { return m.deps }
func (m *testModule) IsInstalled(platform.Platform) bool { return m.installed }
func (m *testModule) Install(InstallContext) error       { return m.installErr }
func (m *testModule) AuditInfo() string                  { return m.auditInfo }

type configuredTestModule struct {
	testModule
	installedForConfig bool
}

// freshHomebrewHandoffModule models the post-install portion of HomebrewModule
// without running its embedded shell script. The module test covers the script
// boundary; this executor test covers the shared-session handoff it enables.
type freshHomebrewHandoffModule struct{}

func (freshHomebrewHandoffModule) ID() ModuleID                       { return ModHomebrew }
func (freshHomebrewHandoffModule) Name() string                       { return "Homebrew" }
func (freshHomebrewHandoffModule) Criticality() Criticality           { return Critical }
func (freshHomebrewHandoffModule) Dependencies() []ModuleID           { return nil }
func (freshHomebrewHandoffModule) IsInstalled(platform.Platform) bool { return false }
func (freshHomebrewHandoffModule) AuditInfo() string                  { return "" }
func (freshHomebrewHandoffModule) Install(ctx InstallContext) error {
	path, err := runner.RefreshBrew(ctx.Platform)
	if err != nil {
		return err
	}
	*ctx.BrewPath = path
	return nil
}

func (m *configuredTestModule) IsInstalledForConfig(platform.Platform, config.Config) bool {
	return m.installedForConfig
}

// collectEvents runs the executor on the given plan and returns all events
// received from the channel, in order.
func collectEvents(plan []Module, ctx InstallContext) []ProgressEvent {
	ch := make(chan ProgressEvent, 100)
	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for e := range ch {
		events = append(events, e)
	}
	return events
}

// defaultContext returns a minimal InstallContext suitable for executor tests.
func defaultContext() InstallContext {
	return InstallContext{
		Platform: platform.Platform{},
		Log:      &bytes.Buffer{},
		Cancel:   context.Background(),
	}
}

func TestRun_AllSucceed(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "A", criticality: Critical},
		&testModule{id: "M-02", name: "B", criticality: NonCritical},
		&testModule{id: "M-03", name: "C", criticality: NonCritical},
	}

	events := collectEvents(plan, defaultContext())

	// 3 running + 3 installed = 6 events
	if len(events) != 6 {
		t.Fatalf("expected 6 events, got %d", len(events))
	}

	// Order: A-running, A-installed, B-running, B-installed, C-running, C-installed
	assertEvent(t, events[0], "M-01", "running", false)
	assertEvent(t, events[1], "M-01", StatusInstalled, false)
	assertEvent(t, events[2], "M-02", "running", false)
	assertEvent(t, events[3], "M-02", StatusInstalled, false)
	assertEvent(t, events[4], "M-03", "running", false)
	assertEvent(t, events[5], "M-03", StatusInstalled, false)
}

func TestRun_NonCriticalFails(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "A", criticality: NonCritical},
		&testModule{id: "M-02", name: "B", criticality: NonCritical, installErr: errors.New("fail")},
		&testModule{id: "M-03", name: "C", criticality: NonCritical},
	}

	events := collectEvents(plan, defaultContext())

	// A-running, A-installed, B-running, B-failed, C-running, C-installed = 6
	if len(events) != 6 {
		t.Fatalf("expected 6 events, got %d", len(events))
	}

	assertEvent(t, events[0], "M-01", "running", false)
	assertEvent(t, events[1], "M-01", StatusInstalled, false)
	assertEvent(t, events[2], "M-02", "running", false)
	assertEvent(t, events[3], "M-02", StatusFailed, true) // has error
	assertEvent(t, events[4], "M-03", "running", false)
	assertEvent(t, events[5], "M-03", StatusInstalled, false)
}

func TestRun_CriticalFailsStopsPipeline(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "Homebrew", criticality: Critical, installErr: errors.New("brew failed")},
		&testModule{id: "M-02", name: "B", criticality: NonCritical},
	}

	events := collectEvents(plan, defaultContext())

	// Only M-01: running + failed = 2 events; M-02 never runs
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	assertEvent(t, events[0], "M-01", "running", false)
	assertEvent(t, events[1], "M-01", StatusFailed, true)
}

func TestRun_AlreadyInstalled(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "A", criticality: NonCritical, installed: true},
		&testModule{id: "M-02", name: "B", criticality: NonCritical},
	}

	events := collectEvents(plan, defaultContext())

	// M-01: skipped (no running event), M-02: running + installed = 3 events
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	assertEvent(t, events[0], "M-01", StatusSkipped, false)
	assertEvent(t, events[1], "M-02", "running", false)
	assertEvent(t, events[2], "M-02", StatusInstalled, false)
}

func TestRun_ConfiguredModuleReconcilesStaleState(t *testing.T) {
	module := &configuredTestModule{
		testModule:         testModule{id: "M-46", name: "Theme", criticality: NonCritical, installed: true},
		installedForConfig: false,
	}

	events := collectEvents([]Module{module}, defaultContext())
	if len(events) != 2 {
		t.Fatalf("expected stale configured state to install, got %d events", len(events))
	}
	assertEvent(t, events[0], "M-46", "running", false)
	assertEvent(t, events[1], "M-46", StatusInstalled, false)
}

func TestRun_FreshHomebrewHandoffUsesDiscoveredPathForChangedBrewModules(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "test-wayland")

	const discoveredBrew = "/test/homebrew/bin/brew"
	tests := []struct {
		name       string
		platform   platform.Platform
		module     Module
		wantBrew   [][]string
		failBrewAt int
	}{
		{
			name:     "BrewModule",
			platform: platform.Platform{OS: platform.Darwin},
			module:   BrewModule{id: ModZoxide, name: "zoxide", formula: "zoxide", checkCommand: "zoxide", deps: []ModuleID{ModHomebrew}},
			wantBrew: [][]string{{"install", "zoxide"}},
		},
		{
			name:     "C++",
			platform: platform.Platform{OS: platform.Darwin},
			module:   modules.CppToolchainModule{},
			wantBrew: [][]string{{"install", "gcc", "cmake", "llvm"}},
		},
		{
			name:     "Clipboard",
			platform: platform.Platform{OS: platform.Linux, Variant: platform.Native},
			module:   modules.ClipboardModule{},
			wantBrew: [][]string{{"install", "wl-clipboard"}},
		},
		{
			name:       "Neovim",
			platform:   platform.Platform{OS: platform.Darwin},
			module:     modules.NeovimModule{},
			wantBrew:   [][]string{{"install", "neovim"}},
			failBrewAt: 1,
		},
		{
			name:       "Zsh",
			platform:   platform.Platform{OS: platform.Darwin},
			module:     modules.ZshModule{},
			wantBrew:   [][]string{{"install", "zsh"}},
			failBrewAt: 1,
		},
		{
			name:       "NerdFont cask and configured-state probe",
			platform:   platform.Platform{OS: platform.Darwin},
			module:     modules.NerdFontModule{},
			wantBrew:   [][]string{{"list", "--cask"}, {"install", "--cask", "font-jetbrains-mono-nerd-font"}},
			failBrewAt: 2,
		},
		{
			name:       "Ghostty cask",
			platform:   platform.Platform{OS: platform.Darwin},
			module:     modules.GhosttyModule{},
			wantBrew:   [][]string{{"install", "--cask", "ghostty"}},
			failBrewAt: 1,
		},
		{
			name:       "Docker cask",
			platform:   platform.Platform{OS: platform.Darwin},
			module:     modules.DockerModule{},
			wantBrew:   [][]string{{"install", "--cask", "docker-desktop"}},
			failBrewAt: 1,
		},
		{
			name:     "GentleAI tap and install",
			platform: platform.Platform{OS: platform.Darwin},
			module:   modules.GentleAiModule{},
			wantBrew: [][]string{{"tap", "Gentleman-Programming/homebrew-tap"}, {"install", "gentle-ai"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []struct {
				name string
				args []string
			}
			setMockRunner(t, &mockExecutor{
				lookPathFunc: func(name string) (string, error) {
					if name == "brew" {
						return discoveredBrew, nil
					}
					return "", errors.New("not found")
				},
				executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
					if name != discoveredBrew {
						return nil, errors.New("unexpected command: " + name)
					}
					calls = append(calls, struct {
						name string
						args []string
					}{name: name, args: args})
					if tt.failBrewAt == len(calls) {
						return nil, errors.New("stop after brew command")
					}
					return nil, nil
				},
			})

			plan := []Module{
				freshHomebrewHandoffModule{},
				tt.module,
			}
			ctx := defaultContext()
			ctx.Platform = tt.platform
			ctx.Config = config.DefaultConfig()
			ch := make(chan ProgressEvent, 4)
			go Run(plan, ctx, ch)
			for range ch {
			}

			if len(calls) != len(tt.wantBrew) {
				t.Fatalf("brew calls = %#v, want %d calls", calls, len(tt.wantBrew))
			}
			for i, wantArgs := range tt.wantBrew {
				if calls[i].name != discoveredBrew || !slices.Equal(calls[i].args, wantArgs) {
					t.Errorf("brew call %d = %q %v, want %q %v", i, calls[i].name, calls[i].args, discoveredBrew, wantArgs)
				}
			}
		})
	}
}

func TestRun_DirectDependencyFailed(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "A", criticality: NonCritical, installErr: errors.New("fail")},
		&testModule{id: "M-02", name: "B", criticality: NonCritical, deps: []ModuleID{"M-01"}},
	}

	events := collectEvents(plan, defaultContext())

	// A-running, A-failed, B-skipped-dependency-failed = 3 events
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	assertEvent(t, events[0], "M-01", "running", false)
	assertEvent(t, events[1], "M-01", StatusFailed, true)
	assertEvent(t, events[2], "M-02", StatusSkippedDependencyFailed, false)
}

func TestRun_TransitiveDependencyFailed(t *testing.T) {
	plan := []Module{
		&testModule{id: "M-01", name: "A", criticality: NonCritical, installErr: errors.New("fail")},
		// B depends on A, C depends on B
		&testModule{id: "M-02", name: "B", criticality: NonCritical, deps: []ModuleID{"M-01"}},
		&testModule{id: "M-03", name: "C", criticality: NonCritical, deps: []ModuleID{"M-02"}},
	}

	events := collectEvents(plan, defaultContext())

	// A-running, A-failed, B-dependency-failed, C-dependency-failed = 4
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	assertEvent(t, events[0], "M-01", "running", false)
	assertEvent(t, events[1], "M-01", StatusFailed, true)
	assertEvent(t, events[2], "M-02", StatusSkippedDependencyFailed, false)
	assertEvent(t, events[3], "M-03", StatusSkippedDependencyFailed, false)
}

func TestRun_ChannelAlwaysClosed(t *testing.T) {
	// Test with critical failure, non-critical failure, and success —
	// channel must close in every case.
	tests := []struct {
		name string
		plan []Module
	}{
		{
			name: "all succeed",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: NonCritical},
				&testModule{id: "M-02", name: "B", criticality: NonCritical},
			},
		},
		{
			name: "critical fails",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: Critical, installErr: errors.New("fail")},
			},
		},
		{
			name: "dependency fails",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: NonCritical, installErr: errors.New("fail")},
				&testModule{id: "M-02", name: "B", criticality: NonCritical, deps: []ModuleID{"M-01"}},
			},
		},
		{
			name: "already installed",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: NonCritical, installed: true},
			},
		},
		{
			name: "single empty plan",
			plan: []Module{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan ProgressEvent, 100)
			go Run(tc.plan, defaultContext(), ch)

			done := make(chan struct{})
			go func() {
				for range ch {
				}
				close(done)
			}()

			select {
			case <-done:
				// Channel closed — correct behaviour
			case <-time.After(2 * time.Second):
				t.Fatal("channel was not closed within 2s — possible goroutine leak")
			}
		})
	}
}

func TestRun_Triangulation(t *testing.T) {
	tests := []struct {
		name      string
		plan      []Module
		wantCount int
	}{
		{
			name: "single module",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: NonCritical},
			},
			wantCount: 2, // running + installed
		},
		{
			name: "three modules in chain with dependencies",
			plan: []Module{
				&testModule{id: "M-01", name: "A", criticality: NonCritical},
				&testModule{id: "M-02", name: "B", criticality: NonCritical, deps: []ModuleID{"M-01"}},
				&testModule{id: "M-03", name: "C", criticality: NonCritical, deps: []ModuleID{"M-02"}},
			},
			wantCount: 6, // 3 running + 3 installed
		},
		{
			name: "five modules with branching deps",
			plan: []Module{
				&testModule{id: "M-01", name: "Root", criticality: NonCritical},
				&testModule{id: "M-02", name: "Left", criticality: NonCritical, deps: []ModuleID{"M-01"}},
				&testModule{id: "M-03", name: "Right", criticality: NonCritical, deps: []ModuleID{"M-01"}},
				&testModule{id: "M-04", name: "LeafL", criticality: NonCritical, deps: []ModuleID{"M-02"}},
				&testModule{id: "M-05", name: "LeafR", criticality: NonCritical, deps: []ModuleID{"M-03"}},
			},
			wantCount: 10, // 5 running + 5 installed
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := collectEvents(tc.plan, defaultContext())
			if len(events) != tc.wantCount {
				t.Fatalf("expected %d events, got %d", tc.wantCount, len(events))
			}
		})
	}
}

func TestRun_EmptyPlan(t *testing.T) {
	plan := []Module{}
	events := collectEvents(plan, defaultContext())
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty plan, got %d", len(events))
	}
}

// assertEvent checks that an event matches expected values.
// If wantErr is true, it only checks that Err is non-nil (not the exact error).
func assertEvent(t *testing.T, e ProgressEvent, wantID ModuleID, wantStatus InstallStatus, wantErr bool) {
	t.Helper()
	if e.ModuleID != wantID {
		t.Errorf("event ModuleID = %s, want %s", e.ModuleID, wantID)
	}
	if e.Status != wantStatus {
		t.Errorf("event Status = %s, want %s", e.Status, wantStatus)
	}
	if wantErr && e.Err == nil {
		t.Error("expected non-nil error on event, got nil")
	}
	if !wantErr && e.Err != nil {
		t.Errorf("expected nil error, got %v", e.Err)
	}
}
