package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"testing"
)

func tnefFixture(attrs ...[]byte) []byte {
	out := []byte{0x78, 0x9f, 0x3e, 0x22, 0, 0}
	out = append(out, tnefAttr(1, 0x00089006, []byte{1, 0, 0, 0})...)
	out = append(out, tnefAttr(1, 0x00069007, []byte{0xe4, 0x04, 0, 0, 0, 0, 0, 0})...)
	for _, attr := range attrs {
		out = append(out, attr...)
	}
	return out
}

func tnefAttr(level byte, tag uint32, value []byte) []byte {
	out := make([]byte, 9+len(value)+2)
	out[0] = level
	binary.LittleEndian.PutUint32(out[1:5], tag)
	binary.LittleEndian.PutUint32(out[5:9], uint32(len(value)))
	copy(out[9:], value)
	var sum uint16
	for _, b := range value {
		sum += uint16(b)
	}
	binary.LittleEndian.PutUint16(out[len(out)-2:], sum)
	return out
}

func tnefProp(typ, id uint16, value []byte) []byte {
	out := make([]byte, 12)
	binary.LittleEndian.PutUint16(out[:2], typ)
	binary.LittleEndian.PutUint16(out[2:4], id)
	binary.LittleEndian.PutUint32(out[4:8], 1)
	binary.LittleEndian.PutUint32(out[8:12], uint32(len(value)))
	out = append(out, value...)
	return append(out, make([]byte, (4-len(value)%4)%4)...)
}

func tnefPropList(props ...[]byte) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, uint32(len(props)))
	for _, prop := range props {
		out = append(out, prop...)
	}
	return out
}

func TestTNEFAttachmentsAndProperties(t *testing.T) {
	e := nativeTestEngine(t)
	attachment := tnefFixture(
		tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}),
		tnefAttr(2, 0x0006800f, gzipFixture(t, []byte(testMarker))),
	)
	property := tnefFixture(
		tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}),
		tnefAttr(2, 0x00069005, tnefPropList(tnefProp(0x0102, 0x3701, gzipFixture(t, []byte(testMarker))))),
	)
	unicodeBody := make([]byte, 0, len(testMarker)*2+2)
	for _, b := range []byte(testMarker) {
		unicodeBody = append(unicodeBody, b, 0)
	}
	unicodeBody = append(unicodeBody, 0, 0)
	message := tnefFixture(tnefAttr(1, 0x00069003, tnefPropList(tnefProp(0x001f, 0x1000, unicodeBody))))
	mail := []byte("MIME-Version: 1.0\r\nContent-Type: application/ms-tnef\r\nContent-Transfer-Encoding: base64\r\n\r\n" + base64.StdEncoding.EncodeToString(attachment))
	for name, data := range map[string][]byte{"attachment": attachment, "property": property, "unicode message": message, "mime": mail} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "Native.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

