package quarantine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Recover reconciles private admission, restore and approved purge journals. It never
// starts an unapproved purge or newly quarantines an original path on restart.
func (s *Store) Recover() ([]RecoveryResult, error) {
	admissions, admissionErr := s.recoverAdmissions()
	restores, restoreErr := s.recoverRestores()
	purges, purgeErr := s.recoverPurges()
	return append(append(admissions, restores...), purges...), errors.Join(admissionErr, restoreErr, purgeErr)
}

func (s *Store) recoverAdmissions() ([]RecoveryResult, error) {
	root, err := s.openRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	journals, err := admissionOpenChild(root, ".admissions")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer journals.Close()
	entries, err := journals.file.ReadDir(maxAdmissionRecovery + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(entries) > maxAdmissionRecovery {
		return nil, errors.New("quarantine recovery queue exceeds its bounded limit")
	}
	var results []RecoveryResult
	var failures error
	for _, entry := range entries {
		id := entry.Name()
		r := RecoveryResult{QuarantineID: id, Status: "needs_attention"}
		if !ordinaryQuarantineID(id) || !entry.IsDir() {
			r.Detail = "unknown recovery entry retained"
			results = append(results, r)
			failures = errors.Join(failures, errors.New(r.Detail))
			continue
		}
		unlock, err := root.lockItem(id)
		if errors.Is(err, ErrQuarantineBusy) {
			continue
		}
		if err != nil {
			failures = errors.Join(failures, err)
			continue
		}
		op, e := admissionOpenChild(journals, id)
		if e == nil {
			m, readErr := readAdmissionManifest(op, id)
			if readErr != nil {
				if os.IsNotExist(readErr) {
					files, listingErr := op.file.ReadDir(1)
					if (listingErr == nil || errors.Is(listingErr, io.EOF)) && len(files) == 0 {
						e = cleanupAdmissionOperation(journals, op, id)
						r.Status = "aborted"
					} else {
						e = readErr
					}
				} else {
					e = readErr
				}
			} else {
				r.OriginalPath = m.Record.OriginalPath
				r.Status, e = s.recoverAdmission(root, journals, op, m)
			}
			_ = op.Close()
		}
		unlock()
		if e != nil {
			r.Status = "needs_attention"
			r.Detail = e.Error()
			failures = errors.Join(failures, fmt.Errorf("quarantine recovery %s: %w", id, e))
		}
		results = append(results, r)
	}
	return results, failures
}

func (s *Store) recoverAdmission(root, journals, op *restoreDirectory, m admissionManifest) (string, error) {
	id := m.Record.QuarantineID
	_, found, err := s.cache.GetQuarantine(id)
	if err != nil {
		return "", err
	}
	if found {
		if err = s.cache.CommitQuarantine(m.Record); err != nil {
			return "", err
		}
		return "committed", cleanupAdmissionOperation(journals, op, id)
	}
	parent, err := platformRestoreDirectory(m.Parent)
	if err != nil {
		return "", err
	}
	defer parent.Close()
	parentID, err := admissionFileID(parent.file)
	if err != nil || parentID != m.ParentID {
		return "", errors.New("original parent directory identity changed; recovery objects retained")
	}
	capture, err := openAdmissionRoot(m.Record.SourceCaptureDir, false)
	if os.IsNotExist(err) && (m.Phase == "preparing" || m.Phase == "ready") {
		return "aborted", cleanupAdmissionOperation(journals, op, id)
	}
	if err != nil {
		return "", err
	}
	defer capture.Close()
	captureID, err := admissionFileID(capture.file)
	if err != nil || m.Record.SourceCaptureDirID != "" && captureID != m.Record.SourceCaptureDirID {
		return "", errors.New("capture directory identity changed; retained")
	}
	input, err := admissionOpenSource(capture, "source")
	if os.IsNotExist(err) && (m.Phase == "preparing" || m.Phase == "ready") {
		_ = capture.Close()
		if err = admissionRemoveDirectory(parent, filepath.Base(m.Record.SourceCaptureDir)); err != nil {
			return "", err
		}
		if err = admissionSyncDirectory(parent); err != nil {
			return "", err
		}
		return "aborted", cleanupAdmissionOperation(journals, op, id)
	}
	if err != nil {
		// A raced symlink/directory is not an authenticated source. Restore its
		// name without following it, or retain it if that name is now occupied.
		if m.Phase == "ready" {
			return rollbackAdmission(journals, op, parent, capture, nil, m, err)
		}
		return "", err
	}
	defer input.Close()
	inputID, err := admissionFileID(input)
	if err != nil || inputID != m.Record.SourceFileID {
		return rollbackAdmission(journals, op, parent, capture, input, m, fmt.Errorf("%w: captured replacement", ErrAdmissionChanged))
	}
	expected, err := restoreDigest(m.Record.SHA256)
	if err != nil {
		return "", err
	}
	if m.Phase == "preparing" {
		return "", errors.New("unexpected capture before durable ready state; retained")
	}
	if m.Phase == "ready" {
		info, e := restoreInputInfo(input)
		if e == nil {
			e = verifiedRestoreHash(input, info, expected)
		}
		if e != nil {
			// A crash after securing a verified source but before recording
			// "claimed" can leave ready metadata. Restore only this proven
			// original object's security before returning changed bytes.
			if securityErr := admissionRestoreSourceSecurity(input, m.Record.OriginalSecurity); securityErr != nil {
				return "", errors.Join(e, securityErr)
			}
			return rollbackAdmission(journals, op, parent, capture, input, m, e)
		}
		if e = admissionProtectCaptured(input); e != nil {
			return "", e
		}
		if e = s.syncAdmissionSource(input); e != nil {
			return "", fmt.Errorf("captured source synchronization failed: %w", e)
		}
		m.Phase = "claimed"
		if err = writeAdmissionManifest(op, m); err != nil {
			return "", err
		}
	}
	stored, err := admissionReadFile(root, id+".quarantined")
	if os.IsNotExist(err) {
		stored, err = admissionReadFile(op, "snapshot")
		if err != nil {
			return "", err
		}
		info, e := restoreInputInfo(stored)
		storedID, idErr := admissionFileID(stored)
		if e == nil && idErr == nil && storedID == m.SnapshotID {
			e = verifiedRestoreHash(stored, info, expected)
		} else {
			e = errors.Join(e, idErr, errors.New("snapshot identity differs"))
		}
		_ = stored.Close()
		if e != nil {
			return "", e
		}
		if err = admissionRename(op, "snapshot", root, id+".quarantined", false); err != nil {
			return "", err
		}
		if err = admissionSyncDirectory(root); err != nil {
			return "", err
		}
		stored, err = admissionReadFile(root, id+".quarantined")
	}
	if err != nil {
		return "", err
	}
	info, e := restoreInputInfo(stored)
	storedID, idErr := admissionFileID(stored)
	if e == nil && idErr == nil && storedID == m.SnapshotID {
		e = verifiedRestoreHash(stored, info, expected)
	} else {
		e = errors.Join(e, idErr, errors.New("published snapshot identity differs"))
	}
	_ = stored.Close()
	if e != nil {
		return "", e
	}
	m.Phase = "published"
	if err = writeAdmissionManifest(op, m); err != nil {
		return "", err
	}
	if err = s.cache.CommitQuarantine(m.Record); err != nil {
		return "", err
	}
	_ = input.Chmod(0600)
	return "committed", cleanupAdmissionOperation(journals, op, id)
}

func rollbackAdmission(journals, op, parent, capture *restoreDirectory, input *os.File, m admissionManifest, reason error) (string, error) {
	// Reuse the captured object's handle on Windows: its mandatory sharing
	// deliberately prevents reopening DELETE access while recovery owns it.
	var err error
	if input != nil {
		err = admissionRollback(input, capture, "source", parent, m.Leaf)
	} else {
		err = admissionRename(capture, "source", parent, m.Leaf, false)
	}
	if err != nil {
		return "", fmt.Errorf("%w; captured entry retained at %q (rollback failed: %v)", reason, m.Record.SourceCaptureDir, err)
	}
	if err := admissionSyncDirectory(parent); err != nil {
		return "", err
	}
	_ = capture.Close()
	if err := admissionRemoveDirectory(parent, filepath.Base(m.Record.SourceCaptureDir)); err != nil {
		return "", err
	}
	return "aborted", cleanupAdmissionOperation(journals, op, m.Record.QuarantineID)
}
