package quarantine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// RestorePendingError never promises rollback: a verified output may already
// have been published. Recovery uses its identity, not its current pathname or
// bytes, to record that historical event. The stored backup survives DB failure.
type RestorePendingError struct {
	QuarantineID, Destination string
	Err                       error
}

func (e *RestorePendingError) Error() string {
	return fmt.Sprintf("restore %s for %q needs recovery; a verified output may already be published and the quarantine record may not have been updated: %v", e.QuarantineID, e.Destination, e.Err)
}
func (e *RestorePendingError) Unwrap() error { return e.Err }

// Paths and the approved record use lossless byte encodings: JSON strings must
// not silently redirect a valid Unix filename containing non-UTF-8 bytes.
type restoreManifest struct {
	Version                     int
	OperationID, Phase          string
	FreshID                     string                 `json:"FreshID,omitempty"`
	Record                      cache.QuarantineRecord `json:"-"`
	RecordData                  []byte
	ParentPath, Leaf            []byte
	ParentID, StageID, OutputID string
	StageState                  string // planned, present, deleting, deleted
	Source                      purgeArtifact
}

func (m restoreManifest) stageName() string      { return ".edr-restore-" + m.OperationID }
func (s *Store) restoreFault(phase string) error { return s.purgeFault("restore-" + phase) }
func writeRestoreManifest(op *restoreDirectory, m restoreManifest) error {
	var err error
	m.RecordData, err = cache.EncodeQuarantinePurgeRecord(m.Record)
	if err != nil {
		return err
	}
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return writePrivateManifest(op, body)
}

func readRestoreManifest(op *restoreDirectory, id string) (restoreManifest, error) {
	var m restoreManifest
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxAdmissionJournal {
		return m, errors.Join(err, errors.New("invalid restore journal file"))
	}
	body, err := io.ReadAll(io.LimitReader(f, maxAdmissionJournal+1))
	if err != nil || len(body) > maxAdmissionJournal {
		return m, errors.Join(err, errors.New("restore journal exceeds limit"))
	}
	return decodeRestoreManifest(body, id)
}

func decodeRestoreManifest(body []byte, id string) (restoreManifest, error) {
	var m restoreManifest
	if len(body) > maxAdmissionJournal {
		return m, errors.New("restore journal exceeds limit")
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return m, err
	}
	var err error
	m.Record, err = cache.DecodeQuarantinePurgeRecord(m.RecordData)
	if err != nil {
		return m, err
	}
	return m, validateRestoreManifest(m, id)
}

func validateRestoreManifest(m restoreManifest, id string) error {
	nonce, err := uuid.Parse(m.OperationID)
	fresh := m.Version == 2 && validReviewUUID(m.FreshID) && m.FreshID == m.OperationID
	if err != nil || nonce.String() != m.OperationID || !(m.Version == 1 && m.FreshID == "" || fresh) || !ordinaryQuarantineID(id) || m.Record.QuarantineID != id || m.Record.Restored || m.Record.Purged || !m.Record.Restorable || m.ParentID == "" || m.Source.ID == "" {
		return errors.New("invalid restore approval identity")
	}
	if _, err = restoreDigest(m.Record.SHA256); err != nil {
		return err
	}
	if m.OutputID != "" && m.OutputID == m.Source.ID {
		return errors.New("restore requires an independent output identity")
	}
	parent, leaf := string(m.ParentPath), string(m.Leaf)
	if !filepath.IsAbs(parent) || filepath.Clean(parent) != parent || leaf == "" {
		return errors.New("invalid restore destination")
	}
	if got, err := restoreLeaf(filepath.Join(parent, leaf)); err != nil || got != leaf {
		return errors.New("invalid restore destination leaf")
	}
	switch m.Source.State {
	case "ready", "claiming", "claimed", "deleted":
	default:
		return errors.New("invalid restore source state")
	}
	if fresh && m.Source.State != "ready" {
		return errors.New("fresh restore must retain the original snapshot")
	}
	switch m.StageState {
	case "planned":
		if m.StageID != "" || m.OutputID != "" {
			return errors.New("uncreated restore stage has identities")
		}
	case "present", "deleting", "deleted":
		if m.StageID == "" {
			return errors.New("missing restore stage identity")
		}
	default:
		return errors.New("invalid restore stage state")
	}
	switch m.Phase {
	case "prepared", "copying", "aborted":
		if m.Source.State != "ready" {
			return errors.New("unpublished restore retired its backup")
		}
	case "verified", "publishing", "published":
		if m.OutputID == "" || m.StageID == "" || m.StageState != "present" || m.Source.State != "ready" {
			return errors.New("restore lacks verified output proof")
		}
	case "committed":
		if m.OutputID == "" || m.StageID == "" {
			return errors.New("committed restore lacks output proof")
		}
	default:
		return errors.New("invalid restore phase")
	}
	return nil
}

