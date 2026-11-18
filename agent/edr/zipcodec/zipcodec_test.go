package zipcodec

import (
	"bytes"
	"errors"
	"testing"
)

// bitWriter writes a bit stream least significant bit first.
type bitWriter struct {
	out []byte
	acc uint64
	n   uint
}

func (w *bitWriter) put(v uint64, n uint) {
	w.acc |= v << w.n
	w.n += n
	for w.n >= 8 {
		w.out = append(w.out, byte(w.acc))
		w.acc >>= 8
		w.n -= 8
	}
}

// huff writes a Huffman code, which deflate stores most significant bit
// first.
func (w *bitWriter) huff(code uint64, n uint) {
	var r uint64
	for i := uint(0); i < n; i++ {
		r = r<<1 | code>>i&1
	}
	w.put(r, n)
}

func (w *bitWriter) bytes() []byte {
	if w.n > 0 {
		return append(w.out, byte(w.acc))
	}
	return w.out
}

// fixedLiteral writes a literal/length symbol with the fixed code.
func (w *bitWriter) fixedLiteral(sym int) {
	switch {
	case sym < 144:
		w.huff(uint64(0x30+sym), 8)
	case sym < 256:
		w.huff(uint64(0x190+sym-144), 9)
	case sym < 280:
		w.huff(uint64(sym-256), 7)
	default:
		w.huff(uint64(0xc0+sym-280), 8)
	}
}

// deflate64Sample is a stored block of 60000 bytes and a fixed block that
// copies with the Deflate64-only codes: length symbol 285 (3 plus 16 extra
// bits) and distance codes 30 and 31.
func deflate64Sample() ([]byte, []byte) {
	stored := make([]byte, 60000)
	for i := range stored {
		stored[i] = byte(i*7 + i/251)
	}
	w := &bitWriter{}
	w.put(0, 3) // stored, not last
	w.out, w.acc, w.n = append(w.out, byte(w.acc)), 0, 0 // to the byte boundary
	w.out = append(w.out, byte(len(stored)), byte(len(stored)>>8), ^byte(len(stored)), ^byte(len(stored)>>8))
	w.out = append(w.out, stored...)
	w.put(1|1<<1, 3) // fixed, last
	want := append([]byte(nil), stored...)
	// length 3+1000 from distance 33000 (code 30: 32769 + 231)
	w.fixedLiteral(285)
	w.put(1000, 16)
	w.huff(30, 5)
	w.put(231, 14)
	for i := 0; i < 1003; i++ {
		want = append(want, want[len(want)-33000])
	}
	// length 3 from distance 49153 + 5 (code 31)
	w.fixedLiteral(257)
	w.huff(31, 5)
	w.put(5, 14)
	for i := 0; i < 3; i++ {
		want = append(want, want[len(want)-49158])
	}
	w.fixedLiteral('Z')
	want = append(want, 'Z')
	w.fixedLiteral(256)
	return w.bytes(), want
}

func TestInflate64(t *testing.T) {
	stream, want := deflate64Sample()
	got, err := Inflate64(stream, 1<<20)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("%d bytes, %v; want %d", len(got), err, len(want))
	}
	// A cut stream gives what decoded; a limit stops it.
	got, err = Inflate64(stream[:len(stream)-3], 1<<20)
	if !errors.Is(err, ErrTruncated) || !bytes.Equal(got, want[:len(got)]) || len(got) < 60000 {
		t.Errorf("cut: %d bytes, %v", len(got), err)
	}
	if _, err := Inflate64(stream, 60500); !errors.Is(err, ErrLimit) {
		t.Errorf("limit: %v", err)
	}
	// Block type 3 and a distance before the start are data errors.
	if _, err := Inflate64([]byte{0x07}, 100); !errors.Is(err, ErrData) {
		t.Errorf("block type 3: %v", err)
	}
	w := &bitWriter{}
	w.put(3, 3)
	w.fixedLiteral('a')
	w.fixedLiteral(257)
	w.huff(4, 5) // distance 5
	w.put(0, 1)
	if got, err := Inflate64(w.bytes(), 100); !errors.Is(err, ErrData) || string(got) != "a" {
		t.Errorf("distance before the start: %q, %v", got, err)
	}
}

// implodeLiterals implodes data as literals only (no literal tree, 4 KiB
// window flags 0): two complete 6-bit trees, then a set bit and eight bits
// per byte.
func implodeLiterals(data []byte) []byte {
	tree := []byte{3, 0xf5, 0xf5, 0xf5, 0xf5}
	w := &bitWriter{}
	for _, b := range append(append([]byte(nil), tree...), tree...) {
		w.put(uint64(b), 8)
	}
	for _, b := range data {
		w.put(1, 1)
		w.put(uint64(b), 8)
	}
	return w.bytes()
}

func TestExplode(t *testing.T) {
	data := []byte("imploded literals, then the end of the data")
	got, err := Explode(implodeLiterals(data), 0, 1<<20)
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("%q %v", got, err)
	}
	// A match: distance 3 (low bits 2, high code 0), length 2+4 = 6.
	tree := []byte{3, 0xf5, 0xf5, 0xf5, 0xf5}
	w := &bitWriter{}
	for _, b := range append(append([]byte(nil), tree...), tree...) {
		w.put(uint64(b), 8)
	}
	for _, b := range []byte("abc") {
		w.put(1, 1)
		w.put(uint64(b), 8)
	}
	w.put(0, 1)
	w.put(2, 6)                   // low distance bits
	w.put(uint64(reverse6(0)), 6) // distance tree value 0
	w.put(uint64(reverse6(4)), 6) // length tree value 4
	got, err = Explode(w.bytes(), 0, 1<<20)
	if err != nil || string(got) != "abcabcabc" {
		t.Errorf("match: %q %v", got, err)
	}
	// Bad tree descriptions fail; the limit stops the output.
	if _, err := Explode([]byte{0, 0xff, 0}, 0, 100); !errors.Is(err, ErrData) {
		t.Errorf("short tree: %v", err)
	}
	if _, err := Explode(implodeLiterals(data), 0, 10); !errors.Is(err, ErrLimit) {
		t.Errorf("limit: %v", err)
	}
}

// reverse6 is the code a complete tree of 64 six-bit values gives value v,
// as read: the Shannon-Fano assignment gives the last value code 0, so
// value v has code 63-v, read bit-reversed.
func reverse6(v int) int {
	c, r := 63-v, 0
	for i := 0; i < 6; i++ {
		r = r<<1 | c>>i&1
	}
	return r
}

func FuzzInflate64(f *testing.F) {
	stream, _ := deflate64Sample()
	f.Add(stream)
	f.Fuzz(func(t *testing.T, b []byte) { Inflate64(b, 1<<20) })
}

func FuzzExplode(f *testing.F) {
	f.Add(implodeLiterals([]byte("seed data for the fuzzer")), uint16(0))
	f.Fuzz(func(t *testing.T, b []byte, flags uint16) { Explode(b, flags, 1<<20) })
}
