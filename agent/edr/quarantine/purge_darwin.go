//go:build darwin

package quarantine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func purgeOpenFile(d *restoreDirectory, name string) (*purgeFile, error) {
	// Darwin has no Linux O_PATH equivalent. Query through the pinned private
	// directory without opening payload data; the claimed entry is checked
	// again before deletion. Other users cannot alter these private contents.
	var stat unix.Stat_t
	if err := unix.Fstatat(int(d.file.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return nil, &os.PathError{Op: "inspect purge identity", Path: filepath.Join(d.path, name), Err: err}
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, errors.New("purge requires a regular stored artifact")
	}
	return &purgeFile{id: fmt.Sprintf("%x:%x", uint64(stat.Dev), stat.Ino)}, nil
}
