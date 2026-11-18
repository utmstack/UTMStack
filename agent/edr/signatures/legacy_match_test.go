package signatures

// This file keeps the matcher as it was before the shared anchor pass, only
// renamed, as the reference the differential tests compare Match against.
// Anchors are each pattern's longest literal (recomputed from the compiled
// nodes), the index records only each literal's first and last occurrence, and
// every candidate pattern tries every start between them.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// graphMatcherProjection preserves the independent pre-native graph oracle.
// Its start-position counts intentionally differ from native multipart event
// counts. Rebuild only this owned test copy's old graph anchors and disable
// the native stream; native pipeline parity is tested against pinned scanner
// controls instead of treating the old graph's counts as that oracle.
func graphMatcherProjection(t testing.TB, original *Database) *Database {
	t.Helper()
	d := *original
	d.patterns = append([]bodyPattern(nil), original.patterns...)
	for id := range d.patterns {
		p := &d.patterns[id]
		if p.simple {
			p.graph = nil
			continue
		}
		if p.graph == nil {
			continue
		}
		graph := *p.graph
		graph.native = nil
		p.graph = &graph
		ts := topLevelTokens(p)
		anchors, _ := longestAnchors(ts, 0, 0)
		var bound, open int64
		stateWindows(ts, 0, 0, &bound, &open)
		if proneFrom(bound, open) == math.MaxInt && len(anchors) > 0 {
			anchors = chooseAnchors(ts, 0, 0, p.nocase)
		}
		for i, a := range anchors {
			if p.nocase && bytes.ContainsFunc(a.literal, func(r rune) bool { return r < 0x80 && asciiLetter(byte(r)) }) {
				anchors[i].literal, anchors[i].fold = asciiLower(a.literal), true
			}
		}
		graph.anchors = anchors
		p.anchored = len(anchors) != 0
	}
	d.anchorGroups = nil
	d.unanchored = nil
	d.index = anchorIndex{}
	d.ruleIndex = ruleIndex{}
	if err := d.buildAnchors(); err != nil {
		t.Fatal(err)
	}
	d.buildRuleIndex()
	d.nativeAC = [maxTargets]*nativeACRoot{}
	return &d
}

// topLevelTokens recovers a compiled pattern's top-level token sequence.
func topLevelTokens(p *bodyPattern) []token {
	if p.simple {
		return []token{{literal: p.prefix}}
	}
	var ts []token
	for id := p.graph.start; id != 0; id = p.graph.nodes[id].next {
		ts = append(ts, p.graph.nodes[id].token)
	}
	return ts
}

// legacyAnchors is how a pattern's anchors were chosen: the longest literal,
// lowercased when the pattern is case-insensitive and the literal has letters.
func legacyAnchors(p *bodyPattern) []patternAnchor {
	anchors := legacyAnchorsFor(topLevelTokens(p))
	for i, a := range anchors {
		if p.nocase && bytes.ContainsFunc(a.literal, func(r rune) bool { return r < 0x80 && asciiLetter(byte(r)) }) {
			anchors[i].literal, anchors[i].fold = asciiLower(a.literal), true
		}
	}
	return anchors
}

func legacyAnchorsFor(ts []token) []patternAnchor {
	var best []patternAnchor
	bestLen, lo, hi := 0, 0, 0
	for _, t := range ts {
		if len(t.literal) > bestLen {
			best, bestLen = []patternAnchor{{literal: t.literal, minPre: lo, maxPre: hi}}, len(t.literal)
		}
		if len(t.alternatives) > 0 && !t.negative {
			var set []patternAnchor
			shortest := 1 << 30
			for _, alternative := range t.alternatives {
				branch := legacyAnchorsFor(alternative)
				if len(branch) == 0 {
					set = nil
					break
				}
				for _, a := range branch {
					a.minPre += lo
					if hi < 0 || a.maxPre < 0 {
						a.maxPre = -1
					} else {
						a.maxPre += hi
					}
					shortest = min(shortest, len(a.literal))
					set = append(set, a)
				}
			}
			if len(set) > 0 && shortest > bestLen {
				best, bestLen = set, shortest
			}
		}
		a, b := tokenWidth([]token{t})
		lo += a
		if hi < 0 || b < 0 {
			hi = -1
		} else {
			hi += b
		}
	}
	return best
}

