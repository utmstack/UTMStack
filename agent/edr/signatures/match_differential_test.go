package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

// ruleGen writes random subsignatures together with a way to produce bytes
// that match them, so generated objects contain real matches, near misses and
// repeated anchors.
type ruleGen struct {
	r        *rand.Rand
	alphabet []byte
	// gappy favors several gaps and alternatives per subsignature, which with
	// a small alphabet makes starts explore (and exhaust) many states.
	gappy bool
}

func (g *ruleGen) literal(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = g.alphabet[g.r.Intn(len(g.alphabet))]
	}
	return b
}

func (g *ruleGen) randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		if g.r.Intn(3) == 0 {
			b[i] = byte(g.r.Intn(256))
		} else {
			b[i] = g.alphabet[g.r.Intn(len(g.alphabet))]
		}
	}
	return b
}

type genPart struct {
	text string
	make func() []byte
}

// sequence returns a body and a generator of bytes it matches.
func (g *ruleGen) sequence(depth, parts int, literalOnly bool) genPart {
	var text strings.Builder
	var makers []func() []byte
	addLiteral := func(n int) {
		b := g.literal(n)
		text.WriteString(hex.EncodeToString(b))
		makers = append(makers, func() []byte { return b })
	}
	if literalOnly {
		addLiteral(1 + g.r.Intn(5))
		return genPart{text.String(), makers[0]}
	}
	addLiteral(1 + g.r.Intn(4))
	for i := 0; i < parts; i++ {
		choice := g.r.Intn(12)
		if g.gappy && choice >= 9 {
			choice = []int{1, 2, 3, 4, 7}[g.r.Intn(5)]
		}
		switch choice {
		case 0:
			n := g.r.Intn(4)
			fmt.Fprintf(&text, "{%d}", n)
			makers = append(makers, func() []byte { return g.randomBytes(n) })
		case 1:
			lo := g.r.Intn(3)
			hi := lo + 1 + g.r.Intn(6)
			fmt.Fprintf(&text, "{%d-%d}", lo, hi)
			makers = append(makers, func() []byte { return g.randomBytes(lo + g.r.Intn(hi-lo+1)) })
		case 2:
			hi := 1 + g.r.Intn(8)
			fmt.Fprintf(&text, "{-%d}", hi)
			makers = append(makers, func() []byte { return g.randomBytes(g.r.Intn(hi + 1)) })
		case 3:
			lo := g.r.Intn(3)
			fmt.Fprintf(&text, "{%d-}", lo)
			makers = append(makers, func() []byte { return g.randomBytes(lo + g.r.Intn(12)) })
		case 4:
			text.WriteString("*")
			makers = append(makers, func() []byte { return g.randomBytes(g.r.Intn(12)) })
		case 5:
			text.WriteString("??")
			makers = append(makers, func() []byte { return g.randomBytes(1) })
		case 6:
			c := g.alphabet[g.r.Intn(len(g.alphabet))]
			if g.r.Intn(2) == 0 {
				fmt.Fprintf(&text, "%x?", c>>4)
				makers = append(makers, func() []byte { return []byte{c&0xf0 | byte(g.r.Intn(16))} })
			} else {
				fmt.Fprintf(&text, "?%x", c&15)
				makers = append(makers, func() []byte { return []byte{c&0x0f | byte(g.r.Intn(16))<<4} })
			}
		case 7, 8:
			if depth >= 2 {
				addLiteral(1 + g.r.Intn(3))
				continue
			}
			var branches []genPart
			for k := 2 + g.r.Intn(3); k > 0; k-- {
				branches = append(branches, g.sequence(depth+1, g.r.Intn(2), false))
			}
			text.WriteString("(")
			for k, b := range branches {
				if k > 0 {
					text.WriteString("|")
				}
				text.WriteString(b.text)
			}
			text.WriteString(")")
			makers = append(makers, func() []byte { return branches[g.r.Intn(len(branches))].make() })
		case 9:
			n := 1 + g.r.Intn(2)
			a, b := g.literal(n), g.literal(n)
			fmt.Fprintf(&text, "!(%x|%x)", a, b)
			makers = append(makers, func() []byte {
				for {
					x := g.randomBytes(n)
					if !bytes.EqualFold(x, a) && !bytes.EqualFold(x, b) {
						return x
					}
				}
			})
		default:
			addLiteral(1 + g.r.Intn(5))
		}
	}
	return genPart{text.String(), func() []byte {
		var out []byte
		for _, m := range makers {
			out = append(out, m()...)
		}
		return out
	}}
}

// genSub is one generated subsignature: its text and a generator of objects
// fragments it matches.
type genSub struct {
	text string
	make func() []byte
}

