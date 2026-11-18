//go:build linux

package quarantine

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func purgeOpenFile(d *restoreDirectory, name string) (*purgeFile, error) {
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open purge identity", Path: filepath.Join(d.path, name), Err: err}
	}
	return purgeFileFromDescriptor(os.NewFile(uintptr(fd), filepath.Join(d.path, name)))
}
