package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/tui/screens"
)

func TestRunRoutes(t *testing.T) {
	for _, test := range []struct {
		name       string
		args       []string
		goos       string
		configure  func(*routeRecorder, *entrypointDependencies)
		wantCode   int
		wantStdout string
		wantStderr string
		wantCalls  []string
		wantStamp  string
	}{
		{
			name:       "version skips every other route",
			args:       []string{"mydots", "--version"},
			goos:       "plan9",
			wantCode:   0,
			wantStdout: "test-version\n",
		},
		{
			name:      "windows shows guide only",
			args:      []string{"mydots"},
			goos:      "windows",
			wantCode:  0,
			wantCalls: []string{"guide"},
		},
		{
			name:       "unsupported platform reports an error",
			args:       []string{"mydots"},
			goos:       "plan9",
			wantCode:   1,
			wantStderr: "Error: unsupported platform: plan9\n",
		},
		{
			name: "unattended missing config preserves FR-22",
			args: []string{"mydots", "--unattended"},
			goos: "linux",
			configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
				dependencies.loadConfig = func(string) (config.Config, error) {
					recorder.record("load")
					return config.Config{}, os.ErrNotExist
				}
			},
			wantCode:   1,
			wantStderr: "Error: no config found. Run 'mydots' to configure first.\n",
			wantCalls:  []string{"load"},
		},
		{
			name:      "unattended runs without TUI",
			args:      []string{"mydots", "--unattended"},
			goos:      "linux",
			wantCode:  0,
			wantCalls: []string{"load", "elevation", "keepalive", "pipeline", "shutdown"},
			wantStamp: "2026-08-10T11-00-00Z",
		},
		{
			name:      "default saves then runs without TUI",
			args:      []string{"mydots", "--default"},
			goos:      "darwin",
			wantCode:  0,
			wantCalls: []string{"save", "elevation", "keepalive", "pipeline", "shutdown"},
			wantStamp: "2026-08-10T11-00-00Z",
		},
		{
			name:      "no flags starts TUI only",
			args:      []string{"mydots"},
			goos:      "linux",
			wantCode:  0,
			wantCalls: []string{"tui"},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			recorder := &routeRecorder{}
			dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
			if test.configure != nil {
				test.configure(recorder, &dependencies)
			}

			originalVersion := buildVersion
			buildVersion = "test-version"
			defer func() { buildVersion = originalVersion }()
			withTestEntrypointDependencies(t, dependencies)

			if got := run(test.args, test.goos); got != test.wantCode {
				t.Errorf("run() = %d, want %d", got, test.wantCode)
			}
			if got := stdout.String(); got != test.wantStdout {
				t.Errorf("stdout = %q, want %q", got, test.wantStdout)
			}
			if got := stderr.String(); got != test.wantStderr {
				t.Errorf("stderr = %q, want %q", got, test.wantStderr)
			}
			if got := strings.Join(recorder.calls, ","); got != strings.Join(test.wantCalls, ",") {
				t.Errorf("calls = %q, want %q", got, strings.Join(test.wantCalls, ","))
			}
			if got := recorder.timestamp; got != test.wantStamp {
				t.Errorf("pipeline timestamp = %q, want %q", got, test.wantStamp)
			}
		})
	}
}

