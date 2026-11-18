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

	"github.com/google/uuid"
)

// RecoveryCancellation approves archiving the original operation, never its
// publication or deletion. The same active hold remains in .purges throughout.
// The versioned receipt extends that hold's binding to exactly two locations of
// the same original directory. Old readers fail closed after the directory move.
type RecoveryCancellation struct {
	Version      int                     `json:"version"`
	Kind         string                  `json:"kind"`
	DecisionID   string                  `json:"decision_id"`
	QuarantineID string                  `json:"quarantine_id"`
	RootID       string                  `json:"root_id"`
	DirectoryID  string                  `json:"directory_id"`
	FileID       string                  `json:"file_id"`
	ApprovedAt   time.Time               `json:"approved_at"`
	ActorID      string                  `json:"actor_id"`
	Reason       string                  `json:"reason"`
	Approval     RecoveryReleaseApproval `json:"approval"`
	Hold         JournalInspection       `json:"hold"`
}

// CancelRecovery requires the exact original and hold identities used by hold
// release, but approves a different action. It never releases the fence, opens
// SQLite, replays recovery, touches the destination/stage or creates a new restore.
func CancelRecovery(dir, id string, approval RecoveryReleaseApproval, reason string) (RecoveryCancellation, error) {
	return cancelRecovery(dir, id, approval, reason, nil)
}

func cancelRecovery(dir, id string, approval RecoveryReleaseApproval, reason string, fault func(string) error) (m RecoveryCancellation, err error) {
	if !ordinaryQuarantineID(id) {
		return m, errors.New("invalid quarantine id")
	}
	if err = validRecoveryReleaseApproval(approval); err != nil {
		return m, err
	}
	if !validReviewReason(reason) {
		return m, errors.New("cancellation reason must contain at most 4096 UTF-8 bytes")
	}
	actor, e := user.Current()
	if e != nil || actor.Uid == "" || len(actor.Uid) > 256 {
		return m, errors.Join(e, errors.New("cannot determine the authorized process identity"))
	}
	absolute, e := filepath.Abs(dir)
	if e != nil {
		return m, e
	}
	resolved, e := filepath.EvalSymlinks(absolute)
	if e != nil {
		return m, e
	}
	root, e := openAdmissionRoot(resolved, false)
	if e != nil {
		return m, e
	}
	defer root.Close()
	rootID, e := admissionFileID(root.file)
	if e != nil || rootID != approval.Original.RootID {
		return m, errors.Join(e, errors.New("cancellation root identity changed"))
	}
	unlock, e := root.lockItem(id)
	if e != nil {
		return m, e
	}
	defer unlock()
	seam := func(at string) error {
		if fault != nil {
			return fault(at)
		}
		return nil
	}
	// The hold UUID is the only history address. A partial or different decision
	// at that address is retained, never overwritten or adopted.
	history, e := admissionOpenChild(root, ".reviews")
	if e != nil && !os.IsNotExist(e) {
		return m, e
	}
	if e == nil {
		defer history.Close()
		op, openErr := admissionOpenChild(history, approval.HoldDecisionID)
		if openErr == nil {
			defer op.Close()
			m, err = readRecoveryCancellation(root, op, id, approval)
			if err != nil {
				return m, err
			}
			if m.Reason != reason || m.ActorID != actor.Uid {
				return m, errors.New("existing cancellation actor or reason changed")
			}
			return finishRecoveryCancellation(root, op, m, seam)
		}
		if !os.IsNotExist(openErr) {
			return m, openErr
		}
	}
	entry, e := activeReleaseEvidence(root, id, approval)
	if e != nil {
		return m, e
	}
	if history == nil {
		history, e = admissionChild(root, ".reviews")
		if e != nil {
			return m, e
		}
		defer history.Close()
	}
	entries, e := history.file.ReadDir(maxAdmissionRecovery + 1)
	if e != nil && !errors.Is(e, io.EOF) || len(entries) >= maxAdmissionRecovery {
		return m, errors.Join(e, errors.New("cancellation history is full; evidence retained"))
	}
	op, e := admissionMakeDirectory(history, approval.HoldDecisionID)
	if e != nil {
		return m, e
	}
	defer op.Close()
	m = RecoveryCancellation{Version: 1, Kind: "restore-cancellation", DecisionID: uuid.NewString(), QuarantineID: id, RootID: rootID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Approval: approval, Hold: entry}
	if m.DirectoryID, e = admissionFileID(op.file); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(history); e != nil {
		return m, e
	}
	if err = seam("cancel-directory-created"); err != nil {
		return m, err
	}
	f, e := admissionCreateFile(op, "manifest.json")
	if e != nil {
		return m, e
	}
	defer f.Close()
	if m.FileID, e = admissionFileID(f); e != nil {
		return m, e
	}
	body, e := json.Marshal(m)
	if e != nil || len(body) > maxRecoveryRelease {
		return m, errors.Join(e, errors.New("cancellation decision exceeds its bound"))
	}
	if _, e = f.Write(body); e != nil {
		return m, e
	}
	if err = seam("cancel-receipt-written"); err != nil {
		return m, err
	}
	if e = f.Sync(); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(op); e != nil {
		return m, e
	}
	if err = seam("cancel-approved"); err != nil {
		return m, err
	}
	return finishRecoveryCancellation(root, op, m, seam)
}

