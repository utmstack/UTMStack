package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFreshExecutionPreservesChangedRawAndUsesIndependentSnapshot(t *testing.T) {
	s, id, destination, stored, a, body := freshExecutionFixture(t)
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
	if err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(old.Record.SourceCaptureDir, "source")
	if err = os.WriteFile(raw, []byte("retained changed raw bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = executeFreshFixture(s, body, nil); err != nil {
		t.Fatal(err)
	}
	requirePurgePayload(t, destination, "payload")
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, raw, "retained changed raw bytes")
}

func TestFreshExecutionUnknownStageChildRemainsHeldAndRetained(t *testing.T) {
	s, id, destination, stored, _, body := freshExecutionFixture(t)
	_, err := executeFreshFixture(s, body, func(at string) error {
		if at == "fresh-execution-verified" {
			return errors.New("owned verified interruption")
		}
		return nil
	})
	if err == nil {
		t.Fatal("verified checkpoint not reached")
	}
	m := restoreFixtureManifest(t, s, id)
	stage := filepath.Join(filepath.Dir(destination), m.stageName())
	unknown := filepath.Join(stage, "unknown-retained-child")
	if err = os.WriteFile(unknown, []byte("unknown retained bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = executeFreshFixture(s, body, nil); err == nil {
		t.Fatal("fresh execution ignored an unknown child")
	}
	requirePurgePayload(t, unknown, "unknown retained bytes")
	requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
	requirePurgePayload(t, stored, "payload")
	rec, found, e := s.cache.GetQuarantine(id)
	if e != nil || !found || rec.Restored || rec.Purged {
		t.Fatal("unknown stage child became completion", rec, e)
	}
}

func TestFreshExecutionCacheFlagIsNotPublicationAuthority(t *testing.T) {
	s, id, destination, stored, _, body := freshExecutionFixture(t)
	if err := s.cache.MarkRestored(id); err != nil {
		t.Fatal(err)
	}
	if _, err := executeFreshFixture(s, body, nil); err == nil {
		t.Fatal("fresh restore adopted an unrelated restored flag")
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatal("unrelated flag caused fresh publication", err)
	}
	if _, err := os.Lstat(filepath.Join(s.dir, ".restores", id)); !os.IsNotExist(err) {
		t.Fatal("stale cache state created a fresh operation", err)
	}
	requirePurgePayload(t, stored, "payload")
}

func TestFreshStageTicketDoesNotTrustRetiredJournalOrCopiedIdentity(t *testing.T) {
	for _, changed := range []string{"retired-journal", "copied-marker", "edited-owner"} {
		t.Run(changed, func(t *testing.T) {
			s, id, destination, stored, a, body := freshExecutionFixture(t)
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-verified" {
					return errors.New("owned verified interruption")
				}
				return nil
			})
			if err == nil {
				t.Fatal("verified checkpoint not reached")
			}
			m := restoreFixtureManifest(t, s, id)
			stagePath := filepath.Join(filepath.Dir(destination), m.stageName())
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			queue, item, decision, err := openFreshApproval(root, id, a.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			defer queue.Close()
			defer item.Close()
			defer decision.Close()
			switch changed {
			case "retired-journal":
				journals, op, e := openRestoreOperation(root, id)
				if e != nil {
					t.Fatal(e)
				}
				m.Phase, m.StageState = "committed", "deleted"
				e = writeRestoreManifest(op, m)
				op.Close()
				journals.Close()
				if e != nil {
					t.Fatal(e)
				}
			case "copied-marker":
				path := filepath.Join(stagePath, freshTicketName)
				raw, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.Rename(path, path+".retained"); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(path, raw, 0600); e != nil {
					t.Fatal(e)
				}
			case "edited-owner":
				if e := os.WriteFile(filepath.Join(decision.path, "ownership.json"), []byte("invalid retained ownership"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			parent, err := openAdmissionRoot(stagePath, false)
			if err != nil {
				t.Fatal(err)
			}
			stageID, err := admissionFileID(parent.file)
			if err != nil {
				t.Fatal(err)
			}
			proof := managedFreshRestoreStage(parent, stageID)
			parent.Close()
			if errors.Is(proof, ErrAlreadyQuarantined) {
				t.Fatal("changed ticket/journal granted fresh ownership", changed, proof)
			}
			if changed == "retired-journal" && !errors.Is(proof, ErrRestoreIntegrity) {
				t.Fatal("retired journal discarded live ticket evidence", proof)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(stagePath, "snapshot"), "payload")
		})
	}
}
