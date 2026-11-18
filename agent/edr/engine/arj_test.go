package engine

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
)

// bitWriter packs bits most significant first, as ARJ reads them.
type bitWriter struct {
	out  []byte
	acc  uint64
	bits int
}

func (w *bitWriter) put(v uint64, n int) {
	for i := n - 1; i >= 0; i-- {
		w.acc = w.acc<<1 | v>>uint(i)&1
		if w.bits++; w.bits == 8 {
			w.out = append(w.out, byte(w.acc))
			w.acc, w.bits = 0, 0
		}
	}
}

func (w *bitWriter) bytes() []byte {
	if w.bits > 0 {
		w.put(0, 8-w.bits)
	}
	return w.out
}

// arjItem is a literal (length 0) or a match of length and distance, where
// distance counts back from the byte before the copy position.
type arjItem struct {
	literal          byte
	length, distance int
}

// arjCompress writes one method 1-3 block. Its code-length table codes are
// themselves sent with a two-symbol table (1 bit each): characters 0 and 1
// get 8-bit codes, every other character and match length a 9-bit code, a
// complete canonical code. Positions use one fixed slot: slot 3 (distances
// 4-7, two extra bits) or slot 0 (distance 0).
func arjCompress(items []arjItem, slot int) []byte {
	var w bitWriter
	w.put(uint64(len(items)), 16)
	// Code-length codes: 12 entries, lengths 0,0,0, a run of 3 zeros, 0,0,0,0,1,1.
	w.put(12, arjTBit)
	for i := 0; i < 3; i++ {
		w.put(0, 3)
	}
	w.put(3, 2)
	for i := 6; i < 10; i++ {
		w.put(0, 3)
	}
	w.put(1, 3)
	w.put(1, 3)
	// Character and length code lengths: symbol 10 (0) means 8, 11 (1) means 9.
	w.put(arjNC, arjCBit)
	for i := 0; i < arjNC; i++ {
		if i < 2 {
			w.put(0, 1)
		} else {
			w.put(1, 1)
		}
	}
	// Position table: one symbol, no bits per position.
	w.put(0, arjPBit)
	w.put(uint64(slot), arjPBit)
	code := func(c int) {
		if c < 2 {
			w.put(uint64(c), 8)
		} else {
			w.put(uint64(c+2), 9)
		}
	}
	for _, it := range items {
		if it.length == 0 {
			code(int(it.literal))
			continue
		}
		code(it.length + 256 - arjThreshold)
		if slot == 3 {
			w.put(uint64(it.distance-4), 2)
		}
	}
	return w.bytes()
}

// arjCompressFast writes method 4 codes.
func arjCompressFast(items []arjItem) []byte {
	var w bitWriter
	variable := func(v, first, stop int) {
		plus, pwr, width := 0, 1<<uint(first), first
		for width < stop && v >= plus+pwr {
			w.put(1, 1)
			plus += pwr
			pwr <<= 1
			width++
		}
		if width < stop {
			w.put(0, 1)
		}
		w.put(uint64(v-plus), width)
	}
	for _, it := range items {
		if it.length == 0 {
			variable(0, 0, 7)
			w.put(uint64(it.literal), 8)
			continue
		}
		variable(it.length-arjThreshold+1, 0, 7)
		variable(it.distance, 9, 13)
	}
	return w.bytes()
}

func arjHeaderBytes(basic []byte) []byte {
	h := binary.LittleEndian.AppendUint16([]byte{0x60, 0xea}, uint16(len(basic)))
	h = append(h, basic...)
	h = binary.LittleEndian.AppendUint32(h, crc32.ChecksumIEEE(basic))
	return append(h, 0, 0) // no extended header
}

type arjMemberFixture struct {
	name          string
	method, flags byte
	stored, plain []byte
	crc           uint32
}

func arjArchive(members ...arjMemberFixture) []byte {
	main := make([]byte, 30)
	main[0], main[6] = 30, arjTypeMain
	b := arjHeaderBytes(append(main, "test.arj\x00\x00"...))
	for _, m := range members {
		basic := make([]byte, 30)
		basic[0], basic[4], basic[5] = 30, m.flags, m.method
		binary.LittleEndian.PutUint32(basic[12:], uint32(len(m.stored)))
		binary.LittleEndian.PutUint32(basic[16:], uint32(len(m.plain)))
		binary.LittleEndian.PutUint32(basic[20:], m.crc)
		b = append(b, arjHeaderBytes(append(basic, m.name+"\x00\x00"...))...)
		b = append(b, m.stored...)
	}
	return append(b, 0x60, 0xea, 0, 0)
}

func arjMemberOf(name string, method byte, plain, stored []byte) arjMemberFixture {
	return arjMemberFixture{name: name, method: method, stored: stored, plain: plain, crc: crc32.ChecksumIEEE(plain)}
}

func arjItems(marker string) ([]arjItem, []byte) {
	var items []arjItem
	plain := []byte(marker + "HELLO")
	for _, c := range plain {
		items = append(items, arjItem{literal: c})
	}
	// Distance 4 copies the five-byte period, overlapping itself.
	items = append(items, arjItem{length: 23, distance: 4}, arjItem{literal: 0}, arjItem{literal: 1}, arjItem{literal: 0xff})
	plain = append(plain, bytes.Repeat([]byte("HELLO"), 5)[:23]...)
	return items, append(plain, 0, 1, 0xff)
}

