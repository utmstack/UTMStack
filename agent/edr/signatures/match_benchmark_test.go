package signatures

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
)

// syntheticFeed is a large rule set shaped like the costly parts of the
// official feed: logical rules of random code literals, patterns anchored on
// zero runs and common opcodes, alternatives of two-byte literals, case
// insensitive and wide strings, and unbounded gaps. object is a 2 MiB
// executable-like object of zero padding, opcode-like bytes and strings that
// contains many of the rules' literals but few complete matches.
var syntheticFeed = sync.OnceValues(func() ([]byte, []byte) {
	r := rand.New(rand.NewSource(20260925))
	randomHex := func(n int) string {
		b := make([]byte, n)
		r.Read(b)
		return hex.EncodeToString(b)
	}
	opcodes := []byte{0x00, 0x00, 0x00, 0x8b, 0x89, 0x48, 0xe8, 0xff, 0x83, 0x45, 0x55, 0xec, 0x74, 0x75, 0x0f, 0x85, 0xc0, 0xcc, 0x90, 0x24, 0x6a}
	opcodeHex := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = opcodes[r.Intn(len(opcodes))]
		}
		return hex.EncodeToString(b)
	}
	words := make([]string, 400)
	for i := range words {
		w := make([]byte, 5+r.Intn(6))
		for j := range w {
			w[j] = byte('a' + r.Intn(26))
		}
		words[i] = string(w)
	}
	var ldb, ndb strings.Builder
	for i := 0; i < 60000; i++ {
		subs := 2 + r.Intn(3)
		fmt.Fprintf(&ldb, "Code.%d;Engine:81-255,Target:1;%s", i, []string{"0&1", "0&1&2", "0&1&2&3"}[subs-2])
		for j := 0; j < subs; j++ {
			fmt.Fprintf(&ldb, ";%s", randomHex(4+r.Intn(12)))
		}
		ldb.WriteString("\n")
	}
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&ldb, "Opcode.%d;Engine:81-255,Target:1;0>2&1;%s{-40}%s;%s\n", i, opcodeHex(3), opcodeHex(4), randomHex(6))
	}
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&ndb, "Virut.%d:1:*:e8??000000{-100}8602(f9|fc|66|eb|89|87){-15}%s\n", i, randomHex(3))
		fmt.Fprintf(&ndb, "Zero.%d:1:*:0000000000000000{2}%s\n", i, randomHex(3))
		fmt.Fprintf(&ndb, "Pad.%d:1:*:ffffffffffffffff{-8}%s??%s\n", i, opcodeHex(2), randomHex(2))
		fmt.Fprintf(&ndb, "Open.%d:0:*:%s*%s\n", i, randomHex(4), randomHex(4))
		fmt.Fprintf(&ndb, "Pair.%d:1:*:(%s|%s|%s)%s\n", i, randomHex(2), randomHex(2), randomHex(2), randomHex(2))
	}
	for i := 0; i < 20000; i++ {
		w := words[r.Intn(len(words))]
		fmt.Fprintf(&ldb, "Text.%d;Engine:81-255,Target:0;0&1;%s{-30}2e657865::i;%s\n", i, hex.EncodeToString([]byte(w)), randomHex(8))
		fmt.Fprintf(&ldb, "Wide.%d;Engine:81-255,Target:1;0&(1>3);%s::wa;%s\n", i, hex.EncodeToString([]byte(words[r.Intn(len(words))])), randomHex(5))
	}
	files := strings.Join([]string{"ldb\x00" + ldb.String(), "ndb\x00" + ndb.String()}, "\x01")

	object := append([]byte(nil), testPE()...)
	for len(object) < 2<<20 {
		switch r.Intn(10) {
		case 0:
			object = append(object, make([]byte, r.Intn(4096))...)
		case 1:
			w := words[r.Intn(len(words))]
			if r.Intn(2) == 0 {
				w = strings.ToUpper(w)
			}
			object = append(object, w+".exe "...)
		case 2:
			for _, c := range words[r.Intn(len(words))] {
				object = append(object, byte(c), 0)
			}
		case 3:
			object = append(object, 0xe8, byte(r.Intn(256)), 0, 0, 0, 0x86, 0x02, 0xf9, 0x86, 0x02)
		default:
			for n := r.Intn(512); n > 0; n-- {
				object = append(object, opcodes[r.Intn(len(opcodes))])
			}
		}
	}
	return []byte(files), object[:2<<20]
})

