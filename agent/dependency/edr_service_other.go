//go:build !windows && !linux

package dependency

import "errors"

// The native EDR service is distributed for Windows and Linux.
func edrServiceStatus() (edrServiceState, error) {
	return edrServiceState{}, errors.New("EDR dependency installation is unavailable on this platform")
}
