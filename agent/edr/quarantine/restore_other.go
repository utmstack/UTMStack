//go:build !linux && !darwin && !windows

package quarantine

import (
	"errors"
	"os"
)

func platformRestoreDirectory(string) (*restoreDirectory, error)   { return nil, errors.ErrUnsupported }
func (d *restoreDirectory) lockItem(string) (func(), error)        { return nil, errors.ErrUnsupported }
func (d *restoreDirectory) removeFile(string) error                { return errors.ErrUnsupported }
func restoreOpenInput(*restoreDirectory, string) (*os.File, error) { return nil, errors.ErrUnsupported }
func restoreCreateOutput(*restoreDirectory) (*os.File, error)      { return nil, errors.ErrUnsupported }
func restoreFinishOutput(*os.File, os.FileInfo, string) error      { return errors.ErrUnsupported }
func restorePublishOutput(*os.File, *restoreDirectory, *restoreDirectory, string) error {
	return errors.ErrUnsupported
}
func (*restoreDirectory) stillNamed() error { return errors.ErrUnsupported }
