//go:build windows

package quarantine

import (
	"golang.org/x/sys/windows"
	"os"
)

func openFreshOutput(stage *restoreDirectory) (*os.File, error) {
	return openFreshWritableFile(stage, "snapshot")
}

func openFreshWritableFile(d *restoreDirectory, name string) (*os.File, error) {
	return restoreNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_WRITE_THROUGH, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_DELETE)
}

func pinFreshDatabaseFile(d *restoreDirectory, name string) (*os.File, error) {
	return admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE)
}
