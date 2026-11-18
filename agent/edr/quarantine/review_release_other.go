//go:build !linux && !darwin && !windows

package quarantine

import (
	"errors"
	"os"
)

func openRecoveryReleaseReceipt(*restoreDirectory) (*os.File, error) {
	return nil, errors.New("recovery release is unsupported on this platform")
}
