package engine

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"testing"
)

func kwajFixture(method uint16, plain []byte, packed []byte) []byte {
	h := append([]byte(nil), kwajMagic...)
	h = binary.LittleEndian.AppendUint16(h, method)
	h = binary.LittleEndian.AppendUint16(h, 0) // filled after extensions
	h = binary.LittleEndian.AppendUint16(h, 0x3f)
	h = binary.LittleEndian.AppendUint32(h, uint32(len(plain)))
	h = binary.LittleEndian.AppendUint16(h, 0x1234)
	h = binary.LittleEndian.AppendUint16(h, 3)
	h = append(h, 1, 2, 3)
	h = append(h, "TEST\x00"...)
	h = append(h, "EX_\x00"...)
	h = binary.LittleEndian.AppendUint16(h, 4)
	h = append(h, "text"...)
	binary.LittleEndian.PutUint16(h[10:12], uint16(len(h)))
	return append(h, packed...)
}

func kwajLiteralLZSS(plain []byte) []byte {
	var out []byte
	for len(plain) > 0 {
		n := min(len(plain), 8)
		out = append(out, byte((1<<n)-1))
		out = append(out, plain[:n]...)
		plain = plain[n:]
	}
	return out
}

type kwajBitWriter struct {
	data []byte
	bits int
}

func (w *kwajBitWriter) put(value, n int) {
	for i := n - 1; i >= 0; i-- {
		if w.bits%8 == 0 {
			w.data = append(w.data, 0)
		}
		if value&(1<<i) != 0 {
			w.data[len(w.data)-1] |= 1 << (7 - w.bits%8)
		}
		w.bits++
	}
}

func kwajLiteralLZH(plain []byte) []byte {
	var w kwajBitWriter
	for i := 0; i < 6; i++ {
		w.put(0, 4) // fixed-length canonical trees
	}
	for len(plain) > 0 {
		n := min(len(plain), 32)
		w.put(0, 4) // MATCHLEN or MATCHLEN2: literal run
		w.put(n-1, 5)
		for _, b := range plain[:n] {
			w.put(int(b), 8)
		}
		plain = plain[n:]
	}
	return w.data
}

func kwajMixedLengthLZH(plain []byte) []byte {
	var w kwajBitWriter
	for _, typ := range []int{1, 2, 3, 0, 0, 0} {
		w.put(typ, 4)
	}
	w.put(4, 4) // MATCHLEN: first length
	for i := 1; i < 16; i++ {
		w.put(0, 1) // repeat preceding length
	}
	w.put(4, 4) // MATCHLEN2: first length
	for i := 1; i < 16; i++ {
		w.put(1, 2) // previous + 0
	}
	for i := 0; i < 32; i++ {
		w.put(5, 4) // LITLEN: uncompressed length list
	}
	for len(plain) > 0 {
		n := min(len(plain), 32)
		w.put(0, 4)
		w.put(n-1, 5)
		for _, b := range plain[:n] {
			w.put(int(b), 8)
		}
		plain = plain[n:]
	}
	return w.data
}

func kwajDeflateBlock(plain, dict []byte) []byte {
	var b bytes.Buffer
	var w *flate.Writer
	var err error
	if dict == nil {
		w, err = flate.NewWriter(&b, flate.BestCompression)
	} else {
		w, err = flate.NewWriterDict(&b, flate.BestCompression, dict)
	}
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(plain); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	block := append([]byte("CK"), b.Bytes()...)
	return append(binary.LittleEndian.AppendUint16(nil, uint16(len(block))), block...)
}

