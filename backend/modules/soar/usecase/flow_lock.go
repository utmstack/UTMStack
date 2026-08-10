package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const flowLockFile = ".flows.lock"

func (s *FlowStore) withFlowLock(fn func() error) error {
	if err := os.MkdirAll(s.userDir, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(s.userDir, flowLockFile), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("soar: open the flow lock: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("soar: take the flow lock: %w", err)
	}
	defer func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }()

	return fn()
}

func (s *FlowStore) flowExistsOnDisk(tenant, relPath string) bool {
	_, err := os.Stat(filepath.Join(s.userDir, tenant, filepath.FromSlash(relPath)))
	return err == nil
}
