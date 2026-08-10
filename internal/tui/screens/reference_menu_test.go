package screens

import (
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
)

func TestReferenceMenuReportsUnavailableAssets(t *testing.T) {
	menu := NewReferenceMenu(nil)

	if view := menu.View().Content; !strings.Contains(view, "Reference not available") {
		t.Errorf("View() = %q, want unavailable reference message", view)
	}
}

func TestReferenceMenuRendersEmbeddedContent(t *testing.T) {
	menu := NewReferenceMenu(fstest.MapFS{
		"assets/cheatsheets/ghostty.md": &fstest.MapFile{Data: []byte("# Ghostty\n\nOpen a terminal.")},
	})

	updated, _ := menu.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	menu = updated.(ReferenceMenu)
	view := menu.View().Content
	if !strings.Contains(view, "# Ghostty") {
		t.Errorf("View() = %q, want embedded reference title", view)
	}
	if !strings.Contains(view, "Open a terminal.") {
		t.Errorf("View() = %q, want embedded reference content", view)
	}
}

func TestReferenceMenuListsAndSelectsEveryEmbeddedTool(t *testing.T) {
	menu := NewReferenceMenu(fstest.MapFS{
		"assets/cheatsheets/ghostty.md":  &fstest.MapFile{Data: []byte("# Ghostty\n\nOpen a terminal.")},
		"assets/cheatsheets/homebrew.md": &fstest.MapFile{Data: []byte("# Homebrew\n\nInstall packages.")},
		"assets/cheatsheets/ignored.txt": &fstest.MapFile{Data: []byte("not a reference")},
	})

	list := menu.View().Content
	for _, tool := range []string{"ghostty", "homebrew"} {
		if !strings.Contains(list, tool) {
			t.Errorf("reference list = %q, want tool %q", list, tool)
		}
	}
	if strings.Contains(list, "Open a terminal.") {
		t.Errorf("reference list = %q, must not render a cheatsheet before selection", list)
	}

	updated, _ := menu.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	menu = updated.(ReferenceMenu)
	updated, _ = menu.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	menu = updated.(ReferenceMenu)
	if view := menu.View().Content; !strings.Contains(view, "# Homebrew") || !strings.Contains(view, "Install packages.") {
		t.Errorf("selected reference = %q, want homebrew cheatsheet", view)
	}
}

func TestReferenceMenuReturnsToMainOnKey(t *testing.T) {
	menu := NewReferenceMenu(nil)
	_, command := menu.Update(tea.KeyPressMsg(tea.Key{Text: "x"}))

	message, ok := commandMessage(t, command).(ChangeScreenMsg)
	if !ok {
		t.Fatalf("key command = %T, want ChangeScreenMsg", commandMessage(t, command))
	}
	if message.Screen != ScreenMain {
		t.Errorf("screen = %q, want %q", message.Screen, ScreenMain)
	}
}
