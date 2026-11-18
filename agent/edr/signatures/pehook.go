package signatures

import (
	"bytes"
	"encoding/binary"
	"slices"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
	"github.com/utmstack/UTMStack/agent/edr/unpack"
)

// PE hook programs run inside the reference's Windows-program parser, which
// reads the header a second time with its own options: sections that start
// past the end of the file are removed, and with broken-executable alerts on
// (the endpoint's configuration) strict checks fail headers the target
// reading accepts. The reading and the record hook programs see are
// described in docs/edr/native-engine/parity/pe-hooks.md (2.4, 2.5).

type hookStatus uint8

const (
	hookNone     hookStatus = iota // no PE header: the parser stops before any hook
	hookFailed                     // the reading fails: no hook runs
	hookPE32Plus                   // a 64-bit program: no hook runs
	hookUnknown                    // a header this engine cannot model
	hookParsed
)

// hookRecordSize is struct cli_pe_hook_data.
const hookRecordSize = 648

type hookHeader struct {
	broken     bool // a strict PE format error eligible for the broken alert
	record     [hookRecordSize]byte
	sections   []section // after removal
	entry      int       // entry point file offset
	headers    uint32
	imageBase  uint32
	epBytes    int  // bytes readable at the entry point, at most 4096
	epReadFail bool // the entry-point read fails
	subsystem  uint16
	entryRVA   uint32
	lowest     uint32 // the lowest aligned section address
	optional   uint16 // SizeOfOptionalHeader as written
}

