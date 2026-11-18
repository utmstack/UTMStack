package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// ptWriter writes the PEtite stream: bits MSB first in tag bytes reserved in
// the byte stream when the previous one fills, literal bytes in between.
type ptWriter struct {
	out []byte
	tag int
	n   int
}

func (w *ptWriter) bit(b int) {
	if w.n == 8 {
		w.tag, w.n = len(w.out), 0
		w.out = append(w.out, 0)
	}
	if b != 0 {
		w.out[w.tag] |= 0x80 >> w.n
	}
	w.n++
}

func (w *ptWriter) gamma(v int) {
	top := 31
	for v>>top == 0 {
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

// ptEncode encodes data greedily with literals, new-offset matches and
// repeated-offset matches.
func ptEncode(data []byte) []byte {
	size := len(data)
	t1, t2, low := uint32(0xfffffc60), uint32(0xffffc060), 5
	switch {
	case size >= 0x40000:
		t1, t2, low = 0xfffffb00, 0xffff8300, 8
	case size >= 0x10000:
		t1, t2, low = 0xfffff980, 0xffff8180, 7
	}
	bonus := func(off int) int {
		v := uint32(-off)
		b := 1
		if v < t1 {
			b++
		}
		if v < t2 {
			b++
		}
		return b
	}
	w := &ptWriter{out: []byte{data[0]}, n: 8}
	left, last := size-1, 0
	for i := 1; i < size; {
		best, off := 0, 0
		match := func(o int) int {
			n := 0
			for i+n < size && data[i+n] == data[i+n-o] && n < 255 {
				n++
			}
			return n
		}
		for o := 1; o <= min(i, 0x7fff); o++ {
			if n := match(o); n > best {
				best, off = n, o
			}
		}
		if last != 0 && last <= i {
			if n := match(last); n >= 2 && n+1 >= best {
				best, off = n, last
			}
		}
		length := 0
		switch {
		case best >= 2 && off == last:
			length = best
			w.bit(1)
			w.gamma(2)
		case best >= 2 && best-bonus(off) >= 1:
			length = best - bonus(off)
			v := off - 1
			w.bit(1)
			w.gamma(v>>low + 3)
			for k := low - 1; k >= 0; k-- {
				w.bit(v >> k & 1)
			}
			last = off
		}
		if length == 0 {
			w.bit(0)
			w.out = append(w.out, data[i]^byte(left))
			left--
			i++
			continue
		}
		if length <= 3 {
			w.bit(length >> 1 & 1)
			w.bit(length & 1)
		} else {
			w.bit(0)
			w.bit(0)
			w.gamma(length - 2)
		}
		i += best
		left -= best
	}
	return w.out
}

func ptPlain(seed byte, n int) []byte {
	out := []byte{}
	for i := 0; len(out) < n; i++ {
		out = append(out, []byte("petite block ")...)
		out = append(out, seed, byte(i*7), byte(i%5), 0, 0)
		if i%6 == 0 {
			out = append(out, bytes.Repeat([]byte{seed, 0x90}, 30)...)
		}
	}
	return out[:n]
}

// ptCode is a block ending with the PEtite code as the reference reads it:
// the entry point mark and encrypted entry point, the strip mark, and the
// import list's address.
func ptCode(head []byte, encrypted, imports uint32) []byte {
	le := binary.LittleEndian
	code := make([]byte, petiteCodeSize)
	copy(code[7:], petiteEPMark)
	le.PutUint32(code[0xf:], encrypted^petiteEPKey)
	copy(code[0x54:], petiteStripMark)
	le.PutUint32(code[0x121:], imports)
	return append(append([]byte{}, head...), code...)
}

type ptBlock struct {
	plain []byte
	dst   uint32
}

// ptProgram lays out a PEtite program like the upstream test program:
// sections at 0x1000 (0x2000 long), 0x3000 (resources), 0x4000 (empty) and
// the stub at 0x5000, whose table at +0x1b8 lists the blocks; the streams
// follow in the stub's section.
func ptProgram(blocks []ptBlock) PetiteProgram {
	le := binary.LittleEndian
	stub := make([]byte, 0x1000)
	at, streams := petiteTable, 0x400
	for _, b := range blocks {
		enc := ptEncode(b.plain)
		copy(stub[streams:], enc)
		le.PutUint32(stub[at:], uint32(0x5000+streams))
		le.PutUint32(stub[at+4:], uint32(len(b.plain)))
		le.PutUint32(stub[at+8:], b.dst)
		at += 16
		streams += len(enc)
	}
	file := make([]byte, 0x400)
	sections := []PetiteSection{
		{RVA: 0x1000, VSize: 0x2000},
		{RVA: 0x3000, VSize: 0x1000, Raw: 0x400, RawSize: 0x200},
		{RVA: 0x4000, VSize: 0x1000},
		{RVA: 0x5000, VSize: 0x1000, Raw: 0x600, RawSize: 0x1000},
	}
	file = append(file, bytes.Repeat([]byte("resource"), 0x40)...)
	file = append(file, stub...)
	return PetiteProgram{File: file, ImageBase: 0x400000, Entry: 0x5042, ImageSize: 0x6000,
		ResourceRVA: 0x3000, ResourceSize: 0xb0, Sections: sections}
}

// ptImports is the head of the block holding the code: two import lists
// (four names, one name) at its start and the list of lists at +0x5c.
func ptImports() []byte {
	le := binary.LittleEndian
	head := ptPlain('i', 0x123)
	for i, t := range []uint32{0x20b6, 0x20c4, 0x20d8, 0x20ea, 0, 0x210a, 0} {
		le.PutUint32(head[4*i:], t)
	}
	le.PutUint32(head[0x5c:], 0x2000)
	le.PutUint32(head[0x60:], 0x2014)
	le.PutUint32(head[0x64:], 0)
	return head
}

func ptSections(t *testing.T, out []byte) (entry uint32, secs [][3]uint32, data [][]byte) {
	t.Helper()
	le := binary.LittleEndian
	e := le.Uint32(out[0x3c:])
	o := out[e+24:]
	n := int(le.Uint16(out[e+6:]))
	for i := 0; i < n; i++ {
		s := o[0xe0+40*i:]
		raw, size := le.Uint32(s[20:]), le.Uint32(s[16:])
		secs = append(secs, [3]uint32{le.Uint32(s[12:]), le.Uint32(s[8:]), size})
		data = append(data, out[raw:raw+size])
	}
	return le.Uint32(o[16:]), secs, data
}

// The reference's PEtite unpacker (parity/unpackers.md): one section per
// block, the first block in a section taking its address, the PEtite code
// stripped, the entry point decrypted with the import list.
func TestPetiteAsReference(t *testing.T) {
	a, b := ptPlain('a', 0xffc), ptPlain('b', 0xf7)
	c := ptCode(ptImports(), 0xdfed1249, 0x205c)
	got, out := Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {c, 0x2000}}))
	if got != PetiteUnpacked {
		t.Fatalf("outcome %d", got)
	}
	entry, secs, data := ptSections(t, out)
	want := [][3]uint32{{0x1000, 0x1000, 0xf7}, {0x2000, 0x2000, 0x123}, {0x4000, 0xffc, 0xffc}}
	if entry != 0x1020 || len(secs) != 3 {
		t.Fatalf("entry %#x, sections %x", entry, secs)
	}
	for i := range want {
		if secs[i] != want[i] {
			t.Fatalf("section %d: %x, want %x", i, secs[i], want[i])
		}
	}
	if !bytes.Equal(data[0], b) || !bytes.Equal(data[1], c[:0x123]) || !bytes.Equal(data[2], a) {
		t.Fatal("section data")
	}
	// A later block in a section keeps its own address; the code in the
	// first block is found there too.
	_, out = Petite(ptProgram([]ptBlock{{c, 0x2000}, {b, 0x1000}, {a, 0x4000}}))
	if _, secs, _ = ptSections(t, out); secs[0] != [3]uint32{0x1000, 0, 0x1123} || secs[1] != [3]uint32{0x1000, 0x3000, 0xf7} {
		t.Fatalf("sections %x", secs)
	}
	// Two names fewer: a different entry point.
	head := ptImports()
	binary.LittleEndian.PutUint32(head[0x60:], 0)
	if _, out = Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {ptCode(head, 0xdfed1249, 0x205c), 0x2000}})); func() uint32 {
		e, _, _ := ptSections(t, out)
		return e
	}() != 0xfffe4f17 {
		t.Error("four names")
	}
	// No entry point mark: entry 0; no strip mark: nothing stripped.
	broken := append([]byte{}, c...)
	broken[0x123+7] ^= 0xff
	if _, out = Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {broken, 0x2000}})); func() bool {
		e, s, _ := ptSections(t, out)
		return e != 0 || s[1][2] != 0x123
	}() {
		t.Error("entry point mark")
	}
	broken = append([]byte{}, c...)
	broken[0x123+0x54] ^= 0xff
	if _, out = Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {broken, 0x2000}})); func() bool {
		e, s, _ := ptSections(t, out)
		return e != 0 || s[1][2] != 0x478
	}() {
		t.Error("strip mark")
	}
	// An import list outside the image: the bogus entry point 0x1000.
	if _, out = Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {ptCode(ptImports(), 0xdfed1249, 0xdf5c), 0x2000}})); func() uint32 {
		e, _, _ := ptSections(t, out)
		return e
	}() != 0x1000 {
		t.Error("bogus entry point")
	}
	// A block outside every section, or past its section's end, fails.
	if got, _ := Petite(ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {c, 0x2c00}})); got != PetiteFailed {
		t.Errorf("across a section: outcome %d", got)
	}
	p := ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {c, 0x2000}})
	p.Sections[2].RVA = 0x9000
	if got, _ := Petite(p); got != PetiteFailed {
		t.Errorf("outside: outcome %d", got)
	}
	// A record of size 0 is not decided.
	p = ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {c, 0x2000}})
	binary.LittleEndian.PutUint32(p.File[0x600+petiteTable+4:], 0)
	if got, _ := Petite(p); got != PetiteUnknown {
		t.Errorf("size 0: outcome %d", got)
	}
}

