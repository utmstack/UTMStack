package engine

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
)

// TNEF is specified by MS-OXTNEF. ClamAV's tnef.c also identifies the
// legacy attachment-data attribute, but the wire specification is the source
// for attribute checksums and encapsulated MAPI properties here.
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxtnef/b03b5704-ef54-4390-b981-dc78f8e79a05
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxtnef/c27c89ec-7f8b-48d3-9e27-dc7600d23854
// https://raw.githubusercontent.com/Cisco-Talos/clamav/main/libclamav/tnef.c
const (
	tnefVersion    = 0x9006
	tnefCodePage   = 0x9007
	tnefBody       = 0x800c
	tnefAttachData = 0x800f
	tnefAttachMeta = 0x8011
	tnefAttachRend = 0x9002
	tnefMsgProps   = 0x9003
	tnefRecipients = 0x9004
	tnefAttachProp = 0x9005
)

func (s *scanState) tnef(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth || len(data) < 6 {
		if depth >= s.maxDepth {
			return "", ErrLimit
		}
		return "", ErrMalformed
	}
	if !bytes.Equal(data[:4], []byte{0x78, 0x9f, 0x3e, 0x22}) {
		return "", ErrMalformed
	}
	pos, attributes, attachments := 6, 0, 0
	var incomplete error
	for pos < len(data) {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		attributes++
		if attributes > s.maxFiles || attachments > s.maxFiles-s.files {
			return "", ErrLimit
		}
		if len(data)-pos < 11 {
			return "", ErrMalformed
		}
		level := data[pos]
		tag := binary.LittleEndian.Uint32(data[pos+1 : pos+5])
		length := binary.LittleEndian.Uint32(data[pos+5 : pos+9])
		pos += 9
		if length > 0x7fffffff || uint64(length)+2 > uint64(len(data)-pos) {
			return "", ErrMalformed
		}
		value := data[pos : pos+int(length)]
		pos += int(length)
		checksum := binary.LittleEndian.Uint16(data[pos : pos+2])
		pos += 2
		var sum uint16
		for _, b := range value {
			sum += uint16(b)
		}
		if sum != checksum && incomplete == nil {
			incomplete = ErrMalformed
		}
		id := uint16(tag)
		// Version and codepage are mandatory and precede message/attachment
		// data. Unsupported variants remain incomplete even if later payloads
		// happen to contain no matching signature.
		if (attributes == 1 && (level != 1 || tag != 0x00089006 || len(value) != 4)) ||
			(attributes == 2 && (level != 1 || tag != 0x00069007 || len(value) != 8)) {
			if incomplete == nil {
				incomplete = ErrMalformed
			}
		}
		if !tnefKnownAttribute(level, id) && incomplete == nil {
			incomplete = ErrUnsupported
		}
		if expected := tnefExpectedTag(id); expected != 0 && tag != expected && incomplete == nil {
			incomplete = ErrUnsupported
		}
		if level == 2 && id == tnefAttachRend {
			attachments++
		} else if level == 2 && attachments == 0 && incomplete == nil {
			incomplete = ErrMalformed
		}
		var sig string
		var err error
		switch {
		case level == 1 && id == tnefBody,
			level == 2 && (id == tnefAttachData || id == tnefAttachMeta):
			if len(value) != 0 {
				sig, err = s.visit(value, depth+1)
			}
		case level == 1 && id == tnefMsgProps, level == 2 && id == tnefAttachProp:
			sig, err = s.tnefProperties(value, depth+1)
		case level == 1 && id == tnefRecipients:
			sig, err = s.tnefRecipients(value, depth+1)
		case level != 1 && level != 2:
			err = ErrUnsupported
		}
		if sig != "" {
			return sig, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	if attributes < 2 {
		return "", ErrMalformed
	}
	return "", incomplete
}

func tnefKnownAttribute(level byte, id uint16) bool {
	if level == 1 {
		switch id {
		case tnefVersion, tnefCodePage, tnefMsgProps, tnefRecipients,
			tnefBody, 0x8000, 0x8004, 0x8005, 0x8006, 0x8007, 0x8008,
			0x8009, 0x800d, 0x8020, 0x0600, 0x0000, 0x0001,
			0x0002, 0x0006, 0x0007, 0x0008, 0x0009:
			return true
		}
	}
	if level == 2 {
		switch id {
		case tnefAttachData, tnefAttachMeta, tnefAttachRend, tnefAttachProp,
			0x8010, 0x8012, 0x8013, 0x9001:
			return true
		}
	}
	return false
}

func tnefExpectedTag(id uint16) uint32 {
	switch id {
	case tnefBody:
		return 0x0002800c
	case tnefAttachData:
		return 0x0006800f
	case tnefAttachMeta:
		return 0x00068011
	case tnefAttachRend:
		return 0x00069002
	case tnefMsgProps:
		return 0x00069003
	case tnefRecipients:
		return 0x00069004
	case tnefAttachProp:
		return 0x00069005
	case tnefVersion:
		return 0x00089006
	case tnefCodePage:
		return 0x00069007
	case 0x8008:
		return 0x00078008
	case 0x8000:
		return 0x00008000
	case 0x8004:
		return 0x00018004
	case 0x8005:
		return 0x00038005
	case 0x8006:
		return 0x00038006
	case 0x8007:
		return 0x00068007
	case 0x8009:
		return 0x00018009
	case 0x800d:
		return 0x0004800d
	case 0x8020:
		return 0x00038020
	case 0x0600:
		return 0x00070600
	case 0x0000, 0x0001, 0x0002:
		return 0x00060000 | uint32(id)
	case 0x0006, 0x0007:
		return 0x00030000 | uint32(id)
	case 0x0008:
		return 0x00050008
	case 0x0009:
		return 0x00040009
	case 0x8010:
		return 0x00018010
	case 0x8012, 0x8013:
		return 0x00038000 | uint32(id)
	case 0x9001:
		return 0x00069001
	}
	return 0
}

func (s *scanState) tnefRecipients(data []byte, depth int) (string, error) {
	if len(data) < 4 {
		return "", ErrMalformed
	}
	count := binary.LittleEndian.Uint32(data[:4])
	if count > uint32(s.maxFiles) || uint64(count)*4 > uint64(len(data)-4) {
		return "", ErrLimit
	}
	pos := 4
	var incomplete error
	for i := uint32(0); i < count; i++ {
		consumed, sig, err := s.tnefPropertyList(data[pos:], depth)
		if sig != "" {
			return sig, nil
		}
		pos += consumed
		if err != nil {
			return "", err
		}
	}
	if pos != len(data) {
		incomplete = ErrMalformed
	}
	return "", incomplete
}

func (s *scanState) tnefProperties(data []byte, depth int) (string, error) {
	used, sig, err := s.tnefPropertyList(data, depth)
	if sig != "" || err != nil {
		return sig, err
	}
	if used != len(data) {
		return "", ErrMalformed
	}
	return "", nil
}

func (s *scanState) tnefPropertyList(data []byte, depth int) (int, string, error) {
	if len(data) < 4 {
		return 0, "", ErrMalformed
	}
	count := binary.LittleEndian.Uint32(data[:4])
	if count > uint32(s.maxFiles) || uint64(count)*4 > uint64(len(data)-4) {
		return 0, "", ErrLimit
	}
	pos := 4
	var incomplete error
	for i := uint32(0); i < count; i++ {
		if err := s.ctx.Err(); err != nil {
			return pos, "", err
		}
		if len(data)-pos < 4 {
			return pos, "", ErrMalformed
		}
		typ := binary.LittleEndian.Uint16(data[pos : pos+2])
		id := binary.LittleEndian.Uint16(data[pos+2 : pos+4])
		pos += 4
		if id >= 0x8000 {
			// Named properties have a namespace GUID and either a number or
			// a padded UTF-16 name before the value.
			if len(data)-pos < 20 {
				return pos, "", ErrMalformed
			}
			nameKind := binary.LittleEndian.Uint32(data[pos+16 : pos+20])
			pos += 20
			switch nameKind {
			case 0:
				if len(data)-pos < 4 {
					return pos, "", ErrMalformed
				}
				pos += 4
			case 1:
				if len(data)-pos < 4 {
					return pos, "", ErrMalformed
				}
				nameLen := binary.LittleEndian.Uint32(data[pos : pos+4])
				pos += 4
				if nameLen > uint32(len(data)-pos) || nameLen%2 != 0 {
					return pos, "", ErrMalformed
				}
				padded := (uint64(nameLen) + 3) &^ 3
				if padded > uint64(len(data)-pos) {
					return pos, "", ErrMalformed
				}
				pos += int(padded)
			default:
				return pos, "", ErrUnsupported
			}
		}
		base := typ &^ 0x1000
		if base == 0x0001 && typ == base { // PtypNull
			continue
		}
		var countValues uint32 = 1
		if typ&0x1000 != 0 {
			if len(data)-pos < 4 {
				return pos, "", ErrMalformed
			}
			countValues = binary.LittleEndian.Uint32(data[pos : pos+4])
			pos += 4
			if countValues > uint32(s.maxFiles) || uint64(countValues)*4 > uint64(len(data)-pos) {
				return pos, "", ErrLimit
			}
		}
		size := 0
		variable := false
		switch base {
		case 0x0002, 0x0003, 0x0004, 0x000a, 0x000b:
			size = 4 // 16-bit scalars are padded to four bytes
		case 0x0005, 0x0006, 0x0007, 0x0014, 0x0040:
			size = 8
		case 0x0048:
			size = 16
		case 0x000d, 0x001e, 0x001f, 0x0102:
			variable = true
		default:
			return pos, "", ErrUnsupported
		}
		if !variable {
			if uint64(countValues)*uint64(size) > uint64(len(data)-pos) {
				return pos, "", ErrMalformed
			}
			pos += int(countValues) * size
			continue
		}
		if typ&0x1000 == 0 {
			if len(data)-pos < 4 || binary.LittleEndian.Uint32(data[pos:pos+4]) != 1 {
				return pos, "", ErrMalformed
			}
			pos += 4
		}
		for j := uint32(0); j < countValues; j++ {
			if len(data)-pos < 4 {
				return pos, "", ErrMalformed
			}
			n := binary.LittleEndian.Uint32(data[pos : pos+4])
			pos += 4
			padded := (uint64(n) + 3) &^ 3
			if padded > uint64(len(data)-pos) {
				return pos, "", ErrMalformed
			}
			value := data[pos : pos+int(n)]
			pos += int(padded)
			if len(value) == 0 {
				continue
			}
			sig, err := s.tnefPropertyValue(value, typ, id, depth)
			if sig != "" {
				return pos, sig, nil
			}
			if err != nil && incomplete == nil {
				incomplete = err
			}
		}
	}
	return pos, "", incomplete
}

func (s *scanState) tnefPropertyValue(value []byte, typ, id uint16, depth int) (string, error) {
	if id == 0x1009 { // PidTagRtfCompressed
		if typ != 0x0102 {
			return "", ErrUnsupported
		}
		decoded, decodeErr := s.decompressRTF(value)
		if len(decoded) != 0 {
			sig, err := s.visit(decoded, depth)
			if sig != "" || err != nil {
				return sig, err
			}
		}
		return "", decodeErr
	}
	if typ&^0x1000 == 0x001f {
		if len(value)%2 != 0 {
			return "", ErrMalformed
		}
		decoded := make([]byte, 0, min(len(value), int(max(0, s.remaining))))
		for i := 0; i < len(value); i += 2 {
			if i&4095 == 0 {
				if err := s.ctx.Err(); err != nil {
					return "", err
				}
			}
			u := binary.LittleEndian.Uint16(value[i : i+2])
			r := rune(u)
			if u >= 0xd800 && u <= 0xdbff {
				if len(value)-i < 4 {
					return "", ErrMalformed
				}
				v := binary.LittleEndian.Uint16(value[i+2 : i+4])
				if v < 0xdc00 || v > 0xdfff {
					return "", ErrMalformed
				}
				r = utf16.DecodeRune(r, rune(v))
				i += 2
			} else if u >= 0xdc00 && u <= 0xdfff {
				return "", ErrMalformed
			}
			if int64(len(decoded)+utf8.RuneLen(r)) > min(s.maxFile, s.remaining) {
				return "", ErrLimit
			}
			decoded = utf8.AppendRune(decoded, r)
		}
		decoded = bytes.TrimRight(decoded, "\x00")
		return s.visit(decoded, depth)
	}
	if typ&^0x1000 == 0x000d { // PtypObject has a 16-byte interface identifier.
		if len(value) < 16 {
			return "", ErrMalformed
		}
		sig, err := s.visit(value[16:], depth)
		if sig != "" || err != nil {
			return sig, err
		}
		return "", ErrUnsupported // arbitrary embedded OLE/Message object
	}
	sig, err := s.visit(value, depth)
	if sig != "" || err != nil {
		return sig, err
	}
	return "", nil
}
