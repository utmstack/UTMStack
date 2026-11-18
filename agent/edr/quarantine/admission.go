package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
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

var ErrAdmissionChanged = errors.New("quarantine target changed or is ineligible")
var ErrAlreadyQuarantined = errors.New("file is already retained by quarantine")

// ContainmentError reports a physically completed, durably journaled operation
// whose metadata/cleanup needs recovery. A caller must not say the file stayed
// in place merely because finalization returned an error.
type ContainmentError struct {
	QuarantineID string
	Err          error
}

func (e *ContainmentError) Error() string {
	return fmt.Sprintf("file contained as %s; finalization needs recovery: %v", e.QuarantineID, e.Err)
}
func (e *ContainmentError) Unwrap() error { return e.Err }

const maxAdmissionJournal = 1 << 20
const maxAdmissionRecovery = 1024

type admissionManifest struct {
	Version    int
	Phase      string
	Record     cache.QuarantineRecord
	Parent     string
	ParentID   string
	Leaf       string
	SnapshotID string
}

type RecoveryResult struct {
	QuarantineID string
	Operation    string
	Status       string
	OriginalPath string
	Detail       string
}

func (s *Store) RecoveryStatus() []RecoveryResult {
	return append([]RecoveryResult(nil), s.recovery...)
}

func (s *Store) openRoot() (*restoreDirectory, error) {
	d, err := openAdmissionRoot(s.dir, false)
	if err != nil {
		return nil, err
	}
	id, err := admissionFileID(d.file)
	if err != nil || id != s.rootID {
		d.Close()
		return nil, fmt.Errorf("%w: quarantine directory changed", ErrAdmissionChanged)
	}
	return d, nil
}

func admissionChild(parent *restoreDirectory, name string) (*restoreDirectory, error) {
	d, err := admissionOpenChild(parent, name)
	if os.IsNotExist(err) {
		d, err = admissionMakeDirectory(parent, name)
		if os.IsExist(err) {
			return admissionOpenChild(parent, name)
		}
		if err == nil {
			err = admissionSyncDirectory(parent)
		}
	}
	if err != nil && d != nil {
		d.Close()
		d = nil
	}
	return d, err
}

func (s *Store) fault(phase string) error {
	if s.admissionFault != nil {
		return s.admissionFault(phase)
	}
	return nil
}

func (s *Store) managedAdmissionParent(root, parent *restoreDirectory) error {
	parentID, err := admissionFileID(parent.file)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(s.dir, parent.path)
	if parentID == s.rootID || err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		return ErrAlreadyQuarantined
	}
	if _, found, err := s.cache.QuarantineByCaptureIdentity(parentID); err != nil {
		return err
	} else if found {
		return ErrAlreadyQuarantined
	}
	if err := s.managedRestoreStage(parent, parentID); err != nil {
		return err
	}
	base := filepath.Base(parent.path)
	if !strings.HasPrefix(base, ".edr-quarantine-") {
		return nil
	}
	id := strings.TrimPrefix(base, ".edr-quarantine-")
	if !ordinaryQuarantineID(id) {
		return nil
	}
	journals, err := admissionOpenChild(root, ".admissions")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer journals.Close()
	op, err := admissionOpenChild(journals, id)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer op.Close()
	m, err := readAdmissionManifest(op, id)
	if err != nil {
		return err
	}
	if m.Record.SourceCaptureDirID == parentID {
		return ErrAlreadyQuarantined
	}
	return nil
}

func writeAdmissionManifest(op *restoreDirectory, m admissionManifest) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return writePrivateManifest(op, body)
}

func writePrivateManifest(op *restoreDirectory, body []byte) error {
	if len(body) > maxAdmissionJournal {
		return errors.New("quarantine recovery metadata exceeds its limit")
	}
	name := ".journal-" + uuid.NewString()
	f, err := admissionCreateFile(op, name)
	if err != nil {
		return err
	}
	_, err = f.Write(body)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err == nil {
		err = admissionRename(op, name, op, "manifest.json", true)
	}
	if err != nil {
		_ = op.removeFile(name)
		return err
	}
	return admissionSyncDirectory(op)
}

