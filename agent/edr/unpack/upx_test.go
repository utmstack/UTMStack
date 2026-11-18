package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// testImage is a decoded image the step rebuilds: one 0x1000-byte section
// holding body at the empty section's address (0x1000), then the original
// header UPX keeps (one section, 0x1000 alignment).
func testImage(body []byte) []byte {
	le := binary.LittleEndian
	img := make([]byte, 0x1000)
	copy(img, body)
	h := make([]byte, 24+0xe0+40)
	copy(h, "PE\x00\x00")
	le.PutUint16(h[4:], 0x14c)
	le.PutUint16(h[6:], 1)
	le.PutUint16(h[20:], 0xe0)
	le.PutUint16(h[22:], 0x102)
	o := h[24:]
	le.PutUint16(o, 0x10b)
	le.PutUint32(o[16:], 0x1000)
	le.PutUint32(o[28:], 0x400000)
	le.PutUint32(o[32:], 0x1000)
	le.PutUint32(o[36:], 0x200)
	le.PutUint32(o[56:], 0x2000)
	le.PutUint32(o[60:], 0x400)
	le.PutUint32(o[92:], 16)
	s := o[0xe0:]
	copy(s, ".text")
	le.PutUint32(s[8:], 0xf00)
	le.PutUint32(s[12:], 0x1000)
	le.PutUint32(s[16:], 0x1000)
	le.PutUint32(s[20:], 0x400)
	return append(img, h...)
}

// upxCase is a program with UPX's layout: the stream at offset start of the
// second section (raw 0x400, 0x1400 bytes), the entry code at its 0x1000th
// byte (0x90 filled), the given routine patterns at entry+0x69 and
// entry+0x71, and a "mov esi, imm32" at entry+1 naming skew when it is not
// zero.
func upxCase(stream []byte, start int, at69, at71 []byte, skew uint32) UPXProgram {
	const raw, size, entry = 0x400, 0x1400, 0x400 + 0x1000
	f := make([]byte, raw+size)
	copy(f[raw+start:], stream)
	code := f[entry:]
	for i := range code {
		code[i] = 0x90
	}
	copy(code[0x69:], at69)
	copy(code[0x71:], at71)
	p := UPXProgram{File: f, Entry: entry, EntryBytes: len(code), ImageBase: 0x400000,
		EmptyRVA: 0x1000, EmptyVSize: 0x4000, NextRVA: 0x5000, NextURVA: 0x5000, NextVSize: 0x2000,
		NextRaw: raw, NextSize: size}
	if skew != 0 {
		code[1] = 0xbe
		binary.LittleEndian.PutUint32(code[2:], p.ImageBase+p.NextRVA+skew)
	}
	return p
}

