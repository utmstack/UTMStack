package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// aspWriter writes bits MSB first.
type aspWriter struct {
	out []byte
	n   uint
}

func (w *aspWriter) bits(v uint32, n uint) {
	for i := int(n) - 1; i >= 0; i-- {
		if w.n%8 == 0 {
			w.out = append(w.out, 0)
		}
		if v>>uint(i)&1 != 0 {
			w.out[len(w.out)-1] |= 0x80 >> (w.n % 8)
		}
		w.n++
	}
}

// aspCodes gives each symbol its canonical code as the decoder assigns it.
func aspCodes(lengths []byte) (codes []uint32) {
	var t aspTable
	if !t.build(lengths) {
		panic("incomplete test table")
	}
	codes = make([]uint32, len(lengths))
	next := t.base
	for i, l := range lengths {
		if l != 0 {
			codes[i] = t.limit[l-1]>>(24-l) + next[l] - t.base[l]
			next[l]++
		}
	}
	return codes
}

// aspTestLengths are complete codes for the four tables: literals and the
// first 47 match symbols in 9 bits, the rest in 10; long lengths in 4 and 5
// bits; aligned offsets all 3 bits (unused); the pretree in 4 and 5 bits.
func aspTestLengths() (pre [19]byte, all [aspTableSize]byte) {
	for i := range pre {
		pre[i] = 4
		if i >= 13 {
			pre[i] = 5
		}
	}
	for i := 0; i < aspMainSize; i++ {
		all[i] = 10
		if i < 0x100+47 {
			all[i] = 9
		}
	}
	for i := 0; i < 28; i++ {
		all[aspMainSize+i] = 5
		if i < 4 {
			all[aspMainSize+i] = 4
		}
	}
	for i := 0x2ed; i < aspTableSize; i++ {
		all[i] = 3
	}
	return
}

// aspEncode encodes data with literals and matches (offsets from scratch,
// lengths up to 8 through the main table only).
func aspEncode(data []byte) []byte {
	w := &aspWriter{}
	pre, all := aspTestLengths()
	w.bits(0, 1)
	for _, l := range pre {
		w.bits(uint32(l), 4)
	}
	preCode := aspCodes(pre[:])
	for _, l := range all {
		w.bits(preCode[l], uint(pre[l]))
	}
	mainCode := aspCodes(all[:aspMainSize])
	sym := func(s int) { w.bits(mainCode[s], uint(all[s])) }
	for i := 0; i < len(data); {
		best, off := 0, 0
		for o := 1; o <= min(i, 0x3fff); o++ {
			n := 0
			for i+n < len(data) && data[i+n] == data[i+n-o] && n < 8 {
				n++
			}
			if n > best {
				best, off = n, o
			}
		}
		if best < 3 {
			sym(int(data[i]))
			i++
			continue
		}
		pos := uint32(off + 2)
		slot := 57
		for slot > 0 && aspPosBase[slot] > pos {
			slot--
		}
		sym(0x100 + slot*8 + best - 2)
		w.bits(pos-aspPosBase[slot], aspPosExtra[slot])
		i += best
	}
	return append(w.out, 0, 0, 0, 0, 0, 0)
}

// The ASPack 2.12 decoder against data encoded with the format's own rules.
func TestASPackDecode(t *testing.T) {
	data := bytes.Repeat([]byte("ASPack test block, "), 40)
	data = append(data, []byte("with a tail of other text 0123456789")...)
	out, r := aspDecode(aspEncode(data), uint32(len(data)))
	if r != aspDecoded || !bytes.Equal(out, data) {
		t.Fatalf("result %d, %d bytes", r, len(out))
	}
	// Input that ends early is not decided; an incomplete pretree fails.
	enc := aspEncode(data)
	if _, r := aspDecode(enc[:len(enc)/2], uint32(len(data))); r != aspUndecided {
		t.Errorf("short input: %d", r)
	}
	bad := append([]byte{}, enc...)
	bad[0] &= 0x80 // the first pretree lengths become 0
	bad[1] = 0
	if _, r := aspDecode(bad, uint32(len(data))); r != aspBadTables {
		t.Errorf("bad tables: %d", r)
	}
}

