package signatures

import (
	"cmp"
	"slices"
)

const nativeACTrackSlots = 9

type nativeACOrder struct{ rule, sub, variant, part int }

func compareNativeACOrder(a, b nativeACOrder) int {
	for _, pair := range [][2]int{{a.rule, b.rule}, {a.sub, b.sub}, {a.variant, b.variant}, {a.part, b.part}} {
		if n := cmp.Compare(pair[0], pair[1]); n != 0 {
			return n
		}
	}
	return 0
}

type nativeACRegistration struct {
	pattern, rule, sub int
	model              *nativeACPattern
	offset             offset
	standalone         bool
	cap                int
}

type nativeACGroup struct {
	head          *nativeACPart
	partno, count int
	order         nativeACOrder
}

type nativeACRoot struct {
	models   map[int]*nativeACPattern
	groups   map[[16]byte]*nativeACGroup
	literals bitset
	patterns []bodyPattern
}

func newNativeACRoot(registrations []nativeACRegistration, patterns []bodyPattern) *nativeACRoot {
	r := &nativeACRoot{models: map[int]*nativeACPattern{}, groups: map[[16]byte]*nativeACGroup{}}
	r.patterns = patterns
	r.literals = newBitset(len(patterns))
	for _, reg := range registrations {
		if reg.model == nil {
			p := &patterns[reg.pattern]
			part := nativeACPart{literal: p.prefix, nocase: p.nocase}
			id := part.fingerprint()
			g := r.groups[id]
			order := nativeACOrder{reg.rule, reg.sub, 0, 0}
			if g == nil {
				g = &nativeACGroup{order: order}
				r.groups[id] = g
			}
			g.count++
			if g.partno > 0 || g.partno == 0 && compareNativeACOrder(order, g.order) < 0 {
				g.head, g.partno, g.order = nil, 0, order
			}
			r.literals.set(reg.pattern)
			continue
		}
		r.models[reg.pattern] = reg.model
		for vi, variant := range reg.model.variants {
			for pi := range variant {
				part := &variant[pi]
				partno := 0
				if len(variant) > 1 {
					partno = pi + 1
				}
				order := nativeACOrder{reg.rule, reg.sub, vi, pi}
				g := r.groups[part.identity]
				if g == nil {
					g = &nativeACGroup{head: part, partno: partno, order: order}
					r.groups[part.identity] = g
				}
				g.count++
				if partno < g.partno || partno == g.partno && compareNativeACOrder(order, g.order) < 0 {
					g.head, g.partno, g.order = part, partno, order
				}
			}
		}
	}
	return r
}

type nativeACHistory struct{ rows [][nativeACTrackSlots + 1]int }

func (h *nativeACHistory) accept(parts []nativeACPart, pi, start, end int) (int, bool) {
	if len(parts) == 1 {
		return start, true
	}
	if h.rows == nil {
		if pi != 0 {
			return 0, false
		}
		h.rows = make([][nativeACTrackSlots + 1]int, len(parts))
		for i := range h.rows {
			for j := 1; j <= nativeACTrackSlots; j++ {
				h.rows[i][j] = -1
			}
		}
	}
	found := 0
	if pi > 0 {
		previous := &h.rows[pi-1]
		if previous[0] == 0 {
			return 0, false
		}
		for j := 1; j <= nativeACTrackSlots && previous[j] >= 0; j++ {
			distance := start - previous[j]
			if distance >= 0 && (parts[pi].maxGap == 0 || distance <= parts[pi].maxGap) && distance >= parts[pi].minGap {
				found = j
				break
			}
		}
	}
	if pi == 1 && found > 1 {
		h.rows[0][1], h.rows[0][found] = h.rows[0][found], h.rows[0][1]
		last := len(parts) - 1
		h.rows[last][1], h.rows[last][found] = h.rows[last][found], h.rows[last][1]
	}
	if pi == 0 || found > 0 && pi != len(parts)-1 {
		row := &h.rows[pi]
		if row[0] == nativeACTrackSlots {
			row[0] = 1
		}
		row[0]++
		row[row[0]] = end
		if pi == 0 {
			h.rows[len(parts)-1][row[0]] = start
		}
		return 0, false
	}
	if found > 0 && pi == len(parts)-1 {
		return h.rows[len(parts)-1][1], true
	}
	return 0, false
}

type nativeACState struct {
	model             *nativeACPattern
	pc                patternCount
	history           []nativeACHistory
	events            [][]nativeACEvent
	lo, hi            int
	starts            []int // VI offsets; nil is an ordinary range
	disabled, unknown bool
	cap               int
}

type nativeACEvent struct {
	serial   uint64
	complete bool
}

func (s *nativeACState) accepts(pi, start int) bool {
	if s.disabled {
		return false
	}
	if s.starts != nil {
		_, ok := slices.BinarySearch(s.starts, start)
		return ok
	}
	return pi != 0 || start >= s.lo && start <= s.hi
}

