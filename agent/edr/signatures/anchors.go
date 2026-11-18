package signatures

import (
	"bytes"
	"errors"
	"hash/maphash"
	"math"
)

// The anchor index finds, in one pass over an object, every occurrence of
// every anchor literal. Each pattern then verifies only the start positions
// its own anchor occurrences allow, and each (pattern, start) pair at most
// once, instead of every pattern scanning the span between its anchor's first
// and last occurrence separately.
//
// All of it is charged to the object's work budget: every filter probe (at
// most two per byte), table lookup, verified anchor and anchored pattern of a
// found anchor, every window visited and every start tried. Occurrences are
// kept as arithmetic runs, at most maxStoredRuns per object; past that, the
// patterns of the literals that could not be kept fall back to trying every
// start in their span, which gives the same results.

// anchorGroup is one distinct anchor literal and every pattern anchored on it,
// with where the literal starts relative to that pattern's match start. keyAt
// is where, inside the literal, its lookup key starts: the rarest window of up
// to four bytes, so crowded keys such as zero runs are avoided when the
// literal has something better. A folded literal is lowercased and compared
// with lowercased data.
type anchorGroup struct {
	literal []byte
	members []anchorMember
	keyAt   int32
	fold    bool
}
type anchorMember struct {
	id             int32
	minPre, maxPre int32 // maxPre < 0: unbounded
}

// newAnchorMember stores an anchor's distances in 32 bits. Gaps can make
// them larger; such a distance may only widen the start span, so a large
// minimum is capped and a large maximum becomes unbounded.
func newAnchorMember(id int, a patternAnchor) anchorMember {
	m := anchorMember{id: int32(id), minPre: int32(min(a.minPre, math.MaxInt32)), maxPre: -1}
	if a.maxPre >= 0 && a.maxPre <= math.MaxInt32 {
		m.maxPre = int32(a.maxPre)
	}
	return m
}

// anchorBucket holds the anchor groups that share a lookup key. A crowded
// bucket (identical keys such as zero runs or wide "C:") moves its longer
// exact anchors into child buckets keyed by the next four bytes after the key,
// or by the next byte for anchors ending sooner, so a hit only verifies
// anchors that agree with the data that follows. Folded anchors stay.
type anchorBucket struct {
	groups []*anchorGroup
	chunk  map[uint32]*anchorBucket
	single map[byte]*anchorBucket
}

const crowdedBucket = 16
const anchorBucketSlabSize = 256

// split distributes a crowded bucket whose anchors agree on the depth bytes
// that begin at their key.
func (b *anchorBucket) split(depth int, index *anchorIndex) {
	if len(b.groups) <= crowdedBucket {
		return
	}
	var keep []*anchorGroup
	for _, g := range b.groups {
		tail := g.literal[g.keyAt:]
		switch rest := len(tail) - depth; {
		case rest <= 0 || g.fold:
			keep = append(keep, g)
		case rest >= 4:
			if b.chunk == nil {
				b.chunk = make(map[uint32]*anchorBucket)
			}
			key := packKey(tail[depth : depth+4])
			if b.chunk[key] == nil {
				b.chunk[key] = index.newBucket()
			}
			b.chunk[key].groups = append(b.chunk[key].groups, g)
		default:
			if b.single == nil {
				b.single = make(map[byte]*anchorBucket)
			}
			key := tail[depth]
			if b.single[key] == nil {
				b.single[key] = index.newBucket()
			}
			b.single[key].groups = append(b.single[key].groups, g)
		}
	}
	b.groups = keep
	for _, child := range b.chunk {
		child.split(depth+4, index)
	}
	for _, child := range b.single {
		child.split(depth+1, index)
	}
}

// anchorIndex finds buckets by the data at a position. Four-byte keys sit
// behind a filter with one bit per hash of the key; keys of one to three bytes
// sit behind a filter with one bit per possible first two bytes. A folded key
// is entered once per spelling of its letters, so the scan needs no lowercased
// pass: each position costs one probe of each filter, and most stop there.
type anchorIndex struct {
	long        map[uint32]*anchorBucket
	longBits    []uint64
	short       [3]map[uint32]*anchorBucket // by key length - 1
	shortBits   []uint64
	bucketSlabs [][]anchorBucket
	bucketUsed  int
}