type legacyGroup struct {
	literal []byte
	members []anchorMember
}

type legacyBucket struct {
	groups []*legacyGroup
	chunk  map[uint32]*legacyBucket
	single map[byte]*legacyBucket
}

func (b *legacyBucket) split(depth int) {
	if len(b.groups) <= crowdedBucket {
		return
	}
	var keep []*legacyGroup
	for _, g := range b.groups {
		switch rest := len(g.literal) - depth; {
		case rest <= 0:
			keep = append(keep, g)
		case rest >= 4:
			if b.chunk == nil {
				b.chunk = make(map[uint32]*legacyBucket)
			}
			key := packKey(g.literal[depth : depth+4])
			if b.chunk[key] == nil {
				b.chunk[key] = &legacyBucket{}
			}
			b.chunk[key].groups = append(b.chunk[key].groups, g)
		default:
			if b.single == nil {
				b.single = make(map[byte]*legacyBucket)
			}
			key := g.literal[depth]
			if b.single[key] == nil {
				b.single[key] = &legacyBucket{}
			}
			b.single[key].groups = append(b.single[key].groups, g)
		}
	}
	b.groups = keep
	for _, child := range b.chunk {
		child.split(depth + 4)
	}
	for _, child := range b.single {
		child.split(depth + 1)
	}
}

// legacyMatcher is the previous matcher over a compiled database.
type legacyMatcher struct {
	d          *Database
	anchors    [2][5]map[uint32]*legacyBucket
	unanchored []int
}

func newLegacyMatcher(d *Database) *legacyMatcher {
	m := &legacyMatcher{d: d}
	groups := [2]map[string]*legacyGroup{{}, {}}
	for id := range d.patterns {
		anchors := legacyAnchors(&d.patterns[id])
		if len(anchors) == 0 {
			m.unanchored = append(m.unanchored, id)
			continue
		}
		for _, a := range anchors {
			fold := 0
			if a.fold {
				fold = 1
			}
			g := groups[fold][string(a.literal)]
			if g == nil {
				g = &legacyGroup{literal: a.literal}
				groups[fold][string(a.literal)] = g
				n := min(4, len(a.literal))
				if m.anchors[fold][n] == nil {
					m.anchors[fold][n] = make(map[uint32]*legacyBucket)
				}
				key := packKey(a.literal[:n])
				if m.anchors[fold][n][key] == nil {
					m.anchors[fold][n][key] = &legacyBucket{}
				}
				m.anchors[fold][n][key].groups = append(m.anchors[fold][n][key].groups, g)
			}
			g.members = append(g.members, newAnchorMember(id, a))
		}
	}
	for fold := range m.anchors {
		for n, index := range m.anchors[fold] {
			for _, bucket := range index {
				bucket.split(n)
			}
		}
	}
	return m
}

type legacyCandidates struct {
	marked bitset
	spans  map[int32]startSpan
}

func (c *legacyCandidates) has(id int) bool { return c.marked.has(id) }
func (c *legacyCandidates) span(id int) startSpan {
	if s, ok := c.spans[int32(id)]; ok {
		return s
	}
	return startSpan{0, math.MaxInt}
}