// readHookHeader reads data the way the reference's Windows-program parser
// does before its hook points; strict adds the checks it applies with
// broken-executable alerts on.
func readHookHeader(data []byte, strict bool) (*hookHeader, hookStatus) {
	le := binary.LittleEndian
	size := uint64(len(data))
	if size < 2 || string(data[:2]) != "MZ" && string(data[:2]) != "ZM" {
		return nil, hookNone
	}
	if size < 64 {
		return &hookHeader{broken: true}, hookFailed // truncated DOS header is CL_EFORMAT
	}
	h := uint64(le.Uint32(data[0x3c:0x40]))
	if h == 0 || h+24 > size || string(data[h:h+4]) != "PE\x00\x00" {
		return nil, hookNone
	}
	count := uint64(le.Uint16(data[h+6 : h+8]))
	optional := uint64(le.Uint16(data[h+20 : h+22]))
	o := h + 24
	if count == 0 || optional < 96 || o+96 > size {
		return &hookHeader{broken: true}, hookFailed
	}
	plus := le.Uint16(data[o:o+2]) == 0x20b
	fixed := uint64(96)
	if plus {
		fixed = 112
		if optional < 112 || o+112 > size {
			return &hookHeader{broken: true}, hookFailed
		}
	}
	if count > maxPESections {
		return nil, hookUnknown
	}
	sectionAlign := le.Uint32(data[o+32 : o+36])
	fileAlign := le.Uint32(data[o+36 : o+40])
	subsystem := le.Uint16(data[o+68 : o+70])
	native := subsystem == 1
	if strict && !native && (sectionAlign == 0 || sectionAlign%0x1000 != 0 || fileAlign == 0 || fileAlign%0x200 != 0) {
		return &hookHeader{broken: true}, hookFailed
	}
	directories := min(uint64(le.Uint32(data[o+fixed-4:o+fixed])), 16)
	if optional < fixed+directories*8 {
		return &hookHeader{broken: true}, hookFailed
	}
	if o+fixed+directories*8 > size {
		return nil, hookFailed // the directories cannot be read
	}
	table := o + optional
	if table+count*40 > size {
		return &hookHeader{broken: true}, hookFailed // section headers cannot be read
	}
	entryRVA := le.Uint32(data[o+16 : o+20])
	headers := peAlignUp(le.Uint32(data[o+60:o+64]), sectionAlign)
	header := func(i uint64) []byte { return data[table+i*40 : table+i*40+40] }
	for i := uint64(0); fileAlign != 0x200 && i < count; i++ {
		rawSize, pointer := le.Uint32(header(i)[16:20]), le.Uint32(header(i)[20:24])
		if fileAlign != 0 && rawSize != 0 && pointer%fileAlign != 0 && pointer%0x200 == 0 {
			fileAlign = 0x200
		}
	}
	hh := &hookHeader{headers: headers, subsystem: subsystem, imageBase: le.Uint32(data[o+28 : o+32])}
	sections := make([]section, 0, count)
	var lowest, high, overlay uint32
	for i := uint64(0); i < count; i++ {
		s := header(i)
		virtualSize, virtualAddress := le.Uint32(s[8:12]), le.Uint32(s[12:16])
		rawSize, pointer := le.Uint32(s[16:20]), le.Uint32(s[20:24])
		sec := section{rva: peAlign(virtualAddress, sectionAlign), raw: peAlign(pointer, fileAlign), size: peAlignUp(rawSize, fileAlign),
			vsz: peAlignUp(virtualSize, sectionAlign), chr: le.Uint32(s[36:40]), urva: virtualAddress, uvsz: virtualSize, uraw: pointer, ursz: rawSize}
		if sec.size != 0 {
			if uint64(sec.raw) >= size || uint64(pointer) >= size {
				if count == 1 {
					return &hookHeader{broken: true}, hookFailed // the only section is outside the file
				}
				continue // removed; later sections move down
			}
			sec.size = uint32(min(uint64(sec.size), size-uint64(sec.raw)))
			sec.ursz = uint32(min(uint64(sec.ursz), size-uint64(pointer)))
		}
		if sec.vsz == 0 && sec.size != 0 {
			sec.vsz = peAlignUp(sec.ursz, sectionAlign)
		}
		if strict && (sectionAlign == 0 || virtualAddress%sectionAlign != 0) {
			return &hookHeader{broken: true}, hookFailed
		}
		if virtualAddress>>31 != 0 || virtualSize>>31 != 0 || sec.size != 0 && pointer>>31 != 0 || sec.ursz>>31 != 0 {
			return &hookHeader{broken: true}, hookFailed
		}
		if len(sections) == 0 {
			if strict && virtualAddress != headers {
				return &hookHeader{broken: true}, hookFailed
			}
			lowest, high = sec.rva, sec.rva+sec.size
		} else {
			previous := sections[len(sections)-1]
			if strict && virtualAddress-previous.urva != previous.vsz {
				return &hookHeader{broken: true}, hookFailed
			}
			lowest = min(lowest, sec.rva)
			if sec.rva+sec.size > high {
				high = sec.rva + sec.size
				overlay = sec.raw + sec.size
			}
		}
		sections = append(sections, sec)
	}
	meta := &peMeta{sections: sections, headers: headers, size: size}
	entry, ok := meta.rawAddr(entryRVA)
	if !ok {
		return &hookHeader{broken: true}, hookFailed
	}
	hh.sections, hh.entry, hh.entryRVA, hh.lowest, hh.optional = sections, entry, entryRVA, lowest, uint16(optional)
	hh.epBytes = int(min(uint64(4096), size-uint64(entry)))
	if plus {
		return hh, hookPE32Plus
	}

	r := hh.record[:]
	le.PutUint32(r[4:], uint32(entry))
	le.PutUint16(r[8:], uint16(len(sections)))
	copy(r[12:36], data[h:h+24])
	copy(r[36:132], data[o:o+96])
	copy(r[264:360], data[o:o+96]) // the reference's union seen as the 64-bit header
	dirs := data[o+fixed : o+fixed+directories*8]
	for _, at := range []int{132, 376, 504} {
		copy(r[at:at+128], dirs)
	}
	le.PutUint32(r[632:], uint32(h))
	le.PutUint32(r[636:], overlay)
	le.PutUint32(r[640:], uint32(size-uint64(overlay)))
	le.PutUint32(r[644:], headers)
	return hh, hookParsed
}

// bytecodePE is what a hook program sees: the full record, the sections of
// this reading, the entry point and the header size.
func (h *hookHeader) bytecodePE() *bytecode.PEInfo {
	pe := &bytecode.PEInfo{EntryPoint: uint32(h.entry), HeaderSize: h.headers, Sections: make([]bytecode.Section, len(h.sections)), Record: h.record[:]}
	for i, s := range h.sections {
		pe.Sections[i] = bytecode.Section{RVA: s.rva, VSize: s.vsz, Raw: s.raw, RSize: s.size, Characteristics: s.chr,
			URVA: s.urva, UVSize: s.uvsz, URaw: s.uraw, URSize: s.ursz}
	}
	return pe
}

