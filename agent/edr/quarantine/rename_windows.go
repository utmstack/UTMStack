//go:build windows

package quarantine

import "golang.org/x/sys/windows"

func renameNoReplace(src, dst string) error {
	from, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return renameError(src, dst, err)
	}
	to, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return renameError(src, dst, err)
	}
	// Omit MOVEFILE_REPLACE_EXISTING and MOVEFILE_COPY_ALLOWED: an existing
	// destination must win, and cross-volume moves use our O_EXCL fallback.
	return renameError(src, dst, windows.MoveFileEx(from, to, windows.MOVEFILE_WRITE_THROUGH))
}
