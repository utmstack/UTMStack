// Package zipcodec decodes the ZIP compression methods the Go library lacks
// and the reference scanner decodes: Deflate64 (method 9) and implode
// (method 6). Both follow public descriptions (RFC 1951 with the Deflate64
// extensions, PKWARE's APPNOTE for implode); how damaged data ends is the
// caller's to decide from the errors.
package zipcodec

import "errors"

var (
	// ErrData: the data cannot be decoded.
	ErrData = errors.New("zipcodec: corrupt data")
	// ErrTruncated: the data ends before the stream does; the output so
	// far is returned with it.
	ErrTruncated = errors.New("zipcodec: data ends early")
	// ErrLimit: the output would exceed the caller's limit.
	ErrLimit = errors.New("zipcodec: output limit")
)

// bits reads a bit stream least significant bit first.
type bits struct {
	src  []byte
	pos  int
	buf  uint32
	n    uint
	over bool
}

func (b *bits) need(n uint) bool {
	for b.n < n {
		if b.pos >= len(b.src) {
			b.over = true
			return false
		}
		b.buf |= uint32(b.src[b.pos]) << b.n
		b.pos++
		b.n += 8
	}
	return true
}

func (b *bits) get(n uint) (uint32, bool) {
	if n == 0 {
		return 0, true
	}
	if !b.need(n) {
		return 0, false
	}
	v := b.buf & (1<<n - 1)
	b.buf >>= n
	b.n -= n
	return v, true
}

// huffman is a canonical code: the count of codes per length and the
// symbols in code order.
type huffman struct {
	count  [16]uint16
	symbol []uint16
}

// build makes a code from lengths, rejecting over-subscribed sets and, as
// zlib does, incomplete ones except a set of a single one-bit code (and never
// for the code-length code).
func build(lengths []uint8, codeLengths bool) (*huffman, bool) {
	h := &huffman{symbol: make([]uint16, len(lengths))}
	for _, l := range lengths {
		h.count[l]++
	}
	if int(h.count[0]) == len(lengths) {
		return h, !codeLengths // no codes: usable only if never read
	}
	left, max := 1, 0
	for l := 1; l < 16; l++ {
		left <<= 1
		left -= int(h.count[l])
		if left < 0 {
			return nil, false
		}
		if h.count[l] != 0 {
			max = l
		}
	}
	if left > 0 && (codeLengths || max != 1) {
		return nil, false
	}
	var offs [16]uint16
	for l := 1; l < 15; l++ {
		offs[l+1] = offs[l] + h.count[l]
	}
	for s, l := range lengths {
		if l != 0 {
			h.symbol[offs[l]] = uint16(s)
			offs[l]++
		}
	}
	return h, true
}

func (h *huffman) decode(b *bits) (int, bool) {
	code, first, index := 0, 0, 0
	for l := 1; l < 16; l++ {
		v, ok := b.get(1)
		if !ok {
			return 0, false
		}
		code |= int(v)
		count := int(h.count[l])
		if code-count < first {
			return int(h.symbol[index+code-first]), true
		}
		index += count
		first += count
		first <<= 1
		code <<= 1
	}
	return -1, true // no such code
}

