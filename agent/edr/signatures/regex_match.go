package signatures

import "errors"

// The backtracking matcher gives a regex subsignature the reference's
// meaning: leftmost match, alternatives and repeats tried in the library's
// priority order, captures and backreferences, lookaround and atomic groups,
// over bytes.
//
// The reference's library gives up on one call when a start position needs
// more than 100,000 backtracking calls or 2,000 nested backtracking frames,
// and the reference then counts no match. Those counts depend on the
// library's own optimizations, which are not modeled. Instead this matcher
// counts calls and nesting at least as generously as the library does, per
// start position as the library does, and stops a margin below the limits:
// when its counts stay within them the library's do too, so the result is
// the reference's; otherwise the result is unknown (errRegexLimit).
//
// regex_oracle_test.go checks the counts against the library's smallest
// sufficient limits: on 30,002 random cases they fall short by at most 3
// calls and 9 frames (patterns mixing lookaround and atomic groups), and on
// repeats of each construct at 500 iterations they are never short.

const (
	rxMatchLimit = 100_000
	rxDepthLimit = 2_000
	// The margins kept below the library's limits.
	rxCallSlack  = 1_000
	rxDepthSlack = 100
)

// errRegexLimit: the reference could have given up on this call.
var errRegexLimit = errors.New("regex backtracking limit")

type rxMatcher struct {
	p     *rxPattern
	data  []byte
	start int // subject start: ^ and \A match here; nothing before it is seen
	end   int // subject end
	caps  []int
	calls int
	depth int
	w     *matchWork
	err   error
	// peakCalls and peakDepth are the largest counts of any start so far.
	peakCalls, peakDepth int
	// callLimit and depthLimit are rxMatchLimit and rxDepthLimit, variables
	// only for tests.
	callLimit, depthLimit int
}

func newRxMatcher(p *rxPattern, data []byte, start, end int, w *matchWork) *rxMatcher {
	m := &rxMatcher{p: p, data: data, start: start, end: end, w: w,
		callLimit: rxMatchLimit - rxCallSlack, depthLimit: rxDepthLimit - rxDepthSlack}
	m.caps = make([]int, 2*(p.groups+1))
	return m
}

// call counts one backtracking call and enters its frame; leave must follow.
func (m *rxMatcher) call() bool {
	m.calls++
	m.depth++
	m.peakCalls = max(m.peakCalls, m.calls)
	m.peakDepth = max(m.peakDepth, m.depth)
	if m.calls > m.callLimit || m.depth > m.depthLimit {
		m.err = errRegexLimit
		return false
	}
	if err := m.w.step(1); err != nil {
		m.err = err
		return false
	}
	return true
}

func (m *rxMatcher) leave() { m.depth-- }

// charge charges n examined bytes to the object's work budget.
func (m *rxMatcher) charge(n int) bool {
	if err := m.w.step(int64(max(n, 1))); err != nil {
		m.err = err
		return false
	}
	return true
}

// matchAt reports the end of the match starting exactly at pos.
func (m *rxMatcher) matchAt(pos int) (int, bool, error) {
	for i := range m.caps {
		m.caps[i] = -1
	}
	m.calls, m.depth, m.err = 0, 0, nil
	end := -1
	// The library's first frame, and the bracket around the whole pattern.
	if !m.call() || !m.call() {
		return 0, false, m.err
	}
	ok := m.match(m.p.root, pos, func(e int) bool { end = e; return true })
	m.leave()
	m.leave()
	if m.err != nil {
		return 0, false, m.err
	}
	return end, ok, nil
}

func (m *rxMatcher) wordAt(pos int) bool {
	return pos >= m.start && pos < m.end && isRxWord(m.data[pos])
}