// Buckets are immutable after compilation. Keeping them in index-owned slabs
// avoids a separate small allocation and GC object for every lookup key.
func (x *anchorIndex) newBucket() *anchorBucket {
	if len(x.bucketSlabs) == 0 || x.bucketUsed == anchorBucketSlabSize {
		x.bucketSlabs = append(x.bucketSlabs, make([]anchorBucket, anchorBucketSlabSize))
		x.bucketUsed = 0
	}
	b := &x.bucketSlabs[len(x.bucketSlabs)-1][x.bucketUsed]
	x.bucketUsed++
	return b
}

func longBit(key uint32) uint32 { return (key * 0x9e3779b1) >> 8 } // 24 bits

func setBit(bits []uint64, bit uint32)      { bits[bit>>6] |= 1 << (bit & 63) }
func hasBit(bits []uint64, bit uint32) bool { return bits[bit>>6]&(1<<(bit&63)) != 0 }

func (x *anchorIndex) add(g *anchorGroup) {
	n := min(4, len(g.literal))
	key := g.literal[g.keyAt : int(g.keyAt)+n]
	// At most four letters produce 16 case variants. Keep their packed keys
	// on the stack instead of allocating byte slices for every anchor group.
	var variants [16]uint32
	variants[0] = packKey(key)
	variantCount := 1
	if g.fold {
		for i, c := range key {
			if !asciiLetter(c) {
				continue
			}
			flip := uint32(0x20) << uint(8*(n-1-i))
			for j, count := 0, variantCount; j < count; j++ {
				variants[variantCount] = variants[j] ^ flip
				variantCount++
			}
		}
	}
	for _, k := range variants[:variantCount] {
		table := &x.long
		if n == 4 {
			if x.longBits == nil {
				x.longBits = make([]uint64, 1<<24/64)
			}
			setBit(x.longBits, longBit(k))
		} else {
			if x.shortBits == nil {
				x.shortBits = make([]uint64, 1<<16/64)
			}
			if n == 1 {
				for next := uint32(0); next < 256; next++ {
					setBit(x.shortBits, k<<8|next)
				}
			} else {
				setBit(x.shortBits, k>>uint(8*(n-2)))
			}
			table = &x.short[n-1]
		}
		if *table == nil {
			*table = map[uint32]*anchorBucket{}
		}
		if (*table)[k] == nil {
			(*table)[k] = x.newBucket()
		}
		(*table)[k].groups = append((*table)[k].groups, g)
	}
}

func (d *Database) buildAnchors() error {
	d.sectionSizes = d.sections.sizes()
	d.patternIDs = nil
	d.expressionIDs = nil
	// Hash the immutable literal bytes for temporary deduplication. A string
	// key would copy every distinct literal while the final anchor index is
	// being built, overlapping another large full-feed allocation. Collisions
	// are resolved by comparing the bytes, never by trusting the hash alone.
	seed := maphash.MakeSeed()
	groups := [2]map[uint32]uint32{{}, {}}
	if len(d.patterns) >= 100_000 && len(d.anchorGroups) == 0 {
		d.anchorGroups = make([]*anchorGroup, 0, len(d.patterns))
	}
	next := make([]uint32, 0, cap(d.anchorGroups))
	for id, p := range d.patterns {
		if p.compiledAnchorCount() == 0 {
			if p.everywhere() {
				d.unanchored = append(d.unanchored, id)
			}
			continue
		}
		for i := 0; i < p.compiledAnchorCount(); i++ {
			a := p.compiledAnchor(i)
			fold := 0
			if a.fold {
				fold = 1
			}
			key := uint32(maphash.Bytes(seed, a.literal))
			g := findAnchorGroup(groups[fold], next, d.anchorGroups, key, a.literal)
			if g == nil {
				if uint64(len(d.anchorGroups)) >= uint64(^uint32(0)) {
					return errors.New("anchor group count limit exceeded")
				}
				g = &anchorGroup{literal: a.literal, fold: a.fold, keyAt: keyWindow(a.literal, a.fold)}
				next = append(next, groups[fold][key])
				groups[fold][key] = uint32(len(d.anchorGroups)) + 1
				d.anchorGroups = append(d.anchorGroups, g)
				d.index.add(g)
			}
			g.members = append(g.members, newAnchorMember(id, a))
		}
		// The index now owns the literals and all required distances. Keep only
		// the filtering flag on the pattern, not a second anchor array for the
		// entire lifetime of every live generation.
		d.patterns[id].releaseCompiledAnchors()
	}
	for _, bucket := range d.index.long {
		bucket.split(4, &d.index)
	}
	for n, table := range d.index.short {
		for _, bucket := range table {
			bucket.split(n+1, &d.index)
		}
	}
	return nil
}

