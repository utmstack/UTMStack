//go:build linux || darwin

package quarantine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

func platformRestoreDirectory(path string) (*restoreDirectory, error) {
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if part == "" {
			continue
		}
		next, e := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		unix.Close(fd)
		if e != nil {
			return nil, &os.PathError{Op: "open restore directory", Path: path, Err: e}
		}
		fd = next
	}
	return &restoreDirectory{file: os.NewFile(uintptr(fd), path), path: path}, nil
}

func (d *restoreDirectory) open(name string, flags int, mode uint32) (*os.File, error) {
	fd, err := unix.Openat(int(d.file.Fd()), name, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, mode)
	if err != nil {
		return nil, &os.PathError{Op: "open restore entry", Path: filepath.Join(d.path, name), Err: err}
	}
	return os.NewFile(uintptr(fd), filepath.Join(d.path, name)), nil
}

func (d *restoreDirectory) private() error {
	info, err := d.file.Stat()
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || st.Uid != uint32(os.Geteuid()) || info.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("%w: quarantine directory must be owned by the service and not writable by other users", ErrRestoreIntegrity)
	}
	return nil
}

// A bounded set of persistent lock files avoids unlink/recreate lock races
// and unbounded per-item lock-file growth. Collisions return a retryable error.
func (d *restoreDirectory) lockItem(id string) (func(), error) {
	if err := d.private(); err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(id))
	f, err := d.open(fmt.Sprintf(".item-lock-%02x", sum[0]), unix.O_RDWR|unix.O_CREAT, 0600)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, fmt.Errorf("%w: invalid quarantine lock", ErrRestoreIntegrity)
	}
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return nil, ErrQuarantineBusy
		}
		return nil, err
	}
	return func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN); _ = f.Close() }, nil
}

func (d *restoreDirectory) removeFile(name string) error {
	if err := unix.Unlinkat(int(d.file.Fd()), name, 0); err != nil {
		return &os.PathError{Op: "remove quarantine entry", Path: filepath.Join(d.path, name), Err: err}
	}
	return nil
}

func (d *restoreDirectory) sameEntry(name string, info os.FileInfo) bool {
	before, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	var after unix.Stat_t
	return unix.Fstatat(int(d.file.Fd()), name, &after, unix.AT_SYMLINK_NOFOLLOW) == nil && uint64(before.Dev) == uint64(after.Dev) && before.Ino == after.Ino && after.Mode&unix.S_IFMT == unix.S_IFREG
}

func (d *restoreDirectory) stillNamed() error {
	now, err := platformRestoreDirectory(d.path)
	if err != nil {
		return fmt.Errorf("destination directory changed: %w", err)
	}
	defer now.Close()
	before, err := d.file.Stat()
	if err != nil {
		return err
	}
	after, err := now.file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, after) {
		return errors.New("destination directory changed during recovery")
	}
	return nil
}

func (d *restoreDirectory) staging() (*restoreDirectory, string, error) {
	name := ".edr-restore-" + uuid.NewString()
	if err := unix.Mkdirat(int(d.file.Fd()), name, 0700); err != nil {
		return nil, "", err
	}
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		_ = unix.Unlinkat(int(d.file.Fd()), name, unix.AT_REMOVEDIR)
		return nil, "", err
	}
	stage := &restoreDirectory{file: os.NewFile(uintptr(fd), filepath.Join(d.path, name)), path: filepath.Join(d.path, name)}
	info, err := stage.file.Stat()
	if err != nil || info.Mode().Perm() != 0700 {
		stage.Close()
		_ = unix.Unlinkat(int(d.file.Fd()), name, unix.AT_REMOVEDIR)
		return nil, "", fmt.Errorf("%w: private staging directory permissions unavailable", ErrRestoreIntegrity)
	}
	return stage, name, nil
}

func cleanupRestoreStage(parent, stage *restoreDirectory, name, entry string) error {
	var err error
	if entry != "" {
		err = stage.removeFile(entry)
		if os.IsNotExist(err) {
			err = nil
		}
	}
	err = errors.Join(err, stage.Close())
	// Never recurse through a destination-controlled name. rmdir cannot delete
	// files if that name was replaced with a non-empty directory or symlink.
	remove := unix.Unlinkat(int(parent.file.Fd()), name, unix.AT_REMOVEDIR)
	if !errors.Is(remove, unix.ENOENT) {
		err = errors.Join(err, remove)
	}
	return err
}

func preserveRestoreOwner(out *os.File, info os.FileInfo) error {
	before, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("stored ownership metadata unavailable")
	}
	after, err := out.Stat()
	if err != nil {
		return err
	}
	current, ok := after.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("snapshot ownership metadata unavailable")
	}
	if current.Uid != before.Uid || current.Gid != before.Gid {
		if err := out.Chown(int(before.Uid), int(before.Gid)); err != nil {
			return err
		}
	}
	if after.Mode().Perm() != 0600 {
		return errors.New("restore requires private readable/writable snapshot permissions")
	}
	return nil
}

// A successful link is already publication. Its private source name stays
// available for cleanup, whose error is reported after recording publication.
func publishRestoreSnapshot(from *restoreDirectory, old string, to *restoreDirectory, name string, rename func(*restoreDirectory, string, *restoreDirectory, string) error) (moved bool, err error) {
	err = rename(from, old, to, name)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.EOPNOTSUPP) {
		return false, err
	}
	return false, unix.Linkat(int(from.file.Fd()), old, int(to.file.Fd()), name, 0)
}

func restoreOpenInput(root *restoreDirectory, name string) (*os.File, error) {
	return admissionReadFile(root, name)
}
func restoreCreateOutput(stage *restoreDirectory) (*os.File, error) {
	return admissionCreateFile(stage, "snapshot")
}
func restoreFinishOutput(out *os.File, info os.FileInfo, _ string) error {
	return preserveRestoreOwner(out, info)
}
func restorePublishOutput(_ *os.File, stage, parent *restoreDirectory, leaf string) error {
	_, err := publishRestoreSnapshot(stage, "snapshot", parent, leaf, renameRestoreEntry)
	return err
}
