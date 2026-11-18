package engine

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"io"
)

// KWAJ is Microsoft's COMPRESS/EXPAND container. The format and its five
// compression methods are specified by libmspack's szdd_kwaj_format.html.
// All decoders below are bounded by the same child and aggregate scan limits
// as the other archive formats.
var kwajMagic = []byte("KWAJ\x88\xf0\x27\xd1")

func (s *scanState) kwaj(data []byte, depth int) (string, error) {
	limit, err := s.childBudget(depth)
	if err != nil {
		return "", err
	}
	if len(data) < 14 || !bytes.HasPrefix(data, kwajMagic) {
		return "", ErrMalformed
	}
	method := binary.LittleEndian.Uint16(data[8:10])
	offset := int(binary.LittleEndian.Uint16(data[10:12]))
	flags := binary.LittleEndian.Uint16(data[12:14])
	if offset < 14 || offset > len(data) {
		return "", ErrMalformed
	}
	if flags&^0x3f != 0 {
		return "", ErrUnsupported
	}
	pos := 14
	var declared int64 = -1
	read := func(n int) ([]byte, bool) {
		if n < 0 || n > len(data)-pos {
			return nil, false
		}
		b := data[pos : pos+n]
		pos += n
		return b, true
	}
	if flags&1 != 0 {
		b, ok := read(4)
		if !ok {
			return "", ErrMalformed
		}
		declared = int64(binary.LittleEndian.Uint32(b))
		if declared > limit {
			return "", ErrLimit
		}
	}
	if flags&2 != 0 {
		if _, ok := read(2); !ok {
			return "", ErrMalformed
		}
	}
	for _, bit := range []uint16{4, 8, 16, 32} {
		if flags&bit == 0 {
			continue
		}
		switch bit {
		case 4, 32:
			b, ok := read(2)
			if !ok {
				return "", ErrMalformed
			}
			if _, ok = read(int(binary.LittleEndian.Uint16(b))); !ok {
				return "", ErrMalformed
			}
		case 8, 16:
			maxName := 9
			if bit == 16 {
				maxName = 4
			}
			// Some real KWAJ headers place the terminating NUL just
			// past data_offset. The reference reader parses the string
			// there, then seeks back to data_offset for decompression.
			b, ok := read(min(maxName, len(data)-pos))
			if !ok {
				return "", ErrMalformed
			}
			end := bytes.IndexByte(b, 0)
			if end < 0 {
				return "", ErrMalformed
			}
			pos -= len(b) - end - 1
		}
	}
	if method > 4 {
		return "", ErrUnsupported
	}
	packed := data[offset:]
	var out []byte
	switch method {
	case 0:
		if int64(len(packed)) > limit {
			return "", ErrLimit
		}
		out = packed
	case 1:
		if int64(len(packed)) > limit {
			return "", ErrLimit
		}
		out = make([]byte, len(packed))
		for i, b := range packed {
			out[i] = b ^ 0xff
		}
	case 2:
		out, err = kwajLZSS(s, packed, limit, declared)
	case 3:
		out, err = kwajLZH(s, packed, limit, declared)
	case 4:
		out, err = kwajMSZIP(s, packed, limit)
	}
	// The optional length is advisory; libmspack decodes to stream EOF.
	// A false short length must not hide bytes after it. Analyze available
	// output even when a late decoder error or length mismatch occurs.
	if err != nil || declared >= 0 && int64(len(out)) != declared {
		if len(out) >= minAnalyzedObject && int64(len(out)) <= limit {
			if sig, _ := s.visit(out, depth+1); sig != "" {
				return sig, nil
			}
		} else {
			s.spend(len(out))
		}
		if err != nil {
			return "", err
		}
		return "", ErrMalformed
	}
	return s.visit(out, depth+1)
}

