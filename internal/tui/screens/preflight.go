package screens

import (
	"io/fs"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const wsl2GuidePath = "assets/wsl2-guide.md"

// Preflight renders the Windows-only WSL2 guide and exits before any install route exists.
type Preflight struct {
	guide string
}

func NewPreflight(assets fs.FS) Preflight {
	guide := "WSL2 setup guide is not available."
	if assets != nil {
		if content, err := fs.ReadFile(assets, wsl2GuidePath); err == nil {
			guide = string(content)
		}
	}
	return Preflight{guide: guide}
}

func (preflight Preflight) Init() tea.Cmd { return nil }

func (preflight Preflight) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := message.(tea.KeyMsg); ok {
		return preflight, tea.Quit
	}
	return preflight, nil
}

func (preflight Preflight) View() tea.View {
	return tea.NewView("WSL2 setup guide\n\n" + strings.TrimSpace(preflight.guide) + "\n\nPress any key to exit.\n")
}
