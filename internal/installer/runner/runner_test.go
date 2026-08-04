package runner

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/platform"
)

// skipIfWindows skips the test on Windows because the runner uses /bin/sh
// and exec.CommandContext, which are unavailable in that environment.
// MyDots does not target Windows (redirects to WSL2).
func skipIfWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: /bin/sh not available")
	}
}

//go:embed testdata/*
var testFS embed.FS

// mockExecutor implements Executor for testing by letting tests supply
// custom Execute and LookPath functions.
type mockExecutor struct {
	executeFunc  func(ctx context.Context, name string, args ...string) ([]byte, error)
	lookPathFunc func(name string) (string, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, name, args...)
	}
	return nil, nil
}

func (m *mockExecutor) LookPath(name string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(name)
	}
	return "", nil
}

// withExecutor sets a mock executor for the test and restores the original on cleanup.
func withExecutor(t *testing.T, e Executor) {
	t.Helper()
	original := currentExecutor
	SetExecutor(e)
	t.Cleanup(func() { SetExecutor(original) })
}

// --- CommandExists ---

func TestCommandExists(t *testing.T) {
	t.Run("binary found on PATH", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		})
		if !CommandExists("git") {
			t.Error("expected true for existing binary")
		}
	})

	t.Run("binary not found", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			lookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		})
		if CommandExists("nonexistent") {
			t.Error("expected false for missing binary")
		}
	})
}

// --- CaptureOutput ---

func TestCaptureOutput(t *testing.T) {
	t.Run("success returns trimmed stdout", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("hello world\n"), nil
			},
		})
		result := CaptureOutput("echo", "hello")
		if result != "hello world" {
			t.Errorf("expected 'hello world', got: %q", result)
		}
	})

	t.Run("non-zero exit returns empty string", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error\n"), errors.New("exit status 1")
			},
		})
		result := CaptureOutput("false")
		if result != "" {
			t.Errorf("expected empty string, got: %q", result)
		}
	})
}

// --- Run ---

func TestRun(t *testing.T) {
	t.Run("writes each output line to logw", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("line1\nline2\nline3\n"), nil
			},
		})
		var buf bytes.Buffer
		err := Run(context.Background(), &buf, "echo", "hello")
		if err != nil {
			t.Fatal(err)
		}
		want := "line1\nline2\nline3\n"
		if buf.String() != want {
			t.Errorf("output:\n  got:  %q\n  want: %q", buf.String(), want)
		}
	})

	t.Run("non-zero exit returns error and writes output", func(t *testing.T) {
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				return []byte("error: something failed\n"), errors.New("exit status 1")
			},
		})
		var buf bytes.Buffer
		err := Run(context.Background(), &buf, "false")
		if err == nil {
			t.Fatal("expected error")
		}
		if buf.String() != "error: something failed\n" {
			t.Errorf("expected output to be written, got: %q", buf.String())
		}
	})
}

func TestRun_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		output     []byte
		execErr    error
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "command not found",
			execErr:    errors.New(`exec: "nonexistent": executable file not found`),
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "non-zero exit code",
			output:     []byte("error: something failed\n"),
			execErr:    errors.New("exit status 1"),
			wantOutput: "error: something failed\n",
			wantErr:    true,
		},
		{
			name:       "no output",
			output:     []byte{},
			wantOutput: "",
			wantErr:    false,
		},
		{
			name:       "partial line without newline",
			output:     []byte("incomplete"),
			wantOutput: "incomplete\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withExecutor(t, &mockExecutor{
				executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
					return tt.output, tt.execErr
				},
			})
			var buf bytes.Buffer
			err := Run(context.Background(), &buf, "somecommand")

			if buf.String() != tt.wantOutput {
				t.Errorf("output:\n  got:  %q\n  want: %q", buf.String(), tt.wantOutput)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("error:\n  got:  %v\n  want err: %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	withExecutor(t, &mockExecutor{
		executeFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err := Run(ctx, &buf, "sleep", "10")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// --- Brew ---

func TestBrew(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	withExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			capturedName = name
			capturedArgs = args
			return nil, nil
		},
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := Brew(ctx, &buf, "install", "zoxide")
	if err != nil {
		t.Fatal(err)
	}
	if capturedName != "brew" {
		t.Errorf("expected command 'brew', got: %q", capturedName)
	}
	if len(capturedArgs) != 2 || capturedArgs[0] != "install" || capturedArgs[1] != "zoxide" {
		t.Errorf("expected args ['install', 'zoxide'], got: %v", capturedArgs)
	}
}

func TestRefreshBrewAndBrewAt(t *testing.T) {
	tests := []struct {
		name     string
		platform platform.Platform
		paths    map[string]string
		wantPath string
	}{
		{"Apple Silicon prefix becomes available to dependent brew commands", platform.Platform{OS: platform.Darwin}, map[string]string{"/opt/homebrew/bin/brew": "/opt/homebrew/bin/brew"}, "/opt/homebrew/bin/brew"},
		{"Linuxbrew prefix becomes available to dependent brew commands", platform.Platform{OS: platform.Linux}, map[string]string{"/home/linuxbrew/.linuxbrew/bin/brew": "/home/linuxbrew/.linuxbrew/bin/brew"}, "/home/linuxbrew/.linuxbrew/bin/brew"},
		{"supported discovered prefix is used before fallback prefixes", platform.Platform{OS: platform.Darwin}, map[string]string{"brew": "/custom/homebrew/bin/brew"}, "/custom/homebrew/bin/brew"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var command string
			var args []string
			withExecutor(t, &mockExecutor{
				lookPathFunc: func(name string) (string, error) {
					if path, ok := tt.paths[name]; ok {
						return path, nil
					}
					return "", errors.New("not found")
				},
				executeFunc: func(_ context.Context, name string, gotArgs ...string) ([]byte, error) {
					command, args = name, gotArgs
					return nil, nil
				},
			})

			brewPath, err := RefreshBrew(tt.platform)
			if err != nil {
				t.Fatalf("RefreshBrew() = %v", err)
			}
			if brewPath != tt.wantPath {
				t.Fatalf("RefreshBrew() = %q, want %q", brewPath, tt.wantPath)
			}
			pathRef := ""
			session := WithBrewPath(context.Background(), &pathRef)
			pathRef = brewPath // Homebrew completes after the installation session starts.
			if err := Brew(session, io.Discard, "install", "zoxide"); err != nil {
				t.Fatalf("Brew() = %v", err)
			}
			if command != tt.wantPath || !slices.Equal(args, []string{"install", "zoxide"}) {
				t.Errorf("dependent command = %q %v, want %q %v", command, args, tt.wantPath, []string{"install", "zoxide"})
			}
		})
	}
}

