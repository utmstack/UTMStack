package quarantine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestAdmissionRestoreStageHintAloneDoesNotGrantOwnership(t *testing.T) {
	s, c, id, original, _ := restoreFixture(t)
	parent, err := openRestoreDirectory(filepath.Dir(original), false)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	stage, err := admissionMakeDirectory(parent, "private-unrecorded-stage")
	if err != nil {
		t.Fatal(err)
	}
	stageID, err := admissionFileID(stage.file)
	if err != nil {
		t.Fatal(err)
	}
	stage.Close()
	file := filepath.Join(parent.path, "private-unrecorded-stage", "foreign")
	if err = os.WriteFile(file, []byte("foreign marker"), 0600); err != nil {
		t.Fatal(err)
	}
	hint := cache.QuarantineRestoreStage{StageID: stageID, OperationID: uuid.NewString(), QuarantineID: id, RootID: s.rootID, RootPath: []byte(s.dir)}
	if err = c.TrackQuarantineRestoreStage(hint); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Quarantine(file, payloadHash("foreign marker"), "Safe.Stale.Hint"); got != "" || !errors.Is(err, ErrRestoreIntegrity) || errors.Is(err, ErrAlreadyQuarantined) {
		t.Fatal("hint granted unproven ownership", got, err)
	}
	requirePurgePayload(t, file, "foreign marker")
	if records, err := c.ListQuarantine(); err != nil || len(records) != 1 {
		t.Fatal("uncertain ownership changed records", err)
	}
}

func TestRestoreStageIndexActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"stage-index-intent", "stage-indexed", "stage-index-deleted"} {
		t.Run(phase, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			if err = c.Close(); err != nil {
				t.Fatal(err)
			}
			db := filepath.Join(filepath.Dir(s.dir), "edr.db")
			cmd := exec.Command(self, "-test.run=^TestRestoreCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_RESTORE_CRASH_WORKER=1", "OPENEDR_RESTORE_DB="+db, "OPENEDR_RESTORE_ROOT="+s.dir, "OPENEDR_RESTORE_ID="+id, "OPENEDR_RESTORE_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 97 {
				t.Fatalf("checkpoint did not terminate process: %v %s", err, output)
			}
			m := restoreFixtureManifest(t, s, id)
			reopened, err := cache.Open(db)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			_, indexed, err := reopened.QuarantineRestoreStageByIdentity(m.StageID)
			if err != nil || indexed != (phase == "stage-indexed") {
				t.Fatalf("unexpected durable index before recovery: %t %v", indexed, err)
			}
			recovered, err := New(s.dir, reopened)
			if err != nil {
				t.Fatal(err)
			}
			results := recovered.RecoveryStatus()
			if len(results) != 1 || results[0].Status == "needs_attention" {
				t.Fatalf("recovery failed: %+v", results)
			}
			if _, indexed, err = reopened.QuarantineRestoreStageByIdentity(m.StageID); err != nil || indexed {
				t.Fatal("recovery left stale index", err)
			}
			record, found, err := reopened.GetQuarantine(id)
			published := phase == "stage-index-deleted"
			if err != nil || !found || record.Restored != published {
				t.Fatalf("incorrect historical publication: %+v %v", record, err)
			}
			if published {
				requirePurgePayload(t, original, "payload")
			} else {
				requirePurgePayload(t, stored, "payload")
				if _, err = os.Lstat(original); !os.IsNotExist(err) {
					t.Fatal("unpublished crash restored original", err)
				}
			}
			requirePurgePayload(t, filepath.Join(record.SourceCaptureDir, "source"), "payload")
			if again, err := recovered.Recover(); err != nil || len(again) != 0 {
				t.Fatal("repeated recovery not empty", again, err)
			}
		})
	}
}

// Directory creation can be noticed before the restore has registered ownership
// or produced output. A request for the future snapshot must recheck ownership
// after opening it and before creating any admission journal or capture.
func TestAdmissionRechecksRestoreStageAfterOpeningSource(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	s.responseFault = func(phase string) error {
		if phase == "restore-stage-index-intent" {
			return errors.New("owned index registration interruption")
		}
		return nil
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("index registration checkpoint was not reached")
	}
	m := restoreFixtureManifest(t, s, id)
	stagePath := filepath.Join(filepath.Dir(original), m.stageName())
	file := filepath.Join(stagePath, "snapshot")
	called := false
	s.admissionFault = func(phase string) error {
		if phase != "parent-checked" {
			return nil
		}
		called = true
		if err := c.TrackQuarantineRestoreStage(s.restoreStageLocator(m)); err != nil {
			return err
		}
		return os.WriteFile(file, []byte("payload"), 0600)
	}
	got, err := s.Quarantine(file, payloadHash("payload"), "Safe.Registration.Race")
	if !called || got != "" || !errors.Is(err, ErrAlreadyQuarantined) {
		t.Fatalf("newly registered stage was re-admitted: hook=%t id=%q err=%v", called, got, err)
	}
	requirePurgePayload(t, file, "payload")
	requirePurgePayload(t, stored, "payload")
	if _, err := os.Lstat(original); !os.IsNotExist(err) {
		t.Fatal("original was published", err)
	}
	if records, err := c.ListQuarantine(); err != nil || len(records) != 1 {
		t.Fatal("race changed quarantine records", err)
	}
	entries, err := os.ReadDir(stagePath)
	if err != nil || len(entries) != 1 || entries[0].Name() != "snapshot" {
		t.Fatal("race created admission objects in restore stage", entries, err)
	}
}