func findAnchorGroup(table map[uint32]uint32, next []uint32, all []*anchorGroup, key uint32, literal []byte) *anchorGroup {
	for slot := table[key]; slot != 0; slot = next[slot-1] {
		g := all[slot-1]
		if bytes.Equal(g.literal, literal) {
			return g
		}
	}
	return nil
}

func packKey(b []byte) uint32 {
	key := uint32(0)
	for _, c := range b {
		key = key<<8 | uint32(c)
	}
	return key
}

func lowerASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 'a' - 'A'
	}
	return b
}

// bitset marks candidate pattern ids.
type bitset []uint64

func newBitset(n int) bitset    { return make(bitset, (n+63)/64) }
func (b bitset) set(i int)      { b[i>>6] |= 1 << uint(i&63) }
func (b bitset) has(i int) bool { return b[i>>6]&(1<<uint(i&63)) != 0 }

// startSpan bounds the start positions at which a pattern can match.
type startSpan struct{ lo, hi int }

var anySpan = startSpan{0, math.MaxInt}

// occRun is an arithmetic run of anchor occurrences: n positions from start,
// stride apart. Periodic data (zero or 0xff runs, repeated instructions)
// yields long runs, so dense occurrences take little memory.
type occRun struct{ start, stride, n int }

func (r occRun) last() int { return r.start + r.stride*(r.n-1) }

// groupHits is one object's occurrences of one anchor literal. The runs are
// dropped (overflow) when the object's stored runs reach maxStoredRuns; its
// patterns then try every start in the span between first and last.
type groupHits struct {
	first, last int
	count       int64
	runs        []occRun
	overflow    bool
}

// maxStoredRuns bounds the occurrence memory of one object's scan (about 24
// MiB). It is a variable only so tests can reach it with small inputs.
var maxStoredRuns = 1 << 20

// anchorRef is one anchor of a candidate pattern and its occurrences.
type anchorRef struct {
	hits           *groupHits
	minPre, maxPre int32
}

// candidateSet is one object's candidate patterns: those whose anchor occurs
// in full, or that have none. anchored lists the former once each. span bounds
// each anchored candidate's matches to the start positions its anchor
// occurrences allow; refs lists the occurrences themselves.
type candidateSet struct {
	marked   bitset
	anchored []int32
	spans    map[int32]startSpan
	refs     map[int32][]anchorRef
}

func (c *candidateSet) has(id int) bool { return c.marked.has(id) }
func (c *candidateSet) span(id int) startSpan {
	if s, ok := c.spans[int32(id)]; ok {
		return s
	}
	return anySpan
}

// anchorScan is the state of one pass over an object. order lists the found
// literals by first occurrence, so what follows the pass never depends on map
// iteration order: the order of a pattern's anchors decides how its windows
// are split, and so the work charged.
type anchorScan struct {
	data   []byte
	w      *matchWork
	hits   map[*anchorGroup]*groupHits
	order  []*anchorGroup
	stored int
}

func (s *anchorScan) record(g *anchorGroup, at int) {
	h := s.hits[g]
	if h == nil {
		h = &groupHits{first: at}
		s.hits[g] = h
		s.order = append(s.order, g)
	}
	h.last = at
	h.count++
	if h.overflow {
		return
	}
	if n := len(h.runs); n > 0 {
		r := &h.runs[n-1]
		if r.n == 1 {
			r.stride, r.n = at-r.start, 2
			return
		}
		if at == r.start+r.stride*r.n {
			r.n++
			return
		}
	}
	if s.stored >= maxStoredRuns {
		s.stored -= len(h.runs)
		h.runs, h.overflow = nil, true
		return
	}
	s.stored++
	h.runs = append(h.runs, occRun{start: at, n: 1})
}

