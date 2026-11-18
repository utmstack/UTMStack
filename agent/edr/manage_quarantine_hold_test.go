package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/release"
)

func TestQuarantineHoldMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_HOLD_MAIN_WORKER") != "1" {
		return
	}
	root := os.Getenv("OPENEDR_HOLD_MAIN_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = filepath.Join(root, "must-not-open-db")
	os.Args = []string{"utmstack_edr", "quarantine", "hold", os.Getenv("OPENEDR_HOLD_MAIN_ID"), "--inspection", filepath.Join(root, "inspection.json"), "--reason", "retain malformed case for independent review"}
	main()
	t.Fatal("hold should exit via the dedicated entry point")
}

func TestQuarantineHoldMainPreservesEvidenceWithoutDatabaseOrLog(t *testing.T) {
	root, id := t.TempDir(), uuid.NewString()
	store := filepath.Join(root, "quarantine")
	c, err := cache.Open(filepath.Join(root, "fixture.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, storeErr := quarantine.New(store, c)
	closeErr := c.Close()
	if storeErr != nil || closeErr != nil {
		t.Fatal(storeErr, closeErr)
	}
	parent := store
	for _, leaf := range []string{".restores", id} {
		private, err := release.MakePrivateTempDir(parent, "hold-fixture-")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(private, filepath.Join(parent, leaf)); err != nil {
			t.Fatal(err)
		}
		parent = filepath.Join(parent, leaf)
	}
	body := []byte("{ owned malformed evidence \xff")
	manifest := filepath.Join(parent, "manifest.json")
	if err := os.WriteFile(manifest, body, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := quarantine.InspectRecovery(store, id)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(report)
	cfg, _ := json.Marshal(map[string]string{"quarantine_dir": store})
	for name, data := range map[string][]byte{"edr.json": cfg, "inspection.json": encoded} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var first quarantine.RecoveryHold
	for attempt := 0; attempt < 2; attempt++ {
		cmd := exec.Command(self, "-test.run=^TestQuarantineHoldMainWorker$")
		cmd.Env = append(os.Environ(), "OPENEDR_HOLD_MAIN_WORKER=1", "OPENEDR_HOLD_MAIN_ROOT="+root, "OPENEDR_HOLD_MAIN_ID="+id)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("hold main: %v %s %s", err, &stdout, &stderr)
		}
		var decision quarantine.RecoveryHold
		if err := json.Unmarshal(stdout.Bytes(), &decision); err != nil || !bytes.Equal(decision.Original.Manifest, body) || stderr.Len() != 0 {
			t.Fatalf("invalid decision: %s %s %v", &stdout, &stderr, err)
		}
		if attempt == 0 {
			first = decision
		} else if first.DecisionID != decision.DecisionID || first.FileID != decision.FileID {
			t.Fatal("lost-return retry created another decision")
		}
	}
	for _, name := range []string{"must-not-create-log", "must-not-open-db", "must-not-open-db-wal", "must-not-open-db-shm"} {
		if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatal("hold created runtime state", name, err)
		}
	}
	current, err := os.ReadFile(manifest)
	if err != nil || !bytes.Equal(current, body) {
		t.Fatal("hold changed original journal", err)
	}
}

func TestQuarantineHoldCLIRejectsInvalidInputBeforeOpeningStore(t *testing.T) {
	for _, args := range [][]string{nil, {"id"}, {"id", "--inspection", "missing", "--reason", "reason"}} {
		var out, diagnostic bytes.Buffer
		if code := runQuarantineHold(args, &out, &diagnostic); code == 0 || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatalf("invalid hold accepted: %v", args)
		}
	}
}
