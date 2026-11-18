package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Historical ownership uses the fixed qid guard and exact original hold UUID.
// It never searches global history, opens SQLite or derives completion from an
// absent live operation, a payload hash, or a database Restored flag.
func retirementCore(root, history *restoreDirectory, m CancelledRecoveryRetirement) (bool, error) {
	r := m.Inspection
	if err := root.stillNamed(); err != nil {
		return false, err
	}
	if err := history.stillNamed(); err != nil {
		return false, err
	}
	rootID, e1 := admissionFileID(root.file)
	hid, e2 := admissionFileID(history.file)
	if e1 != nil || e2 != nil || rootID != m.RootID || hid != m.DirectoryID {
		return false, errors.Join(e1, e2, errors.New("retirement private namespace changed"))
	}
	var hold RecoveryHold
	if err := json.Unmarshal(r.Hold.Manifest, &hold); err != nil {
		return false, err
	}
	a := cancellationApproval(hold, r.Hold)
	if err := validRecoveryReleaseApproval(a); err != nil {
		return false, err
	}
	if _, err := decodeReleaseHold(r.Hold, m.QuarantineID, a); err != nil {
		return false, err
	}
	if hold.DecisionID != m.HoldID || hold.RootID != m.RootID || hold.Original.DirectoryID != r.Original.DirectoryID || hold.Original.ManifestID != r.Original.ManifestID || hold.Original.SHA256 != r.Original.SHA256 || !bytes.Equal(hold.Original.Manifest, r.Original.Manifest) {
		return false, errors.New("retirement original hold binding changed")
	}
	old, err := decodeRestoreManifest(r.Original.Manifest, m.QuarantineID)
	if err != nil || old.Version != 1 || old.StageID != r.OldStage.DirectoryID || old.Source.State != "ready" || old.Source.ID != r.Snapshot.FileID || old.Record.SHA256 != r.Snapshot.SHA256 || old.Record.SourceFileID != r.Raw.FileID || r.Raw.FileID == r.Snapshot.FileID || r.Snapshot.FileID == old.OutputID {
		return false, errors.Join(err, errors.New("retirement independent original evidence identities changed"))
	}
	cancel, err := readCancellationRecord(history, m.QuarantineID, a)
	if err != nil {
		return false, err
	}
	entry, err := readFreshControl(history, "manifest.json", maxRecoveryRelease)
	if err != nil || entry.ManifestID != r.Cancellation.ManifestID || entry.SHA256 != r.Cancellation.SHA256 || cancel.DirectoryID != r.HistoryDirectoryID || r.Cancellation.DirectoryID != r.HistoryDirectoryID {
		return false, errors.Join(err, errors.New("original cancellation receipt changed after retirement approval"))
	}
	original, err := openBoundCancelledOriginal(history, cancel)
	if err != nil {
		return false, err
	}
	original.Close()
	purges, err := admissionOpenChild(root, ".purges")
	if err != nil {
		return false, err
	}
	defer purges.Close()
	pid, err := admissionFileID(purges.file)
	if err != nil || pid != r.Hold.DirectoryID {
		return false, errors.Join(err, errors.New("original retention-fence namespace changed"))
	}
	active, e1 := freshOptionalControl(purges, m.QuarantineID, maxRecoveryReview)
	archived, e2 := freshOptionalControl(history, "hold.json", maxRecoveryReview)
	if e1 != nil || e2 != nil {
		return false, errors.Join(e1, e2)
	}
	match := func(e JournalInspection) bool {
		return e.Status == "captured" && e.ManifestID == r.Hold.ManifestID && e.SHA256 == r.Hold.SHA256 && bytes.Equal(e.Manifest, r.Hold.Manifest)
	}
	if match(active) && archived.Status == "absent" {
		return false, nil
	}
	if active.Status == "absent" && match(archived) {
		return true, nil
	}
	return false, errors.New("same original retirement fence must exist in exactly one approved location")
}

