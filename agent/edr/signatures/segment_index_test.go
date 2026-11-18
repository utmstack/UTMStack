package signatures

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestRuleRefreshRetiresUnreferencedPatternBodiesAndRecompilesReturn(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("Old.Rule:0:*:414243444546\n"),
		"b.ndb": []byte("Kept.Rule:0:*:4d4e4f505152\n"),
	}
	first, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	oldID := first.rules[0].patterns[0]
	if reflect.DeepEqual(first.patterns[oldID], bodyPattern{}) {
		t.Fatal("initial pattern has no compiled body")
	}
	files["a.ndb"] = []byte("New.Rule:0:*:4748494a4b4c\n")
	second, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: first})
	if err != nil || !handled {
		t.Fatalf("first rule refresh: handled=%v err=%v", handled, err)
	}
	if !second.retiredPatterns.has(oldID) || !reflect.DeepEqual(second.patterns[oldID], bodyPattern{}) {
		t.Fatal("inactive pattern body was retained")
	}
	if reflect.DeepEqual(first.patterns[oldID], bodyPattern{}) {
		t.Fatal("prior generation pattern body was mutated")
	}
	keptID := first.rules[1].patterns[0]
	if second.retiredPatterns.has(keptID) {
		t.Fatal("unchanged rule pattern was retired")
	}
	files["a.ndb"] = []byte("Old.Returned:0:*:414243444546\n")
	third, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: second})
	if err != nil || !handled {
		t.Fatalf("returning rule refresh: handled=%v err=%v", handled, err)
	}
	if third.rules[0].patterns[0] != oldID || third.retiredPatterns.has(oldID) ||
		reflect.DeepEqual(third.patterns[oldID], bodyPattern{}) {
		t.Fatal("returning body did not safely reuse the retired pattern slot")
	}
	if !second.retiredPatterns.has(oldID) || !reflect.DeepEqual(second.patterns[oldID], bodyPattern{}) {
		t.Fatal("reusing a slot mutated the previous generation")
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(third.Report(), full.Report()) {
		t.Fatal("returning body refresh differs from full compiler")
	}
	assertMatch(t, first, []byte("ABCDEF"), "Old.Rule")
	assertMatch(t, second, []byte("ABCDEF"), "")
	assertMatch(t, third, []byte("ABCDEF"), "Old.Returned")
	assertMatch(t, third, []byte("MNOPQR"), "Kept.Rule")
}

