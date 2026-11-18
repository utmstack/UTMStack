package unpack

import (
	"bytes"
	"encoding/binary"
)

// FSG2Program describes the empty section and the packed section read by
// ClamAV's FSG 2.0 PE path. The file offsets and sizes are the parser's
// aligned values, after truncation to the file size.
type FSG2Program struct {
	File                       []byte
	ImageBase                  uint32
	EmptyRVA, EmptyVSize       uint32
	NextRVA, NextRaw, NextSize uint32
	Entry                      uint32
}

type FSG2Outcome uint8

const (
	FSG2Failed   FSG2Outcome = iota // precondition/decompression failure; try later PE steps
	FSG2Unpacked                    // rebuilt child to scan, then stop PE parsing
	FSG2Returns                     // size mismatch; reference ends PE parsing
	FSG2Unknown                     // input outside this bounded implementation
)

// FSG2 decodes the FSG 2.0 stream and rebuilds the PE child. Its pointer
// chain follows the xchg [absolute] entry code into the packed section's
// stack and helper table. The stream uses the same aPLib-style bit coding as
// MEW; the caller scans the rebuilt child with ordinary signatures and hooks.
func FSG2(p FSG2Program) (FSG2Outcome, []byte) {
	const maxImage = 64 << 20
	f := p.File
	if p.EmptyVSize > maxImage {
		return FSG2Unknown, nil
	}
	if p.NextSize <= 0x19 || p.EmptyVSize <= p.NextSize {
		return FSG2Returns, nil
	}
	if uint64(p.NextRaw)+uint64(p.NextSize) > uint64(len(f)) || uint64(p.Entry)+6 > uint64(len(f)) {
		return FSG2Unknown, nil
	}
	le := binary.LittleEndian
	if !bytes.Equal(f[p.Entry:p.Entry+2], []byte{0x87, 0x25}) {
		return FSG2Failed, nil
	}
	src := f[p.NextRaw : p.NextRaw+p.NextSize]
	sectionOffset := func(va uint32, length uint64) (uint64, bool) {
		if va < p.ImageBase {
			return 0, false
		}
		rva := uint64(va - p.ImageBase)
		if rva < uint64(p.NextRVA) || rva-uint64(p.NextRVA)+length > uint64(len(src)) {
			return 0, false
		}
		return rva - uint64(p.NextRVA), true
	}
	slot, ok := sectionOffset(le.Uint32(f[p.Entry+2:]), 4)
	if !ok {
		return FSG2Failed, nil
	}
	stack, ok := sectionOffset(le.Uint32(src[slot:]), 32)
	if !ok {
		return FSG2Failed, nil
	}
	destVA := le.Uint32(src[stack:])
	if uint64(destVA) != uint64(p.ImageBase)+uint64(p.EmptyRVA) {
		return FSG2Failed, nil
	}
	stream, ok := sectionOffset(le.Uint32(src[stack+4:]), 1)
	if !ok {
		return FSG2Failed, nil
	}
	helper, ok := sectionOffset(le.Uint32(src[stack+16:]), 16)
	if !ok {
		return FSG2Failed, nil
	}
	oldEP := le.Uint32(src[helper+12:]) - p.ImageBase
	image := make([]byte, p.EmptyVSize)
	d := &mewDecoder{src: src[stream:], image: image}
	_, result := d.block(0)
	if result == MEWFailed {
		return FSG2Failed, nil
	}
	if result == MEWUnknown {
		return FSG2Unknown, nil
	}
	child := Rebuild([]RebuildSection{{RVA: p.EmptyRVA, VSize: p.EmptyVSize, RawSize: p.EmptyVSize, Data: image}}, oldEP, p.ImageBase, 0, 0)
	return FSG2Unpacked, child
}
