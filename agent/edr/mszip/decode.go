// SPDX-License-Identifier: LGPL-2.1-only

// Package mszip decodes CAB MSZIP streams, including libmspack's repair policy.
// The adaptation of Stuart Caie's implementation uses bounded Go storage and
// cancellable input. See NOTICE.md and LICENSE.LGPL-2.1.
package mszip

import (
	"context"
	"errors"
	"io"
)

var (
	ErrFormat = errors.New("invalid MSZIP compressed data")
	ErrInput  = errors.New("MSZIP input ended or failed")
	// ErrLimit lets a bounded input source abort without manufacturing repair
	// output from a resource-policy failure.
	ErrLimit = errors.New("MSZIP resource limit")
)

const frameSize = 32768

var lengthBase = [...]int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
var lengthExtra = [...]uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
var distanceBase = [...]int{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577}
var distanceExtra = [...]uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}
var codeOrder = [...]int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}

type cursor struct {
	owner    *decoder
	pos, end int
	buffer   uint32
	left     uint
}

type decoder struct {
	ctx      context.Context
	input    io.Reader
	bytes    [4096]byte
	bits     cursor
	padded   bool
	fatal    error
	window   [frameSize]byte
	pos      int
	flushed  int
	repaired int
	steps    uint
}

func (d *decoder) refill() error {
	if err := d.ctx.Err(); err != nil {
		d.fatal = err
		return err
	}
	if d.fatal != nil {
		return d.fatal
	}
	n, err := d.input.Read(d.bytes[:])
	if n < 0 || n > len(d.bytes) || err != nil && !errors.Is(err, io.EOF) {
		if errors.Is(err, ErrLimit) {
			d.fatal = ErrLimit
		} else {
			d.fatal = ErrInput
		}
		return d.fatal
	}
	if n == 0 {
		if d.padded {
			d.fatal = ErrInput
			return d.fatal
		}
		d.padded = true
		d.bytes[0], d.bytes[1] = 0, 0
		n = 2 // the reference permits one two-byte EOF lookahead
	}
	d.bits.pos, d.bits.end = 0, n
	return nil
}

func (b *cursor) byte() (byte, error) {
	if b.pos >= b.end {
		if err := b.owner.refill(); err != nil {
			return 0, err
		}
		b.pos, b.end = b.owner.bits.pos, b.owner.bits.end
	}
	v := b.owner.bytes[b.pos]
	b.pos++
	return v, nil
}

func (b *cursor) ensure(n uint) error {
	for b.left < n {
		v, err := b.byte()
		if err != nil {
			return err
		}
		b.buffer |= uint32(v) << b.left
		b.left += 8
	}
	return nil
}

func (b *cursor) drop(n uint) {
	b.buffer >>= n
	b.left -= n
}

func (b *cursor) read(n uint) (int, error) {
	if err := b.ensure(n); err != nil {
		return 0, err
	}
	v := int(b.buffer & ((1 << n) - 1))
	b.drop(n)
	return v, nil
}

type huffman struct {
	count, first, start [17]int
	symbols             []int
}

// Unlike a general DEFLATE reader, libmspack requires a complete tree even
// when a single symbol or an unused distance alphabet would suffice.
func makeHuffman(lengths []uint8) (huffman, error) {
	var h huffman
	for _, n := range lengths {
		if n > 16 {
			return h, ErrFormat
		}
		h.count[n]++
	}
	left, code, offset := 1, 0, 0
	for n := 1; n <= 16; n++ {
		left = left*2 - h.count[n]
		if left < 0 {
			return h, ErrFormat
		}
		code = (code + h.count[n-1]) << 1
		if n == 1 {
			code = 0 // zero-length symbols have no canonical code
		}
		h.first[n], h.start[n] = code, offset
		offset += h.count[n]
	}
	if left != 0 {
		return h, ErrFormat
	}
	h.symbols = make([]int, offset)
	next := h.start
	for symbol, n := range lengths {
		if n != 0 {
			h.symbols[next[n]] = symbol
			next[n]++
		}
	}
	return h, nil
}