func (m *rxMatcher) assertion(a rxAssertKind, pos int) bool {
	switch a {
	case asSubjectStart:
		return pos == m.start
	case asLineStart:
		return pos == m.start || pos < m.end && m.data[pos-1] == '\n'
	case asSubjectEnd:
		return pos == m.end
	case asEndOrFinalNL:
		return pos == m.end || pos == m.end-1 && m.data[pos] == '\n'
	case asLineEnd:
		return pos == m.end || m.data[pos] == '\n'
	case asWordBoundary:
		return m.wordAt(pos-1) != m.wordAt(pos)
	case asNotWordBoundary:
		return m.wordAt(pos-1) == m.wordAt(pos)
	}
	return false
}

// match matches n at pos and then k at the end; it returns k's result for
// the first way that k accepts, in priority order. After m.err is set every
// call returns false.
func (m *rxMatcher) match(n *rxNode, pos int, k func(int) bool) bool {
	if m.err != nil {
		return false
	}
	switch n.kind {
	case rxEmpty:
		return k(pos)
	case rxSet:
		return pos < m.end && n.set.has(m.data[pos]) && k(pos+1)
	case rxString:
		if pos+len(n.seq) > m.end {
			return false
		}
		for i := range n.seq {
			if !n.seq[i].has(m.data[pos+i]) {
				return false
			}
		}
		return k(pos + len(n.seq))
	case rxConcat:
		return m.sequence(n.subs, pos, k)
	case rxAlt:
		for _, alt := range n.subs {
			if !m.call() {
				return false
			}
			ok := m.match(alt, pos, k)
			m.leave()
			if ok || m.err != nil {
				return ok
			}
		}
		return false
	case rxGroup:
		if !m.call() {
			return false
		}
		defer m.leave()
		if n.cap == 0 {
			return m.match(n.subs[0], pos, k)
		}
		i := 2 * n.cap
		old0, old1 := m.caps[i], m.caps[i+1]
		ok := m.match(n.subs[0], pos, func(e int) bool {
			in0, in1 := m.caps[i], m.caps[i+1]
			m.caps[i], m.caps[i+1] = pos, e
			if k(e) {
				return true
			}
			m.caps[i], m.caps[i+1] = in0, in1
			return false
		})
		if !ok {
			m.caps[i], m.caps[i+1] = old0, old1
		}
		return ok
	case rxRepeat:
		return m.repeat(n, pos, k)
	case rxAtomic:
		// The library keeps a frame for the group while the rest of the
		// match runs, besides the frames of the group's own search.
		if !m.call() {
			return false
		}
		defer m.leave()
		saved := m.snapshot()
		if e, ok := m.once(n.subs[0], pos); ok && k(e) {
			return true
		}
		m.restore(saved)
		return false
	case rxLook:
		return m.look(n, pos, k)
	case rxBackref:
		s, e := m.caps[2*n.cap], m.caps[2*n.cap+1]
		if s < 0 || e < 0 {
			return false
		}
		l := e - s
		if pos+l > m.end || !m.charge(l) {
			return false
		}
		if !equalBytes(m.data[pos:pos+l], m.data[s:e], n.caseless) {
			return false
		}
		return k(pos + l)
	case rxAssert:
		return m.assertion(n.assert, pos) && k(pos)
	}
	return false
}

func (m *rxMatcher) sequence(subs []*rxNode, pos int, k func(int) bool) bool {
	if len(subs) == 0 {
		return k(pos)
	}
	return m.match(subs[0], pos, func(e int) bool { return m.sequence(subs[1:], e, k) })
}

// snapshot copies the captures, which an atomic part keeps when it succeeds
// and must give back when what follows it fails.
func (m *rxMatcher) snapshot() []int {
	if m.p.groups == 0 {
		return nil
	}
	return append([]int(nil), m.caps...)
}

func (m *rxMatcher) restore(saved []int) {
	if saved != nil {
		copy(m.caps, saved)
	}
}

// once matches n at pos and keeps its first match: an atomic group.
func (m *rxMatcher) once(n *rxNode, pos int) (int, bool) {
	if !m.call() {
		return 0, false
	}
	defer m.leave()
	end := -1
	ok := m.match(n, pos, func(e int) bool { end = e; return true })
	return end, ok
}

