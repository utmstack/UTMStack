package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// RecoveryInspection is evidence for operator review, not permission to adopt
// or delete an object. Each journal is a separate observation: a running service
// can advance recovery between observations. Stop it for a stable case capture.
type RecoveryInspection struct {
	Version      int                 `json:"version"`
	QuarantineID string              `json:"quarantine_id"`
	ObservedAt   time.Time           `json:"observed_at"`
	RootID       string              `json:"root_id"`
	Journals     []JournalInspection `json:"journals"`
}

type JournalInspection struct {
	Operation   string `json:"operation"`
	EntryKind   string `json:"entry_kind,omitempty"`
	Status      string `json:"status"` // absent, captured, unreadable
	DirectoryID string `json:"directory_id,omitempty"`
	ManifestID  string `json:"manifest_id,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	// JSON encodes bytes as base64, preserving even malformed journals and
	// lossless Unix path encodings. Never interpret this as validated approval.
	Manifest []byte `json:"manifest,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

// InspectRecovery opens existing private directories without hardening them,
// creating locks, opening SQLite, or replaying recovery. Only the three fixed
// journal locations for an ordinary quarantine ID are read. Manifest reads are bounded;
// captured hashes describe exact bytes, without asserting their validity.
func InspectRecovery(dir, id string) (RecoveryInspection, error) {
	report := RecoveryInspection{Version: 1, QuarantineID: id, ObservedAt: time.Now().UTC(), Journals: make([]JournalInspection, 0, 3)}
	if !ordinaryQuarantineID(id) {
		return report, errors.New("invalid quarantine id")
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return report, err
	}
	// Match New's resolved storage location without creating or changing it.
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return report, err
	}
	root, err := openAdmissionRoot(resolved, false)
	if err != nil {
		return report, err
	}
	defer root.Close()
	if report.RootID, err = admissionFileID(root.file); err != nil {
		return report, err
	}
	var failures error
	for _, operation := range []struct{ name, directory string }{
		{"admission", ".admissions"}, {"restore", ".restores"}, {"purge", ".purges"},
	} {
		entry, e := inspectJournal(root, operation.directory, id)
		entry.Operation = operation.name
		if e != nil {
			entry.Status, entry.Detail = "unreadable", e.Error()
			failures = errors.Join(failures, fmt.Errorf("%s inspection: %w", operation.name, e))
		}
		report.Journals = append(report.Journals, entry)
	}
	return report, failures
}

func inspectJournal(root *restoreDirectory, directory, id string) (JournalInspection, error) {
	entry := JournalInspection{Status: "absent"}
	journals, err := admissionOpenChild(root, directory)
	if os.IsNotExist(err) {
		return entry, nil
	}
	if err != nil {
		return entry, err
	}
	defer journals.Close()
	op, err := admissionOpenChild(journals, id)
	if os.IsNotExist(err) {
		return entry, nil
	}
	if err != nil {
		if directory == ".purges" {
			// A reviewed retention fence intentionally occupies the old
			// directory-only namespace as a regular file. Capture even invalid
			// bytes without interpreting them as an approved decision.
			f, fileErr := admissionReadFile(journals, id)
			if fileErr == nil {
				defer f.Close()
				entry.EntryKind = "file"
				entry.DirectoryID, fileErr = admissionFileID(journals.file)
				if fileErr != nil {
					return entry, fileErr
				}
				return inspectJournalFile(entry, f, maxRecoveryReview)
			}
		}
		return entry, err
	}
	defer op.Close()
	if entry.DirectoryID, err = admissionFileID(op.file); err != nil {
		return entry, err
	}
	f, err := admissionReadFile(op, "manifest.json")
	if err != nil {
		return entry, err // An existing operation without a manifest is not absent.
	}
	defer f.Close()
	return inspectJournalFile(entry, f, maxAdmissionJournal)
}

func inspectJournalFile(entry JournalInspection, f *os.File, limit int64) (JournalInspection, error) {
	info, err := f.Stat()
	if err != nil {
		return entry, err
	}
	if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > limit {
		return entry, errors.New("inspection requires a regular journal within the size limit")
	}
	if entry.ManifestID, err = admissionFileID(f); err != nil {
		return entry, err
	}
	body, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return entry, err
	}
	if int64(len(body)) > limit {
		return entry, errors.New("journal grew beyond the inspection size limit")
	}
	sum := sha256.Sum256(body)
	entry.Status, entry.SHA256, entry.Manifest = "captured", hex.EncodeToString(sum[:]), body
	return entry, nil
}
