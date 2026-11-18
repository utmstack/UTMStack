package engine

import (
	"context"
	"encoding/binary"
	"hash/crc32"
)

// ARJ archives are a main header and file headers, each "60 EA", a basic
// header of at most 2600 bytes with its CRC-32 and optional extended headers,
// followed by the file data. Methods 1-3 share a static-Huffman LZ77 coder
// (the "ar002"/LHA family: blocks of code-length tables for characters and
// match lengths, and for match positions) with a 26624-byte window; method 4
// uses fixed variable-length codes. Each file's CRC-32 is checked, so a
// decoding difference can never turn into a clean verdict. The layout
// follows the ARJ technical notes shipped with ARJ (arj.txt).
const (
	arjMaxHeader = 2600
	arjDictSize  = 26624
	arjThreshold = 3
	arjMaxMatch  = 256
	arjNC        = 255 + arjMaxMatch + 2 - arjThreshold // characters and match lengths
	arjNP        = 17                                   // position slots
	arjNT        = 19                                   // code-length codes
	arjCBit      = 9
	arjTBit      = 5
	arjPBit      = 5

	arjGarbled    = 0x01
	arjVolume     = 0x04
	arjExtFile    = 0x08
	arjTypeBinary = 0
	arjTypeText   = 1
	arjTypeMain   = 2
)

func (s *scanState) arj(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	pos := 0
	for count := 0; ; count++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if count > s.maxFiles-s.files {
			return "", ErrLimit
		}
		if len(data)-pos < 4 || data[pos] != 0x60 || data[pos+1] != 0xea {
			return "", ErrMalformed
		}
		size := int(binary.LittleEndian.Uint16(data[pos+2:]))
		pos += 4
		if size == 0 {
			return "", incomplete // end of archive
		}
		if size < 30 || size > arjMaxHeader || len(data)-pos < size+4 {
			return "", ErrMalformed
		}
		basic := data[pos : pos+size]
		if crc32.ChecksumIEEE(basic) != binary.LittleEndian.Uint32(data[pos+size:]) || int(basic[0]) < 30 || int(basic[0]) > size {
			return "", ErrMalformed
		}
		pos += size + 4
		for {
			if len(data)-pos < 2 {
				return "", ErrMalformed
			}
			n := int(binary.LittleEndian.Uint16(data[pos:]))
			pos += 2
			if n == 0 {
				break
			}
			if len(data)-pos < n+4 {
				return "", ErrMalformed
			}
			pos += n + 4
		}
		fileType := basic[6]
		if count == 0 {
			if fileType != arjTypeMain {
				return "", ErrMalformed
			}
			continue // the main header has no data
		}
		stored := uint64(binary.LittleEndian.Uint32(basic[12:]))
		if stored > uint64(len(data)-pos) {
			return "", ErrMalformed
		}
		body := data[pos : pos+int(stored)]
		pos += int(stored)
		if fileType != arjTypeBinary && fileType != arjTypeText {
			continue // directories, labels and chapters hold no file data
		}
		flags, method := basic[4], basic[5]
		if flags&(arjGarbled|arjVolume|arjExtFile) != 0 || method > 4 {
			incomplete = ErrUnsupported // encrypted, split across volumes, or unknown
			continue
		}
		sig, err := s.arjMember(body, method, basic, depth)
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
}

func (s *scanState) arjMember(body []byte, method byte, basic []byte, depth int) (string, error) {
	size := int64(binary.LittleEndian.Uint32(basic[16:]))
	crc := binary.LittleEndian.Uint32(basic[20:])
	if size > min(s.maxFile, s.remaining) {
		return "", ErrLimit
	}
	var out []byte
	var err error
	switch method {
	case 0:
		out = body
		if int64(len(body)) != size {
			err = ErrMalformed
		}
	case 4:
		out, err = arjDecodeFast(body, int(size))
	default:
		d := &arjDecoder{arjBits: arjBits{in: body}, ctx: s.ctx}
		out, err = d.decode(int(size))
	}
	if method != 0 {
		s.spend(len(out)) // held while analyzed, or discarded expansion work
	}
	if err != nil {
		return "", err
	}
	var incomplete error
	if crc32.ChecksumIEEE(out) != crc {
		incomplete = ErrMalformed
	}
	prepaid := int64(len(out))
	if method == 0 {
		prepaid = 0
	}
	if sig, err := s.visitHeld(out, depth+1, prepaid); sig != "" || err != nil {
		return sig, err
	}
	return "", incomplete
}

