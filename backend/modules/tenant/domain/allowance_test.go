package domain

import "testing"

func TestAllowanceOf(t *testing.T) {
	n := func(v int) *int { return &v }
	cases := []struct {
		name     string
		tenant   *int
		instance int
		want     int
	}{
		{"no tenant cap inherits the instance", nil, 100, 100},
		{"no tenant cap on an uncapped instance stays uncapped", nil, -1, -1},
		{"tenant cap applies when the instance is uncapped", n(500), -1, 500},
		{"a cap of none denies even on an uncapped instance", n(0), -1, 0},
		{"tenant cap below the instance applies", n(50), 100, 50},
		{"tenant cap above the instance is clamped", n(500), 100, 100},
	}
	for _, c := range cases {
		got := Limits{MaxAIRequests: c.tenant}.AllowanceOf(c.instance)
		if got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}
