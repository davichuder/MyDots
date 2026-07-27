// Package runtime defines the RuntimeManager interface and its implementations
// for fnm (Node), uv (Python), and sdkman (Java).
//
// Following ADR-011, adding a new runtime manager requires only a new struct
// that satisfies this interface, with no changes to orchestration logic.
package runtime

import (
	"context"
	"io"
	"io/fs"
)

// RuntimeManager abstracts the lifecycle of language runtime managers.
type RuntimeManager interface {
	// Install installs the runtime manager itself (e.g., brew install fnm).
	Install(ctx context.Context, logw io.Writer, assets fs.FS) error

	// IsInstalled checks whether the runtime manager is available on the system.
	IsInstalled() bool

	// InstallRuntime installs a specific runtime version (e.g., fnm install 24).
	InstallRuntime(ctx context.Context, version string, logw io.Writer) error

	// AuditInfo returns the version string of the runtime manager.
	AuditInfo() string
}
