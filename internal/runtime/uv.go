package runtime

import (
	"context"
	"io"
	"io/fs"

	"github.com/davichuder/MyDots/internal/installer/runner"
)

// UvManager implements RuntimeManager for uv (Python package manager).
type UvManager struct{}

// Install installs uv via Homebrew.
func (u UvManager) Install(ctx context.Context, logw io.Writer, _ fs.FS) error {
	return runner.Brew(ctx, logw, "install", "uv")
}

// IsInstalled checks whether uv is on PATH.
func (u UvManager) IsInstalled() bool {
	return runner.CommandExists("uv")
}

// InstallRuntime installs a specific Python version via uv.
func (u UvManager) InstallRuntime(ctx context.Context, version string, logw io.Writer) error {
	return runner.Run(ctx, logw, "uv", "python", "install", version)
}

// AuditInfo returns the uv version string.
func (u UvManager) AuditInfo() string {
	return runner.CaptureOutput("uv", "--version")
}
