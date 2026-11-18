package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func freshExecutionFixture(t *testing.T) (*Store, string, string, string, FreshRestoreApproval, []byte) {
	t.Helper()
	s, id, original, stored, hold := freshInspectionFixture(t)
	destination := filepath.Join(filepath.Dir(original), "independently-approved-new-output")
	r, err := InspectFreshRestore(s.dir, id, hold.HoldDecisionID, destination)
	if err != nil {
		t.Fatal(err)
	}
	a, err := ApproveFreshRestore(s.dir, id, uuid.NewString(), r, "approve a new independent restore")
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return s, id, destination, stored, a, body
}

func executeFreshFixture(s *Store, body []byte, fault func(string) error) (FreshRestoreResult, error) {
	return executeFreshRestore(s.dir, body, "execute reviewed new restore", func(_ *freshRestorePermit, _ restoreManifest) (*cache.Cache, func() error, error) {
		return s.cache, func() error { return nil }, nil
	}, fault)
}

func TestFreshExecutionPublishesNewOutputAndPreservesOriginalEvidence(t *testing.T) {
	s, id, destination, stored, a, body := freshExecutionFixture(t)
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
	if err != nil {
		t.Fatal(err)
	}
	oldStage := filepath.Join(string(old.ParentPath), old.stageName(), "snapshot")
	oldArchive := filepath.Join(s.dir, ".reviews")
	beforeOld := inspectionTree(t, oldArchive)
	result, err := executeFreshFixture(s, body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "committed" || result.OperationID != a.OperationID || result.QuarantineID != id || string(result.Destination) != destination || result.OutputID == "" || result.OutputID == old.OutputID || result.OutputID == a.Inspection.Snapshot.FileID || result.OutputID == old.Record.SourceFileID {
		t.Fatal("fresh result adopted an old identity", result)
	}
	requirePurgePayload(t, destination, "payload")
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, oldStage, "payload")
	requirePurgePayload(t, filepath.Join(old.Record.SourceCaptureDir, "source"), "payload")
	if !reflect.DeepEqual(beforeOld, inspectionTree(t, oldArchive)) {
		t.Fatal("fresh execution changed original cancellation evidence")
	}
	if _, err = os.Lstat(old.Record.OriginalPath); !os.IsNotExist(err) {
		t.Fatal("fresh execution published the old destination", err)
	}
	if _, err = os.Lstat(filepath.Join(s.dir, ".restores", id)); !os.IsNotExist(err) {
		t.Fatal("completed fresh operation was not archived", err)
	}
	archive := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "restore")
	if _, err = os.Lstat(filepath.Join(archive, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID)); !os.IsNotExist(err) {
		t.Fatal("fresh private stage was not retired", err)
	}
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil || !found || !rec.Restored || rec.Purged || rec.OriginalPath != old.Record.OriginalPath {
		t.Fatal("fresh historical receipt is incorrect", rec, err)
	}
	if _, found, err = s.cache.QuarantineRestoreStageByIdentity(old.StageID); err != nil || !found {
		t.Fatal("fresh cleanup retired old-stage ownership", err)
	}
	if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("fresh success released normal recovery", err)
	}
	if err = s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("fresh success released purge", err)
	}
	beforeRetry := inspectionTree(t, filepath.Dir(s.dir))
	again, err := executeFreshFixture(s, body, nil)
	if err != nil || !reflect.DeepEqual(result, again) {
		t.Fatal("lost-return retry changed fresh result", again, err)
	}
	if !reflect.DeepEqual(beforeRetry, inspectionTree(t, filepath.Dir(s.dir))) {
		t.Fatal("completed fresh retry changed evidence")
	}
}

func TestFreshExecutionRejectsStaleAuthorityBeforeDatabase(t *testing.T) {
	for _, changed := range []string{"saved-root", "saved-actor", "saved-operation", "snapshot-bytes", "snapshot-copy", "existing-destination", "hold-copy", "cancellation-copy", "unknown-original-history", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, destination, stored, a, body := freshExecutionFixture(t)
			copyFile := func(path string) {
				t.Helper()
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
			}
			reason := "execute reviewed new restore"
			var hold RecoveryHold
			if err := json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
				t.Fatal(err)
			}
			switch changed {
			case "saved-root":
				a.RootID += "-different"
				body, _ = json.Marshal(a)
			case "saved-actor":
				a.ActorID += "-different"
				body, _ = json.Marshal(a)
			case "saved-operation":
				a.OperationID = uuid.NewString()
				body, _ = json.Marshal(a)
			case "snapshot-bytes":
				if e := os.WriteFile(stored, []byte("changed snapshot"), 0600); e != nil {
					t.Fatal(e)
				}
			case "snapshot-copy":
				copyFile(stored)
			case "existing-destination":
				if e := os.WriteFile(destination, []byte("payload"), 0600); e != nil {
					t.Fatal(e)
				}
			case "hold-copy":
				copyFile(filepath.Join(s.dir, ".purges", id))
			case "cancellation-copy":
				copyFile(filepath.Join(s.dir, ".reviews", hold.DecisionID, "manifest.json"))
			case "unknown-original-history":
				if e := os.WriteFile(filepath.Join(s.dir, ".reviews", hold.DecisionID, "unknown"), []byte("retain"), 0600); e != nil {
					t.Fatal(e)
				}
			case "reason":
				reason = " \t"
			}
			opened := false
			before := inspectionTree(t, filepath.Dir(s.dir))
			_, err := executeFreshRestore(s.dir, body, reason, func(_ *freshRestorePermit, _ restoreManifest) (*cache.Cache, func() error, error) {
				opened = true
				return s.cache, nil, nil
			}, nil)
			if err == nil || opened {
				t.Fatal("stale approval reached metadata storage", opened, err)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("stale execution mutated evidence")
			}
		})
	}
}

