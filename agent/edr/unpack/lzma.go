package unpack

import "errors"

// An LZMA decoder for raw streams of a known size, as packers store them:
// no header and no end marker. It follows the public LZMA specification
// (range coder, literal, match and repeat coding) and fails where that
// specification's reference decoder fails: a first byte other than zero, a
// match reaching before the output start, a stream that needs more input
// than it has, and an end marker before the size is reached.

var (
	ErrLZMAData  = errors.New("lzma: corrupt data")
	ErrLZMAInput = errors.New("lzma: input overrun")
)

// LZMAProps are a stream's literal context bits, literal position bits and
// position bits.
type LZMAProps struct{ LC, LP, PB int }

const (
	lzmaBitModelTotal = 1 << 11
	lzmaMoveBits      = 5
	lzmaTopValue      = 1 << 24
)

type lzmaRange struct {
	src       []byte
	pos       int
	rng, code uint32
	overrun   bool
}

func (r *lzmaRange) next() byte {
	if r.pos >= len(r.src) {
		r.overrun = true
		return 0
	}
	b := r.src[r.pos]
	r.pos++
	return b
}

func (r *lzmaRange) normalize() {
	if r.rng < lzmaTopValue {
		r.rng <<= 8
		r.code = r.code<<8 | uint32(r.next())
	}
}

func (r *lzmaRange) bit(p *uint16) uint32 {
	bound := (r.rng >> 11) * uint32(*p)
	var b uint32
	if r.code < bound {
		r.rng = bound
		*p += (lzmaBitModelTotal - *p) >> lzmaMoveBits
	} else {
		r.rng -= bound
		r.code -= bound
		*p -= *p >> lzmaMoveBits
		b = 1
	}
	r.normalize()
	return b
}

func (r *lzmaRange) direct(n int) uint32 {
	var v uint32
	for ; n > 0; n-- {
		r.rng >>= 1
		r.code -= r.rng
		t := 0 - (r.code >> 31)
		r.code += r.rng & t
		v = v<<1 + t + 1
		r.normalize()
	}
	return v
}

func lzmaProbs(n int) []uint16 {
	p := make([]uint16, n)
	for i := range p {
		p[i] = lzmaBitModelTotal / 2
	}
	return p
}

// lzmaTree decodes n bits most significant first over a bit tree.
func lzmaTree(r *lzmaRange, probs []uint16, n int) uint32 {
	m := uint32(1)
	for i := 0; i < n; i++ {
		m = m<<1 + r.bit(&probs[m])
	}
	return m - 1<<uint(n)
}

// lzmaReverse decodes n bits least significant first over a bit tree.
func lzmaReverse(r *lzmaRange, probs []uint16, n int) uint32 {
	m, v := uint32(1), uint32(0)
	for i := 0; i < n; i++ {
		b := r.bit(&probs[m])
		m = m<<1 + b
		v |= b << uint(i)
	}
	return v
}

type lzmaLength struct {
	choice, choice2 uint16
	low, mid        [16][]uint16
	high            []uint16
}

func newLZMALength() *lzmaLength {
	l := &lzmaLength{choice: lzmaBitModelTotal / 2, choice2: lzmaBitModelTotal / 2, high: lzmaProbs(256)}
	for i := range l.low {
		l.low[i], l.mid[i] = lzmaProbs(8), lzmaProbs(8)
	}
	return l
}

func (l *lzmaLength) decode(r *lzmaRange, posState int) uint32 {
	if r.bit(&l.choice) == 0 {
		return lzmaTree(r, l.low[posState], 3)
	}
	if r.bit(&l.choice2) == 0 {
		return 8 + lzmaTree(r, l.mid[posState], 3)
	}
	return 16 + lzmaTree(r, l.high, 8)
}

