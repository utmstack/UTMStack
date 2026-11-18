package sevenzip

// LZMA and LZMA2 decoding after the public LZMA specification and the LZMA2
// chunk format (xz file format specification): a range coder with 11-bit
// probabilities, literal, match and repeat coding, and LZMA2's chunks of
// stored or LZMA data that reset the dictionary, the state or the
// properties.

const (
	lzmaTop        = 1 << 24
	lzmaModelBits  = 11
	lzmaModelTotal = 1 << lzmaModelBits
	lzmaMoveBits   = 5
)

type rangeDecoder struct {
	src       []byte
	pos       int
	rng, code uint32
	bad       bool
}

func (r *rangeDecoder) init(src []byte) bool {
	r.src, r.pos, r.rng, r.code, r.bad = src, 0, 0xffffffff, 0, false
	if r.next() != 0 {
		return false
	}
	for i := 0; i < 4; i++ {
		r.code = r.code<<8 | uint32(r.next())
	}
	return !r.bad
}

func (r *rangeDecoder) next() byte {
	if r.pos >= len(r.src) {
		r.bad = true
		return 0
	}
	b := r.src[r.pos]
	r.pos++
	return b
}

func (r *rangeDecoder) bit(p *uint16) uint32 {
	bound := (r.rng >> lzmaModelBits) * uint32(*p)
	var b uint32
	if r.code < bound {
		r.rng = bound
		*p += (lzmaModelTotal - *p) >> lzmaMoveBits
	} else {
		r.rng -= bound
		r.code -= bound
		*p -= *p >> lzmaMoveBits
		b = 1
	}
	if r.rng < lzmaTop {
		r.rng <<= 8
		r.code = r.code<<8 | uint32(r.next())
	}
	return b
}

func (r *rangeDecoder) direct(n int) uint32 {
	var v uint32
	for ; n > 0; n-- {
		r.rng >>= 1
		var b uint32
		if r.code >= r.rng {
			r.code -= r.rng
			b = 1
		}
		v = v<<1 | b
		if r.rng < lzmaTop {
			r.rng <<= 8
			r.code = r.code<<8 | uint32(r.next())
		}
	}
	return v
}

func (r *rangeDecoder) tree(p []uint16, n int) uint32 {
	m := uint32(1)
	for i := 0; i < n; i++ {
		m = m<<1 | r.bit(&p[m])
	}
	return m - 1<<uint(n)
}

func (r *rangeDecoder) reverse(p []uint16, n int) uint32 {
	m, v := uint32(1), uint32(0)
	for i := 0; i < n; i++ {
		b := r.bit(&p[m])
		m = m<<1 | b
		v |= b << uint(i)
	}
	return v
}

func probs(n int) []uint16 {
	p := make([]uint16, n)
	for i := range p {
		p[i] = lzmaModelTotal / 2
	}
	return p
}

type lengthCoder struct {
	choice, choice2 uint16
	low, mid        [16][]uint16
	high            []uint16
}

func newLengthCoder() *lengthCoder {
	l := &lengthCoder{choice: lzmaModelTotal / 2, choice2: lzmaModelTotal / 2, high: probs(256)}
	for i := range l.low {
		l.low[i], l.mid[i] = probs(8), probs(8)
	}
	return l
}

func (l *lengthCoder) decode(r *rangeDecoder, posState int) uint32 {
	if r.bit(&l.choice) == 0 {
		return r.tree(l.low[posState], 3)
	}
	if r.bit(&l.choice2) == 0 {
		return 8 + r.tree(l.mid[posState], 3)
	}
	return 16 + r.tree(l.high, 8)
}

// lzmaDecoder keeps the state LZMA2 chunks share; its output buffer is the
// dictionary.
type lzmaDecoder struct {
	lc, lp, pb              int
	literals                []uint16
	isMatch, isRep0Long     [12 << 4]uint16
	isRep, isRepG0, isRepG1 [12]uint16
	isRepG2                 [12]uint16
	slots                   [4][]uint16
	special, align          []uint16
	lengths, repLengths     *lengthCoder
	state                   int
	rep                     [4]uint32
	out                     []byte
	dictStart               int // matches may not reach before it
	done                    <-chan struct{}
	steps                   int
}

