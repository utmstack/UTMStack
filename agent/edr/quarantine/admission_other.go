//go:build !linux && !darwin && !windows

package quarantine

import (
	"errors"
	"os"
)

func admissionFileID(*os.File) (string, error)                  { return "", errors.ErrUnsupported }
func openAdmissionRoot(string, bool) (*restoreDirectory, error) { return nil, errors.ErrUnsupported }
func admissionOpenChild(*restoreDirectory, string) (*restoreDirectory, error) {
	return nil, errors.ErrUnsupported
}
func admissionMakeDirectory(*restoreDirectory, string) (*restoreDirectory, error) {
	return nil, errors.ErrUnsupported
}
func admissionRemoveDirectory(*restoreDirectory, string) error { return errors.ErrUnsupported }
func admissionCreateFile(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}
func admissionReadFile(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}
func admissionReadPinnedSource(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}
func admissionOpenSource(*restoreDirectory, string) (*os.File, error) {
	return nil, errors.ErrUnsupported
}
func admissionSyncDirectory(*restoreDirectory) error     { return errors.ErrUnsupported }
func admissionSyncSource(*os.File) error                 { return errors.ErrUnsupported }
func admissionSnapshotOwner(*os.File, os.FileInfo) error { return errors.ErrUnsupported }
func admissionClaim(*os.File, *restoreDirectory, string, *restoreDirectory, string) error {
	return errors.ErrUnsupported
}
func admissionRollback(*os.File, *restoreDirectory, string, *restoreDirectory, string) error {
	return errors.ErrUnsupported
}
func admissionRename(*restoreDirectory, string, *restoreDirectory, string, bool) error {
	return errors.ErrUnsupported
}
func admissionOriginalSecurity(*os.File) (string, error)    { return "", errors.ErrUnsupported }
func admissionProtectCaptured(*os.File) error               { return errors.ErrUnsupported }
func admissionRestoreSourceSecurity(*os.File, string) error { return errors.ErrUnsupported }