func TestRunPropagatesDetectedWSL2PlatformToTUIAndPipeline(t *testing.T) {
	wantPlatform := platform.Platform{OS: platform.Linux, Variant: platform.WSL2, Arch: "amd64"}
	var stdout, stderr bytes.Buffer
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, &routeRecorder{})
	dependencies.detectPlatform = func(goos string) (platform.Platform, error) {
		if goos != "linux" {
			t.Errorf("detector GOOS = %q, want linux", goos)
		}
		return wantPlatform, nil
	}
	var tuiPlatform, pipelinePlatform platform.Platform
	dependencies.runTUI = func(got platform.Platform) error { tuiPlatform = got; return nil }
	dependencies.runPipeline = func(_ context.Context, _ config.Config, _ string, got platform.Platform) error {
		pipelinePlatform = got
		return nil
	}
	withTestEntrypointDependencies(t, dependencies)

	if code := run([]string{"mydots"}, "linux"); code != 0 {
		t.Errorf("TUI run() = %d, want 0", code)
	}
	if code := run([]string{"mydots", "--unattended"}, "linux"); code != 0 {
		t.Errorf("privileged run() = %d, want 0", code)
	}
	if !reflect.DeepEqual(tuiPlatform, wantPlatform) || !reflect.DeepEqual(pipelinePlatform, wantPlatform) {
		t.Errorf("received TUI/pipeline platforms = %+v/%+v, want %+v", tuiPlatform, pipelinePlatform, wantPlatform)
	}
}

func TestRunWaitsForKeepaliveShutdown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{stopGate: make(chan struct{})}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	withTestEntrypointDependencies(t, dependencies)

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"mydots", "--unattended"}, "linux") }()

	<-recorder.stopStarted
	select {
	case code := <-finished:
		t.Fatalf("run returned %d before keepalive shutdown joined", code)
	default:
	}
	close(recorder.stopGate)
	if code := <-finished; code != 0 {
		t.Errorf("run() = %d, want 0", code)
	}
	if got, want := strings.Join(recorder.calls, ","), "load,elevation,keepalive,pipeline,shutdown"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestRunPipelineFailureWaitsForKeepaliveShutdown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{stopGate: make(chan struct{})}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	dependencies.runPipeline = func(context.Context, config.Config, string, platform.Platform) error {
		recorder.record("pipeline")
		return errors.New("install failed")
	}
	withTestEntrypointDependencies(t, dependencies)

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"mydots", "--unattended"}, "linux") }()

	select {
	case code := <-finished:
		t.Fatalf("run returned %d before failure shutdown started", code)
	case <-recorder.stopStarted:
	}
	select {
	case code := <-finished:
		t.Fatalf("run returned %d before failure shutdown joined", code)
	default:
	}
	close(recorder.stopGate)
	if code := <-finished; code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
	if got, want := strings.Join(recorder.calls, ","), "load,elevation,keepalive,pipeline,shutdown"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestProductionTUICompositionRoutesInstallAndBackupWithInjectedPaths(t *testing.T) {
	root := t.TempDir()
	configPath := root + "/mydots-config.json"
	backupRoot := root + "/backups"

	for _, test := range []struct {
		name     string
		platform platform.Platform
		screen   screens.Screen
		want     string
	}{
		{name: "darwin install", platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native}, screen: screens.ScreenInstall, want: "Installation"},
		{name: "linux backup", platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}, screen: screens.ScreenBackup, want: "Backups"},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := productionTUIApp(test.platform, configPath, backupRoot, func() time.Time {
				return time.Date(2026, 8, 10, 11, 0, 0, 0, time.UTC)
			}, noOpTerminal{})
			updated, _ := app.Update(screens.ChangeScreenMsg{Screen: test.screen, Payload: config.DefaultConfig()})
			if got := updated.View().Content; !strings.Contains(got, test.want) {
				t.Errorf("routed view = %q, want %q", got, test.want)
			}
		})
	}
}

