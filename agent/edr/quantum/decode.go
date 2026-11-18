// SPDX-License-Identifier: LGPL-2.1-only

// Package quantum decodes CAB Quantum folders in Go. The arithmetic models
// and framing follow libmspack's Quantum decoder by Stuart Caie, based on
// Matthew Russotto's research. See NOTICE.md and LICENSE.LGPL-2.1.
package quantum

import (
	"context"
	"errors"
)

var ErrFormat = errors.New("invalid Quantum compressed data")

type symbol struct{ value, cumulative uint32 }
type model struct {
	symbols []symbol // descending cumulative frequencies, including a zero sentinel
	shifts  int
}

func newModel(start, count int) model {
	m := model{symbols: make([]symbol, count+1), shifts: 4}
	for i := range m.symbols {
		m.symbols[i] = symbol{uint32(start + i), uint32(count - i)}
	}
	return m
}

func (m *model) update(index int) {
	for i := 0; i < index; i++ {
		m.symbols[i].cumulative += 8
	}
	if m.symbols[0].cumulative <= 3800 {
		return
	}
	m.shifts--
	last := len(m.symbols) - 1
	if m.shifts != 0 {
		for i := last - 1; i >= 0; i-- {
			m.symbols[i].cumulative = max(m.symbols[i].cumulative/2, m.symbols[i+1].cumulative+1)
		}
		return
	}
	m.shifts = 50
	for i := 0; i < last; i++ {
		m.symbols[i].cumulative = (m.symbols[i].cumulative - m.symbols[i+1].cumulative + 1) / 2
	}
	// Equal-frequency ordering matters. Use the reference's pairwise swaps,
	// which are not equivalent to a stable sort by decreasing frequency.
	for i := 0; i < last-1; i++ {
		for j := i + 1; j < last; j++ {
			if m.symbols[j].cumulative > m.symbols[i].cumulative {
				m.symbols[i], m.symbols[j] = m.symbols[j], m.symbols[i]
			}
		}
	}
	for i := last - 1; i >= 0; i-- {
		m.symbols[i].cumulative += m.symbols[i+1].cumulative
	}
}

// bits reads the CAB blocks as a single stream with libmspack's synthetic
// 0xff byte after each block. At physical EOF its bit reader allows two zero
// lookahead bytes, once only. No unbounded padding can manufacture output.
type bits struct {
	blocks         [][]byte
	block, at, pad int
	buffer         uint64
	available      uint
	err            error
}

func (b *bits) byte() byte {
	for b.block < len(b.blocks) {
		data := b.blocks[b.block]
		if b.at < len(data) {
			v := data[b.at]
			b.at++
			return v
		}
		if b.at == len(data) {
			b.at++
			return 0xff
		}
		b.block++
		b.at = 0
	}
	if b.pad >= 2 {
		b.err = ErrFormat
	} else {
		b.pad++
	}
	return 0
}

func (b *bits) read(n uint) uint32 {
	for b.available < n && b.err == nil {
		b.word()
	}
	if b.err != nil {
		return 0
	}
	b.available -= n
	return uint32(b.buffer>>b.available) & ((1 << n) - 1)
}

func (b *bits) word() {
	hi, lo := b.byte(), b.byte()
	b.buffer = b.buffer<<16 | uint64(hi)<<8 | uint64(lo)
	b.available += 16
}

// The reference's READ_MANY_BITS prefetches a 16-bit word whenever at most
// 16 bits remain, even if the requested field would already fit. Matching
// that lookahead is necessary to reject the same truncated streams.
func (b *bits) many(n uint) uint32 {
	var value uint32
	for n > 0 && b.err == nil {
		if b.available <= 16 {
			b.word()
		}
		run := min(b.available, n)
		value = value<<run | b.read(run)
		n -= run
	}
	return value
}

type decoder struct {
	input              bits
	low, high, current uint32
}

func (d *decoder) symbol(m *model) uint32 {
	if d.input.err != nil {
		return 0
	}
	total := m.symbols[0].cumulative
	span := d.high - d.low + 1
	if d.current < d.low || d.current > d.high || total == 0 || span == 0 {
		d.input.err = ErrFormat
		return 0
	}
	scaled := ((d.current-d.low+1)*total - 1) / span
	i := 1
	for i < len(m.symbols)-1 && m.symbols[i].cumulative > scaled {
		i++
	}
	value := m.symbols[i-1].value
	d.high = (d.low + m.symbols[i-1].cumulative*span/total - 1) & 0xffff
	d.low = (d.low + m.symbols[i].cumulative*span/total) & 0xffff
	m.update(i)
	for {
		if d.low&0x8000 != d.high&0x8000 {
			if d.low&0x4000 == 0 || d.high&0x4000 != 0 {
				break
			}
			d.current ^= 0x4000
			d.low &= 0x3fff
			d.high |= 0x4000
		}
		d.low = d.low << 1 & 0xffff
		d.high = (d.high<<1 | 1) & 0xffff
		d.current = (d.current<<1 | d.input.read(1)) & 0xffff
		if d.input.err != nil {
			break
		}
	}
	return value
}

