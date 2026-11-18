package quarantine

// This helper uses only the v102 API. The lab compiles this exact test against
// both the reviewed v102 source and the candidate, so downgrade checks execute
// the actual earlier recovery and purge implementations.
import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestReviewEarlierRuntimeWorker(t *testing.T) {
	if os.Getenv("OPENEDR_REVIEW_EARLIER_WORKER") != "1" {
		t.Skip("owned earlier-runtime worker")
	}
	dir, id := os.Getenv("OPENEDR_REVIEW_EARLIER_ROOT"), os.Getenv("OPENEDR_REVIEW_EARLIER_ID")
	journal := filepath.Join(dir, ".restores", id, "manifest.json")
	original, err := os.ReadFile(journal)
	if err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(filepath.Dir(dir), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(dir, c)
	if err != nil {
		t.Fatal(err)
	}
	root, err := s.openRoot()
	if err != nil {
		t.Fatal(err)
	}
	journals, op, probeErr := openPurgeOperation(root, id)
	if op != nil {
		op.Close()
	}
	if journals != nil {
		journals.Close()
	}
	root.Close()
	if probeErr == nil || os.IsNotExist(probeErr) {
		t.Fatal("earlier runtime interpreted hold as absent purge", probeErr)
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("earlier runtime restored held case")
	}
	if err := s.Purge(id); err == nil {
		t.Fatal("earlier runtime purged held case")
	}
	_, _ = s.Recover()
	current, err := os.ReadFile(journal)
	if err != nil || !bytes.Equal(original, current) {
		t.Fatal("earlier runtime changed held journal", err)
	}
	r, found, err := c.GetQuarantine(id)
	if err != nil || !found || r.Purged || r.Restored {
		t.Fatal("earlier runtime claimed held completion", r, err)
	}
	t.Log("earlier recovery, restore and purge retain the regular-file fence")
}

func TestReviewEarlierReleasedRuntimeWorker(t *testing.T) {
	if os.Getenv("OPENEDR_REVIEW_EARLIER_RELEASED_WORKER") != "1" {
		t.Skip("owned earlier-runtime released-case worker")
	}
	dir, id := os.Getenv("OPENEDR_REVIEW_EARLIER_ROOT"), os.Getenv("OPENEDR_REVIEW_EARLIER_ID")
	c, err := cache.Open(filepath.Join(filepath.Dir(dir), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(dir, c)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.Recover()
	if _, err := os.Lstat(filepath.Join(dir, ".restores", id)); !os.IsNotExist(err) {
		t.Fatal("earlier runtime did not recover the explicitly released pre-publication operation", err)
	}
	rec, found, err := c.GetQuarantine(id)
	if err != nil || !found || rec.Restored || rec.Purged {
		t.Fatal("earlier runtime claimed publication from a release", rec, err)
	}
	requirePurgePayload(t, filepath.Join(rec.SourceCaptureDir, "source"), "payload")
	t.Log("earlier runtime recovers only after explicit release; no false restored/purged receipt")
}
