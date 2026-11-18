// Package lzx decodes Microsoft LZX data as cabinet files and compiled HTML
// help (CHM) store it, after the public format descriptions (the LZX data
// compression format of the Cabinet SDK; [MS-PATCH], which describes the same
// coding): a bit stream of 16-bit little-endian words read from the most
// significant bit, verbatim, aligned-offset and uncompressed blocks, main,
// length and aligned trees coded through a pretree and kept from block to
// block as deltas, three repeated offsets, 32 KiB frames realigned to 16 bits,
// and the E8 call translation of the frames' output.
package lzx

import "errors"

var (
	// ErrData: the data cannot be decoded.
	ErrData = errors.New("lzx: corrupt data")
	// ErrTruncated: the data ends before the output does.
	ErrTruncated = errors.New("lzx: data ends early")
)

const (
	frameSize   = 32768
	numChars    = 256
	minMatch    = 2
	lengthCodes = 249
	pretreeSize = 20
	alignedSize = 8
	safety      = 64
)

var (
	positionBase [51]uint32
	extraBits    [51]uint8
	slotsFor     = map[int]int{15: 30, 16: 32, 17: 34, 18: 36, 19: 38, 20: 42, 21: 50}
)

func init() {
	// Extra bits: 0, 0, 0, 0, 1, 1, 2, 2, ... up to 17, two slots each.
	for i, j := 0, uint8(0); i < 51; i += 2 {
		extraBits[i] = j
		if i+1 < 51 {
			extraBits[i+1] = j
		}
		if i != 0 && j < 17 {
			j++
		}
	}
	for i, j := 0, uint32(0); i < 51; i++ {
		positionBase[i] = j
		j += 1 << extraBits[i]
	}
}

// bits reads 16-bit little-endian words, most significant bit first. Past
// the end of the data it supplies zero words once, as the reference's
// reader pads, and then fails.
type bits struct {
	src    []byte
	pos    int
	buf    uint32
	left   uint
	padded bool
	over   bool
}

func (b *bits) ensure(n uint) bool {
	for b.left < n {
		var w uint32
		switch {
		case b.pos+1 < len(b.src):
			w = uint32(b.src[b.pos]) | uint32(b.src[b.pos+1])<<8
			b.pos += 2
		case !b.padded:
			if b.pos < len(b.src) { // a last odd byte
				w = uint32(b.src[b.pos])
				b.pos++
			}
			b.padded = true
		default:
			b.over = true
			return false
		}
		b.buf |= w << (16 - b.left)
		b.left += 16
	}
	return true
}

func (b *bits) read(n uint) (uint32, bool) {
	if n == 0 {
		return 0, true
	}
	if !b.ensure(n) {
		return 0, false
	}
	v := b.buf >> (32 - n)
	b.buf <<= n
	b.left -= n
	return v, true
}

// tree is a canonical Huffman code read most significant bit first.
type tree struct {
	count  [17]uint16
	symbol []uint16
	empty  bool
}

// build makes a code from lengths. As the reference's table builder does,
// it fails on over-subscribed and incomplete sets, except a set with no
// codes at all, which only fails when read.
func build(lengths []uint8) (*tree, bool) {
	t := &tree{symbol: make([]uint16, len(lengths))}
	for _, l := range lengths {
		t.count[l]++
	}
	if int(t.count[0]) == len(lengths) {
		t.empty = true
		return t, true
	}
	left := 1
	for l := 1; l <= 16; l++ {
		left = left<<1 - int(t.count[l])
		if left < 0 {
			return nil, false
		}
	}
	if left != 0 {
		return nil, false
	}
	var offs [18]uint16
	for l := 1; l <= 16; l++ {
		offs[l+1] = offs[l] + t.count[l]
	}
	for s, l := range lengths {
		if l != 0 {
			t.symbol[offs[l]] = uint16(s)
			offs[l]++
		}
	}
	return t, true
}

func (t *tree) decode(b *bits) (int, error) {
	if t.empty {
		return 0, ErrData
	}
	code, first, index := 0, 0, 0
	for l := 1; l <= 16; l++ {
		v, ok := b.read(1)
		if !ok {
			return 0, ErrTruncated
		}
		code |= int(v)
		count := int(t.count[l])
		if code-count < first {
			return int(t.symbol[index+code-first]), nil
		}
		index += count
		first = (first + count) << 1
		code <<= 1
	}
	return 0, ErrData
}

// Decoder holds the state LZX keeps between frames.
type Decoder struct {
	windowBits, slots   int
	r0, r1, r2          uint32
	main, length        []uint8
	aligned             [alignedSize]uint8
	mainTree, lenTree   *tree
	alignedTree         *tree
	blockType           int
	blockLength         int
	blockRemaining      int
	headerRead          bool
	intelSize           int32
	intelStarted        bool
	resetFrames, frames int
	in                  bits
	out                 []byte
}

