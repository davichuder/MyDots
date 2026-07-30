package runner

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	writeScriptStub(t, filepath.Join(binDir, "curl"), "#!/bin/sh\nprintf '%s\\n' 'echo downloaded-installer-ran'\n")
	writeScriptStub(t, filepath.Join(binDir, "bash"), "#!/bin/sh\neval \"$(cat \"$1\")\"\n")
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