func TestKWAJMethods(t *testing.T) {
	plain := []byte("prefix-" + payloadMarker + "-suffix")
	xor := append([]byte(nil), plain...)
	for i := range xor {
		xor[i] ^= 0xff
	}
	mszip := append(kwajDeflateBlock(plain[:24], nil), kwajDeflateBlock(plain[24:], plain[:24])...)
	mszip = append(mszip, 0, 0)
	methods := map[string]struct {
		method uint16
		packed []byte
	}{
		"stored": {0, plain},
		"xor":    {1, xor},
		"lzss":   {2, kwajLiteralLZSS(plain)},
		"lzh":    {3, kwajLiteralLZH(plain)},
		"mszip":  {4, mszip},
	}
	for name, test := range methods {
		t.Run(name, func(t *testing.T) {
			fixture := kwajFixture(test.method, plain, test.packed)
			if test.method != 0 {
				requireHidden(t, fixture)
			}
			expectDetected(t, payloadState(t), fixture, "Payload.Test")
		})
	}
	for _, method := range []uint16{2, 3, 4} {
		packed := methods[map[uint16]string{2: "lzss", 3: "lzh", 4: "mszip"}[method]].packed
		shortLength := kwajFixture(method, plain, packed)
		binary.LittleEndian.PutUint32(shortLength[14:18], 0)
		requireHidden(t, shortLength)
		expectDetected(t, payloadState(t), shortLength, "Payload.Test")
		benign := bytes.Repeat([]byte{'a'}, len(plain))
		var benignPacked []byte
		switch method {
		case 2:
			benignPacked = kwajLiteralLZSS(benign)
		case 3:
			benignPacked = kwajLiteralLZH(benign)
		case 4:
			benignPacked = append(kwajDeflateBlock(benign, nil), 0, 0)
		}
		shortLength = kwajFixture(method, benign, benignPacked)
		binary.LittleEndian.PutUint32(shortLength[14:18], 0)
		expectIncomplete(t, payloadState(t), shortLength, ErrMalformed)
	}
	for _, method := range []uint16{2, 3} {
		packed := methods[map[uint16]string{2: "lzss", 3: "lzh"}[method]].packed
		h := binary.LittleEndian.AppendUint16(append([]byte(nil), kwajMagic...), method)
		h = binary.LittleEndian.AppendUint16(h, 14)
		h = binary.LittleEndian.AppendUint16(h, 0)
		expectDetected(t, payloadState(t), append(h, packed...), "Payload.Test")
	}
	mixed := kwajFixture(3, plain, kwajMixedLengthLZH(plain))
	requireHidden(t, mixed)
	expectDetected(t, payloadState(t), mixed, "Payload.Test")
	// A stored KWAJ can itself hold another compressed object. The marker
	// is absent in the KWAJ bytes, so this proves recursive extraction.
	nested := gzipFixture(t, []byte(payloadMarker))
	stored := kwajFixture(0, nested, nested)
	requireHidden(t, stored)
	expectDetected(t, payloadState(t), stored, "Payload.Test")
	// A back-reference into the QBasic window must use its 4078 origin.
	lzss := append([]byte{0x01, 'Z', 0xee, 0xf0}, nil...)
	out, err := kwajLZSS(payloadState(t), lzss, 32, -1)
	if err != nil || string(out) != "ZZZZ" {
		t.Fatalf("QBasic LZSS back-reference: %q, %v", out, err)
	}
}

func TestKWAJIncomplete(t *testing.T) {
	plain := []byte(payloadMarker)
	base := kwajFixture(2, plain, kwajLiteralLZSS(plain))
	badOffset := append([]byte(nil), base...)
	binary.LittleEndian.PutUint16(badOffset[10:12], 0xffff)
	expectIncomplete(t, payloadState(t), badOffset, ErrMalformed)
	badHeader := append([]byte(nil), base...)
	badHeader[29] = 'x' // filename terminator
	expectIncomplete(t, payloadState(t), badHeader, ErrMalformed)
	truncated := base[:len(base)-2]
	expectIncomplete(t, payloadState(t), truncated, ErrMalformed)
	oversized := append([]byte(nil), base...)
	binary.LittleEndian.PutUint32(oversized[14:18], 0xffffffff)
	expectIncomplete(t, payloadState(t), oversized, ErrLimit)
	unknown := kwajFixture(5, plain, nil)
	expectIncomplete(t, payloadState(t), unknown, ErrUnsupported)
	brokenMSZIP := kwajFixture(4, plain, kwajDeflateBlock(plain, nil))
	expectDetected(t, payloadState(t), brokenMSZIP, "Payload.Test")
	benign := bytes.Repeat([]byte{'a'}, len(plain))
	expectIncomplete(t, payloadState(t), kwajFixture(4, benign, kwajDeflateBlock(benign, nil)), ErrMalformed)
	limited := payloadState(t)
	limited.maxFile = 10
	expectIncomplete(t, limited, kwajFixture(2, plain, kwajLiteralLZSS(plain)), ErrLimit)
	if _, err := kwajLZH(payloadState(t), []byte{0}, 100, -1); !errors.Is(err, ErrMalformed) {
		t.Fatalf("short LZH: %v", err)
	}
}

func TestKWAJLibmspackHeaderFixtures(t *testing.T) {
	// libmspack test_files/kwajd/f01.kwj is accepted by the reference
	// despite its extension NUL lying at data_offset. f04.kwj is rejected
	// because no NUL appears within the four-byte extension field.
	decode := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	valid := decode("4b57414a88f027d100000f0010003100ff")
	if sig, err := payloadState(t).visit(valid, 0); sig != "" || err != nil {
		t.Fatalf("libmspack f01.kwj: sig=%q err=%v", sig, err)
	}
	invalid := decode("4b57414a88f027d100001200100031323334ff")
	expectIncomplete(t, payloadState(t), invalid, ErrMalformed)
}
