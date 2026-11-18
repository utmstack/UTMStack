package engine

import (
	"bytes"
	"encoding/binary"
)

// SZDD is the LZSS format of Microsoft's COMPRESS.EXE and EXPAND.EXE (the
// "file.ex_" installer convention): an 8-byte magic, the mode 'A', the
// replaced last character of the name and the 32-bit expanded length, then
// groups of a flag byte and eight items. A set flag bit is a literal byte;
// a clear bit is a 12-bit window position and 4-bit length (plus three) in a
// 4096-byte window that starts filled with spaces, writing from 4080, as
// libmspack's SZDD format notes describe.
var szddMagic = []byte("SZDD\x88\xf0\x27\x33")

func (s *scanState) szdd(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	if len(data) < 14 || !bytes.HasPrefix(data, szddMagic) {
		return "", ErrMalformed
	}
	if data[8] != 'A' {
		return "", ErrUnsupported
	}
	size := int64(binary.LittleEndian.Uint32(data[10:14]))
	if size > min(s.maxFile, s.remaining) {
		return "", ErrLimit
	}
	out, err := szddExpand(data[14:], int(size))
	if err != nil {
		s.spend(len(out))
		return "", err
	}
	return s.visit(out, depth+1)
}

// szddExpand decodes at most size bytes. Input that ends before the declared
// length is truncated; input left after it is not SZDD data.
func szddExpand(in []byte, size int) ([]byte, error) {
	var window [4096]byte
	for i := range window {
		window[i] = ' '
	}
	pos := 4096 - 16
	// A group of 17 input bytes expands to at most 144 bytes, so the declared
	// size reserves no more than the input can produce.
	out := make([]byte, 0, min(size, 9*len(in)))
	i := 0
	for len(out) < size {
		if i >= len(in) {
			return out, ErrMalformed
		}
		flags := in[i]
		i++
		for bit := 0; bit < 8 && len(out) < size; bit++ {
			if flags&(1<<bit) != 0 {
				if i >= len(in) {
					return out, ErrMalformed
				}
				c := in[i]
				i++
				out = append(out, c)
				window[pos] = c
				pos = (pos + 1) & 4095
				continue
			}
			if len(in)-i < 2 {
				return out, ErrMalformed
			}
			match := int(in[i]) | int(in[i+1]&0xf0)<<4
			length := int(in[i+1]&0x0f) + 3
			i += 2
			for j := 0; j < length && len(out) < size; j++ {
				c := window[(match+j)&4095]
				out = append(out, c)
				window[pos] = c
				pos = (pos + 1) & 4095
			}
		}
	}
	// A final group may be padded with unused items; more data than that
	// does not belong to the declared file.
	if len(in)-i > 16 {
		return out, ErrMalformed
	}
	return out, nil
}
