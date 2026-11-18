package quarantine

import (
	"bytes"
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

const retirementReceiptName = "retirement.json"
const retirementPreparedName = ".retirement-guard"
const retirementTicketName = ".edr-retired-restore-owner.json"
const maxRetirementReceipt = maxRetirementInspection + (64 << 10)
const maxRetirementMarker = 8 << 10

var ErrRecoveryRetired = errors.New("cancelled recovery is permanently retired; deletion requires separate review")

// CancelledRecoveryRetirement closes a cancelled original's active hold while
// permanently forbidding ordinary replay. It grants no purge, publication or
// later fresh-execution permission; every original/new data artifact is retained.
type CancelledRecoveryRetirement struct {
	Version       int                           `json:"version"`
	Kind          string                        `json:"kind"`
	DecisionID    string                        `json:"decision_id"`
	RootID        string                        `json:"root_id"`
	QuarantineID  string                        `json:"quarantine_id"`
	HoldID        string                        `json:"hold_id"`
	DirectoryID   string                        `json:"directory_id"`
	FileID        string                        `json:"file_id"`
	MarkerID      string                        `json:"marker_id"`
	StageTicketID string                        `json:"stage_ticket_id"`
	ApprovedAt    time.Time                     `json:"approved_at"`
	ActorID       string                        `json:"actor_id"`
	Reason        string                        `json:"reason"`
	Inspection    CancelledRetirementInspection `json:"inspection"`
}

type retirementMarker struct {
	Version       int    `json:"version"`
	Kind          string `json:"kind"`
	RootPath      []byte `json:"root_path"`
	RootID        string `json:"root_id"`
	QuarantineID  string `json:"quarantine_id"`
	HoldID        string `json:"hold_id"`
	DirectoryID   string `json:"directory_id"`
	FileID        string `json:"file_id"`
	ReceiptID     string `json:"receipt_id"`
	ReceiptSHA256 string `json:"receipt_sha256"`
	OldStageID    string `json:"old_stage_id"`
	OldTicketID   string `json:"old_ticket_id"`
}

type retirementStageTicket struct {
	Version       int    `json:"version"`
	Kind          string `json:"kind"`
	RootPath      []byte `json:"root_path"`
	RootID        string `json:"root_id"`
	QuarantineID  string `json:"quarantine_id"`
	HoldID        string `json:"hold_id"`
	StageID       string `json:"stage_id"`
	FileID        string `json:"file_id"`
	MarkerID      string `json:"marker_id"`
	ReceiptID     string `json:"receipt_id"`
	ReceiptSHA256 string `json:"receipt_sha256"`
}

// RetireCancelledRecovery consumes the complete saved inspection. stageLocator
// may name a subsequently renamed same stage; it is a locator, never a substitute
// identity or a change to the recorded approval. Empty uses the inspected path.
func RetireCancelledRecovery(dir string, r CancelledRetirementInspection, reason, stageLocator string) (CancelledRecoveryRetirement, error) {
	return retireCancelledRecovery(dir, r, reason, stageLocator, nil)
}

func retireCancelledRecovery(dir string, r CancelledRetirementInspection, reason, stageLocator string, fault func(string) error) (m CancelledRecoveryRetirement, err error) {
	if err = validRetirementInspection(r); err != nil || !validReviewReason(reason) {
		return m, errors.Join(err, errors.New("retirement requires exact saved evidence and a bounded new reason"))
	}
	actor, err := user.Current()
	if err != nil || !freshBoundID(actor.Uid) {
		return m, errors.Join(err, errors.New("cannot determine retirement actor"))
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return m, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return m, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return m, err
	}
	defer root.Close()
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != r.RootID {
		return m, errors.Join(err, errors.New("retirement root changed"))
	}
	unlock, err := root.lockItem(r.QuarantineID)
	if err != nil {
		return m, err
	}
	defer unlock()
	historyRoot, err := admissionOpenChild(root, ".reviews")
	if err != nil {
		return m, err
	}
	defer historyRoot.Close()
	history, err := admissionOpenChild(historyRoot, r.HoldID)
	if err != nil {
		return m, err
	}
	defer history.Close()
	seam := func(at string) error {
		if fault != nil {
			return fault("retirement-" + at)
		}
		return nil
	}
	entry, err := readFreshControl(history, retirementReceiptName, maxRetirementReceipt)
	if err == nil {
		m, err = decodeRetirementReceipt(root, history, entry)
		if err != nil {
			return m, err
		}
		if m.ActorID != actor.Uid || m.Reason != reason || !sameCancelledRetirementInspection(m.Inspection, r, false) {
			return m, errors.New("existing retirement actor, reason or exact inspection changed")
		}
		return finishRetirement(root, history, m, stageLocator, seam)
	}
	if !os.IsNotExist(err) {
		return m, err
	}
	if stageLocator == "" {
		stageLocator = string(r.OldStage.Path)
	}
	if stageLocator != string(r.OldStage.Path) {
		return m, errors.New("first retirement decision requires the inspected literal stage; inspect its new location first")
	}
	current, err := inspectCancelledRetirement(root, CancelledRetirementInspection{Version: 1, Kind: "cancelled-retirement-inspection", QuarantineID: r.QuarantineID, HoldID: r.HoldID, ObservedAt: time.Now().UTC(), Fresh: []RetirementFreshOperation{}}, stageLocator)
	if err != nil || !sameCancelledRetirementInspection(current, r, true) {
		return m, errors.Join(err, errors.New("retirement inspection is stale; original fence retained"))
	}
	stage, err := openAdmissionRoot(stageLocator, false)
	if err != nil {
		return m, err
	}
	defer stage.Close()
	stageID, err := admissionFileID(stage.file)
	if err != nil || stageID != r.OldStage.DirectoryID {
		return m, errors.Join(err, errors.New("original stage changed before retirement preparation"))
	}
	for _, name := range []string{retirementReceiptName, retirementPreparedName, "hold.json"} {
		if err = requirePurgeNameAbsent(history, name); err != nil {
			return m, err
		}
	}
	if err = requirePurgeNameAbsent(stage, retirementTicketName); err != nil {
		return m, err
	}
	f, err := admissionCreateFile(history, retirementReceiptName)
	if err != nil {
		return m, err
	}
	defer f.Close()
	m = CancelledRecoveryRetirement{Version: 1, Kind: "cancelled-restore-retirement", DecisionID: uuid.NewString(), RootID: r.RootID, QuarantineID: r.QuarantineID, HoldID: r.HoldID, DirectoryID: r.HistoryDirectoryID, ApprovedAt: time.Now().UTC(), ActorID: actor.Uid, Reason: reason, Inspection: r}
	if m.FileID, err = admissionFileID(f); err != nil {
		return m, err
	}
	if err = seam("receipt-created"); err != nil {
		return m, err
	}
	create := func(parent *restoreDirectory, name string) (string, error) {
		p, e := admissionCreateFile(parent, name)
		if e != nil {
			return "", e
		}
		id, e := admissionFileID(p)
		e = errors.Join(e, p.Sync(), p.Close())
		if e == nil {
			e = admissionSyncDirectory(parent)
		}
		return id, e
	}
	if m.MarkerID, err = create(history, retirementPreparedName); err != nil {
		return m, err
	}
	if err = seam("marker-created"); err != nil {
		return m, err
	}
	if m.StageTicketID, err = create(stage, retirementTicketName); err != nil {
		return m, err
	}
	if err = seam("ticket-created"); err != nil {
		return m, err
	}
	body, err := json.Marshal(m)
	if err != nil || len(body) > maxRetirementReceipt {
		return m, errors.Join(err, errors.New("retirement decision exceeds metadata bound"))
	}
	if _, err = f.Write(body); err != nil {
		return m, err
	}
	if err = seam("receipt-written"); err != nil {
		return m, err
	}
	if err = f.Sync(); err == nil {
		err = admissionSyncDirectory(history)
	}
	if err != nil {
		return m, err
	}
	if err = f.Close(); err != nil {
		return m, err
	}
	if err = seam("approved"); err != nil {
		return m, err
	}
	stage.Close()
	return finishRetirement(root, history, m, stageLocator, seam)
}

func validRetirementInspection(r CancelledRetirementInspection) error {
	if r.Version != 1 || r.Kind != "cancelled-retirement-inspection" || !ordinaryQuarantineID(r.QuarantineID) || !validReviewUUID(r.HoldID) || r.ObservedAt.IsZero() || !freshBoundID(r.RootID) || !freshBoundID(r.RestoreQueueID) || !freshBoundID(r.HistoryDirectoryID) || !freshBoundID(r.OldStage.DirectoryID) || freshLocalDestination(string(r.OldStage.Path)) != nil || len(r.Fresh) > maxAdmissionRecovery {
		return errors.New("invalid complete cancelled-retirement inspection")
	}
	for _, p := range []FreshRestoreInput{r.Raw, r.Snapshot} {
		if !freshBoundID(p.FileID) || p.Size < 0 || p.Size > maxRestoreBytes {
			return errors.New("invalid retirement data observation")
		}
		if _, e := restoreDigest(p.SHA256); e != nil {
			return e
		}
	}
	validPin := func(p RetirementControl, limit int64) bool {
		if p.Status == "absent" {
			return p.FileID == "" && p.SHA256 == "" && p.Bytes == 0
		}
		_, e := restoreDigest(p.SHA256)
		return p.Status == "captured" && freshBoundID(p.FileID) && p.Bytes >= 0 && p.Bytes <= limit && e == nil
	}
	if (r.FreshQueueID == "" || r.FreshItemID == "") && len(r.Fresh) != 0 {
		return errors.New("retirement history lacks its parent identities")
	}
	for i, f := range r.Fresh {
		if !validReviewUUID(f.OperationID) || !freshBoundID(f.DirectoryID) || i > 0 && r.Fresh[i-1].OperationID >= f.OperationID || !validPin(f.Approval, maxFreshRestoreApproval) || f.Approval.Status != "captured" || !validPin(f.Ownership, maxFreshOwnership) || !validPin(f.Journal, maxAdmissionJournal) || !validPin(f.Cancellation, maxFreshCancellation) {
			return errors.New("invalid retirement fresh-history binding")
		}
		switch f.State {
		case "approval-only":
			if f.ArchiveID != "" || f.StageID != "" || f.TicketID != "" || f.Ownership.Status != "absent" || f.Cancellation.Status != "absent" || f.Journal.Status != "absent" {
				return errors.New("unused retirement approval claims an operation")
			}
		case "committed", "cancelled":
			if !freshBoundID(f.ArchiveID) || !freshBoundID(f.StageID) || !freshBoundID(f.TicketID) || f.Ownership.Status != "captured" || f.State == "committed" && (f.Journal.Status != "captured" || f.Cancellation.Status != "absent") || f.State == "cancelled" && f.Cancellation.Status != "captured" {
				return errors.New("retirement fresh archive lacks exact ownership or disposition")
			}
		default:
			return errors.New("unknown retirement fresh-history state")
		}
	}
	body, e := json.Marshal(r)
	if e != nil || len(body) > maxRetirementInspection {
		return errors.Join(e, errors.New("saved retirement inspection exceeds bound"))
	}
	return nil
}

func decodeRetirementReceipt(root, history *restoreDirectory, entry JournalInspection) (m CancelledRecoveryRetirement, err error) {
	if err = json.Unmarshal(entry.Manifest, &m); err != nil {
		return m, err
	}
	rid, e1 := admissionFileID(root.file)
	hid, e2 := admissionFileID(history.file)
	if err = errors.Join(e1, e2); err != nil {
		return m, err
	}
	if m.Version != 1 || m.Kind != "cancelled-restore-retirement" || !validReviewUUID(m.DecisionID) || m.RootID != rid || m.RootID != m.Inspection.RootID || m.QuarantineID != m.Inspection.QuarantineID || m.HoldID != m.Inspection.HoldID || m.DirectoryID != hid || hid != m.Inspection.HistoryDirectoryID || m.FileID != entry.ManifestID || !freshBoundID(m.MarkerID) || !freshBoundID(m.StageTicketID) || m.FileID == m.MarkerID || m.MarkerID == m.StageTicketID || m.FileID == m.StageTicketID || m.ApprovedAt.IsZero() || !freshBoundID(m.ActorID) || !validReviewReason(m.Reason) {
		return m, errors.New("retirement receipt identity changed")
	}
	if err = validRetirementInspection(m.Inspection); err != nil {
		return m, err
	}
	_, body, err := freshJSONHash(m)
	if err != nil || !bytes.Equal(body, entry.Manifest) {
		return m, errors.Join(err, errors.New("retirement receipt is not its exact sealed representation"))
	}
	return m, nil
}

func expectedRetirementMarker(root *restoreDirectory, m CancelledRecoveryRetirement) (retirementMarker, []byte, error) {
	h, _, e := freshJSONHash(m)
	if e != nil {
		return retirementMarker{}, nil, e
	}
	v := retirementMarker{Version: 1, Kind: "cancelled-retirement-guard", RootPath: []byte(root.path), RootID: m.RootID, QuarantineID: m.QuarantineID, HoldID: m.HoldID, DirectoryID: m.DirectoryID, FileID: m.MarkerID, ReceiptID: m.FileID, ReceiptSHA256: h, OldStageID: m.Inspection.OldStage.DirectoryID, OldTicketID: m.StageTicketID}
	body, e := json.Marshal(v)
	return v, body, e
}

func expectedRetirementTicket(root *restoreDirectory, m CancelledRecoveryRetirement) (retirementStageTicket, []byte, error) {
	h, _, e := freshJSONHash(m)
	if e != nil {
		return retirementStageTicket{}, nil, e
	}
	v := retirementStageTicket{Version: 1, Kind: "retired-original-stage", RootPath: []byte(root.path), RootID: m.RootID, QuarantineID: m.QuarantineID, HoldID: m.HoldID, StageID: m.Inspection.OldStage.DirectoryID, FileID: m.StageTicketID, MarkerID: m.MarkerID, ReceiptID: m.FileID, ReceiptSHA256: h}
	body, e := json.Marshal(v)
	return v, body, e
}

// A complete immutable decision already owns these newly created file IDs.
// Retry may only append missing canonical bytes to its own prepared controls;
// different bytes are retained, never truncated, replaced or interpreted.
func completeRetirementControl(d *restoreDirectory, name, id string, body []byte) error {
	f, err := openFreshWritableFile(d, name)
	if err != nil {
		return err
	}
	defer f.Close()
	e, err := inspectJournalFile(JournalInspection{}, f, maxRetirementMarker)
	if err != nil || e.ManifestID != id || len(body) > maxRetirementMarker || !bytes.HasPrefix(body, e.Manifest) {
		return errors.Join(err, errors.New("retirement preparation identity or existing bytes changed"))
	}
	if len(e.Manifest) < len(body) {
		if _, err = f.WriteAt(body[len(e.Manifest):], int64(len(e.Manifest))); err != nil {
			return err
		}
	}
	if err = f.Sync(); err != nil {
		return err
	}
	// inspectJournalFile reads from the handle's current position. WriteAt
	// preserves that position, including the end reached by an exact retry.
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	check, err := inspectJournalFile(JournalInspection{}, f, maxRetirementMarker)
	if err != nil || check.ManifestID != id || !bytes.Equal(check.Manifest, body) {
		return errors.Join(err, errors.New("retirement control did not retain its exact bytes"))
	}
	return admissionSyncDirectory(d)
}

func finishRetirement(root, history *restoreDirectory, m CancelledRecoveryRetirement, stageLocator string, seam func(string) error) (CancelledRecoveryRetirement, error) {
	archived, err := retirementCore(root, history, m)
	if err != nil {
		return m, err
	}
	queue, err := admissionOpenChild(root, ".restores")
	if err != nil {
		return m, err
	}
	defer queue.Close()
	qid, err := admissionFileID(queue.file)
	if err != nil || qid != m.Inspection.RestoreQueueID {
		return m, errors.Join(err, errors.New("retirement compatibility queue changed"))
	}
	_, markerBody, err := expectedRetirementMarker(root, m)
	if err != nil {
		return m, err
	}
	live, e1 := freshOptionalControl(queue, m.QuarantineID, maxRetirementMarker)
	prepared, e2 := freshOptionalControl(history, retirementPreparedName, maxRetirementMarker)
	if e1 != nil || e2 != nil {
		return m, errors.Join(e1, e2)
	}
	published := live.Status == "captured" && live.ManifestID == m.MarkerID && prepared.Status == "absent"
	if !published && !(live.Status == "absent" && prepared.Status == "captured" && prepared.ManifestID == m.MarkerID) {
		return m, errors.New("retirement guard must exist in exactly one bound location")
	}
	if archived {
		if !published || !bytes.Equal(live.Manifest, markerBody) {
			return m, errors.New("archived original hold lacks its exact permanent compatibility guard")
		}
		if err = syncRetirementMetadata(root, history, queue, m); err != nil {
			return m, err
		}
		return m, seam("synced")
	}
	if err = validateRetirementCurrentData(root, history, m, stageLocator); err != nil {
		return m, err
	}
	if stageLocator == "" {
		stageLocator = string(m.Inspection.OldStage.Path)
	}
	stage, err := openAdmissionRoot(stageLocator, false)
	if err != nil {
		return m, err
	}
	defer stage.Close()
	sid, err := admissionFileID(stage.file)
	if err != nil || sid != m.Inspection.OldStage.DirectoryID {
		return m, errors.Join(err, errors.New("retirement retained-stage locator changed"))
	}
	_, ticketBody, err := expectedRetirementTicket(root, m)
	if err != nil {
		return m, err
	}
	if err = syncRetirementReceipt(history, m); err != nil {
		return m, err
	}
	if err = completeRetirementControl(stage, retirementTicketName, m.StageTicketID, ticketBody); err != nil {
		return m, err
	}
	if err = stage.stillNamed(); err != nil {
		return m, err
	}
	if err = seam("ticket-prepared"); err != nil {
		return m, err
	}
	if !published {
		if err = completeRetirementControl(history, retirementPreparedName, m.MarkerID, markerBody); err != nil {
			return m, err
		}
		if err = seam("marker-prepared"); err != nil {
			return m, err
		}
		if err = validateRetirementCurrentData(root, history, m, stageLocator); err != nil {
			return m, err
		}
		if err = admissionRename(history, retirementPreparedName, queue, m.QuarantineID, false); err != nil {
			return m, err
		}
		if err = admissionSyncDirectory(queue); err == nil {
			err = admissionSyncDirectory(history)
		}
		if err != nil {
			return m, err
		}
		if err = seam("marker-published"); err != nil {
			return m, err
		}
	} else if !bytes.Equal(live.Manifest, markerBody) {
		return m, errors.New("published retirement guard changed")
	}
	if err = validateRetirementCurrentData(root, history, m, stageLocator); err != nil {
		return m, err
	}
	if err = seam("before-hold-move"); err != nil {
		return m, err
	}
	if err = validateRetirementCurrentData(root, history, m, stageLocator); err != nil {
		return m, err
	}
	if err = stage.stillNamed(); err != nil {
		return m, err
	}
	ticket, err := readFreshControl(stage, retirementTicketName, maxRetirementMarker)
	if err != nil || ticket.ManifestID != m.StageTicketID || !bytes.Equal(ticket.Manifest, ticketBody) {
		return m, errors.Join(err, errors.New("retirement stage protection changed before original fence move"))
	}
	// Revalidate after the fault seam as well: no callback, stale DB or copied
	// evidence may substitute another fence immediately before its move.
	if archived, err = retirementCore(root, history, m); err != nil || archived {
		return m, errors.Join(err, errors.New("retirement hold location changed before archive"))
	}
	actual, err := readFreshControl(queue, m.QuarantineID, maxRetirementMarker)
	if err != nil || actual.ManifestID != m.MarkerID || !bytes.Equal(actual.Manifest, markerBody) {
		return m, errors.Join(err, errors.New("permanent guard changed before archiving original hold"))
	}
	purges, err := admissionOpenChild(root, ".purges")
	if err != nil {
		return m, err
	}
	defer purges.Close()
	if err = admissionRename(purges, m.QuarantineID, history, "hold.json", false); err != nil {
		return m, fmt.Errorf("retirement approved; hold archive incomplete: %w", err)
	}
	if err = seam("hold-archived"); err != nil {
		return m, err
	}
	if err = syncRetirementMetadata(root, history, queue, m); err != nil {
		return m, err
	}
	if archived, err = retirementCore(root, history, m); err != nil || !archived {
		return m, errors.Join(err, errors.New("retirement original hold archive not proven"))
	}
	return m, seam("synced")
}

func syncRetirementReceipt(history *restoreDirectory, m CancelledRecoveryRetirement) error {
	f, e := openFreshWritableFile(history, retirementReceiptName)
	if e != nil {
		return e
	}
	defer f.Close()
	entry, e := inspectJournalFile(JournalInspection{}, f, maxRetirementReceipt)
	h, _, enc := freshJSONHash(m)
	if e != nil || enc != nil || entry.ManifestID != m.FileID || entry.SHA256 != h {
		return errors.Join(e, enc, errors.New("retirement decision changed before flush"))
	}
	if e = f.Sync(); e != nil {
		return e
	}
	return admissionSyncDirectory(history)
}

func syncRetirementMetadata(root, history, queue *restoreDirectory, m CancelledRecoveryRetirement) error {
	if e := syncRetirementReceipt(history, m); e != nil {
		return e
	}
	purges, e := admissionOpenChild(root, ".purges")
	if e != nil {
		return e
	}
	defer purges.Close()
	return errors.Join(admissionSyncDirectory(queue), admissionSyncDirectory(purges), admissionSyncDirectory(history))
}