// walk verifies the anchors of a bucket whose key starts at keyPos, then
// follows the child buckets that agree with the data after its depth bytes.
// Positions are visited in order and a literal is found at most once per
// position, so each literal's occurrences arrive in order.
func (s *anchorScan) walk(keyPos, depth int, b *anchorBucket) error {
	data := s.data
	for _, g := range b.groups {
		if err := s.w.step(1); err != nil {
			return err
		}
		at := keyPos - int(g.keyAt)
		if at < 0 || at+len(g.literal) > len(data) {
			continue
		}
		matched := true
		if !g.fold {
			matched = bytes.Equal(data[at:at+len(g.literal)], g.literal)
		} else {
			for i, c := range g.literal {
				if lowerASCII(data[at+i]) != c {
					matched = false
					break
				}
			}
		}
		if matched {
			s.record(g, at)
		}
	}
	if b.single != nil && keyPos+depth < len(data) {
		if child := b.single[data[keyPos+depth]]; child != nil {
			if err := s.walk(keyPos, depth+1, child); err != nil {
				return err
			}
		}
	}
	if b.chunk != nil && keyPos+depth+4 <= len(data) {
		if child := b.chunk[packKey(data[keyPos+depth:keyPos+depth+4])]; child != nil {
			return s.walk(keyPos, depth+4, child)
		}
	}
	return nil
}

// lookup charges one step and walks the bucket of key, if any.
func (s *anchorScan) lookup(table map[uint32]*anchorBucket, key uint32, keyPos, depth int) error {
	if err := s.w.step(1); err != nil {
		return err
	}
	if b := table[key]; b != nil {
		return s.walk(keyPos, depth, b)
	}
	return nil
}

// scanChunk is how many positions are charged at once, so cancellation and the
// budget are still observed during a long scan.
const scanChunk = 1 << 16