func validateRetirementCurrentData(root, history *restoreDirectory, m CancelledRecoveryRetirement, stageLocator string) error {
	if archived, err := retirementCore(root, history, m); err != nil || archived {
		return errors.Join(err, errors.New("active original fence changed during retirement"))
	}
	r := m.Inspection
	old, err := decodeRestoreManifest(r.Original.Manifest, m.QuarantineID)
	if err != nil {
		return err
	}
	if stageLocator == "" {
		stageLocator = string(r.OldStage.Path)
	}
	if freshLocalDestination(stageLocator) != nil {
		return errors.New("retirement stage locator is not an ordinary local path")
	}
	stage, err := openAdmissionRoot(stageLocator, false)
	if err != nil {
		return err
	}
	sid, e := admissionFileID(stage.file)
	closeErr := stage.Close()
	if e != nil || closeErr != nil || sid != r.OldStage.DirectoryID {
		return errors.Join(e, closeErr, errors.New("retained original stage was replaced"))
	}
	input, err := restoreOpenInput(root, m.QuarantineID+".quarantined")
	if err != nil {
		return err
	}
	snapshot, err := inspectFreshInput(input)
	err = errors.Join(err, input.Close())
	if err != nil || snapshot != r.Snapshot {
		return errors.Join(err, errors.New("retirement snapshot observation is stale"))
	}
	capture, err := openAdmissionRoot(old.Record.SourceCaptureDir, false)
	if err != nil {
		return err
	}
	defer capture.Close()
	raw, err := restoreOpenInput(capture, "source")
	if err != nil {
		return err
	}
	rawPin, err := inspectFreshInput(raw)
	err = errors.Join(err, raw.Close())
	if err != nil || rawPin != r.Raw {
		return errors.Join(err, errors.New("retirement original raw observation is stale"))
	}
	current := r
	current.FreshQueueID = ""
	current.FreshItemID = ""
	current.Fresh = []RetirementFreshOperation{}
	budget := int64(maxRetirementRead - len(r.Hold.Manifest) - len(r.Cancellation.Manifest) - len(r.Original.Manifest))
	if err = inspectRetirementFresh(root, &current, &budget); err != nil {
		return err
	}
	if !sameCancelledRetirementInspection(current, r, false) {
		return errors.New("retirement fresh history changed after review")
	}
	return nil
}

func readRetirementMarker(root *restoreDirectory, id string) (v retirementMarker, entry JournalInspection, found bool, err error) {
	queue, err := admissionOpenChild(root, ".restores")
	if os.IsNotExist(err) {
		return v, entry, false, nil
	}
	if err != nil {
		return v, entry, false, err
	}
	defer queue.Close()
	op, e := admissionOpenChild(queue, id)
	if e == nil {
		op.Close()
		return v, entry, false, nil
	}
	if os.IsNotExist(e) {
		return v, entry, false, nil
	}
	entry, err = readFreshControl(queue, id, maxRetirementMarker)
	if err != nil {
		return v, entry, true, err
	}
	if err = json.Unmarshal(entry.Manifest, &v); err != nil {
		return v, entry, true, err
	}
	rootID, e := admissionFileID(root.file)
	if e != nil || v.Version != 1 || v.Kind != "cancelled-retirement-guard" || v.RootID != rootID || v.QuarantineID != id || !ordinaryQuarantineID(id) || !validReviewUUID(v.HoldID) || v.FileID != entry.ManifestID || !freshBoundID(v.ReceiptID) || !freshBoundID(v.DirectoryID) || !freshBoundID(v.OldStageID) || !freshBoundID(v.OldTicketID) || !bytes.Equal(v.RootPath, []byte(root.path)) {
		return v, entry, true, errors.Join(e, errors.New("retirement compatibility guard identity changed"))
	}
	if _, e = restoreDigest(v.ReceiptSHA256); e != nil {
		return v, entry, true, e
	}
	_, body, e := freshJSONHash(v)
	if e != nil || !bytes.Equal(body, entry.Manifest) {
		return v, entry, true, errors.Join(e, errors.New("retirement compatibility guard bytes changed"))
	}
	return v, entry, true, nil
}

