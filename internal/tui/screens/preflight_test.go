package screens

import (
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
)

func TestPreflightRendersEmbeddedGuideAndQuitsOnKey(t *testing.T) {
	preflight := NewPreflight(fstest.MapFS{
		"assets/wsl2-guide.md": &fstest.MapFile{Data: []byte("Install WSL2 before continuing.")},
	})

	if view := preflight.View().Content; !strings.Contains(view, "Install WSL2 before continuing.") {
		t.Errorf("View() = %q, want embedded guide", view)
	}
	_, command := preflight.Update(tea.KeyPressMsg(tea.Key{Text: "x"}))
	if message := commandMessage(t, command); message != tea.Quit() {
		t.Errorf("key command = %T, want tea.Quit", message)
	}
}

func TestPreflightReportsMissingGuide(t *testing.T) {
	preflight := NewPreflight(fstest.MapFS{})

	if view := preflight.View().Content; !strings.Contains(view, "not available") {
		t.Errorf("View() = %q, want unavailable guide message", view)
	}
}
