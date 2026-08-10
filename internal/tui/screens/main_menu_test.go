package screens

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
)

var updateMainMenuGolden = flag.Bool("update", false, "update main menu golden files")

type fakeConfigStore struct {
	config config.Config
	err    error
}

func (store fakeConfigStore) Load() (config.Config, error) {
	return store.config, store.err
}

func TestMainMenuViewContainsExactlyFiveChoices(t *testing.T) {
	menu := NewMainMenu(fakeConfigStore{})

	view := menu.View().Content
	choices := []string{"Config", "Install", "Backup", "Reference", "Quit"}
	for _, choice := range choices {
		if count := strings.Count(view, choice); count != 1 {
			t.Errorf("View() contains %q %d times, want exactly once\n%s", choice, count, view)
		}
	}
	if got := strings.Count(view, "\n  ") + strings.Count(view, "\n> "); got != len(choices) {
		t.Errorf("View() renders %d choice rows, want %d\n%s", got, len(choices), view)
	}
}

func TestMainMenuQuitReturnsQuitCommand(t *testing.T) {
	menu := NewMainMenu(fakeConfigStore{})
	for range 4 {
		menu = updateMenu(t, menu, keyPress(tea.KeyDown, ""))
	}

	_, command := menu.Update(keyPress(tea.KeyEnter, ""))
	if _, ok := commandMessage(t, command).(tea.QuitMsg); !ok {
		t.Errorf("Quit command message = %T, want tea.QuitMsg", commandMessage(t, command))
	}
}

func TestMainMenuRoutesStaticChoicesToExpectedScreens(t *testing.T) {
	tests := []struct {
		name string
		down int
		want Screen
	}{
		{name: "config", down: 0, want: ScreenConfig},
		{name: "backup", down: 2, want: ScreenBackup},
		{name: "reference", down: 3, want: ScreenReference},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMainMenu(fakeConfigStore{})
			for range tt.down {
				menu = updateMenu(t, menu, keyPress(tea.KeyDown, ""))
			}

			_, command := menu.Update(keyPress(tea.KeyEnter, ""))
			got := commandMessage(t, command)
			message, ok := got.(ChangeScreenMsg)
			if !ok {
				t.Fatalf("choice command message = %T, want ChangeScreenMsg", got)
			}
			if message.Screen != tt.want {
				t.Errorf("route = %q, want %q", message.Screen, tt.want)
			}
		})
	}
}

func TestMainMenuMissingConfigStaysInline(t *testing.T) {
	menu := NewMainMenu(fakeConfigStore{err: os.ErrNotExist})
	menu = updateMenu(t, menu, keyPress(tea.KeyDown, ""))

	updated, command := menu.Update(keyPress(tea.KeyEnter, ""))
	if command != nil {
		t.Fatalf("Install with missing config returned a command, want nil")
	}
	if _, ok := updated.(MainMenu); !ok {
		t.Fatalf("Install with missing config changed model to %T, want MainMenu", updated)
	}
	if view := updated.View().Content; !strings.Contains(view, "No saved configuration") {
		t.Errorf("View() = %q, want inline missing-config error", view)
	}
}

func TestMainMenuBadConfigRoutesToAssistant(t *testing.T) {
	tests := []struct {
		name  string
		store fakeConfigStore
		want  string
	}{
		{
			name:  "unloadable",
			store: fakeConfigStore{err: errors.New("invalid JSON")},
			want:  "invalid JSON",
		},
		{
			name:  "invalid",
			store: fakeConfigStore{config: config.Config{Font: "not-a-font"}},
			want:  "invalid font",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMainMenu(tt.store)
			menu = updateMenu(t, menu, keyPress(tea.KeyDown, ""))

			_, command := menu.Update(keyPress(tea.KeyEnter, ""))
			message, ok := commandMessage(t, command).(ChangeScreenMsg)
			if !ok {
				t.Fatalf("Install command message = %T, want ChangeScreenMsg", commandMessage(t, command))
			}
			if message.Screen != ScreenConfig {
				t.Errorf("route = %q, want %q", message.Screen, ScreenConfig)
			}
			issue, ok := message.Payload.(ConfigIssue)
			if !ok {
				t.Fatalf("payload = %T, want ConfigIssue", message.Payload)
			}
			if !strings.Contains(issue.Err.Error(), tt.want) {
				t.Errorf("ConfigIssue.Err = %q, want it to contain %q", issue.Err, tt.want)
			}
		})
	}
}

func TestMainMenuValidConfigRoutesToInstall(t *testing.T) {
	valid := config.DefaultConfig()
	menu := NewMainMenu(fakeConfigStore{config: valid})
	menu = updateMenu(t, menu, keyPress(tea.KeyDown, ""))

	_, command := menu.Update(keyPress(tea.KeyEnter, ""))
	message, ok := commandMessage(t, command).(ChangeScreenMsg)
	if !ok {
		t.Fatalf("Install command message = %T, want ChangeScreenMsg", commandMessage(t, command))
	}
	if message.Screen != ScreenInstall {
		t.Errorf("route = %q, want %q", message.Screen, ScreenInstall)
	}
	if got, ok := message.Payload.(config.Config); !ok || got != valid {
		t.Errorf("install payload = %#v, want loaded valid config %#v", message.Payload, valid)
	}
}

func TestMainMenuViewGolden(t *testing.T) {
	menu := NewMainMenu(fakeConfigStore{})
	menu = updateMenu(t, menu, tea.WindowSizeMsg{Width: 80, Height: 24})

	const goldenPath = "testdata/main_menu/default.golden"
	got := menu.View().Content
	if *updateMainMenuGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("View() at 80x24 = %q, want %q", got, want)
	}
}

func updateMenu(t *testing.T, menu MainMenu, message tea.Msg) MainMenu {
	t.Helper()
	updated, _ := menu.Update(message)
	next, ok := updated.(MainMenu)
	if !ok {
		t.Fatalf("Update() model = %T, want MainMenu", updated)
	}
	return next
}

func commandMessage(t *testing.T, command tea.Cmd) tea.Msg {
	t.Helper()
	if command == nil {
		t.Fatal("Update() command = nil")
	}
	return command()
}

func keyPress(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: text})
}
