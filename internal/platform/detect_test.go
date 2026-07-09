package platform

import (
	"errors"
	"testing"
)

func TestDetectDarwin(t *testing.T) {
	p, err := Detect("darwin")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.OS != Darwin {
		t.Errorf("expected OS Darwin, got: %v", p.OS)
	}
	if p.Variant != Native {
		t.Errorf("expected Variant Native, got: %v", p.Variant)
	}
}

func TestDetectWindows(t *testing.T) {
	_, err := Detect("windows")
	if !errors.Is(err, ErrWindowsDetected) {
		t.Errorf("expected ErrWindowsDetected, got: %v", err)
	}
}

func TestDetectLinuxNative(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return []byte("Linux version 5.15.0-91-generic"), nil
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.OS != Linux {
		t.Errorf("expected OS Linux, got: %v", p.OS)
	}
	if p.Variant != Native {
		t.Errorf("expected Variant Native, got: %v", p.Variant)
	}
}

func TestDetectLinuxWSL2(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return []byte("Linux version 5.15.0-91-generic Microsoft (WSL2)"), nil
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.OS != Linux {
		t.Errorf("expected OS Linux, got: %v", p.OS)
	}
	if p.Variant != WSL2 {
		t.Errorf("expected Variant WSL2, got: %v", p.Variant)
	}
}

func TestDetectUnsupportedOS(t *testing.T) {
	_, err := Detect("plan9")
	if _, ok := err.(ErrUnsupportedOS); !ok {
		t.Errorf("expected ErrUnsupportedOS, got: %v (%T)", err, err)
	}
}

func TestDetectLinuxProcVersionUnreadable(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return nil, errors.New("permission denied")
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Variant != Native {
		t.Errorf("expected Variant Native when /proc/version unreadable, got: %v", p.Variant)
	}
}

func TestDetectLinuxProcVersionMissing(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return nil, errors.New("file not found")
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Variant != Native {
		t.Errorf("expected Variant Native when /proc/version missing, got: %v", p.Variant)
	}
}

func TestDetectLinuxProcVersionEmpty(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return []byte(""), nil
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Variant != Native {
		t.Errorf("expected Variant Native when /proc/version empty, got: %v", p.Variant)
	}
}

func TestDetectLinuxProcVersionMicrosoftUppercase(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return []byte("Linux version 5.15.0-91-generic MICROSOFT"), nil
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Variant != WSL2 {
		t.Errorf("expected Variant WSL2 with 'MICROSOFT', got: %v", p.Variant)
	}
}

func TestDetectLinuxProcVersionMultiLineMicrosoft(t *testing.T) {
	orig := readProcVersion
	readProcVersion = func() ([]byte, error) {
		return []byte("First line\nmicrosoft on second line"), nil
	}
	defer func() { readProcVersion = orig }()

	p, err := Detect("linux")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Variant != WSL2 {
		t.Errorf("expected Variant WSL2 when 'microsoft' on line 2, got: %v", p.Variant)
	}
}
