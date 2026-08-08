package screens

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/davichuder/MyDots/internal/config"
)

func TestConfigMenuDelegatesHuhV2FormLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mydots-config.json")
	menu := NewConfigMenu(path)

	if menu.form == nil {
		t.Fatal("NewConfigMenu() form = nil, want Huh v2 form")
	}
	if menu.form.State != huh.StateNormal {
		t.Fatalf("form state = %v, want %v", menu.form.State, huh.StateNormal)
	}

	updated, command := menu.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if command == nil {
		t.Fatal("Update() command = nil, want Huh v2 form lifecycle command")
	}
	if _, ok := updated.(ConfigMenu); !ok {
		t.Fatalf("Update() model = %T, want ConfigMenu", updated)
	}
}

func TestConfigMenuStartsWithSixDefaultChoices(t *testing.T) {
	menu := NewConfigMenu(filepath.Join(t.TempDir(), "mydots-config.json"))
	want := config.DefaultConfig()

	if menu.config != want {
		t.Errorf("initial config = %#v, want DefaultConfig() %#v", menu.config, want)
	}
	if got := len(menu.fontOptions); got != 5 {
		t.Errorf("font options = %d, want 5", got)
	}
}

func TestConfigMenuRejectsHTTPURLWithoutCompleting(t *testing.T) {
	menu := NewConfigMenu(filepath.Join(t.TempDir(), "mydots-config.json"))
	menu.config.Chezmoi.RepoURL = "http://github.com/example/dotfiles"
	menu.form = menu.newForm()

	for range 6 {
		menu, _ = submitConfigMenuStep(t, menu)
	}

	if menu.form.State == huh.StateCompleted {
		t.Fatal("form completed with HTTP URL, want it to remain on the URL step")
	}
	if view := menu.View().Content; !strings.Contains(view, "HTTPS") {
		t.Errorf("View() = %q, want HTTPS validation error", view)
	}
}

func TestConfigMenuSavesValidConfigBeforeReturningToMainMenu(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "mydots-config.json")
	menu := NewConfigMenu(path)

	var command tea.Cmd
	for range 6 {
		menu, command = submitConfigMenuStep(t, menu)
	}

	message := commandMessage(t, command)
	changed, ok := message.(ChangeScreenMsg)
	if !ok {
		t.Fatalf("save command message = %T, want ChangeScreenMsg", message)
	}
	if changed.Screen != ScreenMain {
		t.Errorf("return screen = %q, want %q", changed.Screen, ScreenMain)
	}

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("load saved config: %v", err)
	}
	if saved != config.DefaultConfig() {
		t.Errorf("saved config = %#v, want defaults %#v", saved, config.DefaultConfig())
	}
}

func TestConfigMenuPersistsChangedChoicesThroughHuhLifecycle(t *testing.T) {
	tests := []struct {
		name         string
		selectPHP    bool
		wantLanguage config.LanguageOptions
	}{
		{
			name:         "java only",
			wantLanguage: config.LanguageOptions{Java: true},
		},
		{
			name:         "java and PHP",
			selectPHP:    true,
			wantLanguage: config.LanguageOptions{Java: true, PHP: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mydots-config.json")
			menu := NewConfigMenu(path)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
			menu, _ = submitConfigMenuStep(t, menu)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
			menu, _ = submitConfigMenuStep(t, menu)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: ' '}))
			if tt.selectPHP {
				menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
				menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: ' '}))
			}
			menu, _ = submitConfigMenuStep(t, menu)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
			menu, _ = submitConfigMenuStep(t, menu)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
			menu, _ = submitConfigMenuStep(t, menu)

			menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd}))
			for range config.DefaultChezmoiRepoURL {
				menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
			}
			for _, character := range "https://example.com/dotfiles" {
				menu = updateConfigMenu(t, menu, tea.KeyPressMsg(tea.Key{Code: character, Text: string(character)}))
			}
			_, command := submitConfigMenuStep(t, menu)

			if message := commandMessage(t, command); message != (ChangeScreenMsg{Screen: ScreenMain}) {
				t.Fatalf("save command message = %#v, want ChangeScreenMsg for the main menu", message)
			}

			saved, err := config.Load(path)
			if err != nil {
				t.Fatalf("load saved config: %v", err)
			}
			want := config.DefaultConfig()
			want.Font = config.FontFiraCode
			want.Theme = config.ThemeCatppuccin
			want.Languages = tt.wantLanguage
			want.Nvim.Config = config.NvimConfigPersonal
			want.Nvim.Framework = config.NvimFrameworkLazyVim
			want.Chezmoi.RepoURL = "https://example.com/dotfiles"
			if saved != want {
				t.Errorf("saved config = %#v, want %#v", saved, want)
			}
		})
	}
}

func TestConfigMenuStepGoldens(t *testing.T) {
	menu := NewConfigMenu(filepath.Join(t.TempDir(), "mydots-config.json"))
	menu = updateConfigMenu(t, menu, tea.WindowSizeMsg{Width: 80, Height: 24})

	for step := 1; step <= 6; step++ {
		t.Run("step-"+string(rune('0'+step)), func(t *testing.T) {
			goldenPath := filepath.Join("testdata", "config_menu", "step-"+string(rune('0'+step))+".golden")
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
				t.Errorf("View() for step %d = %q, want %q", step, got, want)
			}
		})
		if step < 6 {
			menu, _ = submitConfigMenuStep(t, menu)
		}
	}
}

func updateConfigMenu(t *testing.T, menu ConfigMenu, message tea.Msg) ConfigMenu {
	t.Helper()
	updated, _ := menu.Update(message)
	next, ok := updated.(ConfigMenu)
	if !ok {
		t.Fatalf("Update() model = %T, want ConfigMenu", updated)
	}
	return next
}

func submitConfigMenuStep(t *testing.T, menu ConfigMenu) (ConfigMenu, tea.Cmd) {
	t.Helper()
	updated, command := menu.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	next, ok := updated.(ConfigMenu)
	if !ok {
		t.Fatalf("Update() model = %T, want ConfigMenu", updated)
	}
	for command != nil {
		updated, command = next.Update(command())
		next, ok = updated.(ConfigMenu)
		if !ok {
			t.Fatalf("command Update() model = %T, want ConfigMenu", updated)
		}
		if next.form.State == huh.StateCompleted {
			return next, command
		}
	}
	return next, nil
}
