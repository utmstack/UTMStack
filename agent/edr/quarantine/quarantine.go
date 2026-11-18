package quarantine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Store struct {
	dir            string
	cache          *cache.Cache
	rootID         string
	recovery       []RecoveryResult
	admissionFault func(string) error
	admissionCopy  func(io.Writer, io.Reader) (int64, error)
	admissionSync  func(*os.File) error
	responseFault  func(string) error
	// Per-store fault seams keep concurrency tests independent of other stores.
	restoreCopy          func(io.Writer, io.Reader) (int64, error)
	restoreBeforePublish func() error
	restoreAfterPublish  func() error
}

func New(dir string, c *cache.Cache) (*Store, error) {
	if c == nil {
		return nil, errors.New("quarantine requires durable metadata storage")
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(absolute, 0700); err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	root, err := openAdmissionRoot(resolved, true)
	if err != nil {
		return nil, err
	}
	id, err := admissionFileID(root.file)
	_ = root.Close()
	if err != nil {
		return nil, err
	}
	s := &Store{dir: resolved, cache: c, rootID: id}
	if results, recoveryErr := s.Recover(); recoveryErr != nil {
		s.recovery = results
		logger.Error("UTMStack EDR: quarantine recovery needs attention: %v", recoveryErr)
	} else {
		s.recovery = results
	}
	return s, nil
}

// Quarantine validates the requested digest before acting. Scanner callers also
// supply their original descriptor identity through QuarantineVerified.
func (s *Store) Quarantine(path, hash, detection string) (string, error) {
	return s.admit(path, hash, detection, nil)
}

// QuarantineVerified requires identity from the original scan descriptor's
// File.Stat, as used by Scanner. Path-based os.Stat may defer its Windows
// file-ID lookup until after a pathname has been replaced.
func (s *Store) QuarantineVerified(path, hash, detection string, scanned os.FileInfo) (string, error) {
	if scanned == nil {
		return "", fmt.Errorf("%w: missing scan identity", ErrAdmissionChanged)
	}
	return s.admit(path, hash, detection, scanned)
}

func (s *Store) Restore(id string) error {
	if !ordinaryQuarantineID(id) {
		return fmt.Errorf("invalid quarantine id")
	}
	root, err := s.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	unlock, err := root.lockItem(id)
	if err != nil {
		return err
	}
	defer unlock()
	if err := checkNoRecoveryHold(root, id); err != nil {
		return err
	}
	if journals, pending, pendingErr := openPurgeOperation(root, id); pendingErr == nil {
		pending.Close()
		journals.Close()
		return &PurgePendingError{QuarantineID: id, Err: errors.New("finish the pending purge before another management operation")}
	} else if !os.IsNotExist(pendingErr) {
		return pendingErr
	}
	return s.restoreJournaled(root, id)
}

// Purge permanently deletes a quarantined item's stored file and marks the
// record purged (kept for audit). Unlike the automatic pipeline — which never
// deletes — this is a DELIBERATE admin action. An already-restored item has no
// stored file; purging it just marks the record.
func (s *Store) Purge(id string) error {
	if !ordinaryQuarantineID(id) {
		return fmt.Errorf("invalid quarantine id")
	}
	root, err := s.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	unlock, err := root.lockItem(id)
	if err != nil {
		return err
	}
	defer unlock()
	if err := checkNoRecoveryHold(root, id); err != nil {
		return err
	}
	if err := checkNoRestoreOperation(root, id); err != nil {
		return err
	}
	return s.purgeJournaled(root, id)
}

// PurgeExpired purges retained artifacts older than retentionDays (relative to
// `now`), including raw source inodes kept after restore. Purged records and
// restored legacy items with no retained raw source are skipped.
// retentionDays <= 0 is a no-op (keep forever). Used by the retention sweep.
func (s *Store) PurgeExpired(retentionDays int, now time.Time) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	recs, err := s.cache.ListQuarantine()
	if err != nil {
		return 0, err
	}
	cutoff := now.Add(-time.Duration(retentionDays) * 24 * time.Hour)
	purged := 0
	var failures error
	for _, r := range recs {
		if r.Restored && r.SourceCaptureDir == "" || r.Purged || r.QuarantinedAt.After(cutoff) {
			continue
		}
		if err := s.Purge(r.QuarantineID); err == nil {
			purged++
		} else if errors.Is(err, ErrRecoveryHeld) || errors.Is(err, ErrRecoveryRetired) {
			continue
		} else {
			failures = errors.Join(failures, fmt.Errorf("purge expired item %s: %w", r.QuarantineID, err))
		}
	}
	return purged, failures
}

// renameFile atomically refuses an occupied destination. Tests replace this
// seam to exercise the exclusive copy fallback used across volumes.
var renameFile = renameNoReplace

func renameError(src, dst string, err error) error {
	if err != nil {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}
	return nil
}

// moveFile moves without replacing any existing destination entry, including a
// symlink, falling back to exclusive copy+remove across volumes. The fallback
// never leaves two copies: when the source cannot be removed after the copy (a
// running image, a file another process holds open) the new copy is deleted
// again and the move fails. That copy is only a duplicate of a file that stays
// in place, so deleting it never loses the original. The copy is created
// exclusively, so the fallback only ever deletes a file it created itself.
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	if err := renameFile(src, dst); err == nil {
		return nil
	} else if os.IsExist(err) {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		in.Close()
		return err
	}
	_, err = io.Copy(out, in)
	in.Close() // Windows refuses to remove a file this process still has open
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		if err = os.Remove(src); err == nil {
			return nil
		}
		err = fmt.Errorf("source cannot be removed after copy: %w", err)
	}
	if rerr := os.Remove(dst); rerr != nil {
		return fmt.Errorf("%w; the copy %s could not be removed: %v", err, dst, rerr)
	}
	return err
}
