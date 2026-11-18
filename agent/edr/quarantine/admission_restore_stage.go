package quarantine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func (s *Store) restoreStageLocator(m restoreManifest) cache.QuarantineRestoreStage {
	return cache.QuarantineRestoreStage{StageID: m.StageID, OperationID: m.OperationID, QuarantineID: m.Record.QuarantineID, RootID: s.rootID, RootPath: []byte(s.dir)}
}

// The SQL row locates proof; it does not grant an exclusion. In particular, a
// reserved-looking name, stale lookup or replacement inode is insufficient.
func (s *Store) managedRestoreStage(parent *restoreDirectory, parentID string) error {
	if err := managedRetiredOriginalStage(parent, parentID); err != nil {
		return err
	}
	if err := managedRetiredFreshStage(parent, parentID); err != nil {
		return err
	}
	if err := managedFreshRestoreStage(parent, parentID); err != nil {
		return err
	}
	locator, found, err := s.cache.QuarantineRestoreStageByIdentity(parentID)
	if err != nil || !found {
		return err
	}
	fail := func(cause error) error {
		return fmt.Errorf("%w: restore stage ownership cannot be proven; objects retained: %v", ErrRestoreIntegrity, cause)
	}
	rootPath := string(locator.RootPath)
	if !ordinaryQuarantineID(locator.QuarantineID) || !filepath.IsAbs(rootPath) || filepath.Clean(rootPath) != rootPath {
		return fail(fmt.Errorf("invalid restore stage lookup path"))
	}
	root, err := openAdmissionRoot(rootPath, false)
	if err != nil {
		return fail(err)
	}
	defer root.Close()
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != locator.RootID {
		return fail(fmt.Errorf("lookup root identity changed: %v", err))
	}
	journals, op, err := openRestoreOperation(root, locator.QuarantineID)
	var m restoreManifest
	if err == nil {
		m, err = readRestoreManifest(op, locator.QuarantineID)
		if err != nil || m.OperationID != locator.OperationID || m.StageID != parentID {
			// A new operation in the live slot cannot supply (or erase)
			// ownership of this exact older cancelled stage.
			op.Close()
			journals.Close()
			err = os.ErrNotExist
		}
	}
	if os.IsNotExist(err) {
		journals, op, err = openArchivedRestoreProof(root, locator.QuarantineID)
		if err == nil {
			m, err = readRestoreManifest(op, locator.QuarantineID)
			if err != nil {
				op.Close()
				journals.Close()
			}
		}
	}
	if err != nil {
		return fail(err)
	}
	defer journals.Close()
	defer op.Close()
	if m.OperationID != locator.OperationID || m.StageID != parentID || m.StageState == "planned" || m.StageState == "deleted" {
		return fail(fmt.Errorf("journal does not identify this stage"))
	}
	if m.Version != 1 {
		// Fresh stages need their separate durable ownership proof. A v2
		// syntax match or a lookup row alone must never grant an exclusion.
		return fail(ErrFreshRestoreApprovalRequired)
	}
	private, err := openAdmissionRoot(parent.path, false)
	if err != nil {
		return fail(err)
	}
	defer private.Close()
	privateID, err := admissionFileID(private.file)
	if err != nil || privateID != parentID {
		return fail(fmt.Errorf("private stage identity changed: %v", err))
	}
	return ErrAlreadyQuarantined
}
