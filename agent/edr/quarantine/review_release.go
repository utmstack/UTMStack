package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const maxRecoveryRelease = 4*maxAdmissionJournal + (128 << 10)

// RecoveryReleaseApproval approves removal of one exact retention fence. It
// does not approve a different hold, cancel the original restore, or prove its
// publication. Normal recovery of the original operation becomes possible.
type RecoveryReleaseApproval struct {
	Original        RecoveryApproval `json:"original"`
	HoldDecisionID  string           `json:"hold_decision_id"`
	HoldDirectoryID string           `json:"hold_directory_id"`
	HoldFileID      string           `json:"hold_file_id"`
	HoldSHA256      string           `json:"hold_sha256"`
}

// RecoveryRelease preserves exact fence evidence in a private history operation.
// HoldFileID is also checked on the archived hold.json: a content match cannot
// adopt a replacement. The original restore journal is never moved or edited.
type RecoveryRelease struct {
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

func validReviewUUID(id string) bool {
	u, err := uuid.Parse(id)
	return err == nil && u != uuid.Nil && u.String() == id
}

func validRecoveryReleaseApproval(a RecoveryReleaseApproval) error {
	if err := validRecoveryApproval(a.Original); err != nil {
		return err
	}
	if !validReviewUUID(a.HoldDecisionID) {
		return errors.New("release requires the exact hold decision UUID")
	}
	return validRecoveryApproval(RecoveryApproval{RootID: a.Original.RootID, DirectoryID: a.HoldDirectoryID, ManifestID: a.HoldFileID, SHA256: a.HoldSHA256})
}

func matchesReleaseEvidence(entry JournalInspection, a RecoveryReleaseApproval) bool {
	return entry.Status == "captured" && entry.ManifestID == a.HoldFileID && entry.SHA256 == a.HoldSHA256
}

// decodeReleaseHold validates retained bytes without requiring the original
// restore to still exist. Only an already approved, identity-bound release
// history may use that property on retry after normal recovery has completed.
func decodeReleaseHold(entry JournalInspection, id string, a RecoveryReleaseApproval) (RecoveryHold, error) {
	var hold RecoveryHold
	if !matchesReleaseEvidence(entry, a) || len(entry.Manifest) > maxRecoveryReview {
		return hold, errors.New("release hold evidence changed")
	}
	sum := sha256.Sum256(entry.Manifest)
	if hex.EncodeToString(sum[:]) != a.HoldSHA256 {
		return hold, errors.New("release hold bytes changed")
	}
	if err := json.Unmarshal(entry.Manifest, &hold); err != nil {
		return hold, err
	}
	if hold.Version != 2 || hold.Kind != "restore-retention-hold" || hold.DecisionID != a.HoldDecisionID || hold.QuarantineID != id || hold.RootID != a.Original.RootID || hold.FileID != a.HoldFileID || hold.ApprovedAt.IsZero() || hold.ActorID == "" || len(hold.ActorID) > 256 || !validReviewReason(hold.Reason) || !reviewMatchesApproval(hold, a.Original) {
		return hold, errors.New("release hold identity does not match the approval")
	}
	originalSum := sha256.Sum256(hold.Original.Manifest)
	if hold.Original.Operation != "restore" || hold.Original.Status != "captured" || len(hold.Original.Manifest) > maxAdmissionJournal || hex.EncodeToString(originalSum[:]) != a.Original.SHA256 {
		return hold, errors.New("release original journal evidence changed")
	}
	return hold, nil
}

// ReleaseRecovery opens existing storage without SQLite, automatic recovery,
// logging or hardening. The explicit approval makes normal recovery possible;
// this call does not run it. The history is addressed by the exact hold UUID.
func ReleaseRecovery(dir, id string, approval RecoveryReleaseApproval, reason string) (RecoveryRelease, error) {
	return releaseRecovery(dir, id, approval, reason, nil)
}

func releaseRecovery(dir, id string, approval RecoveryReleaseApproval, reason string, fault func(string) error) (m RecoveryRelease, err error) {
	if !ordinaryQuarantineID(id) {
		return m, errors.New("invalid quarantine id")
	}
	if err = validRecoveryReleaseApproval(approval); err != nil {
		return m, err
	}
	if !validReviewReason(reason) {
		return m, errors.New("release reason must contain at most 4096 UTF-8 bytes")
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
		return m, errors.Join(e, errors.New("release root identity changed"))
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
	// Check an exact history address before looking at live restore evidence.
	// A completed earlier release remains provable after recovery or a new hold.
	history, e := admissionOpenChild(root, ".reviews")
	if e != nil && !os.IsNotExist(e) {
		return m, e
	}
	if e == nil {
		defer history.Close()
		op, openErr := admissionOpenChild(history, approval.HoldDecisionID)
		if openErr == nil {
			defer op.Close()
			m, err = readRecoveryRelease(root, op, id, approval, reason, actor.Uid)
			if err != nil {
				return m, err // Never adopt a bare or partially written operation.
			}
			return finishRecoveryRelease(root, op, m, seam)
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
		return m, errors.Join(e, errors.New("release history is full; existing evidence retained"))
	}
	op, e := admissionMakeDirectory(history, approval.HoldDecisionID)
	if e != nil {
		return m, e
	}
	defer op.Close()
	m = RecoveryRelease{Version: 1, Kind: "restore-hold-release", DecisionID: uuid.NewString(), QuarantineID: id, RootID: rootID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Approval: approval, Hold: entry}
	if m.DirectoryID, e = admissionFileID(op.file); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(history); e != nil {
		return m, e
	}
	if err = seam("release-directory-created"); err != nil {
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
		return m, errors.Join(e, errors.New("release decision exceeds its bound"))
	}
	if _, e = f.Write(body); e != nil {
		return m, e
	}
	if err = seam("release-receipt-written"); err != nil {
		return m, err
	}
	if e = f.Sync(); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(op); e != nil {
		return m, e
	}
	if err = seam("release-approved"); err != nil {
		return m, err
	}
	return finishRecoveryRelease(root, op, m, seam)
}

func activeReleaseEvidence(root *restoreDirectory, id string, a RecoveryReleaseApproval) (JournalInspection, error) {
	entry, err := inspectJournal(root, ".purges", id)
	entry.Operation = "purge"
	if err != nil || entry.EntryKind != "file" || entry.DirectoryID != a.HoldDirectoryID {
		return entry, errors.Join(err, errors.New("reviewed retention fence location changed"))
	}
	if _, err = decodeReleaseHold(entry, id, a); err != nil {
		return entry, err
	}
	if _, found, err := recoveryHold(root, id); err != nil || !found {
		return entry, errors.Join(err, errors.New("reviewed retention fence or original binding changed"))
	}
	return entry, nil
}

func readRecoveryRelease(root, op *restoreDirectory, id string, a RecoveryReleaseApproval, reason, actor string) (RecoveryRelease, error) {
	var m RecoveryRelease
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	defer f.Close()
	entry, err := inspectJournalFile(JournalInspection{}, f, maxRecoveryRelease)
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(entry.Manifest, &m); err != nil {
		return m, err
	}
	opID, err := admissionFileID(op.file)
	if err != nil {
		return m, err
	}
	if m.Version != 1 || m.Kind != "restore-hold-release" || !validReviewUUID(m.DecisionID) || m.QuarantineID != id || m.RootID != a.Original.RootID || m.DirectoryID != opID || m.FileID != entry.ManifestID || m.ApprovedAt.IsZero() || m.ActorID != actor || m.Reason != reason || m.Approval != a || m.Hold.EntryKind != "file" || m.Hold.Operation != "purge" || m.Hold.DirectoryID != a.HoldDirectoryID {
		return m, errors.New("existing release identity or approval changed; history retained")
	}
	if _, err = decodeReleaseHold(m.Hold, id, a); err != nil {
		return m, err
	}
	entries, err := op.file.ReadDir(3)
	if err != nil && !errors.Is(err, io.EOF) {
		return m, err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() != "manifest.json" && entry.Name() != "hold.json" {
			return m, errors.New("unknown release history entry retained")
		}
	}
	return m, nil
}

func finishRecoveryRelease(root, op *restoreDirectory, m RecoveryRelease, seam func(string) error) (RecoveryRelease, error) {
	// A completed historical release is independent of any later operation for
	// this quarantine ID. It must never move or approve a later retention fence.
	f, err := admissionReadFile(op, "hold.json")
	if err == nil {
		entry, readErr := inspectJournalFile(JournalInspection{}, f, maxRecoveryReview)
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || !matchesReleaseEvidence(entry, m.Approval) {
			return m, errors.Join(readErr, closeErr, errors.New("archived retention fence identity changed"))
		}
		active, inspectErr := inspectJournal(root, ".purges", m.QuarantineID)
		if inspectErr != nil {
			return m, inspectErr
		}
		if active.EntryKind == "file" && active.ManifestID == m.Approval.HoldFileID {
			return m, errors.New("same retention fence remains in both locations; ambiguous move retained")
		}
		if err = syncRecoveryRelease(root, op, m); err != nil {
			return m, err
		}
		return m, nil
	}
	if !os.IsNotExist(err) {
		return m, err
	}
	if _, err = activeReleaseEvidence(root, m.QuarantineID, m.Approval); err != nil {
		return m, err
	}
	// A retry may find a complete receipt written before its original process
	// exited. Flush it and its parent again before authorizing the fence move.
	receipt, err := openRecoveryReleaseReceipt(op)
	if err != nil {
		return m, err
	}
	receiptID, err := admissionFileID(receipt)
	if err != nil || receiptID != m.FileID {
		receipt.Close()
		return m, errors.Join(err, errors.New("release receipt changed before flush"))
	}
	err = receipt.Sync()
	closeErr := receipt.Close()
	if err != nil || closeErr != nil {
		return m, errors.Join(err, closeErr)
	}
	if err = admissionSyncDirectory(op); err != nil {
		return m, err
	}
	if err = seam("release-before-move"); err != nil {
		return m, err
	}
	if _, err = activeReleaseEvidence(root, m.QuarantineID, m.Approval); err != nil {
		return m, err
	}
	purges, err := admissionOpenChild(root, ".purges")
	if err != nil {
		return m, err
	}
	defer purges.Close()
	purgeID, err := admissionFileID(purges.file)
	if err != nil || purgeID != m.Approval.HoldDirectoryID {
		return m, errors.Join(err, errors.New("release fence container changed before move"))
	}
	if err = admissionRename(purges, m.QuarantineID, op, "hold.json", false); err != nil {
		return m, fmt.Errorf("release approved; retention fence move incomplete: %w", err)
	}
	if err = seam("release-published"); err != nil {
		return m, err
	}
	if err = syncRecoveryRelease(root, op, m); err != nil {
		return m, err
	}
	if err = seam("release-synced"); err != nil {
		return m, err
	}
	return m, nil
}

func syncRecoveryRelease(root, op *restoreDirectory, m RecoveryRelease) error {
	if err := admissionSyncDirectory(op); err != nil {
		return err
	}
	purges, err := admissionOpenChild(root, ".purges")
	if err != nil {
		return err
	}
	defer purges.Close()
	identity, err := admissionFileID(purges.file)
	if err != nil || identity != m.Approval.HoldDirectoryID {
		return errors.Join(err, errors.New("released fence container identity changed"))
	}
	return admissionSyncDirectory(purges)
}
