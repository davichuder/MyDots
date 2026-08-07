package screens

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
)

const (
	mainMenuContentWidth  = 28
	mainMenuContentHeight = 9
)

var mainMenuChoices = []string{"Config", "Install", "Backup", "Reference", "Quit"}

// Screen identifies the next Phase-5 TUI screen selected by a child model.
type Screen string

const (
	ScreenConfig    Screen = "config"
	ScreenMain      Screen = "main"
	ScreenInstall   Screen = "install"
	ScreenBackup    Screen = "backup"
	ScreenReference Screen = "reference"
)

// ChangeScreenMsg lets the root app route a screen-owned navigation outcome.
type ChangeScreenMsg struct {
	Screen  Screen
	Payload any
}

// ConfigIssue preserves the saved-configuration error for the configuration assistant.
type ConfigIssue struct {
	Err error
}

// ConfigStore loads the saved configuration without coupling menu behavior to host paths.
type ConfigStore interface {
	Load() (config.Config, error)
}

// MainMenu is the deterministic, no-color Phase-5 menu model.
type MainMenu struct {
	configStore ConfigStore
	cursor      int
	inlineError string
	width       int
	height      int
}

func NewMainMenu(configStore ConfigStore) MainMenu {
	return MainMenu{configStore: configStore}
}

func (menu MainMenu) Init() tea.Cmd {
	return nil
}

func (menu MainMenu) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		menu.width = msg.Width
		menu.height = msg.Height
	case tea.KeyMsg:
		switch msg.Key().Code {
		case tea.KeyUp:
			if menu.cursor > 0 {
				menu.cursor--
			}
		case tea.KeyDown:
			if menu.cursor < len(mainMenuChoices)-1 {
				menu.cursor++
			}
		case tea.KeyEnter:
			return menu.selectChoice()
		}
	}

	return menu, nil
}

func (menu MainMenu) View() tea.View {
	margin := strings.Repeat(" ", max((menu.width-mainMenuContentWidth)/2, 0))
	topPadding := strings.Repeat("\n", max((menu.height-mainMenuContentHeight)/2, 0))

	var content strings.Builder
	content.WriteString(topPadding)
	content.WriteString(margin)
	content.WriteString("MyDots\n\n")
	for index, choice := range mainMenuChoices {
		prefix := "  "
		if index == menu.cursor {
			prefix = "> "
		}
		content.WriteString(margin)
		content.WriteString(prefix)
		content.WriteString(choice)
		content.WriteByte('\n')
	}
	if menu.inlineError != "" {
		content.WriteString("\n")
		content.WriteString(margin)
		content.WriteString("Error: ")
		content.WriteString(menu.inlineError)
		content.WriteByte('\n')
	}
	content.WriteString("\n")
	content.WriteString(margin)
	content.WriteString("↑/↓ navigate • enter select\n")

	return tea.NewView(content.String())
}

func (menu MainMenu) selectChoice() (tea.Model, tea.Cmd) {
	switch menu.cursor {
	case 0:
		return menu, changeScreen(ScreenConfig, nil)
	case 1:
		return menu.selectInstall()
	case 2:
		return menu, changeScreen(ScreenBackup, nil)
	case 3:
		return menu, changeScreen(ScreenReference, nil)
	case 4:
		return menu, tea.Quit
	default:
		return menu, nil
	}
}

func (menu MainMenu) selectInstall() (tea.Model, tea.Cmd) {
	loaded, err := menu.configStore.Load()
	if errors.Is(err, os.ErrNotExist) {
		menu.inlineError = "No saved configuration. Select Config first."
		return menu, nil
	}
	if err != nil {
		return menu, changeScreen(ScreenConfig, ConfigIssue{Err: err})
	}
	if err := config.Validate(loaded); err != nil {
		return menu, changeScreen(ScreenConfig, ConfigIssue{Err: err})
	}
	return menu, changeScreen(ScreenInstall, loaded)
}

func changeScreen(screen Screen, payload any) tea.Cmd {
	return func() tea.Msg {
		return ChangeScreenMsg{Screen: screen, Payload: payload}
	}
}

func (issue ConfigIssue) Error() string {
	return fmt.Sprintf("configuration issue: %v", issue.Err)
}
