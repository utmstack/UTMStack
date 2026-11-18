package unpack

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

// The reference scanner's UPX step, learned from its debug output and the
// rebuilt programs it leaves with --leave-temps, on real UPX-packed programs
// and on crafted ones (docs/edr/native-engine/parity/upx.md). A program has
// an empty section (virtual size, no raw data) followed by one with both.
// The destination holds both sections' aligned virtual sizes; the source is
// the next section's aligned raw data. A decompressor routine the entry code
// holds gets one preferred attempt, its source moved forward by the skew of
// a "mov esi, imm32" at the entry point when that skew is 1 to 0xfff bytes;
// the other decoders are tried at the source and 0x15 bytes into it. The
// first attempt that reads its source to the end marker decompresses. When
// every one fails and the entry code is UPX's older LZMA stub, the source 2
// bytes in (0x17 when the skew is 0x15) is decoded as LZMA. The result is
// rebuilt into a program the reference scans as a child.

// upxRoutines are the decompressor routines the reference recognizes: 13
// bytes at entry+0x69 or entry+0x71, the routine at the later offset winning
// when both hold one (learned from its "Looks like a ... decompression
// routine" lines on stubs with single bytes inverted and on synthetic entry
// code; UPX 4.2.4 stubs use the first NRV2B, NRV2D and NRV2E routines).
var upxRoutines = []struct {
	method Method
	code   []byte
}{
	{NRV2B, []byte{0x11, 0xdb, 0x11, 0xc9, 0x01, 0xdb, 0x75, 0x07, 0x8b, 0x1e, 0x83, 0xee, 0xfc}},
	{NRV2D, []byte{0x83, 0xf0, 0xff, 0x74, 0x78, 0xd1, 0xf8, 0x89, 0xc5, 0xeb, 0x0b, 0x01, 0xdb}},
	{NRV2D, []byte{0xc5, 0xeb, 0x0b, 0x01, 0xdb, 0x75, 0x07, 0x8b, 0x1e, 0x83, 0xee, 0xfc, 0x11}},
	{NRV2E, []byte{0xeb, 0x52, 0x31, 0xc9, 0x83, 0xe8, 0x03, 0x72, 0x11, 0xc1, 0xe0, 0x08, 0x8a}},
	{NRV2E, []byte{0x11, 0xc1, 0xe0, 0x08, 0x8a, 0x06, 0x46, 0x83, 0xf0, 0xff, 0x74, 0x75, 0xd1}},
}

// UPXRoutine is the decompressor routine the entry code holds, if any.
func UPXRoutine(entry []byte) (Method, bool) {
	var m Method
	found := false
	for _, at := range []int{0x69, 0x71} {
		if at+13 > len(entry) {
			break
		}
		for _, r := range upxRoutines {
			if bytes.Equal(entry[at:at+13], r.code) {
				m, found = r.method, true
			}
		}
	}
	return m, found
}

// upxSecondStart is where the non-preferred attempts try the source again.
const upxSecondStart = 0x15

// upxLZMAPattern is the entry code the reference requires at entry+0x2f
// before its LZMA attempt (UPX 3.91 and older stubs; later ones lack the
// five nops): the stub's LZMA settings (lc 3, lp 0, pb 2) are part of it.
// Learned from single inverted bytes and re-encoded streams; the entry
// starts with pusha, the size to decode is the imm32 at entry+0x21.
var upxLZMAPattern = []byte{0x56, 0x83, 0xc3, 0x04, 0x53, 0x50, 0xc7, 0x03, 0x03, 0x00, 0x02, 0x00,
	0x90, 0x90, 0x90, 0x90, 0x90, 0x55, 0x57, 0x56}

// upxLZMAStub reports the entry code the reference's LZMA attempt requires.
func upxLZMAStub(entry []byte) bool {
	return len(entry) >= 0x43 && entry[0] == 0x60 && bytes.Equal(entry[0x2f:0x43], upxLZMAPattern)
}

// UPXProgram is what the step reads of a program, with the section values
// the reference's parser uses (virtual sizes aligned to the section
// alignment, raw pointers aligned down and raw sizes aligned up to the file
// alignment, raw sizes cut at the end of the file).
type UPXProgram struct {
	File       []byte
	Entry      int // entry point file offset
	EntryBytes int // bytes readable there
	ImageBase  uint32
	EmptyRVA   uint32 // the empty section's address
	EmptyVSize uint32
	NextRVA    uint32 // the next section's address and its unaligned value
	NextURVA   uint32
	NextVSize  uint32
	NextRaw    uint32
	NextSize   uint32
}

