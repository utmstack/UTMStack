package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"time"

	"github.com/google/uuid"
)

const maxFreshCancellation = 64 << 20

// FreshRestoreCancellation records a separate decision to retain and archive
// one incomplete new operation. It is not a publication, cleanup, database or
// original-fence release receipt.
type FreshRestoreCancellation struct {
	Version      int                         `json:"version"`
	Kind         string                      `json:"kind"`
	DecisionID   string                      `json:"decision_id"`
	RootID       string                      `json:"root_id"`
	QuarantineID string                      `json:"quarantine_id"`
	OperationID  string                      `json:"operation_id"`
	DirectoryID  string                      `json:"directory_id"`
	FileID       string                      `json:"file_id"`
	ApprovedAt   time.Time                   `json:"approved_at"`
	ActorID      string                      `json:"actor_id"`
	Reason       string                      `json:"reason"`
	Inspection   FreshCancellationInspection `json:"inspection"`
}

func CancelFreshRestore(dir, id, operationID string, inspection FreshCancellationInspection, reason string) (FreshRestoreCancellation, error) {
	return cancelFreshRestore(dir, id, operationID, inspection, reason, nil)
}

func cancelFreshRestore(dir, id, operationID string, inspection FreshCancellationInspection, reason string, fault func(string) error) (m FreshRestoreCancellation, err error) {
	if !validReviewReason(reason) || inspection.Version != 1 || inspection.Kind != "fresh-cancellation-inspection" || inspection.QuarantineID != id || inspection.OperationID != operationID || inspection.ObservedAt.IsZero() || !freshBoundID(inspection.OperationDirectoryID) || len(inspection.Approval.Manifest) > maxFreshRestoreApproval || len(inspection.Ownership.Manifest) > maxFreshOwnership || len(inspection.Journal.Manifest) > maxAdmissionJournal {
		return m, errors.New("fresh cancellation requires the exact bounded saved inspection and reason")
	}
	actor, e := user.Current()
	if e != nil || actor.Uid == "" || len(actor.Uid) > 256 {
		return m, errors.Join(e, errors.New("cannot determine fresh cancellation actor"))
	}
	p, closeProof, e := openFreshRetentionProof(dir, id, operationID)
	if e != nil {
		return m, e
	}
	defer closeProof()
	if inspection.RootID != p.approval.RootID {
		return m, errors.New("fresh cancellation root changed")
	}
	unlock, e := p.root.lockItem(id)
	if e != nil {
		return m, e
	}
	defer unlock()
	seam := func(at string) error {
		if fault != nil {
			return fault("fresh-cancellation-" + at)
		}
		return nil
	}
	entry, e := readFreshControl(p.decision, "cancellation.json", maxFreshCancellation)
	if e == nil {
		m, err = decodeFreshCancellation(p, entry)
		if err != nil {
			return m, err
		}
		if m.ActorID != actor.Uid || m.Reason != reason || !sameFreshCancellationInspection(m.Inspection, inspection, false) {
			return m, errors.New("existing fresh cancellation actor, reason or inspection changed")
		}
		return finishFreshCancellation(p, m, seam)
	}
	if !os.IsNotExist(e) {
		return m, e
	}
	current, e := inspectFreshCancellation(p, FreshCancellationInspection{Version: 1, Kind: "fresh-cancellation-inspection", QuarantineID: id, OperationID: operationID, ObservedAt: time.Now().UTC()})
	if e != nil || !sameFreshCancellationInspection(current, inspection, true) {
		return m, errors.Join(e, errors.New("fresh cancellation inspection is stale; evidence retained"))
	}
	f, e := admissionCreateFile(p.decision, "cancellation.json")
	if e != nil {
		return m, e
	}
	defer f.Close()
	m = FreshRestoreCancellation{Version: 1, Kind: "fresh-restore-cancellation", DecisionID: uuid.NewString(), RootID: p.approval.RootID, QuarantineID: id, OperationID: operationID, DirectoryID: p.approval.DirectoryID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Inspection: inspection}
	if m.FileID, e = admissionFileID(f); e != nil {
		return m, e
	}
	if err = seam("receipt-created"); err != nil {
		return m, err
	}
	body, e := json.Marshal(m)
	if e != nil || len(body) > maxFreshCancellation {
		return m, errors.Join(e, errors.New("fresh cancellation exceeds metadata bound"))
	}
	if _, e = f.Write(body); e != nil {
		return m, e
	}
	if err = seam("receipt-written"); err != nil {
		return m, err
	}
	if e = f.Sync(); e == nil {
		e = admissionSyncDirectory(p.decision)
	}
	if e != nil {
		return m, e
	}
	if e = f.Close(); e != nil {
		return m, e
	}
	if err = seam("approved"); err != nil {
		return m, err
	}
	return finishFreshCancellation(p, m, seam)
}

func decodeFreshCancellation(p *freshRestorePermit, entry JournalInspection) (m FreshRestoreCancellation, err error) {
	m, err = decodeFreshCancellationRecord(p, entry)
	if err != nil {
		return m, err
	}
	return m, checkFreshRetentionPins(p, m.Inspection)
}

