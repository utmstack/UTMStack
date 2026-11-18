// Package lzxtest writes LZX data for tests: literal-only verbatim blocks
// (every byte an 8-bit code), so the data does not appear in the clear, and
// uncompressed blocks.
package lzxtest

// writer writes 16-bit little-endian words, most significant bit first.
type writer struct {
	out  []byte
	acc  uint32
	bits uint
	n    int // bits written in all
}

func (w *writer) put(v uint32, n uint) {
	for i := int(n) - 1; i >= 0; i-- {
		w.acc = w.acc<<1 | v>>uint(i)&1
		w.bits++
		w.n++
		if w.bits == 16 {
			w.out = append(w.out, byte(w.acc), byte(w.acc>>8))
			w.acc, w.bits = 0, 0
		}
	}
}

// align pads to the next 16-bit boundary.
func (w *writer) align() {
	for w.bits != 0 {
		w.put(0, 1)
	}
}

var slots = map[int]int{15: 30, 16: 32, 17: 34, 18: 36, 19: 38, 20: 42, 21: 50}

// Literals encodes data as one verbatim block of literals: the main tree
// gives every byte an 8-bit code (the byte itself) and no matches, the
// length tree is empty. Each 32 KiB frame ends on a 16-bit boundary.
func Literals(data []byte, windowBits int) []byte {
	w := &writer{}
	w.put(0, 1) // no E8 translation
	w.put(1, 3) // verbatim
	w.put(uint32(len(data))>>8, 16)
	w.put(uint32(len(data))&0xff, 8)
	// Pretree: symbols 0 and 9 with one-bit codes (0 -> "0", 9 -> "1").
	pretree := func() {
		for i := 0; i < 20; i++ {
			l := uint32(0)
			if i == 0 || i == 9 {
				l = 1
			}
			w.put(l, 4)
		}
	}
	pretree()
	for i := 0; i < 256; i++ {
		w.put(1, 1) // symbol 9: length (0-9) mod 17 = 8
	}
	pretree()
	for i := 0; i < slots[windowBits]*8; i++ {
		w.put(0, 1) // symbol 0: length 0
	}
	pretree()
	for i := 0; i < 249; i++ {
		w.put(0, 1)
	}
	for i, b := range data {
		w.put(uint32(b), 8)
		if (i+1)%32768 == 0 {
			w.align()
		}
	}
	w.align()
	return w.out
}

// Uncompressed encodes data as one uncompressed block (repeated offsets 1).
func Uncompressed(data []byte) []byte {
	w := &writer{}
	w.put(0, 1)
	w.put(3, 3)
	w.put(uint32(len(data))>>8, 16)
	w.put(uint32(len(data))&0xff, 8)
	w.align()
	if w.n%16 == 0 && w.bits == 0 && len(w.out)%2 == 0 {
		// the header ended exactly on a boundary: a whole padding word
		if (1+3+24)%16 == 0 {
			w.put(0, 16)
		}
	}
	out := append(w.out, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0)
	out = append(out, data...)
	if len(data)&1 != 0 {
		out = append(out, 0)
	}
	return out
}
