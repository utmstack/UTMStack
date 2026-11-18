//go:build linux || darwin

package quarantine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func admissionFileID(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", errors.New("file identity unavailable")
	}
	return fmt.Sprintf("%x:%x", uint64(st.Dev), st.Ino), nil
}
func openAdmissionRoot(path string, _ bool) (*restoreDirectory, error) {
	d, err := platformRestoreDirectory(path)
	if err != nil {
		return nil, err
	}
	if err = d.private(); err != nil {
		d.Close()
		return nil, err
	}
	info, err := d.file.Stat()
	if err != nil || info.Mode().Perm()&0077 != 0 {
		d.Close()
		return nil, errors.New("quarantine storage requires service-only directory permissions")
	}
	return d, nil
}
func admissionOpenChild(d *restoreDirectory, name string) (*restoreDirectory, error) {
	f, err := d.open(name, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	child := &restoreDirectory{file: f, path: filepath.Join(d.path, name)}
	if err = child.private(); err != nil {
		child.Close()
		return nil, err
	}
	info, err := child.file.Stat()
	if err != nil || info.Mode().Perm()&0077 != 0 {
		child.Close()
		return nil, errors.New("admission directory does not provide service-only access")
	}
	return child, nil
}
func admissionMakeDirectory(d *restoreDirectory, name string) (*restoreDirectory, error) {
	if err := unix.Mkdirat(int(d.file.Fd()), name, 0700); err != nil {
		return nil, &os.PathError{Op: "create admission directory", Path: filepath.Join(d.path, name), Err: err}
	}
	child, err := admissionOpenChild(d, name)
	if err != nil {
		_ = unix.Unlinkat(int(d.file.Fd()), name, unix.AT_REMOVEDIR)
	}
	return child, err
}
func admissionRemoveDirectory(d *restoreDirectory, name string) error {
	return unix.Unlinkat(int(d.file.Fd()), name, unix.AT_REMOVEDIR)
}
func admissionCreateFile(d *restoreDirectory, name string) (*os.File, error) {
	return d.open(name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL, 0600)
}
func admissionReadFile(d *restoreDirectory, name string) (*os.File, error) {
	return d.open(name, unix.O_RDONLY, 0)
}
func admissionReadPinnedSource(d *restoreDirectory, name string) (*os.File, error) {
	return admissionReadFile(d, name)
}
func admissionOpenSource(d *restoreDirectory, name string) (*os.File, error) {
	return d.open(name, unix.O_RDONLY, 0)
}
func admissionSyncDirectory(d *restoreDirectory) error { return d.file.Sync() }

// The additional source barrier is currently qualified only on Windows.
func admissionSyncSource(*os.File) error { return nil }
func admissionSnapshotOwner(out *os.File, info os.FileInfo) error {
	return preserveRestoreOwner(out, info)
}
func admissionClaim(_ *os.File, from *restoreDirectory, old string, to *restoreDirectory, name string) error {
	return renameRestoreEntry(from, old, to, name)
}
func admissionRollback(_ *os.File, from *restoreDirectory, old string, to *restoreDirectory, name string) error {
	return renameRestoreEntry(from, old, to, name)
}
func admissionOriginalSecurity(*os.File) (string, error)    { return "", nil }
func admissionProtectCaptured(*os.File) error               { return nil }
func admissionRestoreSourceSecurity(*os.File, string) error { return nil }
func admissionRename(from *restoreDirectory, old string, to *restoreDirectory, name string, replace bool) error {
	if !replace {
		return renameRestoreEntry(from, old, to, name)
	}
	return unix.Renameat(int(from.file.Fd()), old, int(to.file.Fd()), name)
}