func TestProductionPipelinePropagatesPlatformAndCriticalFailure(t *testing.T) {
	for _, test := range []struct {
		name     string
		platform platform.Platform
		wantErr  string
		wantOS   platform.OS
	}{
		{name: "darwin", platform: platform.Platform{OS: platform.Darwin, Variant: platform.Native}, wantErr: "critical install failed", wantOS: platform.Darwin},
		{name: "linux native", platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}, wantErr: "critical install failed", wantOS: platform.Linux},
	} {
		t.Run(test.name, func(t *testing.T) {
			var received installer.InstallContext
			err := runProductionPipelineWith(context.Background(), config.DefaultConfig(), "2026-08-10T11-00-00Z", test.platform, func(plan []installer.Module, installContext installer.InstallContext, progress chan installer.ProgressEvent) {
				received = installContext
				go func() {
					defer close(progress)
					progress <- installer.ProgressEvent{ModuleID: plan[0].ID(), Err: errors.New("critical install failed")}
				}()
			})
			if err == nil || err.Error() != test.wantErr {
				t.Errorf("pipeline error = %v, want %q", err, test.wantErr)
			}
			if received.Platform.OS != test.wantOS || received.Platform.Variant != platform.Native {
				t.Errorf("pipeline platform = %+v, want %s/native", received.Platform, test.wantOS)
			}
		})
	}
}

func TestProductionPipelineReturnsFirstNonCriticalFailureAfterDrainingProgress(t *testing.T) {
	currentPlatform := platform.Platform{OS: platform.Linux, Variant: platform.Native}
	errFirst := errors.New("first noncritical install failed")
	errSecond := errors.New("second noncritical install failed")
	var nonCritical []installer.ModuleID
	for _, module := range installer.BuildPlan(config.DefaultConfig(), currentPlatform) {
		if module.Criticality() == installer.NonCritical {
			nonCritical = append(nonCritical, module.ID())
		}
	}
	if len(nonCritical) < 2 {
		t.Fatalf("plan has %d noncritical modules, want at least 2", len(nonCritical))
	}
	err := runProductionPipelineWith(context.Background(), config.DefaultConfig(), "2026-08-10T11-00-00Z", currentPlatform, func(plan []installer.Module, _ installer.InstallContext, progress chan installer.ProgressEvent) {
		progress <- installer.ProgressEvent{ModuleID: nonCritical[0], Err: errFirst}
		progress <- installer.ProgressEvent{ModuleID: nonCritical[1], Err: errSecond}
		close(progress)
	})
	if !errors.Is(err, errFirst) {
		t.Fatalf("pipeline error = %v, want first noncritical error %v", err, errFirst)
	}
}

func TestRunPrivilegedCancelsPipelineAndFailsWhenKeepaliveBecomesUnhealthy(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	keepalive := &healthKeepalive{done: make(chan struct{}), err: errors.New("sudo refresh failed")}
	pipelineStarted := make(chan struct{})
	dependencies.startKeepalive = func(context.Context) (keepaliveHandle, error) {
		recorder.record("keepalive")
		return keepalive, nil
	}
	dependencies.runPipeline = func(ctx context.Context, _ config.Config, _ string, _ platform.Platform) error {
		recorder.record("pipeline")
		close(pipelineStarted)
		<-ctx.Done()
		return ctx.Err()
	}
	withTestEntrypointDependencies(t, dependencies)

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"mydots", "--unattended"}, "linux") }()
	<-pipelineStarted
	close(keepalive.done)
	select {
	case code := <-finished:
		if code != 1 {
			t.Errorf("run() = %d, want 1", code)
		}
	case <-time.After(time.Second):
		t.Fatal("run did not cancel the pipeline after keepalive failure")
	}
	if !strings.Contains(stderr.String(), "sudo refresh failed") {
		t.Errorf("stderr = %q, want keepalive health error", stderr.String())
	}
	if got, want := strings.Join(recorder.calls, ","), "load,elevation,keepalive,pipeline"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
	if !keepalive.stopped {
		t.Error("keepalive was not stopped after the pipeline joined")
	}
}

