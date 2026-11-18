package unpack

import (
	"bytes"
	"encoding/binary"
	"sort"
)

// The reference scanner's PEtite 2.2 unpacker, learned from its debug output
// and rebuilt programs on the upstream PEtite test program and about 200
// variants of it with re-encoded blocks, changed records, section tables and
// import lists (docs/edr/native-engine/parity/unpackers.md; lab scripts in
// docs/2026-10-14-unpackers/). The decoder follows the PEtite stub inside the
// sample.
//
// The entry point's "mov eax, imm32" names the stub, the last section. A
// table at stub+0x1b8 holds move records (a dword count with bit 31 set, a
// source and a destination: a backwards dword copy) and decompress records
// (source, size, destination, an unused dword), ended by a source of 0. The
// reference loads the sections at their addresses, runs the records, then
// builds one section per decompress record: the first section in the table
// whose aligned extent holds the destination names it (the whole block must
// fit there, else unpacking fails); the first record there takes the
// section's address and its raw data runs from it to the block's end, a
// later one keeps its own address and size. The block whose last 0x355 bytes
// are the PEtite code loses them; that code holds the encrypted entry point
// and the import list the entry point is decrypted with. Sections are sorted
// by address; each one's virtual size reaches the next, the last keeps the
// block size (or the distance to the end of the last move before it). A
// section whose data would end past the stub (counting a 0x400-byte header)
// is written empty. The program is rebuilt the generic way (Rebuild).

// PetiteOutcome is what the reference's parser does at the PEtite step.
type PetiteOutcome uint8

const (
	// PetiteFailed: unpacking fails and the parser goes on to its next step.
	PetiteFailed PetiteOutcome = iota
	// PetiteUnpacked: a rebuilt program, which the reference scans as a
	// child; the parser then returns.
	PetiteUnpacked
	// PetiteUnknown: this engine cannot tell.
	PetiteUnknown
)

// PetiteSection is a section of the packed program as the reference's parser
// reads it: the address, the virtual size aligned to the section alignment,
// and the raw data's offset and size.
type PetiteSection struct {
	RVA, VSize, Raw, RawSize uint32
}

// PetiteProgram is what the PEtite step reads.
type PetiteProgram struct {
	File                      []byte
	ImageBase, Entry          uint32
	ImageSize                 uint32
	ResourceRVA, ResourceSize uint32
	Sections                  []PetiteSection
}

const (
	petiteTable    = 0x1b8
	petiteCodeSize = 0x355
	petiteEPKey    = 0xdfcf64a8
	petiteMaxRecs  = 16
)

var (
	// petiteStripMark is at the PEtite code's offset 0x54, petiteEPMark at
	// its offset 7: the reference strips the code, and decrypts the entry
	// point, only where they are found.
	petiteStripMark = []byte{0x33, 0xc0, 0x5e, 0x64, 0x8b, 0x18, 0x8b, 0x1b}
	petiteEPMark    = []byte{0x02, 0x05, 0xa9, 0x42, 0x2b, 0xa0, 0xc7, 0x36}
)

// petiteMemory is the reference's image of the program. Bytes outside the
// sections' extents are not modeled: reading them is not decided.
type petiteMemory struct {
	buf   []byte
	low   uint32 // the lowest section address
	limit uint32 // reads at or past it are out of the reference's buffer
	bad   bool
}

func (m *petiteMemory) in(at, n uint64) bool {
	if at < uint64(m.low) || at+n > uint64(len(m.buf)) {
		m.bad = true
		return false
	}
	return true
}

func (m *petiteMemory) u32(at uint64) uint32 {
	if !m.in(at, 4) {
		return 0
	}
	return binary.LittleEndian.Uint32(m.buf[at:])
}

type petiteOut struct {
	rva, vsize, rsize uint32
	dst, size         uint32
}