func openRetirementProof(root *restoreDirectory, id string) (m CancelledRecoveryRetirement, history *restoreDirectory, found bool, err error) {
	marker, entry, found, err := readRetirementMarker(root, id)
	if err != nil || !found {
		return m, nil, found, err
	}
	reviews, err := admissionOpenChild(root, ".reviews")
	if err != nil {
		return m, nil, true, err
	}
	defer reviews.Close()
	history, err = admissionOpenChild(reviews, marker.HoldID)
	if err != nil {
		return m, nil, true, err
	}
	ok := false
	openedHistory := history
	defer func() {
		if !ok {
			openedHistory.Close()
			history = nil
		}
	}()
	receipt, err := readFreshControl(history, retirementReceiptName, maxRetirementReceipt)
	if err != nil || receipt.ManifestID != marker.ReceiptID || receipt.SHA256 != marker.ReceiptSHA256 {
		return m, nil, true, errors.Join(err, errors.New("retirement decision referenced by guard changed"))
	}
	m, err = decodeRetirementReceipt(root, history, receipt)
	if err != nil {
		return m, nil, true, err
	}
	_, expected, err := expectedRetirementMarker(root, m)
	if err != nil || !bytes.Equal(entry.Manifest, expected) {
		return m, nil, true, errors.Join(err, errors.New("retirement guard does not match the exact decision"))
	}
	queue, e := admissionOpenChild(root, ".restores")
	if e != nil {
		return m, nil, true, e
	}
	qid, e := admissionFileID(queue.file)
	queue.Close()
	if e != nil || qid != m.Inspection.RestoreQueueID {
		return m, nil, true, errors.Join(e, errors.New("retirement guard parent identity changed"))
	}
	prepared, e := freshOptionalControl(history, retirementPreparedName, maxRetirementMarker)
	if e != nil || prepared.Status != "absent" {
		return m, nil, true, errors.Join(e, errors.New("published retirement guard remains ambiguous"))
	}
	if _, err = retirementCore(root, history, m); err != nil {
		return m, nil, true, err
	}
	ok = true
	return m, history, true, nil
}

// An old v1 stage gets its own pre-publication ticket so its exclusion no longer
// depends on mutable SQLite locators or the still-active original hold.
func managedRetiredOriginalStage(parent *restoreDirectory, parentID string) error {
	private, err := openAdmissionRoot(parent.path, false)
	if err != nil {
		return nil
	}
	defer private.Close()
	id, err := admissionFileID(private.file)
	if err != nil || id != parentID {
		return nil
	}
	entry, err := readFreshControl(private, retirementTicketName, maxRetirementMarker)
	if err != nil {
		return nil
	}
	var ticket retirementStageTicket
	if json.Unmarshal(entry.Manifest, &ticket) != nil || ticket.Version != 1 || ticket.Kind != "retired-original-stage" || ticket.FileID != entry.ManifestID || ticket.StageID != parentID || !ordinaryQuarantineID(ticket.QuarantineID) || !validReviewUUID(ticket.HoldID) || !freshBoundID(ticket.MarkerID) {
		return nil
	}
	path := string(ticket.RootPath)
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || freshLocalDestination(filepath.Join(path, "retirement-root-probe")) != nil {
		return nil
	}
	root, err := openAdmissionRoot(path, false)
	if err != nil {
		return nil
	}
	defer root.Close()
	rid, err := admissionFileID(root.file)
	if err != nil || rid != ticket.RootID {
		return nil
	}
	marker, markerEntry, found, e := readRetirementMarker(root, ticket.QuarantineID)
	fail := func(e error) error { return errors.Join(ErrRestoreIntegrity, e) }
	if !found {
		// A completed stage ticket also protects an interrupted preparation.
		// Its exact receipt supplies a bounded locator before guard publication;
		// a missing guard after the hold moved is an integrity failure.
		return managedPreparedRetirementStage(root, ticket, entry)
	}
	if e != nil {
		return fail(e)
	}
	if marker.OldTicketID != entry.ManifestID || marker.OldStageID != parentID || marker.HoldID != ticket.HoldID || marker.ReceiptID != ticket.ReceiptID || marker.ReceiptSHA256 != ticket.ReceiptSHA256 {
		return nil
	}
	if markerEntry.ManifestID != ticket.MarkerID {
		return fail(errors.New("retired original-stage compatibility guard was replaced"))
	}
	m, history, _, err := openRetirementProof(root, ticket.QuarantineID)
	if err != nil {
		return fail(err)
	}
	defer history.Close()
	_, expected, err := expectedRetirementTicket(root, m)
	if err != nil || !bytes.Equal(expected, entry.Manifest) {
		return fail(errors.Join(err, errors.New("retired original-stage ticket changed")))
	}
	return ErrAlreadyQuarantined
}

