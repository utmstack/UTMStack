package feed

import (
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestUpdateOnceMarksStale(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.Store(cache.VerdictRecord{SHA256: "clean1", Verdict: cache.VerdictClean, SigDBVersion: "100"})

	f := New(config.Default(), c)
	f.runUpdater = func() error { return nil }
	f.currentSigDB = func() (string, error) { return "200", nil }

	ver, err := f.updateOnce()
	if err != nil || ver != "200" {
		t.Fatalf("updateOnce ver=%q err=%v", ver, err)
	}
	if r, _, _ := c.Lookup("clean1"); !r.Stale {
		t.Fatal("clean record should be stale after signature update")
	}
}
