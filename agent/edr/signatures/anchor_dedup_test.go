package signatures

import "testing"

func TestAnchorHashCollisionStillComparesLiteralBytes(t *testing.T) {
	all := []*anchorGroup{{literal: []byte("first")}, {literal: []byte("second")}}
	// Both entries deliberately share one hash bucket; slots are one-based.
	table := map[uint32]uint32{7: 2}
	next := []uint32{0, 1}
	for _, tc := range []struct {
		literal string
		want    *anchorGroup
	}{{"first", all[0]}, {"second", all[1]}, {"third", nil}} {
		if got := findAnchorGroup(table, next, all, 7, []byte(tc.literal)); got != tc.want {
			t.Errorf("lookup %q = %p, want %p", tc.literal, got, tc.want)
		}
	}
}