func TestRunPrivilegedHonorsInjectedSignalCancellation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := false
	dependencies.newContext = func() (context.Context, context.CancelFunc) {
		return ctx, func() { stopped = true }
	}
	dependencies.requestElevation = func(got context.Context) error {
		recorder.record("elevation")
		cancel()
		return got.Err()
	}
	withTestEntrypointDependencies(t, dependencies)

	if code := run([]string{"mydots", "--unattended"}, "linux"); code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
	if !stopped {
		t.Error("signal context stop was not called")
	}
	if got, want := strings.Join(recorder.calls, ","), "load,elevation"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestRunPrivilegedPropagatesInjectedSignalCancellationToPipelineAndKeepalive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := false
	dependencies.newContext = func() (context.Context, context.CancelFunc) {
		return ctx, func() { stopped = true }
	}
	dependencies.runPipeline = func(got context.Context, _ config.Config, _ string, _ platform.Platform) error {
		recorder.record("pipeline")
		cancel()
		<-got.Done()
		return got.Err()
	}
	withTestEntrypointDependencies(t, dependencies)

	if code := run([]string{"mydots", "--unattended"}, "linux"); code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
	if !stopped {
		t.Error("signal context stop was not called")
	}
	if got, want := strings.Join(recorder.calls, ","), "load,elevation,keepalive,pipeline,shutdown"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestRunPrivilegedReturnsShutdownTimeoutForNonCooperativePipeline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	signalContext, cancelSignal := context.WithCancel(context.Background())
	shutdownContext, cancelShutdown := context.WithCancel(context.Background())
	defer cancelShutdown()
	dependencies.newContext = func() (context.Context, context.CancelFunc) {
		return signalContext, func() {}
	}
	dependencies.newShutdownContext = func() (context.Context, context.CancelFunc) {
		return shutdownContext, func() {}
	}
	pipelineStarted := make(chan struct{})
	releasePipeline := make(chan struct{})
	dependencies.runPipeline = func(context.Context, config.Config, string, platform.Platform) error {
		recorder.record("pipeline")
		close(pipelineStarted)
		<-releasePipeline // Deliberately violates the cancellation contract.
		return nil
	}
	withTestEntrypointDependencies(t, dependencies)

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"mydots", "--unattended"}, "linux") }()
	<-pipelineStarted
	cancelSignal()
	select {
	case code := <-finished:
		t.Fatalf("run returned %d before the injected shutdown deadline", code)
	default:
	}
	cancelShutdown()
	if code := <-finished; code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
	if got := stderr.String(); got != "Error: shutdown timed out\n" {
		t.Errorf("stderr = %q, want shutdown timeout", got)
	}
	close(releasePipeline)
}

func TestRunPrivilegedReturnsShutdownTimeoutWhenKeepaliveFailureCannotJoinPipeline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	shutdownContext, cancelShutdown := context.WithCancel(context.Background())
	defer cancelShutdown()
	dependencies.newShutdownContext = func() (context.Context, context.CancelFunc) {
		return shutdownContext, func() {}
	}
	keepalive := &healthKeepalive{done: make(chan struct{}), err: errors.New("sudo refresh failed")}
	dependencies.startKeepalive = func(context.Context) (keepaliveHandle, error) {
		recorder.record("keepalive")
		return keepalive, nil
	}
	pipelineStarted := make(chan struct{})
	releasePipeline := make(chan struct{})
	dependencies.runPipeline = func(context.Context, config.Config, string, platform.Platform) error {
		recorder.record("pipeline")
		close(pipelineStarted)
		<-releasePipeline // Deliberately violates the cancellation contract.
		return nil
	}
	withTestEntrypointDependencies(t, dependencies)

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"mydots", "--unattended"}, "linux") }()
	<-pipelineStarted
	close(keepalive.done)
	select {
	case code := <-finished:
		t.Fatalf("run returned %d before the injected shutdown deadline", code)
	default:
	}
	cancelShutdown()
	if code := <-finished; code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
	if got := stderr.String(); got != "Error: shutdown timed out\n" {
		t.Errorf("stderr = %q, want shutdown timeout", got)
	}
	close(releasePipeline)
}

