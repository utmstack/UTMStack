package engine

import (
	"bytes"
	"encoding/binary"
)

// BinHex 4.0 (RFC 1741) encodes a Macintosh file as six-bit characters
// between two colons, after a line naming the format. The decoded stream uses
// run-length encoding (0x90 marks a repeat count) and holds a header with the
// name and fork lengths, then the data fork and the resource fork, each
// followed by a CRC-16/XMODEM.
// https://www.rfc-editor.org/rfc/rfc1741
const binhexAlphabet = "!\"#$%&'()*+,-012345689@ABCDEFGHIJKLMNPQRSTUVXYZ[`abcdefhijklmpqr"

var (
	binhexMarker = []byte("(This file must be converted with BinHex")
	binhexValues = func() (v [256]int8) {
		for i := range v {
			v[i] = -1
		}
		for i := 0; i < len(binhexAlphabet); i++ {
			v[binhexAlphabet[i]] = int8(i)
		}
		return v
	}()
)

// binhexStart finds encoded data: the marker at the start of a line, then a
// line starting with a colon within the next few lines. Prose that merely
// mentions the marker is not BinHex data.
func binhexStart(data []byte) int {
	for pos := 0; pos < len(data); {
		i := bytes.Index(data[pos:], binhexMarker)
		if i < 0 {
			return -1
		}
		i += pos
		pos = i + len(binhexMarker)
		if i > 0 && data[i-1] != '\n' && data[i-1] != '\r' {
			continue
		}
		line, end := pos, min(len(data), pos+512)
		for lines := 0; lines < 4; lines++ {
			nl := bytes.IndexByte(data[line:end], '\n')
			if nl < 0 {
				break
			}
			line += nl + 1
			first := line
			for first < end && (data[first] == ' ' || data[first] == '\t' || data[first] == '\r') {
				first++
			}
			if first < end && data[first] == ':' {
				return first + 1
			}
		}
	}
	return -1
}

func (s *scanState) binhex(data []byte, start, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	limit := min(s.maxFile, s.remaining)
	if limit < 0 {
		return "", ErrLimit
	}
	decoded, err := binhexDecode(data[start:], limit)
	s.spend(len(decoded)) // held while its forks are analyzed
	if err != nil {
		return "", err
	}
	if len(decoded) < 1 {
		return "", ErrMalformed
	}
	nameLen := int(decoded[0])
	header := 1 + nameLen + 1 + 4 + 4 + 2 + 4 + 4
	if len(decoded) < header+2 {
		return "", ErrMalformed
	}
	var incomplete error
	if binary.BigEndian.Uint16(decoded[header:]) != crcXMODEM(decoded[:header]) {
		incomplete = ErrMalformed
	}
	dataLen := uint64(binary.BigEndian.Uint32(decoded[header-8:]))
	resourceLen := uint64(binary.BigEndian.Uint32(decoded[header-4:]))
	pos := uint64(header + 2)
	for _, n := range []uint64{dataLen, resourceLen} {
		if n+2 > uint64(len(decoded))-pos {
			return "", ErrMalformed
		}
		fork := decoded[pos : pos+n]
		if binary.BigEndian.Uint16(decoded[pos+n:]) != crcXMODEM(fork) {
			incomplete = ErrMalformed
		}
		pos += n + 2
		if n == 0 {
			continue
		}
		sig, err := s.visitHeld(fork, depth+1, int64(n))
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// binhexDecode returns at most limit bytes of the run-length decoded stream.
// White space between characters is ignored; the data ends at a colon.
func binhexDecode(text []byte, limit int64) ([]byte, error) {
	var out []byte
	var acc uint32
	bits := 0
	repeat, last := false, byte(0)
	emit := func(c byte) error {
		switch {
		case repeat:
			repeat = false
			if c == 0 {
				if int64(len(out)) >= limit {
					return ErrLimit
				}
				out = append(out, 0x90)
				last = 0x90
				return nil
			}
			if len(out) == 0 {
				return ErrMalformed
			}
			if int64(len(out))+int64(c)-1 > limit {
				return ErrLimit
			}
			for i := 1; i < int(c); i++ {
				out = append(out, last)
			}
		case c == 0x90:
			repeat = true
		default:
			if int64(len(out)) >= limit {
				return ErrLimit
			}
			out = append(out, c)
			last = c
		}
		return nil
	}
	for _, c := range text {
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		case ':':
			if repeat {
				return out, ErrMalformed
			}
			return out, nil
		}
		v := binhexValues[c]
		if v < 0 {
			return out, ErrMalformed
		}
		acc = acc<<6 | uint32(v)
		bits += 6
		if bits >= 8 {
			bits -= 8
			if err := emit(byte(acc >> bits)); err != nil {
				return out, err
			}
			acc &= 1<<bits - 1
		}
	}
	return out, ErrMalformed // no closing colon
}

func crcXMODEM(b []byte) uint16 {
	var crc uint16
	for _, c := range b {
		crc ^= uint16(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