// arjBits reads bits most significant first. Reading past the data yields
// zero bits, as the reference decoder does; overrun counts those bytes so a
// truncated stream is reported.
type arjBits struct {
	in        []byte
	pos       int
	overrun   int
	bitbuf    uint16
	subbitbuf byte
	bitcount  int
}

func (b *arjBits) fill(n int) {
	b.bitbuf <<= uint(n)
	for n > b.bitcount {
		n -= b.bitcount
		b.bitbuf |= uint16(b.subbitbuf) << uint(n)
		if b.pos < len(b.in) {
			b.subbitbuf = b.in[b.pos]
			b.pos++
		} else {
			b.subbitbuf = 0
			b.overrun++
		}
		b.bitcount = 8
	}
	b.bitcount -= n
	b.bitbuf |= uint16(b.subbitbuf >> uint(b.bitcount))
}

func (b *arjBits) get(n int) int {
	if n == 0 {
		return 0
	}
	x := int(b.bitbuf >> uint(16-n))
	b.fill(n)
	return x
}

func (b *arjBits) start() { b.fill(16) }

// The decoders read up to four bytes ahead of the last code while refilling
// their 16-bit buffers; reading further means the stream is truncated.
const arjMaxOverrun = 8

type arjDecoder struct {
	arjBits
	ctx context.Context
	// LH4-7 use the same block Huffman grammar with a different dictionary
	// width and position-code table. Zero values select ARJ's profile.
	dictSize, np, pBit int
	blocksize          int
	cLen               [arjNC]byte
	ptLen              [arjNT]byte
	cTable             [4096]uint16
	ptTable            [256]uint16
	left, right        [2*arjNC - 1]uint16
}

// arjInitialOutput is the output reserved before decoding, relative to the
// input size: the declared size is untrusted, so output beyond it grows as
// it is produced.
func arjInitialOutput(size, input int) []byte {
	return make([]byte, 0, min(size, max(64<<10, 8*input)))
}

func (d *arjDecoder) decode(size int) ([]byte, error) {
	if d.dictSize == 0 {
		d.dictSize, d.np, d.pBit = arjDictSize, arjNP, arjPBit
	}
	d.start()
	out := arjInitialOutput(size, len(d.in))
	nextContextCheck := 4096
	for len(out) < size {
		if len(out) >= nextContextCheck {
			if d.ctx != nil {
				if err := d.ctx.Err(); err != nil {
					return out, err
				}
			}
			nextContextCheck = len(out) + 4096
		}
		c, err := d.decodeC()
		if err != nil {
			return out, err
		}
		if c <= 255 {
			out = append(out, byte(c))
		} else {
			length := c - (256 - arjThreshold)
			distance, err := d.decodeP()
			if err != nil {
				return out, err
			}
			if distance >= d.dictSize || distance >= len(out) || length > size-len(out) {
				return out, ErrMalformed
			}
			from := len(out) - distance - 1
			for i := 0; i < length; i++ {
				out = append(out, out[from+i])
			}
		}
		if d.overrun > arjMaxOverrun {
			return out, ErrMalformed
		}
	}
	return out, nil
}

