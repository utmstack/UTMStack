package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestFreshCancellationArchivesExactDirectoryAndRetainsEvidence(t *testing.T) {
	for _, phase := range []string{"operation-created", "ownership-created", "verified", "publication-intent", "published-filesystem", "database-committed"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, a, body := freshExecutionFixture(t)
			reached := false
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-"+phase {
					reached = true
					return errors.New("owned cancellation fixture checkpoint")
				}
				return nil
			})
			if err == nil || !reached {
				t.Fatal("fresh checkpoint not reached", err)
			}
			live := filepath.Join(s.dir, ".restores", id)
			if err = os.WriteFile(filepath.Join(live, "unexplained-owned-child"), []byte("preserve unknown evidence"), 0600); err != nil {
				t.Fatal(err)
			}
			beforeOperation := inspectionTree(t, live)
			beforeOld := inspectionTree(t, filepath.Join(s.dir, ".reviews"))
			beforeFence, err := os.ReadFile(filepath.Join(s.dir, ".purges", id))
			if err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID)
			var beforeStage map[string]string
			if _, e := os.Stat(stage); e == nil {
				beforeStage = inspectionTree(t, stage)
			} else if !os.IsNotExist(e) {
				t.Fatal(e)
			}
			beforeRecord, _, err := s.cache.GetQuarantine(id)
			if err != nil {
				t.Fatal(err)
			}
			beforeTree := inspectionTree(t, filepath.Dir(s.dir))
			review, err := InspectFreshCancellation(s.dir, id, a.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(beforeTree, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("read-only fresh cancellation inspection mutated storage")
			}
			receipt, err := CancelFreshRestore(s.dir, id, a.OperationID, review, "retain incomplete fresh operation independently")
			if err != nil {
				t.Fatal(err)
			}
			archive := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "restore")
			afterOperation := inspectionTree(t, archive)
			normalized := map[string]string{}
			for path, value := range afterOperation {
				normalized[live+strings.TrimPrefix(path, archive)] = value
			}
			if !reflect.DeepEqual(beforeOperation, normalized) {
				t.Fatal("fresh cancellation changed archived bytes or children")
			}
			f, err := os.Open(archive)
			if err != nil {
				t.Fatal(err)
			}
			directoryID, e := admissionFileID(f)
			f.Close()
			if e != nil || directoryID != review.OperationDirectoryID {
				t.Fatal("fresh archive is not the same operation", e)
			}
			if _, err = os.Lstat(live); !os.IsNotExist(err) {
				t.Fatal("fresh active operation was not moved", err)
			}
			afterFence, err := os.ReadFile(filepath.Join(s.dir, ".purges", id))
			if err != nil || string(afterFence) != string(beforeFence) {
				t.Fatal("fresh cancellation changed original fence", err)
			}
			if !reflect.DeepEqual(beforeOld, inspectionTree(t, filepath.Join(s.dir, ".reviews"))) {
				t.Fatal("fresh cancellation changed original history")
			}
			if beforeStage != nil && !reflect.DeepEqual(beforeStage, inspectionTree(t, stage)) {
				t.Fatal("fresh cancellation changed external stage")
			}
			afterRecord, _, err := s.cache.GetQuarantine(id)
			if err != nil {
				t.Fatal(err)
			}
			b1, e1 := cache.EncodeQuarantinePurgeRecord(beforeRecord)
			b2, e2 := cache.EncodeQuarantinePurgeRecord(afterRecord)
			if e1 != nil || e2 != nil || string(b1) != string(b2) {
				t.Fatal("fresh cancellation changed database state", e1, e2)
			}
			requirePurgePayload(t, stored, "payload")
			if phase == "published-filesystem" || phase == "database-committed" {
				requirePurgePayload(t, destination, "payload")
			} else if _, err = os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatal("cancellation published a destination", err)
			}
			beforeRetry := inspectionTree(t, filepath.Dir(s.dir))
			again, err := CancelFreshRestore(s.dir, id, a.OperationID, review, receipt.Reason)
			if err != nil || !reflect.DeepEqual(receipt, again) || !reflect.DeepEqual(beforeRetry, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("fresh cancellation lost-return retry changed evidence", err)
			}
			if _, err = executeFreshFixture(s, body, nil); err == nil {
				t.Fatal("cancelled fresh approval resumed execution")
			}
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("fresh cancellation released original restore", err)
			}
			if err = s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("fresh cancellation released purge", err)
			}
		})
	}
}