func position(slot uint32) (base uint32, extra uint) {
	for i := uint32(0); i < slot; i++ {
		base += 1 << uint(max(int(i)-2, 0)/2)
	}
	return base, uint(max(int(slot)-2, 0) / 2)
}

func length(slot uint32) (base uint32, extra uint) {
	if slot == 26 {
		return 254, 0
	}
	for i := uint32(0); i < slot; i++ {
		base += 1 << uint(max(int(i)-2, 0)/4)
	}
	return base, uint(max(int(slot)-2, 0) / 4)
}

// DecodeContext decodes at most outputSize bytes, as requested by the CAB
// folder metadata. Callers must enforce their allocation and expansion limits
// before calling. History and models persist across 32 KiB arithmetic frames.
// Initial history is zero, matching ClamAV's mspack_fmap_alloc adapter.
func DecodeContext(ctx context.Context, blocks [][]byte, windowBits, outputSize int) ([]byte, error) {
	return decodeContext(ctx, blocks, windowBits, outputSize, nil)
}

// DecodeValidatedContext also counts all produced bytes as expansion work.
// On failure its returned prefix stops at a successful symbol/frame boundary;
// an earlier complete member may use it, but a partial member still requires
// the caller to apply Quantum's output-window flushing rules.
func DecodeValidatedContext(ctx context.Context, blocks [][]byte, windowBits, outputSize int) (out []byte, produced int, err error) {
	valid := 0
	out, err = decodeContext(ctx, blocks, windowBits, outputSize, &valid)
	produced = len(out)
	if err != nil {
		out = out[:valid]
	}
	return
}

func decodeContext(ctx context.Context, blocks [][]byte, windowBits, outputSize int, valid *int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if windowBits < 10 || windowBits > 21 || outputSize < 0 {
		return nil, ErrFormat
	}
	literals := [4]model{newModel(0, 64), newModel(64, 64), newModel(128, 64), newModel(192, 64)}
	matches := [3]model{newModel(0, min(windowBits*2, 24)), newModel(0, min(windowBits*2, 36)), newModel(0, windowBits*2)}
	lengths, selectors := newModel(0, 27), newModel(0, 7)
	d := decoder{input: bits{blocks: blocks}}
	// A corrupt member can request a huge output but yield only a short
	// reusable prefix. Grow with actual work so retaining that prefix does
	// not also retain an allocation sized by the untrusted request.
	out := make([]byte, 0, min(outputSize, 256))
	frameLeft, nextCheck := 0, 0
	for len(out) < outputSize {
		if len(out) >= nextCheck {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			nextCheck = len(out) + 4096
		}
		if frameLeft == 0 {
			d.low, d.high, d.current = 0, 0xffff, d.input.read(16)
			frameLeft = 32768
		}
		selector := d.symbol(&selectors)
		if selector < 4 {
			value := d.symbol(&literals[selector])
			if d.input.err != nil {
				return out, d.input.err
			}
			out = append(out, byte(value))
			frameLeft--
		} else {
			if selector > 6 {
				return out, ErrFormat
			}
			n := int(selector) - 1 // selectors 4/5 encode length 3/4
			if selector == 6 {
				base, extra := length(d.symbol(&lengths))
				n = int(base+d.input.many(extra)) + 5
			}
			base, extra := position(d.symbol(&matches[selector-4]))
			offset := int(base+d.input.many(extra)) + 1
			if d.input.err != nil {
				return out, d.input.err
			}
			if n > frameLeft || offset > 1<<windowBits {
				return out, ErrFormat
			}
			frameLeft -= n
			count := min(n, outputSize-len(out))
			for j := 0; j < count; j++ {
				var value byte
				if offset <= len(out) {
					value = out[len(out)-offset]
				}
				out = append(out, value)
			}
		}
		if frameLeft == 0 {
			d.input.read(d.input.available % 8)
			for skipped := 0; ; skipped++ {
				if skipped&4095 == 0 {
					if err := ctx.Err(); err != nil {
						return out, err
					}
				}
				v := d.input.read(8)
				if d.input.err != nil {
					return out, d.input.err
				}
				if v == 0xff {
					break
				}
			}
		}
		if valid != nil {
			*valid = len(out)
		}
	}
	return out, nil
}
