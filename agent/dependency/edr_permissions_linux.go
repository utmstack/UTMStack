//go:build linux

package dependency

import (
	"errors"
	"os"
)

// Only the staged regular file is made executable. The promotion transaction
// still validates its identity and preserves the previous service executable.
func prepareEDRExecutable(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("staged EDR executable is not a regular file")
	}
	return os.Chmod(path, 0755)
}
