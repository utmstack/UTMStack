package quarantine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const freshTicketName = ".edr-restore-owner.json"
const freshTicketTemporary = ".fresh-owner-ticket"
const maxFreshOwnership = 64 << 10

// The immutable ownership record is sealed before any output data is created.
// It binds newly created identities to one exact independent approval. Neither
// it nor its external ticket is permission to replay recovery.
type freshOwnership struct {
	Version              int    `json:"version"`
	Kind                 string `json:"kind"`
	RootPath             []byte `json:"root_path"`
	RootID               string `json:"root_id"`
	QuarantineID         string `json:"quarantine_id"`
	OperationID          string `json:"operation_id"`
	ApprovalDirectoryID  string `json:"approval_directory_id"`
	ApprovalFileID       string `json:"approval_file_id"`
	ApprovalSHA256       string `json:"approval_sha256"`
	FileID               string `json:"file_id"`
	OperationDirectoryID string `json:"operation_directory_id"`
	StageID              string `json:"stage_id"`
	TicketID             string `json:"ticket_id"`
	ActorID              string `json:"actor_id"`
	Reason               string `json:"reason"`
}

type freshOwnerTicket struct {
	Version         int    `json:"version"`
	RootPath        []byte `json:"root_path"`
	RootID          string `json:"root_id"`
	QuarantineID    string `json:"quarantine_id"`
	OperationID     string `json:"operation_id"`
	StageID         string `json:"stage_id"`
	FileID          string `json:"file_id"`
	OwnershipID     string `json:"ownership_id"`
	OwnershipSHA256 string `json:"ownership_sha256"`
}

func freshJSONHash(v any) (string, []byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), body, nil
}

func readFreshControl(d *restoreDirectory, name string, limit int64) (JournalInspection, error) {
	f, err := admissionReadFile(d, name)
	if err != nil {
		return JournalInspection{}, err
	}
	entry, err := inspectJournalFile(JournalInspection{}, f, limit)
	return entry, errors.Join(err, f.Close())
}

func readFreshOwnership(approvalOp *restoreDirectory, a FreshRestoreApproval, approvalHash string) (o freshOwnership, entry JournalInspection, err error) {
	entry, err = readFreshControl(approvalOp, "ownership.json", maxFreshOwnership)
	if err != nil {
		return o, entry, err
	}
	o, err = decodeFreshOwnership(approvalOp, a, approvalHash, entry)
	return o, entry, err
}

func decodeFreshOwnership(approvalOp *restoreDirectory, a FreshRestoreApproval, approvalHash string, entry JournalInspection) (o freshOwnership, err error) {
	if err = json.Unmarshal(entry.Manifest, &o); err != nil {
		return o, err
	}
	opID, err := admissionFileID(approvalOp.file)
	if err != nil {
		return o, err
	}
	if o.Version != 1 || o.Kind != "fresh-restore-ownership" || o.RootID != a.RootID || o.QuarantineID != a.QuarantineID || o.OperationID != a.OperationID || o.ApprovalDirectoryID != opID || opID != a.DirectoryID || o.ApprovalFileID != a.FileID || o.ApprovalSHA256 != approvalHash || o.FileID != entry.ManifestID || o.ActorID != a.ActorID || !validReviewReason(o.Reason) || !freshBoundID(o.OperationDirectoryID) || !freshBoundID(o.StageID) || !freshBoundID(o.TicketID) {
		return o, errors.New("fresh ownership identity or approval changed")
	}
	rootPath := string(o.RootPath)
	if !filepath.IsAbs(rootPath) || filepath.Clean(rootPath) != rootPath || freshLocalDestination(filepath.Join(rootPath, "fresh-root-probe")) != nil {
		return o, errors.New("fresh ownership requires an ordinary local root")
	}
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, a.QuarantineID)
	if err != nil || o.OperationDirectoryID == a.Inspection.Original.DirectoryID || o.StageID == old.StageID || o.StageID == a.RootID || o.TicketID == a.Inspection.Snapshot.FileID || o.TicketID == old.OutputID {
		return o, errors.Join(err, errors.New("fresh ownership reuses an older identity"))
	}
	_, expected, err := freshJSONHash(o)
	if err != nil || !bytes.Equal(expected, entry.Manifest) {
		return o, errors.Join(err, errors.New("fresh ownership bytes are not the sealed record"))
	}
	return o, nil
}

func freshBoundID(id string) bool { return id != "" && len(id) <= 256 }

func validateFreshJournal(m restoreManifest, a FreshRestoreApproval, o freshOwnership) error {
	if err := validateRestoreManifest(m, a.QuarantineID); err != nil {
		return err
	}
	if m.Version != 2 || m.FreshID != a.OperationID || m.OperationID != a.OperationID || !bytes.Equal(m.RecordData, a.Inspection.RecordData) || !bytes.Equal(m.ParentPath, a.Inspection.Destination.ParentPath) || !bytes.Equal(m.Leaf, a.Inspection.Destination.Leaf) || m.ParentID != a.Inspection.Destination.ParentID || m.Source.ID != a.Inspection.Snapshot.FileID || m.Source.State != "ready" || m.StageID != o.StageID {
		return errors.New("fresh journal differs from its exact approval or ownership")
	}
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, a.QuarantineID)
	if err != nil {
		return err
	}
	if m.OutputID != "" && (m.OutputID == old.OutputID || m.OutputID == m.Record.SourceFileID || m.OutputID == o.TicketID) {
		return errors.New("fresh output must have a new independent identity")
	}
	return nil
}

