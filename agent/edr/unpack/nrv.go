// Package unpack holds decompressors for packed Windows programs.
//
// The NRV decoders follow the stream formats and the reference scanner's
// failure rules, learned from its verdicts on crafted streams: a match that
// reaches before the output start fails, an offset prefix above 0x1000002
// fails as soon as it is read (values that would wrap around 32 bits
// included), output past the destination size fails, and so does a stream
// that needs more input than it has.
package unpack

import (
	"encoding/binary"
	"errors"
)

// Method is an NRV bit-stream format: three variants of one LZ77 scheme
// whose control bits come in little-endian 32-bit words, most significant
// bit first.
type Method uint8

const (
	NRV2B Method = iota
	NRV2D
	NRV2E
)

func (m Method) String() string {
	return [...]string{"NRV2B", "NRV2D", "NRV2E"}[m]
}

// Decoding failures.
var (
	ErrInputOverrun  = errors.New("nrv: input overrun")
	ErrOutputOverrun = errors.New("nrv: output overrun")
	ErrLookBehind    = errors.New("nrv: match before the output start")
	ErrOffset        = errors.New("nrv: match offset out of range")
)

// bits reads control bits from 32-bit little-endian words.
type bits struct {
	src  []byte
	pos  int    // next byte
	word uint32 // current control word
	left int    // bits left in word
}

func (b *bits) bit() (uint32, error) {
	if b.left == 0 {
		if b.pos+4 > len(b.src) {
			return 0, ErrInputOverrun
		}
		b.word = binary.LittleEndian.Uint32(b.src[b.pos:])
		b.pos += 4
		b.left = 32
	}
	b.left--
	return b.word >> uint(b.left) & 1, nil
}

func (b *bits) byte() (byte, error) {
	if b.pos >= len(b.src) {
		return 0, ErrInputOverrun
	}
	c := b.src[b.pos]
	b.pos++
	return c, nil
}

// maxPrefix is the largest offset prefix: 0x1000002 with low byte 0xff is
// the end marker.
const maxPrefix = 0xffffff + 3

// Decompress decodes src with the method into at most dstSize bytes. It
// returns the decoded bytes and the number of source bytes read, up to and
// including the end marker.
func Decompress(m Method, src []byte, dstSize int) ([]byte, int, error) {
	b := &bits{src: src}
	dst := make([]byte, 0, min(dstSize, 1<<20))
	var err error
	bit := func() uint32 {
		if err != nil {
			return 0
		}
		var v uint32
		v, err = b.bit()
		return v
	}
	lastOffset := uint32(1)
	for {
		// literals
		for bit() == 1 {
			var c byte
			if c, err = b.byte(); err != nil {
				return dst, b.pos, err
			}
			if len(dst) >= dstSize {
				return dst, b.pos, ErrOutputOverrun
			}
			dst = append(dst, c)
		}
		if err != nil {
			return dst, b.pos, err
		}
		// match offset: an Elias-gamma-like prefix, then a low byte
		offset := uint32(1)
		for {
			offset = offset*2 + bit()
			if offset > maxPrefix {
				return dst, b.pos, ErrOffset
			}
			if bit() == 1 || err != nil {
				break
			}
			if m != NRV2B {
				offset = (offset-1)*2 + bit()
				if offset > maxPrefix {
					return dst, b.pos, ErrOffset
				}
			}
		}
		if err != nil {
			return dst, b.pos, err
		}
		var length uint32
		if offset == 2 {
			offset = lastOffset
			if m != NRV2B {
				length = bit()
			}
		} else {
			c, e := b.byte()
			if e != nil {
				return dst, b.pos, e
			}
			offset = (offset-3)*256 + uint32(c)
			if offset == 0xffffffff {
				return dst, b.pos, nil // end marker
			}
			if m != NRV2B {
				length = (offset ^ 0xffffffff) & 1
				offset >>= 1
			}
			offset++
			lastOffset = offset
		}
		// match length
		switch m {
		case NRV2B:
			length = bit()
			length = length*2 + bit()
			if length == 0 {
				length = 1
				for {
					length = length*2 + bit()
					if bit() == 1 || err != nil {
						break
					}
				}
				length += 2
			}
			if offset > 0xd00 {
				length++
			}
		case NRV2D:
			length = length*2 + bit()
			if length == 0 {
				length = 1
				for {
					length = length*2 + bit()
					if bit() == 1 || err != nil {
						break
					}
				}
				length += 2
			}
			if offset > 0x500 {
				length++
			}
		case NRV2E:
			switch {
			case length != 0:
				length = 1 + bit()
			case bit() == 1:
				length = 3 + bit()
			default:
				length = 1
				for {
					length = length*2 + bit()
					if bit() == 1 || err != nil {
						break
					}
				}
				length += 3
			}
			if offset > 0x500 {
				length++
			}
		}
		if err != nil {
			return dst, b.pos, err
		}
		if uint64(offset) > uint64(len(dst)) {
			return dst, b.pos, ErrLookBehind
		}
		if uint64(len(dst))+uint64(length)+1 > uint64(dstSize) {
			return dst, b.pos, ErrOutputOverrun
		}
		from := len(dst) - int(offset)
		for i := 0; i <= int(length); i++ {
			dst = append(dst, dst[from+i])
		}
	}
}
