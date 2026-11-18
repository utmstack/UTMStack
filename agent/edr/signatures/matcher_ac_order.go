package signatures

import "crypto/sha256"

// Fixed-width AC bodies match when their chosen trie key fires, rather than
// when their full body ends. The pinned profile uses keys of at most three
// bytes, and requires at least two static bytes. Native equal-body grouping
// ignores offsets and includes logical subsignatures, including duplicates.
type fixedACCandidate struct {
	prefix, depth uint32
	group, first  int
}

type fixedACShape struct {
	key           [16]byte
	prefix, depth uint32
}

func fixedACSyntax(body string) bool {
	if len(body) < 4 || len(body)%2 != 0 {
		return false
	}
	for _, c := range body {
		if c != '?' && (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// fixedACShape derives native byte positions from the retained graph, whose
// nodes survive index publication. Raw syntax eligibility is checked first:
// a graph simplified from a gap must not be mistaken for a single AC part.
func (p *bodyPattern) fixedACShape() (fixedACShape, bool) {
	var out fixedACShape
	if !p.nativeFixed || p.holder != nil || p.nocase || p.fullword {
		return out, false
	}
	type maskedByte struct{ value, mask byte }
	var bytes []maskedByte
	appendLiteral := func(b []byte) {
		for _, c := range b {
			bytes = append(bytes, maskedByte{c, 255})
		}
	}
	if p.simple {
		appendLiteral(p.prefix)
	} else if p.graph != nil {
		for id := p.graph.start; id != 0; {
			n := &p.graph.nodes[id]
			t := &n.token
			if t.gap || t.boundary != 0 || t.negative || len(t.alternatives) != 0 || len(n.alternatives) != 0 {
				return out, false
			}
			if len(t.literal) != 0 {
				appendLiteral(t.literal)
			} else {
				bytes = append(bytes, maskedByte{t.value, t.mask})
			}
			id = n.next
		}
	}
	if len(bytes) < 2 {
		return out, false
	}
	// Hash the complete native group body, independent of offset and spelling
	// case. This temporary key is discarded once group counts are published.
	encoded := make([]byte, 0, 2*len(bytes))
	for _, b := range bytes {
		encoded = append(encoded, b.value, b.mask)
	}
	digest := sha256.Sum256(encoded)
	copy(out.key[:], digest[:16])
	wild, zero := false, true
	for i := 0; i < min(3, len(bytes)); i++ {
		if bytes[i].mask != 255 {
			wild = true
			break
		}
		zero = zero && bytes[i].value == 0
	}
	pos := 0
	if wild || zero {
		best, nonzero, nonzeroPos := 0, 0, 0
		for i := 0; i < len(bytes)-1; i++ {
			for j := i; j < min(i+3, len(bytes)); j++ {
				if bytes[j].mask != 255 {
					break
				}
				if j-i+1 >= best {
					best, pos = j-i+1, i
				}
				if bytes[pos].value != 0 || pos+1 < len(bytes) && bytes[pos+1].value != 0 {
					if best >= 3 {
						break
					}
					if best >= 2 && best > nonzero {
						nonzero, nonzeroPos = best, pos
					}
				}
			}
			if best >= 3 && (bytes[pos].value != 0 || bytes[pos+1].value != 0) {
				break
			}
		}
		if nonzero != 0 && pos+1 < len(bytes) && bytes[pos].value == 0 && bytes[pos+1].value == 0 {
			best, pos = nonzero, nonzeroPos
		}
		if best < 2 {
			return out, false
		}
	}
	depth := 0
	for depth < 3 && pos+depth < len(bytes) && bytes[pos+depth].mask == 255 {
		depth++
	}
	if depth < 2 {
		return out, false
	}
	out.prefix, out.depth = uint32(pos), uint32(depth)
	return out, true
}

func acRegistration(d *Database, r *rule) bool {
	return r.expr != nil || r.bytecode != nil || r.unknown == UnknownBytecode || d.standaloneAC(r)
}

// buildFixedAC deliberately enables selection only for wholly modeled roots.
// Unsupported parts can share trie nodes and affect a fixed body's group
// order, even when that unsupported rule never alerts on the scanned object.
// No partial index is published. This gate must expand with the general AC
// part/event implementation, not silently approximate its unmodeled inputs.
func (d *Database) buildFixedAC() {
	d.fixedAC = [maxTargets]map[int]fixedACCandidate{}
	var disabled, standalone [maxTargets]bool
	for i := range d.rules {
		r := &d.rules[i]
		if !acRegistration(d, r) {
			continue
		}
		standalone[r.target] = standalone[r.target] || r.expr == nil && r.bytecode == nil && r.unknown != UnknownBytecode
		if r.dead || r.bytecode != nil || r.unknown != "" {
			disabled[r.target] = true
		}
		for _, id := range r.patterns {
			p := &d.patterns[id]
			if !p.nativeFixed || p.holder != nil || p.nocase || p.fullword {
				disabled[r.target] = true
			}
		}
	}
	type group struct{ count, first int }
	var groups [maxTargets]map[[16]byte]group
	var candidates [maxTargets]map[int]bool
	shapes := map[int]fixedACShape{}
	ordinal := 0
	for i := range d.rules {
		r := &d.rules[i]
		root := r.target
		if disabled[root] || !standalone[root] || !acRegistration(d, r) {
			continue
		}
		for _, id := range r.patterns {
			shape, ok := shapes[id]
			if !ok {
				shape, ok = d.patterns[id].fixedACShape()
				if !ok {
					disabled[root] = true
					break
				}
				shapes[id] = shape
			}
			if groups[root] == nil {
				groups[root] = map[[16]byte]group{}
				candidates[root] = map[int]bool{}
			}
			g := groups[root][shape.key]
			if g.count == 0 {
				g.first = ordinal
			}
			g.count++
			groups[root][shape.key] = g
			ordinal++
			if r.expr == nil {
				candidates[root][id] = true
			}
		}
	}
	for root, ids := range candidates {
		if disabled[root] || len(ids) == 0 {
			continue
		}
		d.fixedAC[root] = make(map[int]fixedACCandidate, len(ids))
		for id := range ids {
			s := shapes[id]
			g := groups[root][s.key]
			d.fixedAC[root][id] = fixedACCandidate{s.prefix, s.depth, g.count, g.first}
		}
	}
}

func (d *Database) matchFixedAC(reached bitset, root int, m *patternCounts, info *objectInfo, w *matchWork) (string, error, error) {
	winner, position := -1, uint64(0)
	var selected fixedACCandidate
	var name string
	var pending error
	for ri := range reached.members() {
		r := &d.rules[ri]
		if r.target != root || r.expr != nil || r.bytecode != nil || r.unknown == UnknownBytecode || !d.standaloneAC(r) {
			continue
		}
		candidate, cause, err := d.evaluateRule(r, m, info, w)
		if err != nil {
			return "", nil, err
		}
		if cause != nil && pending == nil {
			pending = cause
		}
		if candidate == "" {
			continue
		}
		meta, ok := d.fixedAC[root][r.patterns[0]]
		if !ok || len(m.first) != 1 || m.first[0] == missingOffset {
			return "", nil, ErrMatchLimit
		}
		at := uint64(m.first[0]) + uint64(meta.prefix) + uint64(meta.depth)
		tie := meta.depth > selected.depth || meta.depth == selected.depth &&
			(meta.group > selected.group || meta.group == selected.group && meta.first < selected.first)
		if winner < 0 || at < position || at == position && tie {
			winner, position, selected, name = ri, at, meta, candidate
		}
	}
	return name, pending, nil
}
