package exr154

import "math"

type pizBits struct {
	data     []byte
	pos, end int
	err      error
}

func (b *pizBits) get(n int) uint64 {
	if b.err != nil {
		return 0
	}
	if n < 0 || n > b.end-b.pos {
		b.err = ErrFormat
		return 0
	}
	var v uint64
	for i := 0; i < n; i++ {
		v = v<<1 | uint64(b.data[b.pos/8]>>uint(7-b.pos%8)&1)
		b.pos++
	}
	return v
}
func pizHuffman(data []byte, expected int, l Limits) ([]uint16, error) {
	if len(data) < 20 {
		return nil, ErrFormat
	}
	minCode, maxCode := int(le.Uint32(data)), int(le.Uint32(data[4:]))
	declaredBits := uint64(le.Uint32(data[12:]))
	if minCode > maxCode || minCode > 65536 || maxCode > 65536 || declaredBits > math.MaxInt32 {
		return nil, ErrFormat
	}
	data = data[20:]
	if declaredBits > uint64(len(data))*8 {
		return nil, ErrFormat
	}
	b := pizBits{data: data, end: len(data) * 8}
	lengths := make([]uint8, maxCode+1)
	for at := minCode; at <= maxCode; {
		if err := l.charge(1); err != nil {
			return nil, err
		}
		n := int(b.get(6))
		if b.err != nil {
			return nil, b.err
		}
		switch {
		case n == 63:
			run := int(b.get(8)) + 6
			if b.err != nil || run > maxCode+1-at {
				return nil, ErrFormat
			}
			at += run
		case n >= 59:
			run := n - 59 + 2
			if run > maxCode+1-at {
				return nil, ErrFormat
			}
			at += run
		default:
			lengths[at] = uint8(n)
			at++
		}
	}
	// The declared table length is advisory; consume the actual packed table.
	consumed := (b.pos + 7) / 8
	data = data[consumed:]
	if declaredBits > uint64(len(data))*8 {
		return nil, ErrFormat
	}
	padding := int((8 - declaredBits) & 7)
	b = pizBits{data: data, end: len(data)*8 - padding}
	var counts [59]uint64
	var symbols [59][]uint32
	for value, n := range lengths {
		if n != 0 {
			counts[n]++
			symbols[n] = append(symbols[n], uint32(value))
		}
	}
	var first [59]uint64
	var code uint64
	for n := 58; n > 0; n-- {
		first[n] = code
		if code+counts[n] > uint64(1)<<uint(n) {
			return nil, ErrFormat
		}
		code = (code + counts[n]) >> 1
	}
	if err := l.charge(int64(len(lengths))/8 + 1); err != nil {
		return nil, err
	}
	out := make([]uint16, 0, expected)
	for b.pos < b.end {
		var code uint64
		symbol := -1
		n := 1
		for ; n <= 58; n++ {
			code = code<<1 | b.get(1)
			if b.err != nil {
				return nil, b.err
			}
			if code >= first[n] && code-first[n] < counts[n] {
				symbol = int(symbols[n][code-first[n]])
				break
			}
		}
		if symbol < 0 {
			return nil, ErrFormat
		}
		if err := l.charge(int64(n) + 1); err != nil {
			return nil, err
		}
		if symbol == maxCode {
			count := int(b.get(8))
			if b.err != nil || len(out) == 0 || count > expected-len(out) {
				return nil, ErrFormat
			}
			v := out[len(out)-1]
			for i := 0; i < count; i++ {
				out = append(out, v)
			}
		} else {
			if symbol > 65535 || len(out) >= expected {
				return nil, ErrFormat
			}
			out = append(out, uint16(symbol))
		}
	}
	if len(out) != expected {
		return nil, ErrFormat
	}
	return out, nil
}
func unpackPIZ(data []byte, h header, w, ht, expected int, l Limits) ([]byte, error) {
	c := cursor{data: data}
	b := c.take(4)
	if b == nil {
		return nil, c.err
	}
	lo, hi := int(le.Uint16(b)), int(le.Uint16(b[2:]))
	if lo >= 8192 || hi >= 8192 {
		return nil, ErrFormat
	}
	var bitmap [8192]byte
	if lo <= hi {
		b = c.take(hi - lo + 1)
		if b == nil {
			return nil, c.err
		}
		copy(bitmap[lo:hi+1], b)
	}
	c.take(4) // PIZ's declared Huffman length is advisory in this read profile.
	if c.err != nil {
		return nil, c.err
	}
	lookup := make([]uint16, 65536)
	count := 0
	for v := 0; v < 65536; v++ {
		if v == 0 || bitmap[v>>3]&(1<<uint(v&7)) != 0 {
			lookup[count] = uint16(v)
			count++
		}
	}
	if err := l.charge(8193); err != nil {
		return nil, err
	}
	values, err := pizHuffman(data[c.pos:], expected/2, l)
	if err != nil {
		return nil, err
	}
	starts := make([]int, len(h.channels))
	start := 0
	for ci, ch := range h.channels {
		starts[ci] = start
		stride := ch.size / 2
		words := w * ht * stride
		for lane := 0; lane < stride; lane++ {
			if err := pizWavelet(values[start+lane:start+words], w, ht, stride, w*stride, count-1, l); err != nil {
				return nil, err
			}
		}
		start += words
	}
	for i, v := range values {
		if i&4095 == 0 {
			if err := l.charge(513); err != nil {
				return nil, err
			}
		}
		values[i] = lookup[v]
	}
	out := make([]byte, 0, expected)
	for y := 0; y < ht; y++ {
		if err := l.charge(int64(w)*int64(h.bpp)/8 + 1); err != nil {
			return nil, err
		}
		for ci, ch := range h.channels {
			words := w * ch.size / 2
			at := starts[ci] + y*words
			for _, v := range values[at : at+words] {
				out = le.AppendUint16(out, v)
			}
		}
	}
	return out, nil
}
func pizPair(low, high uint16, narrow bool) (uint16, uint16) {
	if narrow {
		h := int32(int16(high))
		a := int32(int16(low)) + (h & 1) + (h >> 1)
		return uint16(a), uint16(a - h)
	}
	m, d := int32(low), int32(high)
	b := (m - (d >> 1)) & 65535
	a := (d + b - 32768) & 65535
	return uint16(a), uint16(b)
}
func pizWavelet(buf []uint16, w, h, sx, sy, maxValue int, l Limits) error {
	n := min(w, h)
	p := 1
	for p <= n {
		p <<= 1
	}
	p >>= 1
	p2 := p
	p >>= 1
	narrow := maxValue < 1<<14
	for p >= 1 {
		dx, dy := sx*p, sy*p
		xx, yy := sx*p2, sy*p2
		y := 0
		for ; y <= sy*(h-p2); y += yy {
			if err := l.charge(int64(w/p2) + 1); err != nil {
				return err
			}
			x := y
			for ; x <= y+sx*(w-p2); x += xx {
				a, c := pizPair(buf[x], buf[x+dy], narrow)
				b, d := pizPair(buf[x+dx], buf[x+dy+dx], narrow)
				a, b = pizPair(a, b, narrow)
				c, d = pizPair(c, d, narrow)
				buf[x], buf[x+dx], buf[x+dy], buf[x+dy+dx] = a, b, c, d
			}
			if w&p != 0 {
				buf[x], buf[x+dy] = pizPair(buf[x], buf[x+dy], narrow)
			}
		}
		if h&p != 0 {
			for x := y; x <= y+sx*(w-p2); x += xx {
				buf[x], buf[x+dx] = pizPair(buf[x], buf[x+dx], narrow)
			}
		}
		p2 = p
		p >>= 1
	}
	return nil
}
