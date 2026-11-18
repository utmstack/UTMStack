package signatures

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestMatchBudgetGrowsWithObjectSize(t *testing.T) {
	d := DefaultLimits()
	if got := d.matchLimit(1 << 20); got != 500_000_000 {
		t.Fatalf("1 MiB budget = %d", got)
	}
	if got := d.matchLimit(100 << 20); got != 16*(100<<20) {
		t.Fatalf("100 MiB budget = %d", got)
	}
	if got := (Limits{MaxMatchSteps: 8}).defaults().matchLimit(1 << 20); got != 8 {
		t.Fatalf("an explicit step cap must stay a hard cap, got %d", got)
	}
	if got := (Limits{MatchStepsPerByte: 4}).defaults().matchLimit(200 << 20); got != 4*(200<<20) {
		t.Fatalf("explicit per-byte budget = %d", got)
	}
}

func TestLargeObjectFinishesWithinScaledBudget(t *testing.T) {
	files := map[string][]byte{"a.ndb": []byte("Tail.Marker:0:*:554e49515545544149" + "4c\n")}
	data := append(bytes.Repeat([]byte("abcdefgh"), 64<<10), []byte("UNIQUETAIL")...)
	fixed, err := CompileWithLimits(files, Limits{MaxMatchSteps: 10_000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixed.Match(context.Background(), data); !errors.Is(err, ErrMatchLimit) {
		t.Fatalf("a small fixed budget must run out on %d bytes, got %v", len(data), err)
	}
	scaled, err := CompileWithLimits(files, Limits{MaxMatchSteps: 10_000, MatchStepsPerByte: 16})
	if err != nil {
		t.Fatal(err)
	}
	if name, err := scaled.Match(context.Background(), data); name != "Tail.Marker" || err != nil {
		t.Fatalf("scaled budget: %q %v", name, err)
	}
}
