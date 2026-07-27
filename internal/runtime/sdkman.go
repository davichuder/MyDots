package runtime

import (
	"context"
	"io"
	"io/fs"
	"os"
	"os/user"

	"github.com/davichuder/MyDots/internal/installer/runner"
)

// SdkmanManager implements RuntimeManager for sdkman (Java SDK Manager).
type SdkmanManager struct{}

// checkSdkmanDir is injectable for testing. Checks whether ~/.sdkman exists.
var checkSdkmanDir = func() bool {
	usr, err := user.Current()
	if err != nil {
		return false
	}
	_, err = os.Stat(usr.HomeDir + "/.sdkman")
	return err == nil
}

// Install installs sdkman via the official curl script.
func (s SdkmanManager) Install(ctx context.Context, logw io.Writer, assets fs.FS) error {
	return runner.Script(ctx, logw, assets, "assets/scripts/sdkman-install.sh", nil)
}

// IsInstalled checks whether ~/.sdkman exists on the filesystem.
func (s SdkmanManager) IsInstalled() bool {
	return checkSdkmanDir()
}

// InstallRuntime installs a specific Java version via sdkman.
func (s SdkmanManager) InstallRuntime(ctx context.Context, version string, logw io.Writer) error {
	return runner.Run(ctx, logw, "sdk", "install", "java", version)
}

// AuditInfo returns the sdkman version string.
func (s SdkmanManager) AuditInfo() string {
	return runner.CaptureOutput("sdk", "version")
}
