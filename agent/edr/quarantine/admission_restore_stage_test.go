package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAdmissionDoesNotRequarantineRecordedRestoreStage(t *testing.T) {
	for _, checkpoint := range []string{"verified", "publication-intent"} {
		t.Run(checkpoint, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			s.responseFault = func(phase string) error {
				if phase == "restore-"+checkpoint {
					return errors.New("owned interruption before publication")
				}
				return nil
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("checkpoint was not reached")
			}
			m := restoreFixtureManifest(t, s, id)
			stage := filepath.Join(filepath.Dir(original), m.stageName())
			output := filepath.Join(stage, "snapshot")
			requirePurgePayload(t, output, "payload")
			before := inspectionTree(t, filepath.Dir(s.dir))
			other, err := s.Quarantine(output, payloadHash("payload"), "Safe.Restore.Stage.Control")
			if !errors.Is(err, ErrAlreadyQuarantined) || other != "" {
				t.Fatalf("recorded restore output was re-admitted: id=%q err=%v", other, err)
			}
			requirePurgePayload(t, output, "payload")
			requirePurgePayload(t, stored, "payload")
			if records, err := c.ListQuarantine(); err != nil || len(records) != 1 {
				t.Fatalf("staged evidence created another quarantine: %+v %v", records, err)
			}
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("rejected re-admission published an original", err)
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatal("refused re-admission changed retained evidence or metadata")
			}
		})
	}
}

func TestAdmissionRestoreStageLookupWorksAcrossStoresAndRename(t *testing.T) {
	s, c, id, original, _ := restoreFixture(t)
	other, err := New(filepath.Join(filepath.Dir(s.dir), "other-quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	s.responseFault = func(phase string) error {
		if phase == "restore-verified" {
			return errors.New("owned retained stage")
		}
		return nil
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("checkpoint was not reached")
	}
	m := restoreFixtureManifest(t, s, id)
	stage := filepath.Join(filepath.Dir(original), m.stageName())
	moved := filepath.Join(filepath.Dir(original), "renamed-private-restore-stage")
	if err := os.Rename(stage, moved); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(moved, "snapshot")
	if _, err := other.Quarantine(output, payloadHash("payload"), "Safe.Renamed.Stage"); !errors.Is(err, ErrAlreadyQuarantined) {
		t.Fatal("renamed recorded stage was not recognized across stores", err)
	}
	// Its former name alone must not exempt a replacement directory.
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(stage, "foreign")
	if err := os.WriteFile(foreign, []byte("foreign marker"), 0600); err != nil {
		t.Fatal(err)
	}
	if foreignID, err := other.Quarantine(foreign, payloadHash("foreign marker"), "Safe.Replacement.Stage"); err != nil || foreignID == "" {
		t.Fatal("restore-looking replacement received an exclusion", foreignID, err)
	}
	requirePurgePayload(t, output, "payload")
}

func TestAdmissionRestoreLookingNameHasNoExclusion(t *testing.T) {
	s, _, path, dir := admissionFixture(t)
	_ = os.Remove(path)
	stage := filepath.Join(dir, ".edr-restore-unregistered")
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(stage, "marker")
	if err := os.WriteFile(file, []byte("foreign marker"), 0600); err != nil {
		t.Fatal(err)
	}
	if id, err := s.Quarantine(file, payloadHash("foreign marker"), "Safe.Unregistered.Stage"); err != nil || id == "" {
		t.Fatal("unregistered restore-looking directory bypassed admission", id, err)
	}
}
