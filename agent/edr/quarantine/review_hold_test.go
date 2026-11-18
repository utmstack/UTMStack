package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func reviewHoldFixture(t *testing.T, phase string) (*Store, string, string, string, RecoveryApproval) {
	t.Helper()
	s, _, id, original, stored := restoreFixture(t)
	s.responseFault = func(at string) error {
		if at == "restore-"+phase {
			return errors.New("owned review interruption")
		}
		return nil
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("review boundary was not reached")
	}
	report, err := InspectRecovery(s.dir, id)
	if err != nil {
		t.Fatal(err)
	}
	var approval RecoveryApproval
	for _, entry := range report.Journals {
		if entry.Operation == "restore" {
			approval = RecoveryApproval{RootID: report.RootID, DirectoryID: entry.DirectoryID, ManifestID: entry.ManifestID, SHA256: entry.SHA256}
		}
	}
	if err := validRecoveryApproval(approval); err != nil {
		t.Fatal(err)
	}
	return s, id, original, stored, approval
}
func TestReviewHoldRetainsRestoreAcrossRecoveryAndRetention(t *testing.T) {
	for _, phase := range []string{"verified", "publication-intent"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, approval := reviewHoldFixture(t, phase)
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			decision, err := HoldRecovery(s.dir, id, approval, "retain pending independent identity review")
			if err != nil {
				t.Fatal(err)
			}
			if decision.DecisionID == "" || decision.FileID == "" || decision.ActorID == "" {
				t.Fatal("incomplete durable decision", decision)
			}
			again, err := HoldRecovery(s.dir, id, approval, decision.Reason)
			if err != nil || !reflect.DeepEqual(decision, again) {
				t.Fatal("lost-return retry changed decision", again, err)
			}
			current, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before.Journals[1], current.Journals[1]) {
				t.Fatal("hold replayed or edited original restore")
			}
			if current.Journals[2].EntryKind != "file" || current.Journals[2].ManifestID != decision.FileID {
				t.Fatal("inspection did not capture exact retention fence", current)
			}
			recovered, err := New(s.dir, s.cache)
			if err != nil {
				t.Fatal(err)
			}
			statuses := recovered.RecoveryStatus()
			if len(statuses) != 2 {
				t.Fatalf("missing held status: %+v", statuses)
			}
			for _, status := range statuses {
				if status.Status != "operator_held" {
					t.Fatalf("hold was replayed or treated as approved purge: %+v", statuses)
				}
			}
			if err = recovered.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("normal restore bypassed hold", err)
			}
			if err = recovered.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("manual purge bypassed hold", err)
			}
			if n, err := recovered.PurgeExpired(1, time.Now().Add(10*365*24*time.Hour)); err != nil || n != 0 {
				t.Fatal("retention bypassed hold", n, err)
			}
			rec, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found || rec.Restored || rec.Purged {
				t.Fatal("hold claimed historical completion", rec, err)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(rec.SourceCaptureDir, "source"), "payload")
			m := restoreFixtureManifest(t, s, id)
			requirePurgePayload(t, filepath.Join(filepath.Dir(original), m.stageName(), "snapshot"), "payload")
			if _, err = os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("hold published original", err)
			}
		})
	}
}
func TestReviewHoldRejectsStaleApprovalBeforeMutation(t *testing.T) {
	for _, field := range []string{"root", "directory", "manifest", "hash", "reason"} {
		t.Run(field, func(t *testing.T) {
			s, id, _, _, approval := reviewHoldFixture(t, "publication-intent")
			reason := "review requested"
			switch field {
			case "root":
				approval.RootID += "-changed"
			case "directory":
				approval.DirectoryID += "-changed"
			case "manifest":
				approval.ManifestID += "-changed"
			case "hash":
				approval.SHA256 = strings.Repeat("0", 64)
			case "reason":
				reason = strings.Repeat("x", 4097)
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := HoldRecovery(s.dir, id, approval, reason); err == nil {
				t.Fatal("stale approval accepted")
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("rejected review changed storage")
			}
		})
	}
}
func TestReviewHoldDoesNotReplaceExistingPurgeEntry(t *testing.T) {
	for _, kind := range []string{"directory", "file"} {
		t.Run(kind, func(t *testing.T) {
			s, id, _, _, approval := reviewHoldFixture(t, "publication-intent")
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			journals, err := admissionChild(root, ".purges")
			if err != nil {
				t.Fatal(err)
			}
			defer journals.Close()
			if kind == "directory" {
				op, err := admissionMakeDirectory(journals, id)
				if err != nil {
					t.Fatal(err)
				}
				op.Close()
			} else {
				f, err := admissionCreateFile(journals, id)
				if err != nil {
					t.Fatal(err)
				}
				f.WriteString("foreign marker")
				f.Close()
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err := HoldRecovery(s.dir, id, approval, "review requested"); err == nil {
				t.Fatal("foreign entry adopted")
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("foreign entry or other evidence changed")
			}
		})
	}
}
func TestReviewHoldCapturesMalformedJournalWithoutReplay(t *testing.T) {
	s, id, _, stored, _ := reviewHoldFixture(t, "publication-intent")
	bad := []byte("{ malformed raw evidence \xff")
	if err := os.WriteFile(filepath.Join(s.dir, ".restores", id, "manifest.json"), bad, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := InspectRecovery(s.dir, id)
	if err != nil {
		t.Fatal(err)
	}
	entry := report.Journals[1]
	approval := RecoveryApproval{RootID: report.RootID, DirectoryID: entry.DirectoryID, ManifestID: entry.ManifestID, SHA256: entry.SHA256}
	decision, err := HoldRecovery(s.dir, id, approval, "preserve malformed evidence")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decision.Original.Manifest, bad) {
		t.Fatal("malformed bytes were normalized")
	}
	recovered, err := New(s.dir, s.cache)
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range recovered.RecoveryStatus() {
		if status.Status != "operator_held" {
			t.Fatal("malformed held journal replayed", status)
		}
	}
	requirePurgePayload(t, stored, "payload")
}
func TestReviewHoldRejectsChangedBindingsAndPreservesFence(t *testing.T) {
	s, id, _, stored, approval := reviewHoldFixture(t, "publication-intent")
	if _, err := HoldRecovery(s.dir, id, approval, "retain for review"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.dir, ".restores", id, "manifest.json"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Purge(id); err == nil || errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("changed evidence silently granted a valid hold", err)
	}
	requirePurgePayload(t, stored, "payload")
	if _, err := os.Lstat(filepath.Join(s.dir, ".purges", id)); err != nil {
		t.Fatal("uncertain fence removed", err)
	}
}

func TestReviewHoldRejectsReplacementFenceWithIdenticalBytes(t *testing.T) {
	s, id, _, stored, approval := reviewHoldFixture(t, "publication-intent")
	if _, err := HoldRecovery(s.dir, id, approval, "retain for review"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.dir, ".purges", id)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, path+"-owned-moved"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Purge(id); err == nil || errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("content match adopted replacement decision", err)
	}
	requirePurgePayload(t, stored, "payload")
	current, err := os.ReadFile(path)
	if err != nil || !reflect.DeepEqual(current, body) {
		t.Fatal("replacement decision changed", err)
	}
}

func TestReviewHoldRespectsActiveItemLock(t *testing.T) {
	s, id, _, _, approval := reviewHoldFixture(t, "publication-intent")
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
	defer func() {
		if unlock != nil {
			unlock()
		}
	}()
	if _, err := HoldRecovery(s.dir, id, approval, "retain for review"); !errors.Is(err, ErrQuarantineBusy) {
		t.Fatal("review bypassed active mutation lock", err)
	}
	// Windows byte-range locks also forbid reading the locked range. Compare
	// evidence only after releasing our owned lock, while no mutator is active.
	unlock()
	unlock = nil
	if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
		t.Fatal("busy review changed evidence")
	}
}
