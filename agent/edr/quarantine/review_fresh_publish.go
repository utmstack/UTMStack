package quarantine

import (
	"bytes"
	"errors"
	"os"
	"time"
)

func (p *freshRestorePermit) copyAndPublish(s *Store, parent, op *restoreDirectory, input *os.File, m *restoreManifest) error {
	if err := p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	stage, err := p.openStage(parent, *m)
	if err != nil {
		return err
	}
	defer stage.Close()
	if err = freshKnownEntries(stage, map[string]bool{freshTicketName: false, "snapshot": false}); err != nil {
		return err
	}
	if err = p.ensureTicket(stage, op, *m); err != nil {
		return err
	}
	if err = s.cache.TrackQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
		return err
	}
	if err = p.seam("stage-indexed"); err != nil {
		return err
	}
	var out *os.File
	if m.OutputID == "" {
		if m.Phase != "copying" {
			return errors.New("fresh verified output identity is missing")
		}
		out, err = restoreCreateOutput(stage)
		if err != nil {
			return err
		}
		defer out.Close()
		m.OutputID, err = admissionFileID(out)
		if err != nil {
			return err
		}
		if err = admissionSyncDirectory(stage); err != nil {
			return err
		}
		if err = p.seam("output-created-filesystem"); err != nil {
			return err
		}
		if err = validateFreshJournal(*m, p.approval, p.owner); err != nil {
			return err
		}
		if err = writeRestoreManifest(op, *m); err != nil {
			return err
		}
		if err = p.seam("output-identified"); err != nil {
			return err
		}
	} else if m.Phase == "copying" {
		out, err = openFreshOutput(stage)
		if err != nil {
			// Only the exact private scratch output may be protected again.
			// A crash can leave its original DACL restored before verified was
			// recorded. The source pin retains writer exclusion during this
			// metadata operation; the subsequent writable open rechecks its ID.
			pin, e := admissionOpenSource(stage, "snapshot")
			if e != nil {
				return errors.Join(err, e)
			}
			id, e := admissionFileID(pin)
			if e == nil && id != m.OutputID {
				e = errors.New("fresh scratch identity changed")
			}
			if e == nil {
				e = admissionProtectCaptured(pin)
			}
			e = errors.Join(e, pin.Close())
			if e != nil {
				return e
			}
			out, err = openFreshOutput(stage)
		}
		if err != nil {
			return err
		}
		defer out.Close()
	} else {
		// A verified output needs a strict read/rename pin, not data-write
		// permission that its restored original DACL may deliberately deny.
		out, err = admissionOpenSource(stage, "snapshot")
		if err != nil {
			return err
		}
		defer out.Close()
	}
	id, err := admissionFileID(out)
	if err != nil || id != m.OutputID {
		return errors.Join(err, errors.New("fresh scratch output identity changed"))
	}
	if committed, e := s.cache.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID); e != nil || committed {
		return errors.Join(e, errors.New("fresh unpublished operation conflicts with metadata"))
	}
	if m.Phase == "copying" {
		if err = out.Truncate(0); err != nil {
			return err
		}
		if _, err = out.Seek(0, 0); err != nil {
			return err
		}
		info, e := restoreInputInfo(input)
		if e != nil {
			return e
		}
		expected, e := restoreDigest(p.approval.Inspection.Snapshot.SHA256)
		if e != nil {
			return e
		}
		if info.Size() != p.approval.Inspection.Snapshot.Size {
			return errors.New("fresh source size changed")
		}
		if err = s.copyRestoreSnapshot(out, input, info, expected); err != nil {
			return err
		}
		if err = p.seam("copied"); err != nil {
			return err
		}
		if err = restoreFinishOutput(out, info, m.Record.OriginalSecurity); err != nil {
			return err
		}
		if err = out.Sync(); err != nil {
			return err
		}
		if err = p.seam("security-restored"); err != nil {
			return err
		}
		m.Phase = "verified"
		if err = writeRestoreManifest(op, *m); err != nil {
			return err
		}
		if err = p.seam("verified"); err != nil {
			return err
		}
	} else {
		pin, e := inspectFreshInput(out)
		if e != nil || pin.FileID != m.OutputID || pin.Size != p.approval.Inspection.Snapshot.Size || pin.SHA256 != p.approval.Inspection.Snapshot.SHA256 {
			return errors.Join(e, errors.New("fresh verified output changed before publication"))
		}
		if err = admissionSyncSource(out); err != nil {
			return err
		}
	}
	if err = p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	if err = p.checkSnapshot(); err != nil {
		return err
	}
	if err = p.checkOutput(out, *m); err != nil {
		return err
	}
	if err = parent.stillNamed(); err != nil {
		return err
	}
	if err = requirePurgeNameAbsent(parent, string(m.Leaf)); err != nil {
		return err
	}
	if committed, e := s.cache.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID); e != nil || committed {
		return errors.Join(e, errors.New("fresh metadata changed before publication"))
	}
	m.Phase = "publishing"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = p.seam("publication-intent"); err != nil {
		return err
	}
	// Recheck after the interruption seam, immediately before no-replace
	// publication. No old destination or output is ever adopted here.
	if err = p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	if err = p.checkOutput(out, *m); err != nil {
		return err
	}
	if err = parent.stillNamed(); err != nil {
		return err
	}
	if err = restorePublishOutput(out, stage, parent, string(m.Leaf)); err != nil {
		return err
	}
	if err = p.seam("published-filesystem"); err != nil {
		return err
	}
	if err = admissionSyncDirectory(parent); err != nil {
		return err
	}
	m.Phase = "published"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	return p.seam("published")
}