func kwajLZSS(s *scanState, in []byte, limit, declared int64) ([]byte, error) {
	var window [4096]byte
	for i := range window {
		window[i] = ' '
	}
	pos := 4096 - 18
	var out []byte
	appendByte := func(b byte) error {
		if int64(len(out)) >= limit {
			return ErrLimit
		}
		out = append(out, b)
		window[pos] = b
		pos = (pos + 1) & 4095
		return nil
	}
	for i := 0; i < len(in); {
		if err := s.ctx.Err(); err != nil {
			return out, err
		}
		flags := in[i]
		i++
		for bit := 0; bit < 8; bit++ {
			if i == len(in) {
				return out, nil // final flag group can end before all eight items
			}
			if flags&(1<<bit) != 0 {
				if err := appendByte(in[i]); err != nil {
					return out, err
				}
				i++
				continue
			}
			if len(in)-i < 2 {
				return out, ErrMalformed
			}
			match := int(in[i]) | int(in[i+1]&0xf0)<<4
			length := int(in[i+1]&0x0f) + 3
			i += 2
			for j := 0; j < length; j++ {
				if err := appendByte(window[(match+j)&4095]); err != nil {
					return out, err
				}
			}
		}
	}
	if declared >= 0 && int64(len(out)) != declared {
		return out, ErrMalformed
	}
	return out, nil
}

func kwajMSZIP(s *scanState, in []byte, limit int64) ([]byte, error) {
	var out []byte
	for len(in) > 0 {
		if err := s.ctx.Err(); err != nil {
			return out, err
		}
		if len(in) == 2 && in[0] == 0 && in[1] == 0 {
			return out, nil
		}
		if len(in) < 4 {
			return out, ErrMalformed
		}
		n := int(binary.LittleEndian.Uint16(in))
		if n < 2 || n > len(in)-2 || n > 32780 || in[2] != 'C' || in[3] != 'K' {
			return out, ErrMalformed
		}
		block := in[4 : 2+n]
		in = in[2+n:]
		history := out[max(0, len(out)-32768):]
		src := bytes.NewReader(block)
		r := flate.NewReaderDict(src, history)
		maxBlock := min(int64(32768), limit-int64(len(out)))
		decoded, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, maxBlock+1))
		closeErr := r.Close()
		out = append(out, decoded...)
		if err := s.ctx.Err(); err != nil {
			return out, err
		}
		if int64(len(decoded)) > maxBlock {
			return out, ErrLimit
		}
		if readErr != nil || closeErr != nil || src.Len() != 0 || len(decoded) == 0 {
			return out, ErrMalformed
		}
	}
	return out, ErrMalformed // MSZIP stream requires the terminating zero word
}

type kwajBits struct {
	data []byte
	pos  int
}

func (r *kwajBits) read(n int) (int, bool) {
	if n < 0 || n > 16 || n > len(r.data)*8-r.pos {
		return 0, false
	}
	var v int
	for i := 0; i < n; i++ {
		v = v<<1 | int(r.data[r.pos/8]>>(7-r.pos%8)&1)
		r.pos++
	}
	return v, true
}

func (r *kwajBits) paddingFrom(start int) bool {
	if len(r.data)*8-start > 7 {
		return false
	}
	for i := start; i < len(r.data)*8; i++ {
		if r.data[i/8]&(1<<(7-i%8)) != 0 {
			return false
		}
	}
	return true
}

type kwajTree struct {
	byLength [16]map[int]int
	maxLen   int
}

