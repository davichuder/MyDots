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
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
	"unicode/utf8"

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

func TestAppRoutesCancelledElevationDirectlyToMainMenu(t *testing.T) {
	app := NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		InstallFactory: func(config.Config) tea.Model {
			return screens.NewInstallScreen(screens.InstallRequest{}, cancelledElevationRunner{}, cancelledElevation{})
		},
	})

	updated, _ := app.Update(screens.ChangeScreenMsg{Screen: screens.ScreenInstall, Payload: config.DefaultConfig()})
	app = updated.(App)
	updated, startCommand := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = updated.(App)
	updated, routeCommand := app.Update(appCommandMessage(t, startCommand))
	app = updated.(App)
	updated, _ = app.Update(appCommandMessage(t, routeCommand))
	app = updated.(App)

	if view := app.View().Content; !strings.Contains(view, "MyDots") {
		t.Errorf("View() = %q, want main menu after cancelled elevation", view)
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

func TestAppRoutesBackupToInjectedFactory(t *testing.T) {
	created := 0
	app := NewApp(Dependencies{
		Platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native},
		BackupFactory: func() tea.Model {
			created++
			return screens.NewPreflight(nil)
		},
	})

	updated, _ := app.Update(screens.ChangeScreenMsg{Screen: screens.ScreenBackup})
	app = updated.(App)
	if created != 1 {
		t.Errorf("backup factory calls = %d, want 1", created)
	}
	if view := app.View().Content; !strings.Contains(view, "WSL2 setup guide") {
		t.Errorf("backup view = %q, want injected backup model", view)
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
	frame := serializeTerminalFrame(canonicalTerminalFrame(output, 80, 24), 80, 24)
	if *updateProgramGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(frame), 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read program golden: %v", err)
	}
	if frame != string(want) {
		t.Errorf("program frame = %q, want golden %q", frame, want)
	}
}

func TestCanonicalTerminalFrameNormalizesEquivalentCursorMotion(t *testing.T) {
	const width, height = 40, 2
	const windowsRenderer = "\x1b[26CMyDots\n\x1b[4DConfig"
	const linuxRenderer = "\x1b[26CMyDots\n\r\x1b[28CConfig"
	const want = "frame 40x2\n00:                          MyDots\n01:                            Config"

	for name, output := range map[string]string{
		"windows-relative-left": windowsRenderer,
		"linux-absolute-right":  linuxRenderer,
	} {
		t.Run(name, func(t *testing.T) {
			frame := canonicalTerminalFrame(output, width, height)
			if got := serializeTerminalFrame(frame, width, height); got != want {
				t.Errorf("serialized terminal frame = %q, want %q", got, want)
			}
		})
	}
}

func TestSerializeTerminalFramePreservesDimensionsAndTrimsOnlyRightPadding(t *testing.T) {
	frame := "  x \n    "
	const want = "frame 4x2\n00:  x\n01:"

	if got := serializeTerminalFrame(frame, 4, 2); got != want {
		t.Errorf("serializeTerminalFrame() = %q, want %q", got, want)
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

func canonicalTerminalFrame(output string, width, height int) string {
	frame := make([][]rune, height)
	for row := range frame {
		frame[row] = make([]rune, width)
		for column := range frame[row] {
			frame[row][column] = ' '
		}
	}

	row, column := 0, 0
	for index := 0; index < len(output); {
		if output[index] == '\x1b' && index+1 < len(output) && output[index+1] == '[' {
			parameters, final, next, ok := parseCSI(output, index+2)
			if ok {
				row, column = applyCSI(frame, row, column, parameters, final)
				index = next
				continue
			}
		}

		r, size := utf8.DecodeRuneInString(output[index:])
		index += size
		switch r {
		case '\r':
			column = 0
		case '\n':
			if row < height-1 {
				row++
			}
		default:
			if r < ' ' || row >= height || column >= width {
				continue
			}
			frame[row][column] = r
			column++
		}
	}

	lines := make([]string, height)
	for row := range frame {
		lines[row] = string(frame[row])
	}
	return strings.Join(lines, "\n")
}

func serializeTerminalFrame(frame string, width, height int) string {
	rows := strings.Split(frame, "\n")
	rowNumberWidth := max(len(strconv.Itoa(height-1)), 2)
	serialized := make([]string, 0, height+1)
	serialized = append(serialized, fmt.Sprintf("frame %dx%d", width, height))
	for row := 0; row < height; row++ {
		line := ""
		if row < len(rows) {
			line = strings.TrimRight(rows[row], " ")
		}
		serialized = append(serialized, fmt.Sprintf("%0*d:%s", rowNumberWidth, row, line))
	}
	return strings.Join(serialized, "\n")
}

func parseCSI(output string, index int) (parameters []int, final byte, next int, ok bool) {
	start := index
	for index < len(output) {
		if output[index] >= 0x40 && output[index] <= 0x7e {
			return csiParameters(output[start:index]), output[index], index + 1, true
		}
		index++
	}
	return nil, 0, index, false
}

func csiParameters(raw string) []int {
	if raw == "" {
		return nil
	}
	parameters := make([]int, 0, strings.Count(raw, ";")+1)
	for _, parameter := range strings.Split(raw, ";") {
		value, err := strconv.Atoi(parameter)
		if err != nil {
			return nil
		}
		parameters = append(parameters, value)
	}
	return parameters
}

func applyCSI(frame [][]rune, row, column int, parameters []int, final byte) (int, int) {
	height, width := len(frame), len(frame[0])
	count := csiCount(parameters)
	switch final {
	case 'C':
		column = min(column+count, width)
	case 'D':
		column = max(column-count, 0)
	case 'G':
		column = min(max(count-1, 0), width)
	case 'H', 'f':
		row = min(max(csiParameter(parameters, 0, 1)-1, 0), height-1)
		column = min(max(csiParameter(parameters, 1, 1)-1, 0), width)
	case 'J':
		if csiParameter(parameters, 0, 0) == 2 {
			for currentRow := range frame {
				for currentColumn := range frame[currentRow] {
					frame[currentRow][currentColumn] = ' '
				}
			}
		}
	case 'K':
		for currentColumn := column; currentColumn < width; currentColumn++ {
			frame[row][currentColumn] = ' '
		}
	}
	return row, column
}

func csiCount(parameters []int) int {
	return csiParameter(parameters, 0, 1)
}

func csiParameter(parameters []int, index, fallback int) int {
	if index >= len(parameters) || parameters[index] == 0 {
		return fallback
	}
	return parameters[index]
}

func appCommandMessage(t *testing.T, command tea.Cmd) tea.Msg {
	t.Helper()
	if command == nil {
		t.Fatal("command = nil")
	}
	return command()
}

var updateProgramGolden = flag.Bool("update", false, "update program golden files")

type sizeStartingModel struct{ started *int }

type cancelledElevation struct{}

func (cancelledElevation) Acquire(context.Context) (screens.Keepalive, error) {
	return nil, context.Canceled
}

type cancelledElevationRunner struct{}

func (cancelledElevationRunner) Start(context.Context, screens.InstallRequest) <-chan screens.InstallEvent {
	return make(chan screens.InstallEvent)
}

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
