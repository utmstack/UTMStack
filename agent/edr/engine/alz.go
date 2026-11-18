package engine

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"encoding/binary"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// ALZ local records use a four-byte signature followed by a nine-byte base
// header. The high nibble of descriptor selects the width of both sizes.
// The layout and method numbers are described by ClamAV's public ALZ tests;
// this parser is independent Go code and keeps archive bytes in memory only
// for the duration of the enclosing bounded scan.
const (
	alzFileHeader  = 0x015a4c41
	alzLocalHeader = 0x015a4c42
	alzCentral     = 0x015a4c43
	alzEnd         = 0x025a4c43
)

func alzHeader(data []byte) bool {
	return len(data) >= 8 && binary.LittleEndian.Uint32(data) == alzFileHeader
}

type alzMember struct {
	name       string
	packed     uint64
	plain      uint64
	method     byte
	attribute  byte
	descriptor byte
	start      int
	next       int
}

func alzMemberOf(data []byte, at int) (alzMember, error) {
	var m alzMember
	if at < 0 || at > len(data) || len(data)-at < 13 || binary.LittleEndian.Uint32(data[at:]) != alzLocalHeader {
		return m, ErrMalformed
	}
	nameSize := int(binary.LittleEndian.Uint16(data[at+4:]))
	m.attribute = data[at+6]
	m.descriptor = data[at+11]
	at += 13
	if nameSize == 0 {
		return m, ErrMalformed
	}
	width := int(m.descriptor >> 4)
	if width != 0 {
		if width != 1 && width != 2 && width != 4 && width != 8 {
			return m, ErrUnsupported
		}
		if len(data)-at < 6+2*width {
			return m, ErrMalformed
		}
		m.method = data[at]
		at += 6 // method, reserved byte and CRC-32; the reference ignores CRC.
		switch width {
		case 1:
			m.packed, m.plain = uint64(data[at]), uint64(data[at+1])
		case 2:
			m.packed = uint64(binary.LittleEndian.Uint16(data[at:]))
			m.plain = uint64(binary.LittleEndian.Uint16(data[at+2:]))
		case 4:
			m.packed = uint64(binary.LittleEndian.Uint32(data[at:]))
			m.plain = uint64(binary.LittleEndian.Uint32(data[at+4:]))
		case 8:
			m.packed = binary.LittleEndian.Uint64(data[at:])
			m.plain = binary.LittleEndian.Uint64(data[at+8:])
		}
		at += 2 * width
	}
	if len(data)-at < nameSize {
		return m, ErrMalformed
	}
	m.name = string(data[at : at+nameSize])
	at += nameSize
	if m.descriptor&1 != 0 {
		if len(data)-at < 12 {
			return m, ErrMalformed
		}
		at += 12 // encrypted header
	}
	m.start = at
	if m.packed > uint64(len(data)-at) {
		return m, ErrMalformed
	}
	m.next = at + int(m.packed)
	return m, nil
}

func (s *scanState) alz(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	if !alzHeader(data) {
		return "", ErrMalformed
	}
	var incomplete error
	position := 1
	for at := 8; at < len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if position > s.maxFiles {
			return "", ErrLimit
		}
		if len(data)-at < 4 {
			return "", ErrMalformed
		}
		switch binary.LittleEndian.Uint32(data[at:]) {
		case alzEnd:
			return "", incomplete
		case alzCentral:
			if len(data)-at < 12 {
				return "", ErrMalformed
			}
			at += 12
			continue
		case alzLocalHeader:
			m, err := alzMemberOf(data, at)
			if err != nil {
				return "", err
			}
			at = m.next
			meta := signatures.MemberMeta{Name: m.name, HasName: true, CompressedSize: int64(m.packed), Size: int64(min(m.plain, uint64(^uint64(0)>>1))), Encrypted: m.descriptor&1 != 0, Index: position}
			position++
			if name, err := s.db.MatchMember(s.ctx, signatures.TypeALZ, int64(len(data)), meta); name != "" || err != nil {
				return name, err
			}
			if m.attribute&0x10 != 0 || m.packed == 0 {
				continue
			}
			if m.descriptor&9 != 0 || m.method > 2 {
				incomplete = ErrUnsupported
				continue
			}
			body := data[m.start:m.next]
			var sig string
			switch m.method {
			case 0:
				sig, err = s.childBytes(body, depth)
			case 1:
				sig, err = s.partialChild(bzip2.NewReader(bytes.NewReader(body)), depth)
			case 2:
				r := flate.NewReader(bytes.NewReader(body))
				sig, err = s.partialChild(r, depth)
				_ = r.Close()
			}
			if sig != "" {
				return sig, nil
			}
			if err != nil {
				incomplete = err
			}
		default:
			return "", ErrMalformed
		}
	}
	// The reference accepts EOF after complete local records.
	return "", incomplete
}
