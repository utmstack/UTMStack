//go:build linux

package proctable

import "testing"

func TestLinuxImageMatchesAreAtomicAndCaseSensitive(t *testing.T) {
	tab := New()
	tab.Add(Proc{PID: 11, Image: "/opt/Tool", OSIdentity: 71, IdentityRequired: true})
	tab.Add(Proc{PID: 12, Image: "/opt/tool", OSIdentity: 72})
	got := tab.FindProcessesByImage("/opt/Tool")
	if len(got) != 1 || got[0].PID != 11 || got[0].OSIdentity != 71 || !got[0].IdentityRequired {
		t.Fatalf("snapshot=%+v", got)
	}
	tab.Add(Proc{PID: 11, Image: "/unrelated", OSIdentity: 73})
	if got[0].OSIdentity != 71 {
		t.Fatal("returned snapshot changed after PID replacement")
	}
}