func TestFreshCancellationAllowsNewIndependentOperationAndProtectsOldStage(t *testing.T) {
	s, id, destination, stored, a, body := freshExecutionFixture(t)
	_, err := executeFreshFixture(s, body, func(at string) error {
		if at == "fresh-execution-verified" {
			return errors.New("owned verified checkpoint")
		}
		return nil
	})
	if err == nil {
		t.Fatal("verified checkpoint not reached")
	}
	r, err := InspectFreshCancellation(s.dir, id, a.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = CancelFreshRestore(s.dir, id, a.OperationID, r, "retain previous fresh operation"); err != nil {
		t.Fatal(err)
	}
	var hold RecoveryHold
	if err = json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID)
	if _, err = InspectFreshRestore(s.dir, id, hold.DecisionID, filepath.Join(stage, "must-not-publish-inside-retained-stage")); err == nil {
		t.Fatal("cancelled fresh stage became a new publication parent")
	}
	nextDestination := filepath.Join(filepath.Dir(destination), "next-independent-destination")
	inspection, err := InspectFreshRestore(s.dir, id, hold.DecisionID, nextDestination)
	if err != nil {
		t.Fatal(err)
	}
	next, err := ApproveFreshRestore(s.dir, id, uuid.NewString(), inspection, "approve an independent subsequent fresh restore")
	if err != nil {
		t.Fatal(err)
	}
	nextBody, err := json.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = executeFreshFixture(s, nextBody, nil); err != nil {
		t.Fatal(err)
	}
	requirePurgePayload(t, nextDestination, "payload")
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
	if _, err = s.Quarantine(filepath.Join(stage, "snapshot"), payloadHash("payload"), "Safe.Cancelled.Fresh.Stage"); !errors.Is(err, ErrAlreadyQuarantined) {
		t.Fatal("cancelled fresh stage lost bounded ownership", err)
	}
	if _, err = CancelFreshRestore(s.dir, id, a.OperationID, r, "retain previous fresh operation"); err != nil {
		t.Fatal("historical cancellation retry was invalidated by later execution", err)
	}
}

func TestFreshCancellationRejectsStaleReviewWithoutMutation(t *testing.T) {
	for _, changed := range []string{"root", "approval-file", "approval-hash", "operation-directory", "journal-file", "journal-hash", "owner-file", "owner-hash", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, _, _, a, body := freshExecutionFixture(t)
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-verified" {
					return errors.New("owned verified checkpoint")
				}
				return nil
			})
			if err == nil {
				t.Fatal("checkpoint not reached")
			}
			r, err := InspectFreshCancellation(s.dir, id, a.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			reason := "retain exact reviewed fresh operation"
			switch changed {
			case "root":
				r.RootID += "-changed"
			case "approval-file":
				r.Approval.ManifestID += "-changed"
			case "approval-hash":
				r.Approval.SHA256 = strings.Repeat("0", 64)
			case "operation-directory":
				r.OperationDirectoryID += "-changed"
			case "journal-file":
				r.Journal.ManifestID += "-changed"
			case "journal-hash":
				r.Journal.SHA256 = strings.Repeat("0", 64)
			case "owner-file":
				r.Ownership.ManifestID += "-changed"
			case "owner-hash":
				r.Ownership.SHA256 = strings.Repeat("0", 64)
			case "reason":
				reason = ""
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err = CancelFreshRestore(s.dir, id, a.OperationID, r, reason); err == nil {
				t.Fatal("stale fresh cancellation accepted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("stale fresh cancellation mutated evidence")
			}
		})
	}
}

func TestFreshCancellationInterruptionRetainsExactDecision(t *testing.T) {
	for _, phase := range []string{"receipt-created", "receipt-written", "approved", "before-move", "archived", "synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, _, stored, a, body := freshExecutionFixture(t)
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-publication-intent" {
					return errors.New("owned publication intent")
				}
				return nil
			})
			if err == nil {
				t.Fatal("checkpoint not reached")
			}
			r, err := InspectFreshCancellation(s.dir, id, a.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			reached := false
			_, err = cancelFreshRestore(s.dir, id, a.OperationID, r, "retain interrupted fresh operation", func(at string) error {
				if at == "fresh-cancellation-"+phase {
					reached = true
					return errors.New("owned cancellation interruption")
				}
				return nil
			})
			if err == nil || !reached {
				t.Fatal("fresh cancellation boundary not reached", err)
			}
			_, err = CancelFreshRestore(s.dir, id, a.OperationID, r, "retain interrupted fresh operation")
			if (err != nil) != (phase == "receipt-created") {
				t.Fatal("fresh cancellation adopted a partial receipt or refused complete exact retry", err)
			}
			requirePurgePayload(t, stored, "payload")
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("interrupted cancellation released restore", err)
			}
		})
	}
}
