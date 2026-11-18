package sevenzip

import "encoding/binary"

// The branch converters of the filters the reference decodes, x86 (BCJ and
// BCJ2) and ARM, as 7-Zip uses them (the same algorithms the xz file format
// specification describes), decoding a whole stream that starts at position
// 0. The reference skips members under the other branch filters (ARM Thumb,
// PowerPC, SPARC, IA-64, ARM64, RISC-V), as it skips Delta and BZip2.

func test86(b byte) bool { return b == 0 || b == 0xff }

func x86Decode(buf []byte) {
	if len(buf) < 5 {
		return
	}
	allowed := [8]bool{true, true, true, false, true, false, false, false}
	bitNumber := [8]uint32{0, 1, 2, 2, 3, 3, 3, 3}
	var prevMask uint32
	prevPos := ^uint32(4) // -5
	for i := 0; i <= len(buf)-5; {
		b := buf[i]
		if b != 0xe8 && b != 0xe9 {
			i++
			continue
		}
		offset := uint32(i) - prevPos
		prevPos = uint32(i)
		if offset > 5 {
			prevMask = 0
		} else {
			for k := uint32(0); k < offset; k++ {
				prevMask &= 0x77
				prevMask <<= 1
			}
		}
		b = buf[i+4]
		if test86(b) && allowed[prevMask>>1&7] && prevMask>>1 < 0x10 {
			src := uint32(b)<<24 | uint32(buf[i+3])<<16 | uint32(buf[i+2])<<8 | uint32(buf[i+1])
			var dest uint32
			for {
				dest = src - (uint32(i) + 5)
				if prevMask == 0 {
					break
				}
				n := bitNumber[prevMask>>1]
				b = byte(dest >> (24 - n*8))
				if !test86(b) {
					break
				}
				src = dest ^ (1<<(32-n*8) - 1)
			}
			buf[i+4] = ^byte((dest>>24)&1 - 1)
			buf[i+3] = byte(dest >> 16)
			buf[i+2] = byte(dest >> 8)
			buf[i+1] = byte(dest)
			i += 5
			prevMask = 0
		} else {
			i++
			prevMask |= 1
			if test86(b) {
				prevMask |= 0x10
			}
		}
	}
}

func armDecode(buf []byte) {
	for i := 0; i+4 <= len(buf); i += 4 {
		if buf[i+3] == 0xeb {
			src := (uint32(buf[i+2])<<16 | uint32(buf[i+1])<<8 | uint32(buf[i])) << 2
			dest := (src - (uint32(i) + 8)) >> 2
			buf[i+2], buf[i+1], buf[i] = byte(dest>>16), byte(dest>>8), byte(dest)
		}
	}
}

// bcj2Decode joins BCJ2's four streams: the main stream, the call and jump
// targets (big-endian, absolute) and the range-coded choices.
func bcj2Decode(main, call, jump, rc []byte, size uint64) ([]byte, error) {
	var p [258]uint16
	for i := range p {
		p[i] = lzmaModelTotal / 2
	}
	r := rangeDecoder{src: rc, rng: 0xffffffff}
	for i := 0; i < 5; i++ {
		r.code = r.code<<8 | uint32(r.next())
	}
	if r.bad {
		return nil, ErrData
	}
	out := make([]byte, 0, size)
	var prev byte
	in := 0
	for uint64(len(out)) < size {
		// Copy until a branch opcode.
		found := false
		for in < len(main) && uint64(len(out)) < size {
			b := main[in]
			out = append(out, b)
			in++
			if b&0xfe == 0xe8 || prev == 0x0f && b&0xf0 == 0x80 {
				found = true
				break
			}
			prev = b
		}
		if !found || uint64(len(out)) == size {
			break
		}
		b := out[len(out)-1]
		var prob *uint16
		switch {
		case b == 0xe8:
			prob = &p[prev]
		case b == 0xe9:
			prob = &p[256]
		default:
			prob = &p[257]
		}
		if r.bit(prob) == 0 {
			prev = b
			continue
		}
		var src []byte
		if b == 0xe8 {
			if len(call) < 4 {
				return nil, ErrData
			}
			src, call = call[:4], call[4:]
		} else {
			if len(jump) < 4 {
				return nil, ErrData
			}
			src, jump = jump[:4], jump[4:]
		}
		dest := binary.BigEndian.Uint32(src) - uint32(len(out)+4)
		for k := 0; k < 4 && uint64(len(out)) < size; k++ {
			out = append(out, byte(dest>>(8*k)))
		}
		prev = byte(dest >> 24)
		if r.bad {
			return nil, ErrData
		}
	}
	if uint64(len(out)) != size {
		return nil, ErrData
	}
	return out, nil
}
