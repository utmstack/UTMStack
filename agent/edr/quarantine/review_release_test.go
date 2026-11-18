package quarantine

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func reviewReleaseFixture(t *testing.T, phase string) (*Store, string, string, string, RecoveryReleaseApproval) {
	t.Helper()
	s, id, original, stored, approval := reviewHoldFixture(t, phase)
	hold, err := HoldRecovery(s.dir, id, approval, "owned hold pending independent review")
	if err != nil {
		t.Fatal(err)
	}
	report, err := InspectRecovery(s.dir, id)
	if err != nil {
		t.Fatal(err)
	}
	fence := report.Journals[2]
	a := RecoveryReleaseApproval{Original: approval, HoldDecisionID: hold.DecisionID, HoldDirectoryID: fence.DirectoryID, HoldFileID: fence.ManifestID, HoldSHA256: fence.SHA256}
	return s, id, original, stored, a
}

func TestReviewReleasePreservesEvidenceAndDoesNotRunRecovery(t *testing.T) {
	for _, phase := range []string{"verified", "publication-intent"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, approval := reviewReleaseFixture(t, phase)
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			decision, err := ReleaseRecovery(s.dir, id, approval, "authorize normal recovery of the original operation")
			if err != nil {
				t.Fatal(err)
			}
			if decision.FileID == "" || decision.DirectoryID == "" || decision.ActorID == "" || decision.DecisionID == "" {
				t.Fatal("incomplete release decision", decision)
			}
			after, err := InspectRecovery(s.dir, id)
			if err != nil || !reflect.DeepEqual(before.Journals[1], after.Journals[1]) || after.Journals[2].Status != "absent" {
				t.Fatal("release replayed the original operation or left the active fence", after, err)
			}
			archive := filepath.Join(s.dir, ".reviews", approval.HoldDecisionID, "hold.json")
			raw, err := os.ReadFile(archive)
			if err != nil || !bytes.Equal(raw, before.Journals[2].Manifest) {
				t.Fatal("release did not retain exact fence evidence", err)
			}
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			f, err := os.Open(archive)
			if err != nil {
				t.Fatal(err)
			}
			archiveID, err := admissionFileID(f)
			f.Close()
			root.Close()
			if err != nil || archiveID != approval.HoldFileID {
				t.Fatal("release adopted a copy instead of moving the same fence", archiveID, err)
			}
			again, err := ReleaseRecovery(s.dir, id, approval, decision.Reason)
			if err != nil || !reflect.DeepEqual(decision, again) {
				t.Fatal("lost-return retry changed the decision", again, err)
			}
			rec, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found || rec.Restored || rec.Purged {
				t.Fatal("release claimed original completion", rec, err)
			}
			requirePurgePayload(t, stored, "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("release published the destination", err)
			}
			_, _ = s.Recover()
			if phase == "verified" {
				if _, err := os.Lstat(filepath.Join(s.dir, ".restores", id)); !os.IsNotExist(err) {
					t.Fatal("explicit later recovery did not retire the pre-publication operation", err)
				}
			} else {
				current, err := os.ReadFile(filepath.Join(s.dir, ".restores", id, "manifest.json"))
				if err != nil || !bytes.Equal(current, before.Journals[1].Manifest) {
					t.Fatal("release made ambiguous publication replayable", err)
				}
			}
			// Historical lookup must also work when normal recovery removed the
			// original journal. This authorizes no later operation for this ID.
			if again, err = ReleaseRecovery(s.dir, id, approval, decision.Reason); err != nil || !reflect.DeepEqual(decision, again) {
				t.Fatal("completed history required old live evidence", err)
			}
		})
	}
}