// The entry point rule against the reference's values for import lists of
// n names (n = 1..12, 300) and of ordinals.
func TestPetiteEntry(t *testing.T) {
	m := &petiteMemory{buf: make([]byte, 0x10000), low: 0x1000, limit: 0x10000}
	le := binary.LittleEndian
	list := func(thunks ...uint32) uint32 {
		le.PutUint32(m.buf[0x1000:], 0x2000)
		for i, t := range append(thunks, 0) {
			le.PutUint32(m.buf[0x2000+4*i:], t)
		}
		return 0x1000
	}
	names := func(n int) []uint32 { return bytes2u32(bytes.Repeat([]byte{0xb6, 0x20, 0, 0}, n)) }
	for _, c := range []struct {
		thunks []uint32
		want   uint32
	}{
		{names(1), 0xfbfdf28f}, {names(5), 0x1020}, {names(6), 0x4841}, {names(7), 0x40004f46}, {names(8), 0xa8005026},
		{names(12), 0xd96ad046}, {names(300), 0xd96ad046}, {names(1000), 0x8cb38cf4},
		{[]uint32{0x80000001}, 0x1bfdf290}, {[]uint32{0x80000001, 0x20b6}, 0xe380048f}, {[]uint32{0x20b6, 0x80000001}, 0xff80048f},
		{nil, 0xdfed6290},
	} {
		clear(m.buf[0x2000:])
		got, ok := petiteEntry(m, 0xdfed1249, list(c.thunks...), 0x5042, [2]uint64{0x10000, 0x10000})
		if !ok || got != c.want {
			t.Errorf("%d imports: %#x, want %#x", len(c.thunks), got, c.want)
		}
	}
}

func bytes2u32(b []byte) []uint32 {
	out := make([]uint32, len(b)/4)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(b[4*i:])
	}
	return out
}

func FuzzPetite(f *testing.F) {
	a, b := ptPlain('a', 0xffc), ptPlain('b', 0xf7)
	p := ptProgram([]ptBlock{{a, 0x4000}, {b, 0x1000}, {ptCode(ptImports(), 0xdfed1249, 0x205c), 0x2000}})
	f.Add(p.File, uint32(0x2000), uint32(0x4000), uint32(0x6000))
	f.Fuzz(func(t *testing.T, file []byte, vsize, rva, image uint32) {
		if len(file) > 1<<16 || image > 1<<20 {
			t.Skip()
		}
		q := p
		q.File, q.ImageSize = file, image
		q.Sections = append([]PetiteSection{}, p.Sections...)
		q.Sections[0].VSize, q.Sections[2].RVA = vsize, rva
		Petite(q)
	})
}
