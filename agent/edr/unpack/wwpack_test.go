package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// wwWriter writes the WWPack32 stream: bits MSB first in 32-bit words, a new
// word reserved the moment the previous one fills (as the decoder loads it
// then), literal bytes in the same byte stream.
type wwWriter struct {
	out  []byte
	word uint32
	n    int
	at   int
}

func newWWWriter() *wwWriter { return &wwWriter{out: make([]byte, 4)} }

func (w *wwWriter) bit(b int) {
	w.word = w.word<<1 | uint32(b&1)
	w.n++
	if w.n == 32 {
		binary.LittleEndian.PutUint32(w.out[w.at:], w.word)
		w.at = len(w.out)
		w.out = append(w.out, 0, 0, 0, 0)
		w.word, w.n = 0, 0
	}
}

func (w *wwWriter) bits(v, n int) {
	for i := n - 1; i >= 0; i-- {
		w.bit(v >> i & 1)
	}
}

func (w *wwWriter) finish() []byte {
	if w.n > 0 {
		binary.LittleEndian.PutUint32(w.out[w.at:], w.word<<(32-w.n))
	}
	w.out = append(w.out, make([]byte, 8)...)
	for len(w.out)%4 != 0 {
		w.out = append(w.out, 0)
	}
	return w.out
}

// wwEncode encodes data with literals, three- and four-byte matches,
// variable-length matches (5-6, 7-13, 14-28 and longer) and two-byte
// matches, then the end marker when end is set.
func wwEncode(data []byte, end bool) []byte {
	w := newWWWriter()
	offset := func(off int) {
		for _, c := range []struct{ code, width, base int }{{0, 5, 1}, {1, 6, 0x21}, {2, 7, 0x61}} {
			if off >= c.base && off < c.base+1<<c.width {
				w.bits(c.code, 3)
				w.bits(off-c.base, c.width)
				return
			}
		}
		for _, c := range []struct{ code, extra, width, base int }{{3, 0, 8, 0xe1}, {3, 1, 9, 0x1e1}, {4, 0, 10, 0x3e1}, {4, 1, 11, 0x7e1}} {
			if off >= c.base && off < c.base+1<<c.width {
				w.bits(c.code, 3)
				w.bit(c.extra)
				w.bits(off-c.base, c.width)
				return
			}
		}
		w.bits(5, 3)
		w.bits(off-0xfe1, 12)
	}
	length := func(n int) {
		switch {
		case n <= 6:
			w.bit(0)
			w.bit(n - 5)
		case n <= 13:
			w.bit(1)
			w.bits(n-6, 3)
		case n <= 28:
			w.bit(1)
			w.bits(0, 3)
			w.bits(n-13, 4)
		default:
			w.bit(1)
			w.bits(0, 3)
			w.bits(0, 4)
			width, dx := 4, 0xd
			for width < 7 {
				dx = (dx+2)*2 - 1
				width++
				if lo := dx & 0xff; n >= lo && n < lo+1<<width {
					w.bit(1)
					w.bits(n-lo, width)
					return
				}
				w.bit(0)
			}
			w.bits(n, 14)
		}
	}
	for i := 0; i < len(data); {
		bestN, bestOff := 0, 0
		for off := 1; off <= min(i, 0xfe0); off++ {
			n := 0
			for i+n < len(data) && data[i+n] == data[i+n-off] && n < 300 {
				n++
			}
			if n > bestN {
				bestN, bestOff = n, off
			}
		}
		switch {
		case bestN >= 3:
			if bestN <= 4 {
				w.bit(1)
				w.bits(bestN-2, 2)
				offset(bestOff)
			} else {
				w.bit(1)
				w.bits(0, 2)
				offset(bestOff)
				length(bestN)
			}
			i += bestN
		case bestN == 2 && bestOff < 0x21:
			w.bit(1)
			w.bits(3, 2)
			w.bits(0, 2)
			w.bits(bestOff-1, 5)
			i += 2
		default:
			w.bit(0)
			w.out = append(w.out, data[i])
			i++
		}
	}
	if end {
		w.bit(1)
		w.bits(3, 2)
		w.bits(3, 2)
		w.bits(0x1ff, 9)
	}
	return w.finish()
}