func (d *arjDecoder) decodeC() (int, error) {
	if d.blocksize == 0 {
		d.blocksize = d.get(16)
		if d.blocksize == 0 {
			d.blocksize = 1 << 16 // a 16-bit count that wraps
		}
		if err := d.readPtLen(arjNT, arjTBit, 3); err != nil {
			return 0, err
		}
		if err := d.readCLen(); err != nil {
			return 0, err
		}
		if err := d.readPtLen(d.np, d.pBit, -1); err != nil {
			return 0, err
		}
	}
	d.blocksize--
	j, err := d.walk(int(d.cTable[d.bitbuf>>4]), arjNC, 1<<3)
	if err != nil {
		return 0, err
	}
	d.fill(int(d.cLen[j]))
	return j, nil
}

func (d *arjDecoder) decodeP() (int, error) {
	j, err := d.walk(int(d.ptTable[d.bitbuf>>8]), d.np, 1<<7)
	if err != nil {
		return 0, err
	}
	d.fill(int(d.ptLen[j]))
	if j != 0 {
		j--
		j = 1<<uint(j) + d.get(j)
	}
	return j, nil
}

// walk follows tree nodes (indices at or above n) for codes longer than the
// lookup table, one bit per step from mask downwards.
func (d *arjDecoder) walk(j, n int, mask uint16) (int, error) {
	for j >= n {
		if mask == 0 || j >= len(d.left) {
			return 0, ErrMalformed
		}
		if d.bitbuf&mask != 0 {
			j = int(d.right[j])
		} else {
			j = int(d.left[j])
		}
		mask >>= 1
	}
	return j, nil
}

func (d *arjDecoder) readPtLen(nn, nbit, special int) error {
	n := d.get(nbit)
	if n == 0 {
		c := d.get(nbit)
		if c >= nn {
			return ErrMalformed
		}
		for i := 0; i < nn; i++ {
			d.ptLen[i] = 0
		}
		for i := range d.ptTable {
			d.ptTable[i] = uint16(c)
		}
		return nil
	}
	if n > nn {
		return ErrMalformed
	}
	i := 0
	for i < n {
		c := int(d.bitbuf >> 13)
		if c == 7 {
			for mask := uint16(1 << 12); mask&d.bitbuf != 0; mask >>= 1 {
				c++
			}
		}
		if c > 16 {
			return ErrMalformed
		}
		if c < 7 {
			d.fill(3)
		} else {
			d.fill(c - 3)
		}
		d.ptLen[i] = byte(c)
		i++
		if i == special {
			for zeros := d.get(2); zeros > 0; zeros-- {
				if i >= nn {
					return ErrMalformed
				}
				d.ptLen[i] = 0
				i++
			}
		}
	}
	for ; i < nn; i++ {
		d.ptLen[i] = 0
	}
	return d.makeTable(nn, d.ptLen[:nn], 8, d.ptTable[:])
}

func (d *arjDecoder) readCLen() error {
	n := d.get(arjCBit)
	if n == 0 {
		c := d.get(arjCBit)
		if c >= arjNC {
			return ErrMalformed
		}
		d.cLen = [arjNC]byte{}
		for i := range d.cTable {
			d.cTable[i] = uint16(c)
		}
		return nil
	}
	if n > arjNC {
		return ErrMalformed
	}
	i := 0
	for i < n {
		c, err := d.walk(int(d.ptTable[d.bitbuf>>8]), arjNT, 1<<7)
		if err != nil {
			return err
		}
		d.fill(int(d.ptLen[c]))
		if c <= 2 {
			switch c {
			case 0:
				c = 1
			case 1:
				c = d.get(4) + 3
			default:
				c = d.get(arjCBit) + 20
			}
			if c > arjNC-i {
				return ErrMalformed
			}
			for ; c > 0; c-- {
				d.cLen[i] = 0
				i++
			}
		} else {
			d.cLen[i] = byte(c - 2)
			i++
		}
	}
	for ; i < arjNC; i++ {
		d.cLen[i] = 0
	}
	return d.makeTable(arjNC, d.cLen[:], 12, d.cTable[:])
}

