//go:build linux || darwin

package scanner

import (
	"golang.org/x/sys/unix"
	"os"
)

func openScannerTarget(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open scan target", Path: path, Err: err}
	}
	return os.NewFile(uintptr(fd), path), nil
}
