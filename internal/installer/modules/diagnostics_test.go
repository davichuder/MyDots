package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestUnsupportedWSLDiagnosticsSurfaceWriterFailures(t *testing.T) {
	writeErr := errors.New("diagnostic sink unavailable")
	ctx := types.InstallContext{
		Cancel:   context.Background(),
		Log:      failingWriter{err: writeErr},
		Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
	}

	t.Run("clipboard skip reports diagnostic failure", func(t *testing.T) {
		originalLookup := lookupEnv
		lookupEnv = func(string) string { return "" }
		t.Cleanup(func() { lookupEnv = originalLookup })

		err := (ClipboardModule{}).Install(ctx)
		if !errors.Is(err, writeErr) {
			t.Fatalf("Install() error = %v, want diagnostic writer error", err)
		}
	})

	t.Run("ghostty skip reports diagnostic failure", func(t *testing.T) {
		originalLookup := lookupEnv
		lookupEnv = func(string) string { return "" }
		t.Cleanup(func() { lookupEnv = originalLookup })

		err := (GhosttyModule{}).Install(ctx)
		if !errors.Is(err, writeErr) {
			t.Fatalf("Install() error = %v, want diagnostic writer error", err)
		}
	})

	t.Run("docker failure preserves systemd failure and reports diagnostic failure", func(t *testing.T) {
		systemdErr := errors.New("systemd unavailable")
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				if name == "systemctl" {
					return nil, systemdErr
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		err := (DockerModule{}).Install(ctx)
		if !errors.Is(err, systemdErr) {
			t.Fatalf("Install() error = %v, want preserved systemd error", err)
		}
		if !errors.Is(err, writeErr) {
			t.Fatalf("Install() error = %v, want diagnostic writer error", err)
		}
	})
}

func TestUnsupportedWSLDiagnosticsKeepExistingModuleSemantics(t *testing.T) {
	originalLookup := lookupEnv
	lookupEnv = func(string) string { return "" }
	t.Cleanup(func() { lookupEnv = originalLookup })

	ctx := types.InstallContext{
		Cancel:   context.Background(),
		Log:      &bytes.Buffer{},
		Platform: platform.Platform{OS: platform.Linux, Variant: platform.WSL2},
	}
	for _, module := range []types.Module{ClipboardModule{}, GhosttyModule{}} {
		if err := module.Install(ctx); err != nil {
			t.Fatalf("%s Install() error = %v, want nil skip", module.Name(), err)
		}
	}
}