func TestRefreshBrewReturnsErrorWhenNoSupportedPrefixExists(t *testing.T) {
	withExecutor(t, &mockExecutor{
		lookPathFunc: func(string) (string, error) { return "", errors.New("not found") },
	})
	if _, err := RefreshBrew(platform.Platform{OS: platform.Linux}); err == nil {
		t.Fatal("RefreshBrew() returned nil when brew was not discoverable")
	}
}

// --- BrewCask ---

func TestBrewCask(t *testing.T) {
	t.Run("installs cask on Darwin", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return nil, nil
			},
		})

		ctx := context.Background()
		var buf bytes.Buffer
		p := platform.Platform{OS: platform.Darwin}

		err := BrewCask(ctx, &buf, p, "ghostty")
		if err != nil {
			t.Fatal(err)
		}
		if capturedName != "brew" {
			t.Errorf("expected command 'brew', got: %q", capturedName)
		}
		if len(capturedArgs) != 3 || capturedArgs[0] != "install" || capturedArgs[1] != "--cask" || capturedArgs[2] != "ghostty" {
			t.Errorf("expected args ['install', '--cask', 'ghostty'], got: %v", capturedArgs)
		}
	})

	t.Run("uses the session brew path and preserves cask argv", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				capturedName = name
				capturedArgs = args
				return nil, nil
			},
		})

		brewPath := "/opt/homebrew/bin/brew"
		ctx := WithBrewPath(context.Background(), &brewPath)
		if err := BrewCask(ctx, io.Discard, platform.Platform{OS: platform.Darwin}, "font-jetbrains-mono-nerd-font"); err != nil {
			t.Fatalf("BrewCask() = %v", err)
		}
		if capturedName != brewPath || !slices.Equal(capturedArgs, []string{"install", "--cask", "font-jetbrains-mono-nerd-font"}) {
			t.Errorf("command = %q %v, want %q %v", capturedName, capturedArgs, brewPath, []string{"install", "--cask", "font-jetbrains-mono-nerd-font"})
		}
	})

	t.Run("propagates cask failure from the session brew path", func(t *testing.T) {
		caskErr := errors.New("cask failed")
		withExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name != "/opt/homebrew/bin/brew" || !slices.Equal(args, []string{"install", "--cask", "ghostty"}) {
					t.Errorf("command = %q %v, want %q %v", name, args, "/opt/homebrew/bin/brew", []string{"install", "--cask", "ghostty"})
				}
				return nil, caskErr
			},
		})

		brewPath := "/opt/homebrew/bin/brew"
		ctx := WithBrewPath(context.Background(), &brewPath)
		if err := BrewCask(ctx, io.Discard, platform.Platform{OS: platform.Darwin}, "ghostty"); !errors.Is(err, caskErr) {
			t.Errorf("BrewCask() = %v, want %v", err, caskErr)
		}
	})

	t.Run("returns ErrNotDarwin on Linux", func(t *testing.T) {
		withExecutor(t, &mockExecutor{})

		ctx := context.Background()
		var buf bytes.Buffer
		p := platform.Platform{OS: platform.Linux}

		err := BrewCask(ctx, &buf, p, "ghostty")
		if !errors.Is(err, ErrNotDarwin) {
			t.Errorf("expected ErrNotDarwin, got: %v", err)
		}
	})
}

// --- BrewTap ---

func TestBrewTap(t *testing.T) {
	var calls []string
	withExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			calls = append(calls, strings.Join(append([]string{name}, args...), " "))
			return nil, nil
		},
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := BrewTap(ctx, &buf, "homebrew/cask", "ghostty")
	if err != nil {
		t.Fatal(err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 executor calls, got: %d", len(calls))
	}

	expectedCmds := []string{"brew tap homebrew/cask", "brew install ghostty"}
	for i, call := range calls {
		if call != expectedCmds[i] {
			t.Errorf("call %d: expected %q, got %q", i, expectedCmds[i], call)
		}
	}
}

func TestBrewTapUsesSessionPathAndPropagatesTapFailure(t *testing.T) {
	tapErr := errors.New("tap failed")
	var calls []struct {
		name string
		args []string
	}
	withExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			calls = append(calls, struct {
				name string
				args []string
			}{name: name, args: args})
			return nil, tapErr
		},
	})

	brewPath := "/opt/homebrew/bin/brew"
	ctx := WithBrewPath(context.Background(), &brewPath)
	err := BrewTap(ctx, io.Discard, "Gentleman-Programming/homebrew-tap", "gentle-ai")
	if !errors.Is(err, tapErr) {
		t.Fatalf("BrewTap() = %v, want %v", err, tapErr)
	}
	if len(calls) != 1 || calls[0].name != brewPath || !slices.Equal(calls[0].args, []string{"tap", "Gentleman-Programming/homebrew-tap"}) {
		t.Errorf("calls = %#v, want one %q %v call", calls, brewPath, []string{"tap", "Gentleman-Programming/homebrew-tap"})
	}
}

// --- Script ---

