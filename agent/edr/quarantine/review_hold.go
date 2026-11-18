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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ErrRecoveryHeld = errors.New("recovery is retained by an operator decision")

const maxRecoveryReview = 2*maxAdmissionJournal + (64 << 10)

type RecoveryApproval struct {
	RootID      string `json:"root_id"`
	DirectoryID string `json:"directory_id"`
	ManifestID  string `json:"manifest_id"`
	SHA256      string `json:"sha256"`
}

// RecoveryHold is an explicit retention decision, not publication, cancellation
// or a restore receipt. Original preserves the exact reviewed journal bytes.
type RecoveryHold struct {
	Version      int               `json:"version"`
	Kind         string            `json:"kind"`
	DecisionID   string            `json:"decision_id"`
	QuarantineID string            `json:"quarantine_id"`
	RootID       string            `json:"root_id"`
	FileID       string            `json:"file_id"`
	ApprovedAt   time.Time         `json:"approved_at"`
	ActorID      string            `json:"actor_id"`
	Reason       string            `json:"reason"`
	Original     JournalInspection `json:"original"`
}

// HoldRecovery opens existing private storage without SQLite, hardening, log
// initialization or automatic recovery. A regular file at .purges/<id> is an
// intentional fence: the earlier directory-only purge/restore paths refuse it.
func HoldRecovery(dir, id string, approval RecoveryApproval, reason string) (RecoveryHold, error) {
	return holdRecovery(dir, id, approval, reason, nil)
}