func (m *rxMatcher) look(n *rxNode, pos int, k func(int) bool) bool {
	// As for an atomic group, the library keeps a frame for the assertion
	// while the rest of the match runs.
	if !m.call() {
		return false
	}
	defer m.leave()
	saved := append([]int(nil), m.caps...)
	found := false
	if !n.behind {
		_, found = m.once(n.subs[0], pos)
	} else {
		alts := []*rxNode{n.subs[0]}
		if n.subs[0].kind == rxAlt {
			alts = n.subs[0].subs
		}
		for i, alt := range alts {
			if !m.call() {
				return false
			}
			if from := pos - n.lengths[i]; from >= m.start {
				found = m.match(alt, from, func(e int) bool { return e == pos })
			}
			m.leave()
			if found || m.err != nil {
				break
			}
		}
	}
	if m.err != nil {
		return false
	}
	if found == n.negate {
		copy(m.caps, saved)
		return false
	}
	if n.negate {
		copy(m.caps, saved)
	}
	if k(pos) {
		return true
	}
	copy(m.caps, saved)
	return false
}

// repeat matches a quantified item.
func (m *rxMatcher) repeat(n *rxNode, pos int, k func(int) bool) bool {
	sub := n.subs[0]
	if sub.kind == rxSet {
		return m.repeatSet(n, &sub.set, pos, k)
	}
	if n.possessive {
		if !m.call() {
			return false
		}
		defer m.leave()
		saved := m.snapshot()
		end := -1
		if m.repeatGroup(n, sub, 0, pos, func(e int) bool { end = e; return true }) && m.err == nil && k(end) {
			return true
		}
		m.restore(saved)
		return false
	}
	return m.repeatGroup(n, sub, 0, pos, k)
}

// repeatSet repeats one byte set: the run is measured once, then the
// continuation is tried at each allowed length in priority order, each try a
// call as in the library.
func (m *rxMatcher) repeatSet(n *rxNode, set *byteSet, pos int, k func(int) bool) bool {
	limit := m.end - pos
	if n.max >= 0 {
		limit = min(limit, n.max)
	}
	run := 0
	if n.greedy || n.possessive {
		for run < limit && set.has(m.data[pos+run]) {
			run++
		}
		if !m.charge(run) || run < n.min {
			return false
		}
		if n.possessive {
			if !m.call() {
				return false
			}
			defer m.leave()
			return k(pos + run)
		}
		for j := run; j >= n.min; j-- {
			if !m.call() {
				return false
			}
			ok := k(pos + j)
			m.leave()
			if ok || m.err != nil {
				return ok
			}
		}
		return false
	}
	for run < n.min {
		if run >= limit || !set.has(m.data[pos+run]) {
			return false
		}
		run++
	}
	for {
		if !m.call() {
			return false
		}
		ok := k(pos + run)
		m.leave()
		if ok || m.err != nil {
			return ok
		}
		if run >= limit || !set.has(m.data[pos+run]) || !m.charge(1) {
			return false
		}
		run++
	}
}

// repeatGroup repeats a general item, count iterations done so far. As in
// the library, the iterations up to the minimum and those of a bounded
// repeat are copies of the item that may each match the empty string; in the
// open-ended part an iteration that matches the empty string ends the loop
// and the match goes on after it.
func (m *rxMatcher) repeatGroup(n, sub *rxNode, count, pos int, k func(int) bool) bool {
	more := func() bool {
		if n.max >= 0 && count >= n.max {
			return false
		}
		if !m.call() {
			return false
		}
		defer m.leave()
		return m.match(sub, pos, func(e int) bool {
			if e == pos && n.max < 0 && count >= n.min {
				return k(e)
			}
			return m.repeatGroup(n, sub, count+1, e, k)
		})
	}
	if count < n.min {
		return more()
	}
	if n.greedy {
		if more() || m.err != nil {
			return m.err == nil
		}
		return k(pos)
	}
	if !m.call() {
		return false
	}
	ok := k(pos)
	m.leave()
	if ok || m.err != nil {
		return ok
	}
	return more()
}
