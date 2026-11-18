package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReviewCancelOwnershipSurvivesIndependentPreparations(t *testing.T) {
	for _, preparation := range []string{"empty-live-slot", "malformed-live-journal", "copied-live-journal", "unknown-history-child"} {
		t.Run(preparation, func(t *testing.T) {
			s, id, original, _, a := reviewReleaseFixture(t, "publication-intent")
			old := restoreFixtureManifest(t, s, id)
			stage := filepath.Join(filepath.Dir(original), old.stageName())
			if _, err := CancelRecovery(s.dir, id, a, "cancel original and preserve ownership"); err != nil {
				t.Fatal(err)
			}
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			if preparation == "unknown-history-child" {
				if err = os.WriteFile(filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "unknown-future-evidence"), []byte("retain without interpreting"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				journals, err := admissionOpenChild(root, ".restores")
				if err != nil {
					t.Fatal(err)
				}
				op, err := admissionMakeDirectory(journals, id)
				if err != nil {
					t.Fatal(err)
				}
				if preparation == "malformed-live-journal" {
					if err = os.WriteFile(filepath.Join(op.path, "manifest.json"), []byte("unidentified preparation"), 0600); err != nil {
						t.Fatal(err)
					}
				} else if preparation == "copied-live-journal" {
					if err = writeRestoreManifest(op, old); err != nil {
						t.Fatal(err)
					}
				}
				op.Close()
				journals.Close()
			}
			root.Close()
			renamed := filepath.Join(filepath.Dir(stage), "retained-renamed-original-stage")
			if err = os.Rename(stage, renamed); err != nil {
				t.Fatal(err)
			}
			if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(old)); err != nil {
				t.Fatal(err)
			}
			// Recovery can refuse the independent preparation. It must still
			// recover the exact archived original's locator before admission.
			recovered, err := New(s.dir, s.cache)
			if err != nil {
				t.Fatal(err)
			}
			if _, found, err := s.cache.QuarantineRestoreStageByIdentity(old.StageID); err != nil || !found {
				t.Fatal("independent preparation erased archived ownership", err)
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if qid, err := recovered.Quarantine(filepath.Join(renamed, "snapshot"), payloadHash("payload"), "Safe.Cancelled.Original.Stage"); qid != "" || !errors.Is(err, ErrAlreadyQuarantined) {
				t.Fatal("independent preparation allowed archived-stage admission", qid, err)
			}
			if err = recovered.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("independent preparation released restore hold", err)
			}
			if err = recovered.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("independent preparation released purge hold", err)
			}
			if _, err = CancelRecovery(s.dir, id, a, "cancel original and preserve ownership"); err == nil {
				t.Fatal("ownership proof authorized cancellation lifecycle retry")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("ownership refusal changed retained evidence")
			}
		})
	}
}
