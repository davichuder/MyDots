package installer

import (
	"context"
	"io"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

// InstallContext bundles shared state for an entire install session.
// It is passed to every Module.Install() call so that all modules
// share the same platform, config, timestamp, log writer, and
// cancellation context.
type InstallContext struct {
	// Platform is the detected OS and variant for this session.
	Platform platform.Platform

	// Config is the user's configuration loaded from mydots-config.json.
	Config config.Config

	// SessionTimestamp is generated once at session start and shared
	// across all backup operations (colons replaced with dashes).
	SessionTimestamp string

	// Log receives stdout/stderr output from the runner for TUI display.
	Log io.Writer

	// Cancel carries the context used by exec.CommandContext.
	// Cancelling this context kills the running subprocess.
	Cancel context.Context
}
