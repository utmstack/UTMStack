package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

const maxFreshRestoreApproval = 24 << 20

// FreshRestoreApproval authorizes one new operation using independently
// reviewed evidence. It does not release the original hold, prove an old
// publication, or itself copy, publish, retire a backup, or open SQLite.
type FreshRestoreApproval struct {
	Version      int                    `json:"version"`
	Kind         string                 `json:"kind"`
	OperationID  string                 `json:"operation_id"`
	QuarantineID string                 `json:"quarantine_id"`
	RootID       string                 `json:"root_id"`
	QueueID      string                 `json:"queue_id"`
	ItemID       string                 `json:"item_id"`
	DirectoryID  string                 `json:"directory_id"`
	FileID       string                 `json:"file_id"`
	ApprovedAt   time.Time              `json:"approved_at"`
	ActorID      string                 `json:"actor_id"`
	Reason       string                 `json:"reason"`
	Inspection   FreshRestoreInspection `json:"inspection"`
}

// ApproveFreshRestore requires a caller-selected new UUID and an independent
// saved inspection. Exact retries return the same historical approval; all
// execution will still require live revalidation under the quarantine item lock.
func ApproveFreshRestore(dir, id, operationID string, inspection FreshRestoreInspection, reason string) (FreshRestoreApproval, error) {
	return approveFreshRestore(dir, id, operationID, inspection, reason, nil)
}

func approveFreshRestore(dir, id, operationID string, inspection FreshRestoreInspection, reason string, fault func(string) error) (m FreshRestoreApproval, err error) {
	if !ordinaryQuarantineID(id) || !validReviewUUID(operationID) || !validReviewReason(reason) {
		return m, errors.New("fresh approval requires an exact quarantine ID, new operation UUID and bounded reason")
	}
	holdID, err := validateFreshInspection(inspection, id)
	if err != nil {
		return m, err
	}
	original, err := decodeRestoreManifest(inspection.Original.Manifest, id)
	if err != nil || operationID == original.OperationID {
		return m, errors.Join(err, errors.New("fresh approval must use a new operation UUID"))
	}
	actor, err := user.Current()
	if err != nil || actor.Uid == "" || len(actor.Uid) > 256 {
		return m, errors.Join(err, errors.New("cannot determine fresh approval actor"))
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return m, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return m, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return m, err
	}
	defer root.Close()
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != inspection.RootID {
		return m, errors.Join(err, errors.New("fresh approval root identity changed"))
	}
	unlock, err := root.lockItem(id)
	if err != nil {
		return m, err
	}
	defer unlock()
	queue, item, op, err := openFreshApproval(root, id, operationID)
	if err == nil {
		defer queue.Close()
		defer item.Close()
		defer op.Close()
		m, err = readFreshApproval(root, queue, item, op, id, operationID)
		if err != nil {
			return m, err
		}
		if !sameFreshInspection(m.Inspection, inspection, false) || m.Reason != reason || m.ActorID != actor.Uid {
			return m, errors.New("existing fresh approval actor, reason or saved inspection changed")
		}
		// Re-flush the exact complete receipt after a lost return or process
		// exit. This cannot authorize a different operation or external action.
		return m, syncFreshApproval(op, m)
	}
	if !os.IsNotExist(err) {
		return m, err
	}
	current, err := inspectFreshRestore(root, id, holdID, filepath.Join(string(inspection.Destination.ParentPath), string(inspection.Destination.Leaf)), FreshRestoreInspection{Version: 1, Kind: "fresh-restore-inspection", QuarantineID: id, ObservedAt: time.Now().UTC()})
	if err != nil || !sameFreshInspection(current, inspection, true) {
		return m, errors.Join(err, errors.New("fresh approval inspection is stale; nothing was approved"))
	}
	// No external or original operation is created here. Separate bounded
	// control storage keeps cancellation's immutable history unchanged.
	queue, err = admissionChild(root, ".fresh-restores")
	if err != nil {
		return m, err
	}
	defer queue.Close()
	if err = freshQueueCapacity(queue); err != nil {
		return m, err
	}
	item, err = admissionChild(queue, id)
	if err != nil {
		return m, err
	}
	defer item.Close()
	if err = freshQueueCapacity(item); err != nil {
		return m, err
	}
	op, err = admissionMakeDirectory(item, operationID)
	if err != nil {
		return m, err
	}
	defer op.Close()
	m = FreshRestoreApproval{Version: 1, Kind: "fresh-restore-approval", OperationID: operationID, QuarantineID: id, RootID: rootID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Inspection: inspection}
	if m.QueueID, err = admissionFileID(queue.file); err == nil {
		m.ItemID, err = admissionFileID(item.file)
	}
	if err == nil {
		m.DirectoryID, err = admissionFileID(op.file)
	}
	if err != nil {
		return m, err
	}
	if err = admissionSyncDirectory(queue); err == nil {
		err = admissionSyncDirectory(item)
	}
	if err == nil {
		err = admissionSyncDirectory(root)
	}
	if err != nil {
		return m, err
	}
	seam := func(phase string) error {
		if fault != nil {
			return fault(phase)
		}
		return nil
	}
	if err = seam("fresh-approval-directory-created"); err != nil {
		return m, err
	}
	f, err := admissionCreateFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	defer f.Close()
	if m.FileID, err = admissionFileID(f); err != nil {
		return m, err
	}
	body, err := json.Marshal(m)
	if err != nil || len(body) > maxFreshRestoreApproval {
		return m, errors.Join(err, errors.New("fresh approval exceeds its metadata limit"))
	}
	if _, err = f.Write(body); err != nil {
		return m, err
	}
	if err = seam("fresh-approval-written"); err != nil {
		return m, err
	}
	if err = f.Sync(); err == nil {
		err = admissionSyncDirectory(op)
	}
	if err != nil {
		return m, err
	}
	return m, seam("fresh-approval-synced")
}

