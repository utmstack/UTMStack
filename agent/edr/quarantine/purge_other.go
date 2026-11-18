//go:build !linux && !darwin && !windows

package quarantine

import (
	"errors"
)

func purgeOpenFile(*restoreDirectory, string) (*purgeFile, error) { return nil, errors.ErrUnsupported }
func purgeRemoveFile(*purgeFile, *restoreDirectory, string) error { return errors.ErrUnsupported }
