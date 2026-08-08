// Package platform detects the current OS and variant (native/WSL2)
// through an injectable goos parameter and optional /proc/version reading.
package platform

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

type OS string

const (
	Darwin  OS = "darwin"
	Linux   OS = "linux"
	Windows OS = "windows"
)

type Variant string

const (
	Native Variant = "native"
	WSL2   Variant = "wsl2"
)

type Platform struct {
	OS      OS
	Variant Variant
	Arch    string
}

var (
	ErrWindowsDetected = errors.New("windows detected: use WSL2")
)

type ErrUnsupportedOS struct {
	GOOS string
}

func (e ErrUnsupportedOS) Error() string {
	return "unsupported OS: " + e.GOOS
}

const procVersionPath string = "/proc/version"

const wsl2Marker string = "microsoft"

var readProcVersion = func() ([]byte, error) {
	return os.ReadFile(procVersionPath)
}

func Detect(goos string) (Platform, error) {
	switch goos {
	case "windows":
		return Platform{}, ErrWindowsDetected

	case "darwin":
		return Platform{OS: Darwin, Variant: Native, Arch: runtime.GOARCH}, nil

	case "linux":
		data, err := readProcVersion()
		variant := Native
		if err == nil && strings.Contains(strings.ToLower(string(data)), wsl2Marker) {
			variant = WSL2
		}
		return Platform{OS: Linux, Variant: variant, Arch: runtime.GOARCH}, nil

	default:
		return Platform{}, ErrUnsupportedOS{GOOS: goos}
	}
}
