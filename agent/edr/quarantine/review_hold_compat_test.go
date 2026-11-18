package quarantine

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReviewHoldEarlierRuntimeCompatibility(t *testing.T) {
	program := os.Getenv("OPENEDR_REVIEW_EARLIER_PROGRAM")
	if program == "" {
		t.Skip("requires hash-pinned v102 program in native acceptance lab")
	}
	s, id, original, stored, approval := reviewHoldFixture(t, "publication-intent")
	m := restoreFixtureManifest(t, s, id)
	if _, err := HoldRecovery(s.dir, id, approval, "owned downgrade retention check"); err != nil {
		t.Fatal(err)
	}
	r, _, err := s.cache.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(program, "-test.run=^TestReviewEarlierRuntimeWorker$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "OPENEDR_REVIEW_EARLIER_WORKER=1", "OPENEDR_REVIEW_EARLIER_ROOT="+s.dir, "OPENEDR_REVIEW_EARLIER_ID="+id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("earlier-runtime worker: %v %s", err, output)
	}
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, filepath.Join(r.SourceCaptureDir, "source"), "payload")
	requirePurgePayload(t, filepath.Join(filepath.Dir(original), m.stageName(), "snapshot"), "payload")
	if _, err := os.Lstat(original); !os.IsNotExist(err) {
		t.Fatal("earlier runtime published destination", err)
	}
	t.Logf("v102 output: %s", output)
}
