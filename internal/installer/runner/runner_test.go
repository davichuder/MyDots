package runner

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"os"
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
		if len(capturedArgs) != 2 || capturedArgs[0] != "--cask" || capturedArgs[1] != "ghostty" {
			t.Errorf("expected args ['--cask', 'ghostty'], got: %v", capturedArgs)
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
