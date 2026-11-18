package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReviewCancelArchivesSameOperationAndKeepsFence(t *testing.T) {
	for _, phase := range []string{"verified", "publication-intent"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, a := reviewReleaseFixture(t, phase)
			manifest := restoreFixtureManifest(t, s, id)
			stage := filepath.Join(filepath.Dir(original), manifest.stageName(), "snapshot")
			unknown := filepath.Join(s.dir, ".restores", id, "unknown-owned-evidence")
			if err := os.WriteFile(unknown, []byte("retained unknown bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			m, err := CancelRecovery(s.dir, id, a, "cancel original operation while retaining ambiguous evidence")
			if err != nil {
				t.Fatal(err)
			}
			after, err := InspectRecovery(s.dir, id)
			if err != nil || after.Journals[1].Status != "absent" || !reflect.DeepEqual(before.Journals[2], after.Journals[2]) {
				t.Fatal("cancellation changed the fence or retained an active operation", after, err)
			}
			archive := filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "restore")
			raw, err := os.ReadFile(filepath.Join(archive, "manifest.json"))
			if err != nil || !bytes.Equal(raw, before.Journals[1].Manifest) {
				t.Fatal("cancellation changed original journal bytes", err)
			}
			f, err := os.Open(archive)
			if err != nil {
				t.Fatal(err)
			}
			archiveID, err := admissionFileID(f)
			f.Close()
			if err != nil || archiveID != a.Original.DirectoryID {
				t.Fatal("archive is not the same operation directory", archiveID, err)
			}
			if raw, err = os.ReadFile(filepath.Join(archive, "unknown-owned-evidence")); err != nil || string(raw) != "retained unknown bytes" {
				t.Fatal("unknown original child was not preserved", err)
			}
			if again, err := CancelRecovery(s.dir, id, a, m.Reason); err != nil || !reflect.DeepEqual(m, again) {
				t.Fatal("cancellation lost-return retry changed the decision", again, err)
			}
			if _, err := ReleaseRecovery(s.dir, id, a, "resume original recovery"); err == nil {
				t.Fatal("ordinary release adopted a cancellation decision")
			}
			recovered, err := New(s.dir, s.cache)
			if err != nil {
				t.Fatal(err)
			}
			if err := recovered.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("new restore bypassed cancellation fence", err)
			}
			if err := recovered.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("purge bypassed cancellation fence", err)
			}
			beforeAdmission := inspectionTree(t, filepath.Dir(s.dir))
			if qid, err := recovered.Quarantine(stage, payloadHash("payload"), "Safe.Archived.Restore.Stage"); !errors.Is(err, ErrAlreadyQuarantined) || qid != "" {
				t.Fatal("archived restore stage lost bounded identity proof", qid, err)
			}
			if !reflect.DeepEqual(beforeAdmission, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("stage refusal changed retained evidence")
			}
			requirePurgePayload(t, stage, "payload")
			requirePurgePayload(t, stored, "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("cancellation published the destination", err)
			}
			rec, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found || rec.Restored || rec.Purged {
				t.Fatal("cancellation claimed historical completion", rec, err)
			}
		})
	}
}

func TestReviewCancelRejectsStaleApprovalBeforeMutation(t *testing.T) {
	for _, changed := range []string{"root", "original-directory", "original-file", "original-hash", "decision", "hold-directory", "hold-file", "hold-hash", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
			reason := "cancel and retain"
			switch changed {
			case "root":
				a.Original.RootID += "-changed"
			case "original-directory":
				a.Original.DirectoryID += "-changed"
			case "original-file":
				a.Original.ManifestID += "-changed"
			case "original-hash":
				a.Original.SHA256 = payloadHash("different")
			case "decision":
				a.HoldDecisionID = "00000000-0000-0000-0000-000000000001"
			case "hold-directory":
				a.HoldDirectoryID += "-changed"
			case "hold-file":
				a.HoldFileID += "-changed"
			case "hold-hash":
				a.HoldSHA256 = payloadHash("different")
			case "reason":
				reason = " \t"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := CancelRecovery(s.dir, id, a, reason); err == nil {
				t.Fatal("stale cancellation accepted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("rejected cancellation mutated evidence")
			}
		})
	}
}

func TestReviewCancelInterruptionPreservesFenceAndRetriesExactDecision(t *testing.T) {
	for _, phase := range []string{"cancel-directory-created", "cancel-receipt-written", "cancel-approved", "cancel-before-move", "cancel-archived", "cancel-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, _, stored, a := reviewReleaseFixture(t, "publication-intent")
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			_, err = cancelRecovery(s.dir, id, a, "cancel and retain", func(at string) error {
				if at == phase {
					return errors.New("owned cancellation interruption")
				}
				return nil
			})
			if err == nil {
				t.Fatal("interruption was not reached")
			}
			after, err := InspectRecovery(s.dir, id)
			if err != nil || !reflect.DeepEqual(before.Journals[2], after.Journals[2]) {
				t.Fatal("interrupted cancellation changed the active fence", err)
			}
			_, err = CancelRecovery(s.dir, id, a, "cancel and retain")
			if (err != nil) != (phase == "cancel-directory-created") {
				t.Fatal("retry adopted partial history or refused complete approval", err)
			}
			if err := s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("interruption opened normal restore", err)
			}
			requirePurgePayload(t, stored, "payload")
		})
	}
}

