//go:build !windows

package event

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func openSpoolRegular(path string, flags int, mode os.FileMode) (*os.File, error) {
	fd, err := unix.Open(path, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, uint32(mode.Perm()))
	if err != nil {
		return nil, &os.PathError{Op: "open event spool", Path: path, Err: err}
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		f.Close()
		return nil, fmt.Errorf("%w: not a regular file", ErrSpoolIntegrity)
	}
	return f, nil
}
func trySpoolLock(f *os.File) (bool, error) {
	err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return false, nil
	}
	return err == nil, err
}
func unlockSpoolFile(f *os.File)             { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }
func replaceSpoolFile(from, to string) error { return os.Rename(from, to) }
func syncSpoolDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
