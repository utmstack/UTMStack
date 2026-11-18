package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAdmissionRecoversLegacyRestoreStageLookup(t *testing.T) {
	for _, kind := range []string{"retained", "renamed", "foreign-journal-entry"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			s.responseFault = func(phase string) error {
				if phase == "restore-publication-intent" {
					return errors.New("owned legacy publication interruption")
				}
				return nil
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("publication checkpoint was not reached")
			}
			m := restoreFixtureManifest(t, s, id)
			// An upgrade creates the lookup table, but version-1 journals produced by
			// older executables never inserted a row. The durable journal is unchanged.
			if err := c.ForgetQuarantineRestoreStage(s.restoreStageLocator(m)); err != nil {
				t.Fatal(err)
			}
			if _, found, err := c.QuarantineRestoreStageByIdentity(m.StageID); err != nil || found {
				t.Fatal("legacy fixture still has a hint", err)
			}
			stage := filepath.Join(filepath.Dir(original), m.stageName())
			if kind == "renamed" {
				moved := filepath.Join(filepath.Dir(original), "renamed-legacy-stage")
				if err := os.Rename(stage, moved); err != nil {
					t.Fatal(err)
				}
				stage = moved
			}
			if kind == "foreign-journal-entry" {
				if err := os.WriteFile(filepath.Join(s.dir, ".restores", id, "unidentified-marker"), []byte("preserve this unknown entry"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			recovered, err := New(s.dir, c)
			if err != nil {
				t.Fatal(err)
			}
			status := recovered.RecoveryStatus()
			if len(status) != 1 || status[0].Status != "needs_attention" {
				t.Fatalf("ambiguous legacy publication was cleared: %+v", status)
			}
			output := filepath.Join(stage, "snapshot")
			requirePurgePayload(t, output, "payload")
			before := inspectionTree(t, filepath.Dir(s.dir))
			other, err := recovered.Quarantine(output, payloadHash("payload"), "Safe.Legacy.Restore.Stage")
			if other != "" || !errors.Is(err, ErrAlreadyQuarantined) {
				t.Fatalf("legacy restore output was re-admitted: id=%q err=%v", other, err)
			}
			requirePurgePayload(t, output, "payload")
			requirePurgePayload(t, stored, "payload")
			if _, err = os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("ambiguous original was published", err)
			}
			if records, err := c.ListQuarantine(); err != nil || len(records) != 1 || records[0].Restored {
				t.Fatal("legacy history was changed", records, err)
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("legacy admission refusal changed retained evidence")
			}
		})
	}
}
