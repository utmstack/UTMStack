package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/release"
)

func TestQuarantineReleaseMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_RELEASE_MAIN_WORKER") != "1" {
		t.Skip("owned standalone release CLI worker")
	}
	root := os.Getenv("OPENEDR_RELEASE_MAIN_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = filepath.Join(root, "must-not-open-db")
	os.Args = []string{"utmstack_edr", "quarantine", "release-hold", os.Getenv("OPENEDR_RELEASE_MAIN_ID"), "--hold", os.Getenv("OPENEDR_RELEASE_MAIN_HOLD"), "--inspection", filepath.Join(root, "inspection.json"), "--reason", "authorize original recovery while retaining malformed evidence"}
	main()
	t.Fatal("release should exit via its standalone entry point")
}

func TestQuarantineReleaseMainPreservesEvidenceWithoutDatabaseOrLog(t *testing.T) {
	root, id := t.TempDir(), uuid.NewString()
	store := filepath.Join(root, "quarantine")
	c, err := cache.Open(filepath.Join(root, "fixture.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, newErr := quarantine.New(store, c)
	if err := c.Close(); err != nil || newErr != nil {
		t.Fatal(newErr, err)
	}
	parent := store
	for _, leaf := range []string{".restores", id} {
		private, err := release.MakePrivateTempDir(parent, "release-fixture-")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(private, filepath.Join(parent, leaf)); err != nil {
			t.Fatal(err)
		}
		parent = filepath.Join(parent, leaf)
	}
	body := []byte("{ malformed original release evidence \xff")
	journal := filepath.Join(parent, "manifest.json")
	if err := os.WriteFile(journal, body, 0600); err != nil {
		t.Fatal(err)
	}
	initial, err := quarantine.InspectRecovery(store, id)
	if err != nil {
		t.Fatal(err)
	}
	entry := initial.Journals[1]
	hold, err := quarantine.HoldRecovery(store, id, quarantine.RecoveryApproval{RootID: initial.RootID, DirectoryID: entry.DirectoryID, ManifestID: entry.ManifestID, SHA256: entry.SHA256}, "retain malformed fixture")
	if err != nil {
		t.Fatal(err)
	}
	report, err := quarantine.InspectRecovery(store, id)
	if err != nil {
		t.Fatal(err)
	}
	inspection, _ := json.Marshal(report)
	cfg, _ := json.Marshal(map[string]string{"quarantine_dir": store})
	for name, raw := range map[string][]byte{"edr.json": cfg, "inspection.json": inspection} {
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var first quarantine.RecoveryRelease
	for attempt := 0; attempt < 2; attempt++ {
		cmd := exec.Command(self, "-test.run=^TestQuarantineReleaseMainWorker$", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENEDR_RELEASE_MAIN_WORKER=1", "OPENEDR_RELEASE_MAIN_ROOT="+root, "OPENEDR_RELEASE_MAIN_ID="+id, "OPENEDR_RELEASE_MAIN_HOLD="+hold.DecisionID)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("release main failed: %v %s %s", err, &stdout, &stderr)
		}
		var decision quarantine.RecoveryRelease
		if err := json.Unmarshal(stdout.Bytes(), &decision); err != nil || stderr.Len() != 0 || !bytes.Equal(decision.Hold.Manifest, report.Journals[2].Manifest) {
			t.Fatalf("invalid release output: %v %s %s", err, &stdout, &stderr)
		}
		if attempt == 0 {
			first = decision
		} else if !reflect.DeepEqual(first, decision) {
			t.Fatal("release retry changed historical decision")
		}
	}
	for _, leaf := range []string{"must-not-open-db", "must-not-open-db-wal", "must-not-open-db-shm", "must-not-create-log"} {
		if _, err := os.Lstat(filepath.Join(root, leaf)); !os.IsNotExist(err) {
			t.Fatal("standalone release created runtime state", leaf, err)
		}
	}
	current, err := os.ReadFile(journal)
	if err != nil || !bytes.Equal(current, body) {
		t.Fatal("release changed malformed original evidence", err)
	}
	if _, err := os.Lstat(filepath.Join(store, ".purges", id)); !os.IsNotExist(err) {
		t.Fatal("release left the active fence", err)
	}
}

func TestQuarantineReleaseCLIRejectsInvalidInputBeforeOpeningStore(t *testing.T) {
	for _, args := range [][]string{nil, {"id"}, {"id", "--hold", "uuid", "--inspection", "missing", "--reason", "reason"}} {
		var out, diagnostic bytes.Buffer
		if code := runQuarantineRelease(args, &out, &diagnostic); code == 0 || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatalf("invalid release accepted: %v", args)
		}
	}
}
