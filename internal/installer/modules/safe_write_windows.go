//go:build windows

package modules

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const replaceFileWriteThrough = 0x00000001

var replaceFileW = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReplaceFileW")

// platformReplaceFile uses ReplaceFileW for existing targets. For a new target it uses
// MoveFileExW with write-through; neither path falls back to deleting a target.
func platformReplaceFile(source, target string) error {
	sourcePtr, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	targetPtr, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return windows.MoveFileEx(sourcePtr, targetPtr, windows.MOVEFILE_WRITE_THROUGH)
	} else if err != nil {
		return err
	}
	r1, _, callErr := replaceFileW.Call(
		uintptr(unsafe.Pointer(targetPtr)), uintptr(unsafe.Pointer(sourcePtr)), 0,
		replaceFileWriteThrough, 0, 0,
	)
	if r1 == 0 {
		return callErr
	}
	return nil
}