// The reference's attempts, learned from its debug output on crafted
// programs (parity/upx.md section 2): a recognized routine (13 bytes at
// entry+0x69 or entry+0x71, the later offset winning) gets one preferred
// attempt at the section start moved by the entry's mov esi skew; the other
// decoders (all three when nothing is recognized) are tried at the section
// start and 0x15 bytes into it, never at the skew.
func TestUPXAttemptsAsReference(t *testing.T) {
	image := testImage([]byte("decoded program body"))
	d := compress(NRV2D, image)
	routine := func(m Method) []byte {
		for _, r := range upxRoutines {
			if r.method == m {
				return r.code
			}
		}
		t.Fatalf("no routine for %v", m)
		return nil
	}
	b, dd := routine(NRV2B), routine(NRV2D)
	cases := []struct {
		name       string
		start      int
		at69, at71 []byte
		skew       uint32
		want       UPXOutcome
	}{
		{"unrecognized, stream at the start", 0, nil, nil, 0, UPXUnpacked},
		{"unrecognized, stream 0x15 bytes in", 0x15, nil, nil, 0, UPXUnpacked},
		{"unrecognized, stream at the skew", 5, nil, nil, 5, UPXFailed},
		{"recognized, stream at the skew", 5, dd, nil, 5, UPXUnpacked},
		{"recognized, stream at the start", 0, dd, nil, 0, UPXUnpacked},
		{"recognized, stream 0x15 bytes in", 0x15, dd, nil, 0, UPXFailed},
		{"other routine, stream 0x15 bytes in", 0x15, b, nil, 0, UPXUnpacked},
		{"other routine, stream at the skew", 5, b, nil, 5, UPXFailed},
		{"second offset wins: other routine", 0x15, dd, b, 0, UPXUnpacked},
		{"second offset wins: the stream's routine", 0x15, b, dd, 0, UPXFailed},
		{"a routine cut short is not recognized", 0x15, dd[:12], nil, 0, UPXUnpacked},
	}
	for _, c := range cases {
		got, out := UPX(upxCase(d, c.start, c.at69, c.at71, c.skew))
		if got != c.want {
			t.Errorf("%s: outcome %d, want %d", c.name, got, c.want)
			continue
		}
		if got == UPXUnpacked && !bytes.Contains(out, []byte("decoded program body")) {
			t.Errorf("%s: the rebuilt program lacks the decoded body", c.name)
		}
	}
	// Two attempts that succeed with different results: the order between
	// them is not known, so the outcome is not either.
	tiny := compress(NRV2B, []byte("x"))
	if len(tiny) > 0x15 {
		t.Fatalf("tiny stream of %d bytes", len(tiny))
	}
	p := upxCase(d, 0x15, nil, nil, 0)
	copy(p.File[p.NextRaw:], tiny)
	if got, _ := UPX(p); got != UPXUnknown {
		t.Errorf("two successful attempts: outcome %d", got)
	}
}

// The reference's LZMA attempt (parity/upx.md section 2): after every NRV
// attempt fails, an entry starting with pusha and holding UPX 3.91's LZMA
// stub at entry+0x2f gets the source 2 bytes in (0x17 when the mov esi
// skew is 0x15) decoded with lc 3, lp 0, pb 2 to the size at entry+0x21,
// which must not exceed the destination.
func TestUPXLZMAAttemptAsReference(t *testing.T) {
	image := testImage([]byte("decoded program body"))
	stream := lzmaRaw(t, image, LZMAProps{3, 0, 2})
	lzmaCase := func(start int, skew uint32, size uint32, stub []byte) UPXProgram {
		p := upxCase(nil, 0, nil, nil, skew)
		copy(p.File[int(p.NextRaw)+start:], stream)
		code := p.File[p.Entry:]
		code[0] = 0x60
		copy(code[0x2f:], stub)
		binary.LittleEndian.PutUint32(code[0x21:], size)
		return p
	}
	later := append(append([]byte(nil), upxLZMAPattern[:12]...), 0x55, 0x57, 0x56, 0x53, 0x83, 0xec, 0x7c, 0x8b)
	n := uint32(len(image))
	cases := []struct {
		name  string
		start int
		skew  uint32
		size  uint32
		stub  []byte
		want  UPXOutcome
	}{
		{"stream 2 bytes in", 2, 0, n, upxLZMAPattern, UPXUnpacked},
		{"stream 0x17 bytes in, skew 0x15", 0x17, 0x15, n, upxLZMAPattern, UPXUnpacked},
		{"stream 0x17 bytes in, skew 0x14", 0x17, 0x14, n, upxLZMAPattern, UPXUnknown},
		{"size above the destination", 2, 0, 0x6001, upxLZMAPattern, UPXFailed},
		{"a later UPX's stub", 2, 0, n, later, UPXFailed},
	}
	for _, c := range cases {
		if got, _ := UPX(lzmaCase(c.start, c.skew, c.size, c.stub)); got != c.want {
			t.Errorf("%s: outcome %d, want %d", c.name, got, c.want)
		}
	}
	p := lzmaCase(2, 0, n, upxLZMAPattern)
	p.File[p.Entry] = 0x90
	if got, _ := UPX(p); got != UPXFailed {
		t.Errorf("no pusha: outcome %d", got)
	}
}