func kwajHuffman(r *kwajBits, typ, count int) (kwajTree, error) {
	var lens [256]int
	if typ == 0 {
		length := 0
		for 1<<length < count {
			length++
		}
		for i := 0; i < count; i++ {
			lens[i] = length
		}
	} else {
		prev := 0
		for i := 0; i < count; i++ {
			n, ok := 0, true
			switch typ {
			case 1:
				if i != 0 {
					b, yes := r.read(1)
					if !yes {
						return kwajTree{}, ErrMalformed
					}
					if b == 0 {
						n = prev
						break
					}
					b, yes = r.read(1)
					if !yes {
						return kwajTree{}, ErrMalformed
					}
					if b == 0 {
						n = prev + 1
						break
					}
				}
				n, ok = r.read(4)
			case 2:
				if i != 0 {
					selector, yes := r.read(2)
					if !yes {
						return kwajTree{}, ErrMalformed
					}
					if selector != 3 {
						n = prev + selector - 1
						break
					}
				}
				n, ok = r.read(4)
			case 3:
				n, ok = r.read(4)
			default:
				return kwajTree{}, ErrMalformed
			}
			if !ok {
				return kwajTree{}, ErrMalformed
			}
			if n < 0 || n > 15 {
				return kwajTree{}, ErrMalformed
			}
			lens[i], prev = n, n
		}
	}
	var counts [16]int
	for _, n := range lens[:count] {
		if n != 0 {
			counts[n]++
		}
	}
	code := 0
	var tree kwajTree
	for length := 1; length <= 15; length++ {
		code = (code + counts[length-1]) << 1
		if code+counts[length] > 1<<length {
			return kwajTree{}, ErrMalformed
		}
		if counts[length] != 0 {
			tree.maxLen = length
			tree.byLength[length] = make(map[int]int, counts[length])
		}
		next := code
		for symbol := 0; symbol < count; symbol++ {
			if lens[symbol] == length {
				tree.byLength[length][next] = symbol
				next++
			}
		}
	}
	if tree.maxLen == 0 {
		return kwajTree{}, ErrMalformed
	}
	return tree, nil
}

func (t kwajTree) symbol(r *kwajBits) (int, bool) {
	code := 0
	for n := 1; n <= t.maxLen; n++ {
		bit, ok := r.read(1)
		if !ok {
			return 0, false
		}
		code = code<<1 | bit
		if symbol, ok := t.byLength[n][code]; ok {
			return symbol, true
		}
	}
	return 0, false
}

func kwajLZH(s *scanState, in []byte, limit, declared int64) ([]byte, error) {
	r := kwajBits{data: in}
	var types [5]int
	for i := 0; i < 6; i++ {
		v, ok := r.read(4)
		if !ok || v > 3 {
			return nil, ErrMalformed
		}
		if i < 5 {
			types[i] = v
		}
	}
	counts := [5]int{16, 16, 32, 64, 256}
	var trees [5]kwajTree
	for i := range trees {
		tree, err := kwajHuffman(&r, types[i], counts[i])
		if err != nil {
			return nil, err
		}
		trees[i] = tree
	}
	var window [4096]byte
	for i := range window {
		window[i] = ' '
	}
	pos, selected := 4096-17, 0
	var out []byte
	put := func(b byte) error {
		if int64(len(out)) >= limit {
			return ErrLimit
		}
		out = append(out, b)
		window[pos] = b
		pos = (pos + 1) & 4095
		return nil
	}
	for r.pos < len(in)*8 {
		if err := s.ctx.Err(); err != nil {
			return out, err
		}
		start := r.pos
		length, ok := trees[selected].symbol(&r)
		if !ok {
			if r.paddingFrom(start) {
				break // unused pad bits in the last byte
			}
			return out, ErrMalformed
		}
		if length != 0 {
			x, ok := trees[3].symbol(&r)
			if !ok {
				if r.paddingFrom(start) {
					break
				}
				return out, ErrMalformed
			}
			y, ok := r.read(6)
			if !ok {
				if r.paddingFrom(start) {
					break
				}
				return out, ErrMalformed
			}
			offset := x<<6 | y
			for i := 0; i < length+2; i++ {
				if err := put(window[(pos-offset)&4095]); err != nil {
					return out, err
				}
			}
			selected = 0
			continue
		}
		run, ok := trees[2].symbol(&r)
		if !ok {
			if r.paddingFrom(start) {
				break
			}
			return out, ErrMalformed
		}
		if run != 31 {
			selected = 1
		}
		for i := 0; i <= run; i++ {
			v, ok := trees[4].symbol(&r)
			if !ok {
				return out, ErrMalformed
			}
			if err := put(byte(v)); err != nil {
				return out, err
			}
		}
	}
	if declared >= 0 && int64(len(out)) != declared {
		return out, ErrMalformed
	}
	return out, nil
}