func (d *lzmaDecoder) setProps(b byte) bool {
	if b >= 9*5*5 {
		return false
	}
	d.lc, d.lp, d.pb = int(b%9), int(b/9%5), int(b/45)
	return true
}

func (d *lzmaDecoder) reset() {
	d.literals = probs(0x300 << uint(d.lc+d.lp))
	for _, p := range [][]uint16{d.isMatch[:], d.isRep0Long[:], d.isRep[:], d.isRepG0[:], d.isRepG1[:], d.isRepG2[:]} {
		for i := range p {
			p[i] = lzmaModelTotal / 2
		}
	}
	for i := range d.slots {
		d.slots[i] = probs(64)
	}
	d.special, d.align = probs(115), probs(16)
	d.lengths, d.repLengths = newLengthCoder(), newLengthCoder()
	d.state = 0
	d.rep = [4]uint32{}
}

// run decodes until the output holds end bytes; it fails on a match before
// the dictionary's start, on the end marker and on missing input.
func (d *lzmaDecoder) run(r *rangeDecoder, end int) error {
	pbMask, lpMask := 1<<uint(d.pb)-1, 1<<uint(d.lp)-1
	for len(d.out) < end {
		if d.steps++; d.steps&0xffff == 0 && stopped(d.done) {
			return ErrAborted
		}
		pos := len(d.out) - d.dictStart
		posState := pos & pbMask
		if r.bit(&d.isMatch[d.state<<4+posState]) == 0 {
			prev := 0
			if len(d.out) > d.dictStart {
				prev = int(d.out[len(d.out)-1])
			}
			p := d.literals[0x300*((pos&lpMask)<<uint(d.lc)+prev>>uint(8-d.lc)):]
			sym := uint32(1)
			if d.state >= 7 {
				at := len(d.out) - 1 - int(d.rep[0])
				if at < d.dictStart {
					return ErrData
				}
				match := uint32(d.out[at])
				for sym < 0x100 {
					mb := match >> 7 & 1
					match <<= 1
					b := r.bit(&p[(1+mb)<<8+sym])
					sym = sym<<1 | b
					if mb != b {
						break
					}
				}
			}
			for sym < 0x100 {
				sym = sym<<1 | r.bit(&p[sym])
			}
			d.out = append(d.out, byte(sym))
			switch {
			case d.state < 4:
				d.state = 0
			case d.state < 10:
				d.state -= 3
			default:
				d.state -= 6
			}
		} else {
			var length uint32
			if r.bit(&d.isRep[d.state]) == 1 {
				if len(d.out) == d.dictStart {
					return ErrData
				}
				if r.bit(&d.isRepG0[d.state]) == 0 {
					if r.bit(&d.isRep0Long[d.state<<4+posState]) == 0 {
						if d.state < 7 {
							d.state = 9
						} else {
							d.state = 11
						}
						at := len(d.out) - 1 - int(d.rep[0])
						if at < d.dictStart {
							return ErrData
						}
						d.out = append(d.out, d.out[at])
						if r.bad {
							return ErrData
						}
						continue
					}
				} else {
					var dist uint32
					if r.bit(&d.isRepG1[d.state]) == 0 {
						dist = d.rep[1]
					} else {
						if r.bit(&d.isRepG2[d.state]) == 0 {
							dist = d.rep[2]
						} else {
							dist = d.rep[3]
							d.rep[3] = d.rep[2]
						}
						d.rep[2] = d.rep[1]
					}
					d.rep[1] = d.rep[0]
					d.rep[0] = dist
				}
				length = d.repLengths.decode(r, posState)
				if d.state < 7 {
					d.state = 8
				} else {
					d.state = 11
				}
			} else {
				d.rep[3], d.rep[2], d.rep[1] = d.rep[2], d.rep[1], d.rep[0]
				length = d.lengths.decode(r, posState)
				if d.state < 7 {
					d.state = 7
				} else {
					d.state = 10
				}
				slot := r.tree(d.slots[min(length, 3)], 6)
				dist := slot
				if slot >= 4 {
					n := int(slot>>1) - 1
					dist = (2 | slot&1) << uint(n)
					if slot < 14 {
						dist += r.reverse(d.special[dist-slot:], n)
					} else {
						dist += r.direct(n-4) << 4
						dist += r.reverse(d.align, 4)
					}
				}
				if dist == 0xffffffff {
					return ErrData // the end marker
				}
				d.rep[0] = dist
			}
			at := len(d.out) - 1 - int(d.rep[0])
			if at < d.dictStart || int(d.rep[0]) < 0 {
				return ErrData
			}
			n := int(length) + 2
			for i := 0; i < n && len(d.out) < end; i++ {
				d.out = append(d.out, d.out[len(d.out)-1-int(d.rep[0])])
			}
		}
		if r.bad {
			return ErrData
		}
	}
	return nil
}