func openRestoreOperation(root *restoreDirectory, id string) (*restoreDirectory, *restoreDirectory, error) {
	journals, err := admissionOpenChild(root, ".restores")
	if err != nil {
		return nil, nil, err
	}
	op, err := admissionOpenChild(journals, id)
	if err != nil {
		journals.Close()
		return nil, nil, err
	}
	return journals, op, nil
}

func checkNoRestoreOperation(root *restoreDirectory, id string) error {
	journals, op, err := openRestoreOperation(root, id)
	if err == nil {
		op.Close()
		journals.Close()
		return &RestorePendingError{QuarantineID: id, Err: errors.New("finish restore recovery before purge")}
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Store) restoreJournaled(root *restoreDirectory, id string) error {
	journals, op, err := openRestoreOperation(root, id)
	if err == nil {
		defer journals.Close()
		defer op.Close()
		m, e := readRestoreManifest(op, id)
		if e == nil {
			e = s.resumeRestore(root, journals, op, &m)
		}
		if e != nil {
			return &RestorePendingError{id, m.Record.OriginalPath, e}
		}
		if m.Phase == "aborted" {
			return s.restoreJournaled(root, id)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil {
		return err
	}
	if !found || rec.Restored || rec.Purged || !rec.Restorable {
		return fmt.Errorf("quarantine id %s is missing, already restored, or not restorable", id)
	}
	expected, err := restoreDigest(rec.SHA256)
	if err != nil {
		return err
	}
	leaf, err := restoreLeaf(rec.OriginalPath)
	if err != nil {
		return err
	}
	in, err := restoreOpenInput(root, id+".quarantined")
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := restoreInputInfo(in)
	if err != nil {
		return err
	}
	sourceID, err := admissionFileID(in)
	if err != nil {
		return err
	}
	parent, err := openRestoreDirectory(filepath.Dir(rec.OriginalPath), true)
	if err != nil {
		return err
	}
	defer parent.Close()
	parentID, err := admissionFileID(parent.file)
	if err != nil {
		return err
	}
	journals, err = admissionChild(root, ".restores")
	if err != nil {
		return err
	}
	defer journals.Close()
	entries, err := journals.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) >= maxAdmissionRecovery {
		return errors.Join(err, errors.New("restore recovery queue is full"))
	}
	op, err = admissionMakeDirectory(journals, id)
	if err != nil {
		return err
	}
	defer op.Close()
	m := restoreManifest{Version: 1, OperationID: uuid.NewString(), Phase: "prepared", Record: rec, ParentPath: []byte(parent.path), Leaf: []byte(leaf), ParentID: parentID, StageState: "planned", Source: purgeArtifact{ID: sourceID, State: "ready"}}
	pending := func(e error) error { return &RestorePendingError{id, rec.OriginalPath, e} }
	if err = admissionSyncDirectory(journals); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("directory-created"); err != nil {
		return pending(err)
	}
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("intent"); err != nil {
		return pending(err)
	}
	stage, err := admissionMakeDirectory(parent, m.stageName())
	if err != nil {
		return pending(err)
	}
	defer stage.Close()
	if err = admissionSyncDirectory(parent); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("stage-created-fs"); err != nil {
		return pending(err)
	}
	m.StageID, err = admissionFileID(stage.file)
	if err != nil {
		return pending(err)
	}
	m.StageState, m.Phase = "present", "copying"
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("stage-index-intent"); err != nil {
		return pending(err)
	}
	if err = s.cache.TrackQuarantineRestoreStage(s.restoreStageLocator(m)); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("stage-indexed"); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("stage-created"); err != nil {
		return pending(err)
	}
	out, err := restoreCreateOutput(stage)
	if err != nil {
		return pending(err)
	}
	defer out.Close()
	m.OutputID, err = admissionFileID(out)
	if err != nil {
		return pending(err)
	}
	if err = admissionSyncDirectory(stage); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("output-created-fs"); err != nil {
		return pending(err)
	}
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("output-created"); err != nil {
		return pending(err)
	}
	abort := func(cause error) error {
		_ = out.Close()
		_ = stage.Close()
		m.Phase = "aborted"
		if e := writeRestoreManifest(op, m); e != nil {
			return pending(errors.Join(cause, e))
		}
		if e := s.cleanupRestoreStage(parent, op, &m); e != nil {
			return pending(errors.Join(cause, e))
		}
		if e := cleanupPurgeOperation(journals, op, id, nil); e != nil {
			return pending(errors.Join(cause, e))
		}
		return cause
	}
	if err = s.copyRestoreSnapshot(out, in, info, expected); err != nil {
		return abort(err)
	}
	if err = restoreFinishOutput(out, info, rec.OriginalSecurity); err != nil {
		return abort(err)
	}
	if err = out.Sync(); err != nil {
		return abort(err)
	}
	m.Phase = "verified"
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("verified"); err != nil {
		return pending(err)
	}
	if s.restoreBeforePublish != nil {
		if err = s.restoreBeforePublish(); err != nil {
			return abort(err)
		}
	}
	if err = parent.stillNamed(); err != nil {
		return abort(err)
	}
	if _, err = s.cache.ValidateQuarantineRestore(rec, m.OperationID, m.ParentID, m.OutputID); err != nil {
		return abort(err)
	}
	m.Phase = "publishing"
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("publication-intent"); err != nil {
		return pending(err)
	}
	if err = restorePublishOutput(out, stage, parent, leaf); err != nil {
		return abort(err)
	}
	if err = s.restoreFault("published-fs"); err != nil {
		return pending(err)
	}
	if err = admissionSyncDirectory(parent); err != nil {
		return pending(err)
	}
	m.Phase = "published"
	if err = writeRestoreManifest(op, m); err != nil {
		return pending(err)
	}
	if err = errors.Join(out.Close(), stage.Close(), in.Close()); err != nil {
		return pending(err)
	}
	if err = s.restoreFault("published"); err != nil {
		return pending(err)
	}
	if s.restoreAfterPublish != nil {
		if err = s.restoreAfterPublish(); err != nil {
			return pending(err)
		}
	}
	if err = s.finishRestore(root, journals, op, &m); err != nil {
		return pending(err)
	}
	return nil
}

func openRestoreParent(m restoreManifest) (*restoreDirectory, error) {
	parent, err := platformRestoreDirectory(string(m.ParentPath))
	if err != nil {
		return nil, err
	}
	id, err := admissionFileID(parent.file)
	if err != nil || id != m.ParentID {
		parent.Close()
		return nil, errors.Join(err, errors.New("restore parent identity changed"))
	}
	return parent, nil
}

func (s *Store) resumeRestore(root, journals, op *restoreDirectory, m *restoreManifest) error {
	if err := validateRestoreManifest(*m, m.Record.QuarantineID); err != nil {
		return err
	}
	if m.Version != 1 {
		// Syntax is never execution authority. Ordinary recovery cannot use
		// a fresh journal to cross its hold or retire the retained snapshot.
		return ErrFreshRestoreApprovalRequired
	}
	// Older executables recorded the stage identity without a lookup row. Import
	// only a locator from their validated private journal, before any recovery
	// refusal can leave a retained snapshot exposed to another scan. Admission
	// still corroborates this journal and the live private directory. Deleting
	// stages are already empty; do not recreate their retired lookup.
	if m.StageState == "present" {
		if err := s.cache.TrackQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
			return fmt.Errorf("restore stage lookup recovery failed: %w", err)
		}
	}
	if err := checkNoRecoveryHold(root, m.Record.QuarantineID); err != nil {
		return err
	}
	if other, pending, e := openPurgeOperation(root, m.Record.QuarantineID); e == nil {
		pending.Close()
		other.Close()
		return errors.New("conflicting purge and restore journals retained")
	} else if !os.IsNotExist(e) {
		return e
	}
	if err := validateRestoreOperation(journals, op, *m); err != nil {
		return err
	}
	switch m.Phase {
	case "prepared", "copying", "verified", "aborted":
		// No publication was attempted. Retain the original quarantine snapshot
		// and discard only an output whose recorded identity is still owned.
		parent, err := openRestoreParent(*m)
		if err != nil {
			return err
		}
		defer parent.Close()
		m.Phase = "aborted"
		if err = writeRestoreManifest(op, *m); err != nil {
			return err
		}
		if err = s.cleanupRestoreStage(parent, op, m); err != nil {
			return err
		}
		return cleanupPurgeOperation(journals, op, m.Record.QuarantineID, nil)
	case "publishing":
		parent, err := openRestoreParent(*m)
		if err != nil {
			return err
		}
		defer parent.Close()
		output, err := inspectPurgeArtifact(parent, string(m.Leaf))
		if err != nil || output.State == "absent" || output.ID != m.OutputID {
			return errors.Join(err, errors.New("publication was interrupted and its output identity cannot be proven; destination preserved"))
		}
		if err = admissionSyncDirectory(parent); err != nil {
			return err
		}
		m.Phase = "published"
		if err = writeRestoreManifest(op, *m); err != nil {
			return err
		}
	}
	return s.finishRestore(root, journals, op, m)
}

func (s *Store) finishRestore(root, journals, op *restoreDirectory, m *restoreManifest) error {
	if m.Version != 1 {
		return ErrFreshRestoreApprovalRequired
	}
	if err := validateRestoreOperation(journals, op, *m); err != nil {
		return err
	}
	committed, err := s.cache.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID)
	if err != nil {
		return fmt.Errorf("quarantine record was not updated: %w", err)
	}
	if !committed {
		if m.Phase != "published" {
			return errors.New("restore lacks durable publication proof")
		}
		if err = s.restoreFault("before-database-commit"); err != nil {
			return err
		}
		if err = s.cache.CommitQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID, time.Now().UTC()); err != nil {
			return fmt.Errorf("quarantine record was not updated: %w", err)
		}
		if err = s.restoreFault("database-committed"); err != nil {
			return err
		}
	}
	m.Phase = "committed"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = s.restoreFault("committed"); err != nil {
		return err
	}
	// The historical restore remains recorded if its output is subsequently
	// edited, moved or replaced. Cleanup never opens that destination file.
	if m.StageState != "deleted" {
		parent, e := openRestoreParent(*m)
		if e != nil {
			return e
		}
		err = s.cleanupRestoreStage(parent, op, m)
		parent.Close()
		if err != nil {
			return err
		}
	}
	check := func() error {
		ok, e := s.cache.ValidateQuarantineRestore(m.Record, m.OperationID, m.ParentID, m.OutputID)
		if e == nil && !ok {
			e = errors.New("restore receipt missing before retiring backup")
		}
		return e
	}
	save := func() error { return writeRestoreManifest(op, *m) }
	if err = retireArtifactStep(root, op, m.Record.QuarantineID+".quarantined", "snapshot", "source", &m.Source, check, save, s.restoreFault); err != nil {
		return fmt.Errorf("%w: %v", ErrRestoreIntegrity, err)
	}
	return cleanupPurgeOperation(journals, op, m.Record.QuarantineID, s.restoreFault)
}

