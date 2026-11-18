package procwatch

import "testing"

func TestConnectorBirthClockBounds(t *testing.T) {
	for _, c := range []struct {
		name                    string
		event                   uint64
		ticks                   int64
		hz, before, boot, after uint64
		want                    bool
	}{
		{"older identity", 2_000_000_000, 100, 100, 3_000_000_000, 3_000_000_010, 3_000_000_020, true},
		{"replacement after queued event", 1_000_000_000, 200, 100, 3_000_000_000, 3_000_000_010, 3_000_000_020, false},
		{"newborn same tick ambiguous", 1_005_000_000, 100, 100, 2_000_000_000, 2_000_000_010, 2_000_000_020, false},
		{"tick upper endpoint", 1_010_000_000, 100, 100, 2_000_000_000, 2_000_000_010, 2_000_000_020, true},
		{"suspend offset", 2_000_000_000, 1090, 100, 3_000_000_000, 13_000_000_010, 3_000_000_020, true},
		{"offset uncertainty", 1_010_000_000, 1101, 100, 2_000_000_000, 12_000_000_010, 2_000_000_020, false},
		{"zero event", 0, 1, 100, 2, 3, 4, false},
		{"future event", 5, 1, 100, 2, 3, 4, false},
		{"overflow ticks", 2, 9223372036854775807, 1, 2, 3, 4, false},
		{"invalid bracket", 2, 1, 100, 4, 3, 2, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			initialOffset := uint64(0)
			if c.name == "suspend offset" {
				initialOffset = 9_999_999_990
			}
			if got := connectorBirthPredates(c.event, c.ticks, c.hz, c.before, c.boot, c.after, initialOffset); got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestQueuedPreSuspendEventCannotAuthorizeReusedPID(t *testing.T) {
	if connectorBirthPredates(100_000_000_000, 10100, 100, 103_000_000_000, 203_000_000_010, 103_000_000_020, 0) {
		t.Fatal("post-suspend clock offset falsely bound an old event to a replacement")
	}
}
