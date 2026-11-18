package unpack

import "encoding/binary"

// The reference scanner's WWPack32 unpacker, learned from its debug output
// and rebuilt programs on the upstream WWPack test program and variants of
// it with re-encoded sections (docs/edr/native-engine/parity/unpackers.md;
// lab scripts in docs/2026-10-14-unpackers/). The decoder follows the
// WWPack32 1.12 stub inside the sample.
//
// The stub is the last section, at the entry point. Block records of 0x11
// bytes start at stub+0x2a1: the area's distance below the stub, the end of
// the area the stub moves the data to, the data's size in dwords, the end
// of the data within the area, and a flag for another record. The reference
// decodes each block from a copy of its data into a buffer the size of the
// section the area starts (its virtual size rounded up to 0x1000): output
// past it fails. Relocations are not applied. The program is rebuilt from
// its own header: one section fewer (the stub's), the entry point from the
// stub's final jump at stub+0x294, every section's virtual and raw size
// rounded up to 0x1000 and its data at its address, the image size up to
// the last section's end, and the file laid out as the memory image. A
// failure ends the reference's parser.

// WWPackOutcome is what the reference's parser does at the WWPack step.
type WWPackOutcome uint8

const (
	// WWPackFailed: unpacking fails and the parser returns.
	WWPackFailed WWPackOutcome = iota
	// WWPackUnpacked: a rebuilt program, which the reference scans as a child.
	WWPackUnpacked
	// WWPackUnknown: this engine cannot tell.
	WWPackUnknown
)

type wwSection struct {
	va, vsize, raw, rawSize uint32
	header                  int // offset of its entry in the section table
}

// WWPack runs the reference's WWPack unpacker on a program whose entry point
// is the start of its last section, holding the WWPack32 stub. The outcome
// is WWPackUnknown wherever the program differs from what was checked
// against the reference: a section alignment other than 0x1000, raw data
// not in 0x200-byte units, sections out of order or over the header, a
// block whose area does not start a section or whose data lies outside its
// raw data, a stream that runs out before its end marker, or a final jump
// that is not there.
func WWPack(f []byte) (WWPackOutcome, []byte) {
	le := binary.LittleEndian
	if len(f) < 0x40 {
		return WWPackUnknown, nil
	}
	e := int(le.Uint32(f[0x3c:]))
	if e < 0 || e+24 > len(f) {
		return WWPackUnknown, nil
	}
	n := int(le.Uint16(f[e+6:]))
	optSize := int(le.Uint16(f[e+20:]))
	o := e + 24
	table := o + optSize
	if optSize < 0x60 || n < 2 || table+40*n > len(f) {
		return WWPackUnknown, nil
	}
	if le.Uint32(f[o+32:]) != 0x1000 {
		return WWPackUnknown, nil
	}
	headers := le.Uint32(f[o+60:])
	if int(headers) > len(f) || headers < uint32(table+40*n) {
		return WWPackUnknown, nil
	}
	secs := make([]wwSection, n)
	var end uint32
	for i := range secs {
		s := f[table+40*i:]
		sec := wwSection{vsize: le.Uint32(s[8:]), va: le.Uint32(s[12:]), rawSize: le.Uint32(s[16:]), raw: le.Uint32(s[20:]), header: table + 40*i}
		size := (uint64(sec.vsize) + 0xfff) &^ 0xfff
		if sec.va%0x1000 != 0 || sec.va < headers || sec.va < end || size == 0 || size > 64<<20 ||
			sec.raw%0x200 != 0 || sec.rawSize%0x200 != 0 || uint64(sec.raw)+uint64(sec.rawSize) > uint64(len(f)) || sec.rawSize > uint32(size) {
			return WWPackUnknown, nil
		}
		end = sec.va + uint32(size)
		secs[i] = sec
	}
	stub := secs[n-1]
	if le.Uint32(f[o+16:]) != stub.va || stub.rawSize < 0x2b2+0x11 {
		return WWPackUnknown, nil
	}
	code := f[stub.raw : stub.raw+stub.rawSize]
	if code[0x294] != 0xe9 {
		return WWPackUnknown, nil
	}
	entry := stub.va + 0x299 + le.Uint32(code[0x295:])
	// The memory image of the sections the program keeps.
	image := map[int][]byte{}
	for i, s := range secs[:n-1] {
		buf := make([]byte, (uint64(s.vsize)+0xfff)&^0xfff)
		copy(buf, f[s.raw:s.raw+s.rawSize])
		image[i] = buf
	}
	for at := 0x2a1; ; at += 0x11 {
		if at+0x11 > len(code) {
			return WWPackUnknown, nil
		}
		r := code[at:]
		area := uint64(stub.va) - uint64(le.Uint32(r[0:]))
		count, srcEnd := uint64(le.Uint32(r[8:])), uint64(le.Uint32(r[12:]))
		index := -1
		for i, s := range secs[:n-1] {
			if uint64(s.va) == area {
				index = i
			}
		}
		if index < 0 || count == 0 || srcEnd+4 < 4*count || srcEnd+4 > uint64(secs[index].rawSize) {
			return WWPackUnknown, nil
		}
		s := secs[index]
		input := f[uint64(s.raw)+srcEnd+4-4*count : uint64(s.raw)+srcEnd+4]
		out := make([]byte, len(image[index]))
		result := wwDecode(input, out)
		if result != WWPackUnpacked {
			return result, nil
		}
		image[index] = out
		if r[0x10] == 0 {
			break
		}
	}
	// The rebuilt program: the original header with one section fewer.
	size := end
	if n >= 2 {
		last := secs[n-2]
		size = last.va + uint32(len(image[n-2]))
	}
	rebuilt := make([]byte, size)
	copy(rebuilt, f[:headers])
	le.PutUint16(rebuilt[e+6:], uint16(n-1))
	le.PutUint32(rebuilt[o+16:], entry)
	le.PutUint32(rebuilt[o+56:], size)
	for i, s := range secs[:n-1] {
		t := rebuilt[s.header:]
		aligned := uint32(len(image[i]))
		le.PutUint32(t[8:], aligned)
		le.PutUint32(t[16:], aligned)
		le.PutUint32(t[20:], s.va)
		copy(rebuilt[s.va:], image[i])
	}
	clear(rebuilt[stub.header : stub.header+40])
	return WWPackUnpacked, rebuilt
}