var (
	lengthBase  = [29]uint16{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 3}
	lengthExtra = [29]uint8{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 16}
	distBase    = [32]uint32{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073,
		4097, 6145, 8193, 12289, 16385, 24577, 32769, 49153}
	distExtra = [32]uint8{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 14, 14}
	clOrder   = [19]uint8{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
)

// Inflate64 decodes a Deflate64 stream into at most limit bytes. On
// ErrTruncated and ErrData the output decoded so far is returned too.
func Inflate64(src []byte, limit int) ([]byte, error) {
	b := &bits{src: src}
	var out []byte
	fixed := fixedCodes()
	for {
		hdr, ok := b.get(3)
		if !ok {
			return out, ErrTruncated
		}
		var err error
		switch hdr >> 1 {
		case 0:
			out, err = stored(b, out, limit)
		case 1:
			out, err = codes(b, out, limit, fixed[0], fixed[1])
		case 2:
			var lit, dist *huffman
			if lit, dist, err = dynamic(b); err == nil {
				out, err = codes(b, out, limit, lit, dist)
			}
		default:
			err = ErrData
		}
		if err != nil {
			return out, err
		}
		if hdr&1 != 0 {
			return out, nil
		}
	}
}

func stored(b *bits, out []byte, limit int) ([]byte, error) {
	b.buf, b.n = 0, 0 // to the byte boundary
	if b.pos+4 > len(b.src) {
		return out, ErrTruncated
	}
	n := int(b.src[b.pos]) | int(b.src[b.pos+1])<<8
	if n != ^(int(b.src[b.pos+2])|int(b.src[b.pos+3])<<8)&0xffff {
		return out, ErrData
	}
	b.pos += 4
	if n > limit-len(out) {
		return out, ErrLimit
	}
	if b.pos+n > len(b.src) {
		return append(out, b.src[b.pos:]...), ErrTruncated
	}
	out = append(out, b.src[b.pos:b.pos+n]...)
	b.pos += n
	return out, nil
}

func fixedCodes() [2]*huffman {
	var l [288]uint8
	for i := range l {
		switch {
		case i < 144:
			l[i] = 8
		case i < 256:
			l[i] = 9
		case i < 280:
			l[i] = 7
		default:
			l[i] = 8
		}
	}
	lit, _ := build(l[:], false)
	var d [32]uint8
	for i := range d {
		d[i] = 5
	}
	dist, _ := build(d[:], false)
	return [2]*huffman{lit, dist}
}

func dynamic(b *bits) (*huffman, *huffman, error) {
	v, ok := b.get(14)
	if !ok {
		return nil, nil, ErrTruncated
	}
	nlen, ndist, ncode := int(v&31)+257, int(v>>5&31)+1, int(v>>10)+4
	if nlen > 286 {
		return nil, nil, ErrData // as zlib: too many length symbols
	}
	var cl [19]uint8
	for i := 0; i < ncode; i++ {
		x, ok := b.get(3)
		if !ok {
			return nil, nil, ErrTruncated
		}
		cl[clOrder[i]] = uint8(x)
	}
	clc, ok := build(cl[:], true)
	if !ok {
		return nil, nil, ErrData
	}
	lengths := make([]uint8, nlen+ndist)
	for i := 0; i < nlen+ndist; {
		sym, ok := clc.decode(b)
		if !ok {
			return nil, nil, ErrTruncated
		}
		switch {
		case sym < 0:
			return nil, nil, ErrData
		case sym < 16:
			lengths[i] = uint8(sym)
			i++
			continue
		}
		var prev uint8
		var n uint32
		switch sym {
		case 16:
			if i == 0 {
				return nil, nil, ErrData
			}
			prev = lengths[i-1]
			n, ok = b.get(2)
			n += 3
		case 17:
			n, ok = b.get(3)
			n += 3
		default:
			n, ok = b.get(7)
			n += 11
		}
		if !ok {
			return nil, nil, ErrTruncated
		}
		if i+int(n) > nlen+ndist {
			return nil, nil, ErrData
		}
		for ; n > 0; n-- {
			lengths[i] = prev
			i++
		}
	}
	if lengths[256] == 0 {
		return nil, nil, ErrData // no end-of-block code
	}
	lit, ok := build(lengths[:nlen], false)
	if !ok {
		return nil, nil, ErrData
	}
	dist, ok := build(lengths[nlen:], false)
	if !ok {
		return nil, nil, ErrData
	}
	return lit, dist, nil
}

func codes(b *bits, out []byte, limit int, lit, dist *huffman) ([]byte, error) {
	for {
		sym, ok := lit.decode(b)
		if !ok {
			return out, ErrTruncated
		}
		switch {
		case sym < 0 || sym > 285:
			return out, ErrData
		case sym < 256:
			if len(out) >= limit {
				return out, ErrLimit
			}
			out = append(out, byte(sym))
			continue
		case sym == 256:
			return out, nil
		}
		sym -= 257
		extra, ok := b.get(uint(lengthExtra[sym]))
		if !ok {
			return out, ErrTruncated
		}
		length := int(lengthBase[sym]) + int(extra)
		d, ok := dist.decode(b)
		if !ok {
			return out, ErrTruncated
		}
		if d < 0 || d > 31 {
			return out, ErrData
		}
		extra, ok = b.get(uint(distExtra[d]))
		if !ok {
			return out, ErrTruncated
		}
		distance := int(distBase[d]) + int(extra)
		if distance > len(out) {
			return out, ErrData // before the start of the output
		}
		if length > limit-len(out) {
			return out, ErrLimit
		}
		for i := 0; i < length; i++ {
			out = append(out, out[len(out)-distance])
		}
	}
}