func (m *legacyMatcher) candidates(data []byte, w *matchWork) (*legacyCandidates, error) {
	c := &legacyCandidates{marked: newBitset(len(m.d.patterns)), spans: map[int32]startSpan{}}
	for _, i := range m.unanchored {
		c.marked.set(i)
	}
	lower := func(b byte) byte {
		if b >= 'A' && b <= 'Z' {
			return b + 'a' - 'A'
		}
		return b
	}
	hits := map[*legacyGroup]*[2]int{}
	for fold := range m.anchors {
		matches := func(at int, literal []byte) bool {
			if at+len(literal) > len(data) {
				return false
			}
			if fold == 0 {
				return bytes.Equal(data[at:at+len(literal)], literal)
			}
			for i, b := range literal {
				if lower(data[at+i]) != b {
					return false
				}
			}
			return true
		}
		var walk func(at, depth int, b *legacyBucket) error
		walk = func(at, depth int, b *legacyBucket) error {
			for _, g := range b.groups {
				if err := w.step(1); err != nil {
					return err
				}
				if !matches(at, g.literal) {
					continue
				}
				if h := hits[g]; h != nil {
					h[1] = at
				} else {
					hits[g] = &[2]int{at, at}
				}
			}
			if b.single != nil && at+depth < len(data) {
				next := data[at+depth]
				if fold == 1 {
					next = lower(next)
				}
				if child := b.single[next]; child != nil {
					if err := walk(at, depth+1, child); err != nil {
						return err
					}
				}
			}
			if b.chunk != nil && at+depth+4 <= len(data) {
				key := uint32(0)
				for _, next := range data[at+depth : at+depth+4] {
					if fold == 1 {
						next = lower(next)
					}
					key = key<<8 | uint32(next)
				}
				if child := b.chunk[key]; child != nil {
					return walk(at, depth+4, child)
				}
			}
			return nil
		}
		for n, index := range m.anchors[fold] {
			if len(index) == 0 {
				continue
			}
			key := uint32(0)
			mask := uint32(0xffffffff)
			if n < 4 {
				mask = (1 << uint(8*n)) - 1
			}
			for i, b := range data {
				if err := w.step(1); err != nil {
					return nil, err
				}
				if fold == 1 {
					b = lower(b)
				}
				key = (key<<8 | uint32(b)) & mask
				if i+1 < n {
					continue
				}
				bucket := index[key]
				if bucket == nil {
					continue
				}
				if err := walk(i+1-n, n, bucket); err != nil {
					return nil, err
				}
			}
		}
	}
	for g, h := range hits {
		for _, mb := range g.members {
			if err := w.step(1); err != nil {
				return nil, err
			}
			lo, hi := h[0]-int(mb.maxPre), h[1]-int(mb.minPre)
			if mb.maxPre < 0 {
				lo = 0
			}
			if s, ok := c.spans[mb.id]; ok {
				lo, hi = min(lo, s.lo), max(hi, s.hi)
			}
			c.spans[mb.id] = startSpan{lo, hi}
			c.marked.set(int(mb.id))
		}
	}
	return c, nil
}

type legacyCount struct {
	n        int
	first    uint32
	next, hi int
	done     bool
	limited  bool
}

func legacyNewCount(p *bodyPattern, data []byte, meta *peMeta, span startSpan) *legacyCount {
	pc := &legacyCount{first: missingOffset}
	lo, hi, ok := p.offset.bounds(len(data), meta)
	lo, hi = max(lo, span.lo), min(hi, len(data)-p.minWidth(), span.hi)
	if !ok || hi < lo {
		pc.done = true
	} else {
		pc.next, pc.hi = lo, hi
	}
	return pc
}

func legacyCountUntil(p *bodyPattern, data []byte, w *matchWork, pc *legacyCount, want int) error {
	for ; pc.n < want && !pc.done && !pc.limited; pc.next++ {
		if pc.next > pc.hi {
			pc.done = true
			break
		}
		if err := w.step(1); err != nil {
			return err
		}
		if len(p.prefix) > 0 && !p.nocase {
			bound := min(len(data), pc.hi+len(p.prefix))
			window := data[pc.next:bound]
			relative := bytes.Index(window, p.prefix)
			searched := len(window) + len(p.prefix)
			if relative >= 0 {
				searched = relative + len(p.prefix)
			}
			if err := w.step(int64(max(1, searched))); err != nil {
				return err
			}
			if relative < 0 {
				pc.done = true
				break
			}
			pc.next += relative
		}
		matched, err := legacyMatchesAt(p, data, pc.next, w)
		if errors.Is(err, errPatternStates) {
			pc.limited = true
			break
		}
		if err != nil {
			return err
		}
		if matched {
			if pc.n == 0 {
				pc.first = uint32(pc.next)
			}
			pc.n++
		}
	}
	return nil
}

