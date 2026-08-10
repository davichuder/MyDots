package screens

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer"
)

func TestInstallScreenConsumesOneEventThenRequeues(t *testing.T) {
	events := make(chan InstallEvent, 2)
	runner := &fakeSessionRunner{events: events}
	keepalive := &fakeKeepalive{}
	screen := NewInstallScreen(InstallRequest{Config: config.DefaultConfig()}, runner, fakeElevation{keepalive: keepalive})

	started, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, started)
	startMessage := commandMessage(t, startCommand)
	if _, ok := startMessage.(installStartedMsg); !ok {
		t.Fatalf("start command message = %T, want installStartedMsg", startMessage)
	}
	started, waitCommand := screen.Update(startMessage)
	screen = installScreen(t, started)

	events <- InstallEvent{Kind: InstallProgressEvent, Progress: installer.ProgressEvent{ModuleID: installer.ModGit, Status: "running", LogLine: "cloning"}}
	updated, requeue := screen.Update(commandMessage(t, waitCommand))
	screen = installScreen(t, updated)

	if got := screen.rows[installer.ModGit].Status; got != "running" {
		t.Errorf("Git row status = %q, want running", got)
	}
	if got := screen.logs.Lines; len(got) != 1 || got[0] != "cloning" {
		t.Errorf("log lines = %#v, want [cloning]", got)
	}
	if requeue == nil {
		t.Fatal("progress update command = nil, want one-event requeue")
	}

	events <- InstallEvent{Kind: InstallDoneEvent}
	completed, doneCommand := screen.Update(commandMessage(t, requeue))
	screen = installScreen(t, completed)
	if !screen.completed {
		t.Fatal("screen completed = false, want true")
	}
	result, ok := commandMessage(t, doneCommand).(ChangeScreenMsg)
	if !ok {
		t.Fatalf("completion command message = %T, want ChangeScreenMsg", commandMessage(t, doneCommand))
	}
	if result.Screen != ScreenResult {
		t.Errorf("completion screen = %q, want %q", result.Screen, ScreenResult)
	}
	if outcome, ok := result.Payload.(InstallResult); !ok || outcome.Cancelled || len(outcome.Rows) != 1 || outcome.Rows[0].Status != "running" {
		t.Errorf("completion payload = %#v, want running non-cancelled InstallResult", result.Payload)
	}
}

func TestInstallScreenClosedEventChannelTransitionsToResult(t *testing.T) {
	events := make(chan InstallEvent)
	close(events)
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: events}, fakeElevation{keepalive: &fakeKeepalive{}})

	started, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, started)
	started, waitCommand := screen.Update(commandMessage(t, startCommand))
	screen = installScreen(t, started)
	completed, resultCommand := screen.Update(commandMessage(t, waitCommand))
	if !installScreen(t, completed).completed {
		t.Fatal("closed event channel did not complete the screen")
	}
	result, ok := commandMessage(t, resultCommand).(ChangeScreenMsg)
	if !ok || result.Screen != ScreenResult {
		t.Errorf("closed channel result = %#v, want ChangeScreenMsg to result", result)
	}
}

func TestInstallScreenInitAndInitialSizeStartOneSession(t *testing.T) {
	runner := &fakeSessionRunner{events: make(chan InstallEvent)}
	elevation := &countingElevation{keepalive: &fakeKeepalive{}}
	screen := NewInstallScreen(InstallRequest{}, runner, elevation)

	if command := screen.Init(); command != nil {
		t.Fatal("Init() command is not nil, want startup to wait for the initial size message")
	}
	updated, command := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, updated)
	_, _ = screen.Update(commandMessage(t, command))

	if elevation.calls != 1 {
		t.Errorf("elevation acquisitions = %d, want 1", elevation.calls)
	}
	if runner.starts != 1 {
		t.Errorf("runner starts = %d, want 1", runner.starts)
	}
}