func managedPreparedRetirementStage(root *restoreDirectory, ticket retirementStageTicket, entry JournalInspection) error {
	fail := func(e error) error { return errors.Join(ErrRestoreIntegrity, e) }
	reviews, err := admissionOpenChild(root, ".reviews")
	if err != nil {
		return fail(err)
	}
	defer reviews.Close()
	history, err := admissionOpenChild(reviews, ticket.HoldID)
	if err != nil {
		return fail(err)
	}
	defer history.Close()
	receipt, err := readFreshControl(history, retirementReceiptName, maxRetirementReceipt)
	if err != nil || receipt.ManifestID != ticket.ReceiptID || receipt.SHA256 != ticket.ReceiptSHA256 {
		return fail(errors.Join(err, errors.New("prepared retirement decision changed")))
	}
	m, err := decodeRetirementReceipt(root, history, receipt)
	if err != nil {
		return fail(err)
	}
	_, expected, err := expectedRetirementTicket(root, m)
	if err != nil || !bytes.Equal(entry.Manifest, expected) {
		return fail(errors.Join(err, errors.New("prepared retirement ticket does not match its decision")))
	}
	archived, err := retirementCore(root, history, m)
	if err != nil || archived {
		return fail(errors.Join(err, errors.New("retirement guard is absent after original hold archival")))
	}
	queue, err := admissionOpenChild(root, ".restores")
	if err != nil {
		return fail(err)
	}
	defer queue.Close()
	qid, err := admissionFileID(queue.file)
	if err != nil || qid != m.Inspection.RestoreQueueID {
		return fail(errors.Join(err, errors.New("prepared retirement guard namespace changed")))
	}
	if err = requirePurgeNameAbsent(queue, m.QuarantineID); err != nil {
		return fail(err)
	}
	prepared, err := readFreshControl(history, retirementPreparedName, maxRetirementMarker)
	_, expected, encodeErr := expectedRetirementMarker(root, m)
	if err != nil || encodeErr != nil || prepared.ManifestID != m.MarkerID || !bytes.HasPrefix(expected, prepared.Manifest) {
		return fail(errors.Join(err, encodeErr, errors.New("prepared retirement guard changed")))
	}
	return ErrAlreadyQuarantined
}

