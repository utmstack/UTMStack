package unpack

import (
	"bytes"
	"encoding/binary"
)

// The reference scanner's yC 1.3 unpacker, learned from its debug output and
// rebuilt programs on the upstream yC test program, a single-byte flip of
// each of its header and stub bytes, and variants with other operation
// blocks (docs/edr/native-engine/parity/unpackers.md section 5; lab scripts
// in docs/2026-10-14-unpackers/). The two decryption layers follow the stub
// inside the sample.
//
// The entry point is the last section's start + 0x60, where the stub begins
// with fixed code (51 bytes, with the constants that give the decryptor's
// place and length) and whose first loop ends with "stosb; loop" at +0xc3.
// The first loop's body, 0x30 bytes at +0x93, is a chain of byte operations
// on al the reference emulates; with cl counting down from 0xb6b it decrypts
// the decryptor at +0xc6. The decryptor's own loop body, 0x30 bytes at
// +0x457, decrypts every section but the last whose name does not start with
// rsrc, .rsr, relo, .rel, yC, .eda or .tls and whose raw pointer and size are
// nonzero, cl counting down from the raw size. The program is rebuilt from the
// file up to the stub's raw data: one section fewer, the entry point the
// dword at decryptor+0xa0f, the image size less the stub's aligned virtual
// size, the import directory cleared.

// YCOutcome is what the reference's parser does at the yC step.
type YCOutcome uint8

const (
	// YCGoesOn: no yC stub, or unpacking failed; the parser goes on.
	YCGoesOn YCOutcome = iota
	// YCUnpacked: a rebuilt program, which the reference scans as a child;
	// the parser then returns.
	YCUnpacked
	// YCUnknown: this engine cannot tell.
	YCUnknown
)

// YCSection is a section as the reference reads it: its name's first four
// bytes, its address and virtual size aligned to the section alignment, its
// raw pointer aligned down to the file alignment, and its raw size as
// written (cut at the file's end).
type YCSection struct {
	Name           [4]byte
	RVA, VSize     uint32
	Raw            uint32
	URaw, URawSize uint32
}

// YCProgram is what the yC step reads: the file, the offset of its PE header
// and its sections (none removed by the parser).
type YCProgram struct {
	File     []byte
	PE       int
	Sections []YCSection
}

var (
	ycStub = []byte{
		0x55, 0x8b, 0xec, 0x53, 0x56, 0x57, 0x60, 0xe8, 0, 0, 0, 0, 0x5d, 0x81, 0xed, 0x8c, 0x21, 0x40, 0, 0xb9,
		0x51, 0x2d, 0x40, 0, 0x81, 0xe9, 0xe6, 0x21, 0x40, 0, 0x8b, 0xd5, 0x81, 0xc2, 0xe6, 0x21, 0x40, 0, 0x8d,
		0x3a, 0x8b, 0xf7, 0x33, 0xc0, 0xeb, 0x04, 0x90, 0xeb, 0x01, 0xc2, 0xac,
	}
	ycLoopEnd = []byte{0xaa, 0xe2, 0xcc}
	ycSkip    = []string{"rsrc", ".rsr", "relo", ".rel", "yC\x00\x00", ".eda", ".tls"}
)

const (
	ycDecryptor = 0xc6
	ycLength    = 0xb6b
)

// ycOp is one emulated operation on al.
type ycOp struct {
	code byte
	v    byte
}

// ycParse reads a 0x30-byte operation block as the reference does: the
// opcode decides, the second byte matters only for the rotates; a short
// jump skips its displacement; any other opcode fails. ok is false where an
// instruction would run past the block (not decided).
func ycParse(block []byte) (ops []ycOp, failed, ok bool) {
	for i := 0; i < len(block); {
		k := block[i]
		need := 1
		switch k {
		case 0x02, 0x2a, 0x32, 0xfe, 0x04, 0x2c, 0x34, 0xd2, 0xeb:
			need = 2
		case 0xc0:
			need = 3
		}
		if i+need > len(block) {
			return nil, false, false
		}
		switch k {
		case 0x02, 0x2a, 0x32, 0xfe:
			ops = append(ops, ycOp{code: k})
		case 0x04, 0x2c, 0x34:
			ops = append(ops, ycOp{code: k, v: block[i+1]})
		case 0xc0:
			op := ycOp{code: 'r', v: block[i+2]} // ror
			if block[i+1] == 0xc0 {
				op.code = 'l' // rol
			}
			ops = append(ops, op)
		case 0xd2:
			op := ycOp{code: 'L'} // rol by cl
			if block[i+1] == 0xc8 {
				op.code = 'R' // ror by cl
			}
			ops = append(ops, op)
		case 0x90, 0xf8, 0xf9:
		case 0xeb:
			i += int(block[i+1])
		default:
			return nil, true, true
		}
		i += need
	}
	return ops, false, true
}