// aspTestProgram lays out an ASPack 2.12 program: .text (0x1000, stored),
// .data (0x2000, one compressed block), the stub (0x3000) and .adata
// (0x5000, no raw data); the entry point is stub+1.
func aspTestProgram(plain []byte, entry uint32) ASPackProgram {
	le := binary.LittleEndian
	file := make([]byte, 0x400)
	text := bytes.Repeat([]byte("stored text "), 0x200/12+1)[:0x200]
	file = append(file, text...)
	enc := aspEncode(plain)
	data := append(enc, make([]byte, 0x400-len(enc)%0x400)...)
	file = append(file, data...)
	stub := make([]byte, 0x800)
	copy(stub[1:], []byte{0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb})
	copy(stub[0x3ba:], aspVersion212)
	copy(stub[0x6d6:], aspTables)
	stub[0x39a] = 0xb8
	le.PutUint32(stub[0x39b:], entry)
	le.PutUint32(stub[0x57c:], 0x2000)
	le.PutUint32(stub[0x580:], uint32(len(plain)))
	stubRaw := uint32(len(file))
	file = append(file, stub...)
	return ASPackProgram{File: file, ImageBase: 0x400000, Entry: 0x3001, ImageSize: 0x6000, Sections: []ASPackSection{
		{RVA: 0x1000, VSize: 0x1000, Raw: 0x400, RawSize: 0x200},
		{RVA: 0x2000, VSize: 0x1000, Raw: 0x600, RawSize: uint32(len(data))},
		{RVA: 0x3000, VSize: 0x2000, Raw: stubRaw, RawSize: 0x800},
		{RVA: 0x5000, VSize: 0x1000, Raw: uint32(len(file))},
	}}
}

// The reference's ASPack unpacker (parity/unpackers.md): blocks decoded in
// place, the first one through the call-address filter, the stub's two
// sections dropped, each kept section written whole.
func TestASPackAsReference(t *testing.T) {
	le := binary.LittleEndian
	plain := bytes.Repeat([]byte("ASPack block "), 60)
	copy(plain[0x10:], []byte{0xe8, 0, 0x11, 0x22, 0x33})
	got, out := ASPack(aspTestProgram(plain, 0x1020))
	if got != ASPackUnpacked {
		t.Fatalf("outcome %d", got)
	}
	e := le.Uint32(out[0x3c:])
	o := out[e+24:]
	if le.Uint16(out[e+6:]) != 2 || le.Uint32(o[16:]) != 0x1020 || le.Uint32(o[56:]) != 0x3000 {
		t.Fatalf("header: %d sections, entry %#x, image %#x", le.Uint16(out[e+6:]), le.Uint32(o[16:]), le.Uint32(o[56:]))
	}
	data := out[0x400+0x1000:]
	want := append([]byte{}, plain...)
	copy(want[0x11:], []byte{0x01, 0x22, 0x33, 0x00}) // rotated right by 8, less 0x10
	if !bytes.Equal(data[:len(want)], want) || !bytes.HasPrefix(out[0x400:], []byte("stored text")) {
		t.Fatal("sections")
	}
	// Another version goes on; a block past the image or of size 0 fails;
	// changed decoder tables are not decided.
	p := aspTestProgram(plain, 0x1020)
	p.File[p.Sections[2].Raw+0x3bb] = 1
	if got, _ := ASPack(p); got != ASPackGoesOn {
		t.Errorf("version: outcome %d", got)
	}
	for _, size := range []uint32{0, 0x5000} {
		p = aspTestProgram(plain, 0x1020)
		le.PutUint32(p.File[p.Sections[2].Raw+0x580:], size)
		if got, _ := ASPack(p); got != ASPackGoesOn {
			t.Errorf("block size %#x: outcome %d", size, got)
		}
	}
	p = aspTestProgram(plain, 0x1020)
	p.File[p.Sections[2].Raw+0x6d8]++
	if got, _ := ASPack(p); got != ASPackUnknown {
		t.Errorf("tables: outcome %d", got)
	}
}