func (h *huffman) symbol(b *cursor, prefetch uint) (int, error) {
	if err := b.ensure(prefetch); err != nil {
		return 0, err
	}
	code := 0
	for n := 1; n <= int(prefetch); n++ {
		code = code<<1 | int(b.buffer>>uint(n-1)&1)
		index := code - h.first[n]
		if index >= 0 && index < h.count[n] {
			b.drop(uint(n))
			return h.symbols[h.start[n]+index], nil
		}
	}
	return 0, ErrFormat
}

func (d *decoder) check() error {
	d.steps++
	if d.steps&4095 == 0 {
		return d.ctx.Err()
	}
	return nil
}

func (d *decoder) flush(n int) error {
	d.flushed += n
	if d.flushed > frameSize {
		return ErrFormat
	}
	return nil
}

func (d *decoder) emit(v byte) error {
	d.window[d.pos] = v
	d.pos++
	if d.pos == frameSize {
		if err := d.flush(frameSize); err != nil {
			return err
		}
		d.pos = 0
	}
	return nil
}

func (d *decoder) readLengths() ([]uint8, []uint8, error) {
	b := d.bits
	read := func(n uint) int {
		v, err := b.read(n)
		if err != nil {
			return -1
		}
		return v
	}
	nlit, ndist, ncode := read(5), read(5), read(4)
	if nlit < 0 || ndist < 0 || ncode < 0 {
		return nil, nil, d.fatal
	}
	nlit, ndist, ncode = nlit+257, ndist+1, ncode+4
	var codeLengths [19]uint8
	for i := 0; i < ncode; i++ {
		v, err := b.read(3)
		if err != nil {
			return nil, nil, err
		}
		codeLengths[codeOrder[i]] = uint8(v)
	}
	tree, err := makeHuffman(codeLengths[:])
	if err != nil {
		return nil, nil, err
	}
	lens := make([]uint8, nlit+ndist)
	last := uint8(0)
	for i := 0; i < len(lens); {
		symbol, err := tree.symbol(&b, 7)
		if err != nil {
			return nil, nil, err
		}
		if symbol < 16 {
			last, lens[i] = uint8(symbol), uint8(symbol)
			i++
			continue
		}
		n, base, value := uint(2), 3, last
		if symbol == 17 {
			n, base, value = 3, 3, 0
		} else if symbol == 18 {
			n, base, value = 7, 11, 0
		} else if symbol != 16 {
			return nil, nil, ErrFormat
		}
		extra, err := b.read(n)
		if err != nil {
			return nil, nil, err
		}
		run := base + extra
		if run > len(lens)-i {
			return nil, nil, ErrFormat
		}
		for j := 0; j < run; j++ {
			lens[i+j] = value
		}
		i += run // reference repeats do not change last, including zero runs
	}
	lit, dist := make([]uint8, 288), make([]uint8, 32)
	copy(lit, lens[:nlit])
	copy(dist, lens[nlit:])
	d.bits = b // failed dynamic headers deliberately do not commit local bits
	return lit, dist, nil
}

