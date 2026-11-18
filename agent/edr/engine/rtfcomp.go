package engine

import (
	"encoding/binary"
	"hash/crc32"
)

// MS-OXRTFCP uses a CRC-32 table with a zero seed and no final inversion,
// unlike the checksum returned by hash/crc32.ChecksumIEEE.
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxrtfcp/a395d090-82f6-4ac0-85c4-057678b2c8e8
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxrtfcp/592e0b55-8b86-45ae-b9d1-b4830e2969ea
func rtfCRC(data []byte) uint32 {
	table := crc32.IEEETable
	var crc uint32
	for _, b := range data {
		crc = table[byte(crc)^b] ^ crc>>8
	}
	return crc
}

const rtfPrebuffer = "{\\rtf1\\ansi\\mac\\deff0\\deftab720{\\fonttbl;}{\\f0\\fnil \\froman \\fswiss \\fmodern \\fscript \\fdecor MS Sans SerifSymbolArialTimes New RomanCourier{\\colortbl\\red0\\green0\\blue0\r\n\\par \\pard\\plain\\f0\\fs20\\b\\i\\u\\tab\\tx"

func (s *scanState) decompressRTF(value []byte) ([]byte, error) {
	if len(value) < 16 {
		return nil, ErrMalformed
	}
	compressedSize := binary.LittleEndian.Uint32(value[:4])
	rawSize := binary.LittleEndian.Uint32(value[4:8])
	kind := binary.LittleEndian.Uint32(value[8:12])
	if uint64(compressedSize)+4 != uint64(len(value)) || compressedSize < 12 {
		return nil, ErrMalformed
	}
	maxOutput := min(s.maxFile, s.remaining)
	if uint64(rawSize) > uint64(max(0, maxOutput)) {
		return nil, ErrLimit
	}
	contents := value[16:]
	switch kind {
	case 0x414c454d: // MELA: uncompressed content
		if uint64(len(contents)) != uint64(rawSize) {
			return contents, ErrMalformed
		}
		return contents, nil
	case 0x75465a4c: // LZFu
	default:
		return nil, ErrUnsupported
	}
	if len(rtfPrebuffer) != 207 {
		return nil, ErrMalformed
	}
	// One control byte and eight 2-byte references produce at most 8*17
	// bytes, so LZFu output never exceeds eight times its input. A larger
	// RawSize is impossible; the output buffer grows with decoded bytes rather
	// than being sized from the header.
	if uint64(rawSize) > 8*uint64(len(contents)) {
		return nil, ErrMalformed
	}
	var dict [4096]byte
	copy(dict[:], rtfPrebuffer)
	writePos, endPos, pos := 207, 207, 0
	out := make([]byte, 0, min(int(rawSize), 4*len(contents)))
	appendByte := func(b byte) error {
		if uint64(len(out)) >= uint64(rawSize) {
			return ErrMalformed
		}
		out = append(out, b)
		dict[writePos] = b
		writePos = (writePos + 1) & 4095
		if endPos < len(dict) {
			endPos++
		}
		return nil
	}
	terminated := false
	for pos < len(contents) && !terminated {
		if err := s.ctx.Err(); err != nil {
			return out, err
		}
		control := contents[pos]
		pos++
		for bit := 0; bit < 8; bit++ {
			if control&(1<<bit) == 0 {
				if pos >= len(contents) {
					return out, ErrMalformed
				}
				if err := appendByte(contents[pos]); err != nil {
					return out, err
				}
				pos++
				continue
			}
			if len(contents)-pos < 2 {
				return out, ErrMalformed
			}
			ref := binary.BigEndian.Uint16(contents[pos : pos+2])
			pos += 2
			readPos := int(ref >> 4)
			if readPos == writePos {
				terminated = true
				break
			}
			for n := int(ref&15) + 2; n > 0; n-- {
				if readPos >= endPos {
					return out, ErrMalformed
				}
				b := dict[readPos]
				readPos = (readPos + 1) & 4095
				if err := appendByte(b); err != nil {
					return out, err
				}
			}
		}
	}
	if !terminated || uint64(len(out)) != uint64(rawSize) || rtfCRC(contents) != binary.LittleEndian.Uint32(value[12:16]) {
		return out, ErrMalformed
	}
	return out, nil
}