type hookReach uint8

const (
	reachHooks   hookReach = iota // the unpacker hooks run
	reachNever                    // the reference returns before them
	reachUnknown                  // a built-in unpacker this engine lacks would be attempted
)

// nsPackEntry is the entry code the reference's NsPack unpacker looks for.
var nsPackEntry = []byte{0x9c, 0x60, 0xe8, 0, 0, 0, 0, 0x5d, 0xb8, 0x07, 0, 0, 0}

// unpackerHookReach follows the reference's parser from its all-PE hook to
// its unpacker hook (pe.c 3266-4361, pe-hooks.md 1.5): each built-in
// unpacker is tried behind a structural test of the header and the entry
// code, and a success ends the parser before the unpacker hook. This engine
// decides the UPX step where every decoder certainly fails (upxReach); where
// another unpacker would be tried, or UPX could succeed, the hooks' outcome
// is unknown; a test that needs more than this engine models answers
// unknown too (a superset of the reference's attempts). pe is the PE word
// of the dynamic configuration.
func (h *hookHeader) unpackerHookReach(data []byte, pe uint32) (hookReach, upxResult) {
	var upx upxResult
	reach := h.unpackerReach(data, pe, &upx)
	return reach, upx
}

// upxResult is what the UPX step did, when it was reached; the MEW step
// reports a rebuilt program the same way.
type upxResult struct {
	tried   bool
	outcome unpack.UPXOutcome
	child   []byte
}

