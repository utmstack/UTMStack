package signatures

func nativeACBoundary(b byte) byte {
	switch b {
	case '.', '<', '@':
		return 1
	case '\n', '\r', '"', '\'', '=', '>':
		return 2
	case ' ', '-', '/', '_':
		return 3
	}
	return 0
}

func nativeACByte(s nativeACSymbol, b byte, nocase bool) bool {
	if s.mask == 255 && nocase && asciiLetter(s.value) {
		return b|32 == s.value|32 && asciiLetter(b)
	}
	return b&s.mask == s.value
}

func (p *nativeACPart) classBoundary(data []byte, at int, class byte, negative, left bool) bool {
	if class == 0 {
		return true
	}
	ok := false
	switch class {
	case 'B':
		if left {
			ok = at == 0 || nativeACBoundary(data[at-1])&1 != 0
		} else {
			ok = at == len(data) || nativeACBoundary(data[at]) >= 2
		}
	case 'L':
		if left {
			ok = at == 0 || data[at-1] == '\n'
		} else {
			ok = at == len(data) || data[at] == '\n' || data[at] == '\r' && at+1 < len(data) && data[at+1] == '\n'
		}
	case 'W':
		kind := boundaryWordEnd
		if left {
			kind = boundaryWordStart
		}
		if p.wide {
			kind = boundaryWideEnd
			if left {
				kind = boundaryWideStart
			}
		}
		ok = boundaryHolds(kind, data, at)
	}
	return ok != negative
}

func (p *nativeACPart) characterAnchor(data []byte, at int, c nativeACChar, left bool, w *matchWork) (bool, error) {
	if !c.set {
		return true, nil
	}
	for distance := c.min; distance <= c.max; distance++ {
		pos := at + distance
		if left {
			pos = at - distance - 1
		}
		if pos < 0 || pos >= len(data) {
			break
		}
		if err := w.step(1); err != nil {
			return false, err
		}
		if nativeACByte(nativeACSymbol{value: c.value, mask: c.mask}, data[pos], p.nocase) {
			return true, nil
		}
	}
	return false, nil
}

// matchSequence walks from a trie key toward one side of a part. Positions
// are byte boundaries in both directions. Alternatives retain expansion
// order, and invoke the remaining sequence before another branch is tried.
// Work is charged for every visited symbol and alternative. No unbounded
// per-start memo or list of all successful paths is retained.
func (p *nativeACPart) matchSequence(data []byte, symbols []nativeACSymbol, pos int, reverse bool, w *matchWork, done func(int) (bool, error)) (bool, error) {
	var walk func([]nativeACSymbol, int, int, func(int) (bool, error)) (bool, error)
	walk = func(ss []nativeACSymbol, index, at int, next func(int) (bool, error)) (bool, error) {
		for index >= 0 && index < len(ss) {
			if err := w.step(1); err != nil {
				return false, err
			}
			s := ss[index]
			step, byteAt := 1, at
			if reverse {
				step, byteAt = -1, at-1
			}
			if len(s.alts) != 0 {
				if s.negative {
					width, _ := nativeACWidth(s.alts[0])
					lo, hi := at, at+width
					if reverse {
						lo, hi = at-width, at
					}
					if lo < 0 || hi > len(data) {
						return false, nil
					}
					for _, alt := range s.alts {
						ok := true
						for j, b := range alt {
							if err := w.step(1); err != nil {
								return false, err
							}
							if !nativeACByte(b, data[lo+j], p.nocase) {
								ok = false
								break
							}
						}
						if ok {
							return false, nil
						}
					}
					at += step * width
					index += step
					continue
				}
				for _, alt := range s.alts {
					if err := w.step(1); err != nil {
						return false, err
					}
					start := 0
					if reverse {
						start = len(alt) - 1
					}
					ok, err := walk(alt, start, at, func(end int) (bool, error) {
						return walk(ss, index+step, end, next)
					})
					if ok || err != nil {
						return ok, err
					}
				}
				return false, nil
			}
			if byteAt < 0 || byteAt >= len(data) {
				return false, nil
			}
			if s.class != 0 {
				width, ok := 1, false
				switch s.class {
				case 'B':
					ok = nativeACBoundary(data[byteAt]) != 0
				case 'W':
					ok = !alphaNumeric(data[byteAt])
				case 'L':
					ok = data[byteAt] == '\n'
					if data[byteAt] == '\r' && byteAt+1 < len(data) && data[byteAt+1] == '\n' {
						ok, width = true, 2
					}
				}
				if s.negative {
					ok, width = !ok, 1
				}
				if !ok {
					return false, nil
				}
				at += step * width
			} else {
				if !nativeACByte(s, data[byteAt], p.nocase) {
					return false, nil
				}
				at += step
			}
			index += step
			if at < 0 || at > len(data) {
				return false, nil
			}
		}
		return next(at)
	}
	start := 0
	if reverse {
		start = len(symbols) - 1
	}
	return walk(symbols, start, pos, done)
}

// matchKey validates the forward body first, then its backward prefix, like
// the native key-driven matcher. One key event chooses one start/end pair;
// variable prefixes do not multiply counts by every possible full start.
func (p *nativeACPart) matchKey(data []byte, keyEnd int, w *matchWork) (int, int, bool, error) {
	keyStart := keyEnd - p.depth
	start, end := 0, 0
	if keyStart < 0 || keyEnd > len(data) {
		return 0, 0, false, nil
	}
	if p.literal != nil {
		start = keyStart - p.keyAt
		end = start + len(p.literal)
		if start < 0 || end > len(data) {
			return 0, 0, false, nil
		}
		if err := w.step(int64(len(p.literal))); err != nil {
			return 0, 0, false, err
		}
		if !equalBytes(data[start:end], p.literal, p.nocase) ||
			!p.classBoundary(data, start, p.leftClass, p.leftNeg, true) ||
			!p.classBoundary(data, end, p.rightClass, p.rightNeg, false) {
			return 0, 0, false, nil
		}
		ok, err := p.characterAnchor(data, start, p.left, true, w)
		if !ok || err != nil {
			return 0, 0, ok, err
		}
		ok, err = p.characterAnchor(data, end, p.right, false, w)
		return start, end, ok, err
	}
	for i := 0; i < p.depth; i++ {
		if err := w.step(1); err != nil {
			return 0, 0, false, err
		}
		if !nativeACByte(nativeACSymbol{value: p.key[i], mask: 255}, data[keyStart+i], p.nocase) {
			return 0, 0, false, nil
		}
	}
	ok, err := p.matchSequence(data, p.symbols[p.keyAt+p.depth:], keyEnd, false, w, func(right int) (bool, error) {
		if !p.classBoundary(data, right, p.rightClass, p.rightNeg, false) {
			return false, nil
		}
		ok, err := p.characterAnchor(data, right, p.right, false, w)
		if !ok || err != nil {
			return ok, err
		}
		return p.matchSequence(data, p.symbols[:p.keyAt], keyStart, true, w, func(left int) (bool, error) {
			if !p.classBoundary(data, left, p.leftClass, p.leftNeg, true) {
				return false, nil
			}
			ok, err := p.characterAnchor(data, left, p.left, true, w)
			if ok && err == nil {
				start, end = left, right
			}
			return ok, err
		})
	})
	return start, end, ok, err
}