// wwBits is the stub's bit reader: bits MSB first from 32-bit words loaded
// from the same byte stream literal bytes come from; the word is reloaded
// the moment it empties, or in the middle of a read that needs more bits.
type wwBits struct {
	in   []byte
	pos  int
	buf  uint32
	left uint
	bad  bool
}

func (b *wwBits) word() uint32 {
	if b.pos+4 > len(b.in) {
		b.bad = true
		return 0
	}
	v := binary.LittleEndian.Uint32(b.in[b.pos:])
	b.pos += 4
	return v
}

func (b *wwBits) bit() uint32 {
	v := b.buf >> 31
	b.buf <<= 1
	b.left--
	if b.left == 0 {
		b.buf, b.left = b.word(), 32
	}
	return v
}

func (b *wwBits) bits(n uint) uint32 {
	if n < b.left {
		v := b.buf >> (32 - n)
		b.buf <<= n
		b.left -= n
		return v
	}
	have := b.left
	var v uint32
	if have > 0 {
		v = b.buf >> (32 - have)
	}
	missing := n - have
	b.buf, b.left = b.word(), 32
	if missing == 0 {
		return v
	}
	v = v<<missing | b.buf>>(32-missing)
	b.buf <<= missing
	b.left = 32 - missing
	return v
}

func (b *wwBits) byte() byte {
	if b.pos >= len(b.in) {
		b.bad = true
		return 0
	}
	v := b.in[b.pos]
	b.pos++
	return v
}

// wwDecode decodes a block into out. The end marker is a two-byte match
// whose 9-bit offset field is all ones.
func wwDecode(in, out []byte) WWPackOutcome {
	b := &wwBits{in: in}
	b.buf, b.left = b.word(), 32
	at := 0
	put := func(v byte) bool {
		if at >= len(out) {
			return false
		}
		out[at] = v
		at++
		return true
	}
	match := func(offset, length int) WWPackOutcome {
		if offset <= 0 || offset > at {
			return WWPackUnknown
		}
		for ; length > 0; length-- {
			if !put(out[at-offset]) {
				return WWPackFailed
			}
		}
		return WWPackUnpacked
	}
	for {
		if b.bad {
			return WWPackUnknown
		}
		if b.bit() == 0 {
			v := b.byte()
			if b.bad {
				return WWPackUnknown
			}
			if !put(v) {
				return WWPackFailed
			}
			continue
		}
		kind := b.bits(2)
		if kind == 3 {
			class := b.bits(2)
			var width uint
			var base int
			if class <= 1 {
				width = uint(class) + 5
				base = 1<<width - 0x1f
			} else {
				width = uint(class) + 6
				base = 1<<width - 0x9f
			}
			v := b.bits(width)
			if b.bad {
				return WWPackUnknown
			}
			if v == 0x1ff {
				return WWPackUnpacked
			}
			if r := match((int(v)+base)&0xffff, 2); r != WWPackUnpacked {
				return r
			}
			continue
		}
		var offset int
		switch c := b.bits(3); {
		case c == 4:
			width := 10 + uint(b.bit())
			offset = int(b.bits(width)) + 1<<width - 0x1f
		case c == 5:
			offset = int(b.bits(12)) + 1<<12 - 0x1f
		case c == 6:
			offset = int(b.bits(14)) + 0x1fe1
		case c == 7:
			offset = int(b.bits(15)) + 0x5fe1
		case c == 3:
			width := 8 + uint(b.bit())
			offset = int(b.bits(width)) + 1<<width - 0x1f
		default:
			width := uint(c) + 5
			offset = int(b.bits(width)) + 1<<width - 0x1f
		}
		offset &= 0xffff
		length := 0
		switch kind {
		case 1:
			length = 3
		case 2:
			length = 4
		default:
			if b.bit() == 0 {
				length = 5 + int(b.bit())
			} else if v := b.bits(3); v != 0 {
				length = int(v) + 6
			} else if v := b.bits(4); v != 0 {
				length = int(v) + 13
			} else {
				width, dx := uint(4), 0xd
				for {
					if width == 7 {
						length = int(b.bits(14))
						break
					}
					dx = (dx+2)*2 - 1
					bit := b.bit()
					width++
					if bit != 0 {
						length = int(b.bits(width)+uint32(dx&0xff)) & 0xff
						break
					}
				}
			}
		}
		if b.bad {
			return WWPackUnknown
		}
		if r := match(offset, length); r != WWPackUnpacked {
			return r
		}
	}
}