func (s *nativeACState) accept(vi, pi, start, end int) bool {
	if !s.accepts(pi, start) {
		return false
	}
	realoff, complete := s.history[vi].accept(s.model.variants[vi], pi, start, end)
	if !complete {
		return false
	}
	partial := len(s.model.variants[vi]) > 1
	if s.pc.n == 0 {
		s.pc.first = uint32(realoff)
	}
	if s.pc.n == 0 || realoff > s.pc.last || partial && realoff == s.pc.last {
		s.pc.n++
		s.pc.last = realoff
	}
	return true
}

type nativeACRef struct {
	state                 *nativeACState
	partno, variant, part int
	rule                  int // negative for a logical subsignature
	order                 nativeACOrder
}

type nativeACActiveGroup struct {
	group *nativeACGroup
	refs  []nativeACRef
	head  *nativeACPart
}

// possibleKey rejects a trie event before walking its body when no legal
// start can follow the object's offset or any retained predecessor. These
// are conservative width bounds: alternatives still choose their exact
// start in matchKey, and accept checks the chosen start again.
func (r *nativeACRef) possibleKey(p *nativeACPart, keyStart int) bool {
	s := r.state
	if s.pc.n >= s.cap {
		return false
	}
	lo, hi := max(0, keyStart-p.prefixMax), keyStart-p.prefixMin
	if hi < lo {
		return false
	}
	if s.starts != nil {
		at, _ := slices.BinarySearch(s.starts, lo)
		return at < len(s.starts) && s.starts[at] <= hi
	}
	if r.part == 0 {
		return lo <= s.hi && hi >= s.lo
	}
	h := &s.history[r.variant]
	if h.rows == nil || h.rows[r.part-1][0] == 0 {
		return false
	}
	part := &s.model.variants[r.variant][r.part]
	for _, end := range h.rows[r.part-1][1:] {
		if end < 0 {
			break
		}
		if hi >= end+part.minGap && (part.maxGap == 0 || lo <= end+part.maxGap) {
			return true
		}
	}
	return false
}

type nativeACResult struct {
	counts map[int]*patternCount
	order  []int
	why    map[int]string
}

