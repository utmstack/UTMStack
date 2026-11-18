package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestQuarantineFreshInspectionMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_INSPECTION_WORKER") != "1" {
		t.Skip("owned standalone fresh inspection worker")
	}
	root := os.Getenv("OPENEDR_FRESH_INSPECTION_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = filepath.Join(root, "must-not-open-db")
	os.Args = []string{"utmstack_edr", "quarantine", "inspect-fresh", "owned-case", "--hold", "00000000-0000-0000-0000-000000000001", "--destination", filepath.Join(root, "new-output")}
	main()
	t.Fatal("fresh inspection must exit via its dedicated entry point")
}

func TestQuarantineFreshInspectionMainFailsBeforeLogDatabaseOrStorageCreation(t *testing.T) {
	root := t.TempDir()
	// A missing storage directory must stay missing. The normal main startup
	// would initialize its log and database before handling management actions.
	if err := os.WriteFile(filepath.Join(root, "edr.json"), []byte(`{"quarantine_dir":"missing-owned-quarantine"}`), 0600); err != nil {
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
	cmd := exec.Command(self, "-test.run=^TestQuarantineFreshInspectionMainWorker$", "-test.count=1")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "OPENEDR_FRESH_INSPECTION_WORKER=1", "OPENEDR_FRESH_INSPECTION_ROOT="+root)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || stdout.Len() != 0 || !bytes.Contains(stderr.Bytes(), []byte("fresh inspection incomplete:")) {
		t.Fatalf("fresh inspection did not fail through its standalone route: %v stdout=%s stderr=%s", err, &stdout, &stderr)
	}
	if after, err := os.ReadDir(root); err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("fresh inspection created local state", err)
	}
}

func TestQuarantineFreshInspectionCLIRejectsIncompleteArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"id"}, {"id", "--wrong", "hold", "--destination", "path"}, {"id", "--hold", "hold", "--wrong", "path"}} {
		var stdout, stderr bytes.Buffer
		if code := runQuarantineFreshInspection(args, &stdout, &stderr); code != 1 || stdout.Len() != 0 || !bytes.Contains(stderr.Bytes(), []byte("usage:")) {
			t.Fatalf("incomplete arguments accepted: %v stdout=%s stderr=%s", args, &stdout, &stderr)
		}
	}
}