func validRecoveryApproval(a RecoveryApproval) error {
	for _, id := range []string{a.RootID, a.DirectoryID, a.ManifestID} {
		if id == "" || len(id) > 256 {
			return errors.New("review requires bounded root, directory and manifest identities")
		}
	}
	raw, err := hex.DecodeString(a.SHA256)
	if err != nil || len(raw) != sha256.Size || hex.EncodeToString(raw) != a.SHA256 {
		return errors.New("review requires the exact lowercase journal SHA-256")
	}
	return nil
}
func validReviewReason(reason string) bool {
	return len(reason) <= 4096 && utf8.ValidString(reason) && strings.TrimSpace(reason) != "" && !strings.ContainsRune(reason, 0)
}
func reviewMatchesApproval(m RecoveryHold, a RecoveryApproval) bool {
	return m.RootID == a.RootID && m.Original.DirectoryID == a.DirectoryID && m.Original.ManifestID == a.ManifestID && m.Original.SHA256 == a.SHA256
}
func holdRecovery(dir, id string, approval RecoveryApproval, reason string, fault func(string) error) (m RecoveryHold, err error) {
	if !ordinaryQuarantineID(id) {
		return m, errors.New("invalid quarantine id")
	}
	if err = validRecoveryApproval(approval); err != nil {
		return m, err
	}
	if !validReviewReason(reason) {
		return m, errors.New("review reason must contain at most 4096 UTF-8 bytes")
	}
	actor, e := user.Current()
	if e != nil || actor.Uid == "" {
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
	if e != nil {
		return m, e
	}
	if rootID != approval.RootID {
		return m, errors.New("review root identity changed")
	}
	unlock, e := root.lockItem(id)
	if e != nil {
		return m, e
	}
	defer unlock()
	original, e := inspectJournal(root, ".restores", id)
	if e != nil {
		return m, e
	}
	original.Operation = "restore"
	if original.Status != "captured" || original.DirectoryID != approval.DirectoryID || original.ManifestID != approval.ManifestID || original.SHA256 != approval.SHA256 {
		return m, errors.New("reviewed restore journal changed; nothing was approved")
	}
	existing, found, e := recoveryHold(root, id)
	if e != nil {
		return m, e
	}
	if found {
		if !reviewMatchesApproval(existing, approval) || existing.Reason != reason || existing.ActorID != actor.Uid {
			return m, errors.New("another operator retention decision already exists")
		}
		return existing, nil
	}
	journals, e := admissionChild(root, ".purges")
	if e != nil {
		return m, e
	}
	defer journals.Close()
	entries, e := journals.file.ReadDir(maxAdmissionRecovery + 1)
	if e != nil && !errors.Is(e, io.EOF) || len(entries) >= maxAdmissionRecovery {
		return m, errors.Join(e, errors.New("review/purge recovery queue is full"))
	}
	// A directory here is an existing purge, never an invitation to replace it.
	if old, e := admissionOpenChild(journals, id); e == nil {
		old.Close()
		return m, errors.New("an existing purge must not be replaced by a hold")
	} else if !os.IsNotExist(e) {
		if _, found, readErr := recoveryHold(root, id); readErr != nil || found {
			return m, errors.Join(readErr, errors.New("existing review/purge entry retained"))
		}
	}
	restoreJournals, sourceOp, e := openRestoreOperation(root, id)
	if e != nil {
		return m, e
	}
	defer restoreJournals.Close()
	defer sourceOp.Close()
	sourceID, e := admissionFileID(sourceOp.file)
	if e != nil || sourceID != approval.DirectoryID {
		return m, errors.Join(e, errors.New("reviewed operation directory changed"))
	}
	decision := uuid.NewString()
	temporary := ".review-" + decision
	// Prepare inside the item-locked original operation. If publication is
	// interrupted, earlier runtimes retain this unknown entry instead of
	// replaying through a partially recorded operator decision.
	f, e := admissionCreateFile(sourceOp, temporary)
	if e != nil {
		return m, e
	}
	defer f.Close()
	fileID, e := admissionFileID(f)
	if e != nil {
		return m, e
	}
	m = RecoveryHold{Version: 2, Kind: "restore-retention-hold", DecisionID: decision, QuarantineID: id, RootID: rootID, FileID: fileID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Original: original}
	seam := func(phase string) error {
		if fault != nil {
			return fault(phase)
		}
		return nil
	}
	if err = seam("review-temp-created"); err != nil {
		return m, err
	}
	body, e := json.Marshal(m)
	if e != nil {
		return m, e
	}
	if len(body) > maxRecoveryReview {
		return m, errors.New("review decision exceeds its bound")
	}
	if _, e = f.Write(body); e != nil {
		return m, e
	}
	if e = f.Sync(); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(sourceOp); e != nil {
		return m, e
	}
	if err = seam("review-prepared"); err != nil {
		return m, err
	}
	// Check the entire reviewed identity again while the item lock is held. A
	// replaced operation or edited journal cannot inherit an earlier approval.
	check, e := inspectJournal(root, ".restores", id)
	if e != nil {
		return m, e
	}
	if check.DirectoryID != approval.DirectoryID || check.ManifestID != approval.ManifestID || check.SHA256 != approval.SHA256 {
		return m, errors.New("reviewed evidence changed before decision publication; preparation retained")
	}
	if e = admissionRename(sourceOp, temporary, journals, id, false); e != nil {
		return m, e
	}
	if err = seam("review-published"); err != nil {
		return m, err
	}
	if e = admissionSyncDirectory(journals); e != nil {
		return m, e
	}
	if e = admissionSyncDirectory(sourceOp); e != nil {
		return m, e
	}
	if err = seam("review-synced"); err != nil {
		return m, err
	}
	return m, nil
}

func readRecoveryHold(root, journals *restoreDirectory, id string) (RecoveryHold, error) {
	var m RecoveryHold
	f, err := admissionReadFile(journals, id)
	if err != nil {
		return m, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxRecoveryReview {
		return m, errors.Join(err, errors.New("invalid review decision file"))
	}
	fileID, err := admissionFileID(f)
	if err != nil {
		return m, err
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxRecoveryReview+1))
	if err != nil || len(raw) > maxRecoveryReview {
		return m, errors.Join(err, errors.New("review decision exceeds its bound"))
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	rootID, err := admissionFileID(root.file)
	if err != nil {
		return m, err
	}
	nonce, nonceErr := uuid.Parse(m.DecisionID)
	if m.Version != 2 || m.Kind != "restore-retention-hold" || m.QuarantineID != id || !ordinaryQuarantineID(id) || m.RootID != rootID || m.FileID != fileID || m.ApprovedAt.IsZero() || m.ActorID == "" || len(m.ActorID) > 256 || nonceErr != nil || nonce == uuid.Nil || nonce.String() != m.DecisionID || !validReviewReason(m.Reason) {
		return m, errors.New("invalid review decision identity")
	}
	a := RecoveryApproval{RootID: m.RootID, DirectoryID: m.Original.DirectoryID, ManifestID: m.Original.ManifestID, SHA256: m.Original.SHA256}
	if err = validRecoveryApproval(a); err != nil {
		return m, err
	}
	digest := sha256.Sum256(m.Original.Manifest)
	if m.Original.Operation != "restore" || m.Original.Status != "captured" || len(m.Original.Manifest) > maxAdmissionJournal || hex.EncodeToString(digest[:]) != m.Original.SHA256 {
		return m, errors.New("invalid exact original review evidence")
	}
	current, err := inspectJournal(root, ".restores", id)
	if err != nil || current.Status != "captured" || current.DirectoryID != a.DirectoryID || current.ManifestID != a.ManifestID || current.SHA256 != a.SHA256 {
		// A distinct, durable cancellation receipt may bind the same original
		// directory at its exact archive address. The v2 fence itself is never
		// edited, moved or treated as permission to release or publish anything.
		parentID, e := admissionFileID(journals.file)
		if e != nil {
			return m, e
		}
		sum := sha256.Sum256(raw)
		entry := JournalInspection{Status: "captured", EntryKind: "file", DirectoryID: parentID, ManifestID: fileID, SHA256: hex.EncodeToString(sum[:])}
		op, _, e := openCancellationProof(root, id, cancellationApproval(m, entry))
		if e != nil {
			// A later independent preparation may occupy the live slot or add
			// history. That must not invalidate the exact archived original's
			// ownership proof. This fallback only retains the fence and stage;
			// cancellation/release actions still use their stricter validation.
			archive, original, ownershipErr := openArchivedRestoreProof(root, id)
			if ownershipErr != nil {
				return m, errors.Join(e, ownershipErr, errors.New("retention fence remains but its original evidence binding changed"))
			}
			original.Close()
			archive.Close()
		} else {
			op.Close()
		}
	}
	return m, nil
}

func recoveryHold(root *restoreDirectory, id string) (RecoveryHold, bool, error) {
	var m RecoveryHold
	journals, err := admissionOpenChild(root, ".purges")
	if os.IsNotExist(err) {
		return m, false, nil
	}
	if err != nil {
		return m, false, err
	}
	defer journals.Close()
	op, dirErr := admissionOpenChild(journals, id)
	if dirErr == nil {
		op.Close()
		return m, false, nil
	}
	if os.IsNotExist(dirErr) {
		return m, false, nil
	}
	m, err = readRecoveryHold(root, journals, id)
	if err != nil {
		return m, false, fmt.Errorf("%w: review/purge entry retained: %v", ErrRestoreIntegrity, err)
	}
	return m, true, nil
}
func checkNoRecoveryHold(root *restoreDirectory, id string) error {
	m, found, err := recoveryHold(root, id)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("%w: %s (%q)", ErrRecoveryHeld, m.DecisionID, m.Reason)
	}
	retired, history, found, err := openRetirementProof(root, id)
	if err != nil {
		return fmt.Errorf("%w: retirement guard retained: %v", ErrRestoreIntegrity, err)
	}
	if found {
		history.Close()
		return fmt.Errorf("%w: %s (%q)", ErrRecoveryRetired, retired.DecisionID, retired.Reason)
	}
	return nil
}
