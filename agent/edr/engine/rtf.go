package engine

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// RTF embeds OLE objects as hexadecimal text in \objdata destinations. The
// decoded bytes are an OLE1 object: version, format (2 = embedded), class,
// topic and item names as length-prefixed strings, then the native data size
// and the native data. The "Package" class holds the packager layout of an
// Ole10Native stream, others usually a compound file.
// Word ignores characters other than hexadecimal digits in the text and reads
// \binN as N raw bytes; both are honored here. The object layout is the
// ObjectHeader and EmbeddedObject structures of [MS-OLEDS] section 2.2.
func (s *scanState) rtf(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	for pos := 0; pos < len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		i := bytes.Index(data[pos:], []byte(`\objdata`))
		if i < 0 {
			break
		}
		at := pos + i
		pos = at + len(`\objdata`)
		if pos < len(data) && isRTFLetter(data[pos]) {
			continue // a longer control word
		}
		escapes := 0
		for j := at - 1; j >= 0 && data[j] == '\\'; j-- {
			escapes++
		}
		if escapes%2 == 1 {
			continue // an escaped backslash followed by text
		}
		if s.files >= s.maxFiles {
			return "", ErrLimit
		}
		limit := min(s.maxFile, s.remaining)
		if limit < 0 {
			return "", ErrLimit
		}
		object, next, err := rtfObjectData(data, pos, limit)
		pos = next
		if err != nil {
			incomplete = err
			if err == ErrLimit {
				s.spend(len(object)) // discarded expansion work
				return "", err
			}
		}
		if len(object) == 0 {
			continue
		}
		// The object's contents are charged when they are analyzed.
		sig, err := s.ole1Object(object, depth+1)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	if sig, err := s.rtfText(data, depth); sig != "" {
		return sig, nil
	} else if err != nil && incomplete == nil {
		incomplete = err
	}
	return "", incomplete
}

func isRTFLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

// rtfObjectData decodes the destination that starts at pos, up to the brace
// that closes its group. Nested groups are decoded as part of it, as Word
// does; control words other than \bin are skipped.
func rtfObjectData(data []byte, pos int, limit int64) (out []byte, next int, err error) {
	nibble, half := byte(0), false
	depth := 0
	for pos < len(data) {
		c := data[pos]
		pos++
		switch {
		case c == '{':
			depth++
		case c == '}':
			if depth == 0 {
				return out, pos, nil
			}
			depth--
		case c == '\\':
			if pos >= len(data) {
				return out, pos, ErrMalformed
			}
			if !isRTFLetter(data[pos]) {
				pos++ // control symbol such as \* or an escaped brace
				continue
			}
			start := pos
			for pos < len(data) && isRTFLetter(data[pos]) {
				pos++
			}
			word := string(data[start:pos])
			argStart := pos
			if pos < len(data) && data[pos] == '-' {
				pos++
			}
			for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' && pos-argStart < 10 {
				pos++
			}
			arg := data[argStart:pos]
			if pos < len(data) && data[pos] == ' ' {
				pos++
			}
			if word == "bin" {
				n := 0
				for _, d := range arg {
					if d == '-' {
						return out, pos, ErrMalformed
					}
					n = n*10 + int(d-'0')
				}
				if n > len(data)-pos {
					return out, len(data), ErrMalformed
				}
				if int64(len(out)+n) > limit {
					return out, pos, ErrLimit
				}
				out = append(out, data[pos:pos+n]...)
				pos += n
				half = false
			}
		default:
			v, ok := hexNibble(c)
			if !ok {
				continue
			}
			if !half {
				nibble, half = v, true
				continue
			}
			if int64(len(out)) >= limit {
				return out, pos, ErrLimit
			}
			out = append(out, nibble<<4|v)
			half = false
		}
	}
	return out, pos, ErrMalformed // the group never closes
}

func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// ole1Object analyzes the native data of an OLE1 embedded object; bytes
// without an OLE1 header are analyzed as they are.
func (s *scanState) ole1Object(b []byte, depth int) (string, error) {
	class, native, format, ok := ole1Native(b)
	switch {
	case !ok:
		return s.visit(b, depth)
	case format == ole1Linked:
		return s.visit(b, depth) // a link holds no embedded data
	case native < 0:
		if sig, err := s.visit(b, depth); sig != "" || err != nil {
			return sig, err
		}
		return "", ErrMalformed
	case strings.EqualFold(class, "Package"):
		// The size field and native data have the Ole10Native stream layout.
		return s.ole10Native(b[native:], depth)
	}
	size := binary.LittleEndian.Uint32(b[native:])
	body := b[native+4:]
	var incomplete error
	if uint64(size) > uint64(len(body)) {
		incomplete = ErrMalformed
	} else {
		body = body[:size]
	}
	if sig, err := s.visit(body, depth); sig != "" || err != nil {
		return sig, err
	}
	return "", incomplete
}

const (
	ole1Linked   = 1
	ole1Embedded = 2
)

// ole1Native parses the OLE1 header. ok reports the OLE1 version field;
// native is the offset of an embedded object's native data size, or -1 when
// the object is linked or its header is not well formed.
func ole1Native(b []byte) (class string, native int, format uint32, ok bool) {
	if len(b) < 8 || binary.LittleEndian.Uint32(b) != 0x0501 {
		return "", -1, 0, false
	}
	format = binary.LittleEndian.Uint32(b[4:])
	if format != ole1Embedded {
		return "", -1, format, true
	}
	pos := 8
	for field := 0; field < 3; field++ {
		if len(b)-pos < 4 {
			return "", -1, format, true
		}
		n := binary.LittleEndian.Uint32(b[pos:])
		pos += 4
		if n > 4096 || int(n) > len(b)-pos {
			return "", -1, format, true
		}
		if field == 0 {
			class = string(bytes.TrimRight(b[pos:pos+int(n)], "\x00"))
		}
		pos += int(n)
	}
	if len(b)-pos < 4 {
		return class, -1, format, true
	}
	return class, pos, format, true
}
