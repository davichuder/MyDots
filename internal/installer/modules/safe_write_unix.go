//go:build !windows

package modules

import "os"

// platformReplaceFile uses same-directory rename on POSIX filesystems.
func platformReplaceFile(source, target string) error {
	return os.Rename(source, target)
}
