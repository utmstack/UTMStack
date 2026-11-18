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

func TestQuarantineInspectionMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_INSPECTION_WORKER") != "1" {
		return
	}
	root := os.Getenv("OPENEDR_INSPECTION_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = filepath.Join(root, "must-not-create-db")
	os.Args = []string{"utmstack_edr", "quarantine", "inspect", os.Getenv("OPENEDR_INSPECTION_ID")}
	main()
	t.Fatal("inspection should exit via its dedicated entry point")
}

func TestQuarantineInspectionMainKeepsEvidenceAndAvoidsLogAndDatabase(t *testing.T) {
	root := t.TempDir()
	id := uuid.NewString()
	store := filepath.Join(root, "quarantine")
	c, err := cache.Open(filepath.Join(root, "fixture.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, storeErr := quarantine.New(store, c)
	closeErr := c.Close()
	if storeErr != nil || closeErr != nil {
		t.Fatalf("private fixture store: %v %v", storeErr, closeErr)
	}
	op := filepath.Join(store, ".restores", id)
	parent := store
	for _, leaf := range []string{".restores", id} {
		// Use the same protected-directory contract on Windows as the store.
		// MkdirAll's Unix mode does not set SE_DACL_PROTECTED there.
		private, err := release.MakePrivateTempDir(parent, "inspection-fixture-")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(private, filepath.Join(parent, leaf)); err != nil {
			t.Fatal(err)
		}
		parent = filepath.Join(parent, leaf)
	}
	// Deliberately invalid approval: inspection captures it without trying to
	// clean it up, open SQLite or treat it as authority for a filesystem action.
	body := []byte("{unidentified journal\n")
	manifest := filepath.Join(op, "manifest.json")
	if err := os.WriteFile(manifest, body, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := json.Marshal(map[string]string{"quarantine_dir": store})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "edr.json"), cfg, 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(self, "-test.run=^TestQuarantineInspectionMainWorker$")
	cmd.Env = append(os.Environ(), "OPENEDR_INSPECTION_WORKER=1", "OPENEDR_INSPECTION_ROOT="+root, "OPENEDR_INSPECTION_ID="+id)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("inspection entry point: %v; stdout=%s stderr=%s", err, &stdout, &stderr)
	}
	var report quarantine.RecoveryInspection
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || len(report.Journals) != 3 || !bytes.Equal(report.Journals[1].Manifest, body) {
		t.Fatalf("inspection output: %s; %v", &stdout, err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected diagnostic: %s", &stderr)
	}
	if after, err := os.ReadDir(root); err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("inspection created local state: before=%v after=%v err=%v", before, after, err)
	}
	if current, err := os.ReadFile(manifest); err != nil || !bytes.Equal(current, body) {
		t.Fatal("inspection altered unresolved journal", err)
	}
}

func TestQuarantineInspectionCLIReportsFailuresOnStderr(t *testing.T) {
	previous := config.ConfigFile
	t.Cleanup(func() { config.ConfigFile = previous })
	config.ConfigFile = filepath.Join(t.TempDir(), "edr.json")
	for _, args := range [][]string{nil, {"id", "extra"}, {"../foreign"}} {
		var stdout, stderr bytes.Buffer
		if code := runQuarantineInspection(args, &stdout, &stderr); code == 0 || stderr.Len() == 0 {
			t.Fatalf("invalid inspection succeeded: %v stdout=%s stderr=%s", args, &stdout, &stderr)
		}
	}
	if err := os.WriteFile(config.ConfigFile, []byte("{bad config"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runQuarantineInspection([]string{uuid.NewString()}, &stdout, &stderr); code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("bad configuration not reported: stdout=%s stderr=%s", &stdout, &stderr)
	}
}
