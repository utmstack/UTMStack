//go:build windows

package engine

import (
	"golang.org/x/sys/windows"
	"os"
)

func lockUpdate(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	overlap := new(windows.Overlapped)
	h := windows.Handle(f.Fd())
	if err = windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlap); err != nil {
		f.Close()
		return nil, err
	}
	return func() { _ = windows.UnlockFileEx(h, 0, 1, 0, overlap); _ = f.Close() }, nil
}

func replaceFile(from, to string) error {
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
func syncDirectory(string) error { return nil }
