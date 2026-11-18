package engine

import (
	"bytes"
	"context"
	"encoding/binary"
)

// lhaHeader recognizes the level-0/1/2 LHA prefix and the distinct level-3
// word-size marker.
func lhaHeader(data []byte) bool {
	if len(data) < 24 || data[2] != '-' || (data[3] != 'l' && data[3] != 'L') || data[6] != '-' {
		return false
	}
	switch data[20] {
	case 0, 1:
		return data[0] >= 22
	case 2:
		return len(data) >= 26 && binary.LittleEndian.Uint16(data[:2]) >= 26
	case 3:
		return len(data) >= 32 && binary.LittleEndian.Uint16(data[:2]) == 4
	}
	return false
}

type lhaMember struct {
	start, packed uint64
	plain         uint64
	crc           uint16
	method        [5]byte
	badHeader     bool
}

// lhaMemberOf validates legacy base headers and level-2/3 total headers.
// Level-1 packed size includes all extended headers; level-2/3 packed size
// excludes them. Level 3 uses four-byte extended-header lengths.
func lhaMemberOf(h []byte) (lhaMember, error) {
	var m lhaMember
	if !lhaHeader(h) {
		return m, ErrMalformed
	}
	copy(m.method[:], h[2:7])
	m.plain = uint64(binary.LittleEndian.Uint32(h[11:15]))
	var baseEnd, next int
	switch h[20] {
	case 0, 1:
		baseEnd = 2 + int(h[0])
		nameSize := int(h[21])
		crcOff := 22 + nameSize
		minimum := crcOff + 2
		if h[20] == 1 {
			minimum += 3 // OS ID and first extended-header size
		}
		if baseEnd > len(h) || baseEnd < minimum {
			return m, ErrMalformed
		}
		m.crc = binary.LittleEndian.Uint16(h[crcOff:])
		var checksum byte
		for _, b := range h[2:baseEnd] {
			checksum += b
		}
		m.badHeader = checksum != h[1]
		if h[20] == 1 {
			next = int(binary.LittleEndian.Uint16(h[baseEnd-2:]))
		}
	case 2:
		baseEnd = int(binary.LittleEndian.Uint16(h[:2]))
		if baseEnd < 26 || baseEnd > len(h) {
			return m, ErrMalformed
		}
		m.crc = binary.LittleEndian.Uint16(h[21:23])
		next = int(binary.LittleEndian.Uint16(h[24:26]))
	case 3:
		baseEnd = int(binary.LittleEndian.Uint32(h[24:28]))
		if baseEnd < 32 || baseEnd > 1<<20 || baseEnd > len(h) {
			return m, ErrMalformed
		}
		m.crc = binary.LittleEndian.Uint16(h[21:23])
		next = int(binary.LittleEndian.Uint32(h[28:32]))
	default:
		return m, ErrUnsupported
	}
	m.start = uint64(baseEnd)
	m.packed = uint64(binary.LittleEndian.Uint32(h[7:11]))
	if h[20] != 0 {
		// LHA for OS-9/68k ('K') writes a level-2 total that omits the
		// final two-byte zero chain terminator. Accept that exact layout
		// without permitting arbitrary bytes beyond the declared header.
		chainEnd := baseEnd
		if h[20] == 2 && h[23] == 'K' && baseEnd+2 <= len(h) {
			chainEnd += 2
		}
		pos := baseEnd
		if h[20] == 2 {
			pos = 26
		} else if h[20] == 3 {
			pos = 32
		}
		for count := 0; next != 0; count++ {
			if count >= 256 {
				return m, ErrLimit
			}
			fieldSize := 2
			if h[20] == 3 {
				fieldSize = 4
			}
			if next < fieldSize+1 || next > len(h)-pos || h[20] >= 2 && next > chainEnd-pos {
				return m, ErrMalformed
			}
			pos += next
			if fieldSize == 4 {
				next = int(binary.LittleEndian.Uint32(h[pos-4:]))
			} else {
				next = int(binary.LittleEndian.Uint16(h[pos-2:]))
			}
			if h[20] == 1 {
				if uint64(pos-baseEnd) > m.packed {
					return m, ErrMalformed
				}
			} else if pos > chainEnd {
				return m, ErrMalformed
			}
		}
		if h[20] == 1 {
			m.packed -= uint64(pos - baseEnd)
			m.start = uint64(pos)
		} else if h[20] == 2 && pos == baseEnd+2 && chainEnd == pos &&
			h[baseEnd] == 0 && h[baseEnd+1] == 0 {
			m.start = uint64(pos)
		} else if pos < baseEnd {
			// The level-2 format permits one zero padding byte to avoid a
			// low-byte-zero header size, but not arbitrary hidden data.
			if h[20] != 2 || baseEnd-pos != 1 || h[pos] != 0 {
				return m, ErrUnsupported
			}
		} else if pos > baseEnd {
			return m, ErrMalformed
		}
	}
	if m.start > uint64(len(h)) || m.packed > uint64(len(h))-m.start {
		return m, ErrMalformed
	}
	return m, nil
}

