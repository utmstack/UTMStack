package cache

import (
	"path/filepath"
	"testing"
)

func TestStoreThenLookup(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer c.Close()

	if err := c.Store(VerdictRecord{SHA256: "hash1", Verdict: VerdictMalicious, Signature: "Win.Test.EICAR", SigDBVersion: "27000"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, found, err := c.Lookup("hash1")
	if err != nil || !found {
		t.Fatalf("Lookup found=%v err=%v", found, err)
	}
	if got.Verdict != VerdictMalicious || got.Signature != "Win.Test.EICAR" {
		t.Fatalf("bad record: %+v", got)
	}
	if _, found, _ := c.Lookup("missing"); found {
		t.Fatal("unexpected hit for missing hash")
	}
}
