package quarantine

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReviewReleaseEarlierRuntimeCompatibility(t *testing.T) {
	program := os.Getenv("OPENEDR_REVIEW_EARLIER_PROGRAM")
	if program == "" {
		t.Skip("requires hash-pinned v102 program in native acceptance lab")
	}
	s, id, _, stored, approval := reviewReleaseFixture(t, "verified")
	decision, err := ReleaseRecovery(s.dir, id, approval, "authorize earlier runtime recovery of original operation")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(program, "-test.run=^TestReviewEarlierReleasedRuntimeWorker$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "OPENEDR_REVIEW_EARLIER_RELEASED_WORKER=1", "OPENEDR_REVIEW_EARLIER_ROOT="+s.dir, "OPENEDR_REVIEW_EARLIER_ID="+id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("earlier released-case worker: %v %s", err, output)
	}
	requirePurgePayload(t, stored, "payload")
	if _, err := os.Lstat(filepath.Join(s.dir, ".restores", id)); !os.IsNotExist(err) {
		t.Fatal("earlier runtime did not retire the approved pre-publication operation", err)
	}
	again, err := ReleaseRecovery(s.dir, id, approval, decision.Reason)
	if err != nil || !reflect.DeepEqual(decision, again) {
		t.Fatal("earlier recovery changed release history", err)
	}
	t.Logf("v102 released-case output: %s", output)
}