func TestScript(t *testing.T) {
	t.Run("runs script and writes output to logw", func(t *testing.T) {
		skipIfWindows(t)
		var buf bytes.Buffer
		ctx := context.Background()

		err := Script(ctx, &buf, testFS, "testdata/hello.sh", nil)
		if err != nil {
			t.Fatal(err)
		}

		output := buf.String()
		if !strings.Contains(output, "hello from script") {
			t.Errorf("expected output to contain 'hello from script', got: %q", output)
		}
	})

	t.Run("temp file has 0700 permissions and is removed", func(t *testing.T) {
		skipIfWindows(t)
		var capturedPath string
		var permsOK bool
		originalCreate := createTempFile
		originalRemove := removeFile

		createTempFile = func(pattern string) (*os.File, error) {
			f, err := os.CreateTemp("", pattern)
			if err == nil {
				capturedPath = f.Name()
			}
			return f, err
		}
		removeFile = func(path string) error {
			info, err := os.Stat(path)
			if err == nil && info.Mode().Perm() == 0700 {
				permsOK = true
			}
			return os.Remove(path)
		}
		t.Cleanup(func() {
			createTempFile = originalCreate
			removeFile = originalRemove
		})

		var buf bytes.Buffer
		ctx := context.Background()
		err := Script(ctx, &buf, testFS, "testdata/hello.sh", nil)
		if err != nil {
			t.Fatal(err)
		}

		if !permsOK {
			t.Error("temp file did not have 0700 permissions before removal")
		}
		if capturedPath != "" {
			if _, statErr := os.Stat(capturedPath); !os.IsNotExist(statErr) {
				t.Error("temp file was not removed after script execution")
			}
		}
	})

	t.Run("env vars are merged with new values taking precedence", func(t *testing.T) {
		skipIfWindows(t)
		var buf bytes.Buffer
		ctx := context.Background()

		err := Script(ctx, &buf, testFS, "testdata/env_test.sh", map[string]string{"MYVAR": "testvalue"})
		if err != nil {
			t.Fatal(err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.HasSuffix(output, "MYVAR=testvalue") {
			t.Errorf("expected output to end with 'MYVAR=testvalue', got: %q", output)
		}
	})
}

func TestShippedInstallerScriptsRunUnderPOSIXSh(t *testing.T) {
	binDir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf '%s\\n' 'echo downloaded-installer-ran'\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n    printf 'GNU bash, version 5.2.0\\n'\n    exit 0\nfi\neval \"$(cat \"$1\")\"\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	scripts := []string{
		"assets/scripts/homebrew-install.sh",
		"assets/scripts/omz-install.sh",
		"assets/scripts/sdkman-install.sh",
	}
	for _, script := range scripts {
		t.Run(script, func(t *testing.T) {
			log, err := runPOSIXScript(t, script, binDir)
			if err != nil {
				t.Fatalf("script %q = %v, want nil; output: %s", script, err, log)
			}
			if !strings.Contains(log, "installation complete") {
				t.Errorf("script %q output = %q, want completion message", script, log)
			}
			if !strings.Contains(log, "downloaded-installer-ran") {
				t.Errorf("script %q output = %q, want downloaded installer execution", script, log)
			}
		})
	}
}

func TestShippedInstallerScriptsReturnDownloadFailure(t *testing.T) {
	binDir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nexit 22\n")

	scripts := []string{
		"assets/scripts/homebrew-install.sh",
		"assets/scripts/omz-install.sh",
		"assets/scripts/sdkman-install.sh",
	}
	for _, script := range scripts {
		t.Run(script, func(t *testing.T) {
			log, err := runPOSIXScript(t, script, binDir)
			if err == nil {
				t.Fatalf("script %q returned nil after the download failed", script)
			}
			if strings.Contains(log, "installation complete") {
				t.Errorf("script %q output = %q, must not report successful installation", script, log)
			}
		})
	}
}

func TestHomebrewInstallScriptValidatesCurlBeforeCreatingTemporaryInstaller(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\nprintf '%s/mydots-homebrew-installer' \"$TMPDIR\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/homebrew-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      tempDir,
	})
	if err == nil {
		t.Fatal("homebrew-install.sh returned nil without curl")
	}
	if got, want := log, "Error: curl is required to install Homebrew.\n"; got != want {
		t.Errorf("homebrew-install.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("mktemp ran without curl; stat error = %v", err)
	}
}

func TestHomebrewInstallScriptValidatesMktempBeforeDownloading(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/homebrew-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      t.TempDir(),
	})
	if err == nil {
		t.Fatal("homebrew-install.sh returned nil without mktemp")
	}
	if got, want := log, "Error: mktemp is required to install Homebrew.\n"; got != want {
		t.Errorf("homebrew-install.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("curl ran without mktemp; stat error = %v", err)
	}
}

func TestHomebrewInstallScriptRunsDownloadedInstallerAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/homebrew-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fsSL\" ] || exit 11\n[ \"$2\" = \"https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh\" ] || exit 12\nprintf 'curl:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\nprintf 'printf upstream-installer-ran\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/homebrew-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      tempDir,
	})
	if err != nil {
		t.Fatalf("homebrew-install.sh = %v, output: %s", err, log)
	}
	if !strings.Contains(log, "upstream-installer-ran") || !strings.Contains(log, "Homebrew installation complete") {
		t.Errorf("homebrew-install.sh output = %q, want installer and completion output", log)
	}
	installerPath := filepath.Join(tempDir, "homebrew-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 2 || entries[0] != "curl:https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh" || !strings.HasPrefix(entries[1], "rm:") {
		t.Errorf("commands = %q, want curl followed by cleanup", commands)
	}
}

func TestHomebrewInstallScriptPropagatesInstallerFailureAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/homebrew-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nprintf 'exit 43\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/homebrew-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      tempDir,
	})
	if err == nil {
		t.Fatal("homebrew-install.sh returned nil after the installer failed")
	}
	if strings.Contains(log, "Homebrew installation complete") {
		t.Errorf("homebrew-install.sh output = %q, must not report success", log)
	}
	installerPath := filepath.Join(tempDir, "homebrew-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed after installer failure; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 2 || entries[0] != "curl" || !strings.HasPrefix(entries[1], "rm:") {
		t.Errorf("commands = %q, want failed download followed by cleanup", commands)
	}
}

func TestOhMyZshInstallScriptRunsOfficialInstallerUnattendedAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/omz-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fsSL\" ] || exit 11\n[ \"$2\" = \"https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh\" ] || exit 12\nprintf 'curl:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\nprintf 'echo upstream-installer-ran\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh:%s:%s:%s\\n' \"$RUNZSH\" \"$CHSH\" \"$1\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/omz-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      tempDir,
	})
	if err != nil {
		t.Fatalf("omz-install.sh = %v, output: %s", err, log)
	}
	if !strings.Contains(log, "upstream-installer-ran") || !strings.Contains(log, "Oh My Zsh installation complete") {
		t.Errorf("omz-install.sh output = %q, want upstream installer and completion output", log)
	}
	installerPath := filepath.Join(tempDir, "omz-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 3 || entries[0] != "curl:https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh" || !strings.HasPrefix(entries[1], "sh:no:no:") || !strings.HasPrefix(entries[2], "rm:") {
		t.Errorf("commands = %q, want official curl, unattended sh, and cleanup", commands)
	}
}

