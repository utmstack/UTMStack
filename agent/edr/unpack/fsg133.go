package unpack

import "encoding/binary"

// FSG133Program is the FSG 1.33 layout: support data in the PE header's
// padding, followed by an empty section and the packed section.
type FSG133Program struct {
	File                       []byte
	ImageBase, HeaderSize      uint32
	EmptyRVA, EmptyVSize       uint32
	NextRVA, NextRaw, NextSize uint32
	Entry, EntryRVA            uint32
}

// FSG133 decodes each compressed section into the reference's sequential
// output buffer, then rebuilds sections in RVA order. FSG 1.33 and 2.0 use
// the same bounded bitstream decoder but different PE metadata layouts.
func FSG133(p FSG133Program) (FSG2Outcome, []byte) {
	const maxImage = 64 << 20
	if p.EmptyVSize > maxImage {
		return FSG2Unknown, nil
	}
	if p.NextSize <= 0x19 || p.EmptyVSize <= p.NextSize {
		return FSG2Returns, nil
	}
	f := p.File
	if uint64(p.NextRaw)+uint64(p.NextSize) > uint64(len(f)) || uint64(p.Entry)+168 > uint64(len(f)) {
		return FSG2Unknown, nil
	}
	le := binary.LittleEndian
	if f[p.Entry] != 0xbe {
		return FSG2Failed, nil
	}
	supportVA := le.Uint32(f[p.Entry+1:])
	if supportVA < p.ImageBase || supportVA-p.ImageBase >= p.HeaderSize {
		return FSG2Failed, nil
	}
	supportAt := supportVA - p.ImageBase
	if supportAt > p.NextRaw || p.NextRaw > uint32(len(f)) {
		return FSG2Unknown, nil
	}
	support := f[supportAt:p.NextRaw]
	if len(support) < 20 {
		return FSG2Unknown, nil
	}
	if uint64(le.Uint32(support[4:])) != uint64(p.ImageBase)+uint64(p.EmptyRVA) {
		return FSG2Failed, nil
	}
	sourceVA := le.Uint32(support[8:])
	if sourceVA < p.ImageBase {
		return FSG2Failed, nil
	}
	sourceRVA := sourceVA - p.ImageBase
	if sourceRVA < p.NextRVA || uint64(sourceRVA-p.NextRVA) >= uint64(p.NextSize) {
		return FSG2Failed, nil
	}
	sections := []RebuildSection{{RVA: p.EmptyRVA}}
	terminated := false
	for at := 12; at+4 < len(support); at += 4 {
		va := le.Uint32(support[at:])
		if va == 0 {
			terminated = true
			break
		}
		if va <= p.ImageBase {
			return FSG2Failed, nil
		}
		rva := va - p.ImageBase - 1
		if rva < p.EmptyRVA || uint64(rva-p.EmptyRVA) >= uint64(p.EmptyVSize) {
			return FSG2Failed, nil
		}
		sections = append(sections, RebuildSection{RVA: rva})
		if len(sections) > 96 {
			return FSG2Unknown, nil
		}
	}
	if !terminated {
		return FSG2Failed, nil
	}
	oldEP := p.EntryRVA + 167 + le.Uint32(f[p.Entry+163:])
	src := f[p.NextRaw : p.NextRaw+p.NextSize]
	return fsgDecodeSections(src, int(sourceRVA-p.NextRVA), p.EmptyRVA, p.EmptyVSize, sections, oldEP, p.ImageBase)
}

// fsgDecodeSections is shared by the 1.31 and 1.33 header layouts. Both
// versions hand their section table and stream to the same reference decoder.
func fsgDecodeSections(src []byte, sourceOffset int, emptyRVA, imageSize uint32, sections []RebuildSection, oldEP, imageBase uint32) (FSG2Outcome, []byte) {
	image := make([]byte, imageSize)
	d := &mewDecoder{src: src, pos: sourceOffset}
	out := 0
	for i := range sections {
		d.image = image[out:]
		d.last = 0
		end, result := d.block(0)
		if result == MEWFailed {
			return FSG2Failed, nil
		}
		if result == MEWUnknown {
			return FSG2Unknown, nil
		}
		sections[i].RawSize = uint32(end)
		sections[i].Data = image[out : out+end]
		out += end
	}
	// ClamAV sorts the records by RVA while leaving each one's data offset
	// attached to that record.
	for i := 0; i < len(sections); i++ {
		for j := i + 1; j < len(sections); j++ {
			if sections[j].RVA < sections[i].RVA {
				sections[i], sections[j] = sections[j], sections[i]
			}
		}
	}
	for i := range sections {
		if i+1 < len(sections) {
			sections[i].VSize = sections[i+1].RVA - sections[i].RVA
		} else {
			sections[i].VSize = imageSize - (sections[i].RVA - emptyRVA)
		}
	}
	return FSG2Unpacked, Rebuild(sections, oldEP, imageBase, 0, 0)
}