// Petite runs the reference's PEtite 2.2 unpacker on a program whose entry
// point names its last section. The outcome is PetiteUnknown wherever the
// program steps outside what was checked against the reference: raw data
// past the image or overlapping in memory, a block past the image, reads
// below the first section or past the image, a record of size 0, more than
// 16 blocks or moves, a decoder that copies from before its block or overruns it, a
// move whose areas overlap the wrong way, or a decision that falls between
// the candidate limits learned.
func Petite(p PetiteProgram) (PetiteOutcome, []byte) {
	le := binary.LittleEndian
	n := len(p.Sections)
	if n < 2 || p.ImageSize == 0 || p.ImageSize > 64<<20 {
		return PetiteUnknown, nil
	}
	stub := p.Sections[n-1]
	m := &petiteMemory{buf: make([]byte, p.ImageSize), low: ^uint32(0)}
	type span struct{ lo, hi uint64 }
	var loaded []span
	for _, s := range p.Sections {
		if s.VSize == 0 || s.RawSize != 0 && uint64(s.RVA)+uint64(s.RawSize) > uint64(p.ImageSize) || s.RawSize > s.VSize ||
			uint64(s.Raw)+uint64(s.RawSize) > uint64(len(p.File)) {
			return PetiteUnknown, nil
		}
		for _, l := range loaded {
			if uint64(s.RVA) < l.hi && l.lo < uint64(s.RVA)+uint64(s.RawSize) {
				return PetiteUnknown, nil
			}
		}
		loaded = append(loaded, span{uint64(s.RVA), uint64(s.RVA) + uint64(s.RawSize)})
		if s.RawSize != 0 {
			copy(m.buf[s.RVA:], p.File[s.Raw:s.Raw+s.RawSize])
		}
		m.low = min(m.low, s.RVA)
	}
	m.limit = uint32(min(uint64(p.ImageSize+0xfff)&^0xfff, 1<<32-1))
	// The records.
	at := uint64(stub.RVA) + petiteTable
	used := make([]bool, n)
	var out []petiteOut
	var moveEnd uint64
	moved, moves := false, 0
	for {
		ecx := m.u32(at)
		if m.bad {
			return PetiteUnknown, nil
		}
		if ecx&0x80000000 != 0 {
			count := uint64(ecx &^ 0x80000000)
			src, dst := uint64(m.u32(at+4)), uint64(m.u32(at+8))
			at += 12
			if moves++; m.bad || moves > petiteMaxRecs || count == 0 || src+4 < 4*count || dst+4 < 4*count {
				return PetiteUnknown, nil
			}
			lo, slo := dst+4-4*count, src+4-4*count
			if !m.in(lo, 4*count) || !m.in(slo, 4*count) || dst < src && lo < src+4 {
				return PetiteUnknown, nil
			}
			copy(m.buf[lo:dst+4], m.buf[slo:src+4])
			moveEnd, moved = dst+4, true
			continue
		}
		if ecx == 0 {
			break
		}
		size, dst := m.u32(at+4), m.u32(at+8)
		at += 16
		if m.bad || size == 0 || len(out) == petiteMaxRecs {
			return PetiteUnknown, nil
		}
		hit := -1
		for i, s := range p.Sections {
			if s.RVA <= dst && uint64(dst) < uint64(s.RVA)+uint64(s.VSize) {
				hit = i
				break
			}
		}
		if hit < 0 || uint64(dst)+uint64(size) > uint64(p.Sections[hit].RVA)+uint64(p.Sections[hit].VSize) {
			return PetiteFailed, nil
		}
		if uint64(dst)+uint64(size) > uint64(p.ImageSize) {
			return PetiteUnknown, nil
		}
		if r := petiteDecode(m, uint64(ecx), dst, size); r != PetiteUnpacked {
			return r, nil
		}
		vsize := size
		if moved && moveEnd > uint64(dst) {
			vsize = uint32(moveEnd - uint64(dst))
		}
		o := petiteOut{rva: dst, vsize: vsize, rsize: size, dst: dst, size: size}
		if !used[hit] {
			used[hit] = true
			o.rva = p.Sections[hit].RVA
			o.rsize = dst + size - o.rva
		}
		out = append(out, o)
	}
	// The PEtite code at the end of a block.
	found := -1
	for k, o := range out {
		mark := uint64(o.dst) + uint64(o.size) - petiteCodeSize + 0x54
		if o.size < petiteCodeSize {
			// A block shorter than the code: memory before it that holds the
			// mark is not decided.
			if uint64(o.dst)+uint64(o.size)+0x54 >= petiteCodeSize && m.in(mark, 8) && bytes.Equal(m.buf[mark:mark+8], petiteStripMark) {
				return PetiteUnknown, nil
			}
			m.bad = false
			continue
		}
		if bytes.Equal(m.buf[mark:mark+8], petiteStripMark) {
			if found >= 0 {
				return PetiteUnknown, nil
			}
			found = k
		}
	}
	var entry uint32
	if found >= 0 {
		o := &out[found]
		code := o.dst + o.size - petiteCodeSize
		o.rsize = code - o.rva
		if bytes.Equal(m.buf[code+7:code+15], petiteEPMark) {
			e := le.Uint32(m.buf[code+0xf:]) ^ petiteEPKey
			var ok bool
			before := p.Sections[n-2]
			ends := [2]uint64{uint64(stub.RVA), uint64(before.RVA) + uint64(before.VSize)}
			if entry, ok = petiteEntry(m, e, le.Uint32(m.buf[code+0x121:]), p.Entry, ends); !ok {
				return PetiteUnknown, nil
			}
		}
	}
	// The sections, sorted by address.
	order := make([]int, len(out))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return out[order[a]].rva < out[order[b]].rva })
	sections := make([]RebuildSection, len(out))
	for i, k := range order {
		o := out[k]
		sections[i] = RebuildSection{RVA: o.rva, VSize: o.vsize, RawSize: o.rsize}
		if i > 0 {
			sections[i-1].VSize = o.rva - sections[i-1].RVA
		}
	}
	// A section whose data would end past the stub is written empty. The
	// limit is the stub's address or the end of the section before it, the
	// header 0x400 bytes or the rebuilt one's size: all must agree.
	header := uint64((0xd0 + 24 + 0xe0 + 40*len(sections) + 0x1ff) &^ 0x1ff)
	before := p.Sections[n-2]
	limits := []uint64{uint64(stub.RVA), uint64(before.RVA) + uint64(before.VSize)}
	var offset uint64
	for i := range sections {
		s := &sections[i]
		if s.RawSize == 0 {
			continue
		}
		skip := 0
		for _, h := range []uint64{0x400, header} {
			for _, limit := range limits {
				if h+offset+uint64(s.RawSize) > limit {
					skip++
				}
			}
		}
		switch skip {
		case 0:
			s.Data = m.buf[s.RVA : uint64(s.RVA)+uint64(s.RawSize)]
			offset += uint64(s.RawSize)
		case 4:
			s.RawSize = 0
		default:
			return PetiteUnknown, nil
		}
	}
	rebuilt := Rebuild(sections, entry, p.ImageBase, p.ResourceRVA, p.ResourceSize)
	return PetiteUnpacked, rebuilt
}