func TestFreshExecutionInterruptionRequiresExactProof(t *testing.T) {
	for _, phase := range []string{"operation-created", "stage-created", "stage-intent", "ownership-created", "ownership-written", "ownership-sealed", "ticket-prepared", "ticket-published", "stage-indexed", "output-created-filesystem", "output-identified", "copied", "security-restored", "verified", "publication-intent", "published-filesystem", "published", "before-database-commit", "database-committed", "committed", "stage-delete-intent", "ticket-retired", "stage-removed-filesystem", "before-archive", "archived"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, _, body := freshExecutionFixture(t)
			reached := false
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-"+phase {
					reached = true
					return errors.New("owned fresh execution interruption")
				}
				return nil
			})
			if err == nil || !reached {
				t.Fatal("fresh checkpoint was not reached", phase, err)
			}
			requirePurgePayload(t, stored, "payload")
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("interruption released ordinary recovery", err)
			}
			_, err = executeFreshFixture(s, body, nil)
			refuse := phase == "operation-created" || phase == "stage-created" || phase == "stage-intent" || phase == "ownership-created" || phase == "output-created-filesystem" || phase == "publication-intent"
			if (err != nil) != refuse {
				t.Fatal("fresh retry adopted unknown evidence or refused identified operation", phase, err)
			}
			rec, found, e := s.cache.GetQuarantine(id)
			if e != nil || !found || rec.Purged || rec.Restored == refuse {
				t.Fatal("fresh retry claimed wrong history", phase, rec, e)
			}
			if !refuse {
				requirePurgePayload(t, destination, "payload")
			} else if _, e = os.Lstat(destination); !os.IsNotExist(e) {
				t.Fatal("unproven fresh retry published output", phase, e)
			}
			requirePurgePayload(t, stored, "payload")
		})
	}
}

func TestFreshExecutionTicketProtectsRenamedStageWithoutDatabaseHint(t *testing.T) {
	s, id, destination, stored, _, body := freshExecutionFixture(t)
	_, err := executeFreshFixture(s, body, func(at string) error {
		if at == "fresh-execution-verified" {
			return errors.New("owned staged output retained")
		}
		return nil
	})
	if err == nil {
		t.Fatal("verified checkpoint not reached")
	}
	m := restoreFixtureManifest(t, s, id)
	if err = s.cache.ForgetQuarantineRestoreStage(s.restoreStageLocator(m)); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(filepath.Dir(destination), m.stageName())
	renamed := filepath.Join(filepath.Dir(destination), "renamed-owned-fresh-stage")
	if err = os.Rename(stage, renamed); err != nil {
		t.Fatal(err)
	}
	foreign, err := New(filepath.Join(filepath.Dir(s.dir), "another-store"), s.cache)
	if err != nil {
		t.Fatal(err)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	if qid, e := foreign.Quarantine(filepath.Join(renamed, "snapshot"), payloadHash("payload"), "Safe.Fresh.Stage"); qid != "" || !errors.Is(e, ErrAlreadyQuarantined) {
		t.Fatal("fresh owner ticket did not protect renamed stage", qid, e)
	}
	if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
		t.Fatal("fresh stage refusal mutated evidence")
	}
	if _, err = executeFreshFixture(s, body, nil); err == nil {
		t.Fatal("execution followed an unapproved renamed stage path")
	}
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, filepath.Join(renamed, "snapshot"), "payload")
}

func TestFreshOwnerMarkerAloneCannotPreventOrdinaryAdmission(t *testing.T) {
	s, id, destination, _, a, _ := freshExecutionFixture(t)
	parent, err := platformRestoreDirectory(filepath.Dir(destination))
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	ordinary, err := admissionMakeDirectory(parent, "ordinary-private-marker-control")
	if err != nil {
		t.Fatal(err)
	}
	dirID, err := admissionFileID(ordinary.file)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := admissionCreateFile(ordinary, freshTicketName)
	if err != nil {
		t.Fatal(err)
	}
	markerID, err := admissionFileID(marker)
	if err != nil {
		t.Fatal(err)
	}
	// Even a self-consistent marker in a private directory cannot invent an
	// immutable ownership record binding that actual directory to approval.
	ticket := freshOwnerTicket{Version: 1, RootPath: []byte(s.dir), RootID: s.rootID, QuarantineID: id, OperationID: a.OperationID, StageID: dirID, FileID: markerID, OwnershipID: "unproven-owned-file", OwnershipSHA256: payloadHash("unproven ownership")}
	_, body, err := freshJSONHash(ticket)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = marker.Write(body); err != nil {
		t.Fatal(err)
	}
	marker.Close()
	ordinary.Close()
	file := filepath.Join(parent.path, "ordinary-private-marker-control", "ordinary-input")
	if err = os.WriteFile(file, []byte("ordinary input marker"), 0600); err != nil {
		t.Fatal(err)
	}
	qid, err := s.Quarantine(file, payloadHash("ordinary input marker"), "Safe.Unproven.Owner.Marker")
	if err != nil || qid == "" || qid == id {
		t.Fatal("unproven owner marker prevented normal response", qid, err)
	}
	if _, err = os.Lstat(file); !os.IsNotExist(err) {
		t.Fatal("ordinary source was not admitted", err)
	}
	requirePurgePayload(t, filepath.Join(s.dir, qid+".quarantined"), "ordinary input marker")
}
