package modules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var replaceFile = platformReplaceFile
var removeTempFile = os.Remove

// atomicWriteFile writes a complete sibling temporary file before replacement.
// POSIX uses rename; Windows uses its strongest replacement primitive without a
// delete-and-recreate fallback, so a reported replacement failure leaves the
// original path untouched by this function.
func atomicWriteFile(path string, data []byte, perm os.FileMode) (err error) {
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".mydots-*.tmp")
	if err != nil {
		return err
	}
	cleanupNeeded := true
	defer func() {
		if !cleanupNeeded {
			return
		}
		if cleanupErr := removeTempFile(tmp.Name()); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("remove temporary file: %w", cleanupErr))
		}
	}()
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	err = replaceFile(tmp.Name(), path)
	if err == nil {
		cleanupNeeded = false
	}
	return err
}