func TestRunFailurePathsPreventLaterCallsAndJoinKeepalive(t *testing.T) {
	for _, test := range []struct {
		name      string
		args      []string
		configure func(*routeRecorder, *entrypointDependencies)
		wantCalls []string
	}{
		{name: "configuration validation", args: []string{"mydots", "--unattended"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.validateConfig = func(config.Config) error { return errors.New("invalid config") }
		}, wantCalls: []string{"load"}},
		{name: "elevation", args: []string{"mydots", "--unattended"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.requestElevation = func(context.Context) error { recorder.record("elevation"); return errors.New("denied") }
		}, wantCalls: []string{"load", "elevation"}},
		{name: "keepalive start error", args: []string{"mydots", "--unattended"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.startKeepalive = func(context.Context) (keepaliveHandle, error) {
				recorder.record("keepalive")
				return nil, errors.New("unavailable")
			}
		}, wantCalls: []string{"load", "elevation", "keepalive"}},
		{name: "keepalive nil handle", args: []string{"mydots", "--unattended"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.startKeepalive = func(context.Context) (keepaliveHandle, error) { recorder.record("keepalive"); return nil, nil }
		}, wantCalls: []string{"load", "elevation", "keepalive"}},
		{name: "default save", args: []string{"mydots", "--default"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.saveConfig = func(string, config.Config) error { recorder.record("save"); return errors.New("read only") }
		}, wantCalls: []string{"save"}},
		{name: "pipeline failure joins", args: []string{"mydots", "--unattended"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.runPipeline = func(context.Context, config.Config, string, platform.Platform) error {
				recorder.record("pipeline")
				return errors.New("install failed")
			}
		}, wantCalls: []string{"load", "elevation", "keepalive", "pipeline", "shutdown"}},
		{name: "tui failure", args: []string{"mydots"}, configure: func(recorder *routeRecorder, dependencies *entrypointDependencies) {
			dependencies.runTUI = func(platform.Platform) error { recorder.record("tui"); return errors.New("terminal failed") }
		}, wantCalls: []string{"tui"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			recorder := &routeRecorder{}
			dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
			test.configure(recorder, &dependencies)
			withTestEntrypointDependencies(t, dependencies)

			if got := run(test.args, "linux"); got != 1 {
				t.Errorf("run() = %d, want 1", got)
			}
			if got := strings.Join(recorder.calls, ","); got != strings.Join(test.wantCalls, ",") {
				t.Errorf("calls = %q, want %q", got, strings.Join(test.wantCalls, ","))
			}
		})
	}
}

func TestProductionDefaultsAndAssetsAreInjected(t *testing.T) {
	dependencies := productionEntrypointDependencies()
	if got, want := dependencies.defaultConfig(), config.DefaultConfig(); !reflect.DeepEqual(got, want) {
		t.Errorf("default config = %#v, want %#v", got, want)
	}
	if dependencies.assets == nil {
		t.Fatal("production assets = nil")
	}
}

func TestRunDefaultSavesExactInjectedPathAndConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recorder := &routeRecorder{}
	dependencies := newRouteDependencies(t.TempDir(), &stdout, &stderr, recorder)
	wantPath := t.TempDir() + "/injected-default.json"
	wantConfig := config.DefaultConfig()
	var pipelinePath string
	var pipelineConfig config.Config
	dependencies.configPath = func() string { return wantPath }
	dependencies.defaultConfig = func() config.Config { return wantConfig }
	dependencies.saveConfig = func(path string, got config.Config) error {
		recorder.record("save")
		recorder.savedPath = path
		recorder.savedConfig = got
		return nil
	}
	dependencies.runPipeline = func(_ context.Context, _ config.Config, _ string, _ platform.Platform) error {
		pipelinePath = recorder.savedPath
		pipelineConfig = recorder.savedConfig
		recorder.record("pipeline")
		return nil
	}
	withTestEntrypointDependencies(t, dependencies)

	if code := run([]string{"mydots", "--default"}, "linux"); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	if recorder.savedPath != wantPath {
		t.Errorf("saved path = %q, want %q", recorder.savedPath, wantPath)
	}
	if !reflect.DeepEqual(recorder.savedConfig, wantConfig) {
		t.Errorf("saved config = %#v, want %#v", recorder.savedConfig, wantConfig)
	}
	if pipelinePath != wantPath || !reflect.DeepEqual(pipelineConfig, wantConfig) {
		t.Errorf("pipeline observed save (%q, %#v), want (%q, %#v)", pipelinePath, pipelineConfig, wantPath, wantConfig)
	}
}