func ycRun(ops []ycOp, al, cl byte) byte {
	rol := func(v, n byte) byte { n &= 7; return v<<n | v>>(8-n) }
	for _, op := range ops {
		switch op.code {
		case 0x02:
			al += cl
		case 0x2a:
			al -= cl
		case 0x32:
			al ^= cl
		case 0xfe:
			al--
		case 0x04:
			al += op.v
		case 0x2c:
			al -= op.v
		case 0x34:
			al ^= op.v
		case 'l':
			al = rol(al, op.v)
		case 'r':
			al = rol(al, 8-op.v&7)
		case 'L':
			al = rol(al, cl)
		case 'R':
			al = rol(al, 8-cl&7)
		}
	}
	return al
}

// YC runs the reference's yC step on a program whose entry point is its last
// section's start + 0x60. The outcome is YCUnknown where the program steps
// outside what was checked: a stub shorter than its decryptor, an operation
// that runs past its block, a decrypted range that overlaps the header or the
// decryptor, a stub raw pointer past the file, or no section to decrypt with
// a second block the reference might or might not read.
func YC(p YCProgram) (YCOutcome, []byte) {
	le := binary.LittleEndian
	n := len(p.Sections)
	f := p.File
	if n < 2 || p.PE < 0 || p.PE+24+0x70 > len(f) {
		return YCUnknown, nil
	}
	stub := p.Sections[n-1]
	base := uint64(stub.Raw)
	if base+ycDecryptor+ycLength > uint64(len(f)) {
		return YCUnknown, nil
	}
	s := f[base:]
	if !bytes.Equal(s[0x60:0x93], ycStub) || !bytes.Equal(s[0xc3:0xc6], ycLoopEnd) {
		return YCGoesOn, nil
	}
	ops, failed, ok := ycParse(s[0x93:0xc3])
	if !ok {
		return YCUnknown, nil
	}
	if failed {
		return YCGoesOn, nil
	}
	if uint64(stub.URaw) > uint64(len(f)) {
		return YCUnknown, nil
	}
	out := append([]byte(nil), f...)
	dec := out[base+ycDecryptor : base+ycDecryptor+ycLength]
	for k := range dec {
		dec[k] = ycRun(ops, dec[k], byte(ycLength-k))
	}
	entry := le.Uint32(out[base+0xa0f:])
	second, failed, ok := ycParse(out[base+0x457 : base+0x487])
	o := p.PE + 24
	header := uint64(o + 0x70) // through the import directory
	decrypted := false
	for _, sec := range p.Sections[:n-1] {
		skip := sec.URaw == 0 || sec.URawSize == 0
		for _, name := range ycSkip {
			if string(sec.Name[:]) == name {
				skip = true
			}
		}
		if skip {
			continue
		}
		if !ok {
			return YCUnknown, nil
		}
		if failed {
			return YCGoesOn, nil
		}
		lo, hi := uint64(sec.Raw), uint64(sec.Raw)+uint64(sec.URawSize)
		if hi > uint64(len(out)) || lo < header || lo < base+ycDecryptor+ycLength && base+ycDecryptor < hi {
			return YCUnknown, nil
		}
		for k := lo; k < hi; k++ {
			out[k] = ycRun(second, out[k], byte(hi-k))
		}
		decrypted = true
	}
	if !decrypted && (failed || !ok) {
		return YCUnknown, nil
	}
	out = out[:stub.URaw]
	if len(out) < o+0x70 {
		return YCUnknown, nil
	}
	le.PutUint16(out[p.PE+6:], uint16(n-1))
	le.PutUint32(out[o+16:], entry)
	le.PutUint32(out[o+56:], le.Uint32(out[o+56:])-stub.VSize)
	le.PutUint64(out[o+96+8:], 0)
	return YCUnpacked, out
}
