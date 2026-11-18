//go:build linux || darwin

package quarantine

import (
	"os"

	"golang.org/x/sys/unix"
)

func openRecoveryReleaseReceipt(op *restoreDirectory) (*os.File, error) {
	return op.open("manifest.json", unix.O_RDWR, 0)
}
