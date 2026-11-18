//go:build !linux && !darwin && !windows

package quarantine

import (
	"errors"
	"os"
)

func openFreshOutput(*restoreDirectory) (*os.File, error) { return nil, errors.ErrUnsupported }

func openFreshWritableFile(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}

func pinFreshDatabaseFile(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}
