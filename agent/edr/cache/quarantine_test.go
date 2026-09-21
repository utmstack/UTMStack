package cache

import (
	"path/filepath"
	"testing"
)

func TestQuarantineRecordRoundTrip(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	rec := QuarantineRecord{QuarantineID: "q1", OriginalPath: `C:\x\a.exe`, SHA256: "h", Detection: "Win.Test.EICAR", Engine: "UTMStack EDR", Restorable: true}
	if err := c.StoreQuarantine(rec); err != nil {
		t.Fatalf("StoreQuarantine: %v", err)
	}
	got, found, err := c.GetQuarantine("q1")
	if err != nil || !found {
		t.Fatalf("GetQuarantine found=%v err=%v", found, err)
	}
	if got.Detection != "Win.Test.EICAR" || got.Engine != "UTMStack EDR" {
		t.Fatalf("bad record: %+v", got)
	}
	if err := c.MarkRestored("q1"); err != nil {
		t.Fatal(err)
	}
	got, _, _ = c.GetQuarantine("q1")
	if !got.Restored {
		t.Fatal("record not marked restored")
	}
}

func TestMarkStaleBySigDB(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.Store(VerdictRecord{SHA256: "clean1", Verdict: VerdictClean, SigDBVersion: "100"})
	_ = c.Store(VerdictRecord{SHA256: "mal1", Verdict: VerdictMalicious, SigDBVersion: "100"})

	n, err := c.MarkStaleBySigDB("200")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 { // only the clean record goes stale
		t.Fatalf("marked %d stale, want 1", n)
	}
	if r, _, _ := c.Lookup("clean1"); !r.Stale {
		t.Fatal("clean1 should be stale")
	}
	if r, _, _ := c.Lookup("mal1"); r.Stale {
		t.Fatal("malicious verdict must persist (not stale)")
	}
}
