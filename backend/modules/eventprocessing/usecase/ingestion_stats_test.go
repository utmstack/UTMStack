package usecase

import (
	"testing"
	"time"
)

// The panel asks for the last 24 hours by default, and used to hand the store
// the literal string "now-24h". Nothing understood it, so no bound was applied
// and the answer covered all of history under a label that said 24 hours.
func TestTheDefaultWindowIsAnActualWindow(t *testing.T) {
	before := time.Now().UTC()
	from, to := resolveWindow("", "")
	after := time.Now().UTC()

	if to.Before(before) || to.After(after.Add(time.Second)) {
		t.Errorf("to = %v, want about now", to)
	}
	if d := to.Sub(from); d < 23*time.Hour+59*time.Minute || d > 24*time.Hour+time.Minute {
		t.Errorf("the window is %v, want 24h", d)
	}
}

func TestTheWindowUnderstandsWhatAskersWrite(t *testing.T) {
	now := time.Now().UTC()

	cases := map[string]time.Duration{
		"now-15m": 15 * time.Minute,
		"now-2h":  2 * time.Hour,
		"now-7d":  7 * 24 * time.Hour,
		"now-2w":  14 * 24 * time.Hour,
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			from, to := resolveWindow(expr, "now")
			if d := to.Sub(from); d < want-time.Second || d > want+time.Second {
				t.Errorf("%s gave %v, want %v", expr, d, want)
			}
		})
	}

	from, to := resolveWindow("2026-08-01T00:00:00Z", "2026-08-02T00:00:00Z")
	if to.Sub(from) != 24*time.Hour {
		t.Errorf("an absolute window gave %v", to.Sub(from))
	}

	// Written backwards, it is still the same window.
	from, to = resolveWindow("2026-08-02T00:00:00Z", "2026-08-01T00:00:00Z")
	if !from.Before(to) {
		t.Error("the bounds were left the wrong way round")
	}

	// Something nobody can read must not become "no bound at all".
	from, to = resolveWindow("last tuesday", "")
	if d := to.Sub(from); d < 23*time.Hour {
		t.Errorf("an unreadable bound gave %v, want the default window", d)
	}
	if now.Sub(to) > time.Minute {
		t.Error("the window does not end near now")
	}
}

// Auto picks the bucket from the window, which only works if the window is
// already resolved — it used to be parsed from the same unreadable strings and
// fell back to an hour every time.
func TestAutoPicksTheBucketFromTheWindow(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		window time.Duration
		want   string
	}{
		{time.Hour, "5m"},
		{24 * time.Hour, "1h"},
		{7 * 24 * time.Hour, "1d"},
		{90 * 24 * time.Hour, "7d"},
	}
	for _, c := range cases {
		if got := resolveInterval("auto", now.Add(-c.window), now); got != c.want {
			t.Errorf("a %v window chose %q, want %q", c.window, got, c.want)
		}
	}

	if got := resolveInterval("15m", now.Add(-time.Hour), now); got != "15m" {
		t.Errorf("an explicit interval was overridden with %q", got)
	}
}
