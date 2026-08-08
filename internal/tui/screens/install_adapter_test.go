package screens

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer"
	"github.com/davichuder/MyDots/internal/platform"
)

func TestModuleSessionRunnerCancellationPreventsLaterStartAndWrite(t *testing.T) {
	firstStarted := make(chan struct{})
	first := &fakeModule{id: installer.ModGit, name: "Git", started: firstStarted, waitForCancel: true}
	second := &fakeModule{id: installer.ModGo, name: "Go"}
	runner := NewModuleSessionRunner(func(InstallRequest) []installer.Module { return []installer.Module{first, second} })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := runner.Start(ctx, InstallRequest{Config: config.DefaultConfig(), Platform: platform.Platform{}})
	<-firstStarted
	cancel()
	for range events {
	}

	if first.starts != 1 {
		t.Errorf("first module starts = %d, want 1", first.starts)
	}
	if second.starts != 0 {
		t.Errorf("second module starts = %d, want 0 after cancellation", second.starts)
	}
	if second.writes != 0 {
		t.Errorf("second module writes = %d, want 0 after cancellation", second.writes)
	}
}

func TestModuleSessionRunnerPreservesInstallerRunSemantics(t *testing.T) {
	failing := &sessionModule{id: installer.ModGit, installErr: errors.New("failed")}
	dependent := &sessionModule{id: installer.ModGo, dependencies: []installer.ModuleID{installer.ModGit}}
	configured := &contextConfiguredModule{sessionModule: sessionModule{id: installer.ModNode}, installed: true}
	runner := NewModuleSessionRunner(func(InstallRequest) []installer.Module {
		return []installer.Module{failing, dependent, configured}
	})

	events := collectInstallEvents(runner.Start(context.Background(), InstallRequest{Config: config.DefaultConfig()}))
	if got := progressStatus(events, installer.ModGo); got != installer.StatusSkippedDependencyFailed {
		t.Errorf("dependent status = %q, want %q", got, installer.StatusSkippedDependencyFailed)
	}
	if configured.starts != 0 {
		t.Errorf("configured module starts = %d, want 0", configured.starts)
	}
	if !configured.probed {
		t.Error("configured module was not probed with its install context")
	}
	if failing.brewPath == nil {
		t.Error("install context did not provide the session BrewPath")
	}
}

func TestModuleSessionRunnerRendersModuleNameWhileRetainingModuleID(t *testing.T) {
	module := &fakeModule{id: installer.ModGit, name: "Git Setup"}
	runner := NewModuleSessionRunner(func(InstallRequest) []installer.Module {
		return []installer.Module{module}
	})

	var progress installer.ProgressEvent
	for event := range runner.Start(context.Background(), InstallRequest{Config: config.DefaultConfig()}) {
		if event.Kind == InstallProgressEvent && event.Progress.Status == installer.StatusInstalled {
			progress = event.Progress
		}
	}
	if progress.ModuleID != installer.ModGit {
		t.Fatalf("progress ModuleID = %q, want %q", progress.ModuleID, installer.ModGit)
	}
	if progress.ModuleName != "Git Setup" {
		t.Fatalf("progress ModuleName = %q, want Module.Name()", progress.ModuleName)
	}

	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{})
	updated, _ := screen.Update(installEventMsg{event: InstallEvent{Kind: InstallProgressEvent, Progress: progress}})
	view := installScreen(t, updated).View().Content
	if !strings.Contains(view, "Git Setup ✅") {
		t.Errorf("View() = %q, want rendered module name", view)
	}
	if strings.Contains(view, string(installer.ModGit)+" ✅") {
		t.Errorf("View() = %q, must not render module ID as the module name", view)
	}
}

func TestModuleSessionRunnerCriticalFailureRendersCriticalResult(t *testing.T) {
	criticalErr := errors.New("critical install failure")
	runner := NewModuleSessionRunner(func(InstallRequest) []installer.Module {
		return []installer.Module{&sessionModule{id: installer.ModHomebrew, criticality: installer.Critical, installErr: criticalErr}}
	})
	screen := NewInstallScreen(InstallRequest{Config: config.DefaultConfig()}, runner, fakeElevation{keepalive: &fakeKeepalive{}})

	updated, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, updated)
	updated, waitCommand := screen.Update(commandMessage(t, startCommand))
	screen = installScreen(t, updated)

	for {
		updated, nextCommand := screen.Update(commandMessage(t, waitCommand))
		screen = installScreen(t, updated)
		if nextCommand == nil {
			t.Fatal("progress update command = nil, want completion drain")
		}
		message := commandMessage(t, nextCommand)
		if result, ok := message.(ChangeScreenMsg); ok {
			outcome, ok := result.Payload.(InstallResult)
			if !ok {
				t.Fatalf("result payload = %T, want InstallResult", result.Payload)
			}
			if !errors.Is(outcome.Err, criticalErr) {
				t.Fatalf("result error = %v, want %v", outcome.Err, criticalErr)
			}
			if view := NewResultScreen(outcome, false).View().Content; !strings.Contains(view, "Installation failed") {
				t.Errorf("result view = %q, want critical failure", view)
			}
			return
		}
		waitCommand = nextCommand
	}
}

