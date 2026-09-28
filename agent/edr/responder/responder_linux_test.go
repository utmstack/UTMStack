//go:build linux

package responder

import (
	"testing"
)

func TestParseStatLine(t *testing.T) {
	// Normal line: ppid is field 4, after the comm.
	if pid, ppid, ok := parseStatLine("100 (sleep) S 99 100 100 0 -1 4194304 0"); !ok || pid != 100 || ppid != 99 {
		t.Fatalf("parseStatLine normal = (%d, %d, %v), want (100, 99, true)", pid, ppid, ok)
	}
	// Comm containing spaces and parens: the ppid must come after the LAST ')'.
	if pid, ppid, ok := parseStatLine("123 (my proc (x)) S 100 123 123 0 -1 4194304"); !ok || pid != 123 || ppid != 100 {
		t.Fatalf("parseStatLine tricky comm = (%d, %d, %v), want (123, 100, true)", pid, ppid, ok)
	}
	// Garbage: no closing paren, non-numeric pid, too few fields after comm.
	for _, bad := range []string{
		"100 (sleep S 99",
		"abc (sleep) S 99 100",
		"100 (sleep) S",
	} {
		if _, _, ok := parseStatLine(bad); ok {
			t.Fatalf("parseStatLine(%q) = ok, want fail", bad)
		}
	}
}

func TestDescendantsFromLeavesFirst(t *testing.T) {
	//        100
	//        /   \
	//      200   300
	//       |
	//      400
	snap := map[int]int{100: 1, 200: 100, 300: 100, 400: 200}
	got, ok := descendantsFrom(snap, 100)
	if !ok {
		t.Fatal("descendantsFrom(root present) = ok=false, want true")
	}
	if len(got) != 4 {
		t.Fatalf("descendantsFrom = %v, want 4 pids", got)
	}
	// Post-order: the two leaves (400, 300) come before their parents, and the
	// root is always last.
	pos := map[int]int{}
	for i, pid := range got {
		pos[pid] = i
	}
	if pos[400] >= pos[200] {
		t.Fatalf("leaf 400 must precede parent 200: %v", got)
	}
	if pos[200] >= pos[100] || pos[300] >= pos[100] {
		t.Fatalf("root 100 must be last: %v", got)
	}
}

func TestDescendantsFromKernelThreadsSkipped(t *testing.T) {
	// pid 200 has ppid 0 (kernel thread) and pid 300 has ppid 2 (kthreadd):
	// neither may appear under root 100.
	snap := map[int]int{100: 1, 200: 0, 300: 2, 400: 100}
	got, ok := descendantsFrom(snap, 100)
	if !ok {
		t.Fatal("descendantsFrom(root present) = ok=false, want true")
	}
	for _, pid := range got {
		if pid == 200 || pid == 300 {
			t.Fatalf("kernel thread %d leaked into tree: %v", pid, got)
		}
	}
	if len(got) != 2 || got[0] != 400 || got[1] != 100 {
		t.Fatalf("descendantsFrom = %v, want [400 100]", got)
	}
}

func TestDescendantsFromSelfParentCycle(t *testing.T) {
	// PID reuse can leave a process pointing at itself: the walk must not
	// recurse forever and must still return the root.
	snap := map[int]int{100: 1, 200: 200, 300: 100}
	got, ok := descendantsFrom(snap, 100)
	if !ok {
		t.Fatal("descendantsFrom(root present) = ok=false, want true")
	}
	for _, pid := range got {
		if pid == 200 {
			t.Fatalf("self-parent cycle pid 200 leaked into tree: %v", got)
		}
	}
	if len(got) != 2 || got[0] != 300 || got[1] != 100 {
		t.Fatalf("descendantsFrom = %v, want [300 100]", got)
	}
}

func TestDescendantsFromMissingRoot(t *testing.T) {
	if _, ok := descendantsFrom(map[int]int{100: 1, 200: 100}, 999); ok {
		t.Fatal("descendantsFrom(root absent) = ok, want false (caller falls back to proctable)")
	}
}
