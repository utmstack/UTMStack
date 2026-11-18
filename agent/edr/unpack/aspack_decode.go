package unpack

// The ASPack 2.12 decompressor, written from the stub inside the upstream
// ASPack test program (docs/edr/native-engine/parity/unpackers.md section 6):
// an LZX-style coder. A 721-symbol main table codes 256 literals, 58
// position slots times 8 length codes, and a symbol that reads new tables; a
// 28-symbol table codes long lengths, an 8-symbol table the low three bits
// of far positions (unless all its lengths are 3), and a 19-symbol pretree
// the other tables' code lengths, delta-coded against the previous ones.
// Positions 0-2 repeat one of the last three offsets. Bits are read MSB
// first through a 24-bit window over a 32-bit register refilled a byte at a
// time.

var (
	aspLenBase  = [28]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 0xa, 0xc, 0xe, 0x10, 0x14, 0x18, 0x1c, 0x20, 0x28, 0x30, 0x38, 0x40, 0x50, 0x60, 0x70, 0x80, 0xa0, 0xc0, 0xe0}
	aspLenExtra = [28]uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5}
	aspPosExtra = [58]uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 14, 14, 15, 15, 16, 16,
		17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 18, 18, 18, 18, 18, 18, 18, 18}
	aspPosBase [58]uint32
)

func init() {
	var b uint32
	for i, e := range aspPosExtra {
		aspPosBase[i] = b
		b += 1 << e
	}
}

const (
	aspMainSize  = 0x2d1
	aspTableSize = 0x2f5
)

// aspDecodeResult tells how a block's decoding ended.
type aspDecodeResult uint8

const (
	aspDecoded   aspDecodeResult = iota
	aspBadTables                 // an incomplete table or a copy outside the block: the reference fails
	aspUndecided                 // the input runs out
)

type aspBits struct {
	d   []byte
	p   int
	bc  uint
	v   uint32
	bad bool
}

func (b *aspBits) window() uint32 {
	for b.bc >= 8 {
		if b.p >= len(b.d) {
			b.bad = true
			return 0
		}
		b.v = b.v<<8 | uint32(b.d[b.p])
		b.p++
		b.bc -= 8
	}
	return b.v >> (8 - b.bc) & 0xffffff
}

func (b *aspBits) get(n uint) uint32 {
	w := b.window()
	b.bc += n
	return w >> (24 - n)
}

type aspTable struct {
	limit, base [16]uint32
	sym         []uint16
}

func (t *aspTable) build(lengths []byte) bool {
	var count [16]uint32
	for _, l := range lengths {
		count[l&0xf]++
	}
	count[0] = 0
	code := uint32(0)
	for l := 1; l < 16; l++ {
		code += count[l] << (24 - l)
		if code > 0x1000000 {
			return false
		}
		t.limit[l] = code
		t.base[l] = t.base[l-1] + count[l-1]
	}
	if code != 0x1000000 {
		return false
	}
	pos := t.base
	t.sym = make([]uint16, len(lengths))
	for i, l := range lengths {
		if l != 0 {
			t.sym[pos[l]] = uint16(i)
			pos[l]++
		}
	}
	return true
}

func (t *aspTable) decode(b *aspBits) uint32 {
	v := b.window() & 0xfffe00
	l := 1
	for l < 15 && v >= t.limit[l] {
		l++
	}
	b.bc += uint(l)
	return uint32(t.sym[(v-t.limit[l-1])>>(24-l)+t.base[l]])
}

type aspDecoder struct {
	bits                aspBits
	old                 [aspTableSize]byte
	rep                 [3]uint32
	main, lens, aligned aspTable
	useAligned          bool
}

func (d *aspDecoder) tables() bool {
	b := &d.bits
	if b.get(1) == 0 {
		d.old = [aspTableSize]byte{}
	}
	var preLen [19]byte
	for i := range preLen {
		preLen[i] = byte(b.get(4))
	}
	var pre aspTable
	if !pre.build(preLen[:]) {
		return false
	}
	var lens [aspTableSize]byte
	for i := 0; i < aspTableSize && !b.bad; {
		s := pre.decode(b)
		switch {
		case s < 16:
			lens[i] = (d.old[i] + byte(s)) & 0xf
			i++
		case s == 16:
			for n := b.get(2) + 3; n > 0 && i < aspTableSize; n-- {
				lens[i] = lens[i-1]
				i++
			}
		default:
			var n uint32
			if s == 17 {
				n = b.get(3) + 3
			} else {
				n = b.get(7) + 11
			}
			for ; n > 0 && i < aspTableSize; n-- {
				lens[i] = 0
				i++
			}
		}
	}
	if !d.main.build(lens[:aspMainSize]) || !d.lens.build(lens[aspMainSize:0x2ed]) || !d.aligned.build(lens[0x2ed:]) {
		return false
	}
	d.useAligned = false
	for _, l := range lens[0x2ed:] {
		if l != 3 {
			d.useAligned = true
		}
	}
	d.old = lens
	return true
}

// aspDecode decodes a block of size bytes from in.
func aspDecode(in []byte, size uint32) ([]byte, aspDecodeResult) {
	d := &aspDecoder{bits: aspBits{d: in, bc: 0x20}}
	if !d.tables() {
		if d.bits.bad {
			return nil, aspUndecided
		}
		return nil, aspBadTables
	}
	b := &d.bits
	out := make([]byte, 0, size)
	for uint32(len(out)) < size {
		if b.bad {
			return nil, aspUndecided
		}
		s := d.main.decode(b)
		if s < 0x100 {
			out = append(out, byte(s))
			continue
		}
		if s >= 0x2d0 {
			if !d.tables() {
				if b.bad {
					return nil, aspUndecided
				}
				return nil, aspBadTables
			}
			continue
		}
		s -= 0x100
		slot, lc := s>>3, s&7
		length := lc + 2
		if lc == 7 {
			l2 := d.lens.decode(b)
			length += aspLenBase[l2] + b.get(aspLenExtra[l2])
		}
		extra := aspPosExtra[slot]
		pos := aspPosBase[slot]
		if d.useAligned && extra >= 3 {
			hi := b.get(extra - 3)
			pos += d.aligned.decode(b) + hi*8
		} else {
			pos += b.get(extra)
		}
		var off uint32
		if pos < 3 {
			off = d.rep[pos]
			if pos != 0 {
				d.rep[pos] = d.rep[0]
				d.rep[0] = off
			}
		} else {
			off = pos - 3
			d.rep[2], d.rep[1], d.rep[0] = d.rep[1], d.rep[0], off
		}
		off++
		if b.bad {
			return nil, aspUndecided
		}
		if uint64(off) > uint64(len(out)) || uint64(len(out))+uint64(length) > uint64(size) {
			return nil, aspBadTables // a copy from before the block or past its end fails
		}
		for ; length > 0; length-- {
			out = append(out, out[len(out)-int(off)])
		}
	}
	if b.bad {
		return nil, aspUndecided
	}
	return out, aspDecoded
}
