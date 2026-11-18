package quarantine

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func (s *Store) recoverPurges() ([]RecoveryResult, error) {
	root, err := s.openRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	journals, err := admissionOpenChild(root, ".purges")
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
		return nil, errors.New("purge recovery queue exceeds its bounded limit")
	}
	var results []RecoveryResult
	var failures error
	for _, entry := range entries {
		id := entry.Name()
		result := RecoveryResult{QuarantineID: id, Operation: "purge", Status: "needs_attention"}
		if ordinaryQuarantineID(id) && !entry.IsDir() {
			unlock, lockErr := root.lockItem(id)
			if errors.Is(lockErr, ErrQuarantineBusy) {
				continue
			}
			if lockErr == nil {
				hold, readErr := readRecoveryHold(root, journals, id)
				if readErr == nil {
					result.Operation, result.Status = "review", "operator_held"
					result.Detail = "retained by decision " + hold.DecisionID
					// Archived cancellation operations no longer appear in the
					// active restore queue. Rebuild their bounded lookup hints from
					// the still-validated fence binding, just as for live holds.
					if m, decodeErr := decodeRestoreManifest(hold.Original.Manifest, id); decodeErr == nil {
						result.OriginalPath = m.Record.OriginalPath
						if m.StageState == "present" {
							if trackErr := s.cache.TrackQuarantineRestoreStage(s.restoreStageLocator(m)); trackErr != nil {
								result.Status, result.Detail = "needs_attention", trackErr.Error()
								failures = errors.Join(failures, trackErr)
							}
						}
					}
					unlock()
					results = append(results, result)
					continue
				}
				unlock()
			}
		}
		if !ordinaryQuarantineID(id) || !entry.IsDir() {
			result.Detail = "unknown purge recovery entry retained"
			results = append(results, result)
			failures = errors.Join(failures, errors.New(result.Detail))
			continue
		}
		unlock, e := root.lockItem(id)
		if errors.Is(e, ErrQuarantineBusy) {
			continue
		}
		if e != nil {
			failures = errors.Join(failures, e)
			continue
		}
		op, e := admissionOpenChild(journals, id)
		if e == nil {
			m, readErr := readPurgeManifest(op, id)
			if os.IsNotExist(readErr) {
				// An empty directory contains no approved destructive work.
				// Cleanup can also have removed the completed manifest last.
				files, listingErr := op.file.ReadDir(1)
				if (listingErr == nil || errors.Is(listingErr, io.EOF)) && len(files) == 0 {
					e = cleanupPurgeOperation(journals, op, id, nil)
					result.Status = "cleaned"
				} else {
					e = readErr
				}
			} else if readErr != nil {
				e = readErr
			} else {
				result.OriginalPath = m.Record.OriginalPath
				e = s.resumePurge(root, journals, op, &m)
				result.Status = "purged"
			}
			op.Close()
		}
		unlock()
		if e != nil {
			result.Status = "needs_attention"
			result.Detail = e.Error()
			failures = errors.Join(failures, fmt.Errorf("purge recovery %s: %w", id, e))
		}
		results = append(results, result)
	}
	return results, failures
}