// petiteEntry decrypts the original entry point with the import list, as the
// stub's import loop changes it: every import subtracts 1 and rotates the
// value right by 3; a name import the stub sends through one of its jump
// trampolines subtracts 2 (a counter from 0 counts names down; below 0 the
// trampoline pointer, from 7, moves 5 bytes and the counter restarts at its
// low 3 bits). The result is relative to the entry point's end of "jmp". A
// read past the reference's buffer gives the bogus entry 0x1000: for the list
// of lists the buffer ends at the stub (ends holds the candidates, the stub's
// address and the end of the section before it), for a thunk list somewhere
// past it, not settled (reads there are not decided), certainly by the
// image's end. Reads below the first section and more than 2^20 imports are
// not decided.
func petiteEntry(m *petiteMemory, e, imports, entry uint32, ends [2]uint64) (uint32, bool) {
	low, high := min(ends[0], ends[1]), max(ends[0], ends[1])
	v, p, c := e, uint32(7), 0
	steps := 0
	for a := uint64(imports); ; a += 4 {
		if a+4 > high {
			return 0x1000, true
		}
		list := m.u32(a)
		if m.bad || a+4 > low {
			return 0, false
		}
		if list == 0 {
			break
		}
		for t := uint64(list); ; t += 4 {
			if t+4 > uint64(m.limit) {
				return 0x1000, true
			}
			thunk := m.u32(t)
			if steps++; m.bad || t+4 > low || steps > 1<<20 {
				return 0, false
			}
			if thunk == 0 {
				break
			}
			d := uint32(1)
			if thunk&0x80000000 == 0 {
				if c--; c < 0 {
					p += 5
					c, d = int(p&7), 2
				}
			}
			v -= d
			v = v>>3 | v<<29
		}
	}
	return v + entry + 5, true
}