func TestARJDecoding(t *testing.T) {
	items, plain := arjItems(payloadMarker)
	for name, stored := range map[string]struct {
		method byte
		data   []byte
	}{
		"method 1": {1, arjCompress(items, 3)},
		"method 3": {3, arjCompress(items, 3)},
		"method 4": {4, arjCompressFast(items)},
	} {
		t.Run(name, func(t *testing.T) {
			b := arjArchive(arjMemberOf("first.txt", 0, []byte("stored member"), []byte("stored member")), arjMemberOf("payload.bin", stored.method, plain, stored.data))
			requireHidden(t, b)
			expectDetected(t, payloadState(t), b, "Payload.Test")
			expectDetected(t, payloadState(t, plain), b, "Payload.Hash0")
		})
	}
	// Distance 0 repeats the previous byte (a run).
	run := append([]byte("ab"), bytes.Repeat([]byte{'b'}, 200)...)
	stored := arjCompress([]arjItem{{literal: 'a'}, {literal: 'b'}, {length: 200}}, 0)
	expectDetected(t, payloadState(t, run), arjArchive(arjMemberOf("run", 1, run, stored)), "Payload.Hash0")
	// A stored member is analyzed as it is.
	gz := gzipFixture(t, []byte(payloadMarker))
	expectDetected(t, payloadState(t), arjArchive(arjMemberOf("a.gz", 0, gz, gz)), "Payload.Test")
}

func TestARJHostileInput(t *testing.T) {
	items, plain := arjItems("harmless text ")
	good := arjCompress(items, 3)
	cases := map[string]struct {
		archive []byte
		want    error
	}{
		"CRC mismatch": {arjArchive(arjMemberFixture{name: "a", method: 1, stored: good, plain: plain, crc: 1}), ErrMalformed},
		"truncated stream": {arjArchive(func() arjMemberFixture {
			m := arjMemberOf("a", 1, append(plain, bytes.Repeat([]byte{'x'}, 4096)...), good)
			return m
		}()), ErrMalformed},
		"distance before start": {arjArchive(arjMemberOf("a", 1, []byte("xxxxxxxxxx"), arjCompress([]arjItem{{literal: 'x'}, {length: 9, distance: 5}}, 3))), ErrMalformed},
		"incomplete code":       {arjArchive(arjMemberOf("a", 1, plain, append([]byte{0, 5, 0x60}, bytes.Repeat([]byte{0xff}, 64)...))), ErrMalformed},
		"encrypted":             {arjArchive(arjMemberFixture{name: "a", method: 1, flags: arjGarbled, stored: good, plain: plain, crc: crc32.ChecksumIEEE(plain)}), ErrUnsupported},
		"split volume":          {arjArchive(arjMemberFixture{name: "a", method: 1, flags: arjVolume, stored: good, plain: plain, crc: crc32.ChecksumIEEE(plain)}), ErrUnsupported},
		"unknown method":        {arjArchive(arjMemberOf("a", 7, plain, good)), ErrUnsupported},
		"missing end":           {func() []byte { b := arjArchive(arjMemberOf("a", 0, plain, plain)); return b[:len(b)-4] }(), ErrMalformed},
		"member past the end": {func() []byte {
			b := arjArchive(arjMemberOf("a", 0, plain, plain))
			return b[:len(b)-10]
		}(), ErrMalformed},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) { expectIncomplete(t, payloadState(t), c.archive, c.want) })
	}
	// A declared size beyond the budget stops before decoding.
	b := arjArchive(arjMemberOf("a", 1, plain, good))
	// original size lives at basic header offset 16, after the main header
	// (4 + 40 + 4 + 2 bytes) and this header's 4-byte prefix.
	off := len(arjHeaderBytes(append(make([]byte, 30), "test.arj\x00\x00"...))) + 4 + 16
	binary.LittleEndian.PutUint32(b[off:], 0x7fffffff)
	basicStart := off - 16
	size := int(binary.LittleEndian.Uint16(b[basicStart-2:]))
	binary.LittleEndian.PutUint32(b[basicStart+size:], crc32.ChecksumIEEE(b[basicStart:basicStart+size]))
	expectIncomplete(t, payloadState(t), b, ErrLimit)
	// The number of members is bounded by the object limit.
	var many []arjMemberFixture
	for i := 0; i < 40; i++ {
		many = append(many, arjMemberOf("m", 0, []byte("member data"), []byte("member data")))
	}
	s := payloadState(t)
	s.maxFiles = 10
	expectIncomplete(t, s, arjArchive(many...), ErrLimit)
}

// A self-extracting ARJ archive is found in an executable's overlay.
func TestARJSelfExtractor(t *testing.T) {
	items, plain := arjItems(payloadMarker)
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = append([]byte("sfx configuration \x60\xea"), arjArchive(arjMemberOf("p", 1, plain, arjCompress(items, 3)))...)
	requireHidden(t, f.overlay)
	expectDetected(t, payloadState(t), f.build(), "Payload.Test")
}