// makeTable builds a lookup table of tablebits bits for canonical Huffman
// codes, with left/right tree nodes for longer codes. The code lengths must
// form a complete code.
func (d *arjDecoder) makeTable(nchar int, bitlen []byte, tablebits int, table []uint16) error {
	var count, weight [17]uint32
	var start [18]uint32
	for _, l := range bitlen[:nchar] {
		if l > 16 {
			return ErrMalformed
		}
		count[l]++
	}
	for i := 1; i <= 16; i++ {
		start[i+1] = start[i] + count[i]<<uint(16-i)
	}
	if start[17] != 1<<16 {
		return ErrMalformed
	}
	jutbits := uint(16 - tablebits)
	for i := 1; i <= tablebits; i++ {
		start[i] >>= jutbits
		weight[i] = 1 << uint(tablebits-i)
	}
	for i := tablebits + 1; i <= 16; i++ {
		weight[i] = 1 << uint(16-i)
	}
	if i := start[tablebits+1] >> jutbits; i != 1<<16 {
		for k := uint32(1) << uint(tablebits); i < k; i++ {
			table[i] = 0
		}
	}
	avail := nchar
	mask := uint32(1) << uint(15-tablebits)
	for ch := 0; ch < nchar; ch++ {
		l := int(bitlen[ch])
		if l == 0 {
			continue
		}
		next := start[l] + weight[l]
		if l <= tablebits {
			if next > uint32(len(table)) {
				return ErrMalformed
			}
			for i := start[l]; i < next; i++ {
				table[i] = uint16(ch)
			}
		} else {
			k := start[l]
			p := &table[k>>jutbits]
			for i := l - tablebits; i > 0; i-- {
				if *p == 0 {
					if avail >= len(d.left) {
						return ErrMalformed
					}
					d.left[avail], d.right[avail] = 0, 0
					*p = uint16(avail)
					avail++
				}
				if int(*p) >= len(d.left) {
					return ErrMalformed
				}
				if k&mask != 0 {
					p = &d.right[*p]
				} else {
					p = &d.left[*p]
				}
				k <<= 1
			}
			*p = uint16(ch)
		}
		start[l] = next
	}
	return nil
}

// arjDecodeFast decodes method 4: a literal is a 0 bit and eight bits; a
// match is a unary-prefixed length and a unary-prefixed position.
func arjDecodeFast(in []byte, size int) ([]byte, error) {
	b := arjBits{in: in}
	b.start()
	var getbuf uint16
	getlen := 0
	refill := func() {
		getbuf |= b.bitbuf >> uint(getlen)
		b.fill(16 - getlen)
		getlen = 16
	}
	bit := func() int {
		if getlen <= 0 {
			refill()
		}
		c := int(getbuf >> 15)
		getbuf <<= 1
		getlen--
		return c
	}
	bits := func(n int) int {
		if getlen < n {
			refill()
		}
		c := int(getbuf >> uint(16-n))
		getbuf <<= uint(n)
		getlen -= n
		return c
	}
	variable := func(first, stop int) int {
		plus, pwr, width := 0, 1<<uint(first), first
		for ; width < stop; width++ {
			if bit() == 0 {
				break
			}
			plus += pwr
			pwr <<= 1
		}
		c := 0
		if width != 0 {
			c = bits(width)
		}
		return c + plus
	}
	out := arjInitialOutput(size, len(in))
	for len(out) < size {
		if b.overrun > arjMaxOverrun {
			return out, ErrMalformed
		}
		c := variable(0, 7)
		if c == 0 {
			out = append(out, byte(bits(8)))
			continue
		}
		length := c - 1 + arjThreshold
		distance := variable(9, 13)
		if distance >= arjDictSize || distance >= len(out) || length > size-len(out) {
			return out, ErrMalformed
		}
		from := len(out) - distance - 1
		for i := 0; i < length; i++ {
			out = append(out, out[from+i])
		}
	}
	return out, nil
}