func TestTNEFMalformedAndUnsupported(t *testing.T) {
	e := nativeTestEngine(t)
	base := tnefFixture(tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}), tnefAttr(2, 0x0006800f, []byte("benign")))
	badChecksum := append([]byte(nil), base...)
	badChecksum[len(badChecksum)-1] ^= 1
	badLength := append([]byte(nil), base...)
	attachmentOffset := 6 + len(tnefAttr(1, 0x00089006, []byte{1, 0, 0, 0})) +
		len(tnefAttr(1, 0x00069007, []byte{0xe4, 0x04, 0, 0, 0, 0, 0, 0})) +
		len(tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}))
	binary.LittleEndian.PutUint32(badLength[attachmentOffset+5:attachmentOffset+9], 0xffffffff)
	unknownProp := tnefFixture(tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}), tnefAttr(2, 0x00069005, tnefPropList([]byte{0xff, 0x7f, 0x01, 0x37})))
	rtf := tnefFixture(tnefAttr(1, 0x00069003, tnefPropList(tnefProp(0x0102, 0x1009, []byte("unexpanded RTF")))))
	for name, data := range map[string][]byte{"bad checksum": badChecksum, "bad length": badLength, "unknown property": unknownProp, "compressed RTF": rtf, "truncated": base[:len(base)-1]} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "" || err == nil {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
	if clean, sig, err := e.Scan(context.Background(), tnefFixture()); !clean || sig != "" || err != nil {
		t.Fatal(clean, sig, err)
	}
	// Even a bad checksum cannot hide a recognized decoded attachment.
	infected := tnefFixture(tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}), tnefAttr(2, 0x0006800f, gzipFixture(t, []byte(testMarker))))
	infected[len(infected)-1] ^= 1
	if clean, sig, err := e.Scan(context.Background(), infected); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
	state := scanState{ctx: context.Background(), db: e.current.Load().db, maxFile: int64(len(base)), remaining: int64(len(base) + 2), maxFiles: 10, maxDepth: 8}
	if _, err := state.visit(base, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestTNEFNamedAndMultivalueProperties(t *testing.T) {
	e := nativeTestEngine(t)
	prop := make([]byte, 4)
	binary.LittleEndian.PutUint16(prop[:2], 0x1102)  // MVBinary
	binary.LittleEndian.PutUint16(prop[2:4], 0x8001) // named property
	prop = append(prop, make([]byte, 16)...)
	prop = binary.LittleEndian.AppendUint32(prop, 1) // string property name
	prop = binary.LittleEndian.AppendUint32(prop, 4)
	prop = append(prop, 'X', 0, 0, 0)
	prop = binary.LittleEndian.AppendUint32(prop, 2)
	prop = binary.LittleEndian.AppendUint32(prop, 4)
	prop = append(prop, []byte("safe")...)
	inner := gzipFixture(t, []byte(testMarker))
	prop = binary.LittleEndian.AppendUint32(prop, uint32(len(inner)))
	prop = append(prop, inner...)
	prop = append(prop, bytes.Repeat([]byte{0}, (4-len(inner)%4)%4)...)
	data := tnefFixture(tnefAttr(2, 0x00069002, []byte{1, 0, 0, 0}), tnefAttr(2, 0x00069005, tnefPropList(prop)))
	if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

func rtfLiteralFixture(raw []byte) []byte {
	contents := make([]byte, 0, len(raw)+len(raw)/8+4)
	writePos := 207
	for len(raw) >= 8 {
		contents = append(contents, 0)
		contents = append(contents, raw[:8]...)
		raw = raw[8:]
		writePos += 8
	}
	contents = append(contents, 1<<len(raw))
	contents = append(contents, raw...)
	writePos += len(raw)
	contents = binary.BigEndian.AppendUint16(contents, uint16((writePos&4095)<<4))
	out := make([]byte, 16)
	binary.LittleEndian.PutUint32(out[:4], uint32(len(contents)+12))
	binary.LittleEndian.PutUint32(out[4:8], uint32(writePos-207))
	binary.LittleEndian.PutUint32(out[8:12], 0x75465a4c)
	binary.LittleEndian.PutUint32(out[12:16], rtfCRC(contents))
	return append(out, contents...)
}

func TestRTFCompressionReferenceAndTNEF(t *testing.T) {
	// Example 3.1.1 in MS-OXRTFCP has an independently published CRC and
	// output; it covers the preloaded dictionary and overlapping references.
	contents, err := hex.DecodeString("03000a007263706731323542320af32068656c090020627705b06c647d0a800fa0")
	if err != nil {
		t.Fatal(err)
	}
	if crc := rtfCRC(contents); crc != 0xa7c7c5f1 {
		t.Fatalf("reference CRC = %08x", crc)
	}
	reference := make([]byte, 16)
	binary.LittleEndian.PutUint32(reference[:4], uint32(len(contents)+12))
	binary.LittleEndian.PutUint32(reference[4:8], 43)
	binary.LittleEndian.PutUint32(reference[8:12], 0x75465a4c)
	binary.LittleEndian.PutUint32(reference[12:16], 0xa7c7c5f1)
	reference = append(reference, contents...)
	s := scanState{ctx: context.Background(), maxFile: 1024, remaining: 1024}
	decoded, err := s.decompressRTF(reference)
	if err != nil || string(decoded) != "{\\rtf1\\ansi\\ansicpg1252\\pard hello world}\r\n" {
		t.Fatalf("decoded %q: %v", decoded, err)
	}
	e := nativeTestEngine(t)
	raw := []byte("{\\rtf1 " + testMarker + "}")
	compressed := rtfLiteralFixture(raw)
	if bytes.Contains(compressed, []byte(testMarker)) {
		t.Fatal("fixture exposes marker without decompression")
	}
	data := tnefFixture(tnefAttr(1, 0x00069003, tnefPropList(tnefProp(0x0102, 0x1009, compressed))))
	if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
	broken := append([]byte(nil), compressed...)
	broken[12] ^= 1
	if decoded, err := s.decompressRTF(broken); !bytes.Equal(decoded, raw) || !errors.Is(err, ErrMalformed) {
		t.Fatalf("crc mismatch: %q %v", decoded, err)
	}
	uncompressed := make([]byte, 16)
	binary.LittleEndian.PutUint32(uncompressed[:4], uint32(len(raw)+12))
	binary.LittleEndian.PutUint32(uncompressed[4:8], uint32(len(raw)))
	binary.LittleEndian.PutUint32(uncompressed[8:12], 0x414c454d)
	uncompressed = append(uncompressed, raw...)
	if decoded, err := s.decompressRTF(uncompressed); !bytes.Equal(decoded, raw) || err != nil {
		t.Fatalf("MELA: %q %v", decoded, err)
	}
}