func TestReviewCancelRejectsChangedHistoryAndArchive(t *testing.T) {
	for _, change := range []string{"receipt-copy", "original-copy", "original-bytes", "hold-copy", "unknown-history", "live-impostor", "reason"} {
		t.Run(change, func(t *testing.T) {
			s, id, _, stored, a := reviewReleaseFixture(t, "publication-intent")
			m, err := CancelRecovery(s.dir, id, a, "retain cancelled operation")
			if err != nil {
				t.Fatal(err)
			}
			history := filepath.Join(s.dir, ".reviews", a.HoldDecisionID)
			copyFile := func(path string) {
				t.Helper()
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(path, path+".retained"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch change {
			case "receipt-copy":
				copyFile(filepath.Join(history, "manifest.json"))
			case "original-copy":
				copyFile(filepath.Join(history, "restore", "manifest.json"))
			case "original-bytes":
				if err := os.WriteFile(filepath.Join(history, "restore", "manifest.json"), []byte("changed original evidence"), 0600); err != nil {
					t.Fatal(err)
				}
			case "hold-copy":
				copyFile(filepath.Join(s.dir, ".purges", id))
			case "unknown-history":
				if err := os.WriteFile(filepath.Join(history, "unrecognized"), []byte("retain"), 0600); err != nil {
					t.Fatal(err)
				}
			case "live-impostor":
				if err := os.Mkdir(filepath.Join(s.dir, ".restores", id), 0700); err != nil {
					t.Fatal(err)
				}
			case "reason":
				m.Reason += " changed"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := CancelRecovery(s.dir, id, a, m.Reason); err == nil {
				t.Fatal("changed cancellation adopted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("refused cancellation changed evidence")
			}
			requirePurgePayload(t, stored, "payload")
		})
	}
}

func TestReviewCancelRechecksReceiptBytesBeforeArchive(t *testing.T) {
	for _, phase := range []string{"cancel-approved", "cancel-before-move"} {
		t.Run(phase, func(t *testing.T) {
			s, id, _, _, a := reviewReleaseFixture(t, "publication-intent")
			_, err := cancelRecovery(s.dir, id, a, "approved cancellation reason", func(at string) error {
				if at != phase {
					return nil
				}
				root, err := s.openRoot()
				if err != nil {
					t.Fatal(err)
				}
				defer root.Close()
				history, err := admissionOpenChild(root, ".reviews")
				if err != nil {
					t.Fatal(err)
				}
				defer history.Close()
				op, err := admissionOpenChild(history, a.HoldDecisionID)
				if err != nil {
					t.Fatal(err)
				}
				defer op.Close()
				// Accommodate the creator's existing DELETE/write access on Windows.
				f, err := openRecoveryReleaseReceipt(op)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				var m RecoveryCancellation
				if err := json.NewDecoder(f).Decode(&m); err != nil {
					t.Fatal(err)
				}
				m.Reason = "unapproved changed reason"
				raw, _ := json.Marshal(m)
				// Modify the same file identity: an identity-only recheck is insufficient.
				if err := f.Truncate(0); err != nil {
					t.Fatal(err)
				}
				if _, err := f.Seek(0, 0); err != nil {
					t.Fatal(err)
				}
				if _, err := f.Write(raw); err != nil {
					t.Fatal(err)
				}
				return nil
			})
			if err == nil {
				t.Fatal("cancellation moved the original after receipt bytes changed")
			}
			if _, err := os.Lstat(filepath.Join(s.dir, ".restores", id, "manifest.json")); err != nil {
				t.Fatal("changed receipt moved the original operation", err)
			}
		})
	}
}

func TestReviewCancelRebuildsArchivedStageLookup(t *testing.T) {
	for _, renamed := range []bool{false, true} {
		t.Run(map[bool]string{false: "original-name", true: "renamed-stage"}[renamed], func(t *testing.T) {
			s, id, original, _, a := reviewReleaseFixture(t, "publication-intent")
			m := restoreFixtureManifest(t, s, id)
			stage := filepath.Join(filepath.Dir(original), m.stageName())
			if _, err := CancelRecovery(s.dir, id, a, "cancel and preserve stage"); err != nil {
				t.Fatal(err)
			}
			if renamed {
				moved := filepath.Join(filepath.Dir(original), "renamed-cancelled-stage")
				if err := os.Rename(stage, moved); err != nil {
					t.Fatal(err)
				}
				stage = moved
			}
			if err := s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(m)); err != nil {
				t.Fatal(err)
			}
			recovered, err := New(s.dir, s.cache)
			if err != nil {
				t.Fatal(err)
			}
			if _, found, err := s.cache.QuarantineRestoreStageByIdentity(m.StageID); err != nil || !found {
				t.Fatal("archived stage lookup not rebuilt", err)
			}
			output := filepath.Join(stage, "snapshot")
			before := inspectionTree(t, filepath.Dir(s.dir))
			if qid, err := recovered.Quarantine(output, payloadHash("payload"), "Safe.Archived.Legacy.Stage"); qid != "" || !errors.Is(err, ErrAlreadyQuarantined) {
				t.Fatal("archived legacy stage was re-admitted", qid, err)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("archived stage refusal changed evidence")
			}
		})
	}
}