func (h *hookHeader) unpackerReach(data []byte, pe uint32, upx *upxResult) hookReach {
	le := binary.LittleEndian
	ep := data[h.entry : h.entry+h.epBytes]
	at := func(i int) byte {
		if i < len(ep) {
			return ep[i]
		}
		return 0
	}
	u32 := func(i int) uint32 {
		if i+4 <= len(ep) {
			return le.Uint32(ep[i:])
		}
		return 0
	}
	has := func(i int, want ...byte) bool { return i+len(want) <= len(ep) && bytes.Equal(ep[i:i+len(want)], want) }
	n := len(h.sections)
	image := h.imageBase
	// The empty-section test: a virtual-only section followed by one with
	// raw data (UPX, FSG and MEW).
	found, index := false, 0
	if pe&(dconfPEUPX|dconfPEFSG|dconfPEMEW) != 0 {
		for i := 0; i+1 < n; i++ {
			s, next := h.sections[i], h.sections[i+1]
			if s.size == 0 && s.vsz != 0 && next.size != 0 && next.vsz != 0 {
				found, index = true, i
				break
			}
		}
	}
	if found && pe&dconfPEMEW != 0 && len(ep) >= 16 && at(0) == 0xe9 {
		if offset := h.entryRVA + u32(1) + 5; offset == 0x154 || offset == 0x158 {
			// The MEW step ends the parser, unpacked or not.
			s, next := h.sections[index], h.sections[index+1]
			if next.raw != next.uraw || next.size != next.ursz {
				return reachUnknown
			}
			outcome, child := unpack.MEW(unpack.MEWProgram{File: data, ImageBase: image, Target: offset,
				EmptyRVA: s.rva, EmptyVSize: s.vsz, NextRVA: next.rva, NextRaw: next.raw, NextSize: next.size, NextVSize: next.vsz})
			switch outcome {
			case unpack.MEWUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.MEWFailed:
				return reachNever
			}
			return reachUnknown
		}
	}
	if len(ep) < 168 {
		return reachNever
	}
	upack := pe&dconfPEUpack != 0 && h.optional == 0x148
	if found || upack {
		if upack && n == 3 && at(0) == 0xbe && u32(1)-image > h.lowest && (at(5) == 0xad && at(6) == 0x50 || at(5) == 0xff && at(6) == 0x36) {
			if at(5) != 0xad {
				return reachUnknown // the other Upack layout
			}
			// Upack: a rebuilt program ends the parser; a failure goes on.
			switch outcome, child := h.upack(data); outcome {
			case unpack.UpackUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.UpackUnknown:
				return reachUnknown
			}
		}
		if !upack && n == 2 && (has(0, 0x60, 0xe8) && u32(2) == 9 ||
			at(0) == 0xbe && u32(1)-image < h.lowest && int32(u32(1)) > int32(image) && has(5, 0xad, 0x8b, 0xf8)) {
			return reachUnknown
		}
		if found && pe&dconfPEFSG != 0 {
			s, next := h.sections[index], h.sections[index+1]
			if has(0, 0x87, 0x25) {
				outcome, child := unpack.FSG2(unpack.FSG2Program{File: data, ImageBase: image,
					EmptyRVA: s.rva, EmptyVSize: s.vsz, NextRVA: next.rva, NextRaw: next.raw,
					NextSize: next.size, Entry: uint32(h.entry)})
				switch outcome {
				case unpack.FSG2Unpacked:
					upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
					return reachNever
				case unpack.FSG2Returns:
					return reachNever
				case unpack.FSG2Unknown:
					return reachUnknown
				}
			}
			if at(0) == 0xbe && u32(1)-image < h.lowest {
				outcome, child := unpack.FSG133(unpack.FSG133Program{File: data, ImageBase: image,
					HeaderSize: h.headers, EmptyRVA: s.rva, EmptyVSize: s.vsz, NextRVA: next.rva,
					NextRaw: next.raw, NextSize: next.size, Entry: uint32(h.entry), EntryRVA: h.entryRVA})
				switch outcome {
				case unpack.FSG2Unpacked:
					upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
					return reachNever
				case unpack.FSG2Returns:
					return reachNever
				case unpack.FSG2Unknown:
					return reachUnknown
				}
			}
			switch {
			case at(0) == 0xbb && u32(1)-image < h.lowest && at(5) == 0xbf && at(10) == 0xbe &&
				h.entryRVA >= next.rva && h.entryRVA-next.rva > next.rva-0xe0:
				outcome, child := unpack.FSG131(unpack.FSG131Program{File: data, ImageBase: image,
					HeaderSize: h.headers, EmptyRVA: s.rva, EmptyVSize: s.vsz, NextRVA: next.rva,
					NextRaw: next.raw, NextSize: next.size, Entry: uint32(h.entry), EntryRVA: h.entryRVA})
				switch outcome {
				case unpack.FSG2Unpacked:
					upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
					return reachNever
				case unpack.FSG2Returns:
					return reachNever
				case unpack.FSG2Unknown:
					return reachUnknown
				}
			}
		}
		if found && pe&dconfPEUPX != 0 {
			upx.tried = true
			upx.outcome, upx.child = h.upx(data, index)
			switch upx.outcome {
			case unpack.UPXReturns, unpack.UPXUnpacked:
				return reachNever
			case unpack.UPXUnknown:
				return reachUnknown
			}
		}
	}
	if len(ep) < 200 {
		return reachNever
	}
	if pe&dconfPEPetite != 0 && at(0) == 0xb8 && n > 0 {
		switch {
		case u32(1) == h.sections[n-1].rva+image:
			// PEtite 2.2: a rebuilt program ends the parser; a failure goes
			// on to the next step.
			switch outcome, child := h.petite(data); outcome {
			case unpack.PetiteUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.PetiteUnknown:
				return reachUnknown
			}
		case n >= 2 && u32(1) == h.sections[n-2].rva+image:
			return reachUnknown // PEtite 2.1
		}
	}
	if n > 1 {
		last := h.sections[n-1]
		if pe&dconfPEPESpin != 0 && h.entryRVA >= last.rva && 0x3217-4 <= last.rva+last.size &&
			h.entryRVA < last.rva+last.size-0x3217-4 && has(4, 0xe8, 0, 0, 0, 0, 0x8b, 0x1c, 0x24, 0x83, 0xc3) {
			// PESpin: a rebuilt program ends the parser; a failure goes on.
			switch outcome, child := h.pespin(data); outcome {
			case unpack.PESpinUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.PESpinUnknown:
				return reachUnknown
			}
		}
		if pe&dconfPEYC != 0 && h.entryRVA == last.rva+0x60 {
			// yC: a rebuilt program ends the parser; no stub or a failure
			// goes on to the next step.
			switch outcome, child := h.yc(data); outcome {
			case unpack.YCUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.YCUnknown:
				return reachUnknown
			}
		}
		if pe&dconfPEWWPack != 0 && h.entryRVA == last.rva && has(0, 0x53, 0x55, 0x8b, 0xe8, 0x33, 0xdb, 0xeb) &&
			has(0x68, 0xe8, 0, 0, 0, 0, 0x58, 0x2d, 0x6d, 0, 0, 0, 0x50, 0x60, 0x33, 0xc9, 0x50, 0x58, 0x50, 0x50) {
			// The WWPack step ends the parser, unpacked or not.
			switch outcome, child := unpack.WWPack(data); outcome {
			case unpack.WWPackUnpacked:
				upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
				return reachNever
			case unpack.WWPackFailed:
				return reachNever
			}
			return reachUnknown
		}
	}
	if pe&dconfPEASPack != 0 && has(0, 0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb) {
		// ASPack 2.12: a rebuilt program ends the parser; another version
		// or a failure goes on to the next step.
		switch outcome, child := h.aspack(data); outcome {
		case unpack.ASPackUnpacked:
			upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
			return reachNever
		case unpack.ASPackUnknown:
			return reachUnknown
		}
	}
	if pe&dconfPENsPack != 0 {
		switch outcome, child := h.nspack(data); outcome {
		case unpack.NsPackUnpacked:
			upx.tried, upx.outcome, upx.child = true, unpack.UPXUnpacked, child
			return reachNever
		case unpack.NsPackUnknown:
			return reachUnknown
		}
	}
	return reachHooks
}