func TestOhMyZshInstallScriptSkipsWhenAlreadyInstalled(t *testing.T) {
	binDir := t.TempDir()
	homeDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	if err := os.Mkdir(filepath.Join(homeDir, ".oh-my-zsh"), 0755); err != nil {
		t.Fatalf("Mkdir existing Oh My Zsh directory: %v", err)
	}
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh\\n' >> \"$COMMAND_LOG\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/omz-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        homeDir,
		"TMPDIR":      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("omz-install.sh = %v, output: %s", err, log)
	}
	if got, want := log, "==> Oh My Zsh is already installed; skipping.\n"; got != want {
		t.Errorf("omz-install.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("installer command ran for existing Oh My Zsh directory; stat error = %v", err)
	}
}

func TestOhMyZshInstallScriptValidatesDependenciesBeforeCreatingTemporaryInstaller(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh\\n' >> \"$COMMAND_LOG\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/omz-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      t.TempDir(),
	})
	if err == nil {
		t.Fatal("omz-install.sh returned nil without curl")
	}
	if got, want := log, "Error: curl is required to install Oh My Zsh.\n"; got != want {
		t.Errorf("omz-install.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("mktemp or sh ran without curl; stat error = %v", err)
	}
}

func TestOhMyZshInstallScriptValidatesRemainingDependenciesBeforeCreatingTemporaryInstaller(t *testing.T) {
	tests := []struct {
		name         string
		installStubs func(t *testing.T, binDir, logFile string)
		wantOutput   string
	}{
		{
			name: "missing mktemp",
			installStubs: func(t *testing.T, binDir, logFile string) {
				writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
				writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh\\n' >> \"$COMMAND_LOG\"\n")
			},
			wantOutput: "Error: mktemp is required to install Oh My Zsh.\n",
		},
		{
			name: "missing sh",
			installStubs: func(t *testing.T, binDir, logFile string) {
				writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
				writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\n")
			},
			wantOutput: "Error: sh is required to install Oh My Zsh.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			logFile := filepath.Join(t.TempDir(), "commands.log")
			tt.installStubs(t, binDir, logFile)

			log, err := runPOSIXScriptWithEnv(t, "assets/scripts/omz-install.sh", binDir, map[string]string{
				"COMMAND_LOG": logFile,
				"HOME":        t.TempDir(),
				"TMPDIR":      t.TempDir(),
			})
			if err == nil {
				t.Fatalf("omz-install.sh returned nil when %s", tt.name)
			}
			if log != tt.wantOutput {
				t.Errorf("omz-install.sh output = %q, want %q", log, tt.wantOutput)
			}
			if _, err := os.Stat(logFile); !os.IsNotExist(err) {
				t.Errorf("installer command ran when %s; stat error = %v", tt.name, err)
			}
		})
	}
}

func TestOhMyZshInstallScriptPropagatesInstallerFailureAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/omz-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nprintf 'exit 43\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh:%s:%s\\n' \"$RUNZSH\" \"$CHSH\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/omz-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        t.TempDir(),
		"TMPDIR":      tempDir,
	})
	if err == nil {
		t.Fatal("omz-install.sh returned nil after the installer failed")
	}
	if strings.Contains(log, "Oh My Zsh installation complete") {
		t.Errorf("omz-install.sh output = %q, must not report success", log)
	}
	installerPath := filepath.Join(tempDir, "omz-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed after installer failure; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 3 || entries[0] != "curl" || entries[1] != "sh:no:no" || !strings.HasPrefix(entries[2], "rm:") {
		t.Errorf("commands = %q, want failed installer execution followed by cleanup", commands)
	}
}

func TestSdkmanInstallScriptRunsOfficialInstallerUnattendedAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	homeDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/sdkman-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fsSL\" ] || exit 11\n[ \"$2\" = \"https://get.sdkman.io\" ] || exit 12\nprintf 'curl:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\nprintf 'echo upstream-installer-ran\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n    printf 'GNU bash, version 5.2.0\\n'\n    exit 0\nfi\nprintf 'bash:%s:%s\\n' \"$SDKMAN_DIR\" \"$1\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/sdkman-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        homeDir,
		"TMPDIR":      tempDir,
	})
	if err != nil {
		t.Fatalf("sdkman-install.sh = %v, output: %s", err, log)
	}
	if !strings.Contains(log, "upstream-installer-ran") || !strings.Contains(log, "SDKMAN installation complete") {
		t.Errorf("sdkman-install.sh output = %q, want installer and completion output", log)
	}
	installerPath := filepath.Join(tempDir, "sdkman-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 3 || entries[0] != "curl:https://get.sdkman.io" || !strings.HasPrefix(entries[1], "bash:") || !strings.Contains(entries[1], ".sdkman:") || !strings.HasPrefix(entries[2], "rm:") {
		t.Errorf("commands = %q, want official curl, SDKMAN_DIR bash, and cleanup", commands)
	}
}

func TestSdkmanInstallScriptSkipsWhenAlreadyInstalled(t *testing.T) {
	binDir := t.TempDir()
	homeDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	initFile := filepath.Join(homeDir, ".sdkman", "bin", "sdkman-init.sh")
	if err := os.MkdirAll(filepath.Dir(initFile), 0755); err != nil {
		t.Fatalf("MkdirAll SDKMAN directory: %v", err)
	}
	if err := os.WriteFile(initFile, []byte("# SDKMAN"), 0600); err != nil {
		t.Fatalf("WriteFile SDKMAN init file: %v", err)
	}
	for _, command := range []string{"curl", "mktemp", "bash"} {
		writeScriptStub(t, filepath.Join(binDir, command), "#!/bin/sh\nprintf '"+command+"\\n' >> \"$COMMAND_LOG\"\n")
	}

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/sdkman-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        homeDir,
		"TMPDIR":      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("sdkman-install.sh = %v, output: %s", err, log)
	}
	if got, want := log, "==> SDKMAN is already installed; skipping.\n"; got != want {
		t.Errorf("sdkman-install.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("installer command ran for existing SDKMAN init file; stat error = %v", err)
	}
}