func TestReviewReleaseRejectsStaleApprovalWithoutHistoryMutation(t *testing.T) {
	for _, field := range []string{"root", "original-directory", "original-file", "original-hash", "decision", "fence-directory", "fence-file", "fence-hash", "reason"} {
		t.Run(field, func(t *testing.T) {
			s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
			reason := "authorize normal recovery"
			switch field {
			case "root":
				a.Original.RootID += "-changed"
			case "original-directory":
				a.Original.DirectoryID += "-changed"
			case "original-file":
				a.Original.ManifestID += "-changed"
			case "original-hash":
				a.Original.SHA256 = strings.Repeat("0", 64)
			case "decision":
				a.HoldDecisionID = "00000000-0000-0000-0000-000000000001"
			case "fence-directory":
				a.HoldDirectoryID += "-changed"
			case "fence-file":
				a.HoldFileID += "-changed"
			case "fence-hash":
				a.HoldSHA256 = strings.Repeat("0", 64)
			case "reason":
				reason = " \t"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := ReleaseRecovery(s.dir, id, a, reason); err == nil {
				t.Fatal("stale release accepted")
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("rejected release changed storage")
			}
		})
	}
}

func TestReviewReleaseRetainsUnknownPreparationAndChangedHistory(t *testing.T) {
	for _, changed := range []string{"bare-directory", "foreign-child", "receipt-replacement", "hold-replacement", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, _, stored, a := reviewReleaseFixture(t, "publication-intent")
			reason := "authorize normal recovery"
			op := filepath.Join(s.dir, ".reviews", a.HoldDecisionID)
			if changed == "bare-directory" {
				if err := os.MkdirAll(op, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				_, err := releaseRecovery(s.dir, id, a, reason, func(at string) error {
					if at == "release-approved" {
						return errors.New("owned approval interruption")
					}
					return nil
				})
				if err == nil {
					t.Fatal("interruption was not reached")
				}
				switch changed {
				case "foreign-child":
					if err := os.WriteFile(filepath.Join(op, "foreign"), []byte("unidentified"), 0600); err != nil {
						t.Fatal(err)
					}
				case "receipt-replacement", "hold-replacement":
					path := filepath.Join(op, "manifest.json")
					if changed == "hold-replacement" {
						path = filepath.Join(s.dir, ".purges", id)
					}
					body, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if err = os.Rename(path, path+"-original"); err != nil {
						t.Fatal(err)
					}
					if err = os.WriteFile(path, body, 0600); err != nil {
						t.Fatal(err)
					}
				case "reason":
					reason = "different approval"
				}
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := ReleaseRecovery(s.dir, id, a, reason); err == nil {
				t.Fatal("uncertain or changed history adopted")
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("uncertain history or fence changed")
			}
			requirePurgePayload(t, stored, "payload")
		})
	}
}

func TestReviewReleaseOldRetryCannotReleaseLaterHold(t *testing.T) {
	s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
	first, err := ReleaseRecovery(s.dir, id, a, "release first hold")
	if err != nil {
		t.Fatal(err)
	}
	later, err := HoldRecovery(s.dir, id, a.Original, "later independent hold")
	if err != nil {
		t.Fatal(err)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	again, err := ReleaseRecovery(s.dir, id, a, first.Reason)
	if err != nil || !reflect.DeepEqual(first, again) {
		t.Fatal("old release could not return its historical result", err)
	}
	if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
		t.Fatal("old release retry altered a later hold")
	}
	if err := s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("later hold was released", later, err)
	}
}

func TestReviewReleaseRejectsChangedArchiveWithIdenticalBytes(t *testing.T) {
	s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
	first, err := ReleaseRecovery(s.dir, id, a, "release reviewed hold")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "hold.json")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, path+"-original"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	if _, err := ReleaseRecovery(s.dir, id, a, first.Reason); err == nil {
		t.Fatal("replacement archive adopted by content hash")
	}
	if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
		t.Fatal("changed historical evidence modified")
	}
}

func TestReviewReleaseRevalidatesFenceBeforeMove(t *testing.T) {
	s, id, _, stored, a := reviewReleaseFixture(t, "publication-intent")
	_, err := releaseRecovery(s.dir, id, a, "release reviewed hold", func(at string) error {
		if at == "release-before-move" {
			return os.WriteFile(filepath.Join(s.dir, ".restores", id, "manifest.json"), []byte("changed original"), 0600)
		}
		return nil
	})
	if err == nil {
		t.Fatal("changed original allowed fence removal")
	}
	if _, err := os.Lstat(filepath.Join(s.dir, ".purges", id)); err != nil {
		t.Fatal("changed original lost its fence", err)
	}
	requirePurgePayload(t, stored, "payload")
}

func TestReviewReleaseRespectsActiveItemLock(t *testing.T) {
	s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
	root, err := s.openRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	before := inspectionTree(t, filepath.Dir(s.dir))
	unlock, err := root.lockItem(id)
	if err != nil {
		t.Fatal(err)
	}
	_, releaseErr := ReleaseRecovery(s.dir, id, a, "release reviewed hold")
	unlock()
	if !errors.Is(releaseErr, ErrQuarantineBusy) {
		t.Fatal("release bypassed active item lock", releaseErr)
	}
	if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
		t.Fatal("busy release changed evidence")
	}
}

func TestReviewReleaseRefusesFenceInBothLocations(t *testing.T) {
	s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
	_, err := releaseRecovery(s.dir, id, a, "release reviewed hold", func(at string) error {
		if at == "release-approved" {
			return errors.New("owned approval interruption")
		}
		return nil
	})
	if err == nil {
		t.Fatal("interruption was not reached")
	}
	// Even an identical file identity in history cannot prove a move when the
	// same fence still occupies its active name (for example, an extra hard link).
	if err := os.Link(filepath.Join(s.dir, ".purges", id), filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "hold.json")); err != nil {
		t.Fatal(err)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	if _, err := ReleaseRecovery(s.dir, id, a, "release reviewed hold"); err == nil {
		t.Fatal("two fence locations were accepted as a completed move")
	}
	if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
		t.Fatal("ambiguous fence locations were changed")
	}
}