// DecompressLZMA decodes size bytes of a raw LZMA stream.
func DecompressLZMA(src []byte, props LZMAProps, size int) ([]byte, error) {
	if props.LC < 0 || props.LC > 8 || props.LP < 0 || props.LP > 4 || props.PB < 0 || props.PB > 4 {
		return nil, ErrLZMAData
	}
	r := &lzmaRange{src: src, rng: 0xffffffff}
	if r.next() != 0 {
		return nil, ErrLZMAData
	}
	for i := 0; i < 4; i++ {
		r.code = r.code<<8 | uint32(r.next())
	}
	if r.overrun {
		return nil, ErrLZMAInput
	}
	out := make([]byte, 0, size)
	literals := lzmaProbs(0x300 << uint(props.LC+props.LP))
	var isMatch, isRep0Long [12 << 4]uint16
	var isRep, isRepG0, isRepG1, isRepG2 [12]uint16
	for _, p := range [][]uint16{isMatch[:], isRep0Long[:], isRep[:], isRepG0[:], isRepG1[:], isRepG2[:]} {
		for i := range p {
			p[i] = lzmaBitModelTotal / 2
		}
	}
	var slots [4][]uint16
	for i := range slots {
		slots[i] = lzmaProbs(64)
	}
	special := lzmaProbs(115) // distance bits of slots 4 to 13
	align := lzmaProbs(16)
	lengths, repLengths := newLZMALength(), newLZMALength()
	pbMask, lpMask := 1<<uint(props.PB)-1, 1<<uint(props.LP)-1
	state := 0
	var rep [4]uint32
	for len(out) < size {
		posState := len(out) & pbMask
		if r.bit(&isMatch[state<<4+posState]) == 0 {
			prev := 0
			if len(out) > 0 {
				prev = int(out[len(out)-1])
			}
			probs := literals[0x300*((len(out)&lpMask)<<uint(props.LC)+prev>>uint(8-props.LC)):]
			sym := uint32(1)
			if state >= 7 {
				if uint64(rep[0]) >= uint64(len(out)) {
					return out, ErrLZMAData
				}
				match := uint32(out[len(out)-1-int(rep[0])])
				for sym < 0x100 {
					mb := match >> 7 & 1
					match <<= 1
					b := r.bit(&probs[(1+mb)<<8+sym])
					sym = sym<<1 | b
					if mb != b {
						break
					}
				}
			}
			for sym < 0x100 {
				sym = sym<<1 | r.bit(&probs[sym])
			}
			out = append(out, byte(sym))
			switch {
			case state < 4:
				state = 0
			case state < 10:
				state -= 3
			default:
				state -= 6
			}
		} else {
			var length uint32
			if r.bit(&isRep[state]) == 1 {
				if len(out) == 0 {
					return out, ErrLZMAData
				}
				if r.bit(&isRepG0[state]) == 0 {
					if r.bit(&isRep0Long[state<<4+posState]) == 0 {
						if state < 7 {
							state = 9
						} else {
							state = 11
						}
						if uint64(rep[0]) >= uint64(len(out)) {
							return out, ErrLZMAData
						}
						out = append(out, out[len(out)-1-int(rep[0])])
						if r.overrun {
							return out, ErrLZMAInput
						}
						continue
					}
				} else {
					var dist uint32
					if r.bit(&isRepG1[state]) == 0 {
						dist = rep[1]
					} else {
						if r.bit(&isRepG2[state]) == 0 {
							dist = rep[2]
						} else {
							dist = rep[3]
							rep[3] = rep[2]
						}
						rep[2] = rep[1]
					}
					rep[1] = rep[0]
					rep[0] = dist
				}
				length = repLengths.decode(r, posState)
				if state < 7 {
					state = 8
				} else {
					state = 11
				}
			} else {
				rep[3], rep[2], rep[1] = rep[2], rep[1], rep[0]
				length = lengths.decode(r, posState)
				if state < 7 {
					state = 7
				} else {
					state = 10
				}
				lenState := min(length, 3)
				slot := lzmaTree(r, slots[lenState], 6)
				dist := slot
				if slot >= 4 {
					n := int(slot>>1) - 1
					dist = (2 | slot&1) << uint(n)
					if slot < 14 {
						dist += lzmaReverse(r, special[dist-slot:], n)
					} else {
						dist += r.direct(n-4) << 4
						dist += lzmaReverse(r, align, 4)
					}
				}
				if dist == 0xffffffff {
					// the end marker
					return out, ErrLZMAData
				}
				rep[0] = dist
			}
			if uint64(rep[0]) >= uint64(len(out)) {
				return out, ErrLZMAData
			}
			n := int(length) + 2
			for i := 0; i < n && len(out) < size; i++ {
				out = append(out, out[len(out)-1-int(rep[0])])
			}
		}
		if r.overrun {
			return out, ErrLZMAInput
		}
	}
	return out, nil
}