func TestInstallScreenQuestionTogglesCurrentModuleReference(t *testing.T) {
	events := make(chan InstallEvent, 1)
	events <- InstallEvent{Kind: InstallProgressEvent, Progress: installer.ProgressEvent{ModuleID: installer.ModGit, Status: "running"}}
	screen := NewInstallScreen(InstallRequest{Assets: fstest.MapFS{"assets/cheatsheets/git.md": {Data: []byte("Git reference\n")}}}, &fakeSessionRunner{events: events}, fakeElevation{keepalive: &fakeKeepalive{}})

	started, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, started)
	started, waitCommand := screen.Update(commandMessage(t, startCommand))
	screen = installScreen(t, started)
	updated, _ := screen.Update(commandMessage(t, waitCommand))
	screen = installScreen(t, updated)

	updated, _ = screen.Update(keyPress('?', "?"))
	screen = installScreen(t, updated)
	if !screen.showReference {
		t.Fatal("question key did not show the reference")
	}
	if view := screen.View().Content; !strings.Contains(view, "Git reference") {
		t.Errorf("View() = %q, want current module reference content", view)
	}

	updated, _ = screen.Update(keyPress('?', "?"))
	if installScreen(t, updated).showReference {
		t.Fatal("second question key did not hide the reference")
	}
}

func TestInstallScreenQuestionReportsUnavailableReference(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{})
	updated, _ := screen.Update(keyPress('?', "?"))
	if view := installScreen(t, updated).View().Content; !strings.Contains(view, "Reference not available") {
		t.Errorf("View() = %q, want explicit unavailable reference", view)
	}
}

func TestInstallScreenDirectInstalledAndFailedEventsRenderMatchingIndicators(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{})
	for _, event := range []installer.ProgressEvent{
		{ModuleID: installer.ModGit, Status: installer.StatusInstalled},
		{ModuleID: installer.ModGo, Status: installer.StatusFailed},
	} {
		updated, _ := screen.Update(installEventMsg{event: InstallEvent{Kind: InstallProgressEvent, Progress: event}})
		screen = installScreen(t, updated)
	}
	view := screen.View().Content
	if !strings.Contains(view, string(installer.ModGit)+" ✅") {
		t.Errorf("View() = %q, want installed indicator for Git", view)
	}
	if !strings.Contains(view, string(installer.ModGo)+" ❌") {
		t.Errorf("View() = %q, want failed indicator for Go", view)
	}
}

func TestInstallScreenCancellationStopsSessionAndDoesNotRequeue(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{keyPress(tea.KeyEscape, ""), keyPress('q', "q")} {
		t.Run(key.Key().String(), func(t *testing.T) {
			events := make(chan InstallEvent)
			runner := &fakeSessionRunner{events: events}
			keepalive := &fakeKeepalive{}
			screen := NewInstallScreen(InstallRequest{}, runner, fakeElevation{keepalive: keepalive})

			started, command := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			screen = installScreen(t, started)
			started, _ = screen.Update(commandMessage(t, command))
			screen = installScreen(t, started)

			updated, requeue := screen.Update(key)
			screen = installScreen(t, updated)
			if !screen.cancelled {
				t.Fatal("screen cancelled = false, want true")
			}
			if requeue != nil {
				t.Error("cancellation command is not nil, want no later event requeue")
			}
			if !runner.cancelled() {
				t.Error("runner context was not cancelled")
			}
			if keepalive.stops != 1 {
				t.Errorf("keepalive stops = %d, want 1", keepalive.stops)
			}
		})
	}
}

