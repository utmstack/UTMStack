package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const maxRetirementInspection = 16 << 20
const maxRetirementRead = 64 << 20

// RetirementControl identifies bounded private metadata without embedding every
// historical approval again. An absent object and an empty file are distinct.
type RetirementControl struct {
	Status string `json:"status"`
	FileID string `json:"file_id,omitempty"`
	Bytes  int64  `json:"bytes,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

type RetirementStage struct {
	Path        []byte `json:"path"`
	DirectoryID string `json:"directory_id"`
}

type RetirementFreshOperation struct {
	OperationID  string            `json:"operation_id"`
	DirectoryID  string            `json:"directory_id"`
	ArchiveID    string            `json:"archive_id,omitempty"`
	State        string            `json:"state"`
	StageID      string            `json:"stage_id,omitempty"`
	TicketID     string            `json:"ticket_id,omitempty"`
	Approval     RetirementControl `json:"approval"`
	Ownership    RetirementControl `json:"ownership"`
	Journal      RetirementControl `json:"journal"`
	Cancellation RetirementControl `json:"cancellation"`
}

// CancelledRetirementInspection is evidence, not permission to move a fence,
// publish an output, delete a backup or reopen SQLite. Incomplete histories
// return the partial report and an error; they never disappear from review.
type CancelledRetirementInspection struct {
	Version            int                        `json:"version"`
	Kind               string                     `json:"kind"`
	QuarantineID       string                     `json:"quarantine_id"`
	HoldID             string                     `json:"hold_id"`
	ObservedAt         time.Time                  `json:"observed_at"`
	RootID             string                     `json:"root_id"`
	RestoreQueueID     string                     `json:"restore_queue_id"`
	HistoryDirectoryID string                     `json:"history_directory_id"`
	FreshQueueID       string                     `json:"fresh_queue_id,omitempty"`
	FreshItemID        string                     `json:"fresh_item_id,omitempty"`
	Hold               JournalInspection          `json:"hold"`
	Cancellation       JournalInspection          `json:"cancellation"`
	Original           JournalInspection          `json:"original"`
	Snapshot           FreshRestoreInput          `json:"snapshot"`
	Raw                FreshRestoreInput          `json:"raw"`
	OldStage           RetirementStage            `json:"old_stage"`
	Fresh              []RetirementFreshOperation `json:"fresh"`
}

// InspectCancelledRetirement opens existing storage only: no database, lock,
// log, hardening, replay or creation. The explicit stage path must name the same
// directory recorded by the original journal, including after a stage rename.
func InspectCancelledRetirement(dir, id, holdID, oldStagePath string) (r CancelledRetirementInspection, err error) {
	r = CancelledRetirementInspection{Version: 1, Kind: "cancelled-retirement-inspection", QuarantineID: id, HoldID: holdID, ObservedAt: time.Now().UTC(), Fresh: []RetirementFreshOperation{}}
	if !ordinaryQuarantineID(id) || !validReviewUUID(holdID) || freshLocalDestination(oldStagePath) != nil {
		return r, errors.New("retirement inspection requires exact IDs and an ordinary absolute retained-stage path")
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return r, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return r, err
	}
	if err = freshLocalDestination(filepath.Join(resolved, "retirement-root-probe")); err != nil {
		return r, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return r, err
	}
	defer root.Close()
	return inspectCancelledRetirement(root, r, oldStagePath)
}

func retirementPin(e JournalInspection) RetirementControl {
	return RetirementControl{Status: e.Status, FileID: e.ManifestID, Bytes: int64(len(e.Manifest)), SHA256: e.SHA256}
}

func inspectCancelledRetirement(root *restoreDirectory, r CancelledRetirementInspection, oldStagePath string) (CancelledRetirementInspection, error) {
	var err error
	if r.RootID, err = admissionFileID(root.file); err != nil {
		return r, err
	}
	if r.Hold, err = inspectJournal(root, ".purges", r.QuarantineID); err != nil {
		return r, err
	}
	if r.Hold.Status != "captured" || r.Hold.EntryKind != "file" {
		return r, errors.New("retirement requires the unchanged active original fence")
	}
	var hold RecoveryHold
	if err = json.Unmarshal(r.Hold.Manifest, &hold); err != nil || hold.DecisionID != r.HoldID {
		return r, errors.Join(err, errors.New("retirement hold UUID changed"))
	}
	a := cancellationApproval(hold, r.Hold)
	if err = validRecoveryReleaseApproval(a); err != nil {
		return r, err
	}
	queue, err := admissionOpenChild(root, ".restores")
	if err != nil {
		return r, err
	}
	defer queue.Close()
	if r.RestoreQueueID, err = admissionFileID(queue.file); err != nil {
		return r, err
	}
	if err = requirePurgeNameAbsent(queue, r.QuarantineID); err != nil {
		return r, errors.Join(err, errors.New("finish or separately cancel the live fresh operation before retirement"))
	}
	// Reserve one compatibility slot without changing the queue. Permanent
	// retirement guards are never silently removed to reclaim this quota.
	if err = freshQueueCapacity(queue); err != nil {
		return r, err
	}
	history, cancellation, err := openCancellationProof(root, r.QuarantineID, a)
	if err != nil {
		return r, err
	}
	defer history.Close()
	archived, err := cancellationOriginal(root, history, cancellation)
	if err != nil || !archived {
		return r, errors.Join(err, errors.New("original cancellation is not completely archived"))
	}
	r.HistoryDirectoryID = cancellation.DirectoryID
	r.Cancellation, err = readFreshControl(history, "manifest.json", maxRecoveryRelease)
	if err != nil {
		return r, err
	}
	r.Cancellation.DirectoryID = r.HistoryDirectoryID
	r.Original = hold.Original
	original, err := decodeRestoreManifest(r.Original.Manifest, r.QuarantineID)
	if err != nil || original.Version != 1 || original.Source.State != "ready" || !freshBoundID(original.StageID) || original.StageState == "planned" || original.StageState == "deleted" {
		return r, errors.Join(err, errors.New("retirement needs a proven retained original snapshot and stage"))
	}
	stage, err := openAdmissionRoot(oldStagePath, false)
	if err != nil {
		return r, err
	}
	defer stage.Close()
	stageID, err := admissionFileID(stage.file)
	if err != nil || stageID != original.StageID || stageID == r.RootID || stageID == original.ParentID {
		return r, errors.Join(err, errors.New("retained original stage identity changed"))
	}
	r.OldStage = RetirementStage{Path: []byte(stage.path), DirectoryID: stageID}
	input, err := restoreOpenInput(root, r.QuarantineID+".quarantined")
	if err != nil {
		return r, err
	}
	r.Snapshot, err = inspectFreshInput(input)
	err = errors.Join(err, input.Close())
	if err != nil || r.Snapshot.FileID != original.Source.ID || r.Snapshot.SHA256 != original.Record.SHA256 || r.Snapshot.FileID == original.Record.SourceFileID || r.Snapshot.FileID == original.OutputID {
		return r, errors.Join(err, errors.New("independent retained snapshot identity or bytes changed"))
	}
	if !filepath.IsAbs(original.Record.SourceCaptureDir) {
		return r, errors.New("original raw capture path is unavailable")
	}
	capture, err := openAdmissionRoot(original.Record.SourceCaptureDir, false)
	if err != nil {
		return r, err
	}
	defer capture.Close()
	raw, err := restoreOpenInput(capture, "source")
	if err != nil {
		return r, err
	}
	r.Raw, err = inspectFreshInput(raw)
	err = errors.Join(err, raw.Close())
	if err != nil || r.Raw.FileID != original.Record.SourceFileID || r.Raw.FileID == r.Snapshot.FileID {
		return r, errors.Join(err, errors.New("original raw capture identity changed"))
	}
	// The raw may legitimately differ after an earlier retained writer; it is
	// independently pinned as observed evidence, never used for publication.
	budget := int64(maxRetirementRead - len(r.Hold.Manifest) - len(r.Cancellation.Manifest) - len(r.Original.Manifest))
	if err = inspectRetirementFresh(root, &r, &budget); err != nil {
		return r, err
	}
	if err = requirePurgeNameAbsent(queue, r.QuarantineID); err == nil {
		err = cancellationFence(root, cancellation)
	}
	if err == nil {
		_, err = cancellationOriginal(root, history, cancellation)
	}
	if err == nil {
		err = stage.stillNamed()
	}
	if err == nil {
		err = root.stillNamed()
	}
	if err != nil {
		return r, err
	}
	body, err := json.Marshal(r)
	if err != nil || len(body) > maxRetirementInspection {
		return r, errors.Join(err, errors.New("complete retirement inspection exceeds its bound"))
	}
	return r, nil
}

func inspectRetirementFresh(root *restoreDirectory, r *CancelledRetirementInspection, budget *int64) error {
	queue, err := admissionOpenChild(root, ".fresh-restores")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer queue.Close()
	if r.FreshQueueID, err = admissionFileID(queue.file); err != nil {
		return err
	}
	item, err := admissionOpenChild(queue, r.QuarantineID)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer item.Close()
	if r.FreshItemID, err = admissionFileID(item.file); err != nil {
		return err
	}
	entries, err := item.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > maxAdmissionRecovery {
		return errors.Join(err, errors.New("retirement fresh history exceeds its bounded item directory"))
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if !entry.IsDir() || !validReviewUUID(entry.Name()) {
			return errors.New("unknown fresh history entry retained")
		}
		record, e := inspectRetirementFreshOperation(root, queue, item, r, entry.Name(), budget)
		r.Fresh = append(r.Fresh, record)
		if e != nil {
			return e
		}
	}
	return item.stillNamed()
}

func inspectRetirementFreshOperation(root, queue, item *restoreDirectory, r *CancelledRetirementInspection, operationID string, budget *int64) (f RetirementFreshOperation, err error) {
	f.OperationID = operationID
	f.Ownership = RetirementControl{Status: "absent"}
	f.Journal = RetirementControl{Status: "absent"}
	f.Cancellation = RetirementControl{Status: "absent"}
	op, err := admissionOpenChild(item, operationID)
	if err != nil {
		return f, err
	}
	defer op.Close()
	if f.DirectoryID, err = admissionFileID(op.file); err != nil {
		return f, err
	}
	read := func(d *restoreDirectory, name string, limit int64) (JournalInspection, error) {
		if *budget < limit {
			limit = *budget
		}
		if limit <= 0 {
			return JournalInspection{}, errors.New("retirement metadata read budget exhausted")
		}
		e, e1 := freshOptionalControl(d, name, limit)
		*budget -= int64(len(e.Manifest))
		return e, e1
	}
	approval, err := read(op, "manifest.json", maxFreshRestoreApproval)
	f.Approval = retirementPin(approval)
	if err != nil || approval.Status != "captured" {
		return f, errors.Join(err, errors.New("fresh approval is incomplete"))
	}
	a, err := decodeFreshApproval(root, queue, item, op, r.QuarantineID, operationID, approval)
	if err != nil {
		return f, err
	}
	approvalHash, canonical, err := freshJSONHash(a)
	if err != nil || !bytes.Equal(approval.Manifest, canonical) || !bytes.Equal(a.Inspection.Hold.Manifest, r.Hold.Manifest) || a.Inspection.Cancellation.ManifestID != r.Cancellation.ManifestID || a.Inspection.Cancellation.SHA256 != r.Cancellation.SHA256 {
		return f, errors.Join(err, errors.New("fresh approval belongs to different original cancellation evidence"))
	}
	owner, err := read(op, "ownership.json", maxFreshOwnership)
	f.Ownership = retirementPin(owner)
	if err != nil {
		return f, err
	}
	cancel, err := read(op, "cancellation.json", maxFreshCancellation)
	f.Cancellation = retirementPin(cancel)
	if err != nil {
		return f, err
	}
	archive, err := admissionOpenChild(op, "restore")
	if os.IsNotExist(err) {
		if owner.Status != "absent" || cancel.Status != "absent" {
			return f, errors.New("incomplete fresh history has no exact archive")
		}
		f.State = "approval-only"
		return f, freshKnownEntries(op, map[string]bool{"manifest.json": false})
	}
	if err != nil {
		return f, err
	}
	defer archive.Close()
	if f.ArchiveID, err = admissionFileID(archive.file); err != nil {
		return f, err
	}
	journal, err := read(archive, "manifest.json", maxAdmissionJournal)
	f.Journal = retirementPin(journal)
	if err != nil {
		return f, err
	}
	o, err := decodeFreshOwnership(op, a, approvalHash, owner)
	if err != nil || o.OperationDirectoryID != f.ArchiveID {
		return f, errors.Join(err, errors.New("archived fresh stage ownership is not provable; keep original fence"))
	}
	f.StageID, f.TicketID = o.StageID, o.TicketID
	if cancel.Status == "captured" {
		p := &freshRestorePermit{root: root, decision: op, approval: a, approvalHash: approvalHash}
		m, e := decodeFreshCancellationRecord(p, cancel)
		if e != nil {
			return f, e
		}
		if m.Inspection.OperationDirectoryID != f.ArchiveID || retirementPin(m.Inspection.Approval) != f.Approval || retirementPin(m.Inspection.Ownership) != f.Ownership || retirementPin(m.Inspection.Journal) != f.Journal || !bytes.Equal(m.Inspection.Approval.Manifest, approval.Manifest) || !bytes.Equal(m.Inspection.Ownership.Manifest, owner.Manifest) || !bytes.Equal(m.Inspection.Journal.Manifest, journal.Manifest) {
			return f, errors.New("fresh cancellation archived metadata differs from the saved exact inspection")
		}
		f.State = "cancelled"
	} else {
		m, e := decodeRestoreManifest(journal.Manifest, r.QuarantineID)
		if e != nil {
			return f, e
		}
		if e = validateFreshJournal(m, a, o); e != nil {
			return f, e
		}
		if m.Phase != "committed" || m.StageState != "deleted" {
			return f, errors.New("fresh publication is not completely archived and cleaned")
		}
		f.State = "committed"
	}
	if err = freshKnownEntries(op, map[string]bool{"manifest.json": false, "ownership.json": false, "cancellation.json": false, "restore": true}); err != nil {
		return f, err
	}
	return f, archive.stillNamed()
}

func sameCancelledRetirementInspection(a, b CancelledRetirementInspection, ignoreObservation bool) bool {
	if ignoreObservation {
		a.ObservedAt, b.ObservedAt = time.Time{}, time.Time{}
	}
	aa, e1 := json.Marshal(a)
	bb, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(aa, bb)
}