// A sealed retirement roster remains an ownership anchor even if a later copy
// or damaged fresh approval can no longer pass the live approval reader. Such
// damage is an integrity refusal, never permission to re-admit retained data.
func managedRetiredFreshStage(parent *restoreDirectory, parentID string) error {
	private, e := openAdmissionRoot(parent.path, false)
	if e != nil {
		return nil
	}
	defer private.Close()
	id, e := admissionFileID(private.file)
	if e != nil || id != parentID {
		return nil
	}
	entry, e := readFreshControl(private, freshTicketName, maxFreshOwnership)
	if e != nil {
		return nil
	}
	var ticket freshOwnerTicket
	if json.Unmarshal(entry.Manifest, &ticket) != nil || ticket.Version != 1 || ticket.FileID != entry.ManifestID || ticket.StageID != parentID || !ordinaryQuarantineID(ticket.QuarantineID) || !validReviewUUID(ticket.OperationID) {
		return nil
	}
	path := string(ticket.RootPath)
	if freshLocalDestination(filepath.Join(path, "retirement-root-probe")) != nil {
		return nil
	}
	root, e := openAdmissionRoot(path, false)
	if e != nil {
		return nil
	}
	defer root.Close()
	rid, e := admissionFileID(root.file)
	if e != nil || rid != ticket.RootID {
		return nil
	}
	_, _, found, e := readRetirementMarker(root, ticket.QuarantineID)
	if !found {
		return nil
	}
	fail := func(e error) error { return errors.Join(ErrRestoreIntegrity, e) }
	if e != nil {
		return fail(e)
	}
	m, history, _, e := openRetirementProof(root, ticket.QuarantineID)
	if e != nil {
		return fail(e)
	}
	defer history.Close()
	var selected *RetirementFreshOperation
	for i := range m.Inspection.Fresh {
		f := &m.Inspection.Fresh[i]
		if f.OperationID == ticket.OperationID && f.StageID == parentID && f.TicketID == entry.ManifestID {
			selected = f
			break
		}
	}
	if selected == nil {
		return nil
	}
	f := *selected
	if f.State != "cancelled" {
		return fail(errors.New("retired completed fresh stage unexpectedly still owns an external ticket"))
	}
	queue, item, op, e := openFreshApproval(root, ticket.QuarantineID, ticket.OperationID)
	if e != nil {
		return fail(e)
	}
	defer queue.Close()
	defer item.Close()
	defer op.Close()
	opID, e := admissionFileID(op.file)
	if e != nil || opID != f.DirectoryID {
		return fail(errors.Join(e, errors.New("retired fresh approval directory changed")))
	}
	for _, control := range []struct {
		name  string
		pin   RetirementControl
		limit int64
	}{{"manifest.json", f.Approval, maxFreshRestoreApproval}, {"ownership.json", f.Ownership, maxFreshOwnership}, {"cancellation.json", f.Cancellation, maxFreshCancellation}} {
		actual, e := freshOptionalControl(op, control.name, control.limit)
		if e != nil || retirementPin(actual) != control.pin {
			return fail(errors.Join(e, errors.New("retired fresh control changed")))
		}
	}
	archive, e := admissionOpenChild(op, "restore")
	if e != nil {
		return fail(e)
	}
	defer archive.Close()
	aid, e := admissionFileID(archive.file)
	if e != nil || aid != f.ArchiveID {
		return fail(errors.Join(e, errors.New("retired fresh operation archive changed")))
	}
	journal, e := freshOptionalControl(archive, "manifest.json", maxAdmissionJournal)
	if e != nil || retirementPin(journal) != f.Journal {
		return fail(errors.Join(e, errors.New("retired fresh raw journal changed")))
	}
	ownerEntry, e := readFreshControl(op, "ownership.json", maxFreshOwnership)
	if e != nil {
		return fail(e)
	}
	if ownerEntry.ManifestID != ticket.OwnershipID || ownerEntry.SHA256 != ticket.OwnershipSHA256 {
		return fail(errors.New("retired fresh stage ticket no longer matches sealed ownership"))
	}
	return ErrAlreadyQuarantined
}

// Only the retired decision's exact roster can redirect a v2 ownership lookup
// past the permanent non-directory legacy guard to its existing archive.
func retiredFreshArchive(root, approval *restoreDirectory, o freshOwnership) (bool, error) {
	m, history, found, err := openRetirementProof(root, o.QuarantineID)
	if err != nil || !found {
		return false, err
	}
	defer history.Close()
	for _, f := range m.Inspection.Fresh {
		if f.OperationID != o.OperationID {
			continue
		}
		id, err := admissionFileID(approval.file)
		if err != nil || id != f.DirectoryID || f.DirectoryID != o.ApprovalDirectoryID || f.ArchiveID != o.OperationDirectoryID || f.StageID != o.StageID || f.TicketID != o.TicketID || f.Ownership.FileID != o.FileID {
			return false, errors.Join(err, errors.New("fresh ownership is not the retired decision's exact archived operation"))
		}
		owner, err := readFreshControl(approval, "ownership.json", maxFreshOwnership)
		if err != nil || retirementPin(owner) != f.Ownership {
			return false, errors.Join(err, errors.New("retired fresh ownership changed"))
		}
		return true, nil
	}
	return false, errors.New("fresh operation is not in the exact retired roster")
}
