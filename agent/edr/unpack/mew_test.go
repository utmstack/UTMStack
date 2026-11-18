package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// mewBits writes the stream format the MEW stub decodes: bits MSB first in
// tag bytes placed where the first bit of each group of eight is needed.
type mewBits struct {
	out []byte
	tag int
	n   int
}

func (w *mewBits) bit(b int) {
	if w.n == 0 || w.n == 8 {
		w.tag = len(w.out)
		w.out = append(w.out, 0)
		w.n = 0
	}
	if b != 0 {
		w.out[w.tag] |= 0x80 >> w.n
	}
	w.n++
}

func (w *mewBits) gamma(v int) {
	top := 30
	for v>>top&1 == 0 {
		top--
	}
	for i := top - 1; i >= 0; i-- {
		w.bit(v >> i & 1)
		if i > 0 {
			w.bit(1)
		} else {
			w.bit(0)
		}
	}
}

// mewEncode encodes one block with every token kind: literals, a zero byte
// and one-byte copies (111), short matches (110), the last offset again
// right after a literal, long matches, then the end marker.
func mewEncode(data []byte) []byte {
	w := &mewBits{}
	w.out = append(w.out, data[0])
	i, afterLiteral, last := 1, true, 0
	match := func(off int) int {
		n := 0
		for i+n < len(data) && data[i+n] == data[i+n-off] && n < 255 {
			n++
		}
		return n
	}
	for i < len(data) {
		if afterLiteral && last > 0 && i >= last {
			if n := match(last); n >= 2 {
				w.bit(1)
				w.bit(0)
				w.gamma(2)
				w.gamma(n)
				i, afterLiteral = i+n, false
				continue
			}
		}
		bestN, bestOff := 0, 0
		for off := 1; off <= min(i, 0x7cff); off++ {
			if n := match(off); n > bestN {
				bestN, bestOff = n, off
			}
		}
		if bestN >= 2 && bestN <= 3 && bestOff <= 127 {
			w.bit(1)
			w.bit(1)
			w.bit(0)
			w.out = append(w.out, byte(bestOff<<1|(bestN-2)))
			i, afterLiteral, last = i+bestN, false, bestOff
			continue
		}
		extra := 0
		switch {
		case bestOff >= 0x7d00:
			extra = 2
		case bestOff >= 0x500:
			extra = 1
		case bestOff <= 0x7f:
			extra = 2
		}
		if bestN >= 4 && bestN-extra >= 2 {
			g := bestOff>>8 + 2
			if afterLiteral {
				g++
			}
			w.bit(1)
			w.bit(0)
			w.gamma(g)
			w.out = append(w.out, byte(bestOff))
			w.gamma(bestN - extra)
			i, afterLiteral, last = i+bestN, false, bestOff
			continue
		}
		short := 0
		for off := 1; off <= min(i, 15); off++ {
			if data[i-off] == data[i] {
				short = off
				break
			}
		}
		switch {
		case data[i] == 0 || short > 0:
			w.bit(1)
			w.bit(1)
			w.bit(1)
			for k := 3; k >= 0; k-- {
				w.bit(short >> k & 1)
			}
		default:
			w.bit(0)
			w.out = append(w.out, data[i])
		}
		i, afterLiteral = i+1, true
	}
	w.bit(1)
	w.bit(1)
	w.bit(0)
	return append(w.out, 0)
}

type mewTestBlock struct {
	rva  uint32
	data []byte
}

// mewTestProgram lays out a MEW program as the upstream sample is laid out:
// the stub at 0x154 in the header, the bit reader at 0x12c with its two
// jump targets, an empty section at 0x1000 of 0x5000 bytes, and a second
// section at 0x6000 holding the gamma reader, the table at 0x1c and the
// blocks.
func mewTestProgram(blocks []mewTestBlock) MEWProgram {
	const base = 0x400000
	f := make([]byte, 0x200+0x800)
	copy(f[0x108:], []byte{0x46, 0x12, 0xd2, 0xc3})
	copy(f[0x12c:], []byte{0x02, 0xd2, 0x75, 0xdb, 0x8a, 0x16, 0xeb, 0xd4})
	copy(f[0x154:], mewStub)
	binary.LittleEndian.PutUint32(f[0x155:], base+0x601c)
	binary.LittleEndian.PutUint32(f[0x154+0x32:], 0x6000-(0x154+0x36))
	s := f[0x200:]
	copy(s, mewGamma)
	binary.LittleEndian.PutUint32(s[0x18:], base+0x6000)
	binary.LittleEndian.PutUint32(s[0x1c:], base+0x12c)
	binary.LittleEndian.PutUint32(s[0x20:], base+0x1020)
	binary.LittleEndian.PutUint32(s[0x24:], base+blocks[0].rva)
	at := 0x28
	for i, b := range blocks {
		at += copy(s[at:], mewEncode(b.data))
		next := uint32(0)
		if i+1 < len(blocks) {
			next = base + blocks[i+1].rva
		}
		binary.LittleEndian.PutUint32(s[at:], next)
		at += 4
	}
	return MEWProgram{File: f, ImageBase: base, Target: 0x154, EmptyRVA: 0x1000, EmptyVSize: 0x5000,
		NextRVA: 0x6000, NextRaw: 0x200, NextSize: 0x800, NextVSize: 0x1000}
}

func repeated(s string, n int) []byte { return bytes.Repeat([]byte(s), n/len(s)+1)[:n] }

