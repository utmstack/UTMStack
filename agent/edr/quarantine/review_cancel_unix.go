//go:build linux || darwin

package quarantine

import "errors"

func moveRecoveryDirectory(from *restoreDirectory, old string, to *restoreDirectory, name, expectedID string) error {
	op, err := admissionOpenChild(from, old)
	if err != nil {
		return err
	}
	defer op.Close()
	id, err := admissionFileID(op.file)
	if err != nil || id != expectedID {
		return errors.Join(err, errors.New("cancellation directory identity changed before move"))
	}
	return renameRestoreEntry(from, old, to, name)
}