// wwTestProgram is a program laid out as the upstream WWPack test program:
// sections .text, .rdata, .rsrc, .clam (0x1000 apart, raw data 0x200 bytes
// each) and the stub .WWP32 at the entry point, whose records at +0x2a1
// describe the given blocks (section index, plain data).
func wwTestProgram(blocks [][2]any, entry uint32) []byte {
	le := binary.LittleEndian
	f := make([]byte, 0x1000)
	copy(f, "MZ")
	le.PutUint32(f[0x3c:], 0x80)
	e := f[0x80:]
	copy(e, "PE\x00\x00")
	le.PutUint16(e[4:], 0x14c)
	le.PutUint16(e[6:], 5)
	le.PutUint16(e[20:], 0xe0)
	o := e[24:]
	le.PutUint16(o, 0x10b)
	le.PutUint32(o[16:], 0x5000)
	le.PutUint32(o[28:], 0x400000)
	le.PutUint32(o[32:], 0x1000)
	le.PutUint32(o[36:], 0x200)
	le.PutUint32(o[56:], 0x6000)
	le.PutUint32(o[60:], 0x400)
	names := []string{".text", ".rdata", ".rsrc", ".clam", ".WWP32"}
	for i, name := range names {
		s := o[0xe0+40*i:]
		copy(s, name)
		le.PutUint32(s[8:], 0x1000)
		le.PutUint32(s[12:], uint32(0x1000*(i+1)))
		le.PutUint32(s[16:], 0x200)
		le.PutUint32(s[20:], uint32(0x400+0x200*i))
		copy(f[0x400+0x200*i:], name+" original raw data")
	}
	le.PutUint32(o[0xe0+40*4+16:], 0x400)
	f = append(f, make([]byte, 0x200)...)
	stub := f[0xc00:]
	copy(stub, []byte{0x53, 0x55, 0x8b, 0xe8, 0x33, 0xdb, 0xeb})
	stub[0x294] = 0xe9
	le.PutUint32(stub[0x295:], entry-(0x5000+0x299))
	at := 0x2a1
	for k, b := range blocks {
		index, plain := b[0].(int), b[1].([]byte)
		stream := wwEncode(plain, true)
		if len(stream) > 0x200 {
			panic("test stream does not fit its section")
		}
		copy(f[0x400+0x200*index:0x400+0x200*(index+1)], append(stream, make([]byte, 0x200)...))
		count := len(stream) / 4
		le.PutUint32(stub[at:], uint32(0x5000-0x1000*(index+1)))
		le.PutUint32(stub[at+4:], 0xffc)
		le.PutUint32(stub[at+8:], uint32(count))
		le.PutUint32(stub[at+12:], uint32((count-1)*4))
		if k+1 < len(blocks) {
			stub[at+16] = 1
		}
		at += 0x11
	}
	return f
}

func wwPlain(seed byte, n int) []byte {
	out := []byte{}
	for i := 0; len(out) < n; i++ {
		out = append(out, []byte("block text ")...)
		out = append(out, seed, byte(i%4), 0, 0, 0, 0, 0)
		if i%7 == 0 {
			out = append(out, bytes.Repeat([]byte{seed}, 40)...)
		}
	}
	return out[:n]
}

// The reference's WWPack unpacker (parity/unpackers.md): the blocks are
// decoded into their sections, which become their aligned size; the stub's
// section goes; the entry point is the stub's final jump.
func TestWWPackAsReference(t *testing.T) {
	le := binary.LittleEndian
	a, b := wwPlain('a', 0x1000), wwPlain('b', 0x700)
	f := wwTestProgram([][2]any{{0, b}, {3, a}}, 0x1234)
	got, out := WWPack(f)
	if got != WWPackUnpacked {
		t.Fatalf("outcome %d", got)
	}
	e := le.Uint32(out[0x3c:])
	o := out[e+24:]
	if le.Uint16(out[e+6:]) != 4 || le.Uint32(o[16:]) != 0x1234 || le.Uint32(o[56:]) != 0x5000 || len(out) != 0x5000 {
		t.Fatalf("header: %d sections, entry %#x, image %#x, %d bytes", le.Uint16(out[e+6:]), le.Uint32(o[16:]), le.Uint32(o[56:]), len(out))
	}
	if !bytes.Equal(out[0x4000:0x5000], a) || !bytes.Equal(out[0x1000:0x1700], b) || !bytes.Equal(out[0x1700:0x2000], make([]byte, 0x900)) {
		t.Fatal("decoded sections")
	}
	if !bytes.HasPrefix(out[0x2000:], []byte(".rdata original raw data")) {
		t.Fatal("a section that is not a block keeps its raw data")
	}
	for i := 0; i < 4; i++ {
		s := o[0xe0+40*i:]
		if le.Uint32(s[8:]) != 0x1000 || le.Uint32(s[16:]) != 0x1000 || le.Uint32(s[20:]) != le.Uint32(s[12:]) {
			t.Fatalf("section %d", i)
		}
	}
	if !bytes.Equal(o[0xe0+40*4:0xe0+40*5], make([]byte, 40)) {
		t.Fatal("the stub's section entry is cleared")
	}
	// Output past the section fails; a stream without its end marker that
	// runs out of data is not decided.
	if got, _ := WWPack(wwTestProgram([][2]any{{3, wwPlain('c', 0x1200)}}, 0x1020)); got != WWPackFailed {
		t.Errorf("overflow: outcome %d", got)
	}
	f = wwTestProgram([][2]any{{3, a}}, 0x1020)
	stream := wwEncode(a[:0x80], false)
	copy(f[0xa00:0xc00], append(stream, make([]byte, 0x200)...))
	binary.LittleEndian.PutUint32(f[0xc00+0x2a1+8:], uint32(len(stream)/4))
	binary.LittleEndian.PutUint32(f[0xc00+0x2a1+12:], uint32((len(stream)/4-1)*4))
	if got, _ := WWPack(f); got != WWPackUnknown {
		t.Errorf("no end marker: outcome %d", got)
	}
	// A record whose area is not a section start is not decided.
	f = wwTestProgram([][2]any{{3, a}}, 0x1020)
	binary.LittleEndian.PutUint32(f[0xc00+0x2a1:], 0x1100)
	if got, _ := WWPack(f); got != WWPackUnknown {
		t.Errorf("area: outcome %d", got)
	}
}
