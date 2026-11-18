//go:build linux

package quarantine

import "golang.org/x/sys/unix"

func renameNoReplace(src, dst string) error {
	return renameError(src, dst, unix.Renameat2(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_NOREPLACE))
}