// NewDecoder starts decoding src with a window of 1<<windowBits bytes (15 to
// 21); resetFrames, when not zero, resets the state every that many frames,
// as CHM files do.
func NewDecoder(src []byte, windowBits, resetFrames int) (*Decoder, error) {
	slots, ok := slotsFor[windowBits]
	if !ok {
		return nil, ErrData
	}
	d := &Decoder{windowBits: windowBits, slots: slots, resetFrames: resetFrames, in: bits{src: src}}
	d.reset()
	return d, nil
}

func (d *Decoder) reset() {
	d.r0, d.r1, d.r2 = 1, 1, 1
	d.main = make([]uint8, numChars+d.slots*8+safety)
	d.length = make([]uint8, lengthCodes+safety)
	d.headerRead = false
	d.blockRemaining, d.blockType = 0, 0
}

// Decode decodes outLen bytes in all (frames of 32 KiB, the last one
// shorter) and returns them; on an error the frames decoded so far are
// returned too.
func Decode(src []byte, windowBits, outLen, resetFrames int) ([]byte, error) {
	d, err := NewDecoder(src, windowBits, resetFrames)
	if err != nil {
		return nil, err
	}
	for len(d.out) < outLen {
		if err := d.frame(min(frameSize, outLen-len(d.out))); err != nil {
			return d.out, err
		}
	}
	return d.out, nil
}

// readLengths reads tree lengths [first, last) through a pretree, each a
// delta from the length the block before left.
func (d *Decoder) readLengths(lens []uint8, first, last int) error {
	var pl [pretreeSize]uint8
	for i := range pl {
		v, ok := d.in.read(4)
		if !ok {
			return ErrTruncated
		}
		pl[i] = uint8(v)
	}
	pre, ok := build(pl[:])
	if !ok {
		return ErrData
	}
	for x := first; x < last; {
		z, err := pre.decode(&d.in)
		if err != nil {
			return err
		}
		switch z {
		case 17, 18:
			n, base := uint(4), 4
			if z == 18 {
				n, base = 5, 20
			}
			y, ok := d.in.read(n)
			if !ok {
				return ErrTruncated
			}
			for i := 0; i < int(y)+base; i++ {
				if x >= len(lens) {
					return ErrData
				}
				lens[x] = 0
				x++
			}
		case 19:
			y, ok := d.in.read(1)
			if !ok {
				return ErrTruncated
			}
			z, err := pre.decode(&d.in)
			if err != nil {
				return err
			}
			if z > 16 {
				return ErrData
			}
			if x >= len(lens) {
				return ErrData
			}
			v := (int(lens[x]) - z + 17) % 17
			for i := 0; i < int(y)+4; i++ {
				if x >= len(lens) {
					return ErrData
				}
				lens[x] = uint8(v)
				x++
			}
		default:
			lens[x] = uint8((int(lens[x]) - z + 17) % 17)
			x++
		}
	}
	return nil
}

func (d *Decoder) readBlockHeader() error {
	if d.blockType == 3 && d.blockLength&1 != 0 {
		// The padding byte after an odd uncompressed block.
		if d.in.pos >= len(d.in.src) {
			return ErrTruncated
		}
		d.in.pos++
	}
	t, ok1 := d.in.read(3)
	hi, ok2 := d.in.read(16)
	lo, ok3 := d.in.read(8)
	if !ok1 || !ok2 || !ok3 {
		return ErrTruncated
	}
	d.blockType = int(t)
	d.blockLength = int(hi)<<8 | int(lo)
	d.blockRemaining = d.blockLength
	switch d.blockType {
	case 2:
		for i := range d.aligned {
			v, ok := d.in.read(3)
			if !ok {
				return ErrTruncated
			}
			d.aligned[i] = uint8(v)
		}
		var ok bool
		if d.alignedTree, ok = build(d.aligned[:]); !ok {
			return ErrData
		}
		fallthrough
	case 1:
		mainCount := numChars + d.slots*8
		if err := d.readLengths(d.main, 0, numChars); err != nil {
			return err
		}
		if err := d.readLengths(d.main, numChars, mainCount); err != nil {
			return err
		}
		var ok bool
		if d.mainTree, ok = build(d.main[:mainCount]); !ok {
			return ErrData
		}
		if d.main[0xe8] != 0 {
			d.intelStarted = true
		}
		if err := d.readLengths(d.length, 0, lengthCodes); err != nil {
			return err
		}
		if d.lenTree, ok = build(d.length[:lengthCodes]); !ok {
			return ErrData
		}
	case 3:
		d.intelStarted = true
		// Realign: drop the bits buffered in the current 16-bit word.
		if !d.in.ensure(16) {
			return ErrTruncated
		}
		if d.in.left > 16 {
			d.in.pos -= 2
		}
		d.in.left, d.in.buf = 0, 0
		if d.in.pos+12 > len(d.in.src) {
			return ErrTruncated
		}
		s := d.in.src[d.in.pos:]
		d.r0 = uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
		d.r1 = uint32(s[4]) | uint32(s[5])<<8 | uint32(s[6])<<16 | uint32(s[7])<<24
		d.r2 = uint32(s[8]) | uint32(s[9])<<8 | uint32(s[10])<<16 | uint32(s[11])<<24
		d.in.pos += 12
	default:
		return ErrData
	}
	return nil
}

