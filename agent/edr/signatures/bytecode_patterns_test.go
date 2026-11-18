package signatures

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestBytecodePatternCountsAndFirstOffsets(t *testing.T) {
	for _, tc := range []struct {
		name, offset, body, data string
		count                    int
		first                    uint32
	}{
		{"overlap", "*", "414141", "xxAAAAA", 3, 2},
		{"leading-wildcard", "*", "??4142", "zxxABxAB", 2, 2},
		{"leading-bounded-gap", "*", "{1-2}4142", "xyzAB", 2, 1},
		{"leading-unbounded-gap", "*", "*4142", "xxABAB", 5, 0},
		{"one-count-per-start", "*", "41*42", "AxyBB", 1, 0},
		{"alternative-lengths", "*", "(4142|414243)", "xxABCAB", 2, 2},
		{"internal-anchor", "*", "41??424344", "xxAzBCD", 1, 2},
		{"nibble", "*", "4?42", "xxABQB", 1, 2},
		{"folded-fullword", "*", "616263::if", "xABC ABC!", 1, 5},
		{"wide-and-ascii", "*", "4142::wa", "xxA\x00B\x00 AB", 2, 2},
		{"absolute", "2", "4142", "ABABAB", 1, 2},
		{"floating-inclusive", "1,3", "4142", "ABABAB", 2, 2},
		{"end-relative", "EOF-2", "4142", "ABABAB", 1, 4},
		{"absent", "*", "4344", "ABABAB", 0, 0xfffffffe},
		{"short-input", "*", "414243", "AB", 0, 0xfffffffe},
		{"empty-input", "*", "4142", "", 0, 0xfffffffe},
		{"offset-outside-input", "20", "4142", "AB", 0, 0xfffffffe},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := compileBody(tc.offset, tc.body)
			if err != nil {
				t.Fatal(err)
			}
			work := &matchWork{ctx: context.Background(), limit: 1_000_000}
			count, first, err := p.countWithOffset([]byte(tc.data), nil, work, len(tc.data)+1)
			if err != nil || count != tc.count || first != tc.first {
				t.Fatalf("countWithOffset = %d, %#x, %v; want %d, %#x, nil", count, first, err, tc.count, tc.first)
			}
			legacy := &matchWork{ctx: context.Background(), limit: 1_000_000}
			oldCount, err := p.count([]byte(tc.data), nil, legacy, len(tc.data)+1)
			if err != nil || oldCount != count || legacy.steps != work.steps {
				t.Fatalf("legacy count/work changed: count %d, steps %d, error %v; want count %d, steps %d", oldCount, legacy.steps, err, count, work.steps)
			}
		})
	}
}

func TestBytecodePatternCapStopsAtRequestedCount(t *testing.T) {
	p, err := compileBody("*", "41")
	if err != nil {
		t.Fatal(err)
	}
	for _, cap := range []int{-1, 0, 1, 2, 5} {
		t.Run(fmt.Sprint(cap), func(t *testing.T) {
			want := 4
			if cap > 0 && cap < want {
				want = cap
			}
			// Candidate searches and literal comparisons also consume work.
			// A positive cap must still stop after the requested match count.
			work := &matchWork{ctx: context.Background(), limit: 1_000_000}
			count, first, err := p.countWithOffset([]byte("AAAA"), nil, work, cap)
			if err != nil || count != want || first != 0 || work.steps == 0 {
				t.Fatalf("countWithOffset = %d, %#x, %v, %d steps; want %d, 0, nil with positive work", count, first, err, work.steps, want)
			}
		})
	}
	work := &matchWork{ctx: context.Background(), limit: 4}
	count, first, err := p.countWithOffset([]byte("AAAA"), nil, work, 5)
	if !errors.Is(err, ErrMatchLimit) || count != 0 || first != 0xfffffffe {
		t.Fatalf("partial count escaped work error: %d, %#x, %v", count, first, err)
	}
}

func TestBytecodePatternPEOffsets(t *testing.T) {
	data := testPE()
	copy(data[512:], "zABCDABCD")
	meta := parsePE(data)
	if meta == nil {
		t.Fatal("invalid PE fixture")
	}
	for _, tc := range []struct {
		offset, body string
		count        int
		first        uint32
	}{
		{"EP", "??41424344", 1, 512},
		{"EP+1", "41424344", 1, 513},
		{"EP-1,3", "41424344", 1, 513},
		{"S0+0,9", "41424344", 2, 513},
		{"SL+1", "41424344", 1, 513},
		{"SE0", "41424344", 2, 513},
		{"S1+0", "41424344", 0, 0xfffffffe},
	} {
		t.Run(tc.offset, func(t *testing.T) {
			p, err := compileBody(tc.offset, tc.body)
			if err != nil {
				t.Fatal(err)
			}
			work := &matchWork{ctx: context.Background(), limit: 1_000_000}
			count, first, err := p.countWithOffset(data, meta, work, len(data)+1)
			if err != nil || count != tc.count || first != tc.first {
				t.Fatalf("PE countWithOffset = %d, %#x, %v; want %d, %#x, nil", count, first, err, tc.count, tc.first)
			}
			work = &matchWork{ctx: context.Background(), limit: 1_000_000}
			count, first, err = p.countWithOffset(data, nil, work, len(data)+1)
			if err != nil || count != 0 || first != 0xfffffffe {
				t.Fatalf("PE offset accepted without PE metadata: %d, %#x, %v", count, first, err)
			}
		})
	}
}

func TestBytecodePatternContextAndWorkErrors(t *testing.T) {
	for _, body := range []string{"4142", "??4142", "{1-2}4142"} {
		t.Run(body, func(t *testing.T) {
			p, err := compileBody("*", body)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			// Match work checks cancellation at each 1024-step boundary.
			// Starting one step before it preserves the existing cadence.
			work := &matchWork{ctx: ctx, steps: 1023, limit: 1_000_000}
			count, first, err := p.countWithOffset([]byte("xxABAB"), nil, work, 7)
			if !errors.Is(err, context.Canceled) || count != 0 || first != 0xfffffffe {
				t.Fatalf("canceled countWithOffset = %d, %#x, %v", count, first, err)
			}
			work = &matchWork{ctx: context.Background(), limit: 0}
			count, first, err = p.countWithOffset([]byte("xxABAB"), nil, work, 7)
			if !errors.Is(err, ErrMatchLimit) || count != 0 || first != 0xfffffffe {
				t.Fatalf("exhausted countWithOffset = %d, %#x, %v", count, first, err)
			}
		})
	}
}
