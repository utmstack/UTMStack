package unpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"testing"
)

// encoder writes NRV streams: control bits in 32-bit little-endian words,
// most significant first, a word reserved when its first bit is written;
// bytes go inline.
type encoder struct {
	m    Method
	out  []byte
	at   int // position of the current control word
	left int
	word uint32
}

func (e *encoder) bit(b uint32) {
	if e.left == 0 {
		e.flush()
		e.at = len(e.out)
		e.out = append(e.out, 0, 0, 0, 0)
		e.left, e.word = 32, 0
	}
	e.left--
	e.word |= (b & 1) << uint(e.left)
}

func (e *encoder) flush() {
	if len(e.out) >= e.at+4 {
		binary.LittleEndian.PutUint32(e.out[e.at:], e.word)
	}
}

func (e *encoder) literal(c byte) {
	e.bit(1)
	e.out = append(e.out, c)
}

// prefix writes an offset prefix v >= 2 as the method reads it.
func (e *encoder) prefix(v uint64) {
	if e.m == NRV2B {
		n := 63
		for v>>uint(n)&1 == 0 {
			n--
		}
		for i := n - 1; i >= 0; i-- {
			e.bit(uint32(v >> uint(i) & 1))
			if i == 0 {
				e.bit(1)
			} else {
				e.bit(0)
			}
		}
		return
	}
	var bits []uint32
	var put func(v uint64, t uint32)
	put = func(v uint64, t uint32) {
		x, b := v>>1, uint32(v&1)
		if x != 1 {
			put((x>>1)+1, 0)
			bits = append(bits, uint32(x&1))
		}
		bits = append(bits, b, t)
	}
	put(v, 1)
	for _, b := range bits {
		e.bit(b)
	}
}

// match copies length (>= 2) bytes from offset back, with a new offset.
func (e *encoder) match(offset, length uint32) {
	e.bit(0)
	o := offset - 1
	big := map[Method]uint32{NRV2B: 0xd00, NRV2D: 0x500, NRV2E: 0x500}[e.m]
	l := length - 1 // the decoder copies l+1
	if offset > big {
		l--
	}
	switch e.m {
	case NRV2B:
		e.prefix(uint64(o>>8) + 3)
		e.out = append(e.out, byte(o))
		e.lengthB(l)
	case NRV2D, NRV2E:
		// low bit of the stored value: inverted first length bit
		var first uint32
		if e.m == NRV2D {
			if l <= 3 {
				first = l >> 1 // l = first*2 + bit; 4 and up take the long code
			}
		} else {
			first = 0
			if l == 1 || l == 2 {
				first = 1
			}
		}
		stored := o<<1 | (first ^ 1)
		e.prefix(uint64(stored>>8) + 3)
		e.out = append(e.out, byte(stored))
		switch e.m {
		case NRV2D:
			if l >= 4 || l == 0 {
				e.bit(0)
				e.gammaLong(l - 2)
			} else {
				e.bit(l - first*2)
			}
		case NRV2E:
			switch {
			case l == 1 || l == 2:
				e.bit(l - 1)
			case l == 3 || l == 4:
				e.bit(1)
				e.bit(l - 3)
			default:
				e.bit(0)
				e.gammaLong(l - 3)
			}
		}
	}
}

func (e *encoder) lengthB(l uint32) {
	if l >= 1 && l <= 3 {
		e.bit(l >> 1)
		e.bit(l & 1)
		return
	}
	e.bit(0)
	e.bit(0)
	e.gammaLong(l - 2)
}

// gammaLong writes v >= 2 as the length loops read it (after a leading 1).
func (e *encoder) gammaLong(v uint32) {
	n := 31
	for v>>uint(n)&1 == 0 {
		n--
	}
	for i := n - 1; i >= 0; i-- {
		e.bit(v >> uint(i) & 1)
		if i == 0 {
			e.bit(1)
		} else {
			e.bit(0)
		}
	}
}

func (e *encoder) end() []byte {
	e.bit(0)
	e.prefix(0xffffff + 3)
	e.out = append(e.out, 0xff)
	e.flush()
	return e.out
}

// compress is a greedy encoder used to make test streams.
func compress(m Method, data []byte) []byte {
	e := &encoder{m: m}
	for i := 0; i < len(data); {
		best, bestLen := 0, 0
		for j := max(0, i-0x4000); j < i; j++ {
			n := 0
			for i+n < len(data) && data[j+n] == data[i+n] && n < 300 {
				n++
			}
			if n > bestLen {
				best, bestLen = j, n
			}
		}
		if bestLen >= 3 {
			e.match(uint32(i-best), uint32(bestLen))
			i += bestLen
			continue
		}
		e.literal(data[i])
		i++
	}
	return e.end()
}

func TestDecompressRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for _, m := range []Method{NRV2B, NRV2D, NRV2E} {
		for trial := 0; trial < 40; trial++ {
			n := r.Intn(20000)
			data := make([]byte, n)
			for i := range data {
				switch {
				case i > 10 && r.Intn(3) == 0:
					data[i] = data[i-1-r.Intn(min(i, 9000))]
				default:
					data[i] = byte(r.Intn(4))
				}
			}
			src := compress(m, data)
			got, used, err := Decompress(m, src, n)
			if err != nil || !bytes.Equal(got, data) || used != len(src) {
				t.Fatalf("%v trial %d: %d bytes used of %d, err %v, equal %v", m, trial, used, len(src), err, bytes.Equal(got, data))
			}
		}
	}
}

// The failures below were checked against the reference: each crafted
// stream, placed where its UPX step reads, made every decoder fail, and
// the same stream without the fault decompressed.
func TestDecompressFailsLikeTheReference(t *testing.T) {
	for _, m := range []Method{NRV2B, NRV2D, NRV2E} {
		stream := func(fault func(e *encoder)) []byte {
			e := &encoder{m: m}
			for _, c := range []byte("ABCD") {
				e.literal(c)
			}
			fault(e)
			e.literal('Q')
			return e.end()
		}
		valid := stream(func(e *encoder) { e.match(2, 3) })
		if got, _, err := Decompress(m, valid, 64); err != nil || string(got) != "ABCDCDCQ" {
			t.Fatalf("%v valid: %q %v", m, got, err)
		}
		for name, c := range map[string]struct {
			fault func(e *encoder)
			err   error
		}{
			"match before the output start": {func(e *encoder) { e.match(9, 2) }, ErrLookBehind},
			"prefix 0x1000003": {func(e *encoder) {
				e.bit(0)
				e.prefix(0xffffff + 4)
				e.out = append(e.out, 2)
			}, ErrOffset},
			"prefix past 2^32": {func(e *encoder) {
				e.bit(0)
				e.prefix(1<<32 + 3)
				e.out = append(e.out, 2)
			}, ErrOffset},
		} {
			if _, _, err := Decompress(m, stream(c.fault), 64); !errors.Is(err, c.err) {
				t.Errorf("%v %s: %v, want %v", m, name, err, c.err)
			}
		}
		// The destination size is a hard limit; the source ends the stream.
		if _, _, err := Decompress(m, valid, 7); !errors.Is(err, ErrOutputOverrun) {
			t.Errorf("%v one byte short: %v", m, err)
		}
		if _, _, err := Decompress(m, valid, 8); err != nil {
			t.Errorf("%v exact size: %v", m, err)
		}
		if _, _, err := Decompress(m, valid[:len(valid)-1], 64); !errors.Is(err, ErrInputOverrun) {
			t.Errorf("%v cut stream: %v", m, err)
		}
	}
}
