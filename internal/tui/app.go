// Package tui provides the interactive MyDots terminal application.
package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/tui/screens"
)

// App routes child screens without reading platform or asset state from the host.
type App struct {
	dependencies  Dependencies
	model         tea.Model
	windowSize    tea.WindowSizeMsg
	hasWindowSize bool
	startupError  error
}

// NewApp selects the safe preflight guide on Windows and the main menu elsewhere.
func NewApp(dependencies Dependencies) App {
	app := App{dependencies: dependencies}
	switch dependencies.Platform.OS {
	case platform.Windows:
		app.model = screens.NewPreflight(dependencies.Assets)
	case platform.Darwin, platform.Linux:
		app.model = screens.NewMainMenu(dependencies.ConfigStore)
	default:
		app.startupError = UnsupportedPlatformError{Platform: dependencies.Platform}
		app.model = startupErrorModel{err: app.startupError}
	}
	return app
}

func (app App) Init() tea.Cmd {
	return app.model.Init()
}

func (app App) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := message.(tea.WindowSizeMsg); ok {
		app.windowSize = size
		app.hasWindowSize = true
	}
	if route, ok := message.(screens.ChangeScreenMsg); ok {
		return app.route(route)
	}
	if issue, ok := message.(screens.ConfigIssue); ok {
		return app.route(screens.ChangeScreenMsg{Screen: screens.ScreenConfig, Payload: issue})
	}

	updated, command := app.model.Update(message)
	app.model = updated
	return app, command
}

func (app App) View() tea.View {
	return app.model.View()
}

func (app App) route(route screens.ChangeScreenMsg) (tea.Model, tea.Cmd) {
	switch route.Screen {
	case screens.ScreenMain:
		app.model = screens.NewMainMenu(app.dependencies.ConfigStore)
	case screens.ScreenConfig:
		if issue, ok := route.Payload.(screens.ConfigIssue); ok {
			app.model = screens.NewConfigMenu(app.dependencies.ConfigPath, issue)
		} else {
			app.model = screens.NewConfigMenu(app.dependencies.ConfigPath)
		}
	case screens.ScreenInstall:
		if configValue, ok := route.Payload.(config.Config); ok && app.dependencies.InstallFactory != nil {
			app.model = app.dependencies.InstallFactory(configValue)
		}
	case screens.ScreenResult:
		if result, ok := route.Payload.(screens.InstallResult); ok {
			app.model = screens.NewResultScreen(result, app.dependencies.Platform.Variant == platform.WSL2)
		}
	case screens.ScreenBackup:
		if app.dependencies.BackupFactory != nil {
			app.model = app.dependencies.BackupFactory()
		}
	case screens.ScreenReference:
		app.model = screens.NewReferenceMenu(app.dependencies.Assets)
	}
	init := app.model.Init()
	if !app.hasWindowSize {
		return app, init
	}
	updated, resize := app.model.Update(app.windowSize)
	app.model = updated
	return app, tea.Batch(init, resize)
}

type startupErrorModel struct{ err error }

// UnsupportedPlatformError reports an injected platform that has no safe TUI flow.
type UnsupportedPlatformError struct{ Platform platform.Platform }

func (err UnsupportedPlatformError) Error() string {
	return fmt.Sprintf("unsupported platform: %s", err.Platform.OS)
}

func (model startupErrorModel) Init() tea.Cmd { return nil }

func (model startupErrorModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return model, nil }

func (model startupErrorModel) View() tea.View {
	return tea.NewView("MyDots cannot start: " + model.err.Error() + "\n")
}
