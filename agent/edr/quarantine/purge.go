package quarantine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// PurgePendingError means an approved purge has a durable recovery record.
// Some artifacts may already be removed; a caller must not report rollback.
type PurgePendingError struct {
	QuarantineID string
	Err          error
}

func (e *PurgePendingError) Error() string {
	return fmt.Sprintf("purge %s needs recovery; some stored artifacts may already be removed: %v", e.QuarantineID, e.Err)
}
func (e *PurgePendingError) Unwrap() error { return e.Err }

type purgeArtifact struct {
	ID    string
	State string // absent, ready, claiming, claimed, deleted
}

type purgeFile struct {
	file *os.File
	id   string
}

func (p *purgeFile) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

func purgeFileFromDescriptor(f *os.File) (*purgeFile, error) {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, errors.Join(err, errors.New("purge requires a regular stored artifact"))
	}
	id, err := admissionFileID(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &purgeFile{file: f, id: id}, nil
}

type purgeManifest struct {
	Version      int
	OperationID  string
	Phase        string
	ApprovedAt   time.Time
	Record       cache.QuarantineRecord `json:"-"`
	RecordData   []byte
	Stored       purgeArtifact
	Raw          purgeArtifact
	CaptureState string // absent, present, deleted
}

func (s *Store) purgeFault(phase string) error {
	if s.responseFault != nil {
		return s.responseFault(phase)
	}
	return nil
}