func (p *freshRestorePermit) checkSnapshot() error {
	f, err := restoreOpenInput(p.root, p.approval.QuarantineID+".quarantined")
	if err != nil {
		return err
	}
	defer f.Close()
	actual, err := inspectFreshInput(f)
	if err != nil || actual != p.approval.Inspection.Snapshot {
		return errors.Join(err, errors.New("fresh retained snapshot changed before publication"))
	}
	return nil
}

func (p *freshRestorePermit) checkOutput(out *os.File, m restoreManifest) error {
	pin, err := inspectFreshInput(out)
	if err != nil || pin.FileID != m.OutputID || pin.Size != p.approval.Inspection.Snapshot.Size || pin.SHA256 != p.approval.Inspection.Snapshot.SHA256 {
		return errors.Join(err, errors.New("fresh output changed before publication"))
	}
	return nil
}

func (p *freshRestorePermit) finish(s *Store, parent, op *restoreDirectory, m *restoreManifest) error {
	if err := p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	committed, err := s.cache.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID)
	if err != nil {
		return err
	}
	if !committed {
		if m.Phase != "published" {
			return errors.New("fresh publication lacks durable proof")
		}
		if err = p.seam("before-database-commit"); err != nil {
			return err
		}
		if err = p.checkOwnedOperation(op, *m); err != nil {
			return err
		}
		if err = s.cache.CommitQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID, time.Now().UTC()); err != nil {
			return err
		}
		if err = p.seam("database-committed"); err != nil {
			return err
		}
	}
	m.Phase = "committed"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = p.seam("committed"); err != nil {
		return err
	}
	if err = p.cleanupStage(s, parent, op, m); err != nil {
		return err
	}
	// In contrast with normal finishRestore, both the independently retained
	// snapshot and original raw capture are deliberately left untouched.
	return nil
}

func (p *freshRestorePermit) cleanupStage(s *Store, parent, op *restoreDirectory, m *restoreManifest) error {
	if err := p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	if m.StageState == "deleted" {
		return nil
	}
	stage, err := p.openStage(parent, *m)
	if os.IsNotExist(err) && m.StageState == "deleting" {
		if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
			return err
		}
		m.StageState = "deleted"
		return writeRestoreManifest(op, *m)
	}
	if err != nil {
		return err
	}
	defer stage.Close()
	if err = freshKnownEntries(stage, map[string]bool{freshTicketName: false, "snapshot": false}); err != nil {
		return err
	}
	output, err := purgeOpenFile(stage, "snapshot")
	if err == nil {
		if output.id != m.OutputID {
			output.Close()
			return errors.New("fresh cleanup found a replacement output")
		}
		err = purgeRemoveFile(output, stage, "snapshot")
		err = errors.Join(err, output.Close())
		if err != nil {
			return err
		}
		if err = requirePurgeNameAbsent(stage, "snapshot"); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	// Retire the owner ticket only after no valuable output remains in the
	// stage. Unknown children are retained, never recursively removed.
	if err = admissionSyncDirectory(stage); err != nil {
		return err
	}
	m.StageState = "deleting"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = p.seam("stage-delete-intent"); err != nil {
		return err
	}
	if err = p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	entry, err := readFreshControl(stage, freshTicketName, maxFreshOwnership)
	if err == nil {
		_, expected, e := p.expectedTicket()
		if e != nil || entry.ManifestID != p.owner.TicketID || !bytes.Equal(entry.Manifest, expected) {
			return errors.Join(e, errors.New("fresh cleanup ticket changed"))
		}
		file, e := purgeOpenFile(stage, freshTicketName)
		if e != nil {
			return e
		}
		if file.id != p.owner.TicketID {
			file.Close()
			return errors.New("fresh cleanup ticket identity changed")
		}
		e = purgeRemoveFile(file, stage, freshTicketName)
		e = errors.Join(e, file.Close())
		if e != nil {
			return e
		}
		if err = requirePurgeNameAbsent(stage, freshTicketName); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = admissionSyncDirectory(stage); err != nil {
		return err
	}
	if err = p.seam("ticket-retired"); err != nil {
		return err
	}
	if err = p.checkOwnedOperation(op, *m); err != nil {
		return err
	}
	if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
		return err
	}
	if err = stage.Close(); err != nil {
		return err
	}
	if err = admissionRemoveDirectory(parent, m.stageName()); err != nil {
		return err
	}
	if err = admissionSyncDirectory(parent); err != nil {
		return err
	}
	if err = p.seam("stage-removed-filesystem"); err != nil {
		return err
	}
	m.StageState = "deleted"
	return writeRestoreManifest(op, *m)
}

func (p *freshRestorePermit) archive(m restoreManifest) error {
	if _, err := p.checkAuthority(); err != nil {
		return err
	}
	if m.Phase != "committed" || m.StageState != "deleted" {
		return errors.New("fresh archive lacks completed cleanup proof")
	}
	journals, err := admissionOpenChild(p.root, ".restores")
	if err != nil {
		return err
	}
	defer journals.Close()
	if err = p.seam("before-archive"); err != nil {
		return err
	}
	current, err := admissionOpenChild(journals, m.Record.QuarantineID)
	if err != nil {
		return err
	}
	err = p.checkOwnedOperation(current, m)
	err = errors.Join(err, current.Close())
	if err != nil {
		return err
	}
	if err = moveRecoveryDirectory(journals, m.Record.QuarantineID, p.decision, "restore", p.owner.OperationDirectoryID); err != nil {
		return err
	}
	if err = admissionSyncDirectory(p.decision); err != nil {
		return err
	}
	if err = admissionSyncDirectory(journals); err != nil {
		return err
	}
	return p.seam("archived")
}