func (pc *legacyCount) value(want int) (int, uint32) {
	switch {
	case pc.n >= want:
		return want, pc.first
	case pc.done:
		return pc.n, pc.first
	}
	return countStateLimited, missingOffset
}

func legacyMatchesAt(p *bodyPattern, data []byte, start int, w *matchWork) (bool, error) {
	if p.simple {
		// Keep the reference evaluator graph-based even when the production
		// pattern stores only literal bytes.
		copy := *p
		copy.graph = &patternGraph{nodes: []patternNode{{next: -1}, {token: token{literal: p.prefix}}}, start: 1}
		p = &copy
	}
	failed := make(map[patternState]bool)
	var visit func(int, int) (bool, error)
	visit = func(id, pos int) (bool, error) {
		if err := w.step(1); err != nil {
			return false, err
		}
		if pos < 0 || pos > len(data) {
			return false, nil
		}
		if id == 0 {
			return true, nil
		}
		state := patternState{id, pos}
		if failed[state] {
			return false, nil
		}
		if len(failed) >= maxPatternStates {
			return false, errPatternStates
		}
		node := p.graph.nodes[id]
		t := node.token
		var matched, expanded bool
		var err error
		switch {
		case t.boundary != 0:
			if boundaryHolds(t.boundary, data, pos) {
				expanded = true
				matched, err = visit(node.next, pos)
			}
		case len(node.alternatives) > 0:
			expanded = true
			for _, alt := range node.alternatives {
				matched, err = visit(alt, pos)
				if matched || err != nil {
					return matched, err
				}
			}
		case len(t.literal) > 0:
			end := pos + len(t.literal)
			if end <= len(data) {
				if err := w.step(int64(len(t.literal))); err != nil {
					return false, err
				}
			}
			if end <= len(data) && equalBytes(data[pos:end], t.literal, p.nocase) {
				expanded = true
				matched, err = visit(node.next, end)
			}
		case t.gap:
			expanded = true
			end := len(data)
			if t.max >= 0 {
				end = min(end, pos+t.max)
			}
			for next := pos + t.min; next <= end; next++ {
				nextNode := p.graph.nodes[node.next]
				if !p.nocase && len(nextNode.token.literal) > 0 {
					needle := nextNode.token.literal
					bound := min(len(data), end+len(needle))
					if next > bound {
						break
					}
					window := data[next:bound]
					relative := bytes.Index(window, needle)
					searched := len(window) + len(needle)
					if relative >= 0 {
						searched = relative + len(needle)
					}
					if err := w.step(int64(max(1, searched))); err != nil {
						return false, err
					}
					if relative < 0 {
						break
					}
					next += relative
				} else if node.follow != nil {
					from := next
					for next <= end && (next == len(data) || !node.follow.has(data[next])) {
						next++
					}
					if err := w.step(int64(max(1, next-from))); err != nil {
						return false, err
					}
					if next > end {
						break
					}
				}
				matched, err = visit(node.next, next)
				if matched || err != nil {
					return matched, err
				}
			}
		case t.negative:
			if pos+t.min <= len(data) {
				allowed := true
				for _, seq := range t.alternatives {
					var literal []byte
					for _, x := range seq {
						literal = append(literal, x.literal...)
					}
					if err := w.step(int64(len(literal))); err != nil {
						return false, err
					}
					if equalBytes(data[pos:pos+t.min], literal, p.nocase) {
						allowed = false
						break
					}
				}
				if allowed {
					expanded = true
					matched, err = visit(node.next, pos+t.min)
				}
			}
		default:
			if pos < len(data) && data[pos]&t.mask == t.value {
				expanded = true
				matched, err = visit(node.next, pos+1)
			}
		}
		if !matched && err == nil && expanded {
			failed[state] = true
		}
		return matched, err
	}
	matched, err := visit(p.graph.start, start)
	if err == errPatternStates {
		return p.forwardMatchesAt(data, start, w) // as the matcher does
	}
	return matched, err
}

