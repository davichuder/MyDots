// Package screens provides Bubble Tea models for interactive TUI workflows.
package screens

import (
	"context"
	"errors"
	"io/fs"
	"sort"
	"strings"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/installer"
	"github.com/davichuder/MyDots/internal/tui/components"
)

// InstallScreen owns one cancellable runner session and its deterministic feedback state.
type InstallScreen struct {
	request       InstallRequest
	runner        SessionRunner
	elevation     Elevation
	ctx           context.Context
	cancel        context.CancelFunc
	events        <-chan InstallEvent
	keepalive     Keepalive
	rows          map[installer.ModuleID]components.ProgressRow
	logs          components.LogPane
	showReference bool
	currentModule installer.ModuleID
	started       bool
	completed     bool
	cancelled     bool
}

// NewInstallScreen creates an idle installation screen. The initial size message starts it once.
func NewInstallScreen(request InstallRequest, runner SessionRunner, elevation Elevation) InstallScreen {
	return InstallScreen{request: request, runner: runner, elevation: elevation, rows: make(map[installer.ModuleID]components.ProgressRow), logs: components.NewLogPane()}
}

func (screen InstallScreen) Init() tea.Cmd { return nil }

func (screen InstallScreen) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		if !screen.started && !screen.cancelled {
			screen.started = true
			screen.ctx, screen.cancel = context.WithCancel(context.Background())
			return screen, screen.startInstall()
		}
	case tea.KeyMsg:
		if msg.Key().Text == "?" {
			screen.showReference = !screen.showReference
			return screen, nil
		}
		if msg.Key().Code == tea.KeyEscape || msg.Key().Text == "q" {
			screen.stopSession()
			return screen, nil
		}
	case installStartedMsg:
		if msg.err != nil {
			screen.completed = true
			screen.finishSession()
			if errors.Is(msg.err, context.Canceled) {
				return screen, changeScreen(ScreenMain, nil)
			}
			var terminalRestoreErr TerminalRestoreError
			if errors.As(msg.err, &terminalRestoreErr) {
				return screen, tea.Quit
			}
			return screen, screen.changeToResult(msg.err)
		}
		if screen.cancelled {
			if msg.keepalive != nil {
				msg.keepalive.Stop()
			}
			return screen, changeScreen(ScreenMain, nil)
		}
		screen.keepalive = msg.keepalive
		screen.events = screen.runner.Start(screen.ctx, screen.request)
		return screen, screen.waitForProgress()
	case installEventMsg:
		screen.applyEvent(msg.event)
		if msg.event.Kind == InstallDoneEvent {
			screen.completed = true
			screen.finishSession()
			return screen, screen.changeToResult(msg.event.Err)
		}
		return screen, screen.waitForProgress()
	}
	return screen, nil
}

func (screen InstallScreen) View() tea.View {
	var content strings.Builder
	content.WriteString("Installation\n\n")
	ids := make([]string, 0, len(screen.rows))
	for id := range screen.rows {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	for _, id := range ids {
		content.WriteString(screen.rows[installer.ModuleID(id)].View())
	}
	for _, line := range screen.logs.Lines {
		content.WriteString(line)
	}
	if screen.showReference {
		content.WriteString("\n")
		content.WriteString(screen.currentReference())
		content.WriteString("\n")
	}
	if screen.cancelled {
		content.WriteString("\nInstallation cancelled\n")
	}
	content.WriteString("\n? reference • q/esc cancel\n")
	return tea.NewView(content.String())
}

type installStartedMsg struct {
	keepalive Keepalive
	err       error
}
type installEventMsg struct{ event InstallEvent }

// InstallResult preserves install feedback for the later result screen without coupling it to this model.
type InstallResult struct {
	Rows      []InstallResultRow
	Cancelled bool
	Err       error
}

// InstallResultRow is one final module state supplied to the result screen.
type InstallResultRow struct {
	ModuleID   installer.ModuleID
	ModuleName string
	Status     installer.InstallStatus
}

func (screen InstallScreen) startInstall() tea.Cmd {
	return func() tea.Msg {
		keepalive, err := screen.elevation.Acquire(screen.ctx)
		return installStartedMsg{keepalive: keepalive, err: err}
	}
}

func (screen InstallScreen) waitForProgress() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-screen.events
		if !ok {
			return installEventMsg{event: InstallEvent{Kind: InstallDoneEvent, Cancelled: screen.cancelled}}
		}
		return installEventMsg{event: event}
	}
}

func (screen *InstallScreen) applyEvent(event InstallEvent) {
	if event.LogLine != "" {
		screen.logs.Append(event.LogLine)
	}
	if event.Progress.ModuleID != "" {
		name := event.Progress.ModuleName
		if name == "" {
			name = string(event.Progress.ModuleID)
		}
		row := components.NewProgressRow(name, string(event.Progress.Status))
		screen.rows[event.Progress.ModuleID] = row
		if event.Progress.Status == "running" {
			screen.currentModule = event.Progress.ModuleID
		}
	}
	if event.Progress.LogLine != "" {
		screen.logs.Append(event.Progress.LogLine)
	}
	if event.Cancelled {
		screen.cancelled = true
	}
}

func (screen InstallScreen) currentReference() string {
	if screen.currentModule == "" || screen.request.Assets == nil {
		return "Reference not available"
	}
	slug := referenceSlug(screen.currentModule)
	if slug == "" {
		return "Reference not available"
	}
	content, err := fs.ReadFile(screen.request.Assets, "assets/cheatsheets/"+slug+".md")
	if err != nil {
		return "Reference not available"
	}
	return string(content)
}

func (screen InstallScreen) changeToResult(err error) tea.Cmd {
	rows := make([]InstallResultRow, 0, len(screen.rows))
	for moduleID, row := range screen.rows {
		rows = append(rows, InstallResultRow{ModuleID: moduleID, ModuleName: row.Name, Status: installer.InstallStatus(row.Status)})
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].ModuleID < rows[right].ModuleID })
	return changeScreen(ScreenResult, InstallResult{Rows: rows, Cancelled: screen.cancelled, Err: err})
}

func (screen *InstallScreen) stopSession() {
	if screen.cancelled && screen.cancel == nil {
		return
	}
	screen.cancelled = true
	if screen.cancel != nil {
		screen.cancel()
	}
	if screen.keepalive != nil {
		screen.keepalive.Stop()
		screen.keepalive = nil
	}
}

func (screen *InstallScreen) finishSession() {
	if screen.cancel != nil {
		screen.cancel()
	}
	if screen.keepalive != nil {
		screen.keepalive.Stop()
		screen.keepalive = nil
	}
}
