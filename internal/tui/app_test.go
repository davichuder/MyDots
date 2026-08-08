package tui

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/tui/screens"
)

func TestProgramStartsDarwinMainMenu(t *testing.T) {
	output := runProgram(t, NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		Assets:   testAssets{},
	}), "\x1b[B\x1b[B\x1b[B\x1b[B\r")

	if !strings.Contains(output, "MyDots") {
		t.Errorf("program output = %q, want main menu", output)
	}
}

func TestAppStartsSupportedPlatformsAndRejectsUnsupportedPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform platform.Platform
		want     string
	}{
		{name: "darwin", platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native}, want: "MyDots"},
		{name: "linux", platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}, want: "MyDots"},
		{name: "windows", platform: platform.Platform{OS: platform.Windows}, want: "WSL2 setup guide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(Dependencies{Platform: tt.platform})
			if view := app.View().Content; !strings.Contains(view, tt.want) {
				t.Errorf("View() = %q, want %q", view, tt.want)
			}
		})
	}

	app := NewApp(Dependencies{Platform: platform.Platform{OS: "plan9"}})
	if app.startupError == nil {
		t.Fatal("unsupported platform startup error = nil, want explicit error")
	}
	var unsupported UnsupportedPlatformError
	if !errors.As(app.startupError, &unsupported) {
		t.Fatalf("startup error = %T, want UnsupportedPlatformError", app.startupError)
	}
	if !strings.Contains(app.startupError.Error(), "unsupported platform") {
		t.Errorf("startup error = %q, want unsupported-platform context", app.startupError)
	}
}

func TestProgramShowsWindowsGuideAndExitsOnKey(t *testing.T) {
	installStarts := 0
	output := runProgram(t, NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Windows},
		Assets: fstest.MapFS{
			"assets/wsl2-guide.md": &fstest.MapFile{Data: []byte("Install WSL2 with: wsl --install")},
		},
		InstallFactory: func(config.Config) tea.Model {
			installStarts++
			return screens.NewPreflight(nil)
		},
	}), "x")

	if !strings.Contains(output, "Install WSL2 with: wsl --install") {
		t.Errorf("program output = %q, want Windows guide", output)
	}
	if installStarts != 0 {
		t.Errorf("Windows program started %d installation screens, want 0", installStarts)
	}
}

func TestAppReplaysWindowSizeToRoutedInstallScreen(t *testing.T) {
	started := 0
	app := NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		InstallFactory: func(config.Config) tea.Model {
			return sizeStartingModel{started: &started}
		},
	})

	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = updated.(App)
	updated, _ = app.Update(screens.ChangeScreenMsg{
		Screen:  screens.ScreenInstall,
		Payload: config.DefaultConfig(),
	})
	app = updated.(App)
	if started != 1 {
		t.Errorf("install starts = %d, want 1 after routed window size replay; app=%T", started, app.model)
	}
}

func TestAppRoutesConfigIssueToAssistantWithOriginalError(t *testing.T) {
	issue := errors.New("saved config contains invalid JSON")
	app := NewApp(Dependencies{ConfigPath: "mydots-config.json"})

	updated, _ := app.Update(screens.ChangeScreenMsg{
		Screen:  screens.ScreenConfig,
		Payload: screens.ConfigIssue{Err: issue},
	})
	routed, ok := updated.(App)
	if !ok {
		t.Fatalf("Update() model = %T, want App", updated)
	}

	if view := routed.View().Content; !strings.Contains(view, issue.Error()) {
		t.Errorf("config assistant view = %q, want original error %q", view, issue)
	}
}

func TestAppRendersConfigIssueEmittedByConfigMenuWithoutRetryLoop(t *testing.T) {
	issue := errors.New("configuration save failed: disk is read-only")
	app := NewApp(Dependencies{Platform: platform.Platform{OS: platform.Darwin}, ConfigPath: "mydots-config.json"})

	updated, _ := app.Update(screens.ConfigIssue{Err: issue})
	routed := updated.(App)
	if view := routed.View().Content; !strings.Contains(view, issue.Error()) {
		t.Errorf("config assistant view = %q, want original error %q", view, issue)
	}
}

func TestProgramMainMenuGolden(t *testing.T) {
	output := runProgram(t, NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
	}), "\x1b[B\x1b[B\x1b[B\x1b[B\r")

	const goldenPath = "testdata/app/main_menu_program.golden"
	if *updateProgramGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(output), 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read program golden: %v", err)
	}
	if output != string(want) {
		t.Errorf("program output = %q, want golden %q", output, want)
	}
}

func TestProgramHarnessIgnoresHostTraceEnvironment(t *testing.T) {
	tracePath := filepath.Join(t.TempDir(), "host-trace.log")
	t.Setenv("TEA_TRACE", tracePath)

	_ = runProgram(t, NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
	}), "\x1b[B\x1b[B\x1b[B\x1b[B\r")

	if _, err := os.Stat(tracePath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("host TEA_TRACE file error = %v, want not exist", err)
	}
	if got := os.Getenv("TEA_TRACE"); got != tracePath {
		t.Errorf("host TEA_TRACE = %q, want restored value %q", got, tracePath)
	}
}

func runProgram(t *testing.T, model tea.Model, input string) string {
	t.Helper()
	trace, traceSet := os.LookupEnv("TEA_TRACE")
	if err := os.Setenv("TEA_TRACE", ""); err != nil {
		t.Fatalf("neutralize TEA_TRACE: %v", err)
	}
	defer func() {
		if traceSet {
			_ = os.Setenv("TEA_TRACE", trace)
			return
		}
		_ = os.Unsetenv("TEA_TRACE")
	}()

	var output bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	program := tea.NewProgram(model,
		tea.WithInput(strings.NewReader(input)),
		tea.WithOutput(&output),
		tea.WithContext(ctx),
		tea.WithWindowSize(80, 24),
		tea.WithEnvironment([]string{"TERM=dumb", "TEA_TRACE="}),
		tea.WithoutSignals(),
		tea.WithoutSignalHandler(),
	)
	if _, err := program.Run(); err != nil {
		t.Fatalf("run program: %v", err)
	}
	return output.String()
}

var updateProgramGolden = flag.Bool("update", false, "update program golden files")

type sizeStartingModel struct{ started *int }

func (model sizeStartingModel) Init() tea.Cmd { return nil }

func (model sizeStartingModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := message.(tea.WindowSizeMsg); ok {
		*model.started++
	}
	return model, nil
}

func (model sizeStartingModel) View() tea.View {
	return tea.NewView(fmt.Sprintf("starts=%d", *model.started))
}

type testAssets struct{}

func (testAssets) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }
