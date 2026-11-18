package quarantine

import (
	"bytes"
	"errors"
	"os"
)

func (p *freshRestorePermit) prepare(parent *restoreDirectory, old restoreManifest) (op *restoreDirectory, m restoreManifest, err error) {
	a := p.approval
	if _, err = p.checkAuthority(); err != nil {
		return nil, m, err
	}
	if err = requirePurgeNameAbsent(parent, string(a.Inspection.Destination.Leaf)); err != nil {
		return nil, m, err
	}
	journals, err := admissionChild(p.root, ".restores")
	if err != nil {
		return nil, m, err
	}
	defer journals.Close()
	if err = freshQueueCapacity(journals); err != nil {
		return nil, m, err
	}
	op, err = admissionMakeDirectory(journals, a.QuarantineID)
	if err != nil {
		return nil, m, err
	}
	ok := false
	defer func() {
		if !ok {
			op.Close()
		}
	}()
	if err = admissionSyncDirectory(journals); err != nil {
		return op, m, err
	}
	if err = p.seam("operation-created"); err != nil {
		return op, m, err
	}
	m = restoreManifest{Version: 2, OperationID: a.OperationID, FreshID: a.OperationID, Phase: "copying", Record: old.Record, RecordData: append([]byte(nil), a.Inspection.RecordData...), ParentPath: append([]byte(nil), a.Inspection.Destination.ParentPath...), Leaf: append([]byte(nil), a.Inspection.Destination.Leaf...), ParentID: a.Inspection.Destination.ParentID, StageState: "present", Source: purgeArtifact{ID: a.Inspection.Snapshot.FileID, State: "ready"}}
	stage, err := admissionMakeDirectory(parent, m.stageName())
	if err != nil {
		return op, m, err
	}
	defer stage.Close()
	if err = admissionSyncDirectory(parent); err != nil {
		return op, m, err
	}
	if err = p.seam("stage-created"); err != nil {
		return op, m, err
	}
	m.StageID, err = admissionFileID(stage.file)
	if err != nil {
		return op, m, err
	}
	if err = writeRestoreManifest(op, m); err != nil {
		return op, m, err
	}
	if err = p.seam("stage-intent"); err != nil {
		return op, m, err
	}
	ticket, err := admissionCreateFile(op, freshTicketTemporary)
	if err != nil {
		return op, m, err
	}
	ticketID, idErr := admissionFileID(ticket)
	err = errors.Join(idErr, ticket.Close())
	if err != nil {
		return op, m, err
	}
	ownerFile, err := admissionCreateFile(p.decision, "ownership.json")
	if err != nil {
		return op, m, err
	}
	defer ownerFile.Close()
	ownerID, err := admissionFileID(ownerFile)
	if err != nil {
		return op, m, err
	}
	opID, err := admissionFileID(op.file)
	if err != nil {
		return op, m, err
	}
	p.owner = freshOwnership{Version: 1, Kind: "fresh-restore-ownership", RootPath: []byte(p.root.path), RootID: a.RootID, QuarantineID: a.QuarantineID, OperationID: a.OperationID, ApprovalDirectoryID: a.DirectoryID, ApprovalFileID: a.FileID, ApprovalSHA256: p.approvalHash, FileID: ownerID, OperationDirectoryID: opID, StageID: m.StageID, TicketID: ticketID, ActorID: a.ActorID, Reason: p.reason}
	if err = p.seam("ownership-created"); err != nil {
		return op, m, err
	}
	_, body, err := freshJSONHash(p.owner)
	if err != nil || len(body) > maxFreshOwnership {
		return op, m, errors.Join(err, errors.New("fresh ownership exceeds metadata bound"))
	}
	if _, err = ownerFile.Write(body); err != nil {
		return op, m, err
	}
	if err = p.seam("ownership-written"); err != nil {
		return op, m, err
	}
	if err = ownerFile.Sync(); err != nil {
		return op, m, err
	}
	if err = ownerFile.Close(); err != nil {
		return op, m, err
	}
	if err = admissionSyncDirectory(p.decision); err != nil {
		return op, m, err
	}
	if err = admissionSyncDirectory(op); err != nil {
		return op, m, err
	}
	if err = p.seam("ownership-sealed"); err != nil {
		return op, m, err
	}
	if err = p.ensureTicket(stage, op, m); err != nil {
		return op, m, err
	}
	if err = p.seam("ticket-published"); err != nil {
		return op, m, err
	}
	ok = true
	return op, m, nil
}

