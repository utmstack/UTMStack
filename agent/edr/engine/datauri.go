package engine

import (
	"bytes"
	"context"
)

// RFC 2397 data URIs carry whole files inside markup: a media type, an
// optional ";base64" flag, a comma, then the data. Browsers download or
// render them like any other link target, so each one in an attribute value
// or CSS url() is decoded and analyzed as an object.
// https://www.rfc-editor.org/rfc/rfc2397
const maxDataURIHeader = 256

func (s *scanState) dataURIs(data []byte, depth int) (string, error) {
	var incomplete error
	for pos := 0; pos < len(data); {
		i := bytes.IndexByte(data[pos:], ':')
		if i < 0 {
			break
		}
		colon := pos + i
		pos = colon + 1
		if colon < 5 || !bytes.EqualFold(data[colon-4:colon], []byte("data")) {
			continue
		}
		open := data[colon-5]
		if open != '"' && open != '\'' && open != '=' && open != '(' {
			continue
		}
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		comma := -1
		for j := pos; j < min(len(data), pos+maxDataURIHeader); j++ {
			c := data[j]
			if c == ',' {
				comma = j
				break
			}
			if !dataURIHeaderByte(c) {
				break
			}
		}
		if comma < 0 {
			continue
		}
		header := data[pos:comma]
		end := dataURIEnd(data, comma+1, open)
		pos = end
		sig, err := s.dataURIContent(header, data[comma+1:end], depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
			if err == ErrLimit {
				return "", err
			}
		}
	}
	return "", incomplete
}

func (s *scanState) dataURIContent(header, body []byte, depth int) (string, error) {
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	limit := min(s.maxFile, s.remaining)
	if limit < 0 {
		return "", ErrLimit
	}
	isBase64 := len(header) >= 7 && bytes.EqualFold(header[len(header)-7:], []byte(";base64"))
	decoded, err := dataURIDecodeContext(s.ctx, body, isBase64, limit)
	if err != nil {
		s.spend(len(decoded))
		return "", err
	}
	return s.visit(decoded, depth+1)
}

func dataURIHeaderByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || bytes.IndexByte([]byte("!#$&^_.+-/;=%"), c) >= 0
}

// dataURIEnd finds where the URI ends: at the closing quote of a quoted
// attribute value, at white space or '>' for an unquoted one, and at ')' or a
// quote inside CSS url().
func dataURIEnd(data []byte, pos int, open byte) int {
	for i := pos; i < len(data); i++ {
		c := data[i]
		switch open {
		case '"', '\'':
			if c == open {
				return i
			}
		case '=':
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' {
				return i
			}
		case '(':
			if c == ')' || c == '"' || c == '\'' {
				return i
			}
		}
	}
	return len(data)
}

// dataURIDecode percent-decodes the data and, for base64, decodes it
// leniently as browsers do: white space is skipped and decoding stops at
// padding or a character outside the alphabet.
func dataURIDecode(text []byte, isBase64 bool, limit int64) ([]byte, error) {
	return dataURIDecodeContext(context.Background(), text, isBase64, limit)
}

func dataURIDecodeContext(ctx context.Context, text []byte, isBase64 bool, limit int64) ([]byte, error) {
	var out []byte
	var acc uint32
	bits := 0
	for i := 0; i < len(text); i++ {
		if i&4095 == 0 {
			if err := ctx.Err(); err != nil {
				return out, err
			}
		}
		c := text[i]
		if c == '%' && i+2 < len(text) {
			hi, ok1 := hexNibble(text[i+1])
			lo, ok2 := hexNibble(text[i+2])
			if ok1 && ok2 {
				c = hi<<4 | lo
				i += 2
			}
		}
		if !isBase64 {
			if int64(len(out)) >= limit {
				return out, ErrLimit
			}
			out = append(out, c)
			continue
		}
		var v byte
		switch {
		case c >= 'A' && c <= 'Z':
			v = c - 'A'
		case c >= 'a' && c <= 'z':
			v = c - 'a' + 26
		case c >= '0' && c <= '9':
			v = c - '0' + 52
		case c == '+' || c == '-':
			v = 62
		case c == '/' || c == '_':
			v = 63
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f':
			continue
		default:
			return out, nil // padding or the end of the encoded data
		}
		acc = acc<<6 | uint32(v)
		bits += 6
		if bits >= 8 {
			bits -= 8
			if int64(len(out)) >= limit {
				return out, ErrLimit
			}
			out = append(out, byte(acc>>bits))
			acc &= 1<<bits - 1
		}
	}
	return out, nil
}