// scan builds a small lookup for the object's candidate parts. Registry
// ranks still include every loaded registration, including offset-ineligible
// and absent patterns. Three-byte trie nodes precede their two-byte failure
// nodes. Each group validates once, then runs its part-number/source chain.
func (root *nativeACRoot) scan(data []byte, regs []nativeACRegistration, info *objectInfo, w *matchWork) (*nativeACResult, error) {
	if err := w.ctx.Err(); err != nil {
		return nil, err
	}
	out := &nativeACResult{counts: map[int]*patternCount{}, why: map[int]string{}}
	states := map[int]*nativeACState{}
	groups := map[[16]byte]*nativeACActiveGroup{}
	for _, reg := range regs {
		if err := w.step(1); err != nil {
			return nil, err
		}
		s := states[reg.pattern]
		if s == nil {
			s = &nativeACState{model: reg.model, pc: patternCount{first: missingOffset, done: true}, history: make([]nativeACHistory, len(reg.model.variants)), events: make([][]nativeACEvent, len(reg.model.variants))}
			for vi, parts := range reg.model.variants {
				s.events[vi] = make([]nativeACEvent, len(parts))
			}
			s.lo, s.hi, _ = reg.offset.bounds(len(data), info.meta)
			_, _, valid := reg.offset.bounds(len(data), info.meta)
			s.disabled = !valid
			if reg.offset.relative() && info.status == exeFailed {
				s.disabled = true
			}
			if reg.offset.executable() && info.status == exeUnknown {
				s.lo, s.hi, s.disabled, s.unknown = 0, len(data), false, true
			}
			if reg.offset.kind == offsetVersionInfo {
				if info.target != 1 || info.status == exeFailed {
					s.disabled = true
				} else if starts, ok := info.versionOffsets(data); ok {
					if err := w.step(int64(len(starts))); err != nil {
						return nil, err
					}
					s.starts = make([]int, len(starts))
					copy(s.starts, starts)
					slices.Sort(s.starts)
					s.disabled = false
				} else {
					s.lo, s.hi, s.disabled, s.unknown = 0, len(data), false, true
				}
			}
			states[reg.pattern] = s
			out.counts[reg.pattern] = &s.pc
		}
		s.cap = max(s.cap, reg.cap)
		if s.disabled {
			continue
		}
		for vi, variant := range reg.model.variants {
			for pi := range variant {
				part := &variant[pi]
				g := groups[part.identity]
				if g == nil {
					g = &nativeACActiveGroup{group: root.groups[part.identity]}
					g.head = g.group.head
					if g.head == nil {
						g.head = part
					}
					groups[part.identity] = g
				}
				partno := 0
				if len(variant) > 1 {
					partno = pi + 1
				}
				ri := -1
				if reg.standalone {
					ri = reg.rule
				}
				g.refs = append(g.refs, nativeACRef{s, partno, vi, pi, ri, nativeACOrder{reg.rule, reg.sub, vi, pi}})
			}
		}
	}
	if len(groups) == 0 {
		return out, w.ctx.Err()
	}
	var exact, folded [2]map[uint32][]*nativeACActiveGroup
	put := func(table *[2]map[uint32][]*nativeACActiveGroup, g *nativeACActiveGroup) {
		p := g.head
		i := p.depth - 2
		if table[i] == nil {
			table[i] = map[uint32][]*nativeACActiveGroup{}
		}
		key := uint32(0)
		for j := 0; j < p.depth; j++ {
			b := p.key[j]
			if p.nocase && asciiLetter(b) {
				b |= 32
			}
			key = key<<8 | uint32(b)
		}
		table[i][key] = append(table[i][key], g)
	}
	for _, g := range groups {
		slices.SortFunc(g.refs, func(a, b nativeACRef) int {
			if n := cmp.Compare(a.partno, b.partno); n != 0 {
				return n
			}
			return compareNativeACOrder(a.order, b.order)
		})
		if g.head.nocase {
			put(&folded, g)
		} else {
			put(&exact, g)
		}
	}
	finish := func() (*nativeACResult, error) {
		for id, s := range states {
			if s.unknown && s.pc.n > 0 {
				s.pc.n = countNeedsHeader
				s.pc.first = missingOffset
				for _, reg := range regs {
					if reg.pattern == id && reg.offset.kind == offsetVersionInfo {
						s.pc.n = countUnsupported
						out.why[id] = UnknownVersionInfo
						break
					}
				}
			}
		}
		if err := w.ctx.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}
	compareGroup := func(a, b *nativeACActiveGroup) int {
		if n := cmp.Compare(a.group.partno, b.group.partno); n != 0 {
			return n
		}
		if n := cmp.Compare(b.group.count, a.group.count); n != 0 {
			return n
		}
		return compareNativeACOrder(a.group.order, b.group.order)
	}
	for _, table := range [][2]map[uint32][]*nativeACActiveGroup{exact, folded} {
		for _, m := range table {
			for _, list := range m {
				slices.SortFunc(list, compareGroup)
			}
		}
	}
	seenRules := map[int]bool{}
	var serial uint64
	var keys, foldKeys uint32
	for at, b := range data {
		if err := w.step(1); err != nil {
			return nil, err
		}
		keys = (keys<<8 | uint32(b)) & 0xffffff
		if asciiLetter(b) {
			b |= 32
		}
		foldKeys = (foldKeys<<8 | uint32(b)) & 0xffffff
		for depth := 3; depth >= 2; depth-- {
			if at+1 < depth {
				continue
			}
			mask := uint32(0xffffff)
			if depth == 2 {
				mask = 0xffff
			}
			a, c := exact[depth-2][keys&mask], folded[depth-2][foldKeys&mask]
			for len(a) > 0 || len(c) > 0 {
				var g *nativeACActiveGroup
				if len(c) == 0 || len(a) > 0 && compareGroup(a[0], c[0]) <= 0 {
					g, a = a[0], a[1:]
				} else {
					g, c = c[0], c[1:]
				}
				if err := w.step(1); err != nil {
					return nil, err
				}
				possible := false
				for _, ref := range g.refs {
					if err := w.step(1); err != nil {
						return nil, err
					}
					if ref.possibleKey(g.head, at+1-depth) {
						possible = true
						break
					}
				}
				if !possible {
					continue
				}
				start, end, ok, err := g.head.matchKey(data, at+1, w)
				if err != nil {
					return nil, err
				}
				if !ok {
					continue
				}
				// Duplicate registrations have independent but identical
				// histories. Advance the shared state once per part event.
				serial++
				for _, ref := range g.refs {
					if err := w.step(1); err != nil {
						return nil, err
					}
					if ref.state.pc.n >= ref.state.cap {
						// The part may have completed earlier in this same
						// event before a duplicate standalone registration.
						cached := &ref.state.events[ref.variant][ref.part]
						if cached.serial != serial {
							continue
						}
					}
					cached := &ref.state.events[ref.variant][ref.part]
					if cached.serial != serial {
						cached.serial = serial
						cached.complete = ref.state.accept(ref.variant, ref.part, start, end)
					}
					hit := cached.complete
					if hit && ref.rule >= 0 && !seenRules[ref.rule] {
						seenRules[ref.rule] = true
						out.order = append(out.order, ref.rule)
						// Default native AC scanning stops at the first
						// standalone match. Logical triggers after that
						// point must remain uncounted, including when a
						// later classifier treats the label as weak.
						if !ref.state.unknown {
							return finish()
						}
					}
				}
			}
		}
	}
	return finish()
}
