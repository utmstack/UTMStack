package unpack

import (
	"bytes"
	"encoding/binary"
)

// The reference scanner's ASPack 2.12 unpacker, learned from its debug
// output and rebuilt programs on the upstream ASPack test program, a
// single-byte flip of each of its 7,680 bytes, and variants with re-encoded
// blocks, other block tables and section tables
// (docs/edr/native-engine/parity/unpackers.md section 6; lab scripts in
// docs/2026-10-14-unpackers/). The decompressor is in aspack_decode.go.
//
// The entry point (one byte into the stub) starts with the ASPack code the
// parser tests; version 2.12 has "push 0; ret" at stub+0x3ba. The stub's
// decoder tables (stub+0x6d6) are read from the file. Block records of 8
// bytes at stub+0x57c (address, size) end with an address of 0; a record
// whose block is empty or leaves the image ends the loop and fails. Each
// block is decoded from the memory image at its address and written back
// there; the first block's output goes through the call-address filter. The
// rebuilt program keeps every section, or every section but the last two
// when the entry point is in the second-to-last, each written whole from the
// memory image, after a section named "empty" (no data) from 0x1000 when the
// first one starts later; the entry point is the stub's "mov eax" at stub+0x39a. A failure,
// or a stub of another version, lets the parser go on.

// ASPackOutcome is what the reference's parser does at the ASPack step.
type ASPackOutcome uint8

const (
	// ASPackGoesOn: not version 2.12, or unpacking failed; the parser goes on.
	ASPackGoesOn ASPackOutcome = iota
	// ASPackUnpacked: a rebuilt program, scanned as a child; the parser
	// then returns.
	ASPackUnpacked
	// ASPackUnknown: this engine cannot tell.
	ASPackUnknown
)

// ASPackSection is a section as the reference's parser keeps it: address and
// virtual size aligned to the section alignment, raw pointer aligned down and
// raw size up to the file alignment (cut at the file's end).
type ASPackSection struct {
	RVA, VSize, Raw, RawSize uint32
}

// ASPackProgram is what the ASPack step reads.
type ASPackProgram struct {
	File                        []byte
	ImageBase, Entry, ImageSize uint32
	Sections                    []ASPackSection
}

var (
	aspVersion212 = []byte{0x68, 0, 0, 0, 0, 0xc3}
	aspTables     []byte
)

func init() {
	for _, b := range aspLenBase {
		aspTables = append(aspTables, byte(b))
	}
	for _, b := range aspLenExtra {
		aspTables = append(aspTables, byte(b))
	}
	for _, b := range aspPosExtra {
		aspTables = append(aspTables, byte(b))
	}
}

// ASPack runs the reference's ASPack step on a program whose entry code the
// parser recognized. The outcome is ASPackUnknown where the program steps
// outside what was checked against the reference: decoder tables other than
// version 2.12's, raw data past the memory image, a block that runs out of
// input, a block between the two candidate image sizes or below the lowest
// section, no block at all, or a first section starting below 0x1000.
func ASPack(p ASPackProgram) (ASPackOutcome, []byte) {
	le := binary.LittleEndian
	n := len(p.Sections)
	if n < 2 || p.Entry == 0 || p.ImageSize == 0 {
		return ASPackUnknown, nil
	}
	size := uint64(p.ImageSize)
	for _, s := range p.Sections {
		size = max(size, uint64(s.RVA)+uint64(s.VSize))
	}
	if size > 64<<20 {
		return ASPackUnknown, nil
	}
	mem := make([]byte, size)
	lowest := uint64(1) << 32
	for _, s := range p.Sections {
		lowest = min(lowest, uint64(s.RVA))
	}
	// Raw data is loaded at its section's address in table order, a later
	// section overwriting what an earlier one's raw data spilled into.
	for _, s := range p.Sections {
		if s.RawSize == 0 {
			continue
		}
		lo, hi := uint64(s.RVA), uint64(s.RVA)+uint64(s.RawSize)
		if uint64(s.Raw)+uint64(s.RawSize) > uint64(len(p.File)) || hi > size {
			return ASPackUnknown, nil
		}
		copy(mem[lo:hi], p.File[s.Raw:s.Raw+s.RawSize])
	}
	stub := uint64(p.Entry - 1)
	if stub+0x748 > size {
		return ASPackUnknown, nil
	}
	if !bytes.Equal(mem[stub+0x3ba:stub+0x3c0], aspVersion212) {
		return ASPackGoesOn, nil
	}
	if !bytes.Equal(mem[stub+0x6d6:stub+0x748], aspTables) {
		return ASPackUnknown, nil
	}
	// The blocks. The image size the reference bounds them by is the
	// header's or the memory image's; both must agree.
	blocks := 0
	for at := stub + 0x57c; ; at += 8 {
		if at+8 > size {
			return ASPackUnknown, nil
		}
		rva, bsize := uint64(le.Uint32(mem[at:])), uint64(le.Uint32(mem[at+4:]))
		if rva == 0 {
			break
		}
		inHeader, inMemory := rva+bsize <= uint64(p.ImageSize), rva+bsize <= size
		if bsize == 0 || !inHeader && !inMemory {
			return ASPackGoesOn, nil
		}
		if inHeader != inMemory || rva < lowest {
			return ASPackUnknown, nil
		}
		out, r := aspDecode(mem[rva:], uint32(bsize))
		switch r {
		case aspBadTables:
			return ASPackGoesOn, nil
		case aspUndecided:
			return ASPackUnknown, nil
		}
		if blocks == 0 {
			aspFilter(out)
		}
		copy(mem[rva:], out)
		blocks++
	}
	if blocks == 0 {
		return ASPackUnknown, nil
	}
	// The rebuilt program.
	keep := p.Sections
	if ep := p.Sections[n-2]; p.Entry >= ep.RVA && uint64(p.Entry) < uint64(ep.RVA)+uint64(ep.VSize) {
		keep = p.Sections[:n-2]
	}
	if len(keep) == 0 || keep[0].RVA < 0x1000 {
		return ASPackUnknown, nil
	}
	var sections []RebuildSection
	if keep[0].RVA > 0x1000 {
		sections = append(sections, RebuildSection{RVA: 0x1000, VSize: keep[0].RVA - 0x1000, Empty: true})
	}
	for _, s := range keep {
		sections = append(sections, RebuildSection{RVA: s.RVA, VSize: s.VSize, RawSize: s.VSize, Data: mem[s.RVA : uint64(s.RVA)+uint64(s.VSize)]})
	}
	return ASPackUnpacked, Rebuild(sections, le.Uint32(mem[stub+0x39b:]), p.ImageBase, 0, 0)
}

// aspFilter undoes the call-address filter on a block: an E8 or E9 byte
// followed by a zero byte (at most 7 bytes before the end) holds an address
// rotated left by 8 and made absolute; others are left alone.
func aspFilter(b []byte) {
	le := binary.LittleEndian
	left := len(b) - 6
	for i := 0; left > 0; {
		c := b[i]
		if (c == 0xe8 || c == 0xe9) && b[i+1] == 0 {
			v := le.Uint32(b[i+1:])
			v = v<<24 | v>>8
			le.PutUint32(b[i+1:], v-uint32(i))
			i += 5
			left -= 5
			continue
		}
		i++
		left--
	}
}