// mewSections reads the rebuilt program's sections: address, size and data.
func mewSections(t *testing.T, out []byte) []RebuildSection {
	t.Helper()
	le := binary.LittleEndian
	n := int(le.Uint16(out[0xd6:]))
	var secs []RebuildSection
	for i := 0; i < n; i++ {
		e := out[0xd0+24+0xe0+40*i:]
		raw, size := le.Uint32(e[20:]), le.Uint32(e[16:])
		secs = append(secs, RebuildSection{RVA: le.Uint32(e[12:]), VSize: le.Uint32(e[8:]), RawSize: size, Data: out[raw : raw+size]})
	}
	return secs
}

// The reference's MEW unpacker on programs laid out as its test program
// (parity/unpackers.md): each block makes a section from where the previous
// one ended to the block's end rounded up to 0x1000.
func TestMEWSectionsAsReference(t *testing.T) {
	mixed := append(append([]byte{}, repeated("\x00\x01abcabc\x02\x00xyzxyz", 0x1000)...), repeated("long run of text ", 0x2000)...)
	cases := []struct {
		name   string
		blocks []mewTestBlock
		want   [][2]uint32 // section address and size
	}{
		{"one page", []mewTestBlock{{0x1000, repeated("A0", 0x400)}}, [][2]uint32{{0x1000, 0x1000}}},
		{"three pages", []mewTestBlock{{0x1000, mixed[:0x2800]}}, [][2]uint32{{0x1000, 0x3000}}},
		{"two blocks in a page", []mewTestBlock{{0x1000, repeated("C0", 0x100)}, {0x1800, repeated("C1", 0x100)}}, [][2]uint32{{0x1000, 0x1000}, {0x2000, 0}}},
		{"pages between", []mewTestBlock{{0x1000, repeated("D0", 0x100)}, {0x4000, repeated("D1", 0x100)}}, [][2]uint32{{0x1000, 0x1000}, {0x2000, 0x3000}}},
		{"last page", []mewTestBlock{{0x5f00, repeated("F0", 0x100)}}, [][2]uint32{{0x1000, 0x5000}}},
		{"overlapping", []mewTestBlock{{0x1000, repeated("I0", 0x100)}, {0x1080, repeated("I1", 0x100)}}, [][2]uint32{{0x1000, 0x1000}, {0x2000, 0}}},
	}
	for _, c := range cases {
		got, out := MEW(mewTestProgram(c.blocks))
		if got != MEWUnpacked {
			t.Errorf("%s: outcome %d", c.name, got)
			continue
		}
		secs := mewSections(t, out)
		if len(secs) != len(c.want) {
			t.Errorf("%s: %d sections, want %d", c.name, len(secs), len(c.want))
			continue
		}
		for i, s := range secs {
			if s.RVA != c.want[i][0] || s.VSize != c.want[i][1] || s.RawSize != c.want[i][1] {
				t.Errorf("%s: section %d at %#x of %#x, want %#x", c.name, i, s.RVA, s.VSize, c.want[i])
			}
		}
		// Each block's data sits at its address, unless a later block wrote over it.
		image := map[uint32]byte{}
		for _, s := range secs {
			for k, b := range s.Data {
				image[s.RVA+uint32(k)] = b
			}
		}
		last := c.blocks[len(c.blocks)-1]
		for k, b := range last.data {
			if image[last.rva+uint32(k)] != b {
				t.Errorf("%s: byte %#x of the last block differs", c.name, k)
				break
			}
		}
		if le := binary.LittleEndian; le.Uint32(out[0xd0+24+16:]) != 0x1020 || le.Uint32(out[0xd0+24+28:]) != 0x400000 {
			t.Errorf("%s: entry or image base", c.name)
		}
	}
}

// Writing outside the empty section's virtual size, or a block that ends
// before the previous section does, fails, and the parser returns.
func TestMEWFailuresAsReference(t *testing.T) {
	for name, blocks := range map[string][]mewTestBlock{
		"past the area":   {{0x5f00, repeated("G0", 0x200)}},
		"before the area": {{0xf00, repeated("L0", 0x200)}},
		"far":             {{0x100000, repeated("M0", 0x100)}},
		"backwards":       {{0x3000, repeated("H0", 0x100)}, {0x1000, repeated("H1", 0x100)}},
	} {
		if got, _ := MEW(mewTestProgram(blocks)); got != MEWFailed {
			t.Errorf("%s: outcome %d", name, got)
		}
	}
}

// Programs that differ from what was checked against the reference are not
// decided.
func TestMEWUnknown(t *testing.T) {
	blocks := []mewTestBlock{{0x1000, repeated("A0", 0x400)}}
	for name, change := range map[string]func(p *MEWProgram){
		"other stub":        func(p *MEWProgram) { p.File[0x154+10] ^= 0xff },
		"other bit reader":  func(p *MEWProgram) { p.File[0x12c] = 0x90 },
		"other gamma":       func(p *MEWProgram) { p.File[0x200+3] ^= 0xff },
		"table outside":     func(p *MEWProgram) { binary.LittleEndian.PutUint32(p.File[0x155:], 0x500000) },
		"stream cut short":  func(p *MEWProgram) { p.NextSize = 0x30 },
		"other jump target": func(p *MEWProgram) { p.Target = 0x160 },
	} {
		p := mewTestProgram(blocks)
		change(&p)
		if got, _ := MEW(p); got != MEWUnknown {
			t.Errorf("%s: outcome %d", name, got)
		}
	}
}