// decodeLZMA decodes a 7z LZMA stream (five property bytes) to size bytes.
func decodeLZMA(props, src []byte, size uint64, done <-chan struct{}) ([]byte, error) {
	if len(props) < 5 {
		return nil, ErrData
	}
	d := &lzmaDecoder{done: done}
	if !d.setProps(props[0]) {
		return nil, ErrData
	}
	d.reset()
	d.out = make([]byte, 0, size)
	var r rangeDecoder
	if !r.init(src) {
		return nil, ErrData
	}
	if err := d.run(&r, int(size)); err != nil {
		return nil, err
	}
	return d.out, nil
}

// decodeLZMA2 decodes a 7z LZMA2 stream to size bytes.
func decodeLZMA2(props, src []byte, size uint64, done <-chan struct{}) ([]byte, error) {
	if len(props) < 1 || props[0] > 40 {
		return nil, ErrData
	}
	d := &lzmaDecoder{out: make([]byte, 0, size), done: done}
	pos, needDict, needProps := 0, true, true
	for {
		if pos >= len(src) {
			return nil, ErrData
		}
		c := src[pos]
		pos++
		if c == 0 {
			break
		}
		if c == 1 || c == 2 {
			if c == 2 && needDict || pos+2 > len(src) {
				return nil, ErrData
			}
			if c == 1 {
				d.dictStart = len(d.out)
				needDict = false
			}
			n := int(src[pos])<<8 | int(src[pos+1]) + 1
			pos += 2
			if pos+n > len(src) || uint64(len(d.out)+n) > size {
				return nil, ErrData
			}
			d.out = append(d.out, src[pos:pos+n]...)
			pos += n
			continue
		}
		if c < 0x80 || pos+4 > len(src) {
			return nil, ErrData
		}
		unpacked := int(c&0x1f)<<16 | int(src[pos])<<8 | int(src[pos+1]) + 1
		packed := int(src[pos+2])<<8 | int(src[pos+3]) + 1
		pos += 4
		reset := c >> 5 & 3
		if reset == 3 {
			d.dictStart = len(d.out)
			needDict = false
		} else if needDict {
			return nil, ErrData
		}
		if reset >= 2 {
			if pos >= len(src) || !d.setProps(src[pos]) || d.lc+d.lp > 4 {
				return nil, ErrData
			}
			pos++
			needProps = false
		} else if needProps {
			return nil, ErrData
		}
		if reset >= 1 {
			d.reset()
		}
		if pos+packed > len(src) || uint64(len(d.out)+unpacked) > size {
			return nil, ErrData
		}
		var r rangeDecoder
		if !r.init(src[pos : pos+packed]) {
			return nil, ErrData
		}
		if err := d.run(&r, len(d.out)+unpacked); err != nil {
			return nil, err
		}
		pos += packed
	}
	if uint64(len(d.out)) != size {
		return nil, ErrData
	}
	return d.out, nil
}