func TestInstallScreenDrainsPostCancellationEventsUntilCompletion(t *testing.T) {
	events := make(chan InstallEvent, 2)
	runner := &fakeSessionRunner{events: events}
	screen := NewInstallScreen(InstallRequest{}, runner, fakeElevation{keepalive: &fakeKeepalive{}})

	started, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, started)
	started, waitCommand := screen.Update(commandMessage(t, startCommand))
	screen = installScreen(t, started)

	cancelled, cancelCommand := screen.Update(keyPress('q', "q"))
	screen = installScreen(t, cancelled)
	if cancelCommand != nil {
		t.Fatal("cancellation command is not nil, want no new work command")
	}

	events <- InstallEvent{Kind: InstallProgressEvent, Progress: installer.ProgressEvent{ModuleID: installer.ModGit, Status: "running"}}
	drained, requeue := screen.Update(commandMessage(t, waitCommand))
	screen = installScreen(t, drained)
	if requeue == nil {
		t.Fatal("post-cancellation non-terminal event did not requeue completion drain")
	}
	if runner.starts != 1 {
		t.Errorf("runner starts = %d, want exactly 1", runner.starts)
	}

	events <- InstallEvent{Kind: InstallDoneEvent, Cancelled: true}
	completed, resultCommand := screen.Update(commandMessage(t, requeue))
	screen = installScreen(t, completed)
	if !screen.completed {
		t.Fatal("completion after cancellation did not complete the screen")
	}
	result, ok := commandMessage(t, resultCommand).(ChangeScreenMsg)
	if !ok || result.Screen != ScreenResult {
		t.Errorf("completion result = %#v, want ChangeScreenMsg to result", result)
	}
}

func TestInstallScreenCancellationCancelsInFlightElevation(t *testing.T) {
	runner := &fakeSessionRunner{events: make(chan InstallEvent)}
	elevation := &blockingElevation{acquired: make(chan context.Context, 1)}
	screen := NewInstallScreen(InstallRequest{}, runner, elevation)

	updated, command := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	screen = installScreen(t, updated)
	message := make(chan tea.Msg, 1)
	go func() { message <- command() }()
	ctx := <-elevation.acquired

	updated, _ = screen.Update(keyPress('q', "q"))
	screen = installScreen(t, updated)
	if !screen.cancelled || ctx.Err() == nil {
		t.Fatal("q did not cancel the elevation context")
	}
	started := <-message
	updated, _ = screen.Update(started)
	screen = installScreen(t, updated)
	if runner.starts != 0 {
		t.Errorf("runner starts = %d, want 0 after acquisition cancellation", runner.starts)
	}
}

func TestInstallScreenCancellationAfterSuccessfulElevationReturnsToMainAndStopsKeepalive(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{keyPress(tea.KeyEscape, ""), keyPress('q', "q")} {
		t.Run(key.Key().String(), func(t *testing.T) {
			runner := &fakeSessionRunner{events: make(chan InstallEvent)}
			keepalive := &fakeKeepalive{}
			elevation := &successAfterCancellationElevation{started: make(chan struct{}), release: make(chan struct{}), keepalive: keepalive}
			screen := NewInstallScreen(InstallRequest{}, runner, elevation)

			updated, command := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			screen = installScreen(t, updated)
			message := make(chan tea.Msg, 1)
			go func() { message <- command() }()
			<-elevation.started

			updated, _ = screen.Update(key)
			screen = installScreen(t, updated)
			close(elevation.release)
			started := <-message

			_, routeCommand := screen.Update(started)
			route, ok := commandMessage(t, routeCommand).(ChangeScreenMsg)
			if !ok {
				t.Fatalf("cancellation route = %T, want ChangeScreenMsg", commandMessage(t, routeCommand))
			}
			if route.Screen != ScreenMain {
				t.Errorf("cancellation route = %q, want %q", route.Screen, ScreenMain)
			}
			if keepalive.stops != 1 {
				t.Errorf("keepalive stops = %d, want 1", keepalive.stops)
			}
			if runner.starts != 0 {
				t.Errorf("runner starts = %d, want 0 after acquisition cancellation", runner.starts)
			}
		})
	}
}

func TestInstallScreenCancelledElevationReturnsDirectlyToMainMenu(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{err: context.Canceled})

	updated, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, routeCommand := installScreen(t, updated).Update(commandMessage(t, startCommand))
	route, ok := commandMessage(t, routeCommand).(ChangeScreenMsg)
	if !ok {
		t.Fatalf("cancellation route = %T, want ChangeScreenMsg", commandMessage(t, routeCommand))
	}
	if route.Screen != ScreenMain {
		t.Errorf("cancellation route = %q, want %q", route.Screen, ScreenMain)
	}
}