func validateRestoreOperation(journals, op *restoreDirectory, m restoreManifest) error {
	fresh, err := admissionOpenChild(journals, m.Record.QuarantineID)
	if err != nil {
		return err
	}
	defer fresh.Close()
	before, e1 := admissionFileID(op.file)
	after, e2 := admissionFileID(fresh.file)
	if e1 != nil || e2 != nil || before != after {
		return errors.Join(e1, e2, errors.New("restore journal directory identity changed"))
	}
	entries, err := fresh.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > maxAdmissionRecovery {
		return errors.Join(err, errors.New("restore journal directory exceeds limit"))
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return errors.New("unexpected restore journal entry type retained")
		}
		if entry.Name() == "snapshot" {
			if m.Phase != "committed" || m.Source.State != "claiming" && m.Source.State != "claimed" {
				return errors.New("unexpected retired restore backup retained")
			}
			item, e := inspectPurgeArtifact(op, "snapshot")
			if e != nil || item.ID != m.Source.ID {
				return errors.Join(e, errors.New("retired restore backup identity changed"))
			}
		} else if entry.Name() != "manifest.json" && !purgeJournalTemporary(entry.Name()) {
			return errors.New("unknown restore journal entry retained")
		}
	}
	return nil
}

// Only the private stage may be cleaned. A replacement stage or output is
// retained for attention; rmdir is bounded to an empty directory, never recursive.
func (s *Store) cleanupRestoreStage(parent, op *restoreDirectory, m *restoreManifest) error {
	if m.StageState == "deleted" {
		return nil
	}
	stage, err := admissionOpenChild(parent, m.stageName())
	if os.IsNotExist(err) {
		if m.StageState != "planned" && m.StageState != "deleting" {
			return errors.New("restore stage disappeared before cleanup intent")
		}
		if m.StageID != "" {
			if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
				return err
			}
		}
		m.StageState = "deleted"
		// A never-created stage has no identity and remains planned in the
		// final manifest until manifest-last cleanup removes the operation.
		if m.StageID == "" {
			m.StageState = "planned"
		}
		return writeRestoreManifest(op, *m)
	}
	if err != nil {
		return err
	}
	defer stage.Close()
	id, err := admissionFileID(stage.file)
	if err != nil || m.StageID == "" || id != m.StageID {
		return errors.Join(err, errors.New("restore stage identity unavailable or changed; retained"))
	}
	entries, err := stage.file.ReadDir(2)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > 1 {
		return errors.Join(err, errors.New("unexpected restore stage entries retained"))
	}
	if len(entries) == 1 {
		if entries[0].Name() != "snapshot" || !entries[0].Type().IsRegular() || m.OutputID == "" {
			return errors.New("unidentified restore output retained")
		}
		item, e := purgeOpenFile(stage, "snapshot")
		if e != nil {
			return e
		}
		if item.id != m.OutputID {
			item.Close()
			return errors.New("restore output identity changed; retained")
		}
		if e = purgeRemoveFile(item, stage, "snapshot"); e != nil {
			item.Close()
			return e
		}
		if e = item.Close(); e != nil {
			return e
		}
		if e = requirePurgeNameAbsent(stage, "snapshot"); e != nil {
			return e
		}
	}
	if err = admissionSyncDirectory(stage); err != nil {
		return err
	}
	m.StageState = "deleting"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	if err = s.restoreFault("stage-delete-intent"); err != nil {
		return err
	}
	// Remove the lookup before unlinking the empty directory, so its inode
	// cannot be reused while a stale lookup still names this operation.
	if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(*m)); err != nil {
		return err
	}
	if err = s.restoreFault("stage-index-deleted"); err != nil {
		return err
	}
	if err = stage.Close(); err != nil {
		return err
	}
	if err = admissionRemoveDirectory(parent, m.stageName()); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err = admissionSyncDirectory(parent); err != nil {
		return err
	}
	if err = s.restoreFault("stage-removed-fs"); err != nil {
		return err
	}
	m.StageState = "deleted"
	if err = writeRestoreManifest(op, *m); err != nil {
		return err
	}
	return s.restoreFault("stage-deleted")
}
