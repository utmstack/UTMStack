package signatures

import (
	"context"
	"errors"
	"testing"
	"unsafe"
)

func TestGraphOnlyStorageDoesNotExpandLiteralPatterns(t *testing.T) {
	if got := unsafe.Sizeof(bodyPattern{}); got != 64 {
		t.Fatalf("body pattern occupies %d bytes, want 64", got)
	}
	for _, tc := range []struct {
		body      string
		wantGraph bool
	}{
		{"414243", false},
		{"4142??43", true},
		{"41(42|43)44", true},
	} {
		p, err := compileBody("*", tc.body)
		if err != nil {
			t.Fatal(err)
		}
		if (p.graph != nil) != tc.wantGraph {
			t.Fatalf("%s graph = %t, want %t", tc.body, p.graph != nil, tc.wantGraph)
		}
	}
	folded, err := compileBody("*", "414243::i")
	if err != nil {
		t.Fatal(err)
	}
	a := folded.compiledAnchor(0)
	if folded.graph != nil || folded.compiledAnchorCount() != 1 || !a.fold || string(a.literal) != "abc" {
		t.Fatalf("simple folded anchor = %+v", a)
	}
}

// A literal-only pattern is stored without graph nodes. Its result and work
// charges must still agree with the graph evaluator at every start and budget.
func TestSimpleLiteralMatchesGraph(t *testing.T) {
	for _, tc := range []struct {
		body string
		data string
	}{
		{"414243", "ABC AB ABCABC"},
		{"414243::i", "abc AB aBcABC"},
		{"414243::w", "A\x00B\x00C\x00 AB\x00C\x00"},
		{"414243::a", "ABC AB ABCABC"},
	} {
		p, err := compileBody("*", tc.body)
		if err != nil {
			t.Fatalf("%s: %v", tc.body, err)
		}
		if !p.simple || p.graph != nil || len(p.prefix) != p.minWidth() || p.longestPart() != p.minWidth() {
			t.Fatalf("%s: literal not stored compactly: %+v", tc.body, p)
		}
		for start := -1; start <= len(tc.data)+1; start++ {
			for budget := int64(0); budget <= int64(len(p.prefix)+3); budget++ {
				w := &matchWork{ctx: context.Background(), limit: budget}
				got, gotErr := p.matchesAt([]byte(tc.data), start, w)
				old := &matchWork{ctx: context.Background(), limit: budget}
				want, wantErr := legacyMatchesAt(&p, []byte(tc.data), start, old)
				if got != want || !errors.Is(gotErr, wantErr) || !errors.Is(wantErr, gotErr) || w.steps != old.steps {
					t.Fatalf("%s at %d, budget %d: compact=(%v,%v,%d), graph=(%v,%v,%d)",
						tc.body, start, budget, got, gotErr, w.steps, want, wantErr, old.steps)
				}
			}
		}
	}
}

func TestSimpleLiteralWithZeroStateLimitUsesGraph(t *testing.T) {
	before := maxPatternStates
	maxPatternStates = 0
	defer func() { maxPatternStates = before }()
	p, err := compileBody("*", "414243")
	if err != nil {
		t.Fatal(err)
	}
	if p.simple || p.graph == nil || len(p.graph.nodes) != 2 {
		t.Fatal("zero state limit must retain the graph's limit behavior")
	}
	_, err = p.matchesAt([]byte("ABC"), 0, &matchWork{ctx: context.Background(), limit: 100})
	if !errors.Is(err, errPatternStates) {
		t.Fatalf("match error = %v, want pattern state limit", err)
	}
}