func readRecoveryCancellation(root, op *restoreDirectory, id string, a RecoveryReleaseApproval) (RecoveryCancellation, error) {
	m, err := readCancellationRecord(op, id, a)
	if err != nil {
		return m, err
	}
	entries, err := op.file.ReadDir(3)
	if err != nil && !errors.Is(err, io.EOF) {
		return m, err
	}
	for _, entry := range entries {
		if entry.Name() != "manifest.json" && entry.Name() != "restore" || entry.Name() == "manifest.json" && entry.IsDir() || entry.Name() == "restore" && !entry.IsDir() {
			return m, errors.New("unknown cancellation history entry retained")
		}
	}
	return m, nil
}

// Immutable ownership evidence is independent of later lifecycle preparations.
// This reader never authorizes cancellation retry, publication, or cleanup.
func readCancellationRecord(op *restoreDirectory, id string, a RecoveryReleaseApproval) (RecoveryCancellation, error) {
	var m RecoveryCancellation
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	entry, readErr := inspectJournalFile(JournalInspection{}, f, maxRecoveryRelease)
	closeErr := f.Close()
	if readErr != nil || closeErr != nil {
		return m, errors.Join(readErr, closeErr)
	}
	if err = json.Unmarshal(entry.Manifest, &m); err != nil {
		return m, err
	}
	opID, err := admissionFileID(op.file)
	if err != nil {
		return m, err
	}
	if m.Version != 1 || m.Kind != "restore-cancellation" || !validReviewUUID(m.DecisionID) || m.QuarantineID != id || m.RootID != a.Original.RootID || m.DirectoryID != opID || m.FileID != entry.ManifestID || m.ApprovedAt.IsZero() || m.ActorID == "" || len(m.ActorID) > 256 || !validReviewReason(m.Reason) || m.Approval != a || m.Hold.EntryKind != "file" || m.Hold.Operation != "purge" || m.Hold.DirectoryID != a.HoldDirectoryID {
		return m, errors.New("cancellation identity or approval changed; history retained")
	}
	if _, err = decodeReleaseHold(m.Hold, id, a); err != nil {
		return m, err
	}
	return m, nil
}

