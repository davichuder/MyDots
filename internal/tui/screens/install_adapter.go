package screens

import (
	"context"
	"io"
	"io/fs"
	"time"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer"
	"github.com/davichuder/MyDots/internal/platform"
)

// InstallRequest contains the immutable inputs for one Phase-5 installation session.
type InstallRequest struct {
	Config    config.Config
	Platform  platform.Platform
	Assets    fs.FS
	Timestamp string
}

// InstallEventKind identifies the one event emitted to the install screen at a time.
type InstallEventKind string

const (
	InstallProgressEvent InstallEventKind = "progress"
	InstallLogEvent      InstallEventKind = "log"
	InstallDoneEvent     InstallEventKind = "done"
)

// InstallEvent is the typed bridge between the session runner and the TUI.
type InstallEvent struct {
	Kind      InstallEventKind
	Progress  installer.ProgressEvent
	LogLine   string
	Err       error
	Cancelled bool
}

// SessionRunner starts one cancellable installation session.
type SessionRunner interface {
	Start(context.Context, InstallRequest) <-chan InstallEvent
}

// BuildPlan creates the session's module plan. It is injectable to keep tests host-free.
type BuildPlan func(InstallRequest) []installer.Module

// ModuleSessionRunner runs Phase-4 modules locally so it can stop before later modules.
type ModuleSessionRunner struct{ buildPlan BuildPlan }

// NewModuleSessionRunner constructs the production adapter with an injected plan seam.
func NewModuleSessionRunner(buildPlan BuildPlan) ModuleSessionRunner {
	return ModuleSessionRunner{buildPlan: buildPlan}
}

// Start emits module progress and exactly one completion event.
func (runner ModuleSessionRunner) Start(ctx context.Context, request InstallRequest) <-chan InstallEvent {
	events := make(chan InstallEvent, 128)
	go func() {
		defer close(events)
		progress := make(chan installer.ProgressEvent, 128)
		installContext := installer.InstallContext{
			Platform:         request.Platform,
			Config:           request.Config,
			SessionTimestamp: request.Timestamp,
			Log:              installEventWriter{events: events},
			Cancel:           ctx,
			Assets:           request.Assets,
		}
		go installer.Run(runner.buildPlan(request), installContext, progress)
		for event := range progress {
			events <- InstallEvent{Kind: InstallProgressEvent, Progress: event}
		}
		events <- InstallEvent{Kind: InstallDoneEvent, Cancelled: ctx.Err() != nil, Err: ctx.Err()}
	}()
	return events
}

type installEventWriter struct{ events chan<- InstallEvent }

func (writer installEventWriter) Write(data []byte) (int, error) {
	writer.events <- InstallEvent{Kind: InstallLogEvent, LogLine: string(data)}
	return len(data), nil
}

var _ io.Writer = installEventWriter{}

// TerminalHandoff releases and restores Bubble Tea raw mode around sudo validation.
type TerminalHandoff interface {
	Release() error
	Restore() error
}

// Clock avoids real timer sleeps in installation tests.
type Clock interface{ Every(time.Duration) Ticker }

// Ticker is the small timer seam used by sudo keepalive.
type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

// SudoValidator runs interactive validation and quiet keepalive checks without exposing commands to screen logic.
type SudoValidator interface {
	Validate(context.Context) error
	ValidateQuiet(context.Context) error
}

// CommandStreams are the terminal streams used for interactive sudo validation.
type CommandStreams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Command captures the only command shape required by sudo validation.
type Command struct {
	Name string
	Args []string
	CommandStreams
}

// CommandExecutor is the narrow command seam used only for sudo validation.
type CommandExecutor interface {
	Execute(context.Context, Command) error
}

// CommandSudoValidator validates an existing sudo credential using injected terminal streams.
type CommandSudoValidator struct {
	commands CommandExecutor
	streams  CommandStreams
}

func NewCommandSudoValidator(commands CommandExecutor, streams CommandStreams) CommandSudoValidator {
	return CommandSudoValidator{commands: commands, streams: streams}
}

func (validator CommandSudoValidator) Validate(ctx context.Context) error {
	return validator.commands.Execute(ctx, Command{Name: "sudo", Args: []string{"-v"}, CommandStreams: validator.streams})
}

func (validator CommandSudoValidator) ValidateQuiet(ctx context.Context) error {
	return validator.commands.Execute(ctx, Command{Name: "sudo", Args: []string{"-v"}, CommandStreams: CommandStreams{Stdout: io.Discard, Stderr: io.Discard}})
}

// Elevation obtains an authenticated sudo session after terminal handoff.
type Elevation interface {
	Acquire(context.Context) (Keepalive, error)
}

// Keepalive stops and joins the 45-second sudo validation lifecycle.
type Keepalive interface{ Stop() }

// SudoElevation releases raw mode, validates sudo, restores raw mode, then keeps sudo alive.
type SudoElevation struct {
	terminal TerminalHandoff
	clock    Clock
	sudo     SudoValidator
}

func NewSudoElevation(terminal TerminalHandoff, clock Clock, sudo SudoValidator) SudoElevation {
	return SudoElevation{terminal: terminal, clock: clock, sudo: sudo}
}

func (elevation SudoElevation) Acquire(ctx context.Context) (Keepalive, error) {
	if err := elevation.terminal.Release(); err != nil {
		return nil, err
	}
	if err := elevation.sudo.Validate(ctx); err != nil {
		_ = elevation.terminal.Restore()
		return nil, err
	}
	if err := elevation.terminal.Restore(); err != nil {
		return nil, err
	}
	return newSudoKeepalive(ctx, elevation.clock, elevation.sudo), nil
}

type sudoKeepalive struct {
	cancel context.CancelFunc
	done   chan struct{}
	ticker Ticker
}

func newSudoKeepalive(parent context.Context, clock Clock, sudo SudoValidator) *sudoKeepalive {
	ctx, cancel := context.WithCancel(parent)
	keepalive := &sudoKeepalive{cancel: cancel, done: make(chan struct{}), ticker: clock.Every(45 * time.Second)}
	go func() {
		defer close(keepalive.done)
		defer keepalive.ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-keepalive.ticker.Chan():
				_ = sudo.ValidateQuiet(ctx)
			}
		}
	}()
	return keepalive
}

func (keepalive *sudoKeepalive) Stop() {
	keepalive.cancel()
	<-keepalive.done
}
