package unpack

import (
	"encoding/binary"
	"errors"
)

// The reference scanner's Upack unpacker for the layout of the upstream
// Upack test program (its "var set" path), learned from its debug output and
// rebuilt programs on that program and a single-byte flip of each of its
// 1,852 bytes (docs/edr/native-engine/parity/unpackers.md section 7; lab
// scripts in docs/2026-10-14-unpackers/). The packer's own stub, run in an
// emulator, decodes the same bytes; its stream is standard LZMA (literal
// context 3, no position bits) without the leading zero byte.
//
// The entry code's "mov esi, imm32" names a variable block in the header
// (0x11b0 past the image base, at file offset 0x1b0): the original entry
// point there, the output end 0x10 after it, the output start 0x18 before,
// the number of call fixups 0x20 before and the stream's address 0x28
// before. The search byte of the call fixups is the copied code's "cmp al"
// at file offset 0x140. The program is rebuilt as one section at 0x1000 as
// long as the output, taken from the memory image the output was written
// into, the entry point the original one less the image base and 0x1000.

// UpackOutcome is what the reference's parser does at the Upack step.
type UpackOutcome uint8

const (
	// UpackGoesOn: unpacking failed; the parser goes on.
	UpackGoesOn UpackOutcome = iota
	// UpackUnpacked: a rebuilt program, scanned as a child; the parser then
	// returns.
	UpackUnpacked
	// UpackUnknown: this engine cannot tell.
	UpackUnknown
)

// UpackSection is a section as the parser keeps it: address and virtual
// size aligned to the section alignment, raw pointer aligned down to the
// file alignment, and the raw pointer and size as written (the size cut at
// the file's end).
type UpackSection struct {
	RVA, VSize, Raw uint32
	URaw, URawSize  uint32
}

// UpackProgram is what the Upack step reads.
type UpackProgram struct {
	File      []byte
	ImageBase uint32
	Sections  []UpackSection
}

// upackFixed are the stub bytes the reference reads at fixed file offsets
// (each change of one of them changed or failed its result), as in the
// upstream sample; another value is not decided here.
var upackFixed = []struct {
	at   int
	want []byte
}{
	{0x18, []byte{0xbe, 0xb0, 0x11, 0x40, 0x00, 0xad, 0x50, 0xff, 0x76, 0x34, 0xeb, 0x7c}},
	{0xa0, []byte{0xff, 0x76, 0x38, 0xad, 0x50, 0x8b, 0x3e, 0xbe, 0xf0, 0xe0, 0x40, 0x00}},
	{0xb8, []byte{0xdf, 0xab, 0xeb, 0x1c}},
	{0xd8, []byte{0x40, 0xab, 0x40, 0xb1, 0x04, 0xf3, 0xab, 0xc1, 0xe0, 0x0a, 0xb5, 0x1c, 0xf3, 0xab, 0x8b, 0x7e, 0x0c,
		0x57, 0x51, 0xe9, 0x37, 0x53, 0x00, 0x00}},
	{0x134, []byte{0x8a, 0x07, 0x47, 0x04, 0x18, 0x3c, 0x02, 0x73, 0xf7, 0x8b, 0x07, 0x3c}},
	{0x1b8, []byte{0x3c, 0x65, 0x40, 0x00}},
	{0x63c, []byte{0xb6, 0x6f, 0xff, 0xc1, 0xed, 0x05, 0x66, 0x69, 0xed, 0x00, 0x03}},
	{0x650, []byte{0x01, 0xe3, 0x1f, 0x2b}},
	{0x678, []byte{0xc9, 0x5f, 0xe9, 0xf2, 0x00, 0x00, 0x00}},
}

