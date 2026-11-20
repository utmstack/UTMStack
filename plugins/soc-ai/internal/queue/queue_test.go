package queue

import "testing"

func TestHasParent(t *testing.T) {
	cases := map[string]bool{
		`{"alertId":"a","parentId":"p1"}`: true,
		`{"alertId":"a","parentId":""}`:   false,
		`{"alertId":"a"}`:                 false,
		`not json`:                        false,
		``:                                false,
	}
	for in, want := range cases {
		if got := hasParent(in); got != want {
			t.Errorf("hasParent(%q) = %v, want %v", in, got, want)
		}
	}
}
