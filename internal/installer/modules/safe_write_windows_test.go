//go:build windows

package modules

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFileWindowsReplacement(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := atomicWriteFile(target, []byte("replacement"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "replacement" {
		t.Fatalf("replacement = %q, %v", data, err)
	}

	originalReplace := replaceFile
	replaceFile = func(_, _ string) error { return errors.New("replace failed") }
	t.Cleanup(func() { replaceFile = originalReplace })
	if err := atomicWriteFile(target, []byte("lost"), 0644); err == nil {
		t.Fatal("expected replacement failure")
	}
	data, err = os.ReadFile(target)
	if err != nil || string(data) != "replacement" {
		t.Fatalf("original after failed replacement = %q, %v", data, err)
	}
}
