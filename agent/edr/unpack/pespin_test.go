package unpack

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
)

// aplibLiterals writes data as aPLib: the first byte as is, then literals,
// a short match when a byte repeats the one before it, and the end marker.
func aplibLiterals(data []byte) []byte {
	var out []byte
	tagAt, bits := -1, 0
	bit := func(b int) {
		if bits == 0 {
			tagAt, bits = len(out), 8
			out = append(out, 0)
		}
		bits--
		out[tagAt] |= byte(b) << bits
	}
	out = append(out, data[0])
	for i := 1; i < len(data); {
		if i+2 <= len(data) && i >= 1 && data[i] == data[i-1] && data[i+1] == data[i-1] {
			bit(1)
			bit(1)
			bit(0)
			out = append(out, 1<<1) // offset 1, length 2
			i += 2
			continue
		}
		bit(0)
		out = append(out, data[i])
		i++
	}
	bit(1)
	bit(1)
	bit(0)
	return append(out, 0)
}

// pespinBuild packs sections into a synthetic program of the layout the
// PESpin step reads: sections at the start of the file, the stub after them
// with its fields and its layers encrypted, the decryptors given.
type pespinBuild struct {
	sections   [][]byte // plaintext (for packed sections, the program data)
	packed     uint32   // depacked sections
	xor, poly  uint32   // bitmaps
	poly1Code  []byte   // POLY1's operations (0x24 bytes)
	poly1      func(al, cl byte) byte
	sectCode   []byte
	sect       func(al, cl byte) byte
	key8       byte
	keyA, keyB uint32
	key32      uint32
}

const pespinTestEP = 0x40

func invertByte(f func(al, cl byte) byte, want, cl byte) byte {
	for c := 0; c < 256; c++ {
		if f(byte(c), cl) == want {
			return byte(c)
		}
	}
	panic("operation not invertible")
}

func (b pespinBuild) build(t *testing.T) (PESpinProgram, [][]byte) {
	le := binary.LittleEndian
	var file []byte
	var secs []PESpinSection
	var want [][]byte
	for j, plain := range b.sections {
		data := append([]byte(nil), plain...)
		vsize := uint32(len(plain)+0xfff) &^ 0xfff
		if b.packed>>j&1 != 0 {
			data = aplibLiterals(plain)
			want = append(want, append(append([]byte(nil), plain...), make([]byte, int(vsize)-len(plain))...))
		} else {
			want = append(want, plain)
		}
		n := len(data)
		if b.poly>>j&1 != 0 {
			for i := range data {
				data[i] = invertByte(b.sect, data[i], byte(n-i))
			}
		}
		raw := uint32(len(file))
		file = append(file, data...)
		secs = append(secs, PESpinSection{RVA: 0x1000 * uint32(j+1), VSize: vsize, Raw: raw, RawSize: uint32(n)})
	}
	stubAt := len(file)
	ep := stubAt + pespinTestEP
	stub := make([]byte, pespinTestEP+0x3217+4+0x40)
	s := stub[pespinTestEP:]
	s[0xdb], s[0xdc], s[0xe0] = 0xbb, b.key8, 0xb9
	le.PutUint32(s[0xe1:], pespinLen8)
	le.PutUint32(s[0x26eb:], b.keyA)
	le.PutUint32(s[0x26f0:], pespinLenA)
	le.PutUint32(s[0x3207:], b.xor)
	le.PutUint32(s[0x644:], pespinLenB)
	le.PutUint32(s[0x650:], b.keyB)
	le.PutUint32(s[0x28dd:], pespinPoly1)
	copy(s[0x28ec:], b.poly1Code)
	s[0x28ec+pespinCode] = 0xaa
	copy(s[0x755:], b.sectCode)
	s[0x755+pespinCode] = 0xaa
	le.PutUint32(s[0x6f1:], b.poly)
	le.PutUint32(s[0x3061:], b.packed)
	adj := uint32(len(stub) - pespinTestEP - 0x3217)
	le.PutUint32(s[0x429:], adj)
	// The layers, last first.
	for i := 0; i < pespinPoly1; i++ {
		s[0x6d4+i] = invertByte(b.poly1, s[0x6d4+i], byte(pespinPoly1-i))
	}
	pespinStream(s[0x28d3:0x28d3+pespinLenB], b.keyB, 0xed43af32, 1)
	pespinStream(s[0x2d5:0x2d5+pespinLenA], b.keyA, 0x8c328834, 1)
	k := b.key8
	for i := pespinLen8; i > 0; i-- {
		s[0x1fe4+i] ^= k
		k--
	}
	file = append(file, stub...)
	secs = append(secs, PESpinSection{RVA: 0x1000 * uint32(len(secs)+1), VSize: uint32(len(stub)+0xfff) &^ 0xfff,
		Raw: uint32(stubAt), RawSize: uint32(len(stub))})
	for j, sec := range secs[:len(secs)-1] {
		if b.xor>>j&1 != 0 {
			pespinStream(file[sec.Raw:sec.Raw+sec.RawSize], b.key32, 0xed43af31, 0)
		}
	}
	le.PutUint32(file[ep+0x3217:], b.key32+pespinChecksum(file[:len(file)-int(adj)]))
	if int(adj) > len(file)-(ep+0x3217) {
		t.Fatal("the checksum would cover the stored key")
	}
	return PESpinProgram{File: file, Entry: ep, Sections: secs}, want
}

