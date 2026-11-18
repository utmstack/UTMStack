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

func TestQuarantineCancelMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_CANCEL_MAIN_WORKER") != "1" {
		t.Skip("owned standalone cancellation CLI worker")
	}
	root := os.Getenv("OPENEDR_CANCEL_MAIN_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = filepath.Join(root, "must-not-open-db")
	os.Args = []string{"utmstack_edr", "quarantine", "cancel-restore", os.Getenv("OPENEDR_CANCEL_MAIN_ID"), "--hold", os.Getenv("OPENEDR_CANCEL_MAIN_HOLD"), "--inspection", filepath.Join(root, "inspection.json"), "--reason", "cancel original operation while retaining malformed evidence"}
	main()
	t.Fatal("cancellation should exit via its standalone entry point")
}

func TestQuarantineCancelMainPreservesEvidenceWithoutDatabaseOrLog(t *testing.T) {
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
	var first quarantine.RecoveryCancellation
	for attempt := 0; attempt < 2; attempt++ {
		cmd := exec.Command(self, "-test.run=^TestQuarantineCancelMainWorker$", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENEDR_CANCEL_MAIN_WORKER=1", "OPENEDR_CANCEL_MAIN_ROOT="+root, "OPENEDR_CANCEL_MAIN_ID="+id, "OPENEDR_CANCEL_MAIN_HOLD="+hold.DecisionID)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("cancellation main failed: %v %s %s", err, &stdout, &stderr)
		}
		var decision quarantine.RecoveryCancellation
		if err := json.Unmarshal(stdout.Bytes(), &decision); err != nil || stderr.Len() != 0 || !bytes.Equal(decision.Hold.Manifest, report.Journals[2].Manifest) {
			t.Fatalf("invalid cancellation output: %v %s %s", err, &stdout, &stderr)
		}
		if attempt == 0 {
			first = decision
		} else if !reflect.DeepEqual(first, decision) {
			t.Fatal("cancellation retry changed historical decision")
		}
	}
	for _, leaf := range []string{"must-not-open-db", "must-not-open-db-wal", "must-not-open-db-shm", "must-not-create-log"} {
		if _, err := os.Lstat(filepath.Join(root, leaf)); !os.IsNotExist(err) {
			t.Fatal("standalone cancellation created runtime state", leaf, err)
		}
	}
	current, err := os.ReadFile(filepath.Join(store, ".reviews", hold.DecisionID, "restore", "manifest.json"))
	if err != nil || !bytes.Equal(current, body) {
		t.Fatal("cancellation changed malformed original evidence", err)
	}
	if raw, err := os.ReadFile(filepath.Join(store, ".purges", id)); err != nil || !bytes.Equal(raw, report.Journals[2].Manifest) {
		t.Fatal("cancellation changed the active fence", err)
	}
	if _, err := os.Lstat(journal); !os.IsNotExist(err) {
		t.Fatal("cancellation did not archive original operation", err)
	}
}

func TestQuarantineCancelCLIRejectsInvalidInputBeforeOpeningStore(t *testing.T) {
	for _, args := range [][]string{nil, {"id"}, {"id", "--hold", "uuid", "--inspection", "missing", "--reason", "reason"}} {
		var out, diagnostic bytes.Buffer
		if code := runQuarantineCancel(args, &out, &diagnostic); code == 0 || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatalf("invalid cancellation accepted: %v", args)
		}
	}
}