func TestSudoElevationReleasesRestoresAndJoinsKeepalive(t *testing.T) {
	terminal := &fakeTerminal{}
	clock := newFakeClock()
	sudo := &fakeSudo{called: make(chan struct{}, 2)}
	elevation := NewSudoElevation(terminal, clock, sudo)

	keepalive, err := elevation.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if terminal.releases != 1 || terminal.restores != 1 {
		t.Errorf("terminal release/restore = %d/%d, want 1/1", terminal.releases, terminal.restores)
	}
	if got := sudo.calls; got != 1 {
		t.Errorf("sudo calls after acquire = %d, want 1", got)
	}

	clock.tick()
	sudo.waitForCalls(t, 2)
	keepalive.Stop()
	if !clock.ticker.stopped {
		t.Error("ticker was not stopped by joined keepalive")
	}
	if got := sudo.calls; got != 2 {
		t.Errorf("sudo calls = %d, want acquire plus one 45-second keepalive", got)
	}
	if clock.every != 45*time.Second {
		t.Errorf("keepalive interval = %s, want 45s", clock.every)
	}
}

func TestSudoElevationRestoresTerminalWhenValidationFails(t *testing.T) {
	terminal := &fakeTerminal{}
	elevation := NewSudoElevation(terminal, newFakeClock(), &fakeSudo{err: errors.New("denied")})

	if _, err := elevation.Acquire(context.Background()); err == nil {
		t.Fatal("Acquire() error = nil, want sudo validation error")
	}
	if terminal.releases != 1 || terminal.restores != 1 {
		t.Errorf("terminal release/restore = %d/%d, want 1/1 on sudo failure", terminal.releases, terminal.restores)
	}
}

func TestSudoElevationAbortsWithPersistentDiagnosticWhenRestoreFails(t *testing.T) {
	restoreErr := errors.New("raw terminal restore failed")
	terminal := &fakeTerminal{restoreErr: restoreErr}
	diagnostics := &bytes.Buffer{}
	elevation := NewSudoElevationWithDiagnostics(terminal, newFakeClock(), &fakeSudo{}, diagnostics)

	if _, err := elevation.Acquire(context.Background()); err == nil {
		t.Fatal("Acquire() error = nil, want terminal recovery error")
	} else {
		var recovery TerminalRestoreError
		if !errors.As(err, &recovery) {
			t.Fatalf("Acquire() error = %T, want TerminalRestoreError", err)
		}
		if !errors.Is(err, restoreErr) {
			t.Errorf("Acquire() error = %v, want wrapped restore error", err)
		}
	}
	if got, want := diagnostics.String(), "mydots: terminal recovery failed after sudo validation: raw terminal restore failed\n"; got != want {
		t.Errorf("diagnostic = %q, want %q", got, want)
	}
}

