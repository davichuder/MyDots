// Package errors defines the MyDotsError interface and concrete error types
// that carry What/Why/Fix context for TUI rendering.
package errors

import (
	"fmt"
	"strings"
)

// MyDotsError is the interface all MyDots-specific errors implement.
// Every error carries a concrete description (What), context about the cause (Why),
// and a suggested next step for the user (Fix).
type MyDotsError interface {
	error
	What() string
	Why() string
	Fix() string
}

// BrewInstallError is returned when a brew install command fails.
type BrewInstallError struct {
	Formula  string
	ExitCode int
	Stderr   string
}

func (e BrewInstallError) Error() string {
	return fmt.Sprintf("brew install %s failed (exit %d)", e.Formula, e.ExitCode)
}

func (e BrewInstallError) What() string {
	return fmt.Sprintf("brew install %s failed with exit code %d.", e.Formula, e.ExitCode)
}

func (e BrewInstallError) Why() string {
	return fmt.Sprintf("Homebrew reported: %s", e.Stderr)
}

func (BrewInstallError) Fix() string {
	return "Run 'brew doctor' to diagnose Homebrew issues, then re-run 'mydots'."
}

// AptInstallError is returned when an apt-get install command fails.
type AptInstallError struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e AptInstallError) Error() string {
	return fmt.Sprintf("apt install failed (exit %d): %s", e.ExitCode, e.Command)
}

func (e AptInstallError) What() string {
	return fmt.Sprintf("apt install failed with exit code %d while running: %s.", e.ExitCode, e.Command)
}

func (e AptInstallError) Why() string {
	return fmt.Sprintf("apt reported: %s", e.Stderr)
}

func (AptInstallError) Fix() string {
	return "Run 'sudo apt update' then re-run 'mydots'. If the issue persists, check your package sources."
}

// CurlScriptError is returned when a curl-script download or execution fails.
type CurlScriptError struct {
	URL      string
	ExitCode int
	Stderr   string
}

func (e CurlScriptError) Error() string {
	return fmt.Sprintf("curl script failed (exit %d) for URL: %s", e.ExitCode, e.URL)
}

func (e CurlScriptError) What() string {
	return fmt.Sprintf("Failed to download or execute script from %s (exit code %d).", e.URL, e.ExitCode)
}

func (e CurlScriptError) Why() string {
	return fmt.Sprintf("curl reported: %s", e.Stderr)
}

func (CurlScriptError) Fix() string {
	return "Check your internet connection and try again. If using a VPN, try disconnecting and re-running 'mydots'."
}

// ConfigWriteError is returned when writing the configuration file fails.
type ConfigWriteError struct {
	Path string
	Err  error
}

func (e ConfigWriteError) Error() string {
	return fmt.Sprintf("failed to write config file %s: %s", e.Path, e.Err.Error())
}

func (e ConfigWriteError) What() string {
	return fmt.Sprintf("Failed to write the configuration file at %s.", e.Path)
}

func (e ConfigWriteError) Why() string {
	return e.Err.Error()
}

func (ConfigWriteError) Fix() string {
	return "Ensure the config directory is writable and try again."
}

// GenericError wraps external errors and provides pattern-matched Fix suggestions.
type GenericError struct {
	Err error
}

func (e GenericError) Error() string {
	return e.Err.Error()
}

func (e GenericError) What() string {
	return "An external error occurred while running an operation."
}

func (e GenericError) Why() string {
	return e.Err.Error()
}

func (e GenericError) Fix() string {
	msg := e.Err.Error()
	switch {
	case strings.Contains(msg, patternNotInPath):
		return fixNotInPath
	case strings.Contains(msg, patternPermissionDenied):
		return fixPermissionDenied
	case strings.Contains(msg, patternConnectionRefused):
		return fixConnectionRefused
	case strings.Contains(msg, patternNoSuchFile):
		return fixNoSuchFile
	default:
		return fallbackFix
	}
}

// Pattern constants for GenericError fix matching.
const (
	patternNotInPath         string = "not found in PATH"
	patternPermissionDenied  string = "permission denied"
	patternConnectionRefused string = "connection refused"
	patternNoSuchFile        string = "no such file or directory"
)

// Fix suggestion constants for GenericError pattern matching.
const (
	fixNotInPath         string = "Make sure the program is installed and available in your PATH."
	fixPermissionDenied  string = "You may need to run this with elevated privileges or check file permissions."
	fixConnectionRefused string = "Check your network connection and ensure the service is running."
	fixNoSuchFile        string = "Check that the file path exists and try again."
	fallbackFix          string = "Check the log above for details and re-run 'mydots'."
)