func writePurgeManifest(op *restoreDirectory, m purgeManifest) error {
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

func readPurgeManifest(op *restoreDirectory, id string) (purgeManifest, error) {
	var m purgeManifest
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxAdmissionJournal {
		return m, errors.New("invalid purge journal file")
	}
	b, err := io.ReadAll(io.LimitReader(f, maxAdmissionJournal+1))
	if err != nil || len(b) > maxAdmissionJournal {
		return m, errors.Join(err, errors.New("purge journal exceeds its limit"))
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	if m.Record, err = cache.DecodeQuarantinePurgeRecord(m.RecordData); err != nil {
		return m, err
	}
	nonce, nonceErr := uuid.Parse(m.OperationID)
	if m.Version != 1 || m.Record.QuarantineID != id || !ordinaryQuarantineID(id) || m.Record.Purged || m.ApprovedAt.IsZero() || nonceErr != nil || nonce == uuid.Nil || nonce.String() != m.OperationID {
		return m, errors.New("invalid purge journal identity")
	}
	for _, artifact := range []purgeArtifact{m.Stored, m.Raw} {
		if artifact.State == "absent" {
			if artifact.ID != "" {
				return m, errors.New("absent purge artifact has an identity")
			}
			continue
		}
		switch artifact.State {
		case "ready", "claiming", "claimed", "deleted":
			if artifact.ID == "" || len(artifact.ID) > 256 {
				return m, errors.New("missing purge artifact identity")
			}
		default:
			return m, errors.New("invalid purge artifact state")
		}
	}
	if err = validPurgeCapture(m.Record); err != nil {
		return m, err
	}
	if m.Raw.State != "absent" && m.Raw.ID != m.Record.SourceFileID {
		return m, errors.New("raw purge identity differs from admission")
	}
	switch m.CaptureState {
	case "present", "deleted":
		if m.Record.SourceCaptureDir == "" {
			return m, errors.New("purge capture location is missing")
		}
		if m.CaptureState == "deleted" && m.Raw.State != "absent" && m.Raw.State != "deleted" {
			return m, errors.New("deleted capture still has an unfinished raw artifact")
		}
	case "absent":
		if m.Raw.State != "absent" {
			return m, errors.New("absent capture has a raw purge artifact")
		}
	default:
		return m, errors.New("invalid purge capture state")
	}
	if m.Phase != "prepared" && m.Phase != "deleting" && m.Phase != "committed" || m.Phase == "committed" && !purgeArtifactsRemoved(m) {
		return m, errors.New("invalid purge journal phase")
	}
	return m, nil
}

func validPurgeCapture(rec cache.QuarantineRecord) error {
	if rec.SourceCaptureDir == "" {
		if rec.SourceCaptureDirID != "" || rec.SourceFileID != "" {
			return errors.New("retained source identities have no directory")
		}
		return nil
	}
	if !filepath.IsAbs(rec.SourceCaptureDir) || filepath.Clean(rec.SourceCaptureDir) != rec.SourceCaptureDir || filepath.Base(rec.SourceCaptureDir) != ".edr-quarantine-"+rec.QuarantineID || rec.SourceCaptureDirID == "" || rec.SourceFileID == "" {
		return errors.New("invalid retained source identity")
	}
	return nil
}

func purgeArtifactsRemoved(m purgeManifest) bool {
	done := func(a purgeArtifact) bool { return a.State == "absent" || a.State == "deleted" }
	return done(m.Stored) && done(m.Raw) && (m.CaptureState == "absent" || m.CaptureState == "deleted")
}

func inspectPurgeArtifact(d *restoreDirectory, name string) (purgeArtifact, error) {
	f, err := purgeOpenFile(d, name)
	if os.IsNotExist(err) {
		return purgeArtifact{State: "absent"}, nil
	}
	if err != nil {
		return purgeArtifact{}, err
	}
	defer f.Close()
	return purgeArtifact{ID: f.id, State: "ready"}, nil
}

func (s *Store) preparePurge(root, journals *restoreDirectory, rec cache.QuarantineRecord) (*restoreDirectory, purgeManifest, error) {
	m := purgeManifest{Version: 1, OperationID: uuid.NewString(), Phase: "prepared", ApprovedAt: time.Now().UTC(), Record: rec, Raw: purgeArtifact{State: "absent"}, CaptureState: "absent"}
	if err := validPurgeCapture(rec); err != nil {
		return nil, m, err
	}
	var err error
	if m.Stored, err = inspectPurgeArtifact(root, rec.QuarantineID+".quarantined"); err != nil {
		return nil, m, err
	}
	if rec.SourceCaptureDir != "" {
		d, e := openAdmissionRoot(rec.SourceCaptureDir, false)
		if e != nil && !os.IsNotExist(e) {
			return nil, m, e
		}
		if e == nil {
			id, idErr := admissionFileID(d.file)
			if idErr != nil || id != rec.SourceCaptureDirID {
				d.Close()
				return nil, m, errors.Join(idErr, errors.New("retained source directory identity changed"))
			}
			m.Raw, err = inspectPurgeArtifact(d, "source")
			if err == nil {
				entries, listErr := d.file.ReadDir(2)
				if listErr != nil && !errors.Is(listErr, io.EOF) {
					err = listErr
				}
				if len(entries) > 1 || len(entries) == 1 && (entries[0].Name() != "source" || !entries[0].Type().IsRegular()) {
					err = errors.New("unexpected retained source entries preserved")
				}
			}
			d.Close()
			if err != nil || m.Raw.State != "absent" && m.Raw.ID != rec.SourceFileID {
				return nil, m, errors.Join(err, errors.New("retained source identity changed"))
			}
			m.CaptureState = "present"
		}
	}
	op, err := admissionMakeDirectory(journals, rec.QuarantineID)
	if err != nil {
		return nil, m, err
	}
	if err = admissionSyncDirectory(journals); err == nil {
		err = s.purgeFault("purge-directory-created")
	}
	if err == nil {
		err = writePurgeManifest(op, m)
	}
	if err == nil {
		err = s.purgeFault("purge-intent")
	}
	return op, m, err
}

func openPurgeOperation(root *restoreDirectory, id string) (*restoreDirectory, *restoreDirectory, error) {
	journals, err := admissionOpenChild(root, ".purges")
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

func (s *Store) purgeJournaled(root *restoreDirectory, id string) error {
	journals, op, err := openPurgeOperation(root, id)
	var m purgeManifest
	if os.IsNotExist(err) {
		rec, found, e := s.cache.GetQuarantine(id)
		if e != nil {
			return e
		}
		if !found || rec.Purged {
			return fmt.Errorf("quarantine id %s is missing or already purged", id)
		}
		journals, err = admissionChild(root, ".purges")
		if err != nil {
			return err
		}
		entries, e := journals.file.ReadDir(maxAdmissionRecovery + 1)
		if e != nil && !errors.Is(e, io.EOF) || len(entries) >= maxAdmissionRecovery {
			journals.Close()
			return errors.Join(e, errors.New("purge recovery queue is full"))
		}
		op, m, err = s.preparePurge(root, journals, rec)
	} else if err == nil {
		m, err = readPurgeManifest(op, id)
	}
	if journals != nil {
		defer journals.Close()
	}
	if op != nil {
		defer op.Close()
	}
	if err == nil {
		err = s.resumePurge(root, journals, op, &m)
	}
	if err != nil && op != nil {
		return &PurgePendingError{QuarantineID: id, Err: err}
	}
	return err
}

// Claim first, persist that fact, then unlink only inside the private retired
// namespace. Once claimed, recovery never deletes a newly occupied source name.
func (s *Store) purgeArtifactStep(op, from, retired *restoreDirectory, source, target, label string, artifact *purgeArtifact, m *purgeManifest) error {
	check := func() error {
		if committed, err := s.cache.ValidateQuarantinePurge(m.Record, m.OperationID); err != nil {
			return err
		} else if committed {
			return errors.New("purge committed before its pending destructive step")
		}
		return nil
	}
	return retireArtifactStep(from, retired, source, target, label, artifact, check, func() error { return writePurgeManifest(op, *m) }, s.purgeFault)
}

// retireArtifactStep is shared by purge and post-receipt restore cleanup. The
// caller supplies its durable approval guard and journal; each claimed identity
// is verified before unlink, and a claimed source name is never reopened.
func retireArtifactStep(from, retired *restoreDirectory, source, target, label string, artifact *purgeArtifact, check, save func() error, fault func(string) error) error {
	if artifact.State == "absent" || artifact.State == "deleted" {
		return nil
	}
	if err := check(); err != nil {
		return err
	}
	if artifact.State == "ready" || artifact.State == "claiming" {
		claimed, err := inspectPurgeArtifact(retired, target)
		if err != nil {
			return err
		}
		if claimed.State == "absent" {
			input, e := purgeOpenFile(from, source)
			if e != nil {
				return fmt.Errorf("purge source disappeared before a durable claim: %w", e)
			}
			defer input.Close()
			if input.id != artifact.ID {
				return errors.New("purge source identity changed")
			}
			artifact.State = "claiming"
			if e = save(); e != nil {
				return e
			}
			if e = fault(label + "-claim-intent"); e != nil {
				return e
			}
			if e = admissionClaim(input.file, from, source, retired, target); e != nil {
				return e
			}
			claimed, e = inspectPurgeArtifact(retired, target)
			if e != nil || claimed.ID != artifact.ID {
				rollback := admissionRollback(input.file, retired, target, from, source)
				return errors.Join(e, rollback, errors.New("purge captured a replacement; it was returned or retained without deletion"))
			}
			if e = errors.Join(admissionSyncDirectory(from), admissionSyncDirectory(retired)); e != nil {
				return e
			}
			if e = fault(label + "-claimed-fs"); e != nil {
				return e
			}
			if e = input.Close(); e != nil {
				return e
			}
		} else if claimed.ID != artifact.ID {
			return errors.New("retired purge artifact identity changed")
		}
		artifact.State = "claimed"
		if err = save(); err != nil {
			return err
		}
		if err = fault(label + "-claimed"); err != nil {
			return err
		}
	}
	f, err := purgeOpenFile(retired, target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		if f.id != artifact.ID {
			f.Close()
			return errors.New("retired purge artifact identity changed")
		}
		if e := fault(label + "-delete-intent"); e != nil {
			f.Close()
			return e
		}
		if e := purgeRemoveFile(f, retired, target); e != nil {
			f.Close()
			return e
		}
		if e := f.Close(); e != nil {
			return e
		}
	}
	// Windows can retain a delete-pending name while another handle is open.
	// Do not record absence until our handles close and the name is absent.
	if err = requirePurgeNameAbsent(retired, target); err != nil {
		return err
	}
	if err = admissionSyncDirectory(retired); err != nil {
		return err
	}
	if err = fault(label + "-unlinked"); err != nil {
		return err
	}
	artifact.State = "deleted"
	if err = save(); err != nil {
		return err
	}
	return fault(label + "-deleted")
}

func (s *Store) purgeCapture(op *restoreDirectory, m *purgeManifest) error {
	if m.CaptureState != "present" {
		return nil
	}
	d, err := openAdmissionRoot(m.Record.SourceCaptureDir, false)
	if os.IsNotExist(err) && (m.Raw.State == "deleted" || m.Raw.State == "absent") {
		m.CaptureState = "deleted"
		return writePurgeManifest(op, *m)
	}
	if err != nil {
		return err
	}
	defer d.Close()
	id, err := admissionFileID(d.file)
	if err != nil || id != m.Record.SourceCaptureDirID {
		return errors.Join(err, errors.New("retained source directory identity changed"))
	}
	if err = s.purgeArtifactStep(op, d, d, "source", ".purge-"+m.OperationID, "raw", &m.Raw, m); err != nil {
		return err
	}
	// A fresh directory stream must see the completed rename/unlink operations.
	fresh, err := openAdmissionRoot(m.Record.SourceCaptureDir, false)
	if err != nil {
		return err
	}
	freshID, err := admissionFileID(fresh.file)
	if err != nil || freshID != id {
		fresh.Close()
		return errors.Join(err, errors.New("retained source directory changed during purge"))
	}
	entries, err := fresh.file.ReadDir(1)
	fresh.Close()
	if err != nil && !errors.Is(err, io.EOF) || len(entries) != 0 {
		return errors.Join(err, errors.New("unexpected retained source entries preserved"))
	}
	parent, err := platformRestoreDirectory(filepath.Dir(m.Record.SourceCaptureDir))
	if err != nil {
		return err
	}
	defer parent.Close()
	if err = s.purgeFault("capture-delete-intent"); err != nil {
		return err
	}
	if err = d.stillNamed(); err != nil {
		return err
	}
	if err = d.Close(); err != nil {
		return err
	}
	if err = admissionRemoveDirectory(parent, filepath.Base(m.Record.SourceCaptureDir)); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err = admissionSyncDirectory(parent); err != nil {
		return err
	}
	if err = s.purgeFault("capture-unlinked"); err != nil {
		return err
	}
	m.CaptureState = "deleted"
	if err = writePurgeManifest(op, *m); err != nil {
		return err
	}
	return s.purgeFault("capture-deleted")
}

func requirePurgeNameAbsent(d *restoreDirectory, name string) error {
	f, err := purgeOpenFile(d, name)
	if os.IsNotExist(err) {
		return nil
	}
	if f != nil {
		f.Close()
	}
	return errors.Join(err, errors.New("replacement at the purged stored name was preserved"))
}

func requirePurgeCaptureAbsent(rec cache.QuarantineRecord) error {
	if rec.SourceCaptureDir == "" {
		return nil
	}
	d, err := openAdmissionRoot(rec.SourceCaptureDir, false)
	if os.IsNotExist(err) {
		return nil
	}
	if d != nil {
		d.Close()
	}
	return errors.Join(err, errors.New("replacement at the retained source directory was preserved"))
}

func purgeJournalTemporary(name string) bool {
	if !strings.HasPrefix(name, ".journal-") {
		return false
	}
	text := strings.TrimPrefix(name, ".journal-")
	id, err := uuid.Parse(text)
	return err == nil && id != uuid.Nil && id.String() == text
}

func validatePurgeOperation(journals, op *restoreDirectory, m purgeManifest) error {
	fresh, err := admissionOpenChild(journals, m.Record.QuarantineID)
	if err != nil {
		return err
	}
	defer fresh.Close()
	oldID, e1 := admissionFileID(op.file)
	newID, e2 := admissionFileID(fresh.file)
	if e1 != nil || e2 != nil || oldID != newID {
		return errors.Join(e1, e2, errors.New("purge journal directory identity changed"))
	}
	entries, err := fresh.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > maxAdmissionRecovery {
		return errors.Join(err, errors.New("purge journal directory exceeds its limit"))
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return errors.New("unexpected purge journal entry type retained")
		}
		if entry.Name() == "snapshot" {
			if m.Stored.State != "claiming" && m.Stored.State != "claimed" {
				return errors.New("unexpected retired snapshot retained")
			}
			artifact, err := inspectPurgeArtifact(op, "snapshot")
			if err != nil || artifact.ID != m.Stored.ID {
				return errors.Join(err, errors.New("retired purge snapshot identity changed"))
			}
		} else if entry.Name() != "manifest.json" && !purgeJournalTemporary(entry.Name()) {
			return errors.New("unknown purge journal entry retained")
		}
	}
	return nil
}

func (s *Store) resumePurge(root, journals, op *restoreDirectory, m *purgeManifest) error {
	if err := checkNoRestoreOperation(root, m.Record.QuarantineID); err != nil {
		return err
	}
	committed, err := s.cache.ValidateQuarantinePurge(m.Record, m.OperationID)
	if err != nil {
		return err
	}
	if err = validatePurgeOperation(journals, op, *m); err != nil {
		return err
	}
	if committed && !purgeArtifactsRemoved(*m) {
		return errors.New("committed purge has inconsistent filesystem journal state")
	}
	if !committed {
		m.Phase = "deleting"
		if err = writePurgeManifest(op, *m); err != nil {
			return err
		}
		if err = s.purgeArtifactStep(op, root, op, m.Record.QuarantineID+".quarantined", "snapshot", "stored", &m.Stored, m); err != nil {
			return err
		}
		if err = s.purgeCapture(op, m); err != nil {
			return err
		}
		if err = requirePurgeNameAbsent(root, m.Record.QuarantineID+".quarantined"); err != nil {
			return err
		}
		if !purgeArtifactsRemoved(*m) {
			return errors.New("purge artifacts have not all been reconciled")
		}
		if err = errors.Join(requirePurgeNameAbsent(op, "snapshot"), requirePurgeCaptureAbsent(m.Record)); err != nil {
			return err
		}
		if err = s.purgeFault("before-database-commit"); err != nil {
			return err
		}
		if err = s.cache.CommitQuarantinePurge(m.Record, m.OperationID, time.Now().UTC()); err != nil {
			return err
		}
		if err = s.purgeFault("database-committed"); err != nil {
			return err
		}
	}
	m.Phase = "committed"
	if err = writePurgeManifest(op, *m); err != nil {
		return err
	}
	if err = s.purgeFault("committed"); err != nil {
		return err
	}
	if err = requirePurgeNameAbsent(root, m.Record.QuarantineID+".quarantined"); err != nil {
		return err
	}
	return cleanupPurgeOperation(journals, op, m.Record.QuarantineID, s.purgeFault)
}

func cleanupPurgeOperation(journals, op *restoreDirectory, id string, fault func(string) error) error {
	fresh, err := admissionOpenChild(journals, id)
	if err != nil {
		return err
	}
	defer fresh.Close()
	before, e1 := admissionFileID(op.file)
	after, e2 := admissionFileID(fresh.file)
	if e1 != nil || e2 != nil || before != after {
		return errors.Join(e1, e2, errors.New("purge journal directory identity changed"))
	}
	entries, err := fresh.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) || len(entries) > maxAdmissionRecovery {
		return errors.Join(err, errors.New("purge journal directory exceeds its limit"))
	}
	for _, entry := range entries {
		if entry.Name() != "manifest.json" && !purgeJournalTemporary(entry.Name()) || !entry.Type().IsRegular() {
			return errors.New("unknown purge journal entry retained")
		}
	}
	for _, entry := range entries {
		if entry.Name() == "manifest.json" {
			continue
		}
		if err = fresh.removeFile(entry.Name()); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err = admissionSyncDirectory(fresh); err != nil {
		return err
	}
	// Remove the manifest last: interrupted cleanup then leaves either a valid
	// committed journal or an empty directory, never unclassified temp files.
	if err = fresh.removeFile("manifest.json"); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err = admissionSyncDirectory(fresh); err != nil {
		return err
	}
	if fault != nil {
		if err = fault("journal-cleanup"); err != nil {
			return err
		}
	}
	if err = fresh.Close(); err != nil {
		return err
	}
	if err = op.Close(); err != nil {
		return err
	}
	if err = admissionRemoveDirectory(journals, id); err != nil && !os.IsNotExist(err) {
		return err
	}
	return admissionSyncDirectory(journals)
}