// petiteDecode runs the stub's decoder: a raw first byte, then bits MSB first
// from tag bytes (each loaded byte carries a sentinel), literals XORed with
// the low byte of the bytes left, matches with a gamma-coded high part, 5, 7
// or 8 low bits by block size and length bonuses for far offsets. The block
// ends when the bytes left reach exactly 0.
func petiteDecode(m *petiteMemory, src uint64, dst, size uint32) PetiteOutcome {
	t1, t2, low := uint32(0xfffffc60), uint32(0xffffc060), 5
	switch {
	case size >= 0x40000:
		t1, t2, low = 0xfffffb00, 0xffff8300, 8
	case size >= 0x10000:
		t1, t2, low = 0xfffff980, 0xffff8180, 7
	}
	tag := uint32(0)
	byteAt := func() byte {
		if !m.in(src, 1) {
			return 0
		}
		v := m.buf[src]
		src++
		return v
	}
	bit := func() uint32 {
		tag = (tag & 0xff) << 1
		if tag&0xff != 0 {
			return tag >> 8 & 1
		}
		tag = uint32(byteAt())<<1 | 1
		return tag >> 8 & 1
	}
	gamma := func() uint32 {
		v := uint32(1)
		for {
			v = v<<1 | bit()
			if bit() == 0 || m.bad || v > 1<<30 {
				return v
			}
		}
	}
	out := uint64(dst)
	m.buf[out] = byteAt()
	out++
	left := int64(size) - 1
	last := uint32(0)
	for left > 0 {
		if m.bad {
			return PetiteUnknown
		}
		if bit() == 0 {
			m.buf[out] = byteAt() ^ byte(left)
			out++
			left--
			continue
		}
		bonus := uint32(0)
		g := gamma()
		if g < 3 {
			if last == 0 {
				return PetiteUnknown
			}
		} else {
			v := g - 3
			for i := 0; i < low; i++ {
				v = v<<1 | bit()
			}
			last = ^v
			bonus = 1
			if last < t1 {
				bonus++
			}
			if last < t2 {
				bonus++
			}
		}
		n := bit()<<1 | bit()
		if n == 0 {
			n = gamma() + 2
		}
		n += bonus
		if m.bad {
			return PetiteUnknown
		}
		offset := uint64(uint32(-last))
		if offset == 0 || offset > out-uint64(dst) || int64(n) > left {
			return PetiteUnknown
		}
		for i := uint32(0); i < n; i++ {
			m.buf[out] = m.buf[out-offset]
			out++
		}
		left -= int64(n)
	}
	return PetiteUnpacked
}
