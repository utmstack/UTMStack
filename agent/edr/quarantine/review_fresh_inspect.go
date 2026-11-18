package quarantine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FreshRestoreInspection describes a new proposed publication after cancellation.
// It is evidence only: obtaining it neither releases the fence nor authorizes a
// restore. Paths use byte encodings so Unix filenames survive saved JSON intact.
type FreshRestoreInspection struct {
	Version      int               `json:"version"`
	Kind         string            `json:"kind"`
	QuarantineID string            `json:"quarantine_id"`
	ObservedAt   time.Time         `json:"observed_at"`
	RootID       string            `json:"root_id"`
	Cancellation JournalInspection `json:"cancellation"`
	Hold         JournalInspection `json:"hold"`
	Original     JournalInspection `json:"original"`
	RecordData   []byte            `json:"record_data"`
	Snapshot     FreshRestoreInput `json:"snapshot"`
	Destination  FreshRestorePath  `json:"destination"`
}

type FreshRestoreInput struct {
	FileID string `json:"file_id"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type FreshRestorePath struct {
	ParentPath []byte `json:"parent_path"`
	ParentID   string `json:"parent_id"`
	Leaf       []byte `json:"leaf"`
}

// InspectFreshRestore reads existing private storage without SQLite, locks,
// hardening, replay or creation. The cancellation must already have archived the
// same original operation. An independent snapshot and an absent, explicit local
// destination are required; the old output never supplies publication evidence.
func InspectFreshRestore(dir, id, holdID, destination string) (FreshRestoreInspection, error) {
	r := FreshRestoreInspection{Version: 1, Kind: "fresh-restore-inspection", QuarantineID: id, ObservedAt: time.Now().UTC()}
	if !ordinaryQuarantineID(id) || !validReviewUUID(holdID) {
		return r, errors.New("fresh inspection requires an exact quarantine ID and hold UUID")
	}
	if err := freshLocalDestination(destination); err != nil {
		return r, err
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return r, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return r, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return r, err
	}
	defer root.Close()
	return inspectFreshRestore(root, id, holdID, destination, r)
}

func inspectFreshRestore(root *restoreDirectory, id, holdID, destination string, r FreshRestoreInspection) (FreshRestoreInspection, error) {
	var err error
	if r.RootID, err = admissionFileID(root.file); err != nil {
		return r, err
	}
	if r.Hold, err = inspectJournal(root, ".purges", id); err != nil {
		return r, err
	}
	r.Hold.Operation = "purge"
	// Decode only the exact regular fence. Its receipt subsequently validates
	// these bytes, including the original journal, against the cancellation.
	var hold RecoveryHold
	if r.Hold.Status != "captured" || r.Hold.EntryKind != "file" {
		return r, errors.New("fresh inspection requires an active retention fence")
	}
	if err = json.Unmarshal(r.Hold.Manifest, &hold); err != nil || hold.DecisionID != holdID {
		return r, errors.Join(err, errors.New("fresh inspection hold UUID changed"))
	}
	a := cancellationApproval(hold, r.Hold)
	if err = validRecoveryReleaseApproval(a); err != nil {
		return r, err
	}
	op, m, err := openCancellationProof(root, id, a)
	if err != nil {
		return r, err
	}
	defer op.Close()
	archived, err := cancellationOriginal(root, op, m)
	if err != nil || !archived {
		return r, errors.Join(err, errors.New("finish cancellation before inspecting a fresh restore"))
	}
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return r, err
	}
	r.Cancellation, err = inspectJournalFile(JournalInspection{Operation: "cancellation", DirectoryID: m.DirectoryID}, f, maxRecoveryRelease)
	err = errors.Join(err, f.Close())
	if err != nil {
		return r, err
	}
	expected, err := json.Marshal(m)
	if err != nil || r.Cancellation.ManifestID != m.FileID || !bytes.Equal(r.Cancellation.Manifest, expected) {
		return r, errors.Join(err, errors.New("cancellation receipt changed during fresh inspection"))
	}
	r.Original = hold.Original
	original, err := decodeRestoreManifest(r.Original.Manifest, id)
	if err != nil || original.Source.State != "ready" {
		return r, errors.Join(err, errors.New("fresh restore requires a validated retained snapshot record"))
	}
	r.RecordData = append([]byte(nil), original.RecordData...)
	input, err := restoreOpenInput(root, id+".quarantined")
	if err != nil {
		return r, err
	}
	defer input.Close()
	r.Snapshot, err = inspectFreshInput(input)
	if err != nil {
		return r, err
	}
	if r.Snapshot.FileID != original.Source.ID || r.Snapshot.SHA256 != original.Record.SHA256 || r.Snapshot.FileID == original.Record.SourceFileID || r.Snapshot.FileID == original.OutputID {
		return r, errors.New("fresh snapshot identity or contents do not match the independent retained backup")
	}
	parent, err := platformRestoreDirectory(filepath.Dir(destination))
	if err != nil {
		return r, err
	}
	defer parent.Close()
	if err = rejectFreshManagedDestination(parent, r.RootID, original.StageID); err != nil {
		return r, err
	}
	r.Destination = FreshRestorePath{ParentPath: []byte(parent.path), Leaf: []byte(filepath.Base(destination))}
	if r.Destination.ParentID, err = admissionFileID(parent.file); err != nil {
		return r, err
	}
	if err = requirePurgeNameAbsent(parent, string(r.Destination.Leaf)); err != nil {
		return r, errors.Join(err, errors.New("fresh destination must be absent; existing objects are retained"))
	}
	if err = parent.stillNamed(); err != nil {
		return r, err
	}
	// Close the observation with the same cancellation/fence/original checks.
	// An inspection is still not a lease; execution must revalidate all pins.
	if err = cancellationFence(root, m); err == nil {
		_, err = cancellationOriginal(root, op, m)
	}
	return r, err
}

func inspectFreshInput(f *os.File) (FreshRestoreInput, error) {
	var pin FreshRestoreInput
	before, err := restoreInputInfo(f)
	if err != nil {
		return pin, err
	}
	if pin.FileID, err = admissionFileID(f); err != nil {
		return pin, err
	}
	h := sha256.New()
	n, err := io.Copy(h, io.NewSectionReader(f, 0, before.Size()))
	if err != nil {
		return pin, err
	}
	after, err := f.Stat()
	if err != nil || n != before.Size() || !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return pin, errors.Join(err, errors.New("retained snapshot changed during fresh inspection"))
	}
	pin.Size, pin.SHA256 = n, hex.EncodeToString(h.Sum(nil))
	return pin, nil
}

func freshLocalDestination(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.ContainsRune(path, 0) {
		return errors.New("fresh destination must be an ordinary clean absolute local path")
	}
	if filepath.Separator == '\\' {
		volume := filepath.VolumeName(path)
		if len(volume) != 2 || volume[1] != ':' || !(volume[0] >= 'A' && volume[0] <= 'Z' || volume[0] >= 'a' && volume[0] <= 'z') {
			return errors.New("fresh destination requires a local drive path")
		}
	}
	_, err := restoreLeaf(path)
	return err
}

// Walk exact ancestor identities, not pathname prefixes: renamed original
// stages and storage aliases must not become a proposed publication directory.
func rejectFreshManagedDestination(parent *restoreDirectory, rootID, oldStageID string) error {
	for path := parent.path; ; path = filepath.Dir(path) {
		d, err := platformRestoreDirectory(path)
		if err != nil {
			return err
		}
		id, err := admissionFileID(d.file)
		if err != nil {
			d.Close()
			return err
		}
		// A separately cancelled fresh stage remains retained evidence too.
		// Follow its exact private ticket address instead of scanning history
		// or trusting a stage-like name or an SQL locator.
		managed := managedFreshRestoreStage(d, id)
		closeErr := d.Close()
		if managed != nil || closeErr != nil {
			return errors.Join(managed, closeErr, errors.New("fresh destination is inside a retained fresh restore stage"))
		}
		if id == rootID || oldStageID != "" && id == oldStageID {
			return errors.New("fresh destination is inside retained quarantine storage or the original restore stage")
		}
		if filepath.Dir(path) == path {
			return nil
		}
	}
}
