package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"testing"
)

// The forward search must answer exactly what the backtracking search
// answers, at every start of every object: it replaces it only when the
// backtracking memory bound is reached.
func TestForwardSearchAgreesWithBacktracking(t *testing.T) {
	defer func(states int) { maxPatternStates = states }(maxPatternStates)
	maxPatternStates = 1 << 30 // the backtracking search never falls back here
	r := rand.New(rand.NewSource(20260927))
	g := &ruleGen{r: r}
	alphabets := [][]byte{[]byte("aAbB.1\x00mMzZ"), []byte("\x00\x00\x00\xff\x90\xcc\x8b\x45\xe8\x55\x89\xecABab"), []byte("AAB\x00")}
	patterns, starts, matches := 0, 0, 0
	for round := 0; round < 3000; round++ {
		g.gappy = round%2 == 0
		g.alphabet = alphabets[round%len(alphabets)]
		sub := g.subsignature(false)
		d, err := Compile(map[string][]byte{"t.ldb": []byte("Fwd;Target:0;0;" + sub.text)})
		if err != nil || len(d.patterns) == 0 || d.patterns[0].holder != nil {
			continue
		}
		p := &d.patterns[0]
		patterns++
		data := g.object([]genSub{sub}, []int{0, 20, 80, 300}[r.Intn(4)])
		for start := 0; start <= len(data); start++ {
			want, err := p.matchesAt(data, start, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if err != nil {
				t.Fatalf("backtracking %q at %d: %v", sub.text, start, err)
			}
			got, err := p.forwardMatchesAt(data, start, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if err != nil || got != want {
				t.Fatalf("%q on %q at %d: forward %v, %v; backtracking %v", sub.text, data, start, got, err, want)
			}
			starts++
			if want {
				matches++
			}
		}
	}
	t.Logf("%d patterns, %d starts, %d matches", patterns, starts, matches)
	if patterns < 1000 || matches < 1000 {
		t.Fatalf("too little coverage: %d patterns, %d matches", patterns, matches)
	}
}

// A pattern with a long gap anchored at the start of the object (as the
// official Win.Packed.Gandcrab-6552923-4) exhausts the backtracking memory on
// objects with long zero runs; the forward search then decides it.
func TestLongGapPatternIsDecided(t *testing.T) {
	defer func(states int) { maxPatternStates = states }(maxPatternStates)
	maxPatternStates = 50
	rule := "Gap.Long;Target:0;0;0:4d5a{-5000}????0080??000080{-255}0000000000000100!(00)000000{-400}!(0000)01"
	d := compileTest(t, "ldb", rule)
	// Header-like records every 16 bytes: each one begins the part after the
	// long gap, so the backtracking search runs out of memory on this object.
	busy := append([]byte("MZ"), make([]byte, 6000)...)
	for off := 100; off < 3000; off += 16 {
		copy(busy[off:], []byte{0, 0, 0, 0x80, 0, 0, 0, 0x80})
	}
	// The later parts occur too, but never in order.
	copy(busy[2000:], []byte{0, 0, 0, 0, 0, 0, 1, 0, 0, 0x55, 0x55, 1})
	name, err := d.Match(context.Background(), busy)
	if name != "" || err != nil {
		t.Fatalf("busy: (%q, %v), want clean", name, err)
	}
	planted := bytes.Clone(busy)
	copy(planted[3100:], []byte{0, 0, 0, 0, 0, 0, 1, 0, 7, 0, 0, 0, 0x55, 0x55, 1})
	if name, err := d.Match(context.Background(), planted); name != "Gap.Long" || err != nil {
		t.Fatalf("planted: (%q, %v), want the detection", name, err)
	}
	// The forward search has its own bound: past it the pattern stays
	// undecided.
	defer func(steps int) { maxForwardSteps = steps }(maxForwardSteps)
	maxForwardSteps = 100
	if name, err := d.Match(context.Background(), busy); name != "" || !errors.Is(err, errPatternStates) || !errors.Is(err, ErrRuleIncomplete) {
		t.Fatalf("forward bound: (%q, %v), want the state limit", name, err)
	}
	if !strings.Contains(errPatternStates.Error(), "pattern state limit") {
		t.Fatal("reason text changed")
	}
}

// Position sets receive additions in any order (a node reached from several
// earlier nodes); reading them must give the sorted, merged ranges. Merging
// on every out-of-order addition cost a 1.3 MB Edge language pack two
// minutes.
func TestForwardRangesMergeInAnyOrder(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for round := 0; round < 2000; round++ {
		var s posRanges
		var want [200]bool
		for k := r.Intn(40); k > 0; k-- {
			lo := r.Intn(190)
			hi := lo + r.Intn(8)
			s.add(lo, hi)
			for p := lo; p <= hi; p++ {
				want[p] = true
			}
		}
		var got [200]bool
		prev := -2
		for _, x := range s.ranges() {
			if x[0] <= prev+1 || x[1] < x[0] {
				t.Fatalf("ranges not sorted, disjoint and apart: %v", s.ranges())
			}
			prev = x[1]
			for p := x[0]; p <= x[1]; p++ {
				got[p] = true
			}
		}
		if got != want {
			t.Fatalf("round %d: ranges %v", round, s.ranges())
		}
	}
}

// The forward search runs only for a rule whose result depends on the
// undecided start: here the other subsignature is absent, so the rule is
// false whatever the long-gap pattern gives, and the object costs a small
// part of what deciding the pattern costs.
func TestForwardSearchOnlyWhenTheRuleDependsOnIt(t *testing.T) {
	defer func(states int) { maxPatternStates = states }(maxPatternStates)
	maxPatternStates = 50
	gap := "0:4d5a{-5000}????0080??000080{-255}0000000000000100!(00)000000{-400}!(0000)01"
	busy := append([]byte("MZ"), make([]byte, 6000)...)
	for off := 100; off < 3000; off += 16 {
		copy(busy[off:], []byte{0, 0, 0, 0x80, 0, 0, 0, 0x80})
	}
	copy(busy[2000:], []byte{0, 0, 0, 0, 0, 0, 1, 0, 0, 0x55, 0x55, 1})
	alone := compileTest(t, "ldb", "Gap.Alone;Target:0;0;"+gap)
	paired := compileTest(t, "ldb", "Gap.Paired;Target:0;0&1;"+gap+";"+hex.EncodeToString([]byte("absent marker")))
	_, decided, err := alone.matchCounted(context.Background(), busy, 1<<40)
	if err != nil {
		t.Fatal(err)
	}
	// The same pattern with a forward search that stops at once.
	defer func(steps int) { maxForwardSteps = steps }(maxForwardSteps)
	maxForwardSteps = 0
	_, undecided, _ := alone.matchCounted(context.Background(), busy, 1<<40)
	maxForwardSteps = 1 << 22
	name, pairedSteps, err := paired.matchCounted(context.Background(), busy, 1<<40)
	if name != "" || err != nil {
		t.Fatalf("paired: (%q, %v), want clean", name, err)
	}
	if decided < undecided+5000 || pairedSteps > undecided+100 {
		t.Fatalf("steps: pattern decided %d, left undecided %d, paired rule %d: the forward search ran for the paired rule", decided, undecided, pairedSteps)
	}
}