func TestSdkmanInstallScriptValidatesDependenciesBeforeCreatingTemporaryInstaller(t *testing.T) {
	tests := []struct {
		name         string
		installStubs func(t *testing.T, binDir, logFile string)
		wantOutput   string
	}{
		{
			name: "missing curl",
			installStubs: func(t *testing.T, binDir, logFile string) {
				writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\n")
				writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash\\n' >> \"$COMMAND_LOG\"\n")
			},
			wantOutput: "Error: curl is required to install SDKMAN.\n",
		},
		{
			name: "missing mktemp",
			installStubs: func(t *testing.T, binDir, logFile string) {
				writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
				writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash\\n' >> \"$COMMAND_LOG\"\n")
			},
			wantOutput: "Error: mktemp is required to install SDKMAN.\n",
		},
		{
			name: "missing bash",
			installStubs: func(t *testing.T, binDir, logFile string) {
				writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
				writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf 'mktemp\\n' >> \"$COMMAND_LOG\"\n")
			},
			wantOutput: "Error: bash is required to install SDKMAN.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			logFile := filepath.Join(t.TempDir(), "commands.log")
			tt.installStubs(t, binDir, logFile)

			log, err := runPOSIXScriptWithEnv(t, "assets/scripts/sdkman-install.sh", binDir, map[string]string{
				"COMMAND_LOG": logFile,
				"HOME":        t.TempDir(),
				"TMPDIR":      t.TempDir(),
			})
			if err == nil {
				t.Fatalf("sdkman-install.sh returned nil when %s", tt.name)
			}
			if log != tt.wantOutput {
				t.Errorf("sdkman-install.sh output = %q, want %q", log, tt.wantOutput)
			}
			if _, err := os.Stat(logFile); !os.IsNotExist(err) {
				t.Errorf("installer command ran when %s; stat error = %v", tt.name, err)
			}
		})
	}
}

func TestSdkmanInstallScriptRequiresBashFourOrNewerBeforeDownloading(t *testing.T) {
	tests := []struct {
		name          string
		bashVersion   string
		wantErr       bool
		wantOutput    string
		wantDownloads bool
	}{
		{
			name:        "rejects Bash 3",
			bashVersion: "GNU bash, version 3.2.57(1)-release (x86_64-apple-darwin23)\n",
			wantErr:     true,
			wantOutput:  "Error: Bash 4 or newer is required to install SDKMAN; found Bash 3. Upgrade Bash and ensure it is first on PATH.\n",
		},
		{
			name:        "rejects invalid version",
			bashVersion: "not a Bash version\n",
			wantErr:     true,
			wantOutput:  "Error: unable to determine the Bash version. SDKMAN requires Bash 4 or newer; install Bash 4+ and ensure it is first on PATH.\n",
		},
		{
			name:          "accepts Bash 4",
			bashVersion:   "GNU bash, version 4.4.23(1)-release (x86_64-pc-linux-gnu)\n",
			wantOutput:    "==> Downloading SDKMAN install script...\nupstream-installer-ran\n==> SDKMAN installation complete\n",
			wantDownloads: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			logFile := filepath.Join(t.TempDir(), "commands.log")
			writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/sdkman-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
			writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nprintf 'echo upstream-installer-ran\\n'\n")
			writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n    printf '%s' \"$BASH_VERSION_OUTPUT\"\n    exit 0\nfi\nprintf 'bash:%s\\n' \"$1\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
			writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

			log, err := runPOSIXScriptWithEnv(t, "assets/scripts/sdkman-install.sh", binDir, map[string]string{
				"BASH_VERSION_OUTPUT": tt.bashVersion,
				"COMMAND_LOG":         logFile,
				"HOME":                t.TempDir(),
				"TMPDIR":              t.TempDir(),
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("sdkman-install.sh error = %v, want error = %t; output: %s", err, tt.wantErr, log)
			}
			if got := log; got != tt.wantOutput {
				t.Errorf("sdkman-install.sh output = %q, want %q", got, tt.wantOutput)
			}

			commands, err := os.ReadFile(logFile)
			if !tt.wantDownloads && os.IsNotExist(err) {
				return
			}
			if err != nil {
				t.Fatalf("ReadFile(command log): %v", err)
			}
			if got := strings.Contains(string(commands), "curl\n"); got != tt.wantDownloads {
				t.Errorf("download invoked = %t, want %t; commands = %q", got, tt.wantDownloads, commands)
			}
		})
	}
}

func TestSdkmanInstallScriptPropagatesInstallerFailureAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	homeDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/sdkman-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nprintf 'exit 43\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n    printf 'GNU bash, version 5.2.0\\n'\n    exit 0\nfi\nprintf 'bash:%s\\n' \"$SDKMAN_DIR\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/sdkman-install.sh", binDir, map[string]string{
		"COMMAND_LOG": logFile,
		"HOME":        homeDir,
		"TMPDIR":      tempDir,
	})
	if err == nil {
		t.Fatal("sdkman-install.sh returned nil after the installer failed")
	}
	if strings.Contains(log, "SDKMAN installation complete") {
		t.Errorf("sdkman-install.sh output = %q, must not report success", log)
	}
	installerPath := filepath.Join(tempDir, "sdkman-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed after installer failure; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 3 || entries[0] != "curl" || !strings.HasPrefix(entries[1], "bash:") || !strings.Contains(entries[1], ".sdkman") || !strings.HasPrefix(entries[2], "rm:") {
		t.Errorf("commands = %q, want failed installer execution followed by cleanup", commands)
	}
}

func TestDockerLinuxScriptInstallsThroughOfficialAptInstallerAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	tempDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/docker-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fsSL\" ] || exit 11\n[ \"$2\" = \"https://get.docker.com\" ] || exit 12\nprintf 'curl:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\nprintf 'printf \\\"apt-repository-setup\\\\napt-install:docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin\\\\n\\\"\\n'\n")
	writeScriptStub(t, filepath.Join(binDir, "sudo"), "#!/bin/sh\nprintf 'sudo:%s\\n' \"$*\" >> \"$COMMAND_LOG\"\nexec \"$@\"\n")
	writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nprintf 'sh:%s\\n' \"$1\" >> \"$COMMAND_LOG\"\nexec /bin/sh \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "id"), "#!/bin/sh\nprintf 'id:%s:%s\\n' \"$1\" \"$2\" >> \"$COMMAND_LOG\"\nprintf '%s\\n' \"$GROUPS_OUTPUT\"\n")
	writeScriptStub(t, filepath.Join(binDir, "getent"), "#!/bin/sh\nexit 0\n")
	writeScriptStub(t, filepath.Join(binDir, "usermod"), "#!/bin/sh\nprintf 'usermod:%s\\n' \"$*\" >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\nprintf 'rm:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n/bin/rm \"$@\"\n")

	log, err := runPOSIXScriptWithEnv(t, "assets/scripts/docker-linux.sh", binDir, map[string]string{
		"COMMAND_LOG":   logFile,
		"GROUPS_OUTPUT": "users wheel",
		"TMPDIR":        tempDir,
		"USER":          "alice",
	})
	if err != nil {
		t.Fatalf("docker-linux.sh = %v, output: %s", err, log)
	}
	if !strings.Contains(log, "apt-repository-setup") || !strings.Contains(log, "apt-install:docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin") || !strings.Contains(log, "Docker installation complete") {
		t.Errorf("docker-linux.sh output = %q, want official apt setup, install sequence, and completion", log)
	}
	installerPath := filepath.Join(tempDir, "docker-installer")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed; stat error = %v", installerPath, err)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	for _, want := range []string{
		"curl:https://get.docker.com",
		"sudo:sh ",
		"id:-nG:alice",
		"sudo:usermod -aG docker alice",
		"rm:",
	} {
		if !strings.Contains(string(commands), want) {
			t.Errorf("commands = %q, want %q", commands, want)
		}
	}
}