// frame decodes the next frame of size bytes.
func (d *Decoder) frame(size int) error {
	if d.resetFrames > 0 && d.frames > 0 && d.frames%d.resetFrames == 0 {
		d.reset()
	}
	if !d.headerRead {
		d.headerRead = true
		v, ok := d.in.read(1)
		if !ok {
			return ErrTruncated
		}
		if v != 0 {
			hi, ok1 := d.in.read(16)
			lo, ok2 := d.in.read(16)
			if !ok1 || !ok2 {
				return ErrTruncated
			}
			d.intelSize = int32(hi<<16 | lo)
		} else {
			d.intelSize = 0
		}
	}
	start := len(d.out)
	window := 1 << d.windowBits
	for todo := size; todo > 0; {
		if d.blockRemaining == 0 {
			if err := d.readBlockHeader(); err != nil {
				return err
			}
		}
		run := min(d.blockRemaining, todo)
		done := 0
		if d.blockType == 3 {
			if d.in.pos+run > len(d.in.src) {
				d.out = append(d.out, d.in.src[d.in.pos:]...)
				return ErrTruncated
			}
			d.out = append(d.out, d.in.src[d.in.pos:d.in.pos+run]...)
			d.in.pos += run
			done = run
		}
		for done < run {
			m, err := d.mainTree.decode(&d.in)
			if err != nil {
				return err
			}
			if m < numChars {
				d.out = append(d.out, byte(m))
				done++
				continue
			}
			m -= numChars
			length := m & 7
			if length == 7 {
				l, err := d.lenTree.decode(&d.in)
				if err != nil {
					return err
				}
				length += l
			}
			length += minMatch
			slot := m >> 3
			var offset uint32
			switch slot {
			case 0:
				offset = d.r0
			case 1:
				offset = d.r1
				d.r1 = d.r0
				d.r0 = offset
			case 2:
				offset = d.r2
				d.r2 = d.r0
				d.r0 = offset
			default:
				extra := uint(extraBits[slot])
				offset = positionBase[slot] - 2
				if d.blockType == 2 && extra >= 3 {
					v, ok := d.in.read(extra - 3)
					if !ok {
						return ErrTruncated
					}
					a, err := d.alignedTree.decode(&d.in)
					if err != nil {
						return err
					}
					offset += v<<3 + uint32(a)
				} else {
					v, ok := d.in.read(extra)
					if !ok {
						return ErrTruncated
					}
					offset += v
				}
				d.r2, d.r1, d.r0 = d.r1, d.r0, offset
			}
			if offset == 0 || int(offset) > len(d.out) || int(offset) > window {
				return ErrData // before the stream's start
			}
			if done+length > run && length-(run-done) > d.blockRemaining-run {
				return ErrData // past the end of the block
			}
			for i := 0; i < length; i++ {
				d.out = append(d.out, d.out[len(d.out)-int(offset)])
			}
			done += length
		}
		if done > run {
			d.blockRemaining -= done
			todo -= done
		} else {
			d.blockRemaining -= run
			todo -= run
		}
		if d.in.over {
			return ErrTruncated
		}
	}
	if len(d.out)-start != size {
		return ErrData // a match ran past the frame's end
	}
	// Realign to 16 bits at the frame's end.
	if d.in.left > 0 {
		d.in.ensure(16)
		if d.in.left&15 != 0 {
			d.in.read(d.in.left & 15)
		}
	}
	if d.intelStarted && d.intelSize != 0 && d.frames < 32768 && size > 10 {
		translateE8(d.out[start:], int32(start), d.intelSize)
	}
	d.frames++
	return nil
}

// translateE8 undoes the E8 call translation over a frame: each E8 byte
// before the last 10 is followed by an absolute address that is made
// relative to its position again when it lies in the translated range.
func translateE8(data []byte, curpos, size int32) {
	end := len(data) - 10
	for i := 0; i < end; {
		if data[i] != 0xe8 {
			i++
			curpos++
			continue
		}
		abs := int32(uint32(data[i+1]) | uint32(data[i+2])<<8 | uint32(data[i+3])<<16 | uint32(data[i+4])<<24)
		if abs >= -curpos && abs < size {
			rel := abs + size
			if abs >= 0 {
				rel = abs - curpos
			}
			data[i+1], data[i+2], data[i+3], data[i+4] = byte(rel), byte(rel>>8), byte(rel>>16), byte(rel>>24)
		}
		i += 5
		curpos += 5
	}
}