func (g *ruleGen) subsignature(target1 bool) genSub {
	if g.r.Intn(16) == 0 {
		return g.unanchored(target1)
	}
	mods := []string{"", "", "", "", "::i", "::f", "::if", "::a", "::w", "::wa", "::wi", "::wia"}[g.r.Intn(12)]
	wide := strings.Contains(mods, "w")
	parts := g.r.Intn(4)
	if g.gappy {
		parts = 2 + g.r.Intn(4)
	}
	part := g.sequence(0, parts, wide)
	offsets := []string{"", "", "", "", "0:", "3:", "2,40:", "EOF-40:", "10,200:"}
	if target1 {
		offsets = append(offsets, "EP+0:", "EP+2,30:", "S0+0:", "SL+4:", "SE0:", "EP-2:")
	}
	offset := offsets[g.r.Intn(len(offsets))]
	body := part.text + mods
	make := func() []byte {
		b := part.make()
		if strings.Contains(mods, "i") {
			b = bytes.Clone(b)
			for i, c := range b {
				if asciiLetter(c) && g.r.Intn(2) == 0 {
					b[i] = c ^ 0x20
				}
			}
		}
		if wide && (!strings.Contains(mods, "a") || g.r.Intn(2) == 0) {
			var w []byte
			for _, c := range b {
				w = append(w, c, 0)
			}
			b = w
		}
		if strings.Contains(mods, "f") {
			b = append(append([]byte{' '}, b...), ' ')
		}
		return b
	}
	return genSub{offset + body, make}
}

// unanchored returns a subsignature of nibbles only: it has no literal to
// anchor on, so it is a candidate on every object and its rules must be
// evaluated everywhere.
func (g *ruleGen) unanchored(target1 bool) genSub {
	var text strings.Builder
	var makers []func() []byte
	for n := 2 + g.r.Intn(3); n > 0; n-- {
		c := g.alphabet[g.r.Intn(len(g.alphabet))]
		if g.r.Intn(2) == 0 {
			fmt.Fprintf(&text, "%x?", c>>4)
			makers = append(makers, func() []byte { return []byte{c&0xf0 | byte(g.r.Intn(16))} })
		} else {
			fmt.Fprintf(&text, "?%x", c&15)
			makers = append(makers, func() []byte { return []byte{c&0x0f | byte(g.r.Intn(16))<<4} })
		}
	}
	offsets := []string{"", "", "0:", "EOF-40:", "10,200:"}
	if target1 {
		offsets = append(offsets, "EP+0:", "S0+0:")
	}
	return genSub{offsets[g.r.Intn(len(offsets))] + text.String(), func() []byte {
		var out []byte
		for _, m := range makers {
			out = append(out, m()...)
		}
		return out
	}}
}

var genExpressions = [][]string{
	{"0", "0", "0>1", "0=2", "0<2", "0=0", "0>3", "0=1"},
	{"0&1", "0|1", "(0|1)>2", "(0|1)>1,2", "0>1&1", "0|1=0", "(0|1)=3", "0&1>2"},
	{"0&1&2", "0|1&2", "(0&1)|2", "(0|1|2)>2,2", "0&(1|2)", "(0|1)>1&2<2", "(0|1|2)>4"},
}

// genZeroExpressions hold with every count zero, so they must be evaluated
// on objects that contain none of their subsignatures. They are drawn rarely:
// such a rule matches most objects and hides the rules after it.
var genZeroExpressions = [][]string{
	{"0<1", "0<3"},
	{"(0|1)<2", "0=0&1", "0=0&1=0", "0<2|1"},
	{"(0|1|2)<2", "(0|1)<1&2=0", "0|1|2=0", "0&1|2=0"},
}

// genRuleSet returns the lines of one random database and a pool of
// subsignatures to plant in objects. Lines the compiler rejects are dropped.
func (g *ruleGen) ruleSet() ([]string, []genSub) {
	pool := make([]genSub, 4+g.r.Intn(8))
	target1 := make([]bool, len(pool))
	for i := range pool {
		target1[i] = g.r.Intn(3) == 0
		pool[i] = g.subsignature(target1[i])
	}
	var lines []string
	for i := 0; i < 6+g.r.Intn(14); i++ {
		n := 1 + g.r.Intn(3)
		var subs []int
		t1 := false
		for j := 0; j < n; j++ {
			k := g.r.Intn(len(pool))
			subs = append(subs, k)
			t1 = t1 || target1[k]
		}
		target := 0
		if t1 || g.r.Intn(4) == 0 {
			target = 1
		}
		var line string
		if n == 1 && g.r.Intn(3) == 0 {
			off, body := "*", pool[subs[0]].text
			if k := strings.IndexByte(body, ':'); k >= 0 && (k+1 == len(body) || body[k+1] != ':') {
				off, body = body[:k], body[k+1:]
			}
			line = fmt.Sprintf("N%d:%d:%s:%s", i, target, off, body)
			if _, err := Compile(map[string][]byte{"one.ndb": []byte(line)}); err == nil {
				lines = append(lines, "ndb\t"+line)
			}
			continue
		}
		expressions := genExpressions[n-1]
		if g.r.Intn(12) == 0 {
			expressions = genZeroExpressions[n-1]
		}
		// Size and header descriptors filter rules before and after their
		// expression; the size filter also decides which rule unknown on an
		// unusable header stands in for the others.
		attributes := ""
		if g.r.Intn(4) == 0 {
			attributes += []string{",FileSize:0-100", ",FileSize:8-8", ",FileSize:50-2000", ",FileSize:1000-100000", ",FileSize:0-0"}[g.r.Intn(5)]
		}
		if target == 1 && g.r.Intn(4) == 0 {
			attributes += []string{",EntryPoint:512-512", ",EntryPoint:0-100", ",NumberOfSections:1-1", ",NumberOfSections:2-3", ",EntryPoint:0-4096,NumberOfSections:1-2"}[g.r.Intn(5)]
		}
		line = fmt.Sprintf("L%d;Engine:81-255,Target:%d%s;%s", i, target, attributes, expressions[g.r.Intn(len(expressions))])
		for _, k := range subs {
			line += ";" + pool[k].text
		}
		if _, err := Compile(map[string][]byte{"one.ldb": []byte(line)}); err == nil {
			lines = append(lines, "ldb\t"+line)
		}
	}
	return lines, pool
}

