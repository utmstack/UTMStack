//go:build linux || darwin

package ransomware

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const canaryPinRequired = true

func duplicateCanaryFile(f *os.File) (*os.File, error) {
	fd, err := unix.FcntlInt(f.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), f.Name()), nil
}

func canaryFileID(info os.FileInfo) (canaryFileIdentity, bool) {
	if info == nil {
		return canaryFileIdentity{}, false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st.Dev == 0 || st.Ino == 0 {
		return canaryFileIdentity{}, false
	}
	return canaryFileIdentity{device: uint64(st.Dev), inode: uint64(st.Ino)}, true
}
