package unpack

import "encoding/binary"

// NsPackOutcome describes the result of a NsPack unpacking attempt.
type NsPackOutcome uint8

// A malformed record can otherwise consume billions of range-decoder steps
// before filling the bounded destination. Stop as undecided instead.
const nsMaxCodeSteps = 100_000_000

const (
	NsPackFailed   NsPackOutcome = iota // the reference continues to later PE steps
	NsPackUnpacked                      // a rebuilt child is scanned, then PE parsing ends
	NsPackUnknown                       // outside this bounded decoder
)

// NsPackProgram is the compressed record selected by the PE parser, and the
// metadata the reference uses when rebuilding its single output section.
type NsPackProgram struct {
	Packed                []byte
	ExpectedSize          uint32
	SectionRVA, ImageBase uint32
	OriginalEntry         uint32
}

// NsPack decodes the range-coded payload and rebuilds its PE child. The
// reference reads a 13-byte record header, then a five-byte range-code seed;
// every output byte or match advances a bounded destination.
func NsPack(p NsPackProgram) (NsPackOutcome, []byte) {
	const maxImage = 64 << 20
	if p.ExpectedSize == 0 || p.ExpectedSize > maxImage {
		return NsPackUnknown, nil
	}
	if len(p.Packed) < 13 {
		return NsPackFailed, nil
	}
	le := binary.LittleEndian
	inputSize := le.Uint32(p.Packed[5:])
	dsize := le.Uint32(p.Packed[9:])
	if inputSize <= 13 || dsize != p.ExpectedSize || inputSize > uint32(len(p.Packed)) {
		return NsPackFailed, nil
	}
	c := uint32(p.Packed[0])
	if c >= 0xe1 {
		return NsPackFailed, nil
	}
	first := uint32(0)
	if c >= 45 {
		first = c / 45
		c -= first * 45
	}
	alloc := uint32(0)
	if c >= 9 {
		alloc = c / 9
		c -= alloc * 9
	}
	tre := c
	shift := tre + alloc
	if shift > 12 {
		return NsPackUnknown, nil
	}
	tableCount := (0x300 << shift) + 0x736
	table := make([]uint16, tableCount)
	for i := range table {
		table[i] = 0x400
	}
	r := nsRange{src: p.Packed[13:inputSize], bitmap: ^uint32(0), table: table}
	for i := 0; i < 5; i++ {
		r.oldval = r.oldval<<8 | uint32(r.byte())
	}
	if r.bad {
		return NsPackFailed, nil
	}
	output := make([]byte, dsize)
	if !r.decode(output, tre, alloc, first, inputSize) {
		if r.limited {
			return NsPackUnknown, nil
		}
		return NsPackFailed, nil
	}
	child := Rebuild([]RebuildSection{{RVA: p.SectionRVA, VSize: dsize, RawSize: dsize, Data: output}}, p.OriginalEntry, p.ImageBase, 0, 0)
	return NsPackUnpacked, child
}

type nsRange struct {
	src     []byte
	pos     int
	oldval  uint32
	bitmap  uint32
	table   []uint16
	bad     bool
	steps   uint64
	limited bool
}

func (r *nsRange) step() bool {
	r.steps++
	if r.steps > nsMaxCodeSteps {
		r.bad, r.limited = true, true
		return false
	}
	return true
}

func (r *nsRange) byte() byte {
	if r.pos >= len(r.src) {
		r.bad = true
		return 0xff
	}
	b := r.src[r.pos]
	r.pos++
	return b
}

func (r *nsRange) bit(at uint32) uint32 {
	if !r.step() {
		return 0
	}
	if at >= uint32(len(r.table)) {
		r.bad = true
		return 0
	}
	prob := uint32(r.table[at])
	cut := prob * (r.bitmap >> 11)
	var bit uint32
	if r.oldval < cut {
		r.bitmap = cut
		prob += (0x800 - prob) >> 5
	} else {
		bit = 1
		r.bitmap -= cut
		r.oldval -= cut
		prob -= prob >> 5
	}
	r.table[at] = uint16(prob)
	if r.bitmap < 0x1000000 {
		r.oldval = r.oldval<<8 | uint32(r.byte())
		r.bitmap <<= 8
	}
	return bit
}

func (r *nsRange) treeByte(base uint32) uint32 {
	pos := uint32(1)
	for pos < 0x100 && !r.bad {
		pos = pos<<1 | r.bit(base+pos)
	}
	return pos & 0xff
}

func (r *nsRange) contextualByte(base, seed uint32) uint32 {
	pos := uint32(1)
	for pos < 0x100 && !r.bad {
		high := (seed >> 7) & 1
		seed <<= 1
		bit := r.bit(base + ((high + 1) << 8) + pos)
		pos = pos<<1 | bit
		if high != bit {
			for pos < 0x100 && !r.bad {
				pos = pos<<1 | r.bit(base+pos)
			}
		}
	}
	return pos & 0xff
}

