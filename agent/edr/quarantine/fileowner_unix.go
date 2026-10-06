//go:build unix

package quarantine

import (
	"os"
	"syscall"
)

// fileOwner returns the numeric owner UID/GID of a file on unix.
func fileOwner(path string) (int, int, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}
	if sys, ok := st.Sys().(*syscall.Stat_t); ok {
		return int(sys.Uid), int(sys.Gid), nil
	}
	return 0, 0, nil
}
