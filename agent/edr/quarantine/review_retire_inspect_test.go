package quarantine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func retirementInspectionFixture(t *testing.T, state string) (*Store, string, string, string, FreshRestoreApproval) {
	t.Helper()
	s, id, _, _, a, body := freshExecutionFixture(t)
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
	if err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(string(old.ParentPath), old.stageName())
	var hold RecoveryHold
	if err = json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
		t.Fatal(err)
	}
	switch state {
	case "approval-only":
	case "committed":
		if _, err = executeFreshFixture(s, body, nil); err != nil {
			t.Fatal(err)
		}
	case "cancelled", "cancelled-malformed", "active":
		_, err = executeFreshFixture(s, body, func(at string) error {
			if at == "fresh-execution-publication-intent" {
				return errors.New("retain independently observed intent")
			}
			return nil
		})
		if err == nil {
			t.Fatal("fresh intent did not stop")
		}
		if state == "cancelled" || state == "cancelled-malformed" {
			if state == "cancelled-malformed" {
				if err = os.WriteFile(filepath.Join(s.dir, ".restores", id, "manifest.json"), []byte("{retain incomplete journal"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(filepath.Join(s.dir, ".restores", id, "unknown-retirement-evidence"), []byte("preserve unknown child"), 0600); err != nil {
				t.Fatal(err)
			}
			r, e := InspectFreshCancellation(s.dir, id, a.OperationID)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = CancelFreshRestore(s.dir, id, a.OperationID, r, "cancel before retirement review"); e != nil {
				t.Fatal(e)
			}
		}
	default:
		t.Fatal("unsupported test history")
	}
	return s, id, hold.DecisionID, stage, a
}

func TestRetirementInspectionReadsCompleteCaseWithoutDatabaseOrMutation(t *testing.T) {
	for _, state := range []string{"approval-only", "committed", "cancelled", "cancelled-malformed"} {
		t.Run(state, func(t *testing.T) {
			s, id, hold, stage, _ := retirementInspectionFixture(t, state)
			if err := s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			// A deliberately unusable database makes any accidental metadata open
			// observable; inspection is solely an existing-filesystem operation.
			db := filepath.Join(filepath.Dir(s.dir), "edr.db")
			if err := os.WriteFile(db, []byte("not a SQLite database"), 0600); err != nil {
				t.Fatal(err)
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
			if err != nil {
				t.Fatal(err)
			}
			expectedState := state
			if state == "cancelled-malformed" {
				expectedState = "cancelled"
			}
			if r.Version != 1 || r.Kind != "cancelled-retirement-inspection" || r.RootID == "" || r.RestoreQueueID == "" || r.HistoryDirectoryID == "" || r.HoldID != hold || len(r.Fresh) != 1 || r.Fresh[0].State != expectedState || r.Snapshot.FileID == r.Raw.FileID || r.OldStage.DirectoryID == "" {
				t.Fatal("incomplete retirement evidence", r)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("read-only retirement inspection changed evidence")
			}
			body, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			var decoded CancelledRetirementInspection
			if err = json.Unmarshal(body, &decoded); err != nil || !sameCancelledRetirementInspection(r, decoded, false) {
				t.Fatal("inspection changed across saved JSON", err)
			}
		})
	}
}

func TestRetirementInspectionFindsOnlyExplicitSameRenamedOldStage(t *testing.T) {
	s, id, hold, stage, _ := retirementInspectionFixture(t, "committed")
	r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(filepath.Dir(stage), "retained-stage-renamed-for-review")
	if err = os.Rename(stage, renamed); err != nil {
		t.Fatal(err)
	}
	if _, err = InspectCancelledRetirement(s.dir, id, hold, stage); err == nil {
		t.Fatal("missing literal stage was accepted")
	}
	again, err := InspectCancelledRetirement(s.dir, id, hold, renamed)
	if err != nil || again.OldStage.DirectoryID != r.OldStage.DirectoryID {
		t.Fatal("same renamed original stage was not independently identified", err)
	}
	copyPath := filepath.Join(filepath.Dir(stage), "different-private-directory")
	if err = os.Mkdir(copyPath, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = InspectCancelledRetirement(s.dir, id, hold, copyPath); err == nil {
		t.Fatal("another private directory inherited stage identity")
	}
}

func TestRetirementInspectionRefusesActiveOrChangedEvidenceWithoutMutation(t *testing.T) {
	for _, change := range []string{"active", "hold-copy", "snapshot-copy", "raw-copy", "owner-bytes", "archive-copy", "unknown-history", "unknown-approval-child", "original-journal", "wrong-hold"} {
		t.Run(change, func(t *testing.T) {
			state := "cancelled"
			if change == "active" {
				state = "active"
			}
			s, id, hold, stage, a := retirementInspectionFixture(t, state)
			copyFile := func(path string) {
				t.Helper()
				b, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.Rename(path, path+".retained"); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(path, b, 0600); e != nil {
					t.Fatal(e)
				}
			}
			op := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID)
			switch change {
			case "hold-copy":
				copyFile(filepath.Join(s.dir, ".purges", id))
			case "snapshot-copy":
				copyFile(filepath.Join(s.dir, id+".quarantined"))
			case "raw-copy":
				old, e := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
				if e != nil {
					t.Fatal(e)
				}
				copyFile(filepath.Join(old.Record.SourceCaptureDir, "source"))
			case "owner-bytes":
				if e := os.WriteFile(filepath.Join(op, "ownership.json"), []byte("{partial ownership"), 0600); e != nil {
					t.Fatal(e)
				}
			case "archive-copy":
				path := filepath.Join(op, "restore")
				if e := os.Rename(path, path+"-retained"); e != nil {
					t.Fatal(e)
				}
				if e := os.Mkdir(path, 0700); e != nil {
					t.Fatal(e)
				}
				b, e := os.ReadFile(filepath.Join(path+"-retained", "manifest.json"))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(path, "manifest.json"), b, 0600); e != nil {
					t.Fatal(e)
				}
			case "unknown-history":
				if e := os.Mkdir(filepath.Join(s.dir, ".fresh-restores", id, "unreviewed-history"), 0700); e != nil {
					t.Fatal(e)
				}
			case "unknown-approval-child":
				if e := os.WriteFile(filepath.Join(op, "unknown"), []byte("retain"), 0600); e != nil {
					t.Fatal(e)
				}
			case "original-journal":
				copyFile(filepath.Join(s.dir, ".reviews", hold, "restore", "manifest.json"))
			case "wrong-hold":
				hold = uuid.NewString()
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, e := InspectCancelledRetirement(s.dir, id, hold, stage); e == nil {
				t.Fatal("changed evidence was accepted", change)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("failed inspection mutated evidence", change)
			}
		})
	}
}

func TestRetirementInspectionIncludesAllFreshApprovalsInStableOrder(t *testing.T) {
	s, id, hold, stage, a := retirementInspectionFixture(t, "committed")
	for i := 0; i < 3; i++ {
		destination := filepath.Join(filepath.Dir(stage), fmt.Sprintf("independent-retirement-output-%d", i))
		r, err := InspectFreshRestore(s.dir, id, hold, destination)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = ApproveFreshRestore(s.dir, id, uuid.NewString(), r, "unused independent approval"); err != nil {
			t.Fatal(err)
		}
	}
	r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Fresh) != 4 {
		t.Fatal("fresh history was omitted", len(r.Fresh))
	}
	found := false
	for i, op := range r.Fresh {
		if i > 0 && r.Fresh[i-1].OperationID >= op.OperationID {
			t.Fatal("unstable history ordering")
		}
		if op.OperationID == a.OperationID {
			found = op.State == "committed"
		} else if op.State != "approval-only" {
			t.Fatal("wrong unused approval disposition")
		}
	}
	if !found {
		t.Fatal("committed history was lost")
	}
}

func TestRetirementInspectionRefusesFullLegacyQueueWithoutChangingIt(t *testing.T) {
	s, id, hold, stage, _ := retirementInspectionFixture(t, "approval-only")
	queue := filepath.Join(s.dir, ".restores")
	for i := 0; i < maxAdmissionRecovery; i++ {
		if e := os.WriteFile(filepath.Join(queue, fmt.Sprintf("occupied-%04d", i)), []byte("permanent retained entry"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	if _, err := InspectCancelledRetirement(s.dir, id, hold, stage); err == nil {
		t.Fatal("unbounded retirement compatibility queue accepted")
	}
	if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
		t.Fatal("capacity refusal reclaimed an existing marker")
	}
}

func TestRetirementInspectionDoesNotCreateMissingStorage(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "missing-store")
	before := inspectionTree(t, base)
	if _, err := InspectCancelledRetirement(path, "item", uuid.NewString(), base); err == nil {
		t.Fatal("missing storage accepted")
	}
	if !reflect.DeepEqual(before, inspectionTree(t, base)) {
		t.Fatal("missing store was created")
	}
}