// nspack follows the reference's entry redirect, offset cell and compressed
// record before calling the bounded range decoder.
func (h *hookHeader) nspack(data []byte) (unpack.NsPackOutcome, []byte) {
	le := binary.LittleEndian
	if uint64(h.entry) > uint64(^uint32(0)) {
		return unpack.NsPackUnknown, nil
	}
	meta := &peMeta{sections: h.sections, headers: h.headers, size: uint64(len(data))}
	rep, eprva := uint32(h.entry), h.entryRVA
	if data[rep] == 0xe9 {
		if uint64(rep)+5 > uint64(len(data)) {
			return unpack.NsPackFailed, nil
		}
		eprva += le.Uint32(data[rep+1:]) + 5
		entry, ok := meta.rawAddr(eprva)
		if !ok || uint64(entry)+24 > uint64(len(data)) {
			return unpack.NsPackFailed, nil
		}
		rep = uint32(entry)
	}
	if uint64(rep)+24 > uint64(len(data)) || !bytes.Equal(data[rep:rep+13], nsPackEntry) {
		return unpack.NsPackFailed, nil
	}
	delta := uint32(0x54) - le.Uint32(data[rep+17:])
	if delta > rep || uint64(rep-delta)+4 > uint64(len(data)) {
		return unpack.NsPackFailed, nil
	}
	start := rep + le.Uint32(data[rep-delta:])
	if uint64(start)+20 > uint64(len(data)) {
		return unpack.NsPackFailed, nil
	}
	if le.Uint32(data[start:]) == 0 {
		start += 4
	}
	if uint64(start)+13 > uint64(len(data)) || len(h.sections) == 0 {
		return unpack.NsPackFailed, nil
	}
	packedSize := le.Uint32(data[start+5:]) | 0xff
	dsize := le.Uint32(data[start+9:])
	if uint64(start)+uint64(packedSize) > uint64(len(data)) {
		return unpack.NsPackUnknown, nil
	}
	if dsize == 0 || dsize != h.sections[0].vsz {
		return unpack.NsPackFailed, nil
	}
	eprva += 0x27a
	oepAt, ok := meta.rawAddr(eprva)
	if !ok || uint64(oepAt)+5 > uint64(len(data)) {
		return unpack.NsPackFailed, nil
	}
	oldEP := eprva + 5 + le.Uint32(data[oepAt+1:])
	return unpack.NsPack(unpack.NsPackProgram{Packed: data[start : start+packedSize], ExpectedSize: dsize,
		SectionRVA: h.sections[0].rva, ImageBase: h.imageBase, OriginalEntry: oldEP})
}