type routeRecorder struct {
	mu          sync.Mutex
	calls       []string
	timestamp   string
	savedPath   string
	savedConfig config.Config
	stopGate    chan struct{}
	stopStarted chan struct{}
}

func (recorder *routeRecorder) record(call string) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.calls = append(recorder.calls, call)
}

func newRouteDependencies(path string, stdout, stderr *bytes.Buffer, recorder *routeRecorder) entrypointDependencies {
	if recorder.stopGate != nil {
		recorder.stopStarted = make(chan struct{})
	}
	return entrypointDependencies{
		stdout: stdout,
		stderr: stderr,
		configPath: func() string {
			return path + "/mydots-config.json"
		},
		loadConfig: func(string) (config.Config, error) {
			recorder.record("load")
			return config.DefaultConfig(), nil
		},
		saveConfig: func(string, config.Config) error {
			recorder.record("save")
			return nil
		},
		defaultConfig:  config.DefaultConfig,
		validateConfig: config.Validate,
		detectPlatform: func(goos string) (platform.Platform, error) {
			return platform.Platform{OS: platform.OS(goos), Variant: platform.Native}, nil
		},
		showGuide: func(fs.FS, string) {
			recorder.record("guide")
		},
		runTUI: func(platform.Platform) error {
			recorder.record("tui")
			return nil
		},
		requestElevation: func(context.Context) error {
			recorder.record("elevation")
			return nil
		},
		startKeepalive: func(context.Context) (keepaliveHandle, error) {
			recorder.record("keepalive")
			return routeKeepalive{recorder: recorder}, nil
		},
		runPipeline: func(_ context.Context, _ config.Config, timestamp string, _ platform.Platform) error {
			recorder.record("pipeline")
			recorder.timestamp = timestamp
			return nil
		},
		newContext: func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
		newShutdownContext: func() (context.Context, context.CancelFunc) {
			return context.Background(), func() {}
		},
		now: func() time.Time { return time.Date(2026, 8, 10, 12, 0, 0, 0, time.FixedZone("non-UTC", 3600)) },
	}
}

type routeKeepalive struct{ recorder *routeRecorder }

type healthKeepalive struct {
	done    chan struct{}
	err     error
	stopped bool
}

func (keepalive *healthKeepalive) Stop(context.Context) error {
	keepalive.stopped = true
	return nil
}
func (keepalive *healthKeepalive) Done() <-chan struct{} { return keepalive.done }
func (keepalive *healthKeepalive) Err() error            { return keepalive.err }

type noOpTerminal struct{}

func (noOpTerminal) Release() error { return nil }
func (noOpTerminal) Restore() error { return nil }

func (keepalive routeKeepalive) Stop(context.Context) error {
	keepalive.recorder.record("shutdown")
	if keepalive.recorder.stopGate != nil {
		close(keepalive.recorder.stopStarted)
		<-keepalive.recorder.stopGate
	}
	return nil
}

func (keepalive routeKeepalive) Done() <-chan struct{} { return nil }
func (keepalive routeKeepalive) Err() error            { return nil }

func withTestEntrypointDependencies(t *testing.T, dependencies entrypointDependencies) {
	t.Helper()
	entrypointFactoryMu.Lock()
	original := entrypointFactory
	entrypointFactory = func() entrypointDependencies { return dependencies }
	t.Cleanup(func() {
		entrypointFactory = original
		entrypointFactoryMu.Unlock()
	})
}