func compileGenerated(lines []string) (*Database, error) {
	files := map[string][]byte{}
	for _, l := range lines {
		ext, line, _ := strings.Cut(l, "\t")
		files["gen."+ext] = append(files["gen."+ext], line+"\n"...)
	}
	return Compile(files)
}

// object builds one object: random filler, periodic runs, and planted
// fragments of the pool, sometimes behind an executable header so target-1
// rules and executable offsets apply.
func (g *ruleGen) object(pool []genSub, size int) []byte {
	var data []byte
	switch g.r.Intn(8) {
	case 0, 1, 2:
		data = append(data, testPE()...)
		if g.r.Intn(2) == 0 {
			// Plant at the entry point, first section and end of the header.
			copy(data[512:], pool[g.r.Intn(len(pool))].make())
		}
	case 3:
		// An MZ object whose header cannot be modeled: executable offsets
		// and descriptors are unknown on it.
		data = append(data, "MZ"...)
	}
	for len(data) < size {
		switch g.r.Intn(6) {
		case 0, 1:
			data = append(data, pool[g.r.Intn(len(pool))].make()...)
		case 2:
			f := pool[g.r.Intn(len(pool))].make()
			if len(f) > 1 {
				f = f[:g.r.Intn(len(f))] // a near miss
			}
			data = append(data, f...)
		case 3:
			period := g.randomBytes(1 + g.r.Intn(3))
			for n := g.r.Intn(200); n > 0; n-- {
				data = append(data, period...)
			}
		default:
			data = append(data, g.randomBytes(1+g.r.Intn(60))...)
		}
	}
	return data
}

type diffStats struct {
	objects, detected, incomplete, limited, windowed, spanned int
	// lazy counts rule evaluations that left a count past the memory bound
	// undecided because the result did not depend on it.
	lazy int
	// windowedLimitable counts windowed counts of patterns a state limit can
	// reach, which keep their longest-literal anchors; windowedBelowLimit
	// those where only the object length kept a skipped start from reaching it.
	windowedLimitable, windowedBelowLimit int
	// zeroLive counts evaluations of applicable rules that none of their
	// subsignatures reached but whose expression is true or unknown.
	zeroLive int
	// standIn counts rules left out that are unknown on an unusable header
	// behind a reached rule reporting that cause.
	standIn int
	// filtered counts generated rules with a size or header descriptor.
	filtered int
	// unanchored counts generated rules with a subsignature without anchors.
	unanchored            int
	fasterOrEqual, slower int
	newSteps, oldSteps    int64
	worstRatio            float64
	worstSize             int
}

// legacyApplies reports whether the previous loop goes on to count rule r
// after its target, size and header descriptor filters, and whether its
// descriptors are then unknown.
func legacyApplies(r *rule, info *objectInfo, size int) (applies, descriptorsUnknown bool) {
	if r.target != 0 && r.target != info.target || r.size != nil && !r.size.contains(int64(size)) {
		return false, false
	}
	if r.entry != nil || r.sections != nil {
		switch info.status {
		case exeParsed:
			if r.entry != nil && !r.entry.contains(int64(info.meta.entry)) || r.sections != nil && !r.sections.contains(int64(len(info.meta.sections))) {
				return false, false
			}
		case exeUnknown:
			return true, true
		default:
			return false, false
		}
	}
	return true, false
}

