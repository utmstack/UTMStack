package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// FreshCancellationInspection captures an incomplete new operation for a
// separate retention decision. Raw partial metadata is evidence, never proof
// that copying or publication completed.
type FreshCancellationInspection struct {
	Version              int               `json:"version"`
	Kind                 string            `json:"kind"`
	QuarantineID         string            `json:"quarantine_id"`
	OperationID          string            `json:"operation_id"`
	RootID               string            `json:"root_id"`
	ObservedAt           time.Time         `json:"observed_at"`
	Approval             JournalInspection `json:"approval"`
	Ownership            JournalInspection `json:"ownership"`
	OperationDirectoryID string            `json:"operation_directory_id"`
	Journal              JournalInspection `json:"journal"`
}

// InspectFreshCancellation reads existing private objects only. It creates no
// lock or receipt, opens no database and never examines or changes a destination.
func InspectFreshCancellation(dir, id, operationID string) (FreshCancellationInspection, error) {
	r := FreshCancellationInspection{Version: 1, Kind: "fresh-cancellation-inspection", QuarantineID: id, OperationID: operationID, ObservedAt: time.Now().UTC()}
	p, closeProof, err := openFreshRetentionProof(dir, id, operationID)
	if err != nil {
		return r, err
	}
	defer closeProof()
	return inspectFreshCancellation(p, r)
}

// The cancellation actor is a new independent decision maker. Opening this
// retention proof deliberately does not impersonate or reuse the approval actor.
func openFreshRetentionProof(dir, id, operationID string) (*freshRestorePermit, func(), error) {
	if !ordinaryQuarantineID(id) || !validReviewUUID(operationID) {
		return nil, nil, errors.New("fresh retention requires an exact quarantine ID and operation UUID")
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, nil, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, nil, err
	}
	if err = freshLocalDestination(filepath.Join(resolved, "fresh-root-probe")); err != nil {
		return nil, nil, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return nil, nil, err
	}
	queue, item, decision, err := openFreshApproval(root, id, operationID)
	if err != nil {
		root.Close()
		return nil, nil, err
	}
	closeProof := func() { decision.Close(); item.Close(); queue.Close(); root.Close() }
	a, err := readFreshApproval(root, queue, item, decision, id, operationID)
	if err != nil {
		closeProof()
		return nil, nil, err
	}
	hash, _, err := freshJSONHash(a)
	if err != nil {
		closeProof()
		return nil, nil, err
	}
	p := &freshRestorePermit{root: root, decision: decision, approval: a, approvalHash: hash}
	if _, err = p.checkAuthority(); err != nil {
		closeProof()
		return nil, nil, err
	}
	return p, closeProof, nil
}

func freshOptionalControl(d *restoreDirectory, name string, limit int64) (JournalInspection, error) {
	e, err := readFreshControl(d, name, limit)
	if os.IsNotExist(err) {
		return JournalInspection{Status: "absent"}, nil
	}
	// Canonical empty payloads must survive receipt JSON round trips without
	// changing the public lost-return result from an empty slice to nil.
	if len(e.Manifest) == 0 {
		e.Manifest = nil
	}
	return e, err
}

func inspectFreshCancellation(p *freshRestorePermit, r FreshCancellationInspection) (FreshCancellationInspection, error) {
	a := p.approval
	r.RootID = a.RootID
	var err error
	if r.Approval, err = readFreshControl(p.decision, "manifest.json", maxFreshRestoreApproval); err != nil {
		return r, err
	}
	r.Approval.DirectoryID = a.DirectoryID
	if r.Approval.ManifestID != a.FileID || r.Approval.SHA256 != p.approvalHash {
		return r, errors.New("fresh retention permission changed")
	}
	if r.Ownership, err = freshOptionalControl(p.decision, "ownership.json", maxFreshOwnership); err != nil {
		return r, err
	}
	queue, op, err := openRestoreOperation(p.root, a.QuarantineID)
	if err != nil {
		return r, err
	}
	defer queue.Close()
	defer op.Close()
	if r.OperationDirectoryID, err = admissionFileID(op.file); err != nil {
		return r, err
	}
	if r.Journal, err = freshOptionalControl(op, "manifest.json", maxAdmissionJournal); err != nil {
		return r, err
	}
	// A recognizable different operation is never swept into this fresh history.
	// Missing/malformed preparation bytes can be retained by exact directory ID.
	if r.Journal.Status == "captured" {
		var intent struct {
			Version              int
			OperationID, FreshID string
		}
		if json.Unmarshal(r.Journal.Manifest, &intent) == nil && (intent.Version != 0 && intent.Version != 2 || intent.OperationID != "" && intent.OperationID != a.OperationID || intent.FreshID != "" && intent.FreshID != a.OperationID) {
			return r, errors.New("active journal names a different restore operation")
		}
	}
	if owner, _, e := readFreshOwnership(p.decision, a, p.approvalHash); e == nil && owner.OperationDirectoryID != r.OperationDirectoryID {
		return r, errors.New("fresh owner binds a different active directory")
	}
	archived, e := admissionOpenChild(p.decision, "restore")
	if e == nil {
		archived.Close()
		return r, errors.New("existing fresh archive retained; no new cancellation inspection")
	}
	if !os.IsNotExist(e) {
		return r, e
	}
	if err = op.stillNamed(); err == nil {
		_, err = p.checkAuthority()
	}
	return r, err
}

func sameFreshCancellationInspection(a, b FreshCancellationInspection, ignoreObservation bool) bool {
	if ignoreObservation {
		a.ObservedAt = time.Time{}
		b.ObservedAt = time.Time{}
	}
	aa, e1 := json.Marshal(a)
	bb, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(aa, bb)
}