func TestDockerLinuxScriptGroupSafetyContracts(t *testing.T) {
	tests := []struct {
		name            string
		groups          string
		idStatus        string
		usermodStatus   string
		user            string
		wantErr         bool
		wantSuccess     bool
		wantUsermod     bool
		wantPackageWork bool
		wantOutput      string
		dockerInstalled bool
	}{
		{name: "pre-existing Docker reconciles missing group membership without package work", groups: "users wheel", user: "alice", wantSuccess: true, wantUsermod: true, dockerInstalled: true},
		{name: "pre-existing Docker with exact group membership skips usermod and package work", groups: "notdocker docker-foo docker", user: "alice", wantSuccess: true, dockerInstalled: true},
		{name: "pre-existing Docker probe failure stops before group mutation", idStatus: "19", user: "alice", wantErr: true, dockerInstalled: true},
		{name: "pre-existing Docker non-member propagates usermod failure without false success", groups: "users wheel", usermodStatus: "23", user: "alice", wantErr: true, wantUsermod: true, dockerInstalled: true},
		{name: "fresh install exact docker token skips redundant group mutation", groups: "notdocker docker-foo docker", user: "alice", wantSuccess: true, wantPackageWork: true},
		{name: "missing USER fails before any mutation even with pre-existing Docker", groups: "users", wantErr: true, wantOutput: "Error: USER must be set to configure Docker group membership.\n", dockerInstalled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			logFile := filepath.Join(t.TempDir(), "commands.log")
			if tt.dockerInstalled {
				writeScriptStub(t, filepath.Join(binDir, "docker"), "#!/bin/sh\nexit 0\n")
			}
			writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\npath=\"$TMPDIR/docker-installer\"\n: > \"$path\"\nprintf '%s\\n' \"$path\"\n")
			writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nprintf 'exit 0\\n'\n")
			writeScriptStub(t, filepath.Join(binDir, "sudo"), "#!/bin/sh\nprintf 'sudo:%s\\n' \"$*\" >> \"$COMMAND_LOG\"\nexec \"$@\"\n")
			writeScriptStub(t, filepath.Join(binDir, "sh"), "#!/bin/sh\nexec /bin/sh \"$1\"\n")
			writeScriptStub(t, filepath.Join(binDir, "id"), "#!/bin/sh\nprintf 'id:%s:%s\\n' \"$1\" \"$2\" >> \"$COMMAND_LOG\"\n[ \"${ID_STATUS:-0}\" -eq 0 ] || exit \"$ID_STATUS\"\nprintf '%s\\n' \"$GROUPS_OUTPUT\"\n")
			writeScriptStub(t, filepath.Join(binDir, "getent"), "#!/bin/sh\nexit 0\n")
			writeScriptStub(t, filepath.Join(binDir, "usermod"), "#!/bin/sh\nprintf 'usermod:%s\\n' \"$*\" >> \"$COMMAND_LOG\"\nexit \"${USERMOD_STATUS:-0}\"\n")
			writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\n/bin/rm \"$@\"\n")

			log, err := runPOSIXScriptWithEnv(t, "assets/scripts/docker-linux.sh", binDir, map[string]string{
				"COMMAND_LOG":    logFile,
				"GROUPS_OUTPUT":  tt.groups,
				"ID_STATUS":      tt.idStatus,
				"TMPDIR":         t.TempDir(),
				"USER":           tt.user,
				"USERMOD_STATUS": tt.usermodStatus,
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("docker-linux.sh error = %v, want error = %t; output: %s", err, tt.wantErr, log)
			}
			if tt.wantOutput != "" && log != tt.wantOutput {
				t.Errorf("docker-linux.sh output = %q, want %q", log, tt.wantOutput)
			}
			if strings.Contains(log, "Docker installation complete") != tt.wantSuccess {
				t.Errorf("success output = %t, want %t; output: %q", strings.Contains(log, "Docker installation complete"), tt.wantSuccess, log)
			}
			commands, readErr := os.ReadFile(logFile)
			if readErr != nil && tt.user != "" {
				t.Fatalf("ReadFile(command log): %v", readErr)
			}
			if tt.user == "" && !os.IsNotExist(readErr) {
				t.Errorf("commands ran before USER validation: %q", commands)
			}
			if tt.user != "" && !strings.Contains(string(commands), "id:-nG:"+tt.user) {
				t.Errorf("commands = %q, want group probe for %q", commands, tt.user)
			}
			packageWorkRan := strings.Contains(string(commands), "curl\n") || strings.Contains(string(commands), "sudo:sh ")
			if packageWorkRan != tt.wantPackageWork {
				t.Errorf("package work ran = %t, want %t; commands = %q", packageWorkRan, tt.wantPackageWork, commands)
			}
			if strings.Contains(string(commands), "usermod:") != tt.wantUsermod {
				t.Errorf("usermod invoked = %t, want %t; commands = %q", strings.Contains(string(commands), "usermod:"), tt.wantUsermod, commands)
			}
			if tt.wantUsermod && !strings.Contains(string(commands), "sudo:usermod -aG docker "+tt.user) {
				t.Errorf("commands = %q, want usermod for %q", commands, tt.user)
			}
		})
	}
}

func TestShippedCavemanScriptReturnsDownloadFailure(t *testing.T) {
	binDir := t.TempDir()
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nexit 22\n")

	log, err := runPOSIXScript(t, "assets/scripts/caveman-install.sh", binDir)
	if err == nil {
		t.Fatal("caveman script returned nil after the download failed")
	}
	if strings.Contains(log, "Caveman installation complete") {
		t.Errorf("caveman script output = %q, must not report successful installation", log)
	}
}

