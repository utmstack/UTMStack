package unpack

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

// upackTestStream is a raw LZMA stream (literal context 3, no position bits,
// leading zero byte dropped as Upack stores it) of upackTestPlain.
const upackTestStream = "3a9ce96fe3a04942f7b31f92e8aef082f8deaf05b0143717000e643cee4c606430cb23e44bfbe22fbe44d1842b71ad761ffd0a4c00"

func upackTestPlain() []byte {
	data := bytes.Repeat([]byte("upack test program "), 40)
	copy(data[0x20:], []byte{0xe8, 0, 0, 0x12, 0x34})
	copy(data[0x60:], []byte{0xe9, 0, 0, 0xab, 0xcd})
	return data
}

// upackTestProgram lays out the checked Upack layout: the stub bytes the
// reference reads, the variable block at 0x1b0 and the stream in the second
// section.
func upackTestProgram() UpackProgram {
	le := binary.LittleEndian
	f := make([]byte, 0x740)
	for _, c := range upackFixed {
		copy(f[c.at:], c.want)
	}
	plain := upackTestPlain()
	le.PutUint32(f[0x1b0:], 0x401020)
	le.PutUint32(f[0x1c0:], 0x401000+uint32(len(plain)))
	le.PutUint32(f[0x198:], 0x401000)
	le.PutUint32(f[0x188:], 0x4060ae)
	le.PutUint32(f[0x190:], 2)
	stream, _ := hex.DecodeString(upackTestStream)
	copy(f[0x2ae:], stream)
	return UpackProgram{File: f, ImageBase: 0x400000, Sections: []UpackSection{
		{RVA: 0x1000, VSize: 0x5000, URaw: 0x10, URawSize: 0x1f0},
		{RVA: 0x6000, VSize: 0x8000, Raw: 0x200, URaw: 0x200, URawSize: 0x53c},
		{RVA: 0xe000, VSize: 0x1000, URaw: 0x10, URawSize: 0x1f0},
	}}
}

// The reference's Upack unpacker for the checked layout (parity/unpackers.md):
// LZMA output, two call fixups, one section at 0x1000.
func TestUpackAsReference(t *testing.T) {
	le := binary.LittleEndian
	got, out := Upack(upackTestProgram())
	if got != UpackUnpacked {
		t.Fatalf("outcome %d", got)
	}
	e := le.Uint32(out[0x3c:])
	o := out[e+24:]
	want := upackTestPlain()
	copy(want[0x21:], []byte{0x0f, 0x12, 0, 0}) // 0x1234 - (0x21 + 4)
	copy(want[0x61:], []byte{0x68, 0xab, 0, 0}) // 0xabcd - (0x61 + 4)
	if le.Uint32(o[16:]) != 0x20 || le.Uint16(out[e+6:]) != 1 || !bytes.Equal(out[0x200:0x200+len(want)], want) {
		t.Fatalf("entry %#x, sections %d", le.Uint32(o[16:]), le.Uint16(out[e+6:]))
	}
	// Three fixups where there are two, an output start past its end, or a
	// corrupt stream fail; other stub bytes are not decided.
	p := upackTestProgram()
	le.PutUint32(p.File[0x190:], 3)
	if got, _ := Upack(p); got != UpackGoesOn {
		t.Errorf("fixups: outcome %d", got)
	}
	p = upackTestProgram()
	le.PutUint32(p.File[0x198:], 0x402000)
	if got, _ := Upack(p); got != UpackGoesOn {
		t.Errorf("start past end: outcome %d", got)
	}
	p = upackTestProgram()
	p.File[0x641] = 4
	if got, _ := Upack(p); got != UpackUnknown {
		t.Errorf("stub: outcome %d", got)
	}
}