// legacyCounts mirrors patternCounts for the previous matcher.
type legacyCounts struct {
	m          *legacyMatcher
	data       []byte
	info       *objectInfo
	w          *matchWork
	candidates *legacyCandidates
	counts     map[int]*legacyCount
	c          []int
	first      []uint32
	why        []string
}

func (m *legacyMatcher) newCounts(data []byte, info *objectInfo, w *matchWork) (*legacyCounts, error) {
	c, err := m.candidates(data, w)
	if err != nil {
		return nil, err
	}
	return &legacyCounts{m: m, data: data, info: info, w: w, candidates: c, counts: map[int]*legacyCount{}}, nil
}

func (lc *legacyCounts) rule(r *rule) error {
	lc.c, lc.first, lc.why = lc.c[:0], lc.first[:0], lc.why[:0]
	for i, id := range r.patterns {
		n, at, why := 0, missingOffset, ""
		pattern := &lc.m.d.patterns[id]
		switch {
		case pattern.holder != nil:
			if n = pattern.holder.count(lc.c, lc.info.typ); n == countUnsupported {
				why = pattern.holder.reason
			}
		case !lc.candidates.has(id):
		case pattern.offset.relative() && lc.info.status == exeFailed:
		case pattern.offset.kind == offsetVersionInfo:
			if lc.info.target == 1 {
				anywhere := *pattern
				anywhere.offset = offset{kind: offsetAny}
				pc := legacyNewCount(&anywhere, lc.data, nil, lc.candidates.span(id))
				if err := legacyCountUntil(&anywhere, lc.data, lc.w, pc, 1); err != nil {
					return err
				}
				if v, _ := pc.value(1); v != 0 {
					n, why = countUnsupported, UnknownVersionInfo
				}
			}
		case pattern.offset.executable() && lc.info.status == exeUnknown:
			// Unknown when the pattern matches at any start of its span.
			anywhere := *pattern
			anywhere.offset = offset{kind: offsetAny}
			pc := legacyNewCount(&anywhere, lc.data, nil, lc.candidates.span(id))
			if err := legacyCountUntil(&anywhere, lc.data, lc.w, pc, 1); err != nil {
				return err
			}
			if v, _ := pc.value(1); v != 0 {
				n = countNeedsHeader
			}
		default:
			maxCount := r.countCap(i)
			if r.bytecode != nil {
				maxCount = len(lc.data) + 1
			}
			pc := lc.counts[id]
			if pc == nil {
				pc = legacyNewCount(pattern, lc.data, lc.info.meta, lc.candidates.span(id))
				lc.counts[id] = pc
			}
			if err := legacyCountUntil(pattern, lc.data, lc.w, pc, maxCount); err != nil {
				return err
			}
			n, at = pc.value(maxCount)
		}
		lc.c, lc.first, lc.why = append(lc.c, n), append(lc.first, at), append(lc.why, why)
	}
	return nil
}

// match is the previous Match with an explicit budget, reporting steps used.
func (m *legacyMatcher) match(ctx context.Context, data []byte, limit int64) (string, int64, error) {
	w := &matchWork{ctx: ctx, limit: limit}
	name, err := m.matchWork(w, data)
	return name, w.steps, err
}

