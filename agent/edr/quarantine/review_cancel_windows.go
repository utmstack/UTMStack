package quarantine

import (
	"errors"

	"golang.org/x/sys/windows"
)

func moveRecoveryDirectory(from *restoreDirectory, old string, to *restoreDirectory, name, expectedID string) error {
	// Pin the exact private directory with DELETE access and rename that handle.
	// Ordinary admission directory handles exclude delete sharing; no such
	// inspection handle may remain open across this separate move operation.
	f, err := admissionNTFile(windows.Handle(from.file.Fd()), old, windows.DELETE|windows.FILE_LIST_DIRECTORY|windows.FILE_TRAVERSE|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = admissionPrivateDirectory(f); err != nil {
		return err
	}
	id, err := admissionFileID(f)
	if err != nil || id != expectedID {
		return errors.Join(err, errors.New("cancellation directory identity changed before move"))
	}
	return renameAdmissionHandle(f, to, name, false)
}
