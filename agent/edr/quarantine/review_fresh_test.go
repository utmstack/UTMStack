package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFreshApprovalRetainsFenceBackupsAndExactLostReturn(t *testing.T) {
	s, id, destination, stored, a := freshInspectionFixture(t)
	r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
	if err != nil {
		t.Fatal(err)
	}
	before, err := InspectRecovery(s.dir, id)
	if err != nil {
		t.Fatal(err)
	}
	operationID := uuid.NewString()
	m, err := ApproveFreshRestore(s.dir, id, operationID, r, "independently approve one new publication")
	if err != nil {
		t.Fatal(err)
	}
	if m.Kind != "fresh-restore-approval" || m.Version != 1 || m.OperationID != operationID || m.QuarantineID != id || m.RootID != s.rootID || m.ActorID == "" || m.FileID == "" || m.DirectoryID == "" || m.ItemID == "" || m.QueueID == "" || !sameFreshInspection(m.Inspection, r, false) {
		t.Fatal("incomplete fresh approval", m)
	}
	if again, err := ApproveFreshRestore(s.dir, id, operationID, r, m.Reason); err != nil || !reflect.DeepEqual(m, again) {
		t.Fatal("lost-return approval changed", again, err)
	}
	after, err := InspectRecovery(s.dir, id)
	if err != nil || !reflect.DeepEqual(before.Journals, after.Journals) {
		t.Fatal("approval changed original operation or hold", err)
	}
	if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("approval released normal restore", err)
	}
	if err = s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("approval released purge", err)
	}
	if _, err = os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatal("approval created a destination", err)
	}
	requirePurgePayload(t, stored, "payload")
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil || !found || rec.Restored || rec.Purged {
		t.Fatal("approval claimed a publication", rec, err)
	}
}

func TestFreshApprovalRejectsStaleSavedInputsBeforeCreation(t *testing.T) {
	for _, changed := range []string{"version", "kind", "quarantine", "observation", "root", "hold-uuid", "hold-hash", "cancellation-hash", "original-hash", "snapshot-id", "snapshot-hash", "snapshot-size", "record", "parent-id", "parent-path", "leaf", "old-operation-uuid", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, destination, _, a := freshInspectionFixture(t)
			r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
			if err != nil {
				t.Fatal(err)
			}
			operationID, reason := uuid.NewString(), "approve reviewed fresh output"
			switch changed {
			case "version":
				r.Version++
			case "kind":
				r.Kind = "different-action"
			case "quarantine":
				r.QuarantineID += "-different"
			case "observation":
				r.ObservedAt = time.Time{}
			case "root":
				r.RootID += "-changed"
			case "hold-uuid":
				var hold RecoveryHold
				if err = json.Unmarshal(r.Hold.Manifest, &hold); err != nil {
					t.Fatal(err)
				}
				hold.DecisionID = uuid.NewString()
				r.Hold.Manifest, _ = json.Marshal(hold)
			case "hold-hash":
				r.Hold.SHA256 = payloadHash("different hold")
			case "cancellation-hash":
				r.Cancellation.SHA256 = payloadHash("different cancellation")
			case "original-hash":
				r.Original.SHA256 = payloadHash("different original")
			case "snapshot-id":
				r.Snapshot.FileID += "-changed"
			case "snapshot-hash":
				r.Snapshot.SHA256 = payloadHash("different snapshot")
			case "snapshot-size":
				r.Snapshot.Size++
			case "record":
				r.RecordData = []byte("different record")
			case "parent-id":
				r.Destination.ParentID += "-changed"
			case "parent-path":
				r.Destination.ParentPath = []byte(filepath.Join(filepath.Dir(destination), "missing-parent"))
			case "leaf":
				r.Destination.Leaf = []byte("../different-leaf")
			case "old-operation-uuid":
				old, err := decodeRestoreManifest(r.Original.Manifest, id)
				if err != nil {
					t.Fatal(err)
				}
				operationID = old.OperationID
			case "reason":
				reason = " \t"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err = ApproveFreshRestore(s.dir, id, operationID, r, reason); err == nil {
				t.Fatal("stale fresh approval accepted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("stale approval created control storage or changed evidence")
			}
		})
	}
}

func TestFreshApprovalInterruptionKeepsEvidenceAndRefusesPartialAdoption(t *testing.T) {
	for _, phase := range []string{"fresh-approval-directory-created", "fresh-approval-written", "fresh-approval-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, a := freshInspectionFixture(t)
			r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
			if err != nil {
				t.Fatal(err)
			}
			operationID := uuid.NewString()
			_, err = approveFreshRestore(s.dir, id, operationID, r, "approve fresh while retaining original", func(at string) error {
				if at == phase {
					return errors.New("owned approval interruption")
				}
				return nil
			})
			if err == nil {
				t.Fatal("approval checkpoint not reached")
			}
			_, err = ApproveFreshRestore(s.dir, id, operationID, r, "approve fresh while retaining original")
			if (err != nil) != (phase == "fresh-approval-directory-created") {
				t.Fatal("retry adopted partial approval or refused complete receipt", err)
			}
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("approval interruption released the hold", err)
			}
			requirePurgePayload(t, stored, "payload")
			if _, err = os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatal("interrupted approval published output", err)
			}
		})
	}
}

func TestFreshApprovalRejectsChangedRetryIdentityAndReason(t *testing.T) {
	for _, changed := range []string{"reason", "actor", "inspection", "receipt-copy", "receipt-bytes"} {
		t.Run(changed, func(t *testing.T) {
			s, id, destination, _, a := freshInspectionFixture(t)
			r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
			if err != nil {
				t.Fatal(err)
			}
			operationID, reason := uuid.NewString(), "approve one reviewed fresh operation"
			m, err := ApproveFreshRestore(s.dir, id, operationID, r, reason)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(s.dir, ".fresh-restores", id, operationID, "manifest.json")
			switch changed {
			case "reason":
				reason += " different"
			case "actor":
				m.ActorID += "-different"
				body, _ := json.Marshal(m)
				if err = os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			case "inspection":
				r.ObservedAt = r.ObservedAt.Add(time.Second)
			case "receipt-copy":
				body, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.Rename(path, path+".retained"); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			case "receipt-bytes":
				if err = os.WriteFile(path, []byte("incomplete owned approval"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err = ApproveFreshRestore(s.dir, id, operationID, r, reason); err == nil {
				t.Fatal("changed approval retry accepted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("changed approval retry mutated evidence")
			}
		})
	}
}
