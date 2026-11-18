package signatures

// buildStandaloneBM models the pinned little-endian profile's BM prefix
// selection. A literal uses its first triplet whose bucket is unoccupied;
// when all its buckets are occupied it retains its original prefix. Generic
// and PE roots have separate buckets. No other matcher registers BM bodies.
// Only nonzero prefixes require retained per-rule storage.
func (d *Database) buildStandaloneBM() {
	var occupied [2][1 << 16]bool
	// Buckets never become empty again. A shared literal can resume after its
	// previous chosen triplet; an exhausted literal always retains prefix0.
	// This makes duplicate-heavy input linear in distinct literal bytes plus
	// rule count instead of repeatedly walking every earlier triplet.
	var next [2][]uint32
	d.bmPrefixes = nil
	for ri := range d.rules {
		r := &d.rules[ri]
		if r.dead || r.expr != nil || r.bytecode != nil || r.unknown == UnknownBytecode || d.standaloneAC(r) {
			continue
		}
		id := r.patterns[0]
		literal := d.patterns[id].prefix
		if len(literal) < 3 {
			// Go also accepts shorter local markers; native BM rejects those
			// source entries, so they cannot occupy a native triplet bucket.
			continue
		}
		if next[r.target] == nil {
			next[r.target] = make([]uint32, len(d.patterns))
		}
		start := next[r.target][id]
		if start == ^uint32(0) {
			continue
		}
		bucket := func(i int) int { return 211*int(literal[i]) + 37*int(literal[i+1]) + int(literal[i+2]) }
		prefix, found := 0, false
		for i := int(start); i+2 < len(literal); i++ {
			if !occupied[r.target][bucket(i)] {
				prefix, found = i, true
				next[r.target][id] = uint32(i + 1)
				break
			}
		}
		if !found {
			next[r.target][id] = ^uint32(0)
		}
		occupied[r.target][bucket(prefix)] = true
		if prefix != 0 {
			if d.bmPrefixes == nil {
				d.bmPrefixes = make(map[int]uint32)
			}
			d.bmPrefixes[ri] = uint32(prefix)
		}
	}
}

// matchStandaloneBM selects a candidate by the position where BM tests its
// chosen triplet. Equal positions correspond to the same matching bucket
// and prefix byte; native insertion puts the later registration first there.
// A work/cancellation failure leaves selection undecided, never clean.
func (d *Database) matchStandaloneBM(reached bitset, root int, m *patternCounts, info *objectInfo, w *matchWork) (string, error, error) {
	winner, position := -1, uint64(0)
	winnerShort := false
	var name string
	var pending error
	for ri := range reached.members() {
		r := &d.rules[ri]
		if r.target != root || r.expr != nil || r.bytecode != nil || r.unknown == UnknownBytecode || d.standaloneAC(r) {
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
		if len(m.first) != 1 || m.first[0] == missingOffset {
			return "", nil, ErrMatchLimit
		}
		at := uint64(m.first[0]) + uint64(d.bmPrefixes[ri])
		short := len(d.patterns[r.patterns[0]].prefix) < 3
		preferTie := ri > winner
		if short || winnerShort {
			// Short local markers extend native's accepted language. Retain
			// their existing first-source preference at the same position.
			preferTie = short && (!winnerShort || ri < winner)
		}
		if winner < 0 || at < position || at == position && preferTie {
			winner, position, name = ri, at, candidate
			winnerShort = short
		}
	}
	return name, pending, nil
}