// compareMatchers checks one object: detection presence and the exact error, the
// counts and first offsets every rule would receive, that every rule left out
// could not change the result, and that a smaller budget never turns into a
// different answer.
func compareMatchers(t *testing.T, r *rand.Rand, d *Database, legacy *legacyMatcher, data []byte, st *diffStats, lines []string) {
	t.Helper()
	ctx := context.Background()
	fail := func(format string, args ...any) {
		t.Helper()
		t.Fatalf("%s\nrules:\n%s\nobject %q", fmt.Sprintf(format, args...), strings.Join(lines, "\n"), data)
	}
	name, steps, err := d.matchCounted(ctx, data, 1<<40)
	oldName, oldSteps, oldErr := legacy.match(ctx, data, 1<<40)
	if (name == "") != (oldName == "") || fmt.Sprint(err) != fmt.Sprint(oldErr) {
		fail("Match = %q, %v; previous matcher %q, %v", name, err, oldName, oldErr)
	}
	st.objects++
	st.newSteps += steps
	st.oldSteps += oldSteps
	if steps <= oldSteps {
		st.fasterOrEqual++
	} else {
		st.slower++
		if ratio := float64(steps) / float64(max(oldSteps, 1)); ratio > st.worstRatio {
			st.worstRatio, st.worstSize = ratio, len(data)
		}
	}
	if name != "" {
		st.detected++
	}
	if errors.Is(err, ErrRuleIncomplete) {
		st.incomplete++
	}
	info := d.objectInfo(data, MatchContext{Type: d.rawScanType(data)})
	w, ow := &matchWork{ctx: ctx, limit: 1 << 40}, &matchWork{ctx: ctx, limit: 1 << 40}
	counts, err := d.newPatternCounts(data, &info, w)
	if err != nil {
		t.Fatal(err)
	}
	oldCounts, err := legacy.newCounts(data, &info, ow)
	if err != nil {
		t.Fatal(err)
	}
	// Match stops at the first detection, so comparing results alone would
	// not notice a later rule wrongly left out. Every rule of the object's
	// roots left out of the reached set must be false there.
	reached, err := d.reachedRules(counts.candidates, info.target, w)
	if err != nil {
		t.Fatal(err)
	}
	// Native standalone/hash/logical precedence intentionally differs from the
	// old source-ordered loop. Every rule still has its counts and first offsets
	// compared below. When labels differ, independently prove both selected
	// rules true; exact winner precedence has separate native controls.
	selected := map[string]bool{}
	for i := range d.rules {
		r := &d.rules[i]
		if err := counts.rule(r); err != nil {
			t.Fatal(err)
		}
		// As evaluateRule does: the forward search resumes the counts
		// stopped at the memory bound when the rule depends on them.
		if counts.unlimit(r) {
			w.forward = true
			err := counts.rule(r)
			w.forward = false
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := oldCounts.rule(r); err != nil {
			t.Fatal(err)
		}
		// The previous matcher decides every start past the memory bound;
		// the new one leaves such a count undecided exactly when the rule's
		// result does not depend on it.
		same := len(counts.c) == len(oldCounts.c)
		for k := 0; same && k < len(counts.c); k++ {
			if counts.c[k] == countStateLimited && oldCounts.c[k] != countStateLimited {
				same = counts.first[k] == missingOffset
				continue
			}
			same = counts.c[k] == oldCounts.c[k] && counts.first[k] == oldCounts.first[k]
		}
		if same && !slices.Equal(counts.c, oldCounts.c) {
			state := ruleState(r, counts.c)
			same = state != logicUnknown && state == ruleState(r, oldCounts.c) && (state == logicFalse || r.bytecode == nil)
			st.lazy++
		}
		if !same {
			fail("rule %s: counts %v first %v; previous matcher %v first %v", r.name, counts.c, counts.first, oldCounts.c, oldCounts.first)
		}
		applies, descriptorsUnknown := legacyApplies(r, &info, len(data))
		if (r.name == name || r.name == oldName) && (r.target == 0 || r.target == info.target) &&
			applies && !descriptorsUnknown && r.inLayers(info.layers) && ruleState(r, oldCounts.c) == logicTrue {
			selected[r.name] = true
		}
		if state := ruleState(r, counts.c); applies && !reached.has(i) && state != logicFalse {
			fail("rule %s left out with counts %v (state %v)", r.name, counts.c, state)
		}
		if slices.Contains(counts.c, countStateLimited) {
			st.limited++
		}
		if applies && slices.Max(counts.c) <= 0 {
			state := logicFalse
			switch {
			case r.expr != nil:
				state = r.expr.evaluate(counts.c).state
			case counts.c[0] < 0:
				state = logicUnknown
			}
			if state != logicFalse {
				st.zeroLive++
			}
		}
	}
	if name != oldName && (!selected[name] || !selected[oldName]) {
		fail("changed label is not independently proved: %q/%q, proved=%v", name, oldName, selected)
	}
	for id, pc := range counts.counts {
		if !pc.windowed {
			st.spanned++
			continue
		}
		st.windowed++
		var bound, open int64
		stateWindows(topLevelTokens(&d.patterns[id]), 0, 0, &bound, &open)
		if proneFrom(bound, open) != math.MaxInt {
			st.windowedLimitable++
		}
		if d.patterns[id].proneFrom() != math.MaxInt {
			st.windowedBelowLimit++
		}
	}
	// A smaller budget may leave the answer unknown, never change it.
	budget := r.Int63n(steps + 1)
	partial, _, perr := d.matchCounted(ctx, data, budget)
	switch {
	case perr == nil && (partial != name || err != nil):
		fail("budget %d of %d gave %q, nil; unlimited %q, %v", budget, steps, partial, name, err)
	case partial != "" && partial != name:
		fail("budget %d of %d gave %q; unlimited %q", budget, steps, partial, name)
	case perr != nil && !errors.Is(perr, ErrMatchLimit) && !errors.Is(perr, ErrRuleIncomplete):
		fail("budget %d gave %v", budget, perr)
	}
}

// TestMatchAgreesWithPreviousMatcher compares Match with the previous matcher
// (legacy_match_test.go) over thousands of generated rule sets and objects,
// including objects of tens of kilobytes, low per-start state limits that
// make limited counts common, and a tiny occurrence store that forces the
// fallback to span counting.
func TestMatchAgreesWithPreviousMatcher(t *testing.T) {
	rounds := 3000
	if testing.Short() {
		rounds = 300
	}
	r := rand.New(rand.NewSource(20260924))
	g := &ruleGen{r: r}
	var st diffStats
	defer func(states, runs, forward int) {
		maxPatternStates, maxStoredRuns, maxForwardSteps = states, runs, forward
	}(maxPatternStates, maxStoredRuns, maxForwardSteps)
	for round := 0; round < rounds; round++ {
		maxPatternStates, maxStoredRuns, maxForwardSteps = 200000, 1<<20, 1<<22
		switch round % 5 {
		case 1:
			// Limited starts: the forward search that decides a start past
			// the memory bound gets a bound of its own that is reached too.
			maxPatternStates, maxForwardSteps = 1+r.Intn(6), 1+r.Intn(60)
		case 2:
			// The forward search decides the starts past the memory bound.
			maxPatternStates = 6 + r.Intn(400)
		case 3:
			maxStoredRuns = 1 + r.Intn(16)
		}
		g.gappy = round%5 == 1 || round%5 == 2
		switch round % 3 {
		case 0:
			g.alphabet = []byte("aAbB.1\x00mMzZ")
		case 1:
			g.alphabet = []byte("\x00\x00\x00\xff\x90\xcc\x8b\x45\xe8\x55\x89\xecABab")
		default:
			g.alphabet = make([]byte, 40)
			r.Read(g.alphabet)
		}
		if g.gappy && r.Intn(2) == 0 {
			g.alphabet = []byte("AAB\x00")
		}
		lines, pool := g.ruleSet()
		if len(lines) == 0 {
			continue
		}
		d, err := compileGenerated(lines)
		if err != nil {
			t.Fatalf("%v\n%s", err, strings.Join(lines, "\n"))
		}
		d = graphMatcherProjection(t, d)
		for i := range d.rules {
			r := &d.rules[i]
			if r.size != nil || r.entry != nil || r.sections != nil {
				st.filtered++
			}
			if slices.ContainsFunc(r.patterns, func(id int) bool { return !d.patterns[id].anchored }) {
				st.unanchored++
			}
		}
		legacy := newLegacyMatcher(d)
		for sample := 0; sample < 8; sample++ {
			size := []int{8, 64, 300, 1500, 6000, 40000}[r.Intn(6)]
			if sample == 0 {
				size = 0
			}
			compareMatchers(t, r, d, legacy, g.object(pool, size), &st, lines)
		}
	}
	t.Logf("%d objects: %d detected, %d incomplete, %d rule evaluations with a limited count, %d true or unknown without a subsignature, %d rules with a size or header descriptor, %d with a subsignature without anchors, %d windowed counts (%d of limitable patterns, %d below their length bound) and %d span counts",
		st.objects, st.detected, st.incomplete, st.limited, st.zeroLive, st.filtered, st.unanchored, st.windowed, st.windowedLimitable, st.windowedBelowLimit, st.spanned)
	t.Logf("%d rule evaluations left a count past the memory bound undecided", st.lazy)
	t.Logf("steps: new matcher %d, previous %d; new <= previous on %d objects, more on %d (at worst %.2fx, on a %d-byte object)",
		st.newSteps, st.oldSteps, st.fasterOrEqual, st.slower, st.worstRatio, st.worstSize)
	if !testing.Short() && (st.detected < st.objects/10 || st.incomplete < 100 || st.limited < 1000 || st.zeroLive < 1000 || st.filtered < 1000 || st.unanchored < 1000 || st.windowed < 10000 || st.windowedLimitable < 1000 || st.spanned < 10000) {
		t.Fatal("weak comparison: too few detections, incomplete results, limited counts, rules true without a subsignature, rules with descriptors or without anchors, or counts of either kind")
	}
}

// A start no anchor occurrence allows can still exhaust the per-start state
// bound before failing. Such starts are skipped only on objects too short for
// that; on longer ones every start in the span is tried, so the limit (and the
// incomplete result it causes) is kept exactly.
func TestSkippedStartsCannotHideStateLimit(t *testing.T) {
	defer func(states, forward int) { maxPatternStates, maxForwardSteps = states, forward }(maxPatternStates, maxForwardSteps)
	maxPatternStates, maxForwardSteps = 50, 10
	// The first branch's anchor (CD) follows unbounded gaps; only the second
	// branch's anchor (EF) occurs in the objects below.
	d := compileTest(t, "ndb", "Branch:0:*:(41*42*4344|4546)47")
	p := &d.patterns[0]
	if p.proneFrom() != 23 {
		t.Fatalf("proneFrom = %d, want 23", p.proneFrom())
	}
	legacy := newLegacyMatcher(d)
	for _, tt := range []struct {
		data              string
		windowed, limited bool
		name              string
	}{
		// Below the bound: only the two EF starts are tried; the A start
		// between them is skipped and could not have been limited.
		{"EFxA" + strings.Repeat("B", 14) + "EFG", true, false, "Branch"},
		{"EFxA" + strings.Repeat("B", 14) + "EFx", true, false, ""},
		// From the bound on, every start in the span is tried as before: the
		// A start is limited after 25 Bs, before the last EF start is reached.
		{"EFxA" + strings.Repeat("B", 30) + "EFx", false, true, ""},
		{"EFxA" + strings.Repeat("B", 30) + "EFG", false, true, ""},
	} {
		data := []byte(tt.data)
		name, _, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, _, oldErr := legacy.match(context.Background(), data, 1<<40)
		if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%q: Match = %q, %v; previous matcher %q, %v", tt.data, name, err, oldName, oldErr)
		}
		if name != tt.name || errors.Is(err, errPatternStates) != tt.limited {
			t.Fatalf("%q: Match = %q, %v; want %q, limited %v", tt.data, name, err, tt.name, tt.limited)
		}
		counts, err := d.newPatternCounts(data, &objectInfo{}, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if err != nil {
			t.Fatal(err)
		}
		if err := counts.rule(&d.rules[0]); err != nil {
			t.Fatal(err)
		}
		if pc := counts.counts[0]; pc.windowed != tt.windowed {
			t.Fatalf("%q: windowed = %v, want %v", tt.data, pc.windowed, tt.windowed)
		}
	}
}

// The work charged for an object must not depend on map iteration order: a
// verdict close to the budget would otherwise differ between two scans of the
// same object. Here the two anchors of an alternative allow windows with the
// same first start, so the order their occurrences are merged in decides how
// the prefix search is split and charged.
func TestMatchStepsAreDeterministic(t *testing.T) {
	d := compileTest(t, "ndb", "Alt:0:*:5a{0-20}(4142434445|{0-10}4647484950)")
	var data []byte
	for i := 0; i < 200; i++ {
		data = append(data, bytes.Repeat([]byte{0x11}, 400)...)
		data = append(data, "ABCDE"...)
		data = append(data, bytes.Repeat([]byte{0x22}, 5)...)
		data = append(data, "FGHIP"...)
	}
	counts, err := d.newPatternCounts(data, &objectInfo{}, &matchWork{ctx: context.Background(), limit: 1 << 40})
	if err != nil {
		t.Fatal(err)
	}
	if err := counts.rule(&d.rules[0]); err != nil {
		t.Fatal(err)
	}
	if pc := counts.counts[0]; !pc.windowed || len(pc.streams) != 2 {
		t.Fatalf("want a windowed count over two anchors, got windowed=%v streams=%d", pc.windowed, len(pc.streams))
	}
	seen := map[int64]int{}
	for i := 0; i < 64; i++ {
		name, steps, err := d.matchCounted(context.Background(), data, 1<<40)
		if name != "" || err != nil {
			t.Fatalf("Match = %q, %v", name, err)
		}
		seen[steps]++
	}
	if len(seen) != 1 {
		t.Fatalf("one object was charged different step counts across scans: %v", seen)
	}
}

// An unbounded gap that failed at one position fails at every later position
// of the same start: it can only reach a subset of the same continuations.
// Case-insensitive patterns cannot jump to the next literal, so without that
// memory each occurrence of the middle literal rescanned the rest of the
// object (as in a 14 MB .NET assembly with a case-insensitive
// "MethodName*Start*cmd" subsignature). Results and state limits must stay
// those of the previous matcher.
func TestUnboundedGapFailureIsRememberedWithinAStart(t *testing.T) {
	defer func(states int) { maxPatternStates = states }(maxPatternStates)
	for _, tt := range []struct {
		states int
		rule   string
		data   string
		name   string
	}{
		{200000, "Gap;Target:0;0;4d6574686f644e616d65*5374617274*636d64::i", "MethodName" + strings.Repeat("start ", 4000) + strings.Repeat("x", 40000), ""},
		{200000, "Gap;Target:0;0;4d6574686f644e616d65*5374617274*636d64::i", "MethodName" + strings.Repeat("start ", 4000) + strings.Repeat("x", 40000) + "CMD", "Gap"},
		{200000, "Gap;Target:0;0;41*42*43", "A" + strings.Repeat("B", 20000), ""},
		{200000, "Gap;Target:0;0;41{2-}42{1-}43", "AAA" + strings.Repeat("BxB", 7000) + "C", "Gap"},
		// With a small state bound the remembered failure must fill the memo
		// exactly as re-exploring would, so the limit is reached at the same
		// start and the rule stays incomplete.
		{50, "Gap;Target:0;0;41*42*43", "A" + strings.Repeat("B", 100), ""},
		{50, "Gap;Target:0;0;41*42*43", "A" + strings.Repeat("B", 20) + "C", "Gap"},
		{50, "Gap;Target:0;0;41*42*43::i", "a" + strings.Repeat("b", 100) + "c", ""},
		{7, "Gap;Target:0;0;41*(42|62)*43", "A" + strings.Repeat("bB", 6) + "C", ""},
	} {
		maxPatternStates = tt.states
		d := compileTest(t, "ldb", tt.rule)
		data := []byte(tt.data)
		name, steps, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), data, 1<<40)
		if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%s on %d bytes: Match = %q, %v; previous matcher %q, %v", tt.rule, len(data), name, err, oldName, oldErr)
		}
		if tt.states == 200000 && (name != tt.name || err != nil) {
			t.Fatalf("%s: Match = %q, %v, want %q", tt.rule, name, err, tt.name)
		}
		if tt.states == 200000 && steps > 20*int64(len(data)) {
			t.Fatalf("%s: %d steps for %d bytes", tt.rule, steps, len(data))
		}
	}
}

