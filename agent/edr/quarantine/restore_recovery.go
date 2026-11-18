package quarantine

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func (s *Store) recoverRestores() ([]RecoveryResult, error) {
	root, err := s.openRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	journals, err := admissionOpenChild(root, ".restores")
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
		return nil, errors.New("restore recovery queue exceeds its bounded limit")
	}
	var results []RecoveryResult
	var failures error
	for _, entry := range entries {
		id := entry.Name()
		result := RecoveryResult{QuarantineID: id, Operation: "restore", Status: "needs_attention"}
		if ordinaryQuarantineID(id) && !entry.IsDir() && entry.Type().IsRegular() {
			unlock, e := root.lockItem(id)
			if errors.Is(e, ErrQuarantineBusy) {
				continue
			}
			if e == nil {
				m, history, found, readErr := openRetirementProof(root, id)
				if readErr == nil && found {
					archived, proofErr := retirementCore(root, history, m)
					if proofErr == nil {
						result.Operation, result.Status = "retirement", "retired"
						if !archived {
							result.Status = "retirement_pending"
						}
						result.Detail = "ordinary replay and deletion disabled by decision " + m.DecisionID
						history.Close()
						unlock()
						results = append(results, result)
						continue
					}
					history.Close()
					readErr = proofErr
				}
				e = readErr
				unlock()
			}
			if e != nil {
				result.Detail = e.Error()
				results = append(results, result)
				failures = errors.Join(failures, e)
				continue
			}
		}
		if !ordinaryQuarantineID(id) || !entry.IsDir() {
			result.Detail = "unknown restore recovery entry retained"
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
		hold, held, holdErr := recoveryHold(root, id)
		if holdErr != nil || held {
			if holdErr == nil {
				// Reconstruct legacy lookup hints even while the operator has
				// deliberately paused normal recovery. Malformed bytes remain
				// retained evidence, never a source of ownership claims.
				if m, decodeErr := decodeRestoreManifest(hold.Original.Manifest, id); decodeErr == nil {
					result.OriginalPath = m.Record.OriginalPath
					if m.StageState == "present" {
						holdErr = s.cache.TrackQuarantineRestoreStage(s.restoreStageLocator(m))
					}
				}
				result.Status = "operator_held"
				result.Detail = "retained by decision " + hold.DecisionID
			}
			if holdErr != nil {
				result.Status = "needs_attention"
				result.Detail = holdErr.Error()
				failures = errors.Join(failures, holdErr)
			}
			unlock()
			results = append(results, result)
			continue
		}
		op, e := admissionOpenChild(journals, id)
		if e == nil {
			m, readErr := readRestoreManifest(op, id)
			if os.IsNotExist(readErr) {
				files, listErr := op.file.ReadDir(1)
				if (listErr == nil || errors.Is(listErr, io.EOF)) && len(files) == 0 {
					e = cleanupPurgeOperation(journals, op, id, nil)
					result.Status = "cleaned"
				} else {
					e = readErr
				}
			} else if readErr != nil {
				e = readErr
			} else {
				result.OriginalPath = m.Record.OriginalPath
				e = s.resumeRestore(root, journals, op, &m)
				result.Status = "restored"
				if m.Phase == "aborted" {
					result.Status = "aborted"
				}
			}
			op.Close()
		}
		unlock()
		if e != nil {
			result.Status = "needs_attention"
			result.Detail = e.Error()
			failures = errors.Join(failures, fmt.Errorf("restore recovery %s: %w", id, e))
		}
		results = append(results, result)
	}
	return results, failures
}
