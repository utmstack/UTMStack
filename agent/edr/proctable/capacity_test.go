package proctable

import (
	"strings"
	"testing"
)

func TestTableBoundsCountAndKeepsMostRecentUpdate(t *testing.T) {
	tab := New()
	for pid := 1; pid <= MaxProcesses; pid++ {
		tab.Add(Proc{PID: pid, OSIdentity: int64(pid)})
	}
	tab.Add(Proc{PID: 1, OSIdentity: 99999})
	tab.Add(Proc{PID: MaxProcesses + 1, OSIdentity: 100000})
	if len(tab.procs) != MaxProcesses || tab.Evictions() != 1 || tab.Health() == nil {
		t.Fatalf("unbounded/silent eviction: count %d, evicted %d", len(tab.procs), tab.Evictions())
	}
	if _, ok := tab.Get(2); ok {
		t.Fatal("oldest unrefreshed record retained")
	}
	if p, ok := tab.Get(1); !ok || p.OSIdentity != 99999 {
		t.Fatal("recent replacement lost")
	}
}
func TestTableBoundsRetainedMetadataAndReleasesAccounting(t *testing.T) {
	tab := New()
	command := strings.Repeat("x", 64<<10)
	for pid := 1; pid <= 300; pid++ {
		tab.Add(Proc{PID: pid, Cmdline: command, OSIdentity: int64(pid)})
	}
	if tab.retained > MaxRetainedBytes || tab.Evictions() == 0 || len(tab.procs) >= 300 {
		t.Fatalf("metadata budget ignored: bytes=%d, entries=%d", tab.retained, len(tab.procs))
	}
	for pid := range tab.procs {
		tab.Remove(pid)
	}
	if tab.retained != 0 || tab.order.Len() != 0 || len(tab.nodes) != 0 {
		t.Fatal("removed metadata retained")
	}
}
