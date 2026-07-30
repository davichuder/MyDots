package modules

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFileJoinsReplacementAndCleanupFailures(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config")
	replaceErr := errors.New("replace failed")
	cleanupErr := errors.New("cleanup failed")
	originalReplace, originalRemove := replaceFile, removeTempFile
	replaceFile = func(_, _ string) error { return replaceErr }
	removeTempFile = func(path string) error {
		if err := os.Remove(path); err != nil {
			return err
		}
		return cleanupErr
	}
	t.Cleanup(func() {
		replaceFile = originalReplace
		removeTempFile = originalRemove
	})

	err := atomicWriteFile(target, []byte("replacement"), 0644)
	if !errors.Is(err, replaceErr) {
		t.Fatalf("atomicWriteFile() error = %v, want replacement error", err)
	}
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("atomicWriteFile() error = %v, want cleanup error", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("target state after failed replacement = %v, want unchanged", statErr)
	}
}
