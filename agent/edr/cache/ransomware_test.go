package cache

import (
	"path/filepath"
	"testing"
)

func TestCanaryAndIncidentPersist(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if err := c.StoreCanary(CanaryRecord{Path: `C:\Users\x\00__a.xlsx`, SHA256: "abc", Volume: "C:"}); err != nil {
		t.Fatal(err)
	}
	cs, err := c.ListCanaries()
	if err != nil || len(cs) != 1 || cs[0].Path != `C:\Users\x\00__a.xlsx` {
		t.Fatalf("canary not persisted: %v %v", cs, err)
	}
	if cs[0].PlacedAt.IsZero() {
		t.Fatal("PlacedAt not stamped")
	}

	if err := c.StoreIncident(RansomwareIncident{
		ID: "i1", PID: 42, Image: `C:\enc.exe`, Signals: "canary,t1490",
		Score: 200, Action: "contained", QuarantineID: "q1",
	}); err != nil {
		t.Fatal(err)
	}
	// second store with same ID must upsert, not error
	if err := c.StoreIncident(RansomwareIncident{ID: "i1", PID: 42, Action: "contained"}); err != nil {
		t.Fatalf("incident upsert failed: %v", err)
	}
}
