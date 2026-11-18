package zipcodec

import "sort"

// Explode decodes imploded data (PKWARE APPNOTE 5.3): flags bit 1 selects
// the 8 KiB window (7 low distance bits instead of 6), bit 2 the literal tree
// (and a minimum match of 3 instead of 2). The trees come first, each a byte
// count less one and then bytes of (count-1)<<4 | (length-1); codes are
// assigned Shannon-Fano style from the longest length down and read
// bit-reversed. A match reaching before the output's start copies zeros (the
// window starts zeroed). As the reference does, the declared size is not
// used: decoding goes on until the data cannot hold another whole token.
func Explode(src []byte, flags uint16, limit int) ([]byte, error) {
	b := &bits{src: src}
	big, literals := flags&2 != 0, flags&4 != 0
	var lit *sfTree
	var err error
	if literals {
		if lit, err = readTree(b, 256); err != nil {
			return nil, err
		}
	}
	length, err := readTree(b, 64)
	if err != nil {
		return nil, err
	}
	dist, err := readTree(b, 64)
	if err != nil {
		return nil, err
	}
	low, minMatch := uint(6), 2
	if big {
		low = 7
	}
	if literals {
		minMatch = 3
	}
	out := make([]byte, 0, min(4*len(src), 1<<20))
	for {
		flag, ok := b.get(1)
		if !ok {
			return out, nil
		}
		if flag == 1 {
			var c int
			if literals {
				if c, ok = lit.decode(b); !ok {
					return out, nil
				}
				if c < 0 {
					return out, ErrData
				}
			} else {
				v, ok := b.get(8)
				if !ok {
					return out, nil
				}
				c = int(v)
			}
			if len(out) >= limit {
				return out, ErrLimit
			}
			out = append(out, byte(c))
			continue
		}
		lo, ok := b.get(low)
		if !ok {
			return out, nil
		}
		hi, ok := dist.decode(b)
		if !ok {
			return out, nil
		}
		if hi < 0 {
			return out, ErrData
		}
		distance := (hi<<low | int(lo)) + 1
		n, ok := length.decode(b)
		if !ok {
			return out, nil
		}
		if n < 0 {
			return out, ErrData
		}
		if n == 63 {
			extra, ok := b.get(8)
			if !ok {
				return out, nil
			}
			n += int(extra)
		}
		n += minMatch
		for i := 0; i < n; i++ {
			if len(out) >= limit {
				return out, ErrLimit
			}
			if distance > len(out) {
				out = append(out, 0)
			} else {
				out = append(out, out[len(out)-distance])
			}
		}
	}
}

// sfTree is a Shannon-Fano tree: codes (bit-reversed, as read) and their
// lengths per value.
type sfTree struct {
	code []uint16
	bits []uint8
	// lookup by length then code
	byLen [17]map[uint16]int
}

func readTree(b *bits, n int) (*sfTree, error) {
	count, ok := b.get(8)
	if !ok {
		return nil, ErrTruncated
	}
	lengths := make([]uint8, 0, n)
	for i := 0; i <= int(count); i++ {
		v, ok := b.get(8)
		if !ok {
			return nil, ErrTruncated
		}
		l, k := uint8(v&15)+1, int(v>>4)+1
		for ; k > 0; k-- {
			if len(lengths) == n {
				return nil, ErrData // more values than the tree has
			}
			lengths = append(lengths, l)
		}
	}
	if len(lengths) != n {
		return nil, ErrData
	}
	// Sort by length (stable), assign codes from the end: the code grows
	// by 1<<(16-length) of the current length before each value.
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return lengths[order[i]] < lengths[order[j]] })
	t := &sfTree{code: make([]uint16, n), bits: lengths}
	code, inc, last := uint32(0), uint32(0), uint8(0)
	for i := n - 1; i >= 0; i-- {
		v := order[i]
		code += inc
		if lengths[v] != last {
			last = lengths[v]
			inc = 1 << (16 - uint(last))
		}
		t.code[v] = reverse16(uint16(code)) & (1<<uint(lengths[v]) - 1)
	}
	for v := 0; v < n; v++ {
		l := lengths[v]
		if t.byLen[l] == nil {
			t.byLen[l] = map[uint16]int{}
		}
		t.byLen[l][t.code[v]] = v
	}
	return t, nil
}

func reverse16(v uint16) uint16 {
	var r uint16
	for i := 0; i < 16; i++ {
		r = r<<1 | v&1
		v >>= 1
	}
	return r
}

// decode reads a value: the code's bits arrive least significant first.
func (t *sfTree) decode(b *bits) (int, bool) {
	var code uint16
	for l := 1; l <= 16; l++ {
		v, ok := b.get(1)
		if !ok {
			return 0, false
		}
		code |= uint16(v) << uint(l-1)
		if m := t.byLen[l]; m != nil {
			if s, ok := m[code]; ok {
				return s, true
			}
		}
	}
	return -1, true // no code of up to 16 bits matched
}