// Every start of a pattern such as ABCD*EFGH searches for EFGH from its own
// position. The first occurrence at or after a position does not depend on the
// start, so one object's searches share what earlier ones found; otherwise
// many starts and no EFGH cost the object length once per start.
func TestGapLiteralSearchIsSharedAcrossStarts(t *testing.T) {
	for _, tt := range []struct {
		rule, data, name string
	}{
		{"Far:0:*:41424344*45464748", strings.Repeat("ABCD"+strings.Repeat("-", 16), 3000), ""},
		{"Far:0:*:41424344*45464748", strings.Repeat("ABCD"+strings.Repeat("-", 16), 3000) + "EFGH", "Far"},
		{"Far:0:*:41424344{2-}45464748", strings.Repeat("ABCDEFGH-", 3000), "Far"},
		{"Far:0:*:41424344{10-}45464748", strings.Repeat("ABCDEFGH-", 3000), "Far"},
		{"Far:0:*:41424344{10-12}45464748", strings.Repeat("ABCDEFGH-", 3000), ""},
		{"Far:0:*:41424344{10-12}45464748", strings.Repeat("ABCDEFGH-", 3000) + "ABCD" + strings.Repeat("-", 10) + "EFGH", "Far"},
		{"Far;Target:0;0>2999;41424344*45464748", strings.Repeat("ABCD"+strings.Repeat("-", 16), 3000) + "EFGH", "Far"},
		{"Far;Target:0;0>3000;41424344*45464748", strings.Repeat("ABCD"+strings.Repeat("-", 16), 3000) + "EFGH", ""},
		// A start searches for the middle literal before its occurrence and,
		// once the rest fails, again after it. Both searches must stay known
		// to the later starts (as in a 63 MB Go program and a mail rule of
		// the form pass*application/octet-stream*...), or each start searches
		// most of the object again.
		{"Far:0:*:70617373*4d4944444c45*54414c4c", strings.Repeat("pass-", 300) + strings.Repeat("-", 30000) + "MIDDLE" + strings.Repeat("-", 60000), ""},
		{"Far:0:*:70617373*4d4944444c45*54414c4c", strings.Repeat("pass-", 300) + strings.Repeat("-", 30000) + "MIDDLE" + strings.Repeat("-", 60000) + "TALL", "Far"},
		{"Far:0:*:70617373*4d4944444c45*54414c4c", strings.Repeat("pass-", 300) + strings.Repeat("-MIDDLE"+strings.Repeat("-", 3000), 10) + strings.Repeat("-", 60000), ""},
		{"Far;Target:0;0>299;70617373*4d4944444c45*54414c4c", strings.Repeat("pass-", 300) + strings.Repeat("-", 30000) + "MIDDLE" + strings.Repeat("-", 60000) + "TALL", "Far"},
		{"Far;Target:0;0>300;70617373*4d4944444c45*54414c4c", strings.Repeat("pass-", 300) + strings.Repeat("-", 30000) + "MIDDLE" + strings.Repeat("-", 60000) + "TALL", ""},
	} {
		ext := "ndb"
		if strings.Contains(tt.rule, ";") {
			ext = "ldb"
		}
		d := compileTest(t, ext, tt.rule)
		d = graphMatcherProjection(t, d)
		data := []byte(tt.data)
		name, steps, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), data, 1<<40)
		if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%s: Match = %q, %v; previous matcher %q, %v", tt.rule, name, err, oldName, oldErr)
		}
		if name != tt.name || err != nil {
			t.Fatalf("%s: Match = %q, %v, want %q", tt.rule, name, err, tt.name)
		}
		if steps > 20*int64(len(data)) {
			t.Fatalf("%s: %d steps for %d bytes", tt.rule, steps, len(data))
		}
	}
}