// UPXOutcome is what the reference's parser does at the UPX step.
type UPXOutcome uint8

const (
	// UPXReturns: a destination not larger than the source; the parser
	// returns ("size mismatch").
	UPXReturns UPXOutcome = iota
	// UPXFailed: every decoder fails; the parser goes on.
	UPXFailed
	// UPXUnpacked: a rebuilt program, which the reference scans as a child
	// before returning.
	UPXUnpacked
	// UPXUnknown: this engine cannot tell.
	UPXUnknown
)

// maxUPXOutput bounds the destination this engine decodes into; larger
// ones meet the reference's scan limits, which are not modeled.
const maxUPXOutput = 64 << 20

// UPX runs the step on a program. The rebuilt program is returned with
// UPXUnpacked.
func UPX(p UPXProgram) (UPXOutcome, []byte) {
	le := binary.LittleEndian
	dst := uint64(p.EmptyVSize) + uint64(p.NextVSize)
	switch {
	case dst <= uint64(p.NextSize):
		return UPXReturns, nil
	case dst > maxUPXOutput:
		return UPXUnknown, nil
	}
	end := uint64(p.NextRaw) + uint64(p.NextSize)
	try := func(m Method, skew uint64) []byte {
		src := p.File[min(uint64(p.NextRaw)+skew, end):end]
		out, _, err := Decompress(m, src, int(dst))
		if err != nil {
			return nil
		}
		return out
	}
	entry := p.File[p.Entry : p.Entry+p.EntryBytes]
	// skew is what the entry's "mov esi, imm32" names past the next
	// section's address, when it names an address there.
	skew := func() (uint32, bool) {
		if len(entry) <= 5 || entry[1] != 0xbe {
			return 0, true
		}
		if p.NextRVA != p.NextURVA {
			return 0, false // which address the reference subtracts is not known
		}
		return le.Uint32(entry[2:]) - p.ImageBase - p.NextRVA, true
	}
	preferred, recognized := UPXRoutine(entry)
	var decoded []byte
	if recognized {
		s, ok := skew()
		if !ok {
			return UPXUnknown, nil
		}
		if s < 1 || s > 0xfff {
			s = 0
		}
		decoded = try(preferred, uint64(s))
	}
	if decoded == nil {
		// The order among the other attempts is not known: two that succeed
		// with different results leave the outcome unknown.
		for _, m := range []Method{NRV2B, NRV2D, NRV2E} {
			if recognized && m == preferred {
				continue
			}
			for _, start := range []uint64{0, upxSecondStart} {
				out := try(m, start)
				switch {
				case out == nil:
				case decoded == nil:
					decoded = out
				case !bytes.Equal(out, decoded):
					return UPXUnknown, nil
				}
			}
		}
	}
	viaLZMA := false
	if decoded == nil && upxLZMAStub(entry) {
		if recognized {
			return UPXUnknown, nil // an NRV routine and the LZMA stub together were not checked
		}
		s, ok := skew()
		if !ok {
			return UPXUnknown, nil
		}
		start := uint64(2)
		if s == upxSecondStart {
			start += upxSecondStart
		}
		size := uint64(le.Uint32(entry[0x21:]))
		switch {
		case size > dst:
			return UPXFailed, nil
		case size == 0:
			return UPXUnknown, nil
		}
		out, err := DecompressLZMA(p.File[min(uint64(p.NextRaw)+start, end):end], LZMAProps{LC: 3, LP: 0, PB: 2}, int(size))
		if err != nil {
			// this decoder's failures are not checked against the reference's
			return UPXUnknown, nil
		}
		decoded, viaLZMA = out, true
	}
	if decoded == nil {
		return UPXFailed, nil
	}
	magic := p.EntryBytes >= 0x10c && p.File[p.Entry+0x106] == 0x8d && p.File[p.Entry+0x107] == 0xbe
	var realstuff uint32
	if magic {
		realstuff = le.Uint32(p.File[p.Entry+0x108:])
	}
	out, ok := rebuildUPX(decoded, p.EmptyRVA, magic, realstuff)
	if !ok {
		return UPXUnknown, nil
	}
	if viaLZMA {
		// After its LZMA attempt the reference's rebuilt program has the
		// destination's size, zero filled past the last section (checked
		// on six UPX 3.91 programs; a longer program was not seen).
		if uint64(len(out)) > dst {
			return UPXUnknown, nil
		}
		out = append(out, make([]byte, dst-uint64(len(out)))...)
	}
	return UPXUnpacked, out
}