func TestRuleRefreshKeepsSharedAndFileTypePatternsActive(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("First.Rule:0:*:414243444546\n"),
		"b.ndb": []byte("Second.Rule:0:*:414243444546\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	shared := prior.rules[0].patterns[0]
	if shared != prior.rules[1].patterns[0] || len(prior.ftypes.patterns) == 0 {
		t.Fatal("fixture lacks shared rule and built-in file-type patterns")
	}
	files["a.ndb"] = []byte("First.New:0:*:4748494a4b4c\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("shared-pattern refresh: handled=%v err=%v", handled, err)
	}
	if next.retiredPatterns.has(shared) {
		t.Fatal("pattern still used by another rule was retired")
	}
	for _, p := range prior.ftypes.patterns {
		if next.retiredPatterns.has(p.id) {
			t.Fatalf("file-type pattern %d was retired", p.id)
		}
	}
	assertMatch(t, next, []byte("ABCDEF"), "Second.Rule")
}

func TestRepeatedRuleRefreshReusesSlotsWithoutAccumulatingPatternBodies(t *testing.T) {
	marker := func(n int) []byte { return []byte(fmt.Sprintf("R%05d", n)) }
	rule := func(n int) []byte { return []byte(fmt.Sprintf("Round.%d:0:*:%x\n", n, marker(n))) }
	files := map[string][]byte{
		"a.ndb": rule(0),
		"b.ndb": []byte("Kept.Rule:0:*:4d4e4f505152\n"),
	}
	previous, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	initialSlots := len(previous.patterns)
	for round := 1; round <= 64; round++ {
		files["a.ndb"] = rule(round)
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: previous})
		if err != nil || !handled {
			t.Fatalf("round %d refresh: handled=%v err=%v", round, handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if len(next.patterns) > initialSlots+1 || !reflect.DeepEqual(next.Report(), full.Report()) {
			t.Fatalf("round %d accumulated %d pattern slots (initial %d)", round, len(next.patterns), initialSlots)
		}
		assertMatch(t, previous, marker(round-1), fmt.Sprintf("Round.%d", round-1))
		assertMatch(t, next, marker(round-1), "")
		assertMatch(t, next, marker(round), fmt.Sprintf("Round.%d", round))
		assertMatch(t, next, []byte("MNOPQR"), "Kept.Rule")
		previous = next
	}
}

func TestRuleRefreshReusesMultipleSlotsAndDeduplicatesPendingPatterns(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("A.One:0:*:414141414141\nA.Two:0:*:424242424242\n"),
		"b.ndb": []byte("B.Kept:0:*:4d4e4f505152\n"),
	}
	initial, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("Mid.One:0:*:434343434343\nMid.Two:0:*:444444444444\n")
	middle, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: initial})
	if err != nil || !handled {
		t.Fatalf("middle refresh: handled=%v err=%v", handled, err)
	}
	files["a.ndb"] = []byte("Final.One:0:*:454545454545\nFinal.Two:0:*:464646464646\nFinal.Shared:0:*:454545454545\n")
	final, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: middle})
	if err != nil || !handled {
		t.Fatalf("final refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.patterns) != len(middle.patterns) || final.rules[0].patterns[0] != final.rules[2].patterns[0] ||
		!reflect.DeepEqual(final.Report(), full.Report()) {
		t.Fatal("multiple retired slots were not reused with correct deduplication")
	}
	for _, data := range [][]byte{[]byte("EEEEEE"), []byte("FFFFFF"), []byte("CCCCCC"), []byte("DDDDDD"), []byte("MNOPQR"), []byte("ordinary clean bytes")} {
		got, gotWork, gotErr := final.matchCounted(context.Background(), data, 10_000)
		want, wantWork, wantErr := full.matchCounted(context.Background(), data, 10_000)
		if got != want || gotWork != wantWork || (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("%q: incremental=(%q,%d,%v) full=(%q,%d,%v)", data, got, gotWork, gotErr, want, wantWork, wantErr)
		}
	}
	assertMatch(t, middle, []byte("CCCCCC"), "Mid.One")
	assertMatch(t, final, []byte("CCCCCC"), "")
	// Both equal literals use the same EEE bucket; the later entry wins.
	assertMatch(t, final, []byte("EEEEEE"), "Final.Shared")
}

func TestRebuiltAnchorIndexPreservesMatcherAndWork(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("A.One:0:*:414243444546\nA.Two:0:*:414243{1-3}4546\n"),
		"b.ndb": []byte("B.One:0:*:4748494a4b4c\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	next := *d
	if err := next.rebuildAnchorIndex(len(d.patterns), next.activePatternsForRules(next.rules)); err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{
		[]byte("ABCDEF"), []byte("ABCDZZEF"), []byte("GHIJKL"),
		[]byte("ordinary harmless content"), []byte("%PDF-1.7\nordinary content"),
	} {
		before, beforeSteps, beforeErr := d.matchCounted(context.Background(), data, 10_000_000)
		after, afterSteps, afterErr := next.matchCounted(context.Background(), data, 10_000_000)
		if before != after || beforeSteps != afterSteps || (beforeErr == nil) != (afterErr == nil) {
			t.Fatalf("%q: before=(%q,%d,%v) after=(%q,%d,%v)", data, before, beforeSteps, beforeErr, after, afterSteps, afterErr)
		}
		beforeType, _ := d.DetermineType(data)
		afterType, _ := next.DetermineType(data)
		if beforeType != afterType {
			t.Fatalf("%q: type %v changed to %v", data, beforeType, afterType)
		}
	}
}

func TestRebuiltAnchorIndexExcludesRemovedSourceRules(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("A.Old:0:*:414243444546\n"),
		"b.ndb": []byte("B.Kept:0:*:4748494a4b4c\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	next := *d
	rangeB := d.sourceStats["b.ndb"]
	next.rules = append([]rule(nil), d.rules[rangeB.ruleStart:rangeB.ruleEnd]...)
	if err := next.rebuildAnchorIndex(len(d.patterns), next.activePatternsForRules(next.rules)); err != nil {
		t.Fatal(err)
	}
	next.buildRuleIndex()
	if len(next.anchorGroups) >= len(d.anchorGroups) {
		t.Fatal("removed source anchor was retained")
	}
	assertMatch(t, d, []byte("ABCDEF"), "A.Old")
	assertMatch(t, &next, []byte("ABCDEF"), "")
	assertMatch(t, &next, []byte("GHIJKL"), "B.Kept")
}

func TestRebuiltAnchorIndexAddsFreshPatternWithoutMutatingPrior(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("A.Old:0:*:414243444546\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	next := *d
	next.rules = append([]rule(nil), d.rules...)
	next.patterns = append([]bodyPattern(nil), d.patterns...)
	next.patternKeys = append([][16]byte(nil), d.patternKeys...)
	next.patternIDs = make(map[[16]byte]int, len(d.patternKeys))
	for id, key := range d.patternKeys {
		next.patternIDs[key] = id
	}
	if err := next.addExtended("B.New:0:*:4748494a4b4c", "B.New:0:*:4748494a4b4c"); err != nil {
		t.Fatal(err)
	}
	if err := next.rebuildAnchorIndex(len(d.patterns), next.activePatternsForRules(next.rules)); err != nil {
		t.Fatal(err)
	}
	next.buildRuleIndex()
	assertMatch(t, d, []byte("GHIJKL"), "")
	assertMatch(t, &next, []byte("GHIJKL"), "B.New")
	assertMatch(t, d, []byte("ABCDEF"), "A.Old")
	assertMatch(t, &next, []byte("ABCDEF"), "A.Old")
}
