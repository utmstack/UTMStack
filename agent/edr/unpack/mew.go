package unpack

import (
	"bytes"
	"encoding/binary"
)

// The reference scanner's MEW unpacker, learned from its debug output and
// the rebuilt programs it leaves with --leave-temps, on the upstream MEW test
// program and on variants of it whose compressed blocks were re-encoded
// (docs/edr/native-engine/parity/unpackers.md; lab scripts in
// docs/2026-10-14-unpackers/). The decoder follows the program's own stub,
// read from the sample: an aPLib-style format whose bits come MSB first from
// tag bytes inside the stream.
//
// The entry point jumps to 0x154 or 0x158 in the header, where the stub's
// "mov esi, imm32" names a table in the second section: the bit reader's
// address, the original entry point and the first block's destination. Each
// block ends with a zero 7-bit offset and is followed by the next block's
// destination, or 0. Blocks are written into the first (empty) section's
// virtual size; one that starts or ends outside it fails. Each block makes a
// rebuilt section from where the previous one ended (the first from the
// empty section's start) to the block's end rounded up to 0x1000; a block
// ending before the previous section's end fails. A failure ends the
// reference's parser.

// MEWOutcome is what the reference's parser does at the MEW step.
type MEWOutcome uint8

const (
	// MEWFailed: unpacking fails and the parser returns.
	MEWFailed MEWOutcome = iota
	// MEWUnpacked: a rebuilt program, which the reference scans as a child.
	MEWUnpacked
	// MEWUnknown: this engine cannot tell.
	MEWUnknown
)

// MEWProgram is what the MEW step reads: the file, its image base, the jump
// target in the header, and the two sections (the empty one and the one
// after it, both with their raw data unaligned as in the header).
type MEWProgram struct {
	File                       []byte
	ImageBase                  uint32
	Target                     uint32
	EmptyRVA, EmptyVSize       uint32
	NextRVA, NextRaw, NextSize uint32
	NextVSize                  uint32
}

// mewStub is the stub's code from its "mov esi" to the end of the block
// loop, as in the upstream sample. Bytes 1-4 (the table) and 0x32-0x35 (a
// call into the second section) vary with the layout. The reference does not
// run the stub; this engine decides only programs with this stub, its bit
// reader and the gamma reader its table names (a superset of the checks).
var mewStub = []byte{
	0xbe, 0, 0, 0, 0, 0x8b, 0xde, 0xad, 0xad, 0x50, 0xad, 0x97, 0xb2, 0x80, 0xa4, 0xb6, 0x80, 0xff, 0x13, 0x73, 0xf9,
	0x33, 0xc9, 0xff, 0x13, 0x73, 0x16, 0x33, 0xc0, 0xff, 0x13, 0x73, 0x21, 0xb6, 0x80, 0x41, 0xb0, 0x10, 0xff, 0x13,
	0x12, 0xc0, 0x73, 0xfa, 0x75, 0x3e, 0xaa, 0xeb, 0xe0, 0xe8, 0, 0, 0, 0, 0x02, 0xf6, 0x83, 0xd9, 0x01, 0x75, 0x0e,
	0xff, 0x53, 0xfc, 0xeb, 0x26, 0xac, 0xd1, 0xe8, 0x74, 0x2f, 0x13, 0xc9, 0xeb, 0x1a, 0x91, 0x48, 0xc1, 0xe0, 0x08,
	0xac, 0xff, 0x53, 0xfc, 0x3d, 0x00, 0x7d, 0x00, 0x00, 0x73, 0x0a, 0x80, 0xfc, 0x05, 0x73, 0x06, 0x83, 0xf8, 0x7f,
	0x77, 0x02, 0x41, 0x41, 0x95, 0x8b, 0xc5, 0xb6, 0x00, 0x56, 0x8b, 0xf7, 0x2b, 0xf0, 0xf3, 0xa4, 0x5e, 0xeb, 0x9b,
	0xad, 0x85, 0xc0, 0x75, 0x90,
}

// mewGamma is the gamma reader the stub calls through the table.
var mewGamma = []byte{0x33, 0xc9, 0x41, 0xff, 0x13, 0x13, 0xc9, 0xff, 0x13, 0x72, 0xf8, 0xc3}

// mewMaxImage bounds the destination this engine decodes; larger ones were
// not checked against the reference.
const mewMaxImage = 64 << 20

