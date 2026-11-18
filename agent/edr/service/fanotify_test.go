package service

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func TestPinnedScannerDoesNotReopenSubstitutedPath(t *testing.T) {
	dir := t.TempDir()
	marker := []byte("OPENEDR_PINNED_SAFE_TEST_MARKER_2026")
	if err := os.WriteFile(filepath.Join(dir, "marker.ndb"), []byte("Pinned.Safe:0:*:"+hex.EncodeToString(marker)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = dir
	eng := engine.New(cfg)
	defer eng.Stop()
	if err := eng.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "inert.fixture")
	if err := os.WriteFile(name, marker, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := os.Rename(name, name+".original"); err != nil {
		t.Skipf("platform does not permit renaming an open fixture: %v", err)
	}
	if err := os.WriteFile(name, []byte("ordinary benign replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	verdict, err := scanPinned(context.Background(), eng, f, 1<<20)
	if err != nil || !verdict.Detected || verdict.Clean || verdict.Generation == "" {
		t.Fatalf("%+v %v", verdict, err)
	}
}
