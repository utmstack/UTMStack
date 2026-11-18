package signatures

import (
	"iter"
	"math/bits"
	"slices"
)

// Most rules cannot become true on a given object. A pattern none of whose
// anchors occurs in an object has a count of 0 there, whatever its offset or
// the object's header, so a rule none of whose subsignatures is a candidate
// has the outcome it has with every count 0: false for almost every rule. The
// rule index lists, per pattern, the rules that use it, and per target the
// rules that are true with every count 0 or that use a pattern without
// anchors (a candidate on every object). Match evaluates, in rule order, only
// those rules and the rules a candidate pattern reaches, for the object's
// typed root and root 0; every rule it leaves out is false there.

type ruleIndex struct {
	// start and refs list the rules using each pattern: pattern id's rules
	// are refs[start[id]:start[id+1]], in rule order, each once.
	start, refs []uint32
	// live lists, per target and in rule order, the rules evaluated on every
	// object matched against that target's root.
	live [maxTargets][]uint32
}

// ruleState is a rule's outcome for its subsignature counts; a rule without
// an expression has one subsignature, true when it occurs.
func ruleState(r *rule, c []int) logicState {
	if r.expr != nil {
		return r.expr.evaluate(c).state
	}
	switch {
	case c[0] < 0:
		return logicUnknown
	case c[0] > 0:
		return logicTrue
	}
	return logicFalse
}

// buildRuleIndex runs once the rules and patterns are final.
func (d *Database) buildRuleIndex() {
	d.buildStandaloneBM()
	d.buildFixedAC()
	d.buildNativeAC()
	x := &d.ruleIndex
	d.legacyMetadataRules = false
	x.start = make([]uint32, len(d.patterns)+1)
	distinct := func(r *rule, k int) bool { return !slices.Contains(r.patterns[:k], r.patterns[k]) }
	for i := range d.rules {
		r := &d.rules[i]
		if r.target == 13 {
			d.legacyMetadataRules = true
		}
		for k, id := range r.patterns {
			if distinct(r, k) {
				x.start[id+1]++
			}
		}
	}
	for id := range d.patterns {
		x.start[id+1] += x.start[id]
	}
	x.refs = make([]uint32, x.start[len(d.patterns)])
	next := slices.Clone(x.start[:len(d.patterns)])
	var c []int
	for i := range d.rules {
		r := &d.rules[i]
		always := false
		for k, id := range r.patterns {
			if distinct(r, k) {
				x.refs[next[id]] = uint32(i)
				next[id]++
			}
			always = always || d.patterns[id].everywhere()
		}
		// Counts on an object where no subsignature occurs: 0, except a
		// regex whose trigger holds without matches, which the reference
		// runs there, so its count is unknown.
		c = c[:0]
		for _, id := range r.patterns {
			n := 0
			if h := d.patterns[id].holder; h != nil && h.reason == UnknownPCRE {
				n = h.count(c, TypeAny)
			}
			c = append(c, n)
		}
		if always || ruleState(r, c) != logicFalse {
			x.live[r.target] = append(x.live[r.target], uint32(i))
		}
	}
}

// reachedRules returns the rules of the target's root and root 0 that Match
// evaluates on one object: those using an anchored candidate pattern and the
// live rules. Every rule list and rule visited, and the set itself, is
// charged.
func (d *Database) reachedRules(c *candidateSet, target int, w *matchWork) (bitset, error) {
	x := &d.ruleIndex
	reached := newBitset(len(d.rules))
	if err := w.step(int64(len(reached) + len(x.live[0]))); err != nil {
		return nil, err
	}
	for _, i := range x.live[0] {
		reached.set(int(i))
	}
	if target != 0 {
		if err := w.step(int64(len(x.live[target]))); err != nil {
			return nil, err
		}
		for _, i := range x.live[target] {
			reached.set(int(i))
		}
	}
	for _, id := range c.anchored {
		rules := x.refs[x.start[id]:x.start[id+1]]
		if err := w.step(int64(1 + len(rules))); err != nil {
			return nil, err
		}
		for _, i := range rules {
			if t := d.rules[i].target; t == 0 || t == target {
				reached.set(int(i))
			}
		}
	}
	return reached, nil
}

// members yields the set bits in increasing order.
func (b bitset) members() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i, word := range b {
			for word != 0 {
				if !yield(i<<6 | bits.TrailingZeros64(word)) {
					return
				}
				word &= word - 1
			}
		}
	}
}