func (p *freshRestorePermit) expectedTicket() (freshOwnerTicket, []byte, error) {
	hash, _, err := freshJSONHash(p.owner)
	if err != nil {
		return freshOwnerTicket{}, nil, err
	}
	t := freshOwnerTicket{Version: 1, RootPath: append([]byte(nil), p.owner.RootPath...), RootID: p.owner.RootID, QuarantineID: p.owner.QuarantineID, OperationID: p.owner.OperationID, StageID: p.owner.StageID, FileID: p.owner.TicketID, OwnershipID: p.owner.FileID, OwnershipSHA256: hash}
	_, body, err := freshJSONHash(t)
	return t, body, err
}

func syncFreshNamedRecord(d *restoreDirectory, name, id, hash string) error {
	f, err := openFreshWritableFile(d, name)
	if err != nil {
		return err
	}
	defer f.Close()
	entry, err := inspectJournalFile(JournalInspection{}, f, maxFreshOwnership)
	if err != nil || entry.ManifestID != id || entry.SHA256 != hash {
		return errors.Join(err, errors.New("fresh control record changed before flush"))
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return admissionSyncDirectory(d)
}

func (p *freshRestorePermit) ensureTicket(stage, op *restoreDirectory, m restoreManifest) error {
	ownerHash, _, err := freshJSONHash(p.owner)
	if err != nil {
		return err
	}
	if err = syncFreshNamedRecord(p.decision, "ownership.json", p.owner.FileID, ownerHash); err != nil {
		return err
	}
	_, body, err := p.expectedTicket()
	if err != nil {
		return err
	}
	marker, markerErr := readFreshControl(stage, freshTicketName, maxFreshOwnership)
	temporary, tempErr := inspectPurgeArtifact(op, freshTicketTemporary)
	if tempErr != nil {
		return tempErr
	}
	if markerErr == nil {
		if marker.ManifestID != p.owner.TicketID || !bytes.Equal(marker.Manifest, body) || temporary.State != "absent" {
			return errors.New("fresh ticket identity, bytes or location changed")
		}
		return syncFreshNamedRecord(stage, freshTicketName, marker.ManifestID, marker.SHA256)
	}
	if !os.IsNotExist(markerErr) || m.Phase != "copying" || m.OutputID != "" || temporary.State == "absent" || temporary.ID != p.owner.TicketID {
		return errors.Join(markerErr, errors.New("fresh stage ticket is missing or ambiguous"))
	}
	if err = requirePurgeNameAbsent(stage, "snapshot"); err != nil {
		return err
	}
	f, err := openFreshWritableFile(op, freshTicketTemporary)
	if err != nil {
		return err
	}
	defer f.Close()
	id, err := admissionFileID(f)
	if err != nil || id != p.owner.TicketID {
		return errors.Join(err, errors.New("fresh temporary ticket identity changed"))
	}
	if err = f.Truncate(0); err != nil {
		return err
	}
	if _, err = f.Seek(0, 0); err != nil {
		return err
	}
	if _, err = f.Write(body); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = admissionSyncDirectory(op); err != nil {
		return err
	}
	if err = p.seam("ticket-prepared"); err != nil {
		return err
	}
	if err = admissionRename(op, freshTicketTemporary, stage, freshTicketName, false); err != nil {
		return err
	}
	if err = admissionSyncDirectory(stage); err != nil {
		return err
	}
	return admissionSyncDirectory(op)
}

func (p *freshRestorePermit) checkStageReadOnly(parent, op *restoreDirectory, m restoreManifest) error {
	if m.StageState == "deleted" {
		return nil
	}
	stage, err := p.openStage(parent, m)
	if os.IsNotExist(err) && m.Phase == "committed" && m.StageState == "deleting" {
		return nil
	}
	if err != nil {
		return err
	}
	defer stage.Close()
	if err = freshKnownEntries(stage, map[string]bool{freshTicketName: false, "snapshot": false}); err != nil {
		return err
	}
	snapshot, err := inspectPurgeArtifact(stage, "snapshot")
	if err != nil {
		return err
	}
	if snapshot.State != "absent" && (m.OutputID == "" || snapshot.ID != m.OutputID) {
		return errors.New("unidentified fresh output retained")
	}
	marker, err := readFreshControl(stage, freshTicketName, maxFreshOwnership)
	if os.IsNotExist(err) && m.Phase == "committed" && m.StageState == "deleting" {
		return nil
	}
	if os.IsNotExist(err) && m.Phase == "copying" && m.OutputID == "" && snapshot.State == "absent" {
		temporary, e := inspectPurgeArtifact(op, freshTicketTemporary)
		if e == nil && temporary.ID == p.owner.TicketID {
			return nil
		}
		return errors.Join(e, errors.New("fresh preparation ticket is not identified"))
	}
	if err != nil {
		return err
	}
	_, body, err := p.expectedTicket()
	if err != nil || marker.ManifestID != p.owner.TicketID || !bytes.Equal(body, marker.Manifest) {
		return errors.Join(err, errors.New("fresh stage ticket changed"))
	}
	return nil
}