// cancellationOriginal validates exactly one location of the original directory.
// It intentionally does not decode or enumerate that directory's children: even
// malformed journals and unknown children are preserved by the directory move.
func cancellationOriginal(root, historyOp *restoreDirectory, m RecoveryCancellation) (archived bool, err error) {
	live, err := inspectJournal(root, ".restores", m.QuarantineID)
	if err != nil {
		return false, err
	}
	saved := JournalInspection{Status: "absent"}
	op, err := admissionOpenChild(historyOp, "restore")
	if err == nil {
		defer op.Close()
		saved.DirectoryID, err = admissionFileID(op.file)
		if err != nil {
			return false, err
		}
		f, e := admissionReadFile(op, "manifest.json")
		if e != nil {
			return false, e
		}
		saved, err = inspectJournalFile(saved, f, maxAdmissionJournal)
		err = errors.Join(err, f.Close())
		if err != nil {
			return false, err
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if (live.Status == "absent") == (saved.Status == "absent") {
		return false, errors.New("original operation must exist in exactly one approved location")
	}
	current := live
	if live.Status == "absent" {
		current, archived = saved, true
	}
	a := m.Approval.Original
	if current.Status != "captured" || current.DirectoryID != a.DirectoryID || current.ManifestID != a.ManifestID || current.SHA256 != a.SHA256 {
		return false, errors.New("cancelled original operation identity or bytes changed")
	}
	return archived, nil
}

func cancellationFence(root *restoreDirectory, m RecoveryCancellation) error {
	entry, err := inspectJournal(root, ".purges", m.QuarantineID)
	if err != nil || entry.EntryKind != "file" || entry.DirectoryID != m.Approval.HoldDirectoryID || !matchesReleaseEvidence(entry, m.Approval) {
		return errors.Join(err, errors.New("cancellation retention fence changed"))
	}
	_, err = decodeReleaseHold(entry, m.QuarantineID, m.Approval)
	return err
}

// openCancellationProof follows one exact hold UUID, never a history scan.
// The returned handles belong to the caller. Both the still-active fence and
// the same original directory in one approved location must remain provable.
func openCancellationProof(root *restoreDirectory, id string, a RecoveryReleaseApproval) (*restoreDirectory, RecoveryCancellation, error) {
	var m RecoveryCancellation
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != a.Original.RootID {
		return nil, m, errors.Join(err, errors.New("cancellation proof root identity changed"))
	}
	history, err := admissionOpenChild(root, ".reviews")
	if err != nil {
		return nil, m, err
	}
	defer history.Close()
	op, err := admissionOpenChild(history, a.HoldDecisionID)
	if err != nil {
		return nil, m, err
	}
	m, err = readRecoveryCancellation(root, op, id, a)
	if err == nil {
		err = cancellationFence(root, m)
	}
	if err == nil {
		_, err = cancellationOriginal(root, op, m)
	}
	if err != nil {
		op.Close()
		return nil, m, err
	}
	return op, m, nil
}

func cancellationApproval(hold RecoveryHold, entry JournalInspection) RecoveryReleaseApproval {
	return RecoveryReleaseApproval{
		Original:       RecoveryApproval{RootID: hold.RootID, DirectoryID: hold.Original.DirectoryID, ManifestID: hold.Original.ManifestID, SHA256: hold.Original.SHA256},
		HoldDecisionID: hold.DecisionID, HoldDirectoryID: entry.DirectoryID, HoldFileID: entry.ManifestID, HoldSHA256: entry.SHA256,
	}
}

func openArchivedRestoreProof(root *restoreDirectory, id string) (*restoreDirectory, *restoreDirectory, error) {
	entry, err := inspectJournal(root, ".purges", id)
	if err != nil || entry.EntryKind != "file" || entry.Status != "captured" {
		return nil, nil, errors.Join(err, errors.New("archived restore has no exact active fence"))
	}
	var hold RecoveryHold
	if err = json.Unmarshal(entry.Manifest, &hold); err != nil {
		return nil, nil, err
	}
	a := cancellationApproval(hold, entry)
	if err = validRecoveryReleaseApproval(a); err != nil {
		return nil, nil, err
	}
	history, err := admissionOpenChild(root, ".reviews")
	if err != nil {
		return nil, nil, err
	}
	defer history.Close()
	op, err := admissionOpenChild(history, a.HoldDecisionID)
	if err != nil {
		return nil, nil, err
	}
	m, err := readCancellationRecord(op, id, a)
	rootID, identityErr := admissionFileID(root.file)
	if err == nil && (identityErr != nil || rootID != m.RootID) {
		err = errors.Join(identityErr, errors.New("archived ownership root changed"))
	}
	if err == nil {
		err = cancellationFence(root, m)
	}
	var original *restoreDirectory
	if err == nil {
		original, err = openBoundCancelledOriginal(op, m)
	}
	if err != nil {
		op.Close()
		return nil, nil, err
	}
	return op, original, nil
}

// Read the one archived original by its exact directory, journal identity and
// hash, without adopting anything in the now-independent live operation slot.
func openBoundCancelledOriginal(op *restoreDirectory, m RecoveryCancellation) (*restoreDirectory, error) {
	original, err := admissionOpenChild(op, "restore")
	if err != nil {
		return nil, err
	}
	directoryID, err := admissionFileID(original.file)
	if err == nil && directoryID != m.Approval.Original.DirectoryID {
		err = errors.New("cancelled original directory identity changed")
	}
	if err == nil {
		var f *os.File
		f, err = admissionReadFile(original, "manifest.json")
		if err == nil {
			var entry JournalInspection
			entry, err = inspectJournalFile(JournalInspection{}, f, maxAdmissionJournal)
			err = errors.Join(err, f.Close())
			if err == nil && (entry.ManifestID != m.Approval.Original.ManifestID || entry.SHA256 != m.Approval.Original.SHA256) {
				err = errors.New("cancelled original journal identity or bytes changed")
			}
		}
	}
	if err != nil {
		original.Close()
		return nil, err
	}
	return original, nil
}

func finishRecoveryCancellation(root, op *restoreDirectory, m RecoveryCancellation, seam func(string) error) (RecoveryCancellation, error) {
	if err := cancellationFence(root, m); err != nil {
		return m, err
	}
	archived, err := cancellationOriginal(root, op, m)
	if err != nil {
		return m, err
	}
	if err = syncCancellationReceipt(op, m); err != nil {
		return m, err
	}
	if !archived {
		if err = seam("cancel-before-move"); err != nil {
			return m, err
		}
		if err = syncCancellationReceipt(op, m); err != nil {
			return m, err
		}
		if err = cancellationFence(root, m); err != nil {
			return m, err
		}
		if archived, err = cancellationOriginal(root, op, m); err != nil || archived {
			return m, errors.Join(err, errors.New("original location changed before cancellation move"))
		}
		journals, e := admissionOpenChild(root, ".restores")
		if e != nil {
			return m, e
		}
		err = moveRecoveryDirectory(journals, m.QuarantineID, op, "restore", m.Approval.Original.DirectoryID)
		journals.Close()
		if err != nil {
			return m, err
		}
		if err = seam("cancel-archived"); err != nil {
			return m, err
		}
	}
	if archived, err = cancellationOriginal(root, op, m); err != nil || !archived {
		return m, errors.Join(err, errors.New("cancelled operation archive is not proven"))
	}
	if err = admissionSyncDirectory(op); err != nil {
		return m, err
	}
	journals, err := admissionOpenChild(root, ".restores")
	if err != nil {
		return m, err
	}
	err = admissionSyncDirectory(journals)
	err = errors.Join(err, journals.Close())
	if err != nil {
		return m, err
	}
	return m, seam("cancel-synced")
}

func syncCancellationReceipt(op *restoreDirectory, m RecoveryCancellation) error {
	// A complete receipt may have survived an exit before Sync. Recheck its
	// exact bytes and identity, then flush it again before any original move.
	f, err := openRecoveryReleaseReceipt(op)
	if err != nil {
		return err
	}
	entry, err := inspectJournalFile(JournalInspection{}, f, maxRecoveryRelease)
	expected, encodeErr := json.Marshal(m)
	if err != nil || encodeErr != nil || entry.ManifestID != m.FileID || !bytes.Equal(entry.Manifest, expected) {
		f.Close()
		return errors.Join(err, encodeErr, errors.New("cancellation receipt identity or bytes changed before flush"))
	}
	err = f.Sync()
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return admissionSyncDirectory(op)
}