func (r *nsRange) treeBits(base, count uint32) uint32 {
	pos := uint32(1)
	for i := uint32(0); i < count && !r.bad; i++ {
		pos = pos<<1 | r.bit(base+pos)
	}
	return pos - (1 << count)
}

func (r *nsRange) length(base, context uint32) uint32 {
	if r.bit(base) == 0 {
		return r.treeBits(base+(context<<3)+2, 3)
	}
	if r.bit(base+1) == 0 {
		return 8 + r.treeBits(base+(context<<3)+0x82, 3)
	}
	return 16 + r.treeBits(base+0x102, 8)
}

func (r *nsRange) reverseBits(base, count uint32) uint32 {
	pos, value := uint32(1), uint32(0)
	for i := uint32(0); i < count && !r.bad; i++ {
		bit := r.bit(base + pos)
		pos = pos<<1 | bit
		value |= bit << i
	}
	return value
}

func (r *nsRange) directBits(count uint32) uint32 {
	var value uint32
	for i := uint32(0); i < count && !r.bad; i++ {
		if !r.step() {
			break
		}
		r.bitmap >>= 1
		value <<= 1
		if r.oldval >= r.bitmap {
			r.oldval -= r.bitmap
			value |= 1
		}
		if r.bitmap < 0x1000000 {
			r.bitmap <<= 8
			r.oldval = r.oldval<<8 | uint32(r.byte())
		}
	}
	return value
}

func (r *nsRange) decode(dst []byte, tre, alloc, first, inputSize uint32) bool {
	firstMask := (uint32(1) << first) - 1
	putMask := (uint32(1) << alloc) - 1
	var previous, unpacked, back, older, oldest, earliest, state, lastByte uint32
	back, older, oldest, earliest = 1, 1, 1, 1
	for !r.bad {
		context := firstMask & unpacked
		if r.bit((state<<4)+context) == 0 {
			shift := (8 - tre) & 0xff
			base := ((lastByte >> shift) + ((putMask & unpacked) << tre)) * 3 << 8
			if state >= 10 {
				state -= 6
			} else if state >= 4 {
				state -= 3
			} else {
				state = 0
			}
			if previous != 0 {
				if back > unpacked {
					return false
				}
				inputSize = inputSize&^0xff | uint32(dst[unpacked-back])
				lastByte = r.contextualByte(base+0x736, inputSize)
				previous = 0
			} else {
				lastByte = r.treeByte(base + 0x736)
			}
			if r.bad || unpacked >= uint32(len(dst)) {
				return false
			}
			dst[unpacked] = byte(lastByte)
			unpacked++
			if unpacked == uint32(len(dst)) {
				return !r.bad
			}
			continue
		}
		previous = 1
		if r.bit(state+0xc0) != 0 {
			if r.bit(state+0xcc) == 0 {
				at := ((state + 0xf) << 4) + context
				if r.bit(at) == 0 {
					if unpacked == 0 || back > unpacked || unpacked >= uint32(len(dst)) {
						return false
					}
					if state >= 7 {
						state = 11
					} else {
						state = 9
					}
					lastByte = uint32(dst[unpacked-back])
					dst[unpacked] = byte(lastByte)
					unpacked++
					if unpacked == uint32(len(dst)) {
						return !r.bad
					}
					continue
				}
				context = r.length(0x534, context)
				state = nsMatchState(state)
			} else {
				var selected uint32
				if r.bit(state+0xd8) == 0 {
					selected = older
				} else {
					if r.bit(state+0xe4) == 0 {
						selected = oldest
					} else {
						selected = earliest
						earliest = oldest
					}
					oldest = older
				}
				older, back = back, selected
				context = r.length(0x534, context)
				state = nsMatchState(state)
			}
		} else {
			earliest, oldest, older = oldest, older, back
			state = nsNewMatchState(state)
			context = r.length(0x332, context)
			positionContext := context
			if positionContext >= 4 {
				positionContext = 3
			}
			slot := r.treeBits(0x1b0+(positionContext<<6), 6)
			var value uint32
			if slot >= 4 {
				bits := (slot >> 1) - 1
				value = ((slot & 1) | 2) << bits
				if slot < 14 {
					value += r.reverseBits(value-slot+0x2af, bits)
				} else {
					value += r.directBits(bits-4)<<4 + r.reverseBits(0x322, 4)
				}
			} else {
				value = slot
			}
			back = value + 1
		}
		if r.bad || back == 0 || back > unpacked {
			return false
		}
		length := context + 2
		if length > uint32(len(dst))-unpacked {
			return false
		}
		for i := uint32(0); i < length; i++ {
			dst[unpacked] = dst[unpacked-back]
			unpacked++
		}
		lastByte = uint32(dst[unpacked-1])
		if unpacked == uint32(len(dst)) {
			return !r.bad
		}
	}
	return false
}

func nsMatchState(state uint32) uint32 {
	if state >= 7 {
		return 11
	}
	return 8
}

func nsNewMatchState(state uint32) uint32 {
	if state >= 7 {
		return 10
	}
	return 7
}
