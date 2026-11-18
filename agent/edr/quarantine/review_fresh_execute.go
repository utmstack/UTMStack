package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type FreshRestoreResult struct {
	Version        int    `json:"version"`
	Kind           string `json:"kind"`
	QuarantineID   string `json:"quarantine_id"`
	OperationID    string `json:"operation_id"`
	ApprovalSHA256 string `json:"approval_sha256"`
	Destination    []byte `json:"destination"`
	ParentID       string `json:"parent_id"`
	OutputID       string `json:"output_id"`
	Phase          string `json:"phase"`
}

type freshCacheOpener func(*freshRestorePermit, restoreManifest) (*cache.Cache, func() error, error)

// ExecuteFreshRestore consumes the exact saved independent approval. It opens
// existing metadata only after validating the permission, actor, active hold,
// original archive, retained snapshot and destination identities. It never calls
// New/Recover, releases the hold, or retires either original backup.
func ExecuteFreshRestore(dir, database string, savedApproval []byte, reason string) (FreshRestoreResult, error) {
	return executeFreshRestore(dir, savedApproval, reason, func(p *freshRestorePermit, m restoreManifest) (*cache.Cache, func() error, error) {
		path, releasePin, err := p.pinDatabase(database, m)
		if err != nil {
			return nil, nil, err
		}
		c, err := cache.Open(path)
		if err != nil {
			releasePin()
			return nil, nil, err
		}
		return c, func() error { return errors.Join(c.Close(), releasePin()) }, nil
	}, nil)
}

// Only a validated instance under the item lock reaches mutation helpers.
// There is deliberately no general-purpose option to bypass recovery holds.
type freshRestorePermit struct {
	root, decision       *restoreDirectory
	approval             FreshRestoreApproval
	approvalHash, reason string
	owner                freshOwnership
	seam                 func(string) error
}

