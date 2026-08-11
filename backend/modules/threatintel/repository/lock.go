package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const lockFileName = configFileName + ".lock"

// withFileLock serialises writers across replicas: the config directory is a
// shared volume, so a second backend saving at the same moment would otherwise
// publish over the first one's file.
func (s *ConfigStore) withFileLock(fn func() error) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(s.dir, lockFileName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("threatintel: open the config lock: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("threatintel: take the config lock: %w", err)
	}
	defer func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }()

	return fn()
}
