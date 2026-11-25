package common_models

import (
	"testing"
	"time"

	"github.com/utmstack/go-sdk/v2/store"
)

func between(v any) []FilterType {
	return []FilterType{{Field: "@timestamp", Operator: OpIsBetween, Value: v}}
}

// "now-24h" was dropped without a word, so a query for the last day scanned the whole table.
func TestRelativeTimeBoundsAreApplied(t *testing.T) {
	scope, rest := SplitTimeBounds(store.Scope{}, between([]any{"now-24h", "now"}))

	if len(rest) != 0 || scope.From.IsZero() || scope.To.IsZero() {
		t.Fatalf("scope = %+v rest = %+v", scope, rest)
	}

	if d := time.Since(scope.From); d < 23*time.Hour+59*time.Minute || d > 24*time.Hour+time.Minute {
		t.Fatalf("from should be about 24h ago, it is %s ago", d)
	}

	for in, want := range map[string]time.Duration{"now-30m": 30 * time.Minute, "now-7d": 7 * 24 * time.Hour, "now-2w": 14 * 24 * time.Hour, "NOW-1H": time.Hour} {
		got := time.Since(asTime(in))
		if got < want-time.Second || got > want+time.Minute {
			t.Errorf("%s: %s ago, want about %s", in, got, want)
		}
	}
}

func TestAbsoluteTimeBoundsStillWork(t *testing.T) {
	scope, _ := SplitTimeBounds(store.Scope{}, between([]any{"2026-11-25T10:00:00Z", "2026-11-25 11:30:00"}))

	if scope.From.Hour() != 10 || scope.To.Hour() != 11 || scope.To.Minute() != 30 {
		t.Fatalf("%+v", scope)
	}
}

func TestValidateTimeBoundsRefusesWhatCannotBeRead(t *testing.T) {
	bad := [][]FilterType{
		between([]any{"yesterday", "now"}),
		between([]any{"now-24h"}),
		between("now-24h"),
		{{Field: "@timestamp", Operator: OpGreater, Value: "soon"}},
	}

	for _, f := range bad {
		if err := ValidateTimeBounds(f); err == nil {
			t.Errorf("%+v should be refused", f)
		}
	}

	good := [][]FilterType{
		nil,
		between([]any{"now-24h", "now"}),
		between([]any{"2026-11-25T10:00:00Z", "2026-11-25T11:00:00Z"}),
		{{Field: "severity", Operator: OpEquals, Value: "high"}},
	}

	for _, f := range good {
		if err := ValidateTimeBounds(f); err != nil {
			t.Errorf("%+v: %v", f, err)
		}
	}
}
