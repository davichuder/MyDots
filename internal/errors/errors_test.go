package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestBrewInstallError_AllMethodsReturnNonEmpty(t *testing.T) {
	e := BrewInstallError{
		Formula:  "zoxide",
		ExitCode: 127,
		Stderr:   "Error: Command failed",
	}

	if e.Error() == "" {
		t.Error("Error() should not be empty")
	}
	if e.What() == "" {
		t.Error("What() should not be empty")
	}
	if e.Why() == "" {
		t.Error("Why() should not be empty")
	}
	if e.Fix() == "" {
		t.Error("Fix() should not be empty")
	}
}

func TestBrewInstallError_FieldsInOutput(t *testing.T) {
	e := BrewInstallError{
		Formula:  "zoxide",
		ExitCode: 127,
		Stderr:   "command not found",
	}

	got := e.Error()
	if !strings.Contains(got, "zoxide") {
		t.Errorf("Error() should contain formula, got: %q", got)
	}
	if !strings.Contains(got, "127") {
		t.Errorf("Error() should contain exit code, got: %q", got)
	}

	what := e.What()
	if !strings.Contains(what, "zoxide") {
		t.Errorf("What() should contain formula, got: %q", what)
	}

	why := e.Why()
	if !strings.Contains(why, "command not found") {
		t.Errorf("Why() should contain stderr, got: %q", why)
	}
}

func TestAptInstallError_AllMethodsReturnNonEmpty(t *testing.T) {
	e := AptInstallError{
		Command:  "apt install -y gcc",
		ExitCode: 1,
		Stderr:   "E: Unable to locate package",
	}

	if e.Error() == "" {
		t.Error("Error() should not be empty")
	}
	if e.What() == "" {
		t.Error("What() should not be empty")
	}
	if e.Why() == "" {
		t.Error("Why() should not be empty")
	}
	if e.Fix() == "" {
		t.Error("Fix() should not be empty")
	}
}

func TestAptInstallError_FieldsInOutput(t *testing.T) {
	e := AptInstallError{
		Command:  "apt install -y gcc",
		ExitCode: 100,
		Stderr:   "E: Broken packages",
	}

	got := e.Error()
	if !strings.Contains(got, "apt install -y gcc") {
		t.Errorf("Error() should contain command, got: %q", got)
	}
	if !strings.Contains(got, "100") {
		t.Errorf("Error() should contain exit code, got: %q", got)
	}

	why := e.Why()
	if !strings.Contains(why, "E: Broken packages") {
		t.Errorf("Why() should contain stderr, got: %q", why)
	}
}

func TestCurlScriptError_AllMethodsReturnNonEmpty(t *testing.T) {
	e := CurlScriptError{
		URL:      "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh",
		ExitCode: 7,
		Stderr:   "curl: (7) Failed to connect",
	}

	if e.Error() == "" {
		t.Error("Error() should not be empty")
	}
	if e.What() == "" {
		t.Error("What() should not be empty")
	}
	if e.Why() == "" {
		t.Error("Why() should not be empty")
	}
	if e.Fix() == "" {
		t.Error("Fix() should not be empty")
	}
}

func TestCurlScriptError_FieldsInOutput(t *testing.T) {
	e := CurlScriptError{
		URL:      "https://example.com/script.sh",
		ExitCode: 22,
		Stderr:   "HTTP 404",
	}

	got := e.Error()
	if !strings.Contains(got, "example.com") {
		t.Errorf("Error() should contain URL, got: %q", got)
	}
	if !strings.Contains(got, "22") {
		t.Errorf("Error() should contain exit code, got: %q", got)
	}

	why := e.Why()
	if !strings.Contains(why, "HTTP 404") {
		t.Errorf("Why() should contain stderr, got: %q", why)
	}
}

func TestConfigWriteError_AllMethodsReturnNonEmpty(t *testing.T) {
	e := ConfigWriteError{
		Path: "/home/user/.config/mydots/mydots-config.json",
		Err:  errors.New("permission denied"),
	}

	if e.Error() == "" {
		t.Error("Error() should not be empty")
	}
	if e.What() == "" {
		t.Error("What() should not be empty")
	}
	if e.Why() == "" {
		t.Error("Why() should not be empty")
	}
	if e.Fix() == "" {
		t.Error("Fix() should not be empty")
	}
}

func TestConfigWriteError_PathAndErrInOutput(t *testing.T) {
	path := "/tmp/mydots-config.json"
	e := ConfigWriteError{
		Path: path,
		Err:  errors.New("disk full"),
	}

	got := e.Error()
	if !strings.Contains(got, path) {
		t.Errorf("Error() should contain path, got: %q", got)
	}

	what := e.What()
	if !strings.Contains(what, path) {
		t.Errorf("What() should contain path, got: %q", what)
	}

	why := e.Why()
	if !strings.Contains(why, "disk full") {
		t.Errorf("Why() should contain underlying error, got: %q", why)
	}
}

