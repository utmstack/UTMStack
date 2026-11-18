package unpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// ycEncrypt finds, byte by byte, the input the operations turn into plain.
func ycEncrypt(t *testing.T, ops []ycOp, plain []byte, count int) []byte {
	t.Helper()
	out := make([]byte, len(plain))
	for k, want := range plain {
		cl := byte(count - k)
		found := false
		for x := 0; x < 256; x++ {
			if ycRun(ops, byte(x), cl) == want {
				out[k], found = byte(x), true
				break
			}
		}
		if !found {
			t.Fatal("operations are not a bijection")
		}
	}
	return out
}

// ycTestProgram is a program like the upstream yC test program: .text (to
// decrypt), .rsrc (skipped), and the yC stub, whose first block is block1
// and whose decryptor carries block2 and the entry point.
func ycTestProgram(t *testing.T, text, block1, block2 []byte, entry uint32) []byte {
	le := binary.LittleEndian
	f := make([]byte, 0x400)
	copy(f, "MZ")
	le.PutUint32(f[0x3c:], 0x80)
	e := f[0x80:]
	copy(e, "PE\x00\x00")
	le.PutUint16(e[4:], 0x14c)
	le.PutUint16(e[6:], 3)
	le.PutUint16(e[20:], 0xe0)
	o := e[24:]
	le.PutUint16(o, 0x10b)
	le.PutUint32(o[16:], 0x3060)
	le.PutUint32(o[28:], 0x400000)
	le.PutUint32(o[32:], 0x1000)
	le.PutUint32(o[36:], 0x200)
	le.PutUint32(o[56:], 0x5000)
	le.PutUint32(o[60:], 0x400)
	le.PutUint32(o[96+8:], 0x3000)
	le.PutUint32(o[96+12:], 0xe0)
	le.PutUint32(o[96+16:], 0x2000)
	le.PutUint32(o[96+20:], 0x40)
	for i, s := range []struct {
		name          string
		va, raw, size uint32
	}{{".text", 0x1000, 0x400, 0x200}, {".rsrc", 0x2000, 0x600, 0x200}, {"yC", 0x3000, 0x800, 0xe00}} {
		h := o[0xe0+40*i:]
		copy(h, s.name)
		le.PutUint32(h[8:], 0x1000)
		le.PutUint32(h[12:], s.va)
		le.PutUint32(h[16:], s.size)
		le.PutUint32(h[20:], s.raw)
	}
	le.PutUint32(o[0xe0+80+8:], 0x2000) // the stub's virtual size
	f = append(f, make([]byte, 0x1200)...)
	stub := f[0x800:]
	copy(stub[0x60:], ycStub)
	copy(stub[0x93:], block1)
	copy(stub[0xc3:], ycLoopEnd)
	dec := make([]byte, ycLength)
	copy(dec[0x457-ycDecryptor:], block2)
	le.PutUint32(dec[0xa0f-ycDecryptor:], entry)
	ops1, failed, ok := ycParse(block1)
	if failed || !ok {
		t.Fatal("block 1")
	}
	copy(stub[ycDecryptor:], ycEncrypt(t, ops1, dec, ycLength))
	if ops2, failed, ok := ycParse(block2); !failed && ok {
		copy(f[0x400:], ycEncrypt(t, ops2, text, len(text)))
	}
	copy(f[0x600:], "resource data")
	return f
}

func ycBlock(ops ...byte) []byte {
	return append(ops, bytes.Repeat([]byte{0x90}, 0x30-len(ops))...)
}

func ycProgram(t *testing.T, f []byte) YCProgram {
	le := binary.LittleEndian
	var secs []YCSection
	for i := 0; i < 3; i++ {
		h := f[0x80+24+0xe0+40*i:]
		s := YCSection{RVA: le.Uint32(h[12:]), VSize: 0x1000, Raw: le.Uint32(h[20:]), URaw: le.Uint32(h[20:]), URawSize: le.Uint32(h[16:])}
		copy(s.Name[:], h)
		secs = append(secs, s)
	}
	secs[2].VSize = 0x2000
	return YCProgram{File: f, PE: 0x80, Sections: secs}
}

// The reference's yC unpacker (parity/unpackers.md): both layers emulated,
// skipped section names, the rebuilt header.
func TestYCAsReference(t *testing.T) {
	le := binary.LittleEndian
	text := bytes.Repeat([]byte("plain code "), 0x200/11+1)[:0x200]
	block1 := ycBlock(0x2a, 0xc1, 0xc0, 0xc0, 0xf5, 0xc0, 0xc8, 0xec, 0xfe, 0xc8, 0x04, 0x11, 0xd2, 0xc8, 0xeb, 0x01, 0xcc)
	block2 := ycBlock(0x34, 0x68, 0xc0, 0xc0, 0x69, 0x02, 0xc1, 0x32, 0xc1, 0x2c, 0x05, 0xd2, 0xc0, 0xf8, 0xf9)
	f := ycTestProgram(t, text, block1, block2, 0x1234)
	got, out := YC(ycProgram(t, f))
	if got != YCUnpacked {
		t.Fatalf("outcome %d", got)
	}
	o := out[0x80+24:]
	if len(out) != 0x800 || le.Uint16(out[0x86:]) != 2 || le.Uint32(o[16:]) != 0x1234 || le.Uint32(o[56:]) != 0x3000 ||
		le.Uint64(o[96+8:]) != 0 || le.Uint32(o[96+16:]) != 0x2000 {
		t.Fatalf("header: %d bytes, %d sections, entry %#x, image %#x", len(out), le.Uint16(out[0x86:]), le.Uint32(o[16:]), le.Uint32(o[56:]))
	}
	if !bytes.Equal(out[0x400:0x600], text) || !bytes.HasPrefix(out[0x600:], []byte("resource data")) {
		t.Fatal("sections")
	}
	// An opcode the reference does not handle, in either block, fails: the
	// parser goes on.
	if got, _ := YC(ycProgram(t, ycTestProgram(t, text, block1, ycBlock(0xf6, 0xd0), 0))); got != YCGoesOn {
		t.Errorf("second block: outcome %d", got)
	}
	bad := ycTestProgram(t, text, block1, block2, 0)
	bad[0x800+0x93] = 0xd0
	if got, _ := YC(ycProgram(t, bad)); got != YCGoesOn {
		t.Errorf("first block: outcome %d", got)
	}
	// No stub: the parser goes on; an operation past the block's end is not
	// decided.
	bad = ycTestProgram(t, text, block1, block2, 0)
	bad[0x800+0x91] ^= 0xff
	if got, _ := YC(ycProgram(t, bad)); got != YCGoesOn {
		t.Errorf("stub: outcome %d", got)
	}
	bad = ycTestProgram(t, text, block1, block2, 0)
	bad[0x800+0xc2] = 0xc0
	if got, _ := YC(ycProgram(t, bad)); got != YCUnknown {
		t.Errorf("past the block: outcome %d", got)
	}
}