// MEW runs the reference's MEW unpacker. The outcome is MEWUnknown wherever
// the program differs from what was checked against the reference: another
// stub, bit reader or gamma reader, a table outside the second section's raw
// data, a stream that reads past it, or a match reaching before the
// destination.
func MEW(p MEWProgram) (MEWOutcome, []byte) {
	le := binary.LittleEndian
	f := p.File
	if p.Target != 0x154 && p.Target != 0x158 || int(p.Target)+len(mewStub) > len(f) {
		return MEWUnknown, nil
	}
	stub := f[p.Target : int(p.Target)+len(mewStub)]
	for i, b := range mewStub {
		if (i < 1 || i > 4) && (i < 0x32 || i > 0x35) && stub[i] != b {
			return MEWUnknown, nil
		}
	}
	if p.NextRaw > uint32(len(f)) || p.NextSize > uint32(len(f))-p.NextRaw || p.EmptyVSize == 0 || p.EmptyVSize > mewMaxImage {
		return MEWUnknown, nil
	}
	source := f[p.NextRaw : p.NextRaw+p.NextSize]
	va := func(rva uint32) uint64 { return uint64(p.ImageBase) + uint64(rva) }
	table := uint64(le.Uint32(stub[1:]))
	if table < va(p.NextRVA) || table-va(p.NextRVA)+12 > uint64(len(source)) {
		return MEWUnknown, nil
	}
	at := int(table - va(p.NextRVA))
	// The bit reader and the gamma reader the table and the stub name.
	reader, ok := mewHeaderCode(f, le.Uint32(source[at:]), p.ImageBase)
	if !ok || !bytes.HasPrefix(reader, []byte{0x02, 0xd2, 0x75}) || len(reader) < 8 || !bytes.Equal(reader[4:6], []byte{0x8a, 0x16}) || reader[6] != 0xeb {
		return MEWUnknown, nil
	}
	ret, ok1 := mewHeaderCode(f, le.Uint32(source[at:])+4+uint32(int8(reader[3])), p.ImageBase)
	more, ok2 := mewHeaderCode(f, le.Uint32(source[at:])+8+uint32(int8(reader[7])), p.ImageBase)
	if !ok1 || !ok2 || len(ret) < 1 || ret[0] != 0xc3 || !bytes.HasPrefix(more, []byte{0x46, 0x12, 0xd2, 0xc3}) {
		return MEWUnknown, nil
	}
	if at < 4 {
		return MEWUnknown, nil
	}
	gamma := uint64(le.Uint32(source[at-4:]))
	if gamma < va(p.NextRVA) || gamma-va(p.NextRVA)+uint64(len(mewGamma)) > uint64(len(source)) ||
		!bytes.Equal(source[gamma-va(p.NextRVA):gamma-va(p.NextRVA)+uint64(len(mewGamma))], mewGamma) {
		return MEWUnknown, nil
	}
	entry := le.Uint32(source[at+4:])
	dest := le.Uint32(source[at+8:])
	image := make([]byte, p.EmptyVSize)
	d := &mewDecoder{src: source, pos: at + 12, image: image}
	var sections []RebuildSection
	previous := uint64(p.EmptyRVA)
	for {
		if uint64(dest) < va(p.EmptyRVA) || uint64(dest)-va(p.EmptyRVA) >= uint64(len(image)) {
			return MEWFailed, nil
		}
		end, result := d.block(int(uint64(dest) - va(p.EmptyRVA)))
		if result != MEWUnpacked {
			return result, nil
		}
		top := (uint64(p.EmptyRVA) + uint64(end) + 0xfff) &^ 0xfff
		if top < previous {
			return MEWFailed, nil
		}
		sections = append(sections, RebuildSection{RVA: uint32(previous), VSize: uint32(top - previous), RawSize: uint32(top - previous)})
		previous = top
		if d.pos+4 > len(source) {
			return MEWUnknown, nil
		}
		dest = le.Uint32(source[d.pos:])
		d.pos += 4
		if dest == 0 {
			break
		}
	}
	for i := range sections {
		s := &sections[i]
		from := uint64(s.RVA) - uint64(p.EmptyRVA)
		to := min(from+uint64(s.VSize), uint64(len(image)))
		if from < to {
			s.Data = image[from:to]
		}
	}
	return MEWUnpacked, Rebuild(sections, entry-p.ImageBase, p.ImageBase, 0, 0)
}

// mewHeaderCode returns the bytes from a virtual address in the header (the
// stub keeps code there).
func mewHeaderCode(f []byte, address, imageBase uint32) ([]byte, bool) {
	if address < imageBase || address-imageBase >= 0x400 || int(address-imageBase) >= len(f) {
		return nil, false
	}
	return f[address-imageBase:], true
}

