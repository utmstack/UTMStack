//go:build linux || darwin

package ransomware

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type canaryDir struct{ file *os.File }

func openCanaryDir(path string) (*canaryDir, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || len(path) > 4096 {
		return nil, errors.New("canary directory must be a clean absolute path")
	}
	// Walk one component per openat. Every lookup rejects symlinks, and the next
	// lookup is relative to an already pinned directory, including during races.
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, component := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if component == "" {
			continue
		}
		next, err := unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		unix.Close(fd)
		if err != nil {
			return nil, err
		}
		fd = next
	}
	return &canaryDir{os.NewFile(uintptr(fd), path)}, nil
}
func (d *canaryDir) Close() error { return d.file.Close() }
func (d *canaryDir) open(name string, create bool) (*os.File, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return nil, errors.New("canary requires a single filename")
	}
	flags := unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_RDONLY
	if create {
		flags = unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_WRONLY | unix.O_CREAT | unix.O_EXCL
	}
	return openPlatformCanaryFile(d.file, name, flags, 0644, create)
}
