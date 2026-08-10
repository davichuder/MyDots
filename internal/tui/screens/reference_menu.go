package screens

import (
	"io/fs"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ReferenceMenu displays embedded reference material when it is available.
type ReferenceMenu struct {
	assets   fs.FS
	tools    []string
	cursor   int
	selected string
}

func NewReferenceMenu(assets fs.FS) ReferenceMenu {
	menu := ReferenceMenu{assets: assets}
	if assets == nil {
		return menu
	}
	entries, err := fs.ReadDir(assets, "assets/cheatsheets")
	if err != nil {
		return menu
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			menu.tools = append(menu.tools, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	sort.Strings(menu.tools)
	return menu
}

func (menu ReferenceMenu) Init() tea.Cmd { return nil }

func (menu ReferenceMenu) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return menu, nil
	}
	if menu.selected != "" || len(menu.tools) == 0 {
		return menu, changeScreen(ScreenMain, nil)
	}
	switch key.Key().Code {
	case tea.KeyDown:
		if menu.cursor < len(menu.tools)-1 {
			menu.cursor++
		}
	case tea.KeyUp:
		if menu.cursor > 0 {
			menu.cursor--
		}
	case tea.KeyEnter:
		menu.selected = menu.tools[menu.cursor]
	case tea.KeyEscape:
		return menu, changeScreen(ScreenMain, nil)
	}
	if key.Key().Text == "q" {
		return menu, changeScreen(ScreenMain, nil)
	}
	return menu, nil
}

func (menu ReferenceMenu) View() tea.View {
	const unavailable = "Reference not available"
	if menu.assets == nil {
		return tea.NewView("Reference\n\n" + unavailable + "\n\nPress any key to return to the menu.\n")
	}

	if len(menu.tools) == 0 {
		return tea.NewView("Reference\n\n" + unavailable + "\n\nPress any key to return to the menu.\n")
	}
	if menu.selected != "" {
		content, err := fs.ReadFile(menu.assets, "assets/cheatsheets/"+menu.selected+".md")
		if err != nil {
			return tea.NewView("Reference\n\n" + unavailable + "\n\nPress any key to return to the menu.\n")
		}
		return tea.NewView("Reference\n\n" + strings.TrimSpace(string(content)) + "\n\nPress any key to return to the menu.\n")
	}

	var content strings.Builder
	content.WriteString("Reference\n\n")
	for index, tool := range menu.tools {
		prefix := "  "
		if index == menu.cursor {
			prefix = "> "
		}
		content.WriteString(prefix)
		content.WriteString(tool)
		content.WriteByte('\n')
	}
	content.WriteString("\n↑/↓ navigate • enter select • q/esc return\n")
	return tea.NewView(content.String())
}