func (d *decoder) inflate() error {
	b := d.bits
	for {
		last, err := b.read(1)
		if err != nil {
			return err
		}
		kind, err := b.read(2)
		if err != nil {
			return err
		}
		switch kind {
		case 0:
			b.drop(b.left & 7)
			var header [4]byte
			i := 0
			for b.left >= 8 {
				if i == len(header) {
					return ErrFormat
				}
				header[i] = byte(b.buffer)
				b.drop(8)
				i++
			}
			for ; i < len(header); i++ {
				v, err := b.byte()
				if err != nil {
					return err
				}
				header[i] = v
			}
			n := int(header[0]) | int(header[1])<<8
			complement := int(header[2]) | int(header[3])<<8
			if n != complement^0xffff {
				return ErrFormat
			}
			for i := 0; i < n; i++ {
				v, err := b.byte()
				if err != nil {
					return err
				}
				if err := d.emit(v); err != nil {
					return err
				}
			}
		case 1, 2:
			litLengths, distLengths := make([]uint8, 288), make([]uint8, 32)
			if kind == 1 {
				for i := range litLengths {
					switch {
					case i < 144:
						litLengths[i] = 8
					case i < 256:
						litLengths[i] = 9
					case i < 280:
						litLengths[i] = 7
					default:
						litLengths[i] = 8
					}
				}
				for i := range distLengths {
					distLengths[i] = 5
				}
			} else {
				d.bits = b
				litLengths, distLengths, err = d.readLengths()
				if err != nil {
					return err
				}
				b = d.bits
			}
			lit, err := makeHuffman(litLengths)
			if err != nil {
				return err
			}
			dist, err := makeHuffman(distLengths)
			if err != nil {
				return err
			}
			for {
				if err := d.check(); err != nil {
					return err
				}
				symbol, err := lit.symbol(&b, 16)
				if err != nil {
					return err
				}
				if symbol == 256 {
					break
				}
				if symbol < 256 {
					if err := d.emit(byte(symbol)); err != nil {
						return err
					}
					continue
				}
				slot := symbol - 257
				if slot >= len(lengthBase) {
					return ErrFormat
				}
				extra, err := b.read(lengthExtra[slot])
				if err != nil {
					return err
				}
				length := lengthBase[slot] + extra
				slot, err = dist.symbol(&b, 16)
				if err != nil {
					return err
				}
				if slot >= len(distanceBase) {
					return ErrFormat
				}
				extra, err = b.read(distanceExtra[slot])
				if err != nil {
					return err
				}
				distance := distanceBase[slot] + extra
				from := (d.pos - distance + frameSize) & (frameSize - 1)
				for i := 0; i < length; i++ {
					if err := d.emit(d.window[from]); err != nil {
						return err
					}
					from = (from + 1) & (frameSize - 1)
				}
			}
		default:
			return ErrFormat
		}
		if last != 0 {
			break
		}
	}
	if d.pos != 0 {
		if err := d.flush(d.pos); err != nil {
			return err
		}
	}
	d.bits = b
	return nil
}

// DecodeContext requests exactly outputSize bytes, scanning CK headers across
// input blocks and preserving the 32 KiB dictionary. Repaired counts frames
// whose invalid DEFLATE output was completed with zeros. Input failure is
// returned even if its repaired last frame supplies the requested bytes.
// Callers must bound outputSize and the input source before calling.
func DecodeContext(ctx context.Context, input io.Reader, outputSize int) (out []byte, repaired int, err error) {
	if outputSize < 0 || input == nil {
		return nil, 0, ErrFormat
	}
	d := &decoder{ctx: ctx, input: input}
	d.bits.owner = d
	for len(out) < outputSize {
		if err := ctx.Err(); err != nil {
			return out, d.repaired, err
		}
		b := d.bits
		b.drop(b.left & 7)
		state := 0
		for state != 2 {
			v, err := b.read(8)
			if err != nil {
				return out, d.repaired, err
			}
			if v == 'C' {
				state = 1
			} else if state == 1 && v == 'K' {
				state = 2
			} else {
				state = 0
			}
		}
		d.bits = b
		d.pos, d.flushed = 0, 0
		frameStart := len(out)
		inflateErr := d.inflate()
		if err := ctx.Err(); err != nil {
			return out, d.repaired, err
		}
		if errors.Is(inflateErr, ErrLimit) {
			return out, d.repaired, ErrLimit
		}
		if inflateErr != nil {
			d.repaired++
			if d.flushed == 0 && d.pos > 0 {
				_ = d.flush(d.pos)
			}
			for i := min(d.flushed, frameSize); i < frameSize; i++ {
				d.window[i] = 0
			}
			d.flushed = frameSize
		}
		n := min(outputSize-len(out), d.flushed)
		out = append(out, d.window[:n]...)
		if d.fatal != nil {
			return out, d.repaired, &InputError{FrameStart: frameStart}
		}
	}
	return out, d.repaired, ctx.Err()
}
