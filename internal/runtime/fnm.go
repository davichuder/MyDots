package runtime

import (
	"context"
	"io"
	"io/fs"

	"github.com/davichuder/MyDots/internal/installer/runner"
)

// FnmManager implements RuntimeManager for fnm (Fast Node Manager).
type FnmManager struct{}

// Install installs fnm via Homebrew.
func (f FnmManager) Install(ctx context.Context, logw io.Writer, _ fs.FS) error {
	return runner.Brew(ctx, logw, "install", "fnm")
}

// IsInstalled checks whether fnm is on PATH.
func (f FnmManager) IsInstalled() bool {
	return runner.CommandExists("fnm")
}

// InstallRuntime installs a specific Node version via fnm.
func (f FnmManager) InstallRuntime(ctx context.Context, version string, logw io.Writer) error {
	return runner.Run(ctx, logw, "fnm", "install", version)
}

// AuditInfo returns the fnm version string.
func (f FnmManager) AuditInfo() string {
	return runner.CaptureOutput("fnm", "--version")
}
