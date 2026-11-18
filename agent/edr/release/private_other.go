//go:build !windows

package release

import (
	"errors"
	"os"
)

// MakePrivateTempDir creates a staging directory inaccessible to other users.
func MakePrivateTempDir(parent, prefix string) (string, error) { return os.MkdirTemp(parent, prefix) }

// CheckPrivateDirectory checks the platform access policy.
func CheckPrivateDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return errors.New("release directory is not private")
	}
	return nil
}

// PreparePublishedFile preserves permissions applied by the Unix updater.
func PreparePublishedFile(path string) error { return nil }
