//go:build darwin

package quarantine

import "golang.org/x/sys/unix"

func renameNoReplace(src, dst string) error {
	return renameError(src, dst, unix.RenamexNp(src, dst, unix.RENAME_EXCL))
}