func sameFreshInspection(a, b FreshRestoreInspection, ignoreObservation bool) bool {
	if ignoreObservation {
		a.ObservedAt, b.ObservedAt = time.Time{}, time.Time{}
	}
	aa, e1 := json.Marshal(a)
	bb, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(aa, bb)
}

func validateFreshInspection(r FreshRestoreInspection, id string) (string, error) {
	if r.Version != 1 || r.Kind != "fresh-restore-inspection" || r.QuarantineID != id || r.ObservedAt.IsZero() || r.RootID == "" || len(r.RootID) > 256 || r.Snapshot.FileID == "" || len(r.Snapshot.FileID) > 256 || r.Snapshot.Size < 0 || r.Snapshot.Size > maxRestoreBytes || r.Destination.ParentID == "" || len(r.Destination.ParentID) > 256 {
		return "", errors.New("invalid saved fresh inspection")
	}
	if _, err := restoreDigest(r.Snapshot.SHA256); err != nil {
		return "", err
	}
	destination := filepath.Join(string(r.Destination.ParentPath), string(r.Destination.Leaf))
	if err := freshLocalDestination(destination); err != nil || filepath.Dir(destination) != string(r.Destination.ParentPath) || filepath.Base(destination) != string(r.Destination.Leaf) {
		return "", errors.Join(err, errors.New("invalid saved fresh destination"))
	}
	body, err := json.Marshal(r)
	if err != nil || len(body) > maxFreshRestoreApproval-(64<<10) {
		return "", errors.Join(err, errors.New("saved fresh inspection exceeds its metadata limit"))
	}
	var hold RecoveryHold
	if err = json.Unmarshal(r.Hold.Manifest, &hold); err != nil || !validReviewUUID(hold.DecisionID) {
		return "", errors.Join(err, errors.New("saved fresh inspection lacks a valid hold UUID"))
	}
	return hold.DecisionID, nil
}

func freshQueueCapacity(d *restoreDirectory) error {
	entries, err := d.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) >= maxAdmissionRecovery {
		return errors.Join(err, errors.New("fresh approval queue is full; existing evidence retained"))
	}
	return nil
}

func openFreshApproval(root *restoreDirectory, id, operationID string) (*restoreDirectory, *restoreDirectory, *restoreDirectory, error) {
	queue, err := admissionOpenChild(root, ".fresh-restores")
	if err != nil {
		return nil, nil, nil, err
	}
	item, err := admissionOpenChild(queue, id)
	if err != nil {
		queue.Close()
		return nil, nil, nil, err
	}
	op, err := admissionOpenChild(item, operationID)
	if err != nil {
		item.Close()
		queue.Close()
		return nil, nil, nil, err
	}
	return queue, item, op, nil
}

func readFreshApproval(root, queue, item, op *restoreDirectory, id, operationID string) (m FreshRestoreApproval, err error) {
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	entry, err := inspectJournalFile(JournalInspection{}, f, maxFreshRestoreApproval)
	err = errors.Join(err, f.Close())
	if err != nil {
		return m, err
	}
	return decodeFreshApproval(root, queue, item, op, id, operationID, entry)
}

func decodeFreshApproval(root, queue, item, op *restoreDirectory, id, operationID string, entry JournalInspection) (m FreshRestoreApproval, err error) {
	if err = json.Unmarshal(entry.Manifest, &m); err != nil {
		return m, err
	}
	rootID, e1 := admissionFileID(root.file)
	queueID, e2 := admissionFileID(queue.file)
	itemID, e3 := admissionFileID(item.file)
	opID, e4 := admissionFileID(op.file)
	if err = errors.Join(e1, e2, e3, e4); err != nil {
		return m, err
	}
	if m.Version != 1 || m.Kind != "fresh-restore-approval" || m.OperationID != operationID || !validReviewUUID(m.OperationID) || m.QuarantineID != id || !ordinaryQuarantineID(id) || m.RootID != rootID || m.Inspection.RootID != rootID || m.QueueID != queueID || m.ItemID != itemID || m.DirectoryID != opID || m.FileID != entry.ManifestID || m.ApprovedAt.IsZero() || m.ActorID == "" || len(m.ActorID) > 256 || !validReviewReason(m.Reason) {
		return m, errors.New("fresh approval identity or actor changed; evidence retained")
	}
	_, err = validateFreshInspection(m.Inspection, id)
	return m, err
}

func syncFreshApproval(op *restoreDirectory, m FreshRestoreApproval) error {
	f, err := openRecoveryReleaseReceipt(op)
	if err != nil {
		return err
	}
	defer f.Close()
	entry, err := inspectJournalFile(JournalInspection{}, f, maxFreshRestoreApproval)
	expected, encodeErr := json.Marshal(m)
	if err != nil || encodeErr != nil || entry.ManifestID != m.FileID || !bytes.Equal(entry.Manifest, expected) {
		return errors.Join(err, encodeErr, errors.New("fresh approval changed before flush"))
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return admissionSyncDirectory(op)
}