func TestCommandSudoValidatorRunsSudoValidationOnly(t *testing.T) {
	commands := &fakeCommandExecutor{}
	streams := CommandStreams{Stdin: strings.NewReader("password\n"), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
	validator := NewCommandSudoValidator(commands, streams)

	if err := validator.Validate(context.Background()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if commands.command.Name != "sudo" || len(commands.command.Args) != 1 || commands.command.Args[0] != "-v" {
		t.Errorf("command = %#v, want sudo [-v]", commands.command)
	}
	if commands.command.Stdin != streams.Stdin || commands.command.Stdout != streams.Stdout || commands.command.Stderr != streams.Stderr {
		t.Errorf("interactive command streams = %#v, want inherited injected streams", commands.command)
	}

	if err := validator.ValidateQuiet(context.Background()); err != nil {
		t.Fatalf("ValidateQuiet() error = %v", err)
	}
	if commands.command.Stdin != nil || commands.command.Stdout != io.Discard || commands.command.Stderr != io.Discard {
		t.Errorf("keepalive command streams = %#v, want nil stdin and discarded output", commands.command)
	}
}

type fakeModule struct {
	id            installer.ModuleID
	name          string
	started       chan struct{}
	waitForCancel bool
	starts        int
	writes        int
}

type sessionModule struct {
	id           installer.ModuleID
	dependencies []installer.ModuleID
	criticality  installer.Criticality
	installErr   error
	starts       int
	brewPath     *string
}

func (module *sessionModule) ID() installer.ModuleID { return module.id }
func (module *sessionModule) Name() string           { return string(module.id) }
func (module *sessionModule) Criticality() installer.Criticality {
	if module.criticality != "" {
		return module.criticality
	}
	return installer.NonCritical
}
func (module *sessionModule) Dependencies() []installer.ModuleID { return module.dependencies }
func (module *sessionModule) IsInstalled(platform.Platform) bool { return false }
func (module *sessionModule) AuditInfo() string                  { return "" }
func (module *sessionModule) Install(ctx installer.InstallContext) error {
	module.starts++
	module.brewPath = ctx.BrewPath
	return module.installErr
}

type contextConfiguredModule struct {
	sessionModule
	installed bool
	probed    bool
}

func (module *contextConfiguredModule) IsInstalledForContext(ctx installer.InstallContext) bool {
	module.probed = true
	module.brewPath = ctx.BrewPath
	return module.installed
}

func collectInstallEvents(events <-chan InstallEvent) []InstallEvent {
	var collected []InstallEvent
	for event := range events {
		collected = append(collected, event)
	}
	return collected
}

func progressStatus(events []InstallEvent, moduleID installer.ModuleID) installer.InstallStatus {
	for _, event := range events {
		if event.Progress.ModuleID == moduleID {
			return event.Progress.Status
		}
	}
	return ""
}

func (module *fakeModule) ID() installer.ModuleID             { return module.id }
func (module *fakeModule) Name() string                       { return module.name }
func (module *fakeModule) Criticality() installer.Criticality { return installer.NonCritical }
func (module *fakeModule) Dependencies() []installer.ModuleID { return nil }
func (module *fakeModule) IsInstalled(platform.Platform) bool { return false }
func (module *fakeModule) AuditInfo() string                  { return "" }
func (module *fakeModule) Install(ctx installer.InstallContext) error {
	module.starts++
	if module.started != nil {
		close(module.started)
	}
	if module.waitForCancel {
		<-ctx.Cancel.Done()
		return ctx.Cancel.Err()
	}
	module.writes++
	return nil
}

type fakeTerminal struct {
	releases, restores int
	restoreErr         error
}

func (terminal *fakeTerminal) Release() error { terminal.releases++; return nil }
func (terminal *fakeTerminal) Restore() error { terminal.restores++; return terminal.restoreErr }

type fakeSudo struct {
	mu     sync.Mutex
	calls  int
	err    error
	called chan struct{}
}

type fakeCommandExecutor struct {
	command Command
}

func (executor *fakeCommandExecutor) Execute(_ context.Context, command Command) error {
	executor.command = command
	return nil
}

func (sudo *fakeSudo) Validate(context.Context) error {
	sudo.mu.Lock()
	defer sudo.mu.Unlock()
	sudo.calls++
	if sudo.called != nil {
		sudo.called <- struct{}{}
	}
	return sudo.err
}

func (sudo *fakeSudo) ValidateQuiet(ctx context.Context) error { return sudo.Validate(ctx) }

func (sudo *fakeSudo) waitForCalls(t *testing.T, want int) {
	t.Helper()
	for {
		sudo.mu.Lock()
		calls := sudo.calls
		sudo.mu.Unlock()
		if calls >= want {
			return
		}
		<-sudo.called
	}
}

type fakeClock struct {
	ticker *fakeTicker
	every  time.Duration
}

func newFakeClock() *fakeClock {
	return &fakeClock{ticker: &fakeTicker{ticks: make(chan time.Time, 1)}}
}
func (clock *fakeClock) Every(interval time.Duration) Ticker {
	clock.every = interval
	return clock.ticker
}
func (clock *fakeClock) tick() { clock.ticker.ticks <- time.Time{} }

type fakeTicker struct {
	ticks   chan time.Time
	stopped bool
}

func (ticker *fakeTicker) Chan() <-chan time.Time { return ticker.ticks }
func (ticker *fakeTicker) Stop()                  { ticker.stopped = true }