// upx runs the reference's UPX step (unpack.UPX) on the empty section at
// index and the section after it.
func (h *hookHeader) upx(data []byte, index int) (unpack.UPXOutcome, []byte) {
	s, next := h.sections[index], h.sections[index+1]
	return unpack.UPX(unpack.UPXProgram{
		File: data, Entry: h.entry, EntryBytes: h.epBytes, ImageBase: h.imageBase,
		EmptyRVA: s.rva, EmptyVSize: s.vsz,
		NextRVA: next.rva, NextURVA: next.urva, NextVSize: next.vsz, NextRaw: next.raw, NextSize: next.size,
	})
}

// petite runs the PEtite 2.2 step on the parser's sections; a section whose
// header fields differ from the aligned values the parser uses is not
// decided.
func (h *hookHeader) petite(data []byte) (unpack.PetiteOutcome, []byte) {
	le := binary.LittleEndian
	sections := make([]unpack.PetiteSection, len(h.sections))
	for i, s := range h.sections {
		if s.rva != s.urva || s.raw != s.uraw || s.size != s.ursz {
			return unpack.PetiteUnknown, nil
		}
		sections[i] = unpack.PetiteSection{RVA: s.rva, VSize: s.vsz, Raw: s.raw, RawSize: s.size}
	}
	r := h.record[:]
	return unpack.Petite(unpack.PetiteProgram{
		File: data, ImageBase: h.imageBase, Entry: h.entryRVA, ImageSize: le.Uint32(r[36+56:]),
		ResourceRVA: le.Uint32(r[132+16:]), ResourceSize: le.Uint32(r[132+20:]), Sections: sections,
	})
}

// upack runs the Upack step on the parser's sections.
func (h *hookHeader) upack(data []byte) (unpack.UpackOutcome, []byte) {
	sections := make([]unpack.UpackSection, len(h.sections))
	for i, s := range h.sections {
		sections[i] = unpack.UpackSection{RVA: s.rva, VSize: s.vsz, Raw: s.raw, URaw: s.uraw, URawSize: s.ursz}
	}
	return unpack.Upack(unpack.UpackProgram{File: data, ImageBase: h.imageBase, Sections: sections})
}

// aspack runs the ASPack step on the parser's sections.
func (h *hookHeader) aspack(data []byte) (unpack.ASPackOutcome, []byte) {
	sections := make([]unpack.ASPackSection, len(h.sections))
	for i, s := range h.sections {
		sections[i] = unpack.ASPackSection{RVA: s.rva, VSize: s.vsz, Raw: s.raw, RawSize: s.size}
	}
	return unpack.ASPack(unpack.ASPackProgram{File: data, ImageBase: h.imageBase, Entry: h.entryRVA,
		ImageSize: binary.LittleEndian.Uint32(h.record[36+56:]), Sections: sections})
}

// pespin runs the PESpin step on the parser's sections; a rebuilt section
// whose header fields differ from the values the parser uses is not decided.
func (h *hookHeader) pespin(data []byte) (unpack.PESpinOutcome, []byte) {
	sections := make([]unpack.PESpinSection, len(h.sections))
	for i, s := range h.sections {
		if i < len(h.sections)-1 && (s.rva != s.urva || s.raw != s.uraw || s.size != s.ursz) {
			return unpack.PESpinUnknown, nil
		}
		sections[i] = unpack.PESpinSection{RVA: s.rva, VSize: s.vsz, Raw: s.raw, RawSize: s.size}
	}
	return unpack.PESpin(unpack.PESpinProgram{File: data, Entry: h.entry, Sections: sections})
}

// yc runs the yC step on the parser's sections, with their names from the
// section table; a program the parser removed sections from is not decided
// (the reference then mixes header and parser indices).
func (h *hookHeader) yc(data []byte) (unpack.YCOutcome, []byte) {
	le := binary.LittleEndian
	e := int(le.Uint32(data[0x3c:]))
	table := e + 24 + int(h.optional)
	if int(le.Uint16(data[e+6:])) != len(h.sections) || table+40*len(h.sections) > len(data) {
		return unpack.YCUnknown, nil
	}
	sections := make([]unpack.YCSection, len(h.sections))
	for i, s := range h.sections {
		sec := unpack.YCSection{RVA: s.rva, VSize: s.vsz, Raw: s.raw, URaw: s.uraw, URawSize: s.ursz}
		copy(sec.Name[:], data[table+40*i:])
		sections[i] = sec
	}
	return unpack.YC(unpack.YCProgram{File: data, PE: e, Sections: sections})
}

