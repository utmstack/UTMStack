package quarantine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type Store struct {
	dir   string
	cache *cache.Cache
}

func New(dir string, c *cache.Cache) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir, cache: c}, nil
}

// Quarantine moves the file into the protected store and records it.
// It never deletes the source on failure.
func (s *Store) Quarantine(originalPath, sha256, detection string) (string, error) {
	id := uuid.NewString()
	dest := filepath.Join(s.dir, id+".quarantined")

	if err := moveFile(originalPath, dest); err != nil {
		return "", fmt.Errorf("quarantine move failed: %w", err)
	}
	// Best-effort: strip execute/normal attributes on the stored copy.
	_ = os.Chmod(dest, 0o600)

	rec := cache.QuarantineRecord{
		QuarantineID:  id,
		OriginalPath:  originalPath,
		SHA256:        sha256,
		Detection:     detection,
		Engine:        "UTMStack EDR",
		QuarantinedAt: time.Now().UTC(),
		Restorable:    true,
	}
	if err := s.cache.StoreQuarantine(rec); err != nil {
		// Roll the file back so we never lose it silently.
		_ = moveFile(dest, originalPath)
		return "", fmt.Errorf("quarantine record failed: %w", err)
	}
	return id, nil
}

func (s *Store) Restore(id string) error {
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("quarantine id not found: %s", id)
	}
	if rec.Restored {
		return fmt.Errorf("quarantine id %s was already restored", id)
	}
	dest := filepath.Join(s.dir, id+".quarantined")
	if err := moveFile(dest, rec.OriginalPath); err != nil {
		return err
	}
	return s.cache.MarkRestored(id)
}

// Purge permanently deletes a quarantined item's stored file and marks the
// record purged (kept for audit). Unlike the automatic pipeline — which never
// deletes — this is a DELIBERATE admin action. An already-restored item has no
// stored file; purging it just marks the record.
func (s *Store) Purge(id string) error {
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("quarantine id not found: %s", id)
	}
	if rec.Purged {
		return fmt.Errorf("quarantine id %s was already purged", id)
	}
	dest := filepath.Join(s.dir, id+".quarantined")
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing quarantined file: %w", err)
	}
	return s.cache.MarkPurged(id)
}

// PurgeExpired purges still-quarantined items older than retentionDays (relative
// to `now`), skipping ones already restored or purged. Returns the count purged.
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
	for _, r := range recs {
		if r.Restored || r.Purged || r.QuarantinedAt.After(cutoff) {
			continue
		}
		if err := s.Purge(r.QuarantineID); err == nil {
			purged++
		}
	}
	return purged, nil
}

// moveFile renames, falling back to copy+remove across volumes.
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		in.Close()
		out.Close()
		return err
	}
	in.Close()
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