// Upack runs the reference's Upack step for the checked layout. The outcome
// is UpackUnknown where the program differs from it: other stub bytes, a
// variable block elsewhere, an image base other than 0x400000, an output
// outside the image, or a stream that runs out.
func Upack(p UpackProgram) (UpackOutcome, []byte) {
	le := binary.LittleEndian
	f := p.File
	if len(f) < 0x6a0 || p.ImageBase != 0x400000 || len(p.Sections) != 3 {
		return UpackUnknown, nil
	}
	for _, c := range upackFixed {
		for i, b := range c.want {
			if f[c.at+i] != b {
				return UpackUnknown, nil
			}
		}
	}
	// The layout: the first section at 0x1000 over the header (raw pointer
	// 0x10), the second with an aligned raw pointer, the sections adjacent.
	s0, s1, s2 := p.Sections[0], p.Sections[1], p.Sections[2]
	if s0.RVA != 0x1000 || s0.Raw != 0 || s0.URaw != 0x10 || s0.URawSize < 0x1c8 || s1.URaw != s1.Raw ||
		s0.RVA+s0.VSize != s1.RVA || s1.RVA+s1.VSize != s2.RVA {
		return UpackUnknown, nil
	}
	oep, end, dest := le.Uint32(f[0x1b0:]), le.Uint32(f[0x1c0:]), le.Uint32(f[0x198:])
	in, loops, search := le.Uint32(f[0x188:]), le.Uint32(f[0x190:]), f[0x140]
	if dest > end {
		return UpackGoesOn, nil // "loc_edi > end_edi"
	}
	image := uint32(0)
	for _, s := range p.Sections {
		image = max(image, s.RVA+s.VSize)
	}
	base := p.ImageBase
	if dest < base+0x1000 || end-base > image || end-dest > 64<<20 {
		return UpackUnknown, nil
	}
	// The stream, read from the section that holds its address.
	var src []byte
	for _, s := range p.Sections {
		if rva := uint64(in - base); rva >= uint64(s.RVA) && rva < uint64(s.RVA)+uint64(s.URawSize) {
			at, end := uint64(s.Raw)+rva-uint64(s.RVA), uint64(s.Raw)+uint64(s.URawSize)
			if end > uint64(len(f)) {
				return UpackUnknown, nil
			}
			src = f[at:end]
		}
	}
	if src == nil {
		return UpackUnknown, nil
	}
	size := int(end - dest)
	out, err := DecompressLZMA(append([]byte{0}, src...), LZMAProps{LC: 3}, size)
	switch {
	case errors.Is(err, ErrLZMAInput):
		return UpackUnknown, nil
	case err != nil:
		return UpackGoesOn, nil
	}
	// Call fixups: an E8 or E9 byte followed by the search byte holds a
	// big-endian address made relative again.
	at := 0
	for n := uint32(0); n < loops; n++ {
		for {
			if at+5 > size {
				return UpackGoesOn, nil // "callfixerr"
			}
			c := out[at]
			at++
			if (c == 0xe8 || c == 0xe9) && out[at] == search {
				break
			}
		}
		v := le.Uint32(out[at:])
		v = v&^0xff | uint32(0) // the search byte becomes the low byte 0
		v = v>>24 | v>>8&0xff00 | v<<8&0xff0000 | v<<24
		le.PutUint32(out[at:], v-uint32(at+4))
		at += 4
	}
	// The section is the memory image from 0x1000, as long as the output:
	// the sections' raw data, then the output written at its address.
	mem := make([]byte, 0x1000+uint64(image))
	for _, s := range p.Sections {
		if s.URawSize == 0 {
			continue
		}
		if uint64(s.RVA)+uint64(s.URawSize) > uint64(len(mem)) || uint64(s.Raw)+uint64(s.URawSize) > uint64(len(f)) {
			return UpackUnknown, nil
		}
		copy(mem[s.RVA:], f[s.Raw:s.Raw+s.URawSize])
	}
	copy(mem[dest-base:], out)
	data := mem[0x1000 : 0x1000+size]
	rebuilt := Rebuild([]RebuildSection{{RVA: 0x1000, VSize: uint32(size), RawSize: uint32(size), Data: data}}, oep-base-0x1000, base, 0, 0)
	return UpackUnpacked, rebuilt
}
