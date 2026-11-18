package engine

import (
	"bytes"
	"encoding/base64"
	"io"
)

// Both historical and begin-base64 envelopes follow POSIX uuencode(1p).
// The decoded bytes enter the same recursive scanner and aggregate budget.
// https://pubs.opengroup.org/onlinepubs/9699919799/utilities/uuencode.html
func (s *scanState) uuencoded(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	for pos := 0; pos < len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		line, next := uuLine(data, pos)
		base64Mode, found, valid := uuBegin(line)
		if !found {
			pos = next
			continue
		}
		if !valid {
			if incomplete == nil {
				incomplete = ErrMalformed
			}
			pos = next
			continue
		}
		maxOutput := min(s.maxFile, s.remaining)
		if maxOutput < 0 {
			return "", ErrLimit
		}
		var decoded []byte
		var err error
		if base64Mode {
			decoded, pos, err = s.uuBase64Block(data, next, maxOutput)
		} else {
			decoded, pos, err = s.uuHistoricalBlock(data, next, maxOutput)
		}
		if len(decoded) != 0 {
			sig, scanErr := s.visit(decoded, depth+1)
			if sig != "" {
				return sig, nil
			}
			if scanErr != nil && incomplete == nil {
				incomplete = scanErr
			}
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	return "", incomplete
}

func uuLine(data []byte, pos int) ([]byte, int) {
	end := bytes.IndexByte(data[pos:], '\n')
	if end < 0 {
		return bytes.TrimSuffix(data[pos:], []byte{'\r'}), len(data)
	}
	line := data[pos : pos+end]
	return bytes.TrimSuffix(line, []byte{'\r'}), pos + end + 1
}

// uuBegin recognizes an envelope header only when the keyword is followed by
// a 3-4 digit mode and a space, as the reference mail decoder does. Other
// lines that start with "begin " (PowerShell begin blocks, SQL transactions,
// prose) are ordinary text. A recognized header is valid when the mode is
// octal and a file name follows, as POSIX requires.
func uuBegin(line []byte) (base64Mode, found, valid bool) {
	switch {
	case bytes.HasPrefix(line, []byte("begin ")):
		line = line[len("begin "):]
	case bytes.HasPrefix(line, []byte("begin-base64 ")):
		line = line[len("begin-base64 "):]
		base64Mode = true
	default:
		return false, false, false
	}
	space := bytes.IndexByte(line, ' ')
	if space < 3 || space > 4 {
		return base64Mode, false, false
	}
	for _, digit := range line[:space] {
		if digit < '0' || digit > '9' {
			return base64Mode, false, false
		}
	}
	if len(line)-space < 2 || len(line)-space > 257 {
		return base64Mode, true, false
	}
	for _, digit := range line[:space] {
		if digit > '7' {
			return base64Mode, true, false
		}
	}
	return base64Mode, true, true
}

func uuValue(c byte) (byte, bool) {
	if c == '`' {
		return 0, true // widely used zero-byte spelling
	}
	if c < ' ' || c > '_' {
		return 0, false
	}
	return c - ' ', true
}

func (s *scanState) uuHistoricalBlock(data []byte, pos int, maxOutput int64) ([]byte, int, error) {
	out := make([]byte, 0)
	zeroLine := false
	for pos < len(data) {
		if err := s.ctx.Err(); err != nil {
			return out, pos, err
		}
		line, next := uuLine(data, pos)
		pos = next
		if zeroLine {
			if bytes.Equal(line, []byte("end")) {
				return out, pos, nil
			}
			return out, pos, ErrMalformed
		}
		if len(line) == 0 {
			return out, pos, ErrMalformed
		}
		n, ok := uuValue(line[0])
		if !ok || n > 45 {
			return out, pos, ErrMalformed
		}
		if n == 0 {
			if len(line) != 1 {
				return out, pos, ErrMalformed
			}
			zeroLine = true
			continue
		}
		groups := (int(n) + 2) / 3
		if len(line) != 1+groups*4 {
			return out, pos, ErrMalformed
		}
		if int64(n) > maxOutput-int64(len(out)) {
			return out, pos, ErrLimit
		}
		for i := 0; i < groups; i++ {
			a, goodA := uuValue(line[1+4*i])
			b, goodB := uuValue(line[2+4*i])
			c, goodC := uuValue(line[3+4*i])
			d, goodD := uuValue(line[4+4*i])
			if !goodA || !goodB || !goodC || !goodD {
				return out, pos, ErrMalformed
			}
			triple := [3]byte{a<<2 | b>>4, b<<4 | c>>2, c<<6 | d}
			out = append(out, triple[:min(3, int(n)-3*i)]...)
		}
	}
	return out, pos, ErrMalformed
}

func (s *scanState) uuBase64Block(data []byte, pos int, maxOutput int64) ([]byte, int, error) {
	start := pos
	for pos < len(data) {
		if err := s.ctx.Err(); err != nil {
			return nil, pos, err
		}
		line, next := uuLine(data, pos)
		if bytes.Equal(line, []byte("====")) {
			decoded, err := s.decodeUUBase64(data[start:pos], maxOutput)
			return decoded, next, err
		}
		pos = next
	}
	decoded, _ := s.decodeUUBase64(data[start:pos], maxOutput)
	return decoded, pos, ErrMalformed
}

func (s *scanState) decodeUUBase64(data []byte, maxOutput int64) ([]byte, error) {
	r := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
	decoded, err := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, maxOutput+1))
	if err != nil {
		if ctxErr := s.ctx.Err(); ctxErr != nil {
			return decoded, ctxErr
		}
		return decoded, ErrMalformed
	}
	if int64(len(decoded)) > maxOutput {
		return decoded[:maxOutput], ErrLimit
	}
	return decoded, nil
}
