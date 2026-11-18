package unpack

import "encoding/binary"

// FSG131Program is the FSG 1.31 PE32 layout: three absolute addresses in
// its entry stub, a 16-bit section table in header padding and a stream in
// the following packed section.
type FSG131Program struct {
	File                       []byte
	ImageBase, HeaderSize      uint32
	EmptyRVA, EmptyVSize       uint32
	NextRVA, NextRaw, NextSize uint32
	Entry, EntryRVA            uint32
}

// FSG131 decodes the 1.31 stream with the same bounded decoder as FSG 1.33.
// Invalid layout data falls through to subsequent PE analysis, as it does in
// the reference parser. Bounds that cannot be established stay incomplete.
func FSG131(p FSG131Program) (FSG2Outcome, []byte) {
	const maxImage = 64 << 20
	if p.EmptyVSize > maxImage || uint64(p.NextRaw)+uint64(p.NextSize) > uint64(len(p.File)) ||
		uint64(p.Entry)+17 > uint64(len(p.File)) {
		return FSG2Unknown, nil
	}
	if p.NextSize <= 0x19 || p.EmptyVSize <= p.NextSize {
		return FSG2Returns, nil
	}
	f := p.File
	le := binary.LittleEndian
	ep := f[p.Entry:]
	if ep[0] != 0xbb || ep[5] != 0xbf || ep[10] != 0xbe {
		return FSG2Failed, nil
	}
	supportVA := le.Uint32(ep[1:])
	destVA := le.Uint32(ep[6:])
	sourceVA := le.Uint32(ep[11:])
	if supportVA < p.ImageBase || destVA != p.ImageBase+p.EmptyRVA || sourceVA < p.ImageBase {
		return FSG2Failed, nil
	}
	supportAt := supportVA - p.ImageBase
	sourceRVA := sourceVA - p.ImageBase
	if supportAt >= p.HeaderSize || supportAt >= p.NextRaw ||
		sourceRVA < p.NextRVA || sourceRVA-p.NextRVA >= p.NextRaw {
		return FSG2Failed, nil
	}
	if uint64(p.NextRaw) > uint64(len(f)) || uint64(supportAt)+10 > uint64(p.NextRaw) {
		return FSG2Unknown, nil
	}
	support := f[supportAt:p.NextRaw]
	sections := []RebuildSection{{RVA: p.EmptyRVA}}
	terminated := false
	for at := 0; at+10 < len(support); at += 2 {
		value := le.Uint16(support[at:])
		if value == 1 || value == 2 {
			if le.Uint32(support[at+6:]) == 2 {
				terminated = true
			}
			break
		}
		if value < 2 {
			return FSG2Failed, nil
		}
		address := uint64(value-2) << 12
		if address < uint64(p.ImageBase) {
			return FSG2Failed, nil
		}
		rva := address - uint64(p.ImageBase)
		if rva < uint64(p.EmptyRVA) || rva-uint64(p.EmptyRVA) >= uint64(p.EmptyVSize) {
			return FSG2Failed, nil
		}
		sections = append(sections, RebuildSection{RVA: uint32(rva)})
		if len(sections) > 96 {
			return FSG2Unknown, nil
		}
	}
	if !terminated || p.EntryRVA < p.NextRVA {
		return FSG2Failed, nil
	}
	src := f[p.NextRaw : p.NextRaw+p.NextSize]
	oldEPAt := uint64(p.EntryRVA-p.NextRVA) + 0xda + 2
	gp := uint32(0xda)
	if ep[16] == 0xe8 {
		oldEPAt += 6
		gp += 6
	}
	if oldEPAt+4 > uint64(len(src)) {
		return FSG2Unknown, nil
	}
	oldEP := p.EntryRVA + gp + 6 + le.Uint32(src[oldEPAt:])
	return fsgDecodeSections(src, int(sourceRVA-p.NextRVA), p.EmptyRVA, p.EmptyVSize, sections, oldEP, p.ImageBase)
}