func compileSyntheticFeed(tb testing.TB) (*Database, []byte) {
	files, object := syntheticFeed()
	input := map[string][]byte{}
	for _, f := range strings.Split(string(files), "\x01") {
		ext, text, _ := strings.Cut(f, "\x00")
		input["synthetic."+ext] = []byte(text)
	}
	d, err := Compile(input)
	if err != nil {
		tb.Fatal(err)
	}
	return d, object
}

// TestSyntheticFeedWorkRegression guards the shared anchor pass: on a large
// feed-shaped rule set it must agree with the previous matcher and stay far
// below its work. The previous matcher needs about 37 billion steps on the
// whole object, so it is compared on a prefix.
func TestSyntheticFeedWorkRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("large rule set")
	}
	d, object := compileSyntheticFeed(t)
	_, steps, err := d.matchCounted(context.Background(), object, 1<<40)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d rules, %d patterns: %d steps on %d bytes (%.1f per byte)", len(d.rules), len(d.patterns), steps, len(object), float64(steps)/float64(len(object)))
	if steps > 40*int64(len(object)) {
		t.Fatalf("%d steps on %d bytes", steps, len(object))
	}
	prefix := object[:96<<10]
	name, steps, err := d.matchCounted(context.Background(), prefix, 1<<40)
	oldName, oldSteps, oldErr := newLegacyMatcher(d).match(context.Background(), prefix, 1<<40)
	if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
		t.Fatalf("Match = %q, %v; previous matcher %q, %v", name, err, oldName, oldErr)
	}
	t.Logf("first %d bytes: %d steps, previous matcher %d", len(prefix), steps, oldSteps)
	if steps*5 > oldSteps {
		t.Fatalf("%d steps; previous matcher %d", steps, oldSteps)
	}
}

func BenchmarkMatchSyntheticFeed(b *testing.B) {
	d, object := compileSyntheticFeed(b)
	b.SetBytes(int64(len(object)))
	b.ResetTimer()
	var steps int64
	for i := 0; i < b.N; i++ {
		_, n, err := d.matchCounted(context.Background(), object, 1<<40)
		if err != nil {
			b.Fatal(err)
		}
		steps = n
	}
	b.ReportMetric(float64(steps), "steps/op")
}

// errAfter is a context that reports cancellation after a number of checks,
// so cancellation is observed in the middle of the anchor pass or counting.
type errAfter struct {
	context.Context
	checks, calls int
}

func (c *errAfter) Err() error {
	if c.calls++; c.calls > c.checks {
		return context.Canceled
	}
	return nil
}

func TestCancellationStopsMatching(t *testing.T) {
	d := compileTest(t, "ldb", "Late;Target:0;0&1;41414141;42424242{-2000}43434343")
	data := make([]byte, 8<<20)
	for i := 0; i < len(data); i += 1000 {
		copy(data[i:], "AAAA..BBBB")
	}
	all := &errAfter{Context: context.Background(), checks: 1 << 30}
	if name, _, err := d.matchCounted(all, data, 1<<40); name != "" || err != nil {
		t.Fatalf("Match = %q, %v", name, err)
	}
	scan := &errAfter{Context: context.Background(), checks: 1 << 30}
	if _, err := d.newPatternCounts(data, &objectInfo{}, &matchWork{ctx: scan, limit: 1 << 40}); err != nil {
		t.Fatal(err)
	}
	if all.calls < scan.calls+1000 {
		t.Fatalf("%d cancellation checks in all, %d in the anchor pass", all.calls, scan.calls)
	}
	// Cancel before, during the anchor pass, and while counting.
	for _, checks := range []int{0, 1, scan.calls / 2, scan.calls + 10, (scan.calls + all.calls) / 2, all.calls - 2} {
		name, _, err := d.matchCounted(&errAfter{Context: context.Background(), checks: checks}, data, 1<<40)
		if name != "" || !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled after %d of %d checks: Match = %q, %v", checks, all.calls, name, err)
		}
	}
}
