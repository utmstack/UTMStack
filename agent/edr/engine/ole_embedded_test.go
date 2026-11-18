package engine

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"math/rand"
	"testing"
)

func biffRecord(typ uint16, body []byte) []byte {
	h := make([]byte, 4, 4+len(body))
	binary.LittleEndian.PutUint16(h, typ)
	binary.LittleEndian.PutUint16(h[2:], uint16(len(body)))
	return append(h, body...)
}

func biffGlobalsBOF() []byte {
	bof := make([]byte, 16)
	binary.LittleEndian.PutUint16(bof[0:], 0x0600) // BIFF8
	binary.LittleEndian.PutUint16(bof[2:], 0x0005) // workbook globals
	binary.LittleEndian.PutUint16(bof[4:], 0x0dbb)
	binary.LittleEndian.PutUint16(bof[6:], 0x07cc)
	return bof
}

// Encrypted legacy Excel (BIFF8 FILEPASS) and PowerPoint (encrypted
// CurrentUserAtom token or EncryptedSummary) binaries are incomplete, never
// clean: NATIVE_ENGINE.md states "Encrypted Office content is incomplete".
func TestOLEEncryptedLegacyExcelAndPowerPointAreIncomplete(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	cipher := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(rng.Intn(255) + 1)
		}
		return b
	}
	// RC4 FILEPASS: wEncryptionType=1, vMajor=1, vMinor=1, salt/verifier/hash.
	rc4 := append([]byte{1, 0, 1, 0, 1, 0}, cipher(48)...)
	// XOR obfuscation FILEPASS: wEncryptionType=0, key, verificationBytes.
	xor := []byte{0, 0, 0x34, 0x12, 0x78, 0x56}
	interfaceHdr := biffRecord(0x00e1, []byte{0xb0, 0x04})
	for _, tc := range []struct {
		name, stream     string
		before, filepass []byte
	}{{"xls-rc4", "Workbook", nil, rc4}, {"xls-xor", "Workbook", nil, xor}, {"xls-later-record", "Workbook", interfaceHdr, rc4}, {"xls-book", "Book", nil, xor}} {
		t.Run(tc.name, func(t *testing.T) {
			stream := append(biffRecord(0x0809, biffGlobalsBOF()), tc.before...)
			stream = append(stream, biffRecord(0x002f, tc.filepass)...)
			stream = append(stream, cipher(5000-len(stream))...)
			b := oleFixture(t, 3, tc.stream, stream, false, false, false)
			sig, err := oleTestState(t).visit(b, 0)
			if sig != "" || !errors.Is(err, ErrUnsupported) {
				t.Fatalf("encrypted BIFF8 workbook: sig=%q err=%v; want incomplete", sig, err)
			}
		})
	}
	t.Run("xls-plain", func(t *testing.T) {
		stream := append(biffRecord(0x0809, biffGlobalsBOF()), biffRecord(0x00e1, []byte{0xb0, 0x04})...)
		stream = append(stream, biffRecord(0x000a, nil)...) // EOF
		stream = append(stream, bytes.Repeat([]byte{'x'}, 5000-len(stream))...)
		b := oleFixture(t, 3, "Workbook", stream, false, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || err != nil {
			t.Fatalf("unencrypted workbook: sig=%q err=%v", sig, err)
		}
	})
	t.Run("xls-filepass-past-the-end", func(t *testing.T) {
		// A FILEPASS record header marks the workbook encrypted even when its
		// data runs past the stream.
		stream := biffRecord(0x0809, biffGlobalsBOF())
		stream = append(stream, 0x2f, 0, 0xff, 0x7f)
		b := oleFixture(t, 3, "Workbook", stream, true, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("sig=%q err=%v", sig, err)
		}
	})
	t.Run("xls-records-break-off", func(t *testing.T) {
		// Records that break off before any FILEPASS show no encryption: the
		// reference reads the stream as it is.
		stream := append(biffRecord(0x0809, biffGlobalsBOF()), biffRecord(0x00e1, []byte{0xb0, 0x04})...)
		stream = append(stream, 0x3c, 0, 0xff, 0x7f)
		b := oleFixture(t, 3, "Workbook", stream, true, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || err != nil {
			t.Fatalf("sig=%q err=%v", sig, err)
		}
	})
	currentUser := func(token uint32) []byte {
		// CurrentUserAtom: RecordHeader(recVer/instance=0, recType=0x0FF6,
		// len), size=0x14, headerToken, offsetToCurrentEdit, lenUserName,
		// docFileVersion, major/minor version, unused, user name.
		hdr := make([]byte, 8)
		binary.LittleEndian.PutUint16(hdr[2:], 0x0ff6)
		body := make([]byte, 24)
		binary.LittleEndian.PutUint32(body[0:], 0x14)
		binary.LittleEndian.PutUint32(body[4:], token)
		binary.LittleEndian.PutUint32(body[8:], 0x1000)
		binary.LittleEndian.PutUint16(body[12:], 4)
		binary.LittleEndian.PutUint16(body[14:], 0x03f4)
		body[16], body[17] = 3, 0
		copy(body[20:], "user")
		binary.LittleEndian.PutUint32(hdr[4:], uint32(len(body)))
		return append(hdr, body...)
	}
	t.Run("ppt-current-user-encrypted-token", func(t *testing.T) {
		b := oleFixture(t, 3, "Current User", currentUser(0xf3d1c4df), true, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("encrypted CurrentUserAtom: sig=%q err=%v; want incomplete", sig, err)
		}
	})
	t.Run("ppt-current-user-plain-token", func(t *testing.T) {
		b := oleFixture(t, 3, "Current User", currentUser(0xe391c05f), true, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || err != nil {
			t.Fatalf("unencrypted CurrentUserAtom: sig=%q err=%v", sig, err)
		}
	})
	t.Run("ppt-encrypted-summary", func(t *testing.T) {
		b := oleFixture(t, 3, "EncryptedSummary", cipher(200), true, false, false)
		if sig, err := oleTestState(t).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("EncryptedSummary: sig=%q err=%v; want incomplete", sig, err)
		}
	})
}

// ole10Native builds the OLE1 packager layout produced by Office "Insert
// Object": total size, flags, label, source path, two reserved words, temp
// path, then the embedded data.
func ole10Native(label string, payload []byte) []byte {
	var body bytes.Buffer
	u16 := func(v uint16) { body.Write(binary.LittleEndian.AppendUint16(nil, v)) }
	u32 := func(v uint32) { body.Write(binary.LittleEndian.AppendUint32(nil, v)) }
	u16(2)
	body.WriteString(label + "\x00")
	body.WriteString("C:\\" + label + "\x00")
	u16(0)
	u16(3)
	tmp := "C:\\Temp\\" + label + "\x00"
	u32(uint32(len(tmp)))
	body.WriteString(tmp)
	u32(uint32(len(payload)))
	body.Write(payload)
	stream := binary.LittleEndian.AppendUint32(nil, uint32(body.Len()))
	return append(stream, body.Bytes()...)
}

// An OLE "\x01Ole10Native" packager stream wraps an embedded file behind a
// header. The payload is unwrapped and analyzed from its own first byte, so
// container dispatch and whole-object signatures apply to it.
func TestOLE10NativePackagedPayloadIsUnwrapped(t *testing.T) {
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	if _, err := w.Write([]byte(oleTestMarker)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	payload := gz.Bytes()
	control := oleFixture(t, 3, "Contents", payload, true, false, false)
	if sig, err := oleTestState(t).visit(control, 0); sig != "OLE.Test" {
		t.Fatalf("fixture control not detected: sig=%q err=%v", sig, err)
	}
	withNames := func(stream []byte) []byte {
		// Current packagers append length-prefixed Unicode copies of the
		// names after the data; the total size covers them.
		stream = append(append([]byte(nil), stream...), 4, 0, 0, 0, 'a', 0, 'b', 0)
		binary.LittleEndian.PutUint32(stream, uint32(len(stream)-4))
		return stream
	}
	for name, stream := range map[string][]byte{
		"packager":                             ole10Native("invoice.gz", payload),
		"unicode names":                        withNames(ole10Native("invoice.gz", payload)),
		"upper-case name":                      ole10Native("invoice.gz", payload),
		"native data without a package header": append(binary.LittleEndian.AppendUint32(nil, uint32(len(payload))), payload...),
	} {
		t.Run(name, func(t *testing.T) {
			streamName := "\x01Ole10Native"
			if name == "upper-case name" {
				streamName = "\x01OLE10NATIVE"
			}
			b := oleFixture(t, 3, streamName, stream, true, false, false)
			if sig, err := oleTestState(t).visit(b, 0); sig != "OLE.Test" || err != nil {
				t.Fatalf("sig=%q err=%v; want the packaged gzip detected", sig, err)
			}
		})
	}
	// Truncated streams still have their embedded bytes analyzed.
	for name, stream := range map[string][]byte{
		"data length past the end": func() []byte {
			b := ole10Native("invoice.gz", append(append([]byte(nil), payload...), bytes.Repeat([]byte{0}, 64)...))
			b = b[:len(b)-64]
			binary.LittleEndian.PutUint32(b, uint32(len(b)-4))
			return b
		}(),
		"total size past the end": func() []byte {
			b := ole10Native("invoice.gz", payload)
			binary.LittleEndian.PutUint32(b, uint32(len(b)))
			return b
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			b := oleFixture(t, 3, "\x01Ole10Native", stream, true, false, false)
			if sig, err := oleTestState(t).visit(b, 0); sig != "OLE.Test" || err != nil {
				t.Fatalf("sig=%q err=%v; want the truncated package's gzip detected", sig, err)
			}
		})
	}
	benign := oleFixture(t, 3, "\x01Ole10Native", ole10Native("note.txt", []byte("harmless text")), true, false, false)
	if sig, err := oleTestState(t).visit(benign, 0); sig != "" || err != nil {
		t.Fatalf("benign package: sig=%q err=%v", sig, err)
	}
}

// A native data stream whose packager header does not validate is never a
// clean verdict: its layout is unknown, although its bytes are still scanned.
func TestOLE10NativeUnrecognizedHeadersAreIncomplete(t *testing.T) {
	valid := ole10Native("a.bin", []byte("harmless payload"))
	for name, tc := range map[string]struct {
		stream []byte
		err    error
	}{
		"truncated":         {valid[:3], ErrMalformed},
		"total too large":   {append(binary.LittleEndian.AppendUint32(nil, 1<<20), valid[4:]...), ErrMalformed},
		"total too small":   {append(binary.LittleEndian.AppendUint32(nil, 2), valid[4:]...), ErrUnsupported},
		"unterminated name": {append(append(binary.LittleEndian.AppendUint32(nil, 40), 2, 0), bytes.Repeat([]byte{'a'}, 38)...), ErrUnsupported},
		"data too long": {func() []byte {
			b := append([]byte(nil), valid...)
			binary.LittleEndian.PutUint32(b[len(b)-len("harmless payload")-4:], 1<<20)
			return b
		}(), ErrMalformed},
		"temp path length mismatch": {func() []byte {
			b := append([]byte(nil), valid...)
			at := bytes.Index(b, []byte("C:\\Temp\\")) - 4
			binary.LittleEndian.PutUint32(b[at:], 3)
			return b
		}(), ErrUnsupported},
	} {
		t.Run(name, func(t *testing.T) {
			b := oleFixture(t, 3, "\x01Ole10Native", tc.stream, true, false, false)
			if sig, err := oleTestState(t).visit(b, 0); sig != "" || !errors.Is(err, tc.err) {
				t.Fatalf("sig=%q err=%v; want %v", sig, err, tc.err)
			}
		})
	}
}

// PowerPoint 97-2003 keeps embedded OLE objects and VBA storages in
// ExOleObjStg records. The reference reads every such record as a size and
// zlib data, whatever its instance, and analyzes whatever decodes; an
// uncompressed record yields nothing (docs/2026-10-14-cfb-oracle/pptlab.py).
func TestPowerPointCompressedExOleObjStgIsUnwrapped(t *testing.T) {
	// The embedded storage's stream spans non-adjacent mini sectors, so only
	// its reconstruction exposes the marker.
	payload := bytes.Repeat([]byte{'x'}, 128)
	copy(payload[56:], oleTestMarker)
	inner := oleFixture(t, 3, "Contents", payload, true, false, false)
	if bytes.Contains(inner, []byte(oleTestMarker)) {
		t.Fatal("fixture must require stream reconstruction")
	}
	record := func(verInstance uint16, typ uint16, body []byte) []byte {
		h := binary.LittleEndian.AppendUint16(nil, verInstance)
		h = binary.LittleEndian.AppendUint16(h, typ)
		h = binary.LittleEndian.AppendUint32(h, uint32(len(body)))
		return append(h, body...)
	}
	compressed := binary.LittleEndian.AppendUint32(nil, uint32(len(inner)))
	compressed = append(compressed, zlibFixture(t, inner)...)
	pad := record(0x000f, 0x03e8, bytes.Repeat([]byte{0}, 4096)) // DocumentContainer
	stream := func(records ...[]byte) []byte {
		return bytes.Join(append([][]byte{pad}, records...), nil)
	}
	for name, tc := range map[string]struct {
		stream []byte
		sig    string
		err    error
	}{
		"compressed":          {stream(record(0x0010, 0x1011, compressed)), "OLE.Test", nil},
		"any instance":        {stream(record(0x0020, 0x1011, compressed)), "OLE.Test", nil},
		"wrong size":          {stream(record(0x0010, 0x1011, append(binary.LittleEndian.AppendUint32(nil, uint32(len(inner)-1)), zlibFixture(t, inner)...))), "OLE.Test", nil},
		"uncompressed":        {stream(record(0x0000, 0x1011, inner)), "", nil},
		"benign":              {stream(record(0x0010, 0x1011, append(binary.LittleEndian.AppendUint32(nil, 99), zlibFixture(t, oleFixture(t, 3, "Contents", []byte("harmless"), true, false, false))...))), "", nil},
		"not zlib":            {stream(record(0x0010, 0x1011, append(binary.LittleEndian.AppendUint32(nil, 99), 1, 2, 3))), "", nil},
		"encrypted session":   {stream(record(0x000f, 0x2f14, make([]byte, 16))), "", ErrUnsupported},
		"truncated record":    {stream(record(0x0010, 0x1011, compressed)[:100]), "", nil},
		"after an empty one":  {stream(record(0x000f, 0x03f8, nil), record(0x0010, 0x1011, compressed)), "", nil},
		"before an empty one": {stream(record(0x0010, 0x1011, compressed), record(0x000f, 0x03f8, nil)), "", nil},
		"zero stream padding": {append(stream(record(0x0010, 0x1011, compressed)), 0, 0, 0), "OLE.Test", nil},
	} {
		t.Run(name, func(t *testing.T) {
			b := oleFixture(t, 3, "PowerPoint Document", tc.stream, false, false, false)
			sig, err := oleTestState(t).visit(b, 0)
			if sig != tc.sig || !errors.Is(err, tc.err) {
				t.Fatalf("sig=%q err=%v; want sig=%q err=%v", sig, err, tc.sig, tc.err)
			}
		})
	}
}
