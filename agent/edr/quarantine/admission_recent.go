package quarantine

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// RecentQuarantineByPath also consults durable, verified admissions when SQL
// finalization was interrupted. A running malicious image must not lose its
// verdict merely because containment removed its pathname before SQL committed.
func (s *Store) RecentQuarantineByPath(path string, within time.Duration) (cache.QuarantineRecord, bool, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return cache.QuarantineRecord{}, false, err
	}
	best, found, cacheErr := s.cache.RecentQuarantineByPath(path, within)
	if cacheErr == nil && found {
		return best, true, nil
	}
	root, err := s.openRoot()
	if err != nil {
		return best, found, errors.Join(cacheErr, err)
	}
	defer root.Close()
	journals, err := admissionOpenChild(root, ".admissions")
	if os.IsNotExist(err) {
		return best, found, cacheErr
	}
	if err != nil {
		return best, found, errors.Join(cacheErr, err)
	}
	defer journals.Close()
	entries, err := journals.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return best, found, err
	}
	if len(entries) > maxAdmissionRecovery {
		return best, found, errors.New("quarantine recovery queue exceeds its limit")
	}
	cutoff := time.Now().UTC().Add(-within)
	for _, entry := range entries {
		if !entry.IsDir() || !ordinaryQuarantineID(entry.Name()) {
			continue
		}
		op, e := admissionOpenChild(journals, entry.Name())
		if e != nil {
			continue
		}
		m, e := readAdmissionManifest(op, entry.Name())
		_ = op.Close()
		if e != nil || m.Phase != "claimed" && m.Phase != "published" {
			continue
		}
		match := m.Record.OriginalPath == path
		if runtime.GOOS == "windows" {
			match = strings.EqualFold(m.Record.OriginalPath, path)
		}
		if match && !m.Record.QuarantinedAt.Before(cutoff) && (!found || m.Record.QuarantinedAt.After(best.QuarantinedAt)) {
			best, found = m.Record, true
		}
	}
	if found {
		return best, true, nil
	}
	return best, false, cacheErr
}
