package quarantine

import (
	"os"

	"golang.org/x/sys/windows"
)

func openRecoveryReleaseReceipt(op *restoreDirectory) (*os.File, error) {
	// FlushFileBuffers requires write access, including when retrying a complete
	// receipt left by a process that exited before its original flush returned.
	return admissionNTFile(windows.Handle(op.file.Fd()), "manifest.json", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
}
