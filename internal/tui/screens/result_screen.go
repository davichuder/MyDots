package screens

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/installer"
)

// ResultScreen renders the terminal outcome of an installation without host-dependent state.
type ResultScreen struct {
	result InstallResult
	wsl2   bool
	width  int
	height int
}

// NewResultScreen creates a result view from the install screen's typed outcome.
func NewResultScreen(result InstallResult, wsl2 bool) ResultScreen {
	return ResultScreen{result: result, wsl2: wsl2}
}

func (screen ResultScreen) Init() tea.Cmd { return nil }

func (screen ResultScreen) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		size := msg
		screen.width = size.Width
		screen.height = size.Height
	case tea.KeyMsg:
		return screen, changeScreen(ScreenMain, nil)
	}
	return screen, nil
}

func (screen ResultScreen) View() tea.View {
	var content strings.Builder
	content.WriteString(screen.outcome())
	if screen.wsl2 {
		content.WriteString("\nWSL2 font note: install the selected Nerd Font in Windows Terminal for best results.\n")
	}
	content.WriteString("\nPress any key to return to the menu.\n")
	return tea.NewView(content.String())
}

func (screen ResultScreen) outcome() string {
	if screen.result.Cancelled {
		return fmt.Sprintf("Installation cancelled\n%s completed\n", moduleCount(screen.completedCount()))
	}
	if screen.result.Err != nil {
		return fmt.Sprintf("Installation failed\n%s\n%s\n", screen.failedModuleNames(), screen.result.Err)
	}
	if failures := screen.failedModuleNames(); failures != "" {
		return fmt.Sprintf("Installation completed with warnings\n%s\nFailed modules: %s\n", warningCount(screen.failedCount()), failures)
	}
	return fmt.Sprintf("Installation complete\n%s completed\n", moduleCount(screen.completedCount()))
}

func (screen ResultScreen) completedCount() int {
	completed := 0
	for _, row := range screen.result.Rows {
		if row.Status == installer.StatusInstalled || strings.HasPrefix(string(row.Status), "skipped") {
			completed++
		}
	}
	return completed
}

func (screen ResultScreen) failedCount() int {
	failed := 0
	for _, row := range screen.result.Rows {
		if row.Status == installer.StatusFailed {
			failed++
		}
	}
	return failed
}

func (screen ResultScreen) failedModuleNames() string {
	failed := make([]string, 0, screen.failedCount())
	for _, row := range screen.result.Rows {
		if row.Status == installer.StatusFailed {
			name := row.ModuleName
			if name == "" {
				name = string(row.ModuleID)
			}
			failed = append(failed, name)
		}
	}
	return strings.Join(failed, ", ")
}

func moduleCount(count int) string {
	if count == 1 {
		return "1 module"
	}
	return fmt.Sprintf("%d modules", count)
}

func warningCount(count int) string {
	if count == 1 {
		return "1 warning"
	}
	return fmt.Sprintf("%d warnings", count)
}