func TestGenericError_KnownPatternsReturnActionableFix(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantFix string
	}{
		{
			name:    "not found in PATH",
			err:     errors.New("zsh: command not found in PATH"),
			wantFix: fixNotInPath,
		},
		{
			name:    "permission denied",
			err:     errors.New("open /etc/shells: permission denied"),
			wantFix: fixPermissionDenied,
		},
		{
			name:    "connection refused",
			err:     errors.New("dial tcp 127.0.0.1:443: connect: connection refused"),
			wantFix: fixConnectionRefused,
		},
		{
			name:    "no such file or directory",
			err:     errors.New("stat /nonexistent: no such file or directory"),
			wantFix: fixNoSuchFile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ge := GenericError{Err: tc.err}
			if ge.Error() == "" {
				t.Error("Error() should not be empty")
			}
			if ge.What() == "" {
				t.Error("What() should not be empty")
			}
			if ge.Why() == "" {
				t.Error("Why() should not be empty")
			}
			got := ge.Fix()
			if got != tc.wantFix {
				t.Errorf("Fix() = %q, want %q", got, tc.wantFix)
			}
		})
	}
}

func TestGenericError_UnknownPatternReturnsFallback(t *testing.T) {
	ge := GenericError{Err: errors.New("something unexpected happened")}

	got := ge.Fix()
	if got != fallbackFix {
		t.Errorf("Fix() = %q, want fallback %q", got, fallbackFix)
	}
}

func TestGenericError_AllMethodsNonEmpty(t *testing.T) {
	ge := GenericError{Err: errors.New("random error")}

	if ge.Error() == "" {
		t.Error("Error() should not be empty")
	}
	if ge.What() == "" {
		t.Error("What() should not be empty")
	}
	if ge.Why() == "" {
		t.Error("Why() should not be empty")
	}
	if ge.Fix() == "" {
		t.Error("Fix() should not be empty")
	}
}

func TestErrorsAsBrewInstallError(t *testing.T) {
	err := BrewInstallError{
		Formula:  "bat",
		ExitCode: 1,
		Stderr:   "some error",
	}

	var target BrewInstallError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match BrewInstallError")
	}
	if target.Formula != "bat" {
		t.Errorf("expected formula 'bat', got: %q", target.Formula)
	}
}

func TestErrorsAsAptInstallError(t *testing.T) {
	err := AptInstallError{
		Command:  "apt install",
		ExitCode: 1,
		Stderr:   "error",
	}

	var target AptInstallError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match AptInstallError")
	}
	if target.Command != "apt install" {
		t.Errorf("expected command 'apt install', got: %q", target.Command)
	}
}

func TestErrorsAsCurlScriptError(t *testing.T) {
	err := CurlScriptError{
		URL:      "https://example.com",
		ExitCode: 7,
		Stderr:   "error",
	}

	var target CurlScriptError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match CurlScriptError")
	}
	if target.URL != "https://example.com" {
		t.Errorf("expected URL, got: %q", target.URL)
	}
}

func TestErrorsAsConfigWriteError(t *testing.T) {
	err := ConfigWriteError{
		Path: "/tmp/config.json",
		Err:  errors.New("disk full"),
	}

	var target ConfigWriteError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match ConfigWriteError")
	}
	if target.Path != "/tmp/config.json" {
		t.Errorf("expected path, got: %q", target.Path)
	}
}

func TestErrorsAsGenericError(t *testing.T) {
	err := GenericError{Err: errors.New("something")}

	var target GenericError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match GenericError")
	}
}

func TestTriangulationMultipleErrorTypes(t *testing.T) {
	// At least 3 different error types with different values
	errs := []error{
		BrewInstallError{Formula: "atuin", ExitCode: 1, Stderr: "err1"},
		AptInstallError{Command: "apt-get", ExitCode: 2, Stderr: "err2"},
		CurlScriptError{URL: "https://example.com", ExitCode: 3, Stderr: "err3"},
	}

	for _, err := range errs {
		if err.Error() == "" {
			t.Errorf("Error() should be non-empty for %T", err)
		}
		mdErr, ok := err.(MyDotsError)
		if !ok {
			t.Errorf("expected %T to implement MyDotsError", err)
			continue
		}
		if mdErr.What() == "" {
			t.Errorf("What() should be non-empty for %T", err)
		}
		if mdErr.Why() == "" {
			t.Errorf("Why() should be non-empty for %T", err)
		}
		if mdErr.Fix() == "" {
			t.Errorf("Fix() should be non-empty for %T", err)
		}
	}
}

func TestGenericErrorNotInPathContainsAllKeywords(t *testing.T) {
	// Verify the pattern matching catches "not found in PATH" anywhere in the message
	ge := GenericError{Err: errors.New("some tool was not found in PATH because it's missing")}

	fix := ge.Fix()
	if !strings.Contains(fix, "PATH") {
		t.Errorf("Fix() for 'not found in PATH' should mention PATH, got: %q", fix)
	}
}