func executeFreshRestore(dir string, savedApproval []byte, reason string, openCache freshCacheOpener, fault func(string) error) (result FreshRestoreResult, err error) {
	if len(savedApproval) > maxFreshRestoreApproval || !validReviewReason(reason) {
		return result, errors.New("fresh execution requires bounded saved approval and reason")
	}
	var a FreshRestoreApproval
	if err = json.Unmarshal(savedApproval, &a); err != nil {
		return result, err
	}
	approvalHash, canonical, err := freshJSONHash(a)
	if err != nil || !bytes.Equal(bytes.TrimSpace(savedApproval), canonical) || !ordinaryQuarantineID(a.QuarantineID) || !validReviewUUID(a.OperationID) {
		return result, errors.Join(err, errors.New("fresh execution requires the exact saved approval JSON"))
	}
	actor, err := user.Current()
	if err != nil || actor.Uid != a.ActorID || len(actor.Uid) > 256 {
		return result, errors.Join(err, errors.New("fresh execution actor differs from the independent approval"))
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return result, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return result, err
	}
	if err = freshLocalDestination(filepath.Join(resolved, "fresh-root-probe")); err != nil {
		return result, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return result, err
	}
	defer root.Close()
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != a.RootID {
		return result, errors.Join(err, errors.New("fresh execution root changed"))
	}
	queue, item, decision, err := openFreshApproval(root, a.QuarantineID, a.OperationID)
	if err != nil {
		return result, err
	}
	defer queue.Close()
	defer item.Close()
	defer decision.Close()
	actual, err := readFreshApproval(root, queue, item, decision, a.QuarantineID, a.OperationID)
	if err != nil {
		return result, err
	}
	actualHash, _, err := freshJSONHash(actual)
	if err != nil || actualHash != approvalHash {
		return result, errors.Join(err, errors.New("saved fresh approval differs from private storage"))
	}
	unlock, err := root.lockItem(a.QuarantineID)
	if err != nil {
		return result, err
	}
	defer unlock()
	p := &freshRestorePermit{root: root, decision: decision, approval: a, approvalHash: approvalHash, reason: reason}
	p.seam = func(phase string) error {
		if fault != nil {
			return fault("fresh-execution-" + phase)
		}
		return nil
	}
	old, err := p.checkAuthority()
	if err != nil {
		return result, err
	}
	if err = freshKnownEntries(decision, map[string]bool{"manifest.json": false, "ownership.json": false, "restore": true}); err != nil {
		return result, err
	}
	parent, err := platformRestoreDirectory(string(a.Inspection.Destination.ParentPath))
	if err != nil {
		return result, err
	}
	defer parent.Close()
	parentID, err := admissionFileID(parent.file)
	if err != nil || parentID != a.Inspection.Destination.ParentID {
		return result, errors.Join(err, errors.New("fresh destination parent changed"))
	}
	if err = freshLocalDestination(filepath.Join(parent.path, string(a.Inspection.Destination.Leaf))); err != nil {
		return result, err
	}
	if err = rejectFreshManagedDestination(parent, a.RootID, old.StageID); err != nil {
		return result, err
	}
	input, err := restoreOpenInput(root, a.QuarantineID+".quarantined")
	if err != nil {
		return result, err
	}
	defer input.Close()
	pin, err := inspectFreshInput(input)
	if err != nil || pin != a.Inspection.Snapshot {
		return result, errors.Join(err, errors.New("approved fresh snapshot changed"))
	}
	var m restoreManifest
	var op *restoreDirectory
	var archived bool
	p.owner, _, err = readFreshOwnership(decision, a, approvalHash)
	newOperation := os.IsNotExist(err)
	if err != nil && !newOperation {
		return result, err
	}
	if !newOperation {
		if p.owner.Reason != reason || string(p.owner.RootPath) != root.path {
			return result, errors.New("fresh execution reason or root path changed")
		}
		op, archived, err = openFreshOwnedOperation(root, decision, p.owner)
		if err != nil {
			return result, err
		}
		defer func() {
			if op != nil {
				op.Close()
			}
		}()
		m, err = readRestoreManifest(op, a.QuarantineID)
		if err != nil {
			return result, err
		}
		if err = validateFreshJournal(m, a, p.owner); err != nil {
			return result, err
		}
		if err = freshKnownOperation(op); err != nil {
			return result, err
		}
		if archived && (m.Phase != "committed" || m.StageState != "deleted") {
			return result, errors.New("fresh archived operation lacks completion proof")
		}
		if !archived {
			if unexpected, e := admissionOpenChild(decision, "restore"); e == nil {
				unexpected.Close()
				return result, errors.New("unknown fresh archive retained")
			} else if !os.IsNotExist(e) {
				return result, e
			}
			if err = p.checkStageReadOnly(parent, op, m); err != nil {
				return result, err
			}
		}
	} else {
		if err = freshKnownEntries(decision, map[string]bool{"manifest.json": false}); err != nil {
			return result, err
		}
		if err = checkNoRestoreOperation(root, a.QuarantineID); err != nil {
			return result, err
		}
		if err = requirePurgeNameAbsent(parent, string(a.Inspection.Destination.Leaf)); err != nil {
			return result, err
		}
	}
	if err = syncFreshApproval(decision, a); err != nil {
		return result, err
	}
	if openCache == nil {
		return result, errors.New("fresh execution requires metadata storage")
	}
	c, closeCache, err := openCache(p, m)
	if err != nil {
		return result, err
	}
	if c == nil {
		return result, errors.New("fresh metadata opener returned no database")
	}
	if closeCache != nil {
		defer func() { err = errors.Join(err, closeCache()) }()
	}
	s := &Store{dir: root.path, rootID: a.RootID, cache: c}
	rec, found, err := c.GetQuarantine(a.QuarantineID)
	if err != nil || !found || rec.Purged {
		return result, errors.Join(err, errors.New("fresh quarantine record is missing or purged"))
	}
	if newOperation || m.Phase != "published" && m.Phase != "committed" && m.Phase != "publishing" {
		body, e := cache.EncodeQuarantinePurgeRecord(rec)
		if e != nil || rec.Restored || !bytes.Equal(body, a.Inspection.RecordData) {
			return result, errors.Join(e, errors.New("fresh quarantine record changed before publication"))
		}
	}
	pending := func(cause error) (FreshRestoreResult, error) {
		return p.result(m), &RestorePendingError{a.QuarantineID, filepath.Join(parent.path, string(a.Inspection.Destination.Leaf)), cause}
	}
	if newOperation {
		op, m, err = p.prepare(parent, old)
		if err != nil {
			return pending(err)
		}
		defer func() {
			if op != nil {
				op.Close()
			}
		}()
	}
	if archived {
		ok, e := c.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID)
		if e != nil || !ok {
			return pending(errors.Join(e, errors.New("fresh completion receipt is missing")))
		}
		return p.result(m), nil
	}
	if err = p.checkOwnedOperation(op, m); err != nil {
		return pending(err)
	}
	if m.Phase == "copying" || m.Phase == "verified" {
		if err = p.copyAndPublish(s, parent, op, input, &m); err != nil {
			return pending(err)
		}
	} else if m.Phase == "publishing" {
		stage, e := p.openStage(parent, m)
		if e != nil {
			return pending(e)
		}
		source, e1 := inspectPurgeArtifact(stage, "snapshot")
		target, e2 := inspectPurgeArtifact(parent, string(m.Leaf))
		stage.Close()
		if e1 != nil || e2 != nil || source.State != "absent" || target.State == "absent" || target.ID != m.OutputID {
			return pending(errors.Join(e1, e2, errors.New("fresh publication is ambiguous; both source and destination are retained")))
		}
		if err = admissionSyncDirectory(parent); err != nil {
			return pending(err)
		}
		m.Phase = "published"
		if err = writeRestoreManifest(op, m); err != nil {
			return pending(err)
		}
	}
	if m.Phase != "published" && m.Phase != "committed" {
		return pending(errors.New("fresh execution phase cannot be replayed"))
	}
	if err = p.finish(s, parent, op, &m); err != nil {
		return pending(err)
	}
	// Windows directory inspection handles must be closed before the exact
	// same operation directory is renamed by its separate DELETE handle.
	op.Close()
	op = nil
	if err = p.archive(m); err != nil {
		return pending(err)
	}
	return p.result(m), nil
}