// rebuildDOSHeader is the DOS header the reference writes on a rebuilt
// program, its own stub text included (its output, byte for byte).
var rebuildDOSHeader, _ = hex.DecodeString("4d5a90000200000004000f00ffff0000b00000000000000040001a000000000000000000000000000000000000000000000000000000000000000000d00000000e1fb409ba0d00cd21b44ccd21546869732066696c6520776173206372656174656420627920436c616d415620666f7220696e7465726e616c2075736520616e642073686f756c64206e6f742062652072756e2e0d0a436c616d4156202d20412047504c207669727573207363616e6e6572202d20687474703a2f2f7777772e636c616d61762e6e65740d0a24000000")

// rebuildUPX rebuilds a program from the decoded image, which starts at the
// empty section's address. The reference has three routes to the original
// header UPX keeps after its import list: the import list at the entry
// code's "lea edi, [esi+realstuff]" (when that magic is there) with the
// header right after its end; otherwise a search for the import list; and
// last the final "PE\0\0" of the decoded data whose section count is not
// zero. The search never succeeded on the reference (99 programs from UPX
// 1.25 to 5.2.1 and 13 crafted layouts, the import list right before the
// header among them), so without the magic the last copy is taken. With
// the magic, where the import list could lead to another copy the result is
// uncertain: ok requires none between realstuff and the last copy. The
// program is the reference's
// DOS header, the header copy with the time stamp "CLAM", the file
// alignment set to the section alignment and each section's virtual and raw
// size aligned and its data placed right after the previous one's, the
// header area aligned, then each section's data from the decoded image.
// Anything outside what was checked against the reference (no copy, an
// optional header other than 224 bytes, a section alignment other than
// 0x200 or 0x1000, empty or overlapping sections, a section outside the
// decoded data or over the header copy) is uncertain.
func rebuildUPX(decoded []byte, emptyRVA uint32, magic bool, realstuff uint32) ([]byte, bool) {
	le := binary.LittleEndian
	sig := []byte("PE\x00\x00")
	at := len(decoded)
	for {
		at = bytes.LastIndex(decoded[:at], sig)
		if at < 0 || at+8 > len(decoded) {
			return nil, false
		}
		if le.Uint16(decoded[at+6:]) != 0 {
			break
		}
	}
	if magic && int64(realstuff) < int64(at) && bytes.Contains(decoded[realstuff:at], sig) {
		return nil, false
	}
	const optionalSize = 0xe0
	if at+24 > len(decoded) || le.Uint16(decoded[at+20:]) != optionalSize {
		return nil, false
	}
	n := int(le.Uint16(decoded[at+6:]))
	length := 24 + optionalSize + 40*n
	if n == 0 || at+length > len(decoded) {
		return nil, false
	}
	copyOfHeader := decoded[at : at+length]
	salign := le.Uint32(copyOfHeader[24+32:])
	if salign != 0x1000 && salign != 0x200 {
		return nil, false
	}
	align := func(v uint32) uint64 { return (uint64(v) + uint64(salign) - 1) / uint64(salign) * uint64(salign) }
	headerArea := align(uint32(len(rebuildDOSHeader) + length))
	type section struct{ from, size uint64 }
	sections := make([]section, n)
	total := headerArea
	var previousEnd uint64
	for i := range sections {
		s := copyOfHeader[24+optionalSize+40*i:]
		vsize, rva := le.Uint32(s[8:]), le.Uint32(s[12:])
		if vsize == 0 || rva < emptyRVA {
			return nil, false
		}
		from, size := uint64(rva-emptyRVA), align(vsize)
		if from < previousEnd || from+size > uint64(len(decoded)) ||
			from < uint64(at+length) && uint64(at) < from+size {
			return nil, false
		}
		previousEnd = from + size
		sections[i] = section{from, size}
		total += size
	}
	if total > maxUPXOutput {
		return nil, false
	}
	out := make([]byte, total)
	copy(out, rebuildDOSHeader)
	h := out[len(rebuildDOSHeader):]
	copy(h, copyOfHeader)
	copy(h[8:12], "CLAM")
	le.PutUint32(h[24+36:], salign)
	raw := headerArea
	for i, s := range sections {
		e := h[24+optionalSize+40*i:]
		le.PutUint32(e[8:], uint32(s.size))
		le.PutUint32(e[16:], uint32(s.size))
		le.PutUint32(e[20:], uint32(raw))
		copy(out[raw:raw+s.size], decoded[s.from:s.from+s.size])
		raw += s.size
	}
	return out, true
}