// The literal searches one object's counts keep beyond the first of each
// literal are bounded per object, as its anchor occurrences are, and the bound
// changes only the work charged. Without it every candidate pattern with a
// common literal after a gap keeps maxLiteralHits of them for the whole scan.
func TestLiteralSearchMemoryIsBoundedPerObject(t *testing.T) {
	var rules strings.Builder
	var data []byte
	for i := 0; i < 300; i++ {
		unique := fmt.Sprintf("<%08d-unique>", i)
		fmt.Fprintf(&rules, "Gap.%d:0:*:%s*2d2d*0a0b\n", i, hex.EncodeToString([]byte(unique)))
		data = append(data, unique...)
	}
	data = append(data, strings.Repeat("--.", 100)...)
	d := compileTest(t, "ndb", rules.String())
	d = graphMatcherProjection(t, d)
	// kept evaluates every rule as Match does and returns the hits kept
	// beyond the first of each literal, every rule's counts and the steps.
	kept := func() (int, []int, int64) {
		w := &matchWork{ctx: context.Background(), limit: 1 << 40}
		m, err := d.newPatternCounts(data, &objectInfo{}, w)
		if err != nil {
			t.Fatal(err)
		}
		var counts []int
		for i := range d.rules {
			if err := m.rule(&d.rules[i]); err != nil {
				t.Fatal(err)
			}
			counts = append(counts, m.c...)
		}
		extra := 0
		for _, pc := range m.counts {
			for _, hits := range pc.searches {
				extra += len(hits) - 1
			}
		}
		return extra, counts, w.steps
	}
	defer func(n int) { maxSearchHits = n }(maxSearchHits)
	maxSearchHits = math.MaxInt
	unbounded, want, _ := kept()
	if unbounded < 300*(maxLiteralHits-1) {
		t.Fatalf("the object keeps only %d extra hits without a bound; the test no longer fills them", unbounded)
	}
	for _, bound := range []int{0, 1, 1000} {
		maxSearchHits = bound
		extra, counts, steps := kept()
		t.Logf("bound %d: %d extra hits kept (%d without a bound), %d steps", bound, extra, unbounded, steps)
		if extra > bound {
			t.Fatalf("bound %d: %d extra hits kept", bound, extra)
		}
		if !slices.Equal(counts, want) {
			t.Fatalf("bound %d changed the counts", bound)
		}
		name, _, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), data, 1<<40)
		if name != "" || err != nil || oldName != "" || oldErr != nil {
			t.Fatalf("bound %d: Match = %q, %v; previous matcher %q, %v", bound, name, err, oldName, oldErr)
		}
	}
}

