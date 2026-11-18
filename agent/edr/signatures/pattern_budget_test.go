package signatures

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// A failed bytes.Index still examines its window. The work cap must account
// for the whole search and for long literal comparisons near a match.
func TestPatternSearchChargesFailedWindow(t *testing.T) {
	data := bytes.Repeat([]byte{'A'}, 1<<20)
	p, err := compileBody("*", "4141414142")
	if err != nil {
		t.Fatal(err)
	}
	w := &matchWork{ctx: context.Background(), limit: 1000}
	if _, _, err := p.countWithOffset(data, nil, w, 0); !errors.Is(err, ErrMatchLimit) {
		t.Fatal("failed prefix search did not exhaust budget", err, w.steps)
	}
	gap, err := compileBody("*", "4142{0-}4141414142")
	if err != nil {
		t.Fatal(err)
	}
	data[1] = 'B'
	w = &matchWork{ctx: context.Background(), limit: 1000}
	if _, err := gap.matchesAt(data, 0, w); !errors.Is(err, ErrMatchLimit) {
		t.Fatal("failed gap search did not exhaust budget", err, w.steps)
	}
}

func TestPatternLiteralComparisonChargesBytes(t *testing.T) {
	p, err := compileBody("*", "414141414141414141414141414141414141414141414141414141414141414142")
	if err != nil {
		t.Fatal(err)
	}
	w := &matchWork{ctx: context.Background(), limit: 16}
	if _, err := p.matchesAt(bytes.Repeat([]byte{'A'}, 33), 0, w); !errors.Is(err, ErrMatchLimit) {
		t.Fatal("long literal comparison did not exhaust budget", err, w.steps)
	}
}
