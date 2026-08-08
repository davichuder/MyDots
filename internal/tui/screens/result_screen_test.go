package screens

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/installer"
)

func TestResultScreenOutcomePrecedence(t *testing.T) {
	tests := []struct {
		name      string
		result    InstallResult
		wantTitle string
		wantText  []string
	}{
		{
			name:      "success",
			result:    InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModGo, Status: installer.StatusSkipped}}},
			wantTitle: "Installation complete",
			wantText:  []string{"2 modules completed"},
		},
		{
			name:      "warnings",
			result:    InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModGo, Status: installer.StatusFailed}, {ModuleID: installer.ModDocker, Status: installer.StatusFailed}}},
			wantTitle: "Installation completed with warnings",
			wantText:  []string{"2 warnings", string(installer.ModGo), string(installer.ModDocker)},
		},
		{
			name:      "critical failure",
			result:    InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModHomebrew, Status: installer.StatusFailed}}, Err: errors.New("brew unavailable")},
			wantTitle: "Installation failed",
			wantText:  []string{string(installer.ModHomebrew), "brew unavailable"},
		},
		{
			name:      "cancellation overrides critical failure",
			result:    InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModHomebrew, Status: installer.StatusFailed}}, Cancelled: true, Err: errors.New("brew unavailable")},
			wantTitle: "Installation cancelled",
			wantText:  []string{"1 module completed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := NewResultScreen(tt.result, false)
			view := screen.View().Content

			if !strings.Contains(view, tt.wantTitle) {
				t.Errorf("View() = %q, want outcome title %q", view, tt.wantTitle)
			}
			for _, text := range tt.wantText {
				if !strings.Contains(view, text) {
					t.Errorf("View() = %q, want %q", view, text)
				}
			}
			if tt.result.Cancelled && strings.Contains(view, "Installation failed") {
				t.Errorf("View() = %q, cancellation must not appear as failure", view)
			}
		})
	}
}

func TestResultScreenShowsWSL2FontNoteForEveryOutcome(t *testing.T) {
	screen := NewResultScreen(InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}}}, true)

	if view := screen.View().Content; !strings.Contains(view, "WSL2 font note") {
		t.Errorf("View() = %q, want WSL2 font note", view)
	}
}

func TestResultScreenAnyKeyReturnsToMainMenu(t *testing.T) {
	screen := NewResultScreen(InstallResult{}, false)

	_, command := screen.Update(keyPress('x', "x"))
	message := commandMessage(t, command)
	if message != (ChangeScreenMsg{Screen: ScreenMain}) {
		t.Errorf("key command message = %#v, want ChangeScreenMsg to main menu", message)
	}
}

func TestResultScreenViewGoldens(t *testing.T) {
	tests := []struct {
		name   string
		result InstallResult
		wsl2   bool
	}{
		{name: "success", result: InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModGo, Status: installer.StatusSkipped}}}},
		{name: "warnings", result: InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModGo, Status: installer.StatusFailed}, {ModuleID: installer.ModDocker, Status: installer.StatusFailed}}}},
		{name: "critical-failure", result: InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModHomebrew, Status: installer.StatusFailed}}, Err: errors.New("brew unavailable")}},
		{name: "cancellation-progress", result: InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}, {ModuleID: installer.ModHomebrew, Status: installer.StatusFailed}}, Cancelled: true, Err: errors.New("brew unavailable")}},
		{name: "wsl2-note", result: InstallResult{Rows: []InstallResultRow{{ModuleID: installer.ModGit, Status: installer.StatusInstalled}}}, wsl2: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := NewResultScreen(tt.result, tt.wsl2)
			updated, _ := screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			screen = resultScreen(t, updated)
			got := screen.View().Content
			goldenPath := filepath.Join("testdata", "result_screen", tt.name+".golden")
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
		})
	}
}

func resultScreen(t *testing.T, model tea.Model) ResultScreen {
	t.Helper()
	screen, ok := model.(ResultScreen)
	if !ok {
		t.Fatalf("Update() model = %T, want ResultScreen", model)
	}
	return screen
}
