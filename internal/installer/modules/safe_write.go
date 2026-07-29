package modules

import (
	"os"
	"path/filepath"
)

var replaceFile = platformReplaceFile

// atomicWriteFile writes a complete sibling temporary file before replacement.
// POSIX uses rename; Windows uses its strongest replacement primitive without a
// delete-and-recreate fallback, so a reported replacement failure leaves the
// original path untouched by this function.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".mydots-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
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
	return replaceFile(tmp.Name(), path)
}
