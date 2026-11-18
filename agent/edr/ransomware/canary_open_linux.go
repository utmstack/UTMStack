//go:build linux

package ransomware

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func canaryDirectoryOwner(dir *os.File) (uint32, error) {
	var st unix.Stat_t
	if err := unix.Fstat(int(dir.Fd()), &st); err != nil {
		return 0, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return 0, errors.New("canary ownership requires a pinned directory")
	}
	return st.Uid, nil
}

func openPlatformCanaryFile(dir *os.File, name string, flags int, mode uint32, create bool) (*os.File, error) {
	owner, err := canaryDirectoryOwner(dir)
	if err != nil {
		return nil, fmt.Errorf("canary directory initial ownership metadata: %w", err)
	}
	fd, err := unix.Openat(int(dir.Fd()), name, flags, mode)
	if err != nil {
		return nil, err
	}
	// No pathname chown/unlink: a replacement name cannot redirect this operation
	// to an existing user file. A failed admission leaves only an unregistered inert
	// newly created file; deleting its potentially replaced name would be unsafe.
	fileNow := func() (canaryOwnerState, error) {
		var st unix.Stat_t
		if err := unix.Fstat(fd, &st); err != nil {
			return canaryOwnerState{}, err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFREG {
			return canaryOwnerState{}, errors.New("canary descriptor is not a regular file")
		}
		return canaryOwnerState{uid: st.Uid, mode: st.Mode}, nil
	}
	err = checkCanaryOwnership(owner, create, func() (uint32, error) { return canaryDirectoryOwner(dir) }, fileNow, func(uid uint32) error { return unix.Fchown(fd, int(uid), -1) })
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	return os.NewFile(uintptr(fd), filepath.Join(dir.Name(), name)), nil
}
