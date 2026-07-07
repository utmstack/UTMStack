package proctable

import (
	"reflect"
	"sort"
	"testing"
)

func TestDescendantsLeavesFirst(t *testing.T) {
	tab := New()
	tab.Add(Proc{PID: 100, PPID: 4, Image: `C:\a.exe`})
	tab.Add(Proc{PID: 200, PPID: 100, Image: `C:\b.exe`})
	tab.Add(Proc{PID: 300, PPID: 200, Image: `C:\c.exe`})
	tab.Add(Proc{PID: 400, PPID: 100, Image: `C:\d.exe`})

	got := tab.Descendants(100) // want children before parents; 100 last
	if len(got) != 4 || got[len(got)-1] != 100 {
		t.Fatalf("descendants = %v (100 must be last)", got)
	}
	idx := func(v int) int {
		for i, x := range got {
			if x == v {
				return i
			}
		}
		return -1
	}
	if idx(300) > idx(200) || idx(200) > idx(100) || idx(400) > idx(100) {
		t.Fatalf("not leaves-first: %v", got)
	}
}

func TestFindByImageCaseInsensitive(t *testing.T) {
	tab := New()
	tab.Add(Proc{PID: 1, Image: `C:\Users\x\A.EXE`})
	tab.Add(Proc{PID: 2, Image: `C:\Users\x\other.exe`})
	got := tab.FindByImage(`c:\users\x\a.exe`)
	sort.Ints(got)
	if !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("FindByImage = %v, want [1]", got)
	}
}