func decodeFreshCancellationRecord(p *freshRestorePermit, entry JournalInspection) (m FreshRestoreCancellation, err error) {
	if err = json.Unmarshal(entry.Manifest, &m); err != nil {
		return m, err
	}
	a := p.approval
	if m.Version != 1 || m.Kind != "fresh-restore-cancellation" || !validReviewUUID(m.DecisionID) || m.RootID != a.RootID || m.QuarantineID != a.QuarantineID || m.OperationID != a.OperationID || m.DirectoryID != a.DirectoryID || m.FileID != entry.ManifestID || m.ApprovedAt.IsZero() || m.ActorID == "" || len(m.ActorID) > 256 || !validReviewReason(m.Reason) || m.Inspection.Version != 1 || m.Inspection.Kind != "fresh-cancellation-inspection" || m.Inspection.RootID != a.RootID || m.Inspection.QuarantineID != a.QuarantineID || m.Inspection.OperationID != a.OperationID || m.Inspection.ObservedAt.IsZero() || !freshBoundID(m.Inspection.OperationDirectoryID) {
		return m, errors.New("fresh cancellation identity changed")
	}
	_, expected, err := freshJSONHash(m)
	if err != nil || !bytes.Equal(expected, entry.Manifest) {
		return m, errors.Join(err, errors.New("fresh cancellation bytes changed"))
	}
	return m, nil
}

func checkFreshRetentionPins(p *freshRestorePermit, r FreshCancellationInspection) error {
	if _, err := p.checkAuthority(); err != nil {
		return err
	}
	a, err := readFreshControl(p.decision, "manifest.json", maxFreshRestoreApproval)
	if err != nil {
		return err
	}
	a.DirectoryID = p.approval.DirectoryID
	owner, err := freshOptionalControl(p.decision, "ownership.json", maxFreshOwnership)
	if err != nil {
		return err
	}
	before := r
	before.Approval = a
	before.Ownership = owner
	if !sameFreshCancellationInspection(before, r, false) {
		return errors.New("fresh cancellation permission or ownership bytes changed")
	}
	return nil
}

func syncFreshCancellation(p *freshRestorePermit, m FreshRestoreCancellation) error {
	f, err := openFreshWritableFile(p.decision, "cancellation.json")
	if err != nil {
		return err
	}
	defer f.Close()
	entry, err := inspectJournalFile(JournalInspection{}, f, maxFreshCancellation)
	_, body, encodeErr := freshJSONHash(m)
	if err != nil || encodeErr != nil || entry.ManifestID != m.FileID || !bytes.Equal(entry.Manifest, body) {
		return errors.Join(err, encodeErr, errors.New("fresh cancellation changed before flush"))
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return admissionSyncDirectory(p.decision)
}

// An unrelated later live operation cannot invalidate a completed historical
// move. No child of that unrelated directory is opened or interpreted here.
func freshCancellationLocation(p *freshRestorePermit, r FreshCancellationInspection) (bool, error) {
	var live, saved *restoreDirectory
	var liveID, savedID string
	queue, op, err := openRestoreOperation(p.root, p.approval.QuarantineID)
	if err == nil {
		queue.Close()
		live = op
		defer live.Close()
		liveID, err = admissionFileID(live.file)
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	saved, err = admissionOpenChild(p.decision, "restore")
	if err == nil {
		defer saved.Close()
		savedID, err = admissionFileID(saved.file)
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	archived := savedID == r.OperationDirectoryID && liveID != r.OperationDirectoryID
	if !archived && !(liveID == r.OperationDirectoryID && saved == nil) {
		return false, errors.New("fresh cancelled operation is absent or ambiguous")
	}
	selected := live
	if archived {
		selected = saved
	}
	entry, err := freshOptionalControl(selected, "manifest.json", maxAdmissionJournal)
	if err != nil {
		return false, err
	}
	observed := r
	observed.Journal = entry
	if !sameFreshCancellationInspection(observed, r, false) {
		return false, errors.New("fresh cancelled operation journal changed")
	}
	return archived, nil
}

func finishFreshCancellation(p *freshRestorePermit, m FreshRestoreCancellation, seam func(string) error) (FreshRestoreCancellation, error) {
	if err := checkFreshRetentionPins(p, m.Inspection); err != nil {
		return m, err
	}
	archived, err := freshCancellationLocation(p, m.Inspection)
	if err != nil {
		return m, err
	}
	if err = syncFreshCancellation(p, m); err != nil {
		return m, err
	}
	if !archived {
		if err = seam("before-move"); err != nil {
			return m, err
		}
		if err = checkFreshRetentionPins(p, m.Inspection); err != nil {
			return m, err
		}
		if archived, err = freshCancellationLocation(p, m.Inspection); err != nil || archived {
			return m, errors.Join(err, errors.New("fresh operation moved before reviewed cancellation"))
		}
		queue, e := admissionOpenChild(p.root, ".restores")
		if e != nil {
			return m, e
		}
		err = moveRecoveryDirectory(queue, m.QuarantineID, p.decision, "restore", m.Inspection.OperationDirectoryID)
		queue.Close()
		if err != nil {
			return m, err
		}
		if err = seam("archived"); err != nil {
			return m, err
		}
	}
	if err = checkFreshRetentionPins(p, m.Inspection); err != nil {
		return m, err
	}
	if archived, err = freshCancellationLocation(p, m.Inspection); err != nil || !archived {
		return m, errors.Join(err, errors.New("fresh cancellation archive not proven"))
	}
	queue, err := admissionOpenChild(p.root, ".restores")
	if err != nil {
		return m, err
	}
	err = admissionSyncDirectory(queue)
	err = errors.Join(err, queue.Close())
	if err == nil {
		err = admissionSyncDirectory(p.decision)
	}
	if err != nil {
		return m, err
	}
	return m, seam("synced")
}
