//go:build windows

package fanotify

import (
	"golang.org/x/sys/windows"
	"os"
)

func duplicateFixtureFile(f *os.File) (*os.File, error) {
	var copy windows.Handle
	proc := windows.CurrentProcess()
	if err := windows.DuplicateHandle(proc, windows.Handle(f.Fd()), proc, &copy, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(copy), f.Name()), nil
}