func pespinTestBuild() pespinBuild {
	text := bytes.Repeat([]byte("PESPIN-TEST section data, aaa bbb "), 40)
	return pespinBuild{
		sections: [][]byte{text[:0x200], text[:0x180], text[:0x300]},
		packed:   0b100, xor: 0b011, poly: 0b101,
		// add 0x13; nop; sub cl; jump over one byte; rol 3; dec; xor 0x5a
		poly1Code: append([]byte{0x04, 0x13, 0x90, 0x2a, 0xc1, 0xeb, 0x01, 0xe9, 0xc0, 0xc0, 0x03, 0xfe, 0xc8, 0x34, 0x5a},
			bytes.Repeat([]byte{0xf9}, pespinCode-15)...),
		poly1: func(al, cl byte) byte {
			al = al + 0x13 - cl
			al = al<<3 | al>>5
			return (al - 1) ^ 0x5a
		},
		// xor cl; inc; ror 2 (c0 c8 02); add cl; or al,al
		sectCode: append([]byte{0x32, 0xc1, 0xfe, 0xc0, 0xc0, 0xc8, 0x02, 0x02, 0xc1, 0x0a, 0xc0},
			bytes.Repeat([]byte{0x90}, pespinCode-11)...),
		sect: func(al, cl byte) byte {
			al = (al ^ cl) + 1
			return (al>>2 | al<<6) + cl
		},
		key8: 0x91, keyA: 0x47b3f060, keyB: 0x43a806db, key32: 0x3523a0f5,
	}
}

func TestPESpinUnpacksTheLayers(t *testing.T) {
	p, want := pespinTestBuild().build(t)
	outcome, rebuilt := PESpin(p)
	if outcome != PESpinUnpacked {
		t.Fatalf("outcome %v", outcome)
	}
	le := binary.LittleEndian
	e := int(le.Uint32(rebuilt[0x3c:]))
	if n := int(le.Uint16(rebuilt[e+6:])); n != len(want) {
		t.Fatalf("%d sections, want %d", n, len(want))
	}
	if ep := le.Uint32(rebuilt[e+40:]); ep != 0x1000 {
		t.Errorf("entry point %#x", ep)
	}
	for j, w := range want {
		h := rebuilt[e+24+0xe0+40*j:]
		raw, size := le.Uint32(h[20:]), le.Uint32(h[16:])
		if size != uint32(len(w)) || !bytes.Equal(rebuilt[raw:raw+size], w) {
			t.Errorf("section %d: %d bytes, want %d", j, size, len(w))
		}
	}
}

func TestPESpinChecksAndFailures(t *testing.T) {
	le := binary.LittleEndian
	for name, c := range map[string]struct {
		change func(b *pespinBuild, p *PESpinProgram)
		want   PESpinOutcome
	}{
		"length 8 differs": {func(b *pespinBuild, p *PESpinProgram) { le.PutUint32(p.File[p.Entry+0xe1:], 0x11ff) }, PESpinGoesOn},
		"key 8 opcode":     {func(b *pespinBuild, p *PESpinProgram) { p.File[p.Entry+0xdb] = 0xb8 }, PESpinGoesOn},
		"bogus operation":  {func(b *pespinBuild, _ *PESpinProgram) { b.sectCode[0] = 0xd0 }, PESpinGoesOn},
		"early stosb":      {func(b *pespinBuild, _ *PESpinProgram) { b.sectCode[4] = 0xaa }, PESpinGoesOn},
		"checksum length": {func(b *pespinBuild, p *PESpinProgram) {
			// the length is encrypted; any change garbles it
			p.File[p.Entry+0x429+3] ^= 0x80
		}, PESpinGoesOn},
		"too short": {func(b *pespinBuild, p *PESpinProgram) { p.File = p.File[:p.Entry+0x3000] }, PESpinUnknown},
	} {
		t.Run(name, func(t *testing.T) {
			b := pespinTestBuild()
			b.sectCode = append([]byte(nil), b.sectCode...)
			var p PESpinProgram
			if name == "bogus operation" || name == "early stosb" {
				c.change(&b, nil)
				p, _ = b.build(t)
			} else {
				p, _ = b.build(t)
				c.change(&b, &p)
			}
			if outcome, _ := PESpin(p); outcome != c.want {
				t.Errorf("outcome %v, want %v", outcome, c.want)
			}
		})
	}
}

func TestPESpinChecksum(t *testing.T) {
	for _, c := range []struct {
		data []byte
		want uint32
	}{
		{nil, 0xffffffff},
		{[]byte("PESpin"), 0xe9dd6670},
		{bytes.Repeat([]byte{0x5a}, 300), 0x945de418},
	} {
		if got := pespinChecksum(c.data); got != c.want {
			t.Errorf("%q: %#x, want %#x", c.data, got, c.want)
		}
	}
}

func TestAPLibDepack(t *testing.T) {
	data := []byte("aPLib aaaa round trip, aaaa again")
	got, ok := aplibDepack(aplibLiterals(data), len(data))
	if !ok || !bytes.Equal(got, data) {
		t.Fatalf("%q %v", got, ok)
	}
	if _, ok := aplibDepack(aplibLiterals(data), len(data)-1); ok {
		t.Error("output past the limit accepted")
	}
	if _, ok := aplibDepack(aplibLiterals(data)[:10], len(data)); ok {
		t.Error("cut input accepted")
	}
}

func FuzzPESpin(f *testing.F) {
	f.Add(int64(1), 4)
	f.Fuzz(func(t *testing.T, seed int64, flips int) {
		p, _ := pespinTestBuild().build(t)
		r := rand.New(rand.NewSource(seed))
		for i := 0; i < flips%64; i++ {
			p.File[r.Intn(len(p.File))] ^= byte(1 + r.Intn(255))
		}
		PESpin(p)
	})
}