func readAdmissionManifest(op *restoreDirectory, id string) (admissionManifest, error) {
	var m admissionManifest
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return m, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxAdmissionJournal {
		return m, errors.New("invalid quarantine recovery metadata")
	}
	b, err := io.ReadAll(io.LimitReader(f, maxAdmissionJournal+1))
	if err != nil {
		return m, err
	}
	if len(b) > maxAdmissionJournal {
		return m, errors.New("oversized quarantine recovery metadata")
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	leaf, leafErr := restoreLeaf(filepath.Join(m.Parent, m.Leaf))
	if m.Version != 1 || m.Record.QuarantineID != id || !ordinaryQuarantineID(id) || !filepath.IsAbs(m.Record.OriginalPath) || filepath.Base(m.Record.OriginalPath) != m.Leaf || !filepath.IsAbs(m.Parent) || filepath.Clean(m.Parent) != m.Parent || m.ParentID == "" || leafErr != nil || leaf != m.Leaf || !m.Record.Restorable || m.Record.Restored || m.Record.Purged || m.Record.Engine != "UTMStack EDR" || m.Record.SourceCaptureDir != filepath.Join(m.Parent, ".edr-quarantine-"+id) || m.Record.SourceFileID == "" {
		return m, errors.New("invalid quarantine recovery identities")
	}
	if _, err = restoreDigest(m.Record.SHA256); err != nil {
		return m, err
	}
	switch m.Phase {
	case "preparing", "ready", "claimed", "published":
	default:
		return m, errors.New("invalid quarantine recovery phase")
	}
	if m.Phase != "preparing" && (m.SnapshotID == "" || m.Record.SourceCaptureDirID == "") {
		return m, errors.New("missing prepared object identities")
	}
	return m, nil
}

func cleanupAdmissionOperation(journals, op *restoreDirectory, id string) error {
	// Enumerate through a newly opened handle after journal rename/publication.
	// A directory stream opened before those mutations can miss new entries on
	// Linux tmpfs. Keep the original identity pinned and refuse any replacement.
	fresh, err := admissionOpenChild(journals, id)
	if err != nil {
		return err
	}
	defer fresh.Close()
	before, beforeErr := admissionFileID(op.file)
	after, afterErr := admissionFileID(fresh.file)
	if beforeErr != nil || afterErr != nil || before != after {
		return errors.Join(beforeErr, afterErr, errors.New("recovery directory identity changed"))
	}
	entries, err := fresh.file.ReadDir(maxAdmissionRecovery + 1)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		return err
	}
	if len(entries) > maxAdmissionRecovery {
		return errors.New("recovery operation exceeds its bounded entry limit")
	}
	for _, e := range entries {
		if e.Name() != "manifest.json" && e.Name() != "snapshot" && !strings.HasPrefix(e.Name(), ".journal-") {
			return fmt.Errorf("unknown recovery entry %q retained", e.Name())
		}
		if e.IsDir() {
			return errors.New("unexpected recovery subdirectory retained")
		}
	}
	// Validate the complete bounded listing before deleting any journal data.
	for _, e := range entries {
		if err = fresh.removeFile(e.Name()); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err = admissionSyncDirectory(fresh); err != nil {
		return err
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

func sameAdmissionInfo(before, after os.FileInfo) bool {
	return before != nil && after != nil && os.SameFile(before, after) && before.Size() == after.Size() && before.ModTime().Equal(after.ModTime())
}

func (s *Store) copyAdmission(out, in *os.File, info os.FileInfo, expected [sha256.Size]byte) error {
	copyFn := s.admissionCopy
	if copyFn == nil {
		copyFn = io.Copy
	}
	h := sha256.New()
	n, err := copyFn(io.MultiWriter(out, h), io.NewSectionReader(in, 0, info.Size()))
	if err != nil {
		return err
	}
	if n != info.Size() || !equalRestoreHash(h.Sum(nil), expected) {
		return fmt.Errorf("%w: source no longer holds the analyzed bytes", ErrAdmissionChanged)
	}
	after, err := in.Stat()
	if err != nil {
		return err
	}
	if !sameAdmissionInfo(info, after) {
		return fmt.Errorf("%w during snapshot copy", ErrAdmissionChanged)
	}
	if err = admissionSnapshotOwner(out, info); err != nil {
		return err
	}
	return out.Sync()
}

func (s *Store) syncAdmissionSource(f *os.File) error {
	if s.admissionSync != nil {
		return s.admissionSync(f)
	}
	return admissionSyncSource(f)
}

func (s *Store) admit(path, hash, detection string, scanned os.FileInfo) (id string, result error) {
	expected, err := restoreDigest(hash)
	if err != nil {
		return "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	leaf, err := restoreLeaf(path)
	if err != nil {
		return "", err
	}
	root, err := s.openRoot()
	if err != nil {
		return "", err
	}
	defer root.Close()
	// A known unavailable database must not cause an unnecessary namespace move.
	if _, _, err = s.cache.GetQuarantine("admission-health-probe"); err != nil {
		return "", fmt.Errorf("quarantine record preparation failed: %w", err)
	}
	parent, err := openRestoreDirectory(filepath.Dir(path), false)
	if err != nil {
		return "", err
	}
	defer parent.Close()
	if err = s.managedAdmissionParent(root, parent); err != nil {
		return "", err
	}
	if err = s.fault("parent-checked"); err != nil {
		return "", err
	}
	in, err := admissionOpenSource(parent, leaf)
	if err != nil {
		return "", err
	}
	defer in.Close()
	// A restore can register this previously empty directory and create its
	// snapshot after our first lookup. Recheck after opening the source: restore
	// records ownership before creating any output, and no namespace has moved.
	if err = s.managedAdmissionParent(root, parent); err != nil {
		return "", err
	}
	info, err := in.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxRestoreBytes || scanned != nil && !sameAdmissionInfo(scanned, info) {
		return "", fmt.Errorf("%w: source identity, type or size differs from the scan", ErrAdmissionChanged)
	}
	fileID, err := admissionFileID(in)
	if err != nil {
		return "", err
	}
	parentID, err := admissionFileID(parent.file)
	if err != nil {
		return "", err
	}
	originalSecurity, err := admissionOriginalSecurity(in)
	if err != nil {
		return "", err
	}
	// Persist the pinned raw input before recording intent or moving a name.
	// A durable independent snapshot cannot flush the original file's data.
	if err = s.syncAdmissionSource(in); err != nil {
		return "", fmt.Errorf("quarantine source synchronization failed: %w", err)
	}
	if err = s.fault("source-synced"); err != nil {
		return "", err
	}
	id = uuid.NewString()
	operationID := id
	confirmedCapture := false
	defer func() {
		if result != nil && confirmedCapture {
			var contained *ContainmentError
			if !errors.As(result, &contained) {
				result = &ContainmentError{operationID, result}
			}
			id = operationID
		}
	}()
	unlock, err := root.lockItem(id)
	if err != nil {
		return "", err
	}
	defer unlock()
	journals, err := admissionChild(root, ".admissions")
	if err != nil {
		return "", err
	}
	defer journals.Close()
	queued, queueErr := journals.file.ReadDir(maxAdmissionRecovery + 1)
	if queueErr != nil && !errors.Is(queueErr, io.EOF) {
		return "", queueErr
	}
	if len(queued) >= maxAdmissionRecovery {
		return "", errors.New("quarantine recovery queue is full")
	}
	op, err := admissionMakeDirectory(journals, id)
	if err != nil {
		return "", err
	}
	defer op.Close()
	var capture *restoreDirectory
	sourceClaimed := false
	defer func() {
		if sourceClaimed {
			return
		}
		if capture != nil {
			_ = capture.Close()
			cleanupErr := admissionRemoveDirectory(parent, ".edr-quarantine-"+operationID)
			if cleanupErr != nil && !os.IsNotExist(cleanupErr) {
				result = errors.Join(result, fmt.Errorf("uncommitted capture retained for recovery: %w", cleanupErr))
				return // Never discard the journal while an entry may be captured.
			}
		}
		if _, openErr := op.file.Stat(); errors.Is(openErr, os.ErrClosed) {
			return
		}
		result = errors.Join(result, cleanupAdmissionOperation(journals, op, operationID))
	}()
	m := admissionManifest{Version: 1, Phase: "preparing", Parent: parent.path, ParentID: parentID, Leaf: leaf, Record: cache.QuarantineRecord{QuarantineID: id, OriginalPath: path, SHA256: hex.EncodeToString(expected[:]), Detection: detection, Engine: "UTMStack EDR", QuarantinedAt: time.Now().UTC(), Restorable: true, SourceCaptureDir: filepath.Join(parent.path, ".edr-quarantine-"+id), SourceFileID: fileID}}
	m.Record.OriginalSecurity = originalSecurity
	if err = writeAdmissionManifest(op, m); err != nil {
		return "", err
	}
	if err = admissionSyncDirectory(journals); err != nil {
		return "", err
	}
	// All abrupt-exit seams run after the preceding durable boundary.
	if err = s.fault("intent"); err != nil {
		return "", err
	}
	capture, err = admissionMakeDirectory(parent, filepath.Base(m.Record.SourceCaptureDir))
	if err != nil {
		_ = cleanupAdmissionOperation(journals, op, id)
		return "", err
	}
	defer capture.Close()
	m.Record.SourceCaptureDirID, err = admissionFileID(capture.file)
	if err != nil {
		return "", err
	}
	out, err := admissionCreateFile(op, "snapshot")
	if err != nil {
		return "", err
	}
	m.SnapshotID, err = admissionFileID(out)
	if err == nil {
		err = s.fault("snapshot-created")
	}
	if err == nil {
		err = s.copyAdmission(out, in, info, expected)
	}
	err = errors.Join(err, out.Close())
	if err != nil {
		_ = capture.Close()
		_ = admissionRemoveDirectory(parent, filepath.Base(m.Record.SourceCaptureDir))
		_ = cleanupAdmissionOperation(journals, op, id)
		return "", err
	}
	if err = s.fault("snapshot"); err != nil {
		return "", err
	}
	m.Phase = "ready"
	if err = admissionSyncDirectory(parent); err == nil {
		err = writeAdmissionManifest(op, m)
	}
	if err != nil {
		return "", err
	}
	if err = s.fault("ready"); err != nil {
		return "", err
	}
	if err = parent.stillNamed(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrAdmissionChanged, err)
	}
	if err = admissionClaim(in, parent, leaf, capture, "source"); err != nil {
		return "", err
	}
	sourceClaimed = true
	if err = s.fault("claimed-before-check"); err != nil {
		return "", err
	}
	current, openErr := admissionReadPinnedSource(capture, "source")
	if openErr == nil {
		currentID, e := admissionFileID(current)
		if e != nil || currentID != fileID {
			openErr = fmt.Errorf("%w: replacement entry captured", ErrAdmissionChanged)
		}
		_ = current.Close()
	}
	if openErr == nil {
		openErr = parent.stillNamed()
	}
	if openErr == nil {
		after, e := in.Stat()
		if e != nil {
			openErr = e
		} else if !sameAdmissionInfo(info, after) {
			openErr = ErrAdmissionChanged
		} else {
			openErr = verifiedRestoreHash(in, info, expected)
		}
	}
	if openErr != nil {
		if rollback := admissionRollback(in, capture, "source", parent, leaf); rollback != nil {
			return "", fmt.Errorf("%w; captured entry retained at %s (rollback failed: %v)", openErr, m.Record.SourceCaptureDir, rollback)
		}
		_ = admissionSyncDirectory(parent)
		_ = capture.Close()
		_ = admissionRemoveDirectory(parent, filepath.Base(m.Record.SourceCaptureDir))
		_ = cleanupAdmissionOperation(journals, op, id)
		return "", openErr
	}
	// The source name is already captured and its identity/content proved.
	// Later ACL or sync failures must report that physical result accurately.
	confirmedCapture = true
	if err = s.fault("verified"); err != nil {
		return "", err
	}
	if err = admissionProtectCaptured(in); err != nil {
		return "", err
	}
	if err = s.syncAdmissionSource(in); err != nil {
		return "", fmt.Errorf("captured source synchronization failed: %w", err)
	}
	if err = s.fault("captured-synced"); err != nil {
		return "", err
	}
	if err = admissionSyncDirectory(capture); err == nil {
		err = admissionSyncDirectory(parent)
	}
	if err != nil {
		return "", err
	}
	m.Phase = "claimed"
	if err = writeAdmissionManifest(op, m); err != nil {
		return "", err
	}
	if err = s.fault("claimed"); err != nil {
		return "", err
	}
	if err = admissionRename(op, "snapshot", root, id+".quarantined", false); err != nil {
		return "", err
	}
	if err = admissionSyncDirectory(root); err != nil {
		return id, &ContainmentError{id, err}
	}
	m.Phase = "published"
	if err = writeAdmissionManifest(op, m); err != nil {
		return id, &ContainmentError{id, err}
	}
	if err = s.fault("published"); err != nil {
		return id, &ContainmentError{id, err}
	}
	if err = s.cache.CommitQuarantine(m.Record); err != nil {
		return id, &ContainmentError{id, err}
	}
	_ = in.Chmod(0600)
	if err = s.fault("recorded"); err != nil {
		return id, &ContainmentError{id, err}
	}
	if err = cleanupAdmissionOperation(journals, op, id); err != nil {
		return id, &ContainmentError{id, err}
	}
	return id, nil
}