func TestShippedCavemanScriptRunsDownloadedInstallerWithOpenclawOnly(t *testing.T) {
	binDir := t.TempDir()
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\ncat <<'EOF'\n#!/bin/sh\nif [ \"$1\" = \"--only\" ] && [ \"$2\" = \"openclaw\" ]; then\n    echo caveman-installer-ran\nelse\n    exit 9\nfi\nEOF\n")

	log, err := runPOSIXScript(t, "assets/scripts/caveman-install.sh", binDir)
	if err != nil {
		t.Fatalf("caveman script = %v, want nil; output: %s", err, log)
	}
	if !strings.Contains(log, "caveman-installer-ran") {
		t.Errorf("caveman script output = %q, want downloaded installer output", log)
	}
	if !strings.Contains(log, "Caveman installation complete") {
		t.Errorf("caveman script output = %q, want completion message", log)
	}
}

func TestShippedGhosttyLinuxScriptDownloadsAndRunsUbuntuInstaller(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fsSL\" ] || exit 11\n[ \"$2\" = \"-o\" ] || exit 12\n[ \"$4\" = \"https://raw.githubusercontent.com/mkasberg/ghostty-ubuntu/HEAD/install.sh\" ] || exit 13\nprintf '#!/bin/sh\\necho upstream-installer-ran\\n' > \"$3\"\nprintf 'curl:%s\\n' \"$3\" >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash:%s\\n' \"$1\" >> \"$COMMAND_LOG\"\nexec sh \"$1\"\n")
	t.Setenv("COMMAND_LOG", logFile)

	log, err := runPOSIXScript(t, "assets/scripts/ghostty-linux.sh", binDir)
	if err != nil {
		t.Fatalf("ghostty-linux.sh = %v, output: %s", err, log)
	}
	if !strings.Contains(log, "upstream-installer-ran") || !strings.Contains(log, "Ghostty installation complete") {
		t.Errorf("ghostty-linux.sh output = %q, want installer and completion output", log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 2 || !strings.HasPrefix(entries[0], "curl:") || entries[1] != "bash:"+strings.TrimPrefix(entries[0], "curl:") {
		t.Errorf("commands = %q, want curl followed by bash for the same temporary installer", commands)
		return
	}
	installerPath := strings.TrimPrefix(entries[0], "curl:")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed; stat error = %v", installerPath, err)
	}
}

func TestShippedGhosttyLinuxScriptStopsWhenDownloadFails(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl:%s\\n' \"$3\" >> \"$COMMAND_LOG\"\nexit 22\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash\\n' >> \"$COMMAND_LOG\"\n")
	t.Setenv("COMMAND_LOG", logFile)

	log, err := runPOSIXScript(t, "assets/scripts/ghostty-linux.sh", binDir)
	if err == nil {
		t.Fatal("ghostty-linux.sh returned nil after the download failed")
	}
	if strings.Contains(log, "Ghostty installation complete") {
		t.Errorf("ghostty-linux.sh output = %q, must not report success", log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	installerPath := strings.TrimPrefix(strings.TrimSpace(string(commands)), "curl:")
	if installerPath == strings.TrimSpace(string(commands)) {
		t.Fatalf("commands = %q, want curl temporary installer path", commands)
	}
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed after download failure; stat error = %v", installerPath, err)
	}
}

func TestShippedGhosttyLinuxScriptValidatesRequiredToolsBeforeDownloading(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf '%s\\n' \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash\\n' >> \"$COMMAND_LOG\"\n")
	t.Setenv("COMMAND_LOG", logFile)

	log, err := runPOSIXScriptWithPath(t, "assets/scripts/ghostty-linux.sh", binDir)
	if err == nil {
		t.Fatal("ghostty-linux.sh returned nil without curl")
	}
	if !strings.Contains(log, "curl is required") {
		t.Errorf("ghostty-linux.sh output = %q, want missing curl error", log)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("installer ran without curl; stat error = %v", err)
	}
}