func lhaCRC16(data []byte) uint16 {
	var crc uint16
	for _, b := range data {
		crc ^= uint16(b)
		for range 8 {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xa001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func (s *scanState) lha(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	for at := 0; at < len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if data[at] == 0 && !lhaHeader(data[at:]) { // LHA terminator, unless level-2 size has low byte zero
			return "", incomplete
		}
		h := data[at:]
		m, err := lhaMemberOf(h)
		if err != nil {
			return "", err
		}
		if m.badHeader {
			incomplete = ErrMalformed
		}
		body := h[int(m.start):int(m.start+m.packed)]
		switch {
		case bytes.Equal(m.method[:], []byte("-lh0-")), bytes.Equal(m.method[:], []byte("-lz4-")):
			if m.packed != m.plain {
				incomplete = ErrMalformed
			} else {
				if sig, err := s.childBytes(body, depth); sig != "" {
					return sig, nil
				} else if err != nil && incomplete == nil {
					incomplete = err
				}
				if lhaCRC16(body) != m.crc {
					incomplete = ErrMalformed
				}
			}
		case bytes.Equal(m.method[:], []byte("-lhd-")):
			// Directory records have no child data.
		case bytes.Equal(m.method[:], []byte("-lh1-")):
			if m.plain > uint64(min(s.maxFile, s.remaining)) || s.files >= s.maxFiles {
				return "", ErrLimit
			}
			out, err := lh1Expand(s.ctx, body, int(m.plain))
			s.spend(len(out))
			if err != nil {
				incomplete = err
				break
			}
			if sig, err := s.visitHeld(out, depth+1, int64(len(out))); sig != "" {
				return sig, nil
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
			if lhaCRC16(out) != m.crc {
				incomplete = ErrMalformed
			}
		case bytes.Equal(m.method[:], []byte("-lzs-")), bytes.Equal(m.method[:], []byte("-lz5-")):
			if m.plain > uint64(min(s.maxFile, s.remaining)) || s.files >= s.maxFiles {
				return "", ErrLimit
			}
			var out []byte
			if m.method[3] == 's' {
				out, err = lzsExpand(s.ctx, body, int(m.plain))
			} else {
				out, err = lz5Expand(s.ctx, body, int(m.plain))
			}
			s.spend(len(out))
			if err != nil {
				incomplete = err
				break
			}
			if sig, err := s.visitHeld(out, depth+1, int64(len(out))); sig != "" {
				return sig, nil
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
			if lhaCRC16(out) != m.crc {
				incomplete = ErrMalformed
			}
		case bytes.Equal(m.method[:], []byte("-lh4-")), bytes.Equal(m.method[:], []byte("-lh5-")),
			bytes.Equal(m.method[:], []byte("-lh6-")), bytes.Equal(m.method[:], []byte("-lh7-")):
			var dictBits int
			switch m.method[3] {
			case '4':
				dictBits = 12
			case '5':
				dictBits = 13
			case '6':
				dictBits = 15
			case '7':
				dictBits = 16
			}
			pBit := 4
			if dictBits >= 15 {
				pBit = 5
			}
			if m.plain > uint64(min(s.maxFile, s.remaining)) || s.files >= s.maxFiles {
				return "", ErrLimit
			}
			d := &arjDecoder{arjBits: arjBits{in: body}, ctx: s.ctx, dictSize: 1 << dictBits, np: dictBits + 1, pBit: pBit}
			out, err := d.decode(int(m.plain))
			s.spend(len(out))
			if err != nil {
				incomplete = err
				break
			}
			if sig, err := s.visitHeld(out, depth+1, int64(len(out))); sig != "" {
				return sig, nil
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
			if lhaCRC16(out) != m.crc {
				incomplete = ErrMalformed
			}
		case m.method[0] == '-' && m.method[4] == '-':
			incomplete = ErrUnsupported
		default:
			incomplete = ErrMalformed
		}
		at += int(m.start + m.packed)
	}
	// The format permits EOF immediately after a complete member.
	return "", incomplete
}

// LArc's LZS stream uses MSB-first literal and copy commands. The 2 KiB
// history starts with spaces and writes from offset 2031. A copy encodes an
// absolute 11-bit history offset and a four-bit length plus two.
func lzsExpand(ctx context.Context, in []byte, size int) ([]byte, error) {
	var history [2048]byte
	for i := range history {
		history[i] = ' '
	}
	pos, bit := 2048-17, 0
	out := arjInitialOutput(size, len(in))
	read := func(n int) (int, bool) {
		if len(in)*8-bit < n {
			return 0, false
		}
		v := 0
		for range n {
			v = v<<1 | int((in[bit>>3]>>(7-(bit&7)))&1)
			bit++
		}
		return v, true
	}
	put := func(c byte) {
		out = append(out, c)
		history[pos] = c
		pos = (pos + 1) & 2047
	}
	nextContextCheck := 0
	for len(out) < size {
		if len(out) >= nextContextCheck {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			nextContextCheck = len(out) + 4096
		}
		kind, ok := read(1)
		if !ok {
			return out, ErrMalformed
		}
		if kind == 1 {
			c, ok := read(8)
			if !ok {
				return out, ErrMalformed
			}
			put(byte(c))
			continue
		}
		start, ok := read(11)
		if !ok {
			return out, ErrMalformed
		}
		length, ok := read(4)
		if !ok {
			return out, ErrMalformed
		}
		for j := 0; j < length+2 && len(out) < size; j++ {
			put(history[(start+j)&2047])
		}
	}
	return out, nil
}

// LArc's LZ5 stream groups eight commands under a little-endian flag byte.
// Its 4 KiB dictionary has a format-defined preset; unlike SZDD, a match
// may refer to that preset before any literal is emitted.
func lz5Expand(ctx context.Context, in []byte, size int) ([]byte, error) {
	var history [4096]byte
	for c := 0; c < 256; c++ {
		for j := 0; j < 13; j++ {
			history[c*13+j] = byte(c)
		}
		history[3328+c] = byte(c)
		history[3584+c] = byte(255 - c)
	}
	for i := 3840; i < 3968; i++ {
		history[i] = 0
	}
	for i := 3968; i < 4078; i++ {
		history[i] = ' '
	}
	pos, at := 4096-18, 0
	out := arjInitialOutput(size, len(in))
	put := func(c byte) {
		out = append(out, c)
		history[pos] = c
		pos = (pos + 1) & 4095
	}
	nextContextCheck := 0
	for len(out) < size {
		if len(out) >= nextContextCheck {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			nextContextCheck = len(out) + 4096
		}
		if at >= len(in) {
			return out, ErrMalformed
		}
		flags := in[at]
		at++
		for bit := 0; bit < 8 && len(out) < size; bit++ {
			if flags&(1<<bit) != 0 {
				if at >= len(in) {
					return out, ErrMalformed
				}
				put(in[at])
				at++
				continue
			}
			if len(in)-at < 2 {
				return out, ErrMalformed
			}
			start := int(in[at]) | int(in[at+1]&0xf0)<<4
			length := int(in[at+1]&0x0f) + 3
			at += 2
			for j := 0; j < length && len(out) < size; j++ {
				put(history[(start+j)&4095])
			}
		}
	}
	return out, nil
}