type mewDecoder struct {
	src   []byte
	pos   int
	bits  byte
	image []byte
	last  int
}

func (d *mewDecoder) byte() (byte, bool) {
	if d.pos >= len(d.src) {
		return 0, false
	}
	b := d.src[d.pos]
	d.pos++
	return b, true
}

// bit is the stub's bit reader: "add dl, dl", and when the register empties,
// the next byte shifted in with the carry.
func (d *mewDecoder) bit() (uint32, bool) {
	carry := d.bits >> 7
	d.bits <<= 1
	if d.bits != 0 {
		return uint32(carry), true
	}
	b, ok := d.byte()
	if !ok {
		return 0, false
	}
	d.bits = b<<1 | carry
	return uint32(b >> 7), true
}

func (d *mewDecoder) gamma() (uint32, bool) {
	v := uint32(1)
	for {
		b, ok := d.bit()
		if !ok {
			return 0, false
		}
		v = v<<1 | b
		more, ok := d.bit()
		if !ok {
			return 0, false
		}
		if more == 0 {
			return v, true
		}
		if v > 1<<30 {
			return 0, false
		}
	}
}

// block decodes one block to image[at:], returning where it ended.
func (d *mewDecoder) block(at int) (int, MEWOutcome) {
	out := at
	d.bits = 0x80
	put := func(b byte) bool {
		if out >= len(d.image) {
			return false
		}
		d.image[out] = b
		out++
		return true
	}
	copyFrom := func(offset, length int) MEWOutcome {
		// Before a previous match sets last, a repeat-last token uses offset
		// zero. The reference copies the destination's existing (zeroed)
		// byte in that case; it is a valid stream, not a bad back-reference.
		if offset < 0 || offset > out {
			return MEWUnknown
		}
		for ; length > 0; length-- {
			if out >= len(d.image) {
				return MEWFailed
			}
			if !put(d.image[out-offset]) {
				return MEWFailed
			}
		}
		return MEWUnpacked
	}
	literal := func() MEWOutcome {
		b, ok := d.byte()
		if !ok {
			return MEWUnknown
		}
		if !put(b) {
			return MEWFailed
		}
		return MEWUnpacked
	}
	if r := literal(); r != MEWUnpacked {
		return 0, r
	}
	afterLiteral := true
	for {
		b, ok := d.bit()
		if !ok {
			return 0, MEWUnknown
		}
		if b == 0 {
			if r := literal(); r != MEWUnpacked {
				return 0, r
			}
			afterLiteral = true
			continue
		}
		if b, ok = d.bit(); !ok {
			return 0, MEWUnknown
		}
		if b == 0 {
			g, ok := d.gamma()
			if !ok {
				return 0, MEWUnknown
			}
			high := g - 1
			if afterLiteral {
				high--
			}
			var offset int
			var length uint32
			if high == 0 {
				if length, ok = d.gamma(); !ok {
					return 0, MEWUnknown
				}
				offset = d.last
			} else {
				low, ok := d.byte()
				if !ok || high > 1<<22 {
					return 0, MEWUnknown
				}
				offset = int(high-1)<<8 | int(low)
				if length, ok = d.gamma(); !ok {
					return 0, MEWUnknown
				}
				switch {
				case offset >= 0x7d00:
					length += 2
				case offset>>8 >= 5:
					length++
				case offset <= 0x7f:
					length += 2
				}
				d.last = offset
			}
			if r := copyFrom(offset, int(length)); r != MEWUnpacked {
				return 0, r
			}
			afterLiteral = false
			continue
		}
		if b, ok = d.bit(); !ok {
			return 0, MEWUnknown
		}
		if b == 0 {
			v, ok := d.byte()
			if !ok {
				return 0, MEWUnknown
			}
			offset := int(v >> 1)
			if offset == 0 {
				return out, MEWUnpacked
			}
			d.last = offset
			if r := copyFrom(offset, 2+int(v&1)); r != MEWUnpacked {
				return 0, r
			}
			afterLiteral = false
			continue
		}
		v := 0
		for i := 0; i < 4; i++ {
			b, ok := d.bit()
			if !ok {
				return 0, MEWUnknown
			}
			v = v<<1 | int(b)
		}
		if v == 0 {
			if !put(0) {
				return 0, MEWFailed
			}
		} else if r := copyFrom(v, 1); r != MEWUnpacked {
			return 0, r
		}
		afterLiteral = true
	}
}
