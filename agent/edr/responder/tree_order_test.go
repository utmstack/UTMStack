package responder

import (
	"reflect"
	"testing"
)

func TestMergeProcessTreesPreservesAncestry(t *testing.T) {
	for _, tc := range []struct {
		name        string
		table, live []processParent
		want        []int
	}{
		{"live-root-table-child", []processParent{{100, 4}, {200, 100}}, []processParent{{100, 4}}, []int{200, 100}},
		{"table-grandchild-of-live-child", []processParent{{100, 4}, {200, 100}, {300, 200}}, []processParent{{200, 100}, {100, 4}}, []int{300, 200, 100}},
		{"both-views-add-children", []processParent{{100, 4}, {200, 100}}, []processParent{{300, 100}, {100, 4}}, []int{200, 300, 100}},
		{"live-parent-wins", []processParent{{100, 4}, {200, 100}, {300, 200}}, []processParent{{200, 300}, {300, 100}, {100, 4}}, []int{200, 300, 100}},
		{"disconnected-cycle", []processParent{{100, 4}, {200, 300}, {300, 200}}, nil, []int{100}},
		{"missing-root-record", []processParent{{200, 100}}, nil, []int{200, 100}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mergeProcessTrees(100, tc.table, tc.live); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("response order=%v want=%v", got, tc.want)
			}
		})
	}
}