func TestShippedGhosttyLinuxScriptPropagatesInstallerFailureAndCleansUp(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf '#!/bin/sh\\nexit 43\\n' > \"$3\"\nprintf 'curl:%s\\n' \"$3\" >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\nprintf 'bash:%s\\n' \"$1\" >> \"$COMMAND_LOG\"\nexit 43\n")
	t.Setenv("COMMAND_LOG", logFile)

	log, err := runPOSIXScript(t, "assets/scripts/ghostty-linux.sh", binDir)
	if err == nil {
		t.Fatal("ghostty-linux.sh returned nil after the installer failed")
	}
	if strings.Contains(log, "Ghostty installation complete") {
		t.Errorf("ghostty-linux.sh output = %q, must not report success", log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	entries := strings.Split(strings.TrimSpace(string(commands)), "\n")
	if len(entries) != 2 || !strings.HasPrefix(entries[0], "curl:") || entries[1] != "bash:"+strings.TrimPrefix(entries[0], "curl:") {
		t.Errorf("commands = %q, want curl followed by bash for the same temporary installer", commands)
		return
	}
	installerPath := strings.TrimPrefix(entries[0], "curl:")
	if _, err := os.Stat(installerPath); !os.IsNotExist(err) {
		t.Errorf("temporary installer %q was not removed after installer failure; stat error = %v", installerPath, err)
	}
}

func TestShippedFontLinuxScriptInstallsAndRefreshesFontCache(t *testing.T) {
	binDir := t.TempDir()
	fontHome := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\n[ \"$1\" = \"-fL\" ] || exit 11\n[ \"$2\" = \"-o\" ] || exit 12\nprintf 'archive' > \"$3\"\nprintf 'curl:%s\\n' \"$4\" >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "unzip"), "#!/bin/sh\n[ \"$1\" = \"-o\" ] || exit 21\n[ \"$3\" = \"-d\" ] || exit 22\nmkdir -p \"$4\"\nprintf 'font' > \"$4/font.ttf\"\nprintf 'unzip:%s:%s\\n' \"$2\" \"$4\" >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "fc-cache"), "#!/bin/sh\n[ \"$1\" = \"-fv\" ] || exit 31\nprintf 'cache:%s\\n' \"$2\" >> \"$COMMAND_LOG\"\n")
	t.Setenv("HOME", fontHome)
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "FiraCode.zip")

	log, err := runPOSIXScript(t, "assets/scripts/font-linux.sh", binDir)
	if err != nil {
		t.Fatalf("font-linux.sh = %v, output: %s", err, log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	fontDir := filepath.Join(fontHome, ".local", "share", "fonts")
	if !strings.Contains(string(commands), "curl:https://github.com/ryanoasis/nerd-fonts/releases/latest/download/FiraCode.zip") ||
		!strings.Contains(string(commands), "unzip:") ||
		!strings.Contains(string(commands), "cache:") ||
		!strings.Contains(string(commands), "/.local/share/fonts\n") {
		t.Errorf("commands = %q, want download, extract, and cache refresh", commands)
	}
	if _, err := os.Stat(filepath.Join(fontDir, "font.ttf")); err != nil {
		t.Errorf("extracted font = %v, want file", err)
	}
}

func TestShippedFontLinuxScriptStopsWhenDownloadFails(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\nexit 22\n")
	writeScriptStub(t, filepath.Join(binDir, "unzip"), "#!/bin/sh\nprintf 'unzip\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "fc-cache"), "#!/bin/sh\nprintf 'cache\\n' >> \"$COMMAND_LOG\"\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "Hack.zip")

	if _, err := runPOSIXScript(t, "assets/scripts/font-linux.sh", binDir); err == nil {
		t.Fatal("font-linux.sh returned nil after the download failed")
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	if got := string(commands); got != "curl\n" {
		t.Errorf("commands = %q, want curl only", got)
	}
}

func TestShippedFontLinuxScriptStopsWhenUnzipIsMissing(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "mktemp"), "#!/bin/sh\nprintf '%s\\n' \"$1\"\n")
	writeScriptStub(t, filepath.Join(binDir, "mkdir"), "#!/bin/sh\n")
	writeScriptStub(t, filepath.Join(binDir, "rm"), "#!/bin/sh\n")
	writeScriptStub(t, filepath.Join(binDir, "fc-cache"), "#!/bin/sh\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "Hack.zip")

	log, err := runPOSIXScriptWithPath(t, "assets/scripts/font-linux.sh", binDir)
	if err == nil {
		t.Fatal("font-linux.sh returned nil when unzip was unavailable")
	}
	if got, want := log, "Error: unzip is required to install Nerd Fonts.\n"; got != want {
		t.Errorf("font-linux.sh output = %q, want %q", got, want)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("curl ran without unzip; stat error = %v", err)
	}
}

func TestShippedFontLinuxScriptStopsWhenExtractionFails(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'archive' > \"$3\"\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "unzip"), "#!/bin/sh\nprintf 'unzip\\n' >> \"$COMMAND_LOG\"\nexit 32\n")
	writeScriptStub(t, filepath.Join(binDir, "fc-cache"), "#!/bin/sh\nprintf 'cache\\n' >> \"$COMMAND_LOG\"\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "Hack.zip")

	log, err := runPOSIXScript(t, "assets/scripts/font-linux.sh", binDir)
	if err == nil {
		t.Fatal("font-linux.sh returned nil after unzip failed")
	}
	if strings.Contains(log, "Nerd Font installation complete") {
		t.Errorf("font-linux.sh output = %q, must not report success", log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	if got := string(commands); got != "curl\nunzip\n" {
		t.Errorf("commands = %q, want curl and unzip only", got)
	}
}

func TestShippedFontLinuxScriptStopsWhenFontCacheRefreshFails(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'archive' > \"$3\"\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "unzip"), "#!/bin/sh\nmkdir -p \"$4\"\nprintf 'unzip\\n' >> \"$COMMAND_LOG\"\n")
	writeScriptStub(t, filepath.Join(binDir, "fc-cache"), "#!/bin/sh\nprintf 'cache\\n' >> \"$COMMAND_LOG\"\nexit 43\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "Hack.zip")

	log, err := runPOSIXScript(t, "assets/scripts/font-linux.sh", binDir)
	if err == nil {
		t.Fatal("font-linux.sh returned nil after fc-cache failed")
	}
	if strings.Contains(log, "Nerd Font installation complete") {
		t.Errorf("font-linux.sh output = %q, must not report success", log)
	}
	commands, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(command log): %v", err)
	}
	if got := string(commands); got != "curl\nunzip\ncache\n" {
		t.Errorf("commands = %q, want curl, unzip, and failed cache refresh", got)
	}
}

func TestShippedFontLinuxScriptRejectsUnsafeFontName(t *testing.T) {
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "commands.log")
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf 'curl\\n' >> \"$COMMAND_LOG\"\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COMMAND_LOG", logFile)
	t.Setenv("FONT_NAME", "../unsafe.zip")

	if _, err := runPOSIXScript(t, "assets/scripts/font-linux.sh", binDir); err == nil {
		t.Fatal("font-linux.sh returned nil for an unsafe font name")
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("curl ran for an unsafe font name; stat error = %v", err)
	}
}

func writeScriptStub(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0700); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func runPOSIXScript(t *testing.T, script, binDir string) (string, error) {
	t.Helper()
	return runPOSIXScriptWithPath(t, script, scriptTestPath(binDir))
}

func runPOSIXScriptWithPath(t *testing.T, script, path string) (string, error) {
	t.Helper()
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Abs repository root: %v", err)
	}

	cmd := exec.Command(posixShell(t), filepath.Join(repoRoot, script))
	cmd.Env = replaceEnv(os.Environ(), "PATH", path)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runPOSIXScriptWithEnv(t *testing.T, script, path string, env map[string]string) (string, error) {
	t.Helper()
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Abs repository root: %v", err)
	}

	cmd := exec.Command(posixShell(t), filepath.Join(repoRoot, script))
	cmd.Env = replaceEnv(os.Environ(), "PATH", path)
	for key, value := range env {
		cmd.Env = replaceEnv(cmd.Env, key, value)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func posixShell(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		return "/bin/sh"
	}

	gitDash := filepath.Join(os.Getenv("ProgramFiles"), "Git", "usr", "bin", "dash.exe")
	if _, err := os.Stat(gitDash); err != nil {
		t.Fatalf("Git POSIX dash is required for shipped-script tests: %v", err)
	}
	return gitDash
}

func scriptTestPath(binDir string) string {
	path := binDir + string(os.PathListSeparator)
	if runtime.GOOS == "windows" {
		path += filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin") + string(os.PathListSeparator)
	}
	return path + os.Getenv("PATH")
}

func replaceEnv(env []string, key, value string) []string {
	prefix := key + "="
	replaced := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			replaced = append(replaced, entry)
		}
	}
	return append(replaced, prefix+value)
}