// Ownership lookup has two exact possible locations, never a history scan. A
// different live operation cannot invalidate a proven older archived stage.
func openFreshOwnedOperation(root, approvalOp *restoreDirectory, o freshOwnership) (*restoreDirectory, bool, error) {
	var live, saved *restoreDirectory
	var liveID, savedID string
	journals, op, err := openRestoreOperation(root, o.QuarantineID)
	if err == nil {
		journals.Close()
		live = op
		liveID, err = admissionFileID(live.file)
		if err != nil {
			live.Close()
			return nil, false, err
		}
	} else if !os.IsNotExist(err) {
		retired, proofErr := retiredFreshArchive(root, approvalOp, o)
		if proofErr != nil || !retired {
			return nil, false, errors.Join(err, proofErr)
		}
	}
	saved, err = admissionOpenChild(approvalOp, "restore")
	if err == nil {
		savedID, err = admissionFileID(saved.file)
	}
	if err != nil && !os.IsNotExist(err) {
		if live != nil {
			live.Close()
		}
		if saved != nil {
			saved.Close()
		}
		return nil, false, err
	}
	if liveID == o.OperationDirectoryID && savedID != o.OperationDirectoryID {
		if saved != nil {
			saved.Close()
		}
		return live, false, nil
	}
	if savedID == o.OperationDirectoryID && liveID != o.OperationDirectoryID {
		if live != nil {
			live.Close()
		}
		return saved, true, nil
	}
	if live != nil {
		live.Close()
	}
	if saved != nil {
		saved.Close()
	}
	return nil, false, errors.New("fresh operation identity is absent or ambiguous")
}

// Invalid unrelated markers are not an exclusion. Only an agent-private stage,
// the same ticket file, protected root, exact approval and immutable ownership
// can establish the bounded address of the mutable operation journal.
func managedFreshRestoreStage(parent *restoreDirectory, parentID string) error {
	private, err := openAdmissionRoot(parent.path, false)
	if err != nil {
		return nil
	}
	defer private.Close()
	id, err := admissionFileID(private.file)
	if err != nil || id != parentID {
		return nil
	}
	ticketEntry, err := readFreshControl(private, freshTicketName, maxFreshOwnership)
	if err != nil {
		return nil
	}
	var ticket freshOwnerTicket
	if json.Unmarshal(ticketEntry.Manifest, &ticket) != nil || ticket.Version != 1 || ticket.FileID != ticketEntry.ManifestID || ticket.StageID != parentID || !ordinaryQuarantineID(ticket.QuarantineID) || !validReviewUUID(ticket.OperationID) || !freshBoundID(ticket.RootID) {
		return nil
	}
	rootPath := string(ticket.RootPath)
	if !filepath.IsAbs(rootPath) || filepath.Clean(rootPath) != rootPath || freshLocalDestination(filepath.Join(rootPath, "fresh-root-probe")) != nil {
		return nil
	}
	root, err := openAdmissionRoot(rootPath, false)
	if err != nil {
		return nil
	}
	defer root.Close()
	rootID, err := admissionFileID(root.file)
	if err != nil || rootID != ticket.RootID {
		return nil
	}
	queue, item, approvalOp, err := openFreshApproval(root, ticket.QuarantineID, ticket.OperationID)
	if err != nil {
		return nil
	}
	defer queue.Close()
	defer item.Close()
	defer approvalOp.Close()
	a, err := readFreshApproval(root, queue, item, approvalOp, ticket.QuarantineID, ticket.OperationID)
	if err != nil {
		return nil
	}
	approvalHash, _, err := freshJSONHash(a)
	if err != nil {
		return nil
	}
	o, entry, err := readFreshOwnership(approvalOp, a, approvalHash)
	if err != nil || entry.ManifestID != ticket.OwnershipID || entry.SHA256 != ticket.OwnershipSHA256 || o.StageID != parentID || o.TicketID != ticketEntry.ManifestID || !bytes.Equal(o.RootPath, ticket.RootPath) {
		return nil
	}
	fail := func(cause error) error { return errors.Join(ErrRestoreIntegrity, cause) }
	op, _, err := openFreshOwnedOperation(root, approvalOp, o)
	if err != nil {
		return fail(err)
	}
	defer op.Close()
	m, err := readRestoreManifest(op, a.QuarantineID)
	if err != nil {
		return fail(err)
	}
	if err = validateFreshJournal(m, a, o); err != nil {
		return fail(err)
	}
	if m.StageState != "present" && m.StageState != "deleting" {
		return fail(errors.New("retired fresh stage still has its exact ownership ticket"))
	}
	return ErrAlreadyQuarantined
}
