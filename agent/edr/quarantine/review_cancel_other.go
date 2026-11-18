//go:build !linux && !darwin && !windows

package quarantine

import "errors"

func moveRecoveryDirectory(*restoreDirectory, string, *restoreDirectory, string, string) error {
	return errors.ErrUnsupported
}