func (m *legacyMatcher) matchWork(w *matchWork, data []byte) (string, error) {
	d, ctx := m.d, w.ctx
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var pending error
	incomplete := func(err error) {
		if pending == nil {
			pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, err)
		}
	}
	if d.hashes.len() > 0 || d.allow.len() > 0 {
		digests, err := hashesFor(ctx, data)
		if err != nil {
			return "", err
		}
		for _, digest := range digests {
			if d.allow.lookup(digest, int64(len(data))) != "" {
				return "", nil
			}
		}
		for _, digest := range digests {
			if name := d.hashes.lookup(digest, int64(len(data))); name != "" {
				return name, nil
			}
		}
	}
	info := d.objectInfo(data, MatchContext{Type: d.rawScanType(data)})
	if info.target == 1 && d.sections.len() > 0 && bytes.HasPrefix(data, []byte("MZ")) {
		if info.status != exeParsed {
			incomplete(errExecutableHeader)
		} else {
			for _, s := range info.meta.sections {
				if s.size == 0 || !d.sectionSizes[int64(s.size)] && !d.sectionSizes[-1] {
					continue
				}
				digests, err := hashesFor(ctx, data[s.raw:int64(s.raw)+int64(s.size)])
				if err != nil {
					return "", err
				}
				for _, digest := range digests {
					if name := d.sections.lookup(digest, int64(s.size)); name != "" {
						return name, nil
					}
				}
			}
		}
	}
	if reason := d.unknownHooks[info.target]; reason != "" {
		incomplete(unknownBecause(reason))
	}
	lc, err := m.newCounts(data, &info, w)
	if err != nil {
		return "", err
	}
	evaluate := func(r *rule) (string, error, error) {
		if err := w.step(1); err != nil {
			return "", nil, err
		}
		applies, descriptorsUnknown := legacyApplies(r, &info, len(data))
		if !applies || !r.inLayers(info.layers) {
			return "", nil, nil
		}
		if err := lc.rule(r); err != nil {
			return "", nil, err
		}
		c, first := lc.c, lc.first
		state := logicFalse
		if r.expr == nil {
			switch {
			case c[0] < 0:
				state = logicUnknown
			case c[0] > 0:
				state = logicTrue
			}
		} else {
			state = r.expr.evaluate(c).state
		}
		if state == logicFalse {
			return "", nil, nil
		}
		if state == logicUnknown {
			return "", unknownCause(c, lc.why), nil
		}
		if descriptorsUnknown {
			return "", errExecutableHeader, nil
		}
		if r.icons != nil {
			switch d.iconCondition(&info, data, *r.icons) {
			case logicFalse:
				return "", nil, nil
			case logicUnknown:
				return "", unknownBecause(UnknownIconGroup), nil
			}
		}
		switch {
		case r.unknown != "":
			return "", unknownBecause(r.unknown), nil
		case r.bytecode == nil:
			return r.name, nil, nil
		}
		matches := bytecode.EmptyMatchContext()
		for i := range r.patterns {
			if c[i] < 0 {
				return "", unknownCause(c, lc.why), nil
			}
			if c[i] > 0 {
				if uint64(c[i]) > uint64(^uint32(0)) {
					return "", ErrMatchLimit, nil
				}
				matches.Counts[i] = uint32(c[i])
				matches.Offsets[i] = first[i]
			}
		}
		remaining := w.limit - w.steps
		if remaining <= 0 {
			return "", nil, ErrMatchLimit
		}
		allowance := min(remaining, 1_000_000)
		result, err := r.bytecode.RunWithMatchContext(ctx, 0, nil, data, matches, uint64(allowance))
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return "", nil, ctxErr
			}
			if e := w.step(allowance); e != nil || w.steps >= w.limit {
				return "", nil, ErrMatchLimit
			}
			return "", fmt.Errorf("UTMStack EDR bytecode analysis failed: %w", err), nil
		}
		if err = w.step(int64(result.WorkUsed)); err != nil {
			return "", nil, err
		}
		if !result.VirusNameSet {
			return "", nil, nil
		}
		if err := validName(result.VirusName); err != nil {
			return "", fmt.Errorf("UTMStack EDR bytecode detection label: %w", err), nil
		}
		for _, prefix := range []string{"PUA.", "Heuristics.", "BC.Heuristics.", "Weak."} {
			if strings.HasPrefix(result.VirusName, prefix) {
				return "", fmt.Errorf("UTMStack EDR bytecode detection classification: %w", bytecode.ErrUnsupported), nil
			}
		}
		return result.VirusName, nil, nil
	}
	roots := []int{0}
	if info.target != 0 {
		roots = []int{info.target, 0}
	}
	for _, root := range roots {
		for i := range d.rules {
			r := &d.rules[i]
			if r.target != root {
				continue
			}
			name, cause, err := evaluate(r)
			if err != nil {
				return "", err
			}
			if name != "" {
				return name, nil
			}
			if cause != nil {
				incomplete(cause)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", pending
}
