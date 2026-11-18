// Package msziptest supplies an independent RFC 1951 fixed-tree test encoder.
// It is not used by production scanning.
package msziptest

type writer struct {
	data  []byte
	bits  uint64
	count uint
}

func (w *writer) put(v uint32, n uint) {
	w.bits |= uint64(v) << w.count
	w.count += n
	for w.count >= 8 {
		w.data = append(w.data, byte(w.bits))
		w.bits >>= 8
		w.count -= 8
	}
}

func reverse(v uint32, n uint) uint32 {
	var r uint32
	for i := uint(0); i < n; i++ {
		r = r<<1 | v&1
		v >>= 1
	}
	return r
}

func (w *writer) symbol(s int) {
	var code uint32
	var n uint
	switch {
	case s < 144:
		code, n = uint32(0x30+s), 8
	case s < 256:
		code, n = uint32(0x190+s-144), 9
	case s < 280:
		code, n = uint32(s-256), 7
	default:
		code, n = uint32(0xc0+s-280), 8
	}
	w.put(reverse(code, n), n)
}

func (w *writer) match(length, distance int) {
	lb := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
	le := []uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
	db := []int{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577}
	de := []uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}
	i := len(lb) - 1
	for lb[i] > length {
		i--
	}
	w.symbol(257 + i)
	w.put(uint32(length-lb[i]), le[i])
	i = len(db) - 1
	for db[i] > distance {
		i--
	}
	w.put(reverse(uint32(i), 5), 5)
	w.put(uint32(distance-db[i]), de[i])
}

// Encode writes one final fixed-tree frame. History lists the previous
// dictionary bytes from oldest to newest; overlapping LZ77 copies are allowed.
func Encode(data, history []byte) []byte {
	if len(history) > 32768 {
		history = history[len(history)-32768:]
	}
	plain := append(append([]byte(nil), history...), data...)
	key := func(i int) uint32 { return uint32(plain[i])<<16 | uint32(plain[i+1])<<8 | uint32(plain[i+2]) }
	seen := make(map[uint32]int)
	for i := 0; i < len(history) && i+2 < len(plain); i++ {
		seen[key(i)] = i
	}
	w := writer{}
	w.put(3, 3) // BFINAL=1, BTYPE=01
	for at := len(history); at < len(plain); {
		n, from := 0, 0
		if at+2 < len(plain) {
			if prior, ok := seen[key(at)]; ok && at-prior <= 32768 {
				for n < 258 && at+n < len(plain) && plain[prior+n] == plain[at+n] {
					n++
				}
				from = prior
			}
		}
		if n >= 3 {
			w.match(n, at-from)
		} else {
			n = 1
			w.symbol(int(plain[at]))
		}
		for i := at; i < at+n && i+2 < len(plain); i++ {
			seen[key(i)] = i
		}
		at += n
	}
	w.symbol(256)
	if w.count != 0 {
		w.data = append(w.data, byte(w.bits))
	}
	return append([]byte("CK"), w.data...)
}

// InitialMatch emits a fixed-tree match against the zeroed initial window.
func InitialMatch(length, distance int) []byte {
	return InitialMatchWithSuffix(length, distance, nil)
}

// InitialMatchWithSuffix appends literal output after a zero-history match.
func InitialMatchWithSuffix(length, distance int, suffix []byte) []byte {
	w := writer{}
	w.put(3, 3)
	w.match(length, distance)
	for _, v := range suffix {
		w.symbol(int(v))
	}
	w.symbol(256)
	if w.count != 0 {
		w.data = append(w.data, byte(w.bits))
	}
	return append([]byte("CK"), w.data...)
}