// candidates finds every occurrence of every anchor in one pass that probes
// the index filters at every position. Every filter probe, every table lookup,
// every verified anchor and every anchored pattern of a found anchor is
// charged.
func (d *Database) candidates(data []byte, w *matchWork) (*candidateSet, error) {
	c := &candidateSet{marked: newBitset(len(d.patterns)), spans: map[int32]startSpan{}, refs: map[int32][]anchorRef{}}
	for _, i := range d.unanchored {
		c.marked.set(i)
	}
	s := &anchorScan{data: data, w: w, hits: map[*anchorGroup]*groupHits{}}
	x := &d.index
	probes := 0
	if x.longBits != nil {
		probes++
	}
	if x.shortBits != nil {
		probes++
	}
	for i := 0; i < len(data) && probes > 0; i++ {
		if i%scanChunk == 0 {
			if err := w.step(int64(min(scanChunk, len(data)-i) * probes)); err != nil {
				return nil, err
			}
		}
		if x.longBits != nil && i+4 <= len(data) {
			key := packKey(data[i : i+4])
			if hasBit(x.longBits, longBit(key)) {
				if err := s.lookup(x.long, key, i, 4); err != nil {
					return nil, err
				}
			}
		}
		if x.shortBits == nil || i+1 < len(data) && !hasBit(x.shortBits, uint32(data[i])<<8|uint32(data[i+1])) {
			continue
		}
		for n := 1; n <= 3 && i+n <= len(data); n++ {
			if x.short[n-1] != nil {
				if err := s.lookup(x.short[n-1], packKey(data[i:i+n]), i, n); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, g := range s.order {
		h := s.hits[g]
		for _, m := range g.members {
			if err := w.step(1); err != nil {
				return nil, err
			}
			lo, hi := h.first-int(m.maxPre), h.last-int(m.minPre)
			if m.maxPre < 0 {
				lo = 0
			}
			if span, ok := c.spans[m.id]; ok {
				lo, hi = min(lo, span.lo), max(hi, span.hi)
			}
			c.spans[m.id] = startSpan{lo, hi}
			c.refs[m.id] = append(c.refs[m.id], anchorRef{h, m.minPre, m.maxPre})
			if !c.marked.has(int(m.id)) {
				c.marked.set(int(m.id))
				c.anchored = append(c.anchored, m.id)
			}
		}
	}
	return c, nil
}

// startCount begins counting a candidate pattern. Its start positions are
// bounded by its offset and by the span of its anchor occurrences. When a
// start that no occurrence allows could be state-limited on this object (see
// proneFrom), or the occurrences were not kept, every start in that span is
// tried exactly as it always has been. Otherwise only the starts some anchor
// occurrence allows are tried, when that is expected to be less work. Both
// give identical counts, first offsets and state limits: every match contains
// an anchor occurrence, and a skipped start can neither match nor be limited.
func (c *candidateSet) startCount(p *bodyPattern, id int, data []byte, meta *peMeta) *patternCount {
	pc := p.newCount(data, meta, c.span(id))
	if pc.done || len(data) >= p.proneFrom() {
		return pc
	}
	refs := c.refs[int32(id)]
	if len(refs) == 0 {
		return pc
	}
	var occurrences, widest int64
	for _, r := range refs {
		if r.hits.overflow || r.maxPre < 0 {
			return pc
		}
		occurrences += r.hits.count
		widest = max(widest, int64(r.maxPre-r.minPre)+1)
	}
	// Windowed counting visits every occurrence and at most the starts the
	// windows cover; the span scan visits every start once.
	span := int64(pc.hi-pc.lo) + 1
	if occurrences*int64(len(refs)) >= span {
		return pc
	}
	covered := span
	if widest <= span/occurrences {
		covered = occurrences * widest
	}
	if occurrences*int64(len(refs))+covered >= span {
		return pc
	}
	pc.streams = make([]windowStream, len(refs))
	for i, r := range refs {
		pc.streams[i] = windowStream{runs: r.hits.runs, minPre: int(r.minPre), maxPre: int(r.maxPre)}
	}
	pc.windowed, pc.winHi = true, pc.next-1
	return pc
}

// windowStream walks one anchor's occurrences in order. Each occurrence at a
// allows the starts [a-maxPre, a-minPre].
type windowStream struct {
	runs           []occRun
	run, k         int
	minPre, maxPre int
	positioned     bool
}

func (s *windowStream) peek() (int, bool) {
	if s.run >= len(s.runs) {
		return 0, false
	}
	r := s.runs[s.run]
	return r.start + r.stride*s.k, true
}

func (s *windowStream) advance() {
	if s.k++; s.k >= s.runs[s.run].n {
		s.run, s.k = s.run+1, 0
	}
}

// seek skips the occurrences before target by binary search over the runs,
// charging each probe.
func (s *windowStream) seek(target int, w *matchWork) error {
	lo, hi := 0, len(s.runs)
	for lo < hi {
		if err := w.step(1); err != nil {
			return err
		}
		mid := int(uint(lo+hi) >> 1)
		if s.runs[mid].last() < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	s.run, s.k = lo, 0
	if lo < len(s.runs) {
		if r := s.runs[lo]; r.start < target {
			s.k = (target - r.start + r.stride - 1) / r.stride
		}
	}
	return nil
}

// nextWindow makes the next window of allowed starts current: the one that
// begins first among all anchors' next occurrences. Windows are visited in
// order of their first start and pc.next only grows, so every allowed start is
// tried once, in increasing order, as the span scan tries them.
func (pc *patternCount) nextWindow(w *matchWork) (bool, error) {
	for {
		best, bestLo := -1, 0
		for i := range pc.streams {
			if err := w.step(1); err != nil {
				return false, err
			}
			s := &pc.streams[i]
			if !s.positioned {
				// Occurrences whose windows end before the first start are skipped.
				if err := s.seek(pc.lo+s.minPre, w); err != nil {
					return false, err
				}
				s.positioned = true
			}
			if a, ok := s.peek(); ok && (best < 0 || a-s.maxPre < bestLo) {
				best, bestLo = i, a-s.maxPre
			}
		}
		if best < 0 || bestLo > pc.hi {
			return false, nil
		}
		s := &pc.streams[best]
		a, _ := s.peek()
		s.advance()
		lo, hi := max(bestLo, pc.lo, pc.next), min(a-s.minPre, pc.hi)
		if lo > hi {
			continue
		}
		pc.next, pc.winHi = lo, hi
		return true, nil
	}
}

// countWindows is countUntil restricted to the current and later windows.
func (p *bodyPattern) countWindows(data []byte, w *matchWork, pc *patternCount, want int) error {
	for pc.n < want && !pc.done && !pc.limited {
		if pc.next > pc.winHi {
			more, err := pc.nextWindow(w)
			if err != nil {
				return err
			}
			if !more {
				pc.done = true
			}
			continue
		}
		if err := w.step(1); err != nil {
			return err
		}
		if len(p.prefix) > 0 && !p.nocase {
			bound := min(len(data), pc.winHi+len(p.prefix))
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
				pc.next = pc.winHi + 1
				continue
			}
			pc.next += relative
		}
		matched, err := p.matchesAtWith(data, pc.next, w, &pc.searches)
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
			pc.last = pc.next
		}
		pc.next++
	}
	return nil
}