// parserResult is what the reference's Windows-program parser adds to an
// object's raw scan: a detection or an undecided cause from its hooks or its
// built-in unpackers, the program its UPX step rebuilt (scanned as a child),
// and whether this engine could follow the parser to its unpackers
// (upxDecided). headerRejected records a proved header-format failure so
// the enclosing scanner does not mistake it for an undecoded PE image.
type parserResult struct {
	name           string
	deferred       bool
	headerRejected bool
	cause          error
	unpacked       []byte
	upxDecided     bool
}

// peParser follows the reference's parser on a 32-bit program (pe-hooks.md
// 1.4-2.7): its own header reading, no further step for a failed reading or
// a 64-bit program, the all-PE hooks whose triggers held, the built-in
// unpackers, then the unpacker hooks unless the parser returned first. A
// built-in unpacker this engine lacks makes the object incomplete, with or
// without hook programs. Each hook gets zero match
// counts and offsets and the full header record; the first detection wins.
// A failed run leaves the object undecided (the reference ignores it) and
// ends the phase.
func (d *Database) peParser(w *matchWork, data []byte, flagged []*rule, permissive bool) (parserResult, error) {
	res := parserResult{upxDecided: true}
	if d.dconf.PE == 0 {
		return res, nil // the parser is off
	}
	slices.SortFunc(flagged, func(a, b *rule) int { return a.hookOrder - b.hookOrder })
	h, st := readHookHeader(data, !permissive)
	res.headerRejected = st == hookFailed && h != nil && h.broken
	switch {
	case st == hookFailed && h != nil && h.broken && !permissive:
		// PE-header CL_EFORMAT becomes this heuristic with --alert-broken.
		// The enclosing scanner first analyzes embedded archive members, so
		// the label is deferred until that scan has found no detection.
		res.name, res.deferred = "Heuristics.Broken.Executable", true
		return res, nil
	case st == hookUnknown:
		res.upxDecided = false
		if len(flagged) > 0 {
			res.cause = errExecutableHeader
		}
		return res, nil
	case st != hookParsed || h.epReadFail:
		return res, nil
	}
	pe := h.bytecodePE()
	run := func(kind uint16) (string, bool, error, error) {
		broke := false
		for _, r := range flagged {
			if r.hook != kind {
				continue
			}
			name, value, cause, err := runBytecode(w, r.bytecode, data, bytecode.MatchContext{PE: pe, Hook: true})
			if w.hookTrace != nil {
				w.hookTrace(r.name, r.hook, value, cause)
			}
			if name != "" || cause != nil || err != nil {
				return name, false, cause, err
			}
			broke = broke || value == 0xcea5e
		}
		return "", broke, nil, nil
	}
	name, broke, cause, err := run(259)
	if name != "" || cause != nil || err != nil || broke {
		res.name, res.cause = name, cause
		return res, err
	}
	reach, upx := h.unpackerHookReach(data, d.dconf.PE)
	if upx.tried {
		switch upx.outcome {
		case unpack.UPXUnpacked:
			res.unpacked = upx.child
		case unpack.UPXUnknown:
			// a program may be rebuilt and scanned that this engine
			// cannot produce
			res.upxDecided = false
			res.cause = unknownBecause(UnknownPEUnpacker)
			return res, nil
		}
	}
	switch reach {
	case reachNever:
		return res, nil
	case reachUnknown:
		// the reference would hand the program to a built-in unpacker this
		// engine lacks, and scan what it produces
		res.cause = unknownBecause(UnknownPEUnpacker)
		return res, nil
	}
	if !slices.ContainsFunc(flagged, func(r *rule) bool { return r.hook == 257 }) {
		return res, nil
	}
	name, _, cause, err = run(257)
	res.name, res.cause = name, cause
	return res, err
}
