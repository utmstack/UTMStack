//go:build darwin

package quarantine

import (
	"golang.org/x/sys/unix"
)

func renameRestoreEntry(from *restoreDirectory, old string, to *restoreDirectory, name string) error {
	return unix.RenameatxNp(int(from.file.Fd()), old, int(to.file.Fd()), name, unix.RENAME_EXCL)
}
