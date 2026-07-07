// edr/cache/netblock_test.go
package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func openTmp(t *testing.T) *Cache {
	t.Helper()
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestIndicatorRoundTrip(t *testing.T) {
	c := openTmp(t)
	now := time.Now().UTC()
	in := []IndicatorRecord{
		{Value: "1.2.3.4", Type: "ip", Level: 1, ListName: "level1/ip", FirstSeen: now, LastSeen: now},
		{Value: "10.0.0.0/8", Type: "cidr", Level: 1, ListName: "level1/ip", FirstSeen: now, LastSeen: now},
	}
	if err := c.PutIndicators(in); err != nil {
		t.Fatal(err)
	}
	got, err := c.ListIndicators()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
}

func TestFeedState(t *testing.T) {
	c := openTmp(t)
	if err := c.SaveFeedState(FeedState{Feed: "level1/ip", IndicatorNum: 5, Checksum: "abc"}); err != nil {
		t.Fatal(err)
	}
	fs, err := c.GetFeedState("level1/ip")
	if err != nil {
		t.Fatal(err)
	}
	if fs.IndicatorNum != 5 || fs.Checksum != "abc" {
		t.Fatalf("bad feedstate: %+v", fs)
	}
}
