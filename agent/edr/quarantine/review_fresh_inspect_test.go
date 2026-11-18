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

func freshInspectionFixture(t *testing.T) (*Store, string, string, string, RecoveryReleaseApproval) {
	t.Helper()
	s, id, original, stored, a := reviewReleaseFixture(t, "publication-intent")
	if _, err := CancelRecovery(s.dir, id, a, "retain original case before an independent fresh review"); err != nil {
		t.Fatal(err)
	}
	return s, id, original, stored, a
}

func TestFreshInspectionPinsIndependentBackupWithoutMutation(t *testing.T) {
	s, id, destination, stored, a := freshInspectionFixture(t)
	// Close SQLite before observation: inspection must be independent of it.
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 1 || r.Kind != "fresh-restore-inspection" || r.QuarantineID != id || r.RootID != s.rootID || r.Snapshot.FileID == "" || r.Snapshot.Size != 7 || r.Snapshot.SHA256 != payloadHash("payload") || r.Cancellation.ManifestID == "" || r.Hold.ManifestID != a.HoldFileID || r.Original.ManifestID != a.Original.ManifestID || string(r.Destination.Leaf) != filepath.Base(destination) {
		t.Fatalf("incomplete inspection: %+v", r)
	}
	f, err := os.Open(stored)
	if err != nil {
		t.Fatal(err)
	}
	inputID, err := admissionFileID(f)
	f.Close()
	if err != nil || inputID != r.Snapshot.FileID {
		t.Fatal("snapshot identity was not observed independently", err)
	}
	body, err := json.Marshal(r)
	var saved FreshRestoreInspection
	if err != nil || json.Unmarshal(body, &saved) != nil || !bytes.Equal(saved.RecordData, r.RecordData) || !reflect.DeepEqual(saved.Destination, r.Destination) {
		t.Fatal("inspection lost byte-safe approval inputs", err)
	}
	if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
		t.Fatal("fresh inspection created state or changed retained evidence")
	}
}

func TestFreshInspectionRejectsChangedOrAmbiguousEvidence(t *testing.T) {
	for _, changed := range []string{"hold-uuid", "not-cancelled", "cancel-bytes", "cancel-copy", "hold-bytes", "hold-copy", "original-bytes", "original-copy", "snapshot-bytes", "snapshot-copy", "existing-output", "matching-output", "directory-output", "missing-parent", "storage-destination", "old-stage-destination", "renamed-stage-destination", "relative-destination"} {
		t.Run(changed, func(t *testing.T) {
			s, id, destination, stored, a := freshInspectionFixture(t)
			op := filepath.Join(s.dir, ".reviews", a.HoldDecisionID)
			old, err := os.ReadFile(filepath.Join(op, "restore", "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			m, err := decodeRestoreManifest(old, id)
			if err != nil {
				t.Fatal(err)
			}
			change := func(path string, copyOnly bool) {
				t.Helper()
				body, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if copyOnly {
					if err = os.Rename(path, path+"-retained"); err != nil {
						t.Fatal(err)
					}
				} else {
					body = []byte("changed owned evidence")
				}
				if err = os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch changed {
			case "hold-uuid":
				a.HoldDecisionID = "00000000-0000-0000-0000-000000000001"
			case "not-cancelled":
				if err = os.Rename(filepath.Join(op, "restore"), filepath.Join(s.dir, ".restores", id)); err != nil {
					t.Fatal(err)
				}
			case "cancel-bytes", "cancel-copy":
				change(filepath.Join(op, "manifest.json"), changed == "cancel-copy")
			case "hold-bytes", "hold-copy":
				change(filepath.Join(s.dir, ".purges", id), changed == "hold-copy")
			case "original-bytes", "original-copy":
				change(filepath.Join(op, "restore", "manifest.json"), changed == "original-copy")
			case "snapshot-bytes", "snapshot-copy":
				change(stored, changed == "snapshot-copy")
			case "existing-output", "matching-output":
				body := []byte("existing destination")
				if changed == "matching-output" {
					body = []byte("payload")
				}
				if err = os.WriteFile(destination, body, 0600); err != nil {
					t.Fatal(err)
				}
			case "directory-output":
				if err = os.Mkdir(destination, 0700); err != nil {
					t.Fatal(err)
				}
			case "missing-parent":
				destination = filepath.Join(filepath.Dir(destination), "missing-parent", "new-output")
			case "storage-destination":
				destination = filepath.Join(op, "new-output")
			case "old-stage-destination", "renamed-stage-destination":
				stage := filepath.Join(string(m.ParentPath), m.stageName())
				if changed == "renamed-stage-destination" {
					renamed := filepath.Join(filepath.Dir(stage), "ordinary-renamed-owned-stage")
					if err = os.Rename(stage, renamed); err != nil {
						t.Fatal(err)
					}
					stage = renamed
				}
				destination = filepath.Join(stage, "new-output")
			case "relative-destination":
				destination = "relative-output"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err = InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination); err == nil {
				t.Fatal("fresh inspection accepted changed or ambiguous evidence")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("failed fresh inspection modified evidence")
			}
		})
	}
}

func TestFreshInspectionDoesNotAuthorizeNormalRestoreOrPurge(t *testing.T) {
	s, id, destination, stored, a := freshInspectionFixture(t)
	if _, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("inspection bypassed restore fence", err)
	}
	if err := s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("inspection bypassed purge fence", err)
	}
	requirePurgePayload(t, stored, "payload")
}