func (p *freshRestorePermit) result(m restoreManifest) FreshRestoreResult {
	a := p.approval
	return FreshRestoreResult{Version: 1, Kind: "fresh-restore-result", QuarantineID: a.QuarantineID, OperationID: a.OperationID, ApprovalSHA256: p.approvalHash, Destination: []byte(filepath.Join(string(a.Inspection.Destination.ParentPath), string(a.Inspection.Destination.Leaf))), ParentID: a.Inspection.Destination.ParentID, OutputID: m.OutputID, Phase: m.Phase}
}

func (p *freshRestorePermit) checkAuthority() (restoreManifest, error) {
	var original restoreManifest
	a := p.approval
	if err := p.root.stillNamed(); err != nil {
		return original, err
	}
	if err := p.decision.stillNamed(); err != nil {
		return original, err
	}
	entry, err := readFreshControl(p.decision, "manifest.json", maxFreshRestoreApproval)
	if err != nil || entry.ManifestID != a.FileID || entry.SHA256 != p.approvalHash {
		return original, errors.Join(err, errors.New("fresh permission changed before action"))
	}
	var hold RecoveryHold
	if err = json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
		return original, err
	}
	approval := cancellationApproval(hold, a.Inspection.Hold)
	if err = validRecoveryReleaseApproval(approval); err != nil {
		return original, err
	}
	history, err := admissionOpenChild(p.root, ".reviews")
	if err != nil {
		return original, err
	}
	defer history.Close()
	cancel, err := admissionOpenChild(history, hold.DecisionID)
	if err != nil {
		return original, err
	}
	defer cancel.Close()
	m, err := readRecoveryCancellation(p.root, cancel, a.QuarantineID, approval)
	if err != nil {
		return original, err
	}
	if err = cancellationFence(p.root, m); err != nil {
		return original, err
	}
	archived, err := openBoundCancelledOriginal(cancel, m)
	if err != nil {
		return original, err
	}
	defer archived.Close()
	entry, err = readFreshControl(cancel, "manifest.json", maxRecoveryRelease)
	if err != nil || entry.ManifestID != a.Inspection.Cancellation.ManifestID || entry.SHA256 != a.Inspection.Cancellation.SHA256 || m.DirectoryID != a.Inspection.Cancellation.DirectoryID {
		return original, errors.Join(err, errors.New("fresh cancellation evidence changed"))
	}
	if !bytes.Equal(hold.Original.Manifest, a.Inspection.Original.Manifest) || hold.Original.DirectoryID != a.Inspection.Original.DirectoryID || hold.Original.ManifestID != a.Inspection.Original.ManifestID {
		return original, errors.New("fresh original evidence changed")
	}
	original, err = decodeRestoreManifest(hold.Original.Manifest, a.QuarantineID)
	if err != nil || !bytes.Equal(original.RecordData, a.Inspection.RecordData) || original.Source.State != "ready" || original.Source.ID != a.Inspection.Snapshot.FileID {
		return original, errors.Join(err, errors.New("fresh retained record changed"))
	}
	return original, nil
}

