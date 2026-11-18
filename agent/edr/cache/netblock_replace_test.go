package cache

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestReplaceIndicatorsRemovalPersistsAcrossRestartIncludingEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.ReplaceIndicators([]IndicatorRecord{{Value: "192.0.2.1", Type: "ip"}, {Value: "192.0.2.2", Type: "ip"}}); err != nil {
		t.Fatal(err)
	}
	if err = c.ReplaceIndicators([]IndicatorRecord{{Value: "192.0.2.2", Type: "ip"}}, FeedState{Feed: "level1/ip", Checksum: "receipt"}); err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	c, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	rows, err := c.ListIndicators()
	if err != nil || len(rows) != 1 || rows[0].Value != "192.0.2.2" {
		t.Fatalf("removed IOC resurrected: %v %v", rows, err)
	}
	if err = c.ReplaceIndicators(nil); err != nil {
		t.Fatal(err)
	}
	rows, err = c.ListIndicators()
	if err != nil || len(rows) != 0 {
		t.Fatalf("empty successful snapshot retained old IOCs %v %v", rows, err)
	}
}

func TestReplaceIndicatorsFailedLaterBatchRestoresOldRowsAndReceipts(t *testing.T) {
	c := openTmp(t)
	if err := c.ReplaceIndicators([]IndicatorRecord{{Value: "198.51.100.1", Type: "ip"}}, FeedState{Feed: "level1/ip", Checksum: "old"}); err != nil {
		t.Fatal(err)
	}
	// Duplicate primary keys in the second SQL batch must roll back the earlier
	// delete, first successful insert batch, and receipt replacement together.
	var next []IndicatorRecord
	for i := 0; i < 600; i++ {
		next = append(next, IndicatorRecord{Value: fmt.Sprintf("203.0.%d.%d", i/256, i%256), Type: "ip"})
	}
	next = append(next, next[0])
	if err := c.ReplaceIndicators(next, FeedState{Feed: "level1/ip", Checksum: "new"}); err == nil {
		t.Fatal("duplicate input unexpectedly committed")
	}
	rows, err := c.ListIndicators()
	if err != nil || len(rows) != 1 || rows[0].Value != "198.51.100.1" {
		t.Fatalf("rollback lost old snapshot %v %v", rows, err)
	}
	state, err := c.GetFeedState("level1/ip")
	if err != nil || state.Checksum != "old" {
		t.Fatalf("receipt advanced after rollback %+v %v", state, err)
	}
}

func TestReplaceIndicatorsRejectsBoundViolationsBeforeMutation(t *testing.T) {
	c := openTmp(t)
	if err := c.ReplaceIndicators([]IndicatorRecord{{Value: "192.0.2.1", Type: "ip"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.ReplaceIndicators(make([]IndicatorRecord, MaxIndicatorRecords+1)); err == nil {
		t.Fatal("record budget ignored")
	}
	if err := c.ReplaceIndicators([]IndicatorRecord{{Value: "x", Type: "unknown"}}); err == nil {
		t.Fatal("invalid type ignored")
	}
	rows, _ := c.ListIndicators()
	if len(rows) != 1 || rows[0].Value != "192.0.2.1" {
		t.Fatal("invalid input changed cache")
	}
}
