package engine

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// nrv2bStream is an NRV2B stream of the bytes as literals and the end
// marker (control bits in 32-bit little-endian words, most significant
// first; the marker is a 0 bit, the offset prefix 0x1000002 and 0xff).
func nrv2bStream(data []byte) []byte {
	var out []byte
	at, left, word := 0, 0, uint32(0)
	bit := func(b uint32) {
		if left == 0 {
			if len(out) > 0 {
				binary.LittleEndian.PutUint32(out[at:], word)
			}
			at, left, word = len(out), 32, 0
			out = append(out, 0, 0, 0, 0)
		}
		left--
		word |= b << uint(left)
	}
	for _, c := range data {
		bit(1)
		out = append(out, c)
	}
	bit(0)
	const v = 0xffffff + 3
	for i := 23; i >= 0; i-- {
		bit(v >> uint(i) & 1)
		if i == 0 {
			bit(1)
		} else {
			bit(0)
		}
	}
	out = append(out, 0xff)
	binary.LittleEndian.PutUint32(out[at:], word)
	return out
}

// upxImage is a decoded image the UPX step rebuilds: one 0x1000-byte
// section holding body at the empty section's address (0x1000), then the
// original header UPX keeps (one section, 0x1000 alignment).
func upxImage(body []byte) []byte {
	le := binary.LittleEndian
	img := make([]byte, 0x1000)
	copy(img, body)
	h := make([]byte, 24+0xe0+40)
	copy(h, "PE\x00\x00")
	le.PutUint16(h[4:], 0x14c)
	le.PutUint16(h[6:], 1)
	le.PutUint16(h[20:], 0xe0)
	le.PutUint16(h[22:], 0x102)
	o := h[24:]
	le.PutUint16(o, 0x10b)
	le.PutUint32(o[16:], 0x1000) // entry point
	le.PutUint32(o[28:], 0x400000)
	le.PutUint32(o[32:], 0x1000)
	le.PutUint32(o[36:], 0x200)
	le.PutUint32(o[56:], 0x2000)
	le.PutUint32(o[60:], 0x400)
	le.PutUint32(o[92:], 16)
	s := o[0xe0:]
	copy(s, ".text")
	le.PutUint32(s[8:], 0xf00)
	le.PutUint32(s[12:], 0x1000)
	le.PutUint32(s[16:], 0x1000)
	le.PutUint32(s[20:], 0x400)
	le.PutUint32(s[36:], 0x60000020)
	return append(img, h...)
}

// upxPacked is a 32-bit program with UPX's layout: UPX0 (virtual only, at
// 0x1000) and UPX1 holding the stream at its start and plain code at the
// entry point after it.
func upxPacked(stream []byte) []byte {
	le := binary.LittleEndian
	const raw1 = 0x400
	codeAt := (len(stream) + 0x1ff) &^ 0x1ff
	rsz := codeAt + 0x400
	b := make([]byte, raw1+rsz)
	copy(b, "MZ")
	le.PutUint32(b[0x3c:], 0x40)
	copy(b[0x40:], "PE\x00\x00")
	le.PutUint16(b[0x44:], 0x14c)
	le.PutUint16(b[0x46:], 2)
	le.PutUint16(b[0x54:], 0xe0)
	le.PutUint16(b[0x56:], 0x102)
	o := b[0x58:]
	le.PutUint16(o, 0x10b)
	rva1 := uint32(0x1000 + 0x4000)
	le.PutUint32(o[16:], rva1+uint32(codeAt))
	le.PutUint32(o[28:], 0x400000)
	le.PutUint32(o[32:], 0x1000)
	le.PutUint32(o[36:], 0x200)
	le.PutUint32(o[56:], rva1+uint32((rsz+0xfff)&^0xfff))
	le.PutUint32(o[60:], raw1)
	le.PutUint32(o[92:], 16)
	sec := func(i int, name string, vsz, rva, rawSize, raw uint32) {
		s := b[0x58+0xe0+40*i:]
		copy(s, name)
		le.PutUint32(s[8:], vsz)
		le.PutUint32(s[12:], rva)
		le.PutUint32(s[16:], rawSize)
		le.PutUint32(s[20:], raw)
		le.PutUint32(s[36:], 0xe0000040)
	}
	sec(0, "UPX0", 0x4000, 0x1000, 0, 0)
	sec(1, "UPX1", uint32((rsz+0xfff)&^0xfff), rva1, uint32(rsz), raw1)
	copy(b[raw1:], stream)
	copy(b[raw1+codeAt:], bytes.Repeat([]byte{0x90}, 0x3ff))
	b[raw1+codeAt+0x3ff] = 0xc3
	return b
}

// A program the UPX step rebuilds is scanned as a child: a pattern only the
// decoded image holds (split by control words in the stream) is found.
func TestUPXRebuiltProgramIsScanned(t *testing.T) {
	marker := "this marker is only readable after unpacking the image"
	s := pipelineState(t, map[string]string{"a.ndb": "UPX.Child.Marker:0:*:" + hx(marker) + "\n"})
	packed := upxPacked(nrv2bStream(upxImage([]byte(marker))))
	if bytes.Contains(packed, []byte(marker)) {
		t.Fatal("the marker is readable in the packed program")
	}
	expectScan(t, s, packed, "UPX.Child.Marker", "")
	// Two header copies and no "lea edi" magic in the entry code: the
	// reference takes the last one, and the child is scanned.
	image := upxImage([]byte(marker))
	twice := append(append([]byte(nil), image[:0x800]...), image[0x1000:]...)
	twice = append(twice, image[0x800:]...)
	expectScan(t, s, upxPacked(nrv2bStream(twice)), "UPX.Child.Marker", "")
	// A UPX layout every decoder fails on is plain data to the reference.
	failing := upxPacked(bytes.Repeat([]byte{0}, 64))
	expectScan(t, s, failing, "", "")
}