func freshKnownEntries(d *restoreDirectory, allowed map[string]bool, allowJournalTemps ...bool) error {
	fresh, err := openAdmissionRoot(d.path, false)
	if err != nil {
		return err
	}
	defer fresh.Close()
	before, e1 := admissionFileID(d.file)
	after, e2 := admissionFileID(fresh.file)
	if e1 != nil || e2 != nil || before != after {
		return errors.Join(e1, e2, errors.New("fresh directory identity changed"))
	}
	entries, err := fresh.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > maxAdmissionRecovery {
		return errors.Join(err, errors.New("fresh directory exceeds inspection bound"))
	}
	for _, entry := range entries {
		directory, known := allowed[entry.Name()]
		if !known && len(allowJournalTemps) == 1 && allowJournalTemps[0] && purgeJournalTemporary(entry.Name()) {
			directory, known = false, true
		}
		if !known || directory != entry.IsDir() || !directory && !entry.Type().IsRegular() {
			return errors.New("unknown fresh operation child retained")
		}
	}
	return nil
}

func freshKnownOperation(op *restoreDirectory) error {
	return freshKnownEntries(op, map[string]bool{"manifest.json": false, freshTicketTemporary: false}, true)
}

func (p *freshRestorePermit) checkOwnedOperation(op *restoreDirectory, m restoreManifest) error {
	if _, err := p.checkAuthority(); err != nil {
		return err
	}
	if err := validateFreshJournal(m, p.approval, p.owner); err != nil {
		return err
	}
	owner, _, err := readFreshOwnership(p.decision, p.approval, p.approvalHash)
	if err != nil {
		return err
	}
	_, before, e1 := freshJSONHash(p.owner)
	_, after, e2 := freshJSONHash(owner)
	if e1 != nil || e2 != nil || !bytes.Equal(before, after) {
		return errors.Join(e1, e2, errors.New("fresh ownership changed before action"))
	}
	id, err := admissionFileID(op.file)
	if err != nil || id != p.owner.OperationDirectoryID {
		return errors.Join(err, errors.New("fresh operation directory changed"))
	}
	if err = op.stillNamed(); err != nil {
		return err
	}
	entry, err := readFreshControl(op, "manifest.json", maxAdmissionJournal)
	hash, _, encodeErr := freshJSONHash(m)
	if err != nil || encodeErr != nil || entry.SHA256 != hash {
		return errors.Join(err, encodeErr, errors.New("fresh mutable journal changed before action"))
	}
	return freshKnownOperation(op)
}

func (p *freshRestorePermit) openStage(parent *restoreDirectory, m restoreManifest) (*restoreDirectory, error) {
	stage, err := admissionOpenChild(parent, m.stageName())
	if err != nil {
		return nil, err
	}
	id, err := admissionFileID(stage.file)
	if err != nil || id != p.owner.StageID || id != m.StageID {
		stage.Close()
		return nil, errors.Join(err, errors.New("fresh stage identity changed"))
	}
	return stage, nil
}
