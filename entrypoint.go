package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/sudo"
	"github.com/davichuder/MyDots/internal/tui"
	"github.com/davichuder/MyDots/internal/tui/screens"
)

const missingConfigMessage = "Error: no config found. Run 'mydots' to configure first."

type keepaliveHandle interface{ Stop() }

type entrypointDependencies struct {
	stdout, stderr   io.Writer
	assets           fs.FS
	configPath       func() string
	loadConfig       func(string) (config.Config, error)
	saveConfig       func(string, config.Config) error
	defaultConfig    func() config.Config
	validateConfig   func(config.Config) error
	detectPlatform   func(string) (platform.Platform, error)
	showGuide        func(fs.FS, string)
	runTUI           func(platform.Platform) error
	requestElevation func(context.Context) error
	startKeepalive   func(context.Context) (keepaliveHandle, error)
	runPipeline      func(context.Context, config.Config, string, platform.Platform) error
	newContext       func() context.Context
	now              func() time.Time
}

var entrypointFactoryMu sync.Mutex
var entrypointFactory = productionEntrypointDependencies

func run(args []string, goos string) int {
	dependencies := entrypointFactory()
	flags := flag.NewFlagSet("mydots", flag.ContinueOnError)
	flags.SetOutput(dependencies.stderr)
	unattended := flags.Bool("unattended", false, "run saved configuration")
	defaults := flags.Bool("default", false, "save and run default configuration")
	version := flags.Bool("version", false, "print version")
	if len(args) == 0 {
		args = []string{"mydots"}
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 1
	}
	if *version {
		_, _ = fmt.Fprintln(dependencies.stdout, buildVersion)
		return 0
	}
	if goos == "windows" {
		dependencies.showGuide(dependencies.assets, "assets/wsl2-guide.md")
		return 0
	}
	if goos != "darwin" && goos != "linux" {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: unsupported platform: %s\n", goos)
		return 1
	}
	currentPlatform, err := dependencies.detectPlatform(goos)
	if err != nil {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	if *unattended {
		configuration, err := dependencies.loadConfig(dependencies.configPath())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				_, _ = fmt.Fprintln(dependencies.stderr, missingConfigMessage)
			} else {
				_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
			}
			return 1
		}
		return runPrivileged(dependencies, configuration, currentPlatform)
	}
	if *defaults {
		configuration := dependencies.defaultConfig()
		if err := dependencies.saveConfig(dependencies.configPath(), configuration); err != nil {
			_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
			return 1
		}
		return runPrivileged(dependencies, configuration, currentPlatform)
	}
	if err := dependencies.runTUI(currentPlatform); err != nil {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func runPrivileged(dependencies entrypointDependencies, configuration config.Config, currentPlatform platform.Platform) int {
	if err := dependencies.validateConfig(configuration); err != nil {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	ctx := dependencies.newContext()
	if err := dependencies.requestElevation(ctx); err != nil {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	keepalive, err := dependencies.startKeepalive(ctx)
	if err != nil || keepalive == nil {
		if err == nil {
			err = errors.New("could not start sudo keepalive")
		}
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	defer keepalive.Stop()
	timestamp := dependencies.now().UTC().Format("2006-01-02T15-04-05Z")
	if err := dependencies.runPipeline(ctx, configuration, timestamp, currentPlatform); err != nil {
		_, _ = fmt.Fprintf(dependencies.stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func productionEntrypointDependencies() entrypointDependencies {
	return entrypointDependencies{
		stdout:         os.Stdout,
		stderr:         os.Stderr,
		assets:         assets,
		configPath:     config.DefaultConfigPath,
		loadConfig:     config.Load,
		saveConfig:     config.Save,
		defaultConfig:  config.DefaultConfig,
		validateConfig: config.Validate,
		detectPlatform: platform.Detect,
		showGuide: func(assets fs.FS, path string) {
			guide, err := fs.ReadFile(assets, path)
			if err != nil {
				guide = []byte("WSL2 setup guide is not available.\n")
			}
			_, _ = fmt.Fprint(os.Stdout, string(guide))
		},
		runTUI:           runProductionTUI,
		requestElevation: func(context.Context) error { return sudo.RequestElevation() },
		startKeepalive: func(ctx context.Context) (keepaliveHandle, error) {
			return sudo.StartKeepalive(ctx)
		},
		runPipeline: runProductionPipeline,
		newContext:  context.Background,
		now:         time.Now,
	}
}

func runProductionTUI(currentPlatform platform.Platform) error {
	path := config.DefaultConfigPath()
	terminal := &programTerminal{}
	app := productionTUIApp(currentPlatform, path, filepath.Join(filepath.Dir(path), "backups"), time.Now, terminal)
	program := tea.NewProgram(app)
	terminal.program = program
	_, err := program.Run()
	return err
}

func productionTUIApp(currentPlatform platform.Platform, configPath, backupRoot string, now func() time.Time, terminal screens.TerminalHandoff) tui.App {
	if now == nil {
		now = time.Now
	}
	installerRunner := screens.NewModuleSessionRunner(func(request screens.InstallRequest) []installer.Module {
		return installer.BuildPlan(request.Config, request.Platform)
	})
	sudoValidator := screens.NewCommandSudoValidator(commandExecutor{}, screens.CommandStreams{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
	elevation := screens.NewSudoElevation(terminal, systemClock{}, sudoValidator)
	return tui.NewApp(tui.Dependencies{
		Platform:    currentPlatform,
		Assets:      assets,
		ConfigPath:  configPath,
		ConfigStore: fileConfigStore{path: configPath},
		InstallFactory: func(configuration config.Config) tea.Model {
			return screens.NewInstallScreen(screens.InstallRequest{
				Config:    configuration,
				Platform:  currentPlatform,
				Assets:    assets,
				Timestamp: now().UTC().Format("2006-01-02T15-04-05Z"),
			}, installerRunner, elevation)
		},
		BackupFactory: func() tea.Model {
			return screens.NewBackupMenu(screens.NewFileBackupStore(backupRoot, []screens.ManagedPath{{Source: configPath, RelativeDestination: "mydots-config.json"}}, func() string {
				return now().UTC().Format("2006-01-02T15-04-05Z")
			}))
		},
	})
}

type fileConfigStore struct{ path string }

func (store fileConfigStore) Load() (config.Config, error) { return config.Load(store.path) }

func runProductionPipeline(ctx context.Context, configuration config.Config, timestamp string, currentPlatform platform.Platform) error {
	return runProductionPipelineWith(ctx, configuration, timestamp, currentPlatform, installer.Run)
}

func runProductionPipelineWith(ctx context.Context, configuration config.Config, timestamp string, currentPlatform platform.Platform, runInstaller func([]installer.Module, installer.InstallContext, chan installer.ProgressEvent)) error {
	plan := installer.BuildPlan(configuration, currentPlatform)
	progress := make(chan installer.ProgressEvent, len(plan)*2)
	go runInstaller(plan, installer.InstallContext{
		Platform:         currentPlatform,
		Config:           configuration,
		SessionTimestamp: timestamp,
		Log:              os.Stdout,
		Cancel:           ctx,
		Assets:           assets,
	}, progress)
	criticalModules := make(map[installer.ModuleID]bool, len(plan))
	for _, module := range plan {
		criticalModules[module.ID()] = module.Criticality() == installer.Critical
	}
	for event := range progress {
		if event.Err != nil && criticalModules[event.ModuleID] {
			return event.Err
		}
	}
	return nil
}

type programTerminal struct{ program *tea.Program }

func (terminal *programTerminal) Release() error { return terminal.program.ReleaseTerminal() }
func (terminal *programTerminal) Restore() error { return terminal.program.RestoreTerminal() }

type commandExecutor struct{}

func (commandExecutor) Execute(ctx context.Context, command screens.Command) error {
	execution := exec.CommandContext(ctx, command.Name, command.Args...)
	execution.Stdin = command.Stdin
	execution.Stdout = command.Stdout
	execution.Stderr = command.Stderr
	return execution.Run()
}

type systemClock struct{}

func (systemClock) Every(interval time.Duration) screens.Ticker {
	return systemTicker{ticker: time.NewTicker(interval)}
}

type systemTicker struct{ ticker *time.Ticker }

func (ticker systemTicker) Chan() <-chan time.Time { return ticker.ticker.C }
func (ticker systemTicker) Stop()                  { ticker.ticker.Stop() }
