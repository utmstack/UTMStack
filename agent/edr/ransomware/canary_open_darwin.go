//go:build darwin

package ransomware

import (
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// macOS keeps its existing descriptor-only creation/adoption behavior. Linux
// directory-owner coverage policy is not inferred for another operating system.
func openPlatformCanaryFile(dir *os.File, name string, flags int, mode uint32, _ bool) (*os.File, error) {
	fd, err := unix.Openat(int(dir.Fd()), name, flags, mode)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), filepath.Join(dir.Name(), name)), nil
}