func TestInstallScreenElevationFailureStillRoutesToResult(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{err: context.DeadlineExceeded})

	updated, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, routeCommand := installScreen(t, updated).Update(commandMessage(t, startCommand))
	route, ok := commandMessage(t, routeCommand).(ChangeScreenMsg)
	if !ok {
		t.Fatalf("failure route = %T, want ChangeScreenMsg", commandMessage(t, routeCommand))
	}
	if route.Screen != ScreenResult {
		t.Errorf("failure route = %q, want %q", route.Screen, ScreenResult)
	}
}

func TestInstallScreenTerminalRestoreFailureQuitsWithoutRenderingAResult(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{err: TerminalRestoreError{Err: context.DeadlineExceeded}})

	updated, startCommand := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, quitCommand := installScreen(t, updated).Update(commandMessage(t, startCommand))
	if _, ok := commandMessage(t, quitCommand).(tea.QuitMsg); !ok {
		t.Errorf("terminal restore failure command = %T, want tea.QuitMsg", commandMessage(t, quitCommand))
	}
}

func TestInstallScreenCarriesHumanModuleNameIntoResultRows(t *testing.T) {
	screen := NewInstallScreen(InstallRequest{}, &fakeSessionRunner{events: make(chan InstallEvent)}, fakeElevation{})
	updated, _ := screen.Update(installEventMsg{event: InstallEvent{Kind: InstallProgressEvent, Progress: installer.ProgressEvent{ModuleID: installer.ModGit, ModuleName: "Git Setup", Status: installer.StatusFailed}}})
	result := commandMessage(t, installScreen(t, updated).changeToResult(nil)).(ChangeScreenMsg).Payload.(InstallResult)

	if got, want := result.Rows[0].ModuleName, "Git Setup"; got != want {
		t.Errorf("result module name = %q, want %q", got, want)
	}
	if got, want := result.Rows[0].ModuleID, installer.ModGit; got != want {
		t.Errorf("result module ID = %q, want %q", got, want)
	}
}

type fakeSessionRunner struct {
	events chan InstallEvent
	ctx    context.Context
	starts int
}

func (runner *fakeSessionRunner) Start(ctx context.Context, _ InstallRequest) <-chan InstallEvent {
	runner.ctx = ctx
	runner.starts++
	return runner.events
}

func (runner *fakeSessionRunner) cancelled() bool {
	return runner.ctx != nil && runner.ctx.Err() != nil
}

type fakeElevation struct {
	keepalive Keepalive
	err       error
}

type countingElevation struct {
	keepalive Keepalive
	calls     int
}

func (elevation *countingElevation) Acquire(context.Context) (Keepalive, error) {
	elevation.calls++
	return elevation.keepalive, nil
}

type blockingElevation struct{ acquired chan context.Context }

func (elevation *blockingElevation) Acquire(ctx context.Context) (Keepalive, error) {
	elevation.acquired <- ctx
	<-ctx.Done()
	return nil, ctx.Err()
}

type successAfterCancellationElevation struct {
	started   chan struct{}
	release   chan struct{}
	keepalive Keepalive
}

func (elevation *successAfterCancellationElevation) Acquire(context.Context) (Keepalive, error) {
	close(elevation.started)
	<-elevation.release
	return elevation.keepalive, nil
}

func (elevation fakeElevation) Acquire(context.Context) (Keepalive, error) {
	return elevation.keepalive, elevation.err
}

type fakeKeepalive struct{ stops int }

func (keepalive *fakeKeepalive) Stop() { keepalive.stops++ }

func installScreen(t *testing.T, model tea.Model) InstallScreen {
	t.Helper()
	screen, ok := model.(InstallScreen)
	if !ok {
		t.Fatalf("Update() model = %T, want InstallScreen", model)
	}
	return screen
}
