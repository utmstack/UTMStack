//go:build windows

package event

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func openSpoolRegular(path string, flags int, mode os.FileMode) (*os.File, error) {
	before, err := os.Lstat(path)
	if err == nil && !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: not a regular file", ErrSpoolIntegrity)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	f, err := os.OpenFile(path, flags, mode)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		f.Close()
		return nil, ErrSpoolIntegrity
	}
	return f, nil
}
func trySpoolLock(f *os.File) (bool, error) {
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, new(windows.Overlapped))
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return false, nil
	}
	return err == nil, err
}
func unlockSpoolFile(f *os.File) {
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, new(windows.Overlapped))
}
func replaceSpoolFile(from, to string) error {
	f, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	t, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(f, t, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
func syncSpoolDirectory(string) error { return nil }