// Every remembered literal search must answer exactly as a new search would,
// whatever order the searches come in and after the per-literal or the
// per-object bound fills.
func TestLiteralSearchesAnswerAsSearching(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	known := 0
	for round := 0; round < 200; round++ {
		data := make([]byte, 50+r.Intn(3000))
		for i := range data {
			data[i] = "ab"[r.Intn(2)]
		}
		// Needles that overlap themselves occur at adjacent positions, where
		// a search that finds at is next to one that finds at+1.
		needle := []byte([]string{"ab", "abb", "abba", "b", "aa", "aaa", "aba"}[r.Intn(7)])
		first := func(from int) int {
			if i := bytes.Index(data[from:], needle); i >= 0 {
				return from + i
			}
			return -1
		}
		var s literalSearches
		var searched []int
		var w matchWork
		if round%2 == 1 {
			// The object's other counts have (nearly) used its extra hits.
			w.searchHits = maxSearchHits - r.Intn(8)
		}
		for op := 0; op < 400; op++ {
			pos := r.Intn(len(data) + 1)
			if len(searched) > 0 && r.Intn(2) == 0 {
				// Near an earlier search, where it may be remembered.
				pos = min(len(data), searched[r.Intn(len(searched))]+r.Intn(4))
			}
			if at, ok := s.lookup(7, pos); ok {
				known++
				if at != first(pos) {
					t.Fatalf("lookup(%d) = %d, searching finds %d", pos, at, first(pos))
				}
				continue
			}
			s.record(7, pos, first(pos), &w)
			searched = append(searched, pos)
			hits := s[7]
			if len(hits) > maxLiteralHits || w.searchHits > maxSearchHits || !slices.IsSortedFunc(hits, func(a, b literalHit) int { return a.from - b.from }) {
				t.Fatalf("%d hits, sorted %v", len(hits), slices.IsSortedFunc(hits, func(a, b literalHit) int { return a.from - b.from }))
			}
		}
	}
	if known < 10000 {
		t.Fatalf("only %d lookups were answered from memory", known)
	}
}
