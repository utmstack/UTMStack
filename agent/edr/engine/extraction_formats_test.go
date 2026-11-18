package engine

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func gzipWriter(w io.Writer) *gzip.Writer { return gzip.NewWriter(w) }

func payloadStateDB(tb testing.TB) *signatures.Database {
	tb.Helper()
	db, err := signatures.Compile(map[string][]byte{"test.ndb": []byte("Payload.Test:0:*:" + hex.EncodeToString([]byte(payloadMarker)) + "\n")})
	if err != nil {
		tb.Fatal(err)
	}
	return db
}

// Formats below hide the payload from raw matching (compression, text
// encodings or wrappers), so a detection proves the decoder ran. Each test
// also checks that the fixture itself does not contain the marker.
func requireHidden(t *testing.T, b []byte) {
	t.Helper()
	if bytes.Contains(b, []byte(payloadMarker)) {
		t.Fatal("fixture leaks the marker")
	}
}

func expectDetected(t *testing.T, s *scanState, b []byte, want string) {
	t.Helper()
	if sig, err := s.visit(b, 0); sig != want || err != nil {
		t.Fatalf("sig=%q err=%v, want %q", sig, err, want)
	}
}

func expectIncomplete(t *testing.T, s *scanState, b []byte, want error) {
	t.Helper()
	if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, want) {
		t.Fatalf("sig=%q err=%v, want %v", sig, err, want)
	}
}

// cpioArchive writes members in one of the four header formats.
func cpioArchive(kind cpioKind, names []string, contents [][]byte, trailer bool) []byte {
	var b []byte
	pad := func(align int) {
		for len(b)%align != 0 {
			b = append(b, 0)
		}
	}
	member := func(name string, data []byte) {
		name += "\x00"
		switch kind {
		case cpioODC:
			b = append(b, fmt.Sprintf("070707%06o%06o%06o%06o%06o%06o%06o%011o%06o%011o", 1, 2, 0o100644, 0, 0, 1, 0, 0, len(name), len(data))...)
			b = append(append(b, name...), data...)
		case cpioNewc:
			b = append(b, fmt.Sprintf("070701%08X%08X%08X%08X%08X%08X%08X%08X%08X%08X%08X%08X%08X", 2, 0o100644, 0, 0, 1, 0, len(data), 0, 0, 0, 0, len(name), 0)...)
			b = append(b, name...)
			pad(4)
			b = append(b, data...)
			pad(4)
		default:
			var order binary.ByteOrder = binary.LittleEndian
			if kind == cpioBinaryBE {
				order = binary.BigEndian
			}
			h := make([]byte, 26)
			order.PutUint16(h, 0o070707)
			order.PutUint16(h[6:], 0o100644)
			order.PutUint16(h[20:], uint16(len(name)))
			order.PutUint16(h[22:], uint16(len(data)>>16))
			order.PutUint16(h[24:], uint16(len(data)))
			b = append(append(b, h...), name...)
			pad(2)
			b = append(b, data...)
			pad(2)
		}
	}
	for i, name := range names {
		member(name, contents[i])
	}
	if trailer {
		member("TRAILER!!!", nil)
	}
	return append(b, make([]byte, 64)...)
}

func TestCPIOFormats(t *testing.T) {
	payload := gzipFixture(t, []byte(payloadMarker))
	exact := []byte("an exactly bounded cpio member")
	for name, kind := range map[string]cpioKind{"odc": cpioODC, "newc": cpioNewc, "binary LE": cpioBinaryLE, "binary BE": cpioBinaryBE} {
		t.Run(name, func(t *testing.T) {
			// Odd name and data lengths exercise each format's padding.
			b := cpioArchive(kind, []string{"dir/a.txt", "b.gz"}, [][]byte{[]byte("odd"), payload}, true)
			requireHidden(t, b)
			expectDetected(t, payloadState(t), b, "Payload.Test")
			b = cpioArchive(kind, []string{"x", "exact.bin"}, [][]byte{[]byte("first member"), exact}, true)
			expectDetected(t, payloadState(t, exact), b, "Payload.Hash0")
			// The reference accepts EOF after complete members without a
			// trailer, but a full broken later header remains incomplete.
			b = cpioArchive(kind, []string{"a", "b"}, [][]byte{[]byte("first"), []byte("second")}, false)
			if sig, err := payloadState(t).visit(b[:len(b)-64], 0); sig != "" || err != nil {
				t.Fatalf("complete CPIO without trailer: sig=%q err=%v", sig, err)
			}
			// Incomplete trailing bytes shorter than the next header do
			// not change the result or hide a scanned child.
			if sig, err := payloadState(t).visit(b[:len(b)-64+1], 0); sig != "" || err != nil {
				t.Fatalf("short next CPIO header: sig=%q err=%v", sig, err)
			}
			noTrailer := cpioArchive(kind, []string{"payload.gz"}, [][]byte{payload}, false)
			requireHidden(t, noTrailer)
			expectDetected(t, payloadState(t), noTrailer[:len(noTrailer)-64], "Payload.Test")
			b = cpioArchive(kind, []string{"a", "b"}, [][]byte{[]byte("first"), []byte("second")}, true)
			second := bytes.Index(b[4:], b[:4]) + 4
			b[second] ^= 0xff
			expectIncomplete(t, payloadState(t), b, ErrMalformed)
		})
	}
	// Declared sizes beyond the data and names without a terminator are not
	// accepted as a first header; later ones make the archive malformed.
	odc := cpioArchive(cpioODC, []string{"a"}, [][]byte{[]byte("data")}, true)
	huge := append([]byte(nil), odc...)
	copy(huge[65:76], "77777777777")
	if _, recognized, _ := payloadState(t).cpio(huge, cpioODC, 0); recognized {
		t.Fatal("oversized first member recognized")
	}
	noNUL := append([]byte(nil), odc...)
	noNUL[76+1] = 'x'
	if _, recognized, _ := payloadState(t).cpio(noNUL, cpioODC, 0); recognized {
		t.Fatal("unterminated name recognized")
	}
	// Two bytes of the binary magic alone are ordinary data.
	expectDetected(t, payloadState(t), append([]byte{0xc7, 0x71, 1, 2, 3, 4}, payloadMarker...), "Payload.Test")
	// Member count is bounded by the object limit.
	names := make([]string, 50)
	contents := make([][]byte, 50)
	for i := range names {
		names[i], contents[i] = fmt.Sprint(i), []byte("member data")
	}
	s := payloadState(t)
	s.maxFiles = 10
	expectIncomplete(t, s, cpioArchive(cpioNewc, names, contents, true), ErrLimit)
}

// szddCompress emits literals, with a back-reference for each run the caller
// marks: a real SZDD stream's items, not a stored copy.
func szddCompress(items []any) []byte {
	var out, group []byte
	var flags byte
	n, size := 0, 0
	flush := func() {
		out = append(append(out, flags), group...)
		flags, group, n = 0, nil, 0
	}
	for _, it := range items {
		switch v := it.(type) {
		case byte:
			flags |= 1 << n
			group = append(group, v)
			size++
		case [2]int: // window position, length
			group = append(group, byte(v[0]), byte(v[0]>>8&0x0f<<4)|byte(v[1]-3))
			size += v[1]
		}
		if n++; n == 8 {
			flush()
		}
	}
	if n > 0 {
		flush()
	}
	h := append([]byte(nil), szddMagic...)
	h = append(h, 'A', 'e')
	return append(binary.LittleEndian.AppendUint32(h, uint32(size)), out...)
}

func TestSZDDExpansion(t *testing.T) {
	var items []any
	for _, c := range []byte(payloadMarker) {
		items = append(items, c)
	}
	// Copy "INSTALLER" again from where it was written (window 4080), then
	// three of the initial spaces, then an overlapping copy of the last byte.
	// The window wraps: the marker overwrites its first positions.
	items = append(items, [2]int{4080, 9}, [2]int{2000, 3}, byte('Z'), [2]int{(4080 + len(payloadMarker) + 9 + 3) & 4095, 5})
	want := payloadMarker + payloadMarker[:9] + "   " + "ZZZZZZ"
	b := szddCompress(items)
	out, err := szddExpand(b[14:], len(want))
	if err != nil || string(out) != want {
		t.Fatalf("expanded %q err=%v", out, err)
	}
	expectDetected(t, payloadState(t, []byte(want)), b, "Payload.Hash0")
	gz := gzipFixture(t, []byte(payloadMarker))
	items = nil
	for _, c := range gz {
		items = append(items, c)
	}
	b = szddCompress(items)
	requireHidden(t, b)
	expectDetected(t, payloadState(t), b, "Payload.Test")
	// Hostile inputs never come back clean.
	huge := append([]byte(nil), b...)
	binary.LittleEndian.PutUint32(huge[10:], 0xffffffff)
	expectIncomplete(t, payloadState(t), huge, ErrLimit)
	expectIncomplete(t, payloadState(t), b[:len(b)-4], ErrMalformed)
	expectIncomplete(t, payloadState(t), append(append([]byte(nil), b...), bytes.Repeat([]byte{0xff}, 64)...), ErrMalformed)
	other := append([]byte(nil), b...)
	other[8] = 'B'
	expectIncomplete(t, payloadState(t), other, ErrUnsupported)
}

// binhexEncode writes a BinHex 4.0 file with run-length encoding.
func binhexEncode(name string, dataFork, resourceFork []byte, corruptCRC bool) []byte {
	var plain []byte
	header := append([]byte{byte(len(name))}, name...)
	header = append(header, 0)
	header = append(header, "TEXTttxt\x00\x00"...)
	header = binary.BigEndian.AppendUint32(header, uint32(len(dataFork)))
	header = binary.BigEndian.AppendUint32(header, uint32(len(resourceFork)))
	plain = binary.BigEndian.AppendUint16(append(plain, header...), crcXMODEM(header))
	crc := crcXMODEM(dataFork)
	if corruptCRC {
		crc++
	}
	plain = binary.BigEndian.AppendUint16(append(plain, dataFork...), crc)
	plain = binary.BigEndian.AppendUint16(append(plain, resourceFork...), crcXMODEM(resourceFork))
	var rle []byte
	for i := 0; i < len(plain); {
		c := plain[i]
		run := 1
		for i+run < len(plain) && plain[i+run] == c && run < 255 {
			run++
		}
		if c == 0x90 {
			rle = append(rle, 0x90, 0)
			run = 1
		} else if run > 2 {
			rle = append(rle, c, 0x90, byte(run))
		} else {
			rle = append(rle, c)
			run = 1
		}
		i += run
	}
	out := []byte("(This file must be converted with BinHex 4.0)\r\n\r\n:")
	var acc uint32
	bits, column := 0, 1
	emit := func(v byte) {
		out = append(out, binhexAlphabet[v])
		if column++; column == 64 {
			out = append(out, '\r', '\n')
			column = 0
		}
	}
	for _, c := range rle {
		acc = acc<<8 | uint32(c)
		bits += 8
		for bits >= 6 {
			bits -= 6
			emit(byte(acc>>bits) & 63)
		}
	}
	if bits > 0 {
		emit(byte(acc<<(6-bits)) & 63)
	}
	return append(out, ':', '\r', '\n')
}

func TestBinHexDecoding(t *testing.T) {
	// Runs and 0x90 bytes exercise the run-length encoding.
	fork := append(append(bytes.Repeat([]byte{'A'}, 300), 0x90, 0x90, 'x'), payloadMarker...)
	b := binhexEncode("clean.txt", fork, []byte("resource fork"), false)
	requireHidden(t, b)
	expectDetected(t, payloadState(t), b, "Payload.Test")
	expectDetected(t, payloadState(t, fork), b, "Payload.Hash0")
	// The resource fork is analyzed as well.
	b = binhexEncode("r", []byte("data"), gzipFixture(t, []byte(payloadMarker)), false)
	expectDetected(t, payloadState(t), append([]byte("Mail text before the attachment\n"), b...), "Payload.Test")
	// A checksum error is incomplete but content found is still reported.
	b = binhexEncode("x", fork, nil, true)
	expectDetected(t, payloadState(t), b, "Payload.Test")
	b = binhexEncode("x", []byte("harmless"), nil, true)
	expectIncomplete(t, payloadState(t), b, ErrMalformed)
	// Truncation, invalid characters and run-length bombs.
	b = binhexEncode("x", []byte("harmless data fork"), nil, false)
	expectIncomplete(t, payloadState(t), b[:len(b)-10], ErrMalformed)
	open := binhexStart(b)
	bad := append(append(append([]byte(nil), b[:open+5]...), '~'), b[open+5:]...)
	expectIncomplete(t, payloadState(t), bad, ErrMalformed)
	bomb := []byte("A")
	for i := 0; i < 2000; i++ {
		bomb = append(bomb, 0x90, 0xff)
	}
	s := payloadState(t)
	s.maxFile, s.remaining = 64<<10, 64<<10
	if _, err := binhexDecode(binhexText(bomb), 64<<10); !errors.Is(err, ErrLimit) {
		t.Fatalf("run-length bomb err=%v", err)
	}
	// Prose that mentions the format is not BinHex data.
	prose := "The attachment said\n(This file must be converted with BinHex 4.0)\nbut it was lost.\n" + payloadMarker
	expectDetected(t, payloadState(t), []byte(prose), "Payload.Test")
	if binhexStart([]byte(prose)) >= 0 {
		t.Fatal("prose recognized as BinHex")
	}
}

// binhexText six-bit encodes already run-length encoded bytes.
func binhexText(rle []byte) []byte {
	var out []byte
	var acc uint32
	bits := 0
	for _, c := range rle {
		acc = acc<<8 | uint32(c)
		bits += 8
		for bits >= 6 {
			bits -= 6
			out = append(out, binhexAlphabet[byte(acc>>bits)&63])
		}
	}
	return append(out, ':')
}

func ole1Header(class string) []byte {
	b := binary.LittleEndian.AppendUint32(nil, 0x0501)
	b = binary.LittleEndian.AppendUint32(b, 2)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(class)+1))
	b = append(append(b, class...), 0)
	return binary.LittleEndian.AppendUint32(binary.LittleEndian.AppendUint32(b, 0), 0)
}

// rtfDocument wraps object data in an embedded-object group, with the hex
// text split into lines and interrupted the way editors and evasions do.
func rtfDocument(objdata string) []byte {
	return []byte(`{\rtf1\ansi{\fonttbl{\f0 Times;}}{\pard Text\par}{\object\objemb\objw855\objh765{\*\objclass Package}{\*\objdata ` + objdata + `}{\result{\pict\wmetafile8 0102}}}}`)
}

func splitHex(b []byte) string {
	h := hex.EncodeToString(b)
	var out strings.Builder
	for i := 0; i < len(h); i += 64 {
		out.WriteString(h[i:min(len(h), i+64)])
		out.WriteString("\r\n")
	}
	return out.String()
}

func TestRTFEmbeddedObjects(t *testing.T) {
	gz := gzipFixture(t, []byte(payloadMarker))
	packaged := append(ole1Header("Package"), ole10Native("invoice.gz", gz)...)
	b := rtfDocument(splitHex(packaged))
	requireHidden(t, b)
	expectDetected(t, payloadState(t), b, "Payload.Test")
	// The packaged file is unwrapped with its exact bounds.
	exact := []byte("the packaged file itself")
	b = rtfDocument(splitHex(append(ole1Header("Package"), ole10Native("a.txt", exact)...)))
	expectDetected(t, payloadState(t, exact), b, "Payload.Hash0")
	// Other classes: the native data is analyzed as an object.
	native := binary.LittleEndian.AppendUint32(ole1Header("Word.Document.8"), uint32(len(gz)))
	b = rtfDocument(splitHex(append(native, gz...)))
	expectDetected(t, payloadState(t), b, "Payload.Test")
	// Non-hex characters, control words and \bin raw bytes inside the data.
	h := splitHex(packaged)
	evasive := h[:10] + "zz\\par " + h[10:20] + "{\\*\\junk}" + h[20:40] + "\\bin3 " + string(packaged[20:23]) + h[46:]
	expectDetected(t, payloadState(t), rtfDocument(evasive), "Payload.Test")
	// Data without an OLE1 header is analyzed as it is.
	expectDetected(t, payloadState(t), rtfDocument(splitHex(gz)), "Payload.Test")
	// An escaped backslash is text, not a control word.
	expectDetected(t, payloadState(t), []byte(`{\rtf1 \\objdata `+hex.EncodeToString([]byte("x"))+` `+payloadMarker+`}`), "Payload.Test")
	// Incomplete: an object group that never closes, and a truncated header.
	expectIncomplete(t, payloadState(t), []byte(`{\rtf1{\object{\*\objdata `+hex.EncodeToString(packaged[:40])), ErrMalformed)
	expectIncomplete(t, payloadState(t), rtfDocument(splitHex(ole1Header("Package")[:14])), ErrMalformed)
	s := payloadState(t)
	s.maxFile, s.remaining = 32<<10, 32<<10
	expectIncomplete(t, s, rtfDocument(strings.Repeat("00", 20<<10)), ErrLimit)
}

func TestHTMLDataURIs(t *testing.T) {
	gz := gzipFixture(t, []byte(payloadMarker))
	enc := base64.StdEncoding.EncodeToString(gz)
	for name, doc := range map[string]string{
		"link":           `<a href="data:application/octet-stream;base64,` + enc + `">t</a>`,
		"single quotes":  `<img src='data:image/png;BASE64,` + enc[:20] + "\n  " + enc[20:] + `'>`,
		"unquoted":       `<iframe src=data:;base64,` + enc + `></iframe>`,
		"css url":        `<div style="background:url(data:application/x-gzip;base64,` + enc + `)"></div>`,
		"url-safe":       `<a href="data:application/octet-stream;base64,` + strings.NewReplacer("+", "-", "/", "_").Replace(enc) + `">`,
		"percent":        `<a href="data:text/plain,` + urlEscape(gz) + `">`,
		"many resources": strings.Repeat(`<img src="data:image/gif;base64,R0lGODlhAQABAAAAACw=">`, 20) + `<a href="data:,` + urlEscape(gz) + `">`,
	} {
		t.Run(name, func(t *testing.T) {
			requireHidden(t, []byte(doc))
			expectDetected(t, payloadState(t), []byte(doc), "Payload.Test")
		})
	}
	// "data:" in text or in other attributes' prose is not a data URI.
	for _, doc := range []string{`<p>data: ` + enc + `</p>`, `<p title="metadata: ` + enc + `">`} {
		if sig, err := payloadState(t).visit([]byte(doc), 0); sig != "" || err != nil {
			t.Fatalf("%q: sig=%q err=%v", doc[:20], sig, err)
		}
	}
	s := payloadState(t)
	s.maxFile, s.remaining = 16<<10, 16<<10
	expectIncomplete(t, s, []byte(`<a href="data:;base64,`+strings.Repeat("QUFB", 8<<10)+`">`), ErrLimit)
}

func urlEscape(b []byte) string {
	var out strings.Builder
	for _, c := range b {
		fmt.Fprintf(&out, "%%%02X", c)
	}
	return out.String()
}

func TestMboxMessages(t *testing.T) {
	attachment := base64.StdEncoding.EncodeToString(gzipFixture(t, []byte(payloadMarker)))
	msg := "From: a@example.com\r\nMIME-Version: 1.0\r\nContent-Type: application/octet-stream; name=\"a.gz\"\r\nContent-Transfer-Encoding: base64\r\n\r\n" + attachment + "\r\n"
	single := "From a@example.com  Thu Jul 31 13:49:50 2008\n" + msg
	requireHidden(t, []byte(single))
	expectDetected(t, payloadState(t), []byte(single), "Payload.Test")
	// The second message of a mailbox, after a plain first message.
	multi := "From a@example.com  Thu Jul 31 13:49:50 2008\nFrom: a@example.com\nSubject: hi\n\nFrom here on, a quoted line.\n\n" +
		"From b@example.com  Fri Aug  1 10:00:00 2008\n" + msg
	expectDetected(t, payloadState(t), []byte(multi), "Payload.Test")
	// A plain message's text is still analyzed for encodings.
	uu := "From a@example.com  Thu Jul 31 13:49:50 2008\nSubject: x\n\n" + uuencodeFixture(gzipFixture(t, []byte(payloadMarker)))
	expectDetected(t, payloadState(t), []byte(uu), "Payload.Test")
	// A MIME message that cannot be parsed is incomplete.
	broken := "From a@example.com  Thu Jul 31 13:49:50 2008\nMIME-Version: 1.0\nContent-Type: multipart/mixed\n\nno boundary\n"
	expectIncomplete(t, payloadState(t), []byte(broken), ErrMalformed)
}

func uuencodeFixture(b []byte) string {
	var out strings.Builder
	out.WriteString("begin 644 a.gz\n")
	for i := 0; i < len(b); i += 45 {
		line := b[i:min(len(b), i+45)]
		out.WriteByte(byte(' ' + len(line)))
		for j := 0; j < len(line); j += 3 {
			var g [3]byte
			copy(g[:], line[j:])
			for _, v := range []byte{g[0] >> 2, g[0]<<4&0x30 | g[1]>>4, g[1]<<2&0x3c | g[2]>>6, g[2] & 0x3f} {
				if v == 0 {
					out.WriteByte('`')
				} else {
					out.WriteByte(' ' + v)
				}
			}
		}
		out.WriteByte('\n')
	}
	out.WriteString("`\nend\n")
	return out.String()
}

// bzip2 of "BZIP2-ZIP-HARMLESS-MEMBER-MARKER", made with the bzip2 tool: the
// standard library has no bzip2 compressor.
const bzip2Member = "BZIP2-ZIP-HARMLESS-MEMBER-MARKER"

var bzip2MemberData, _ = hex.DecodeString("425a6839314159265359b78127cd0000041e0000021000326e58102000310000089a7a9918ca5e1e034f0fec349a64170983265b8bf8bb9229c28485bc093e68")

func bzip2ZIP(t *testing.T, plain, compressed []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	w.RegisterCompressor(zipBzip2, func(out io.Writer) (io.WriteCloser, error) { return &fixedCompressor{out: out, data: compressed}, nil })
	f, err := w.CreateHeader(&zip.FileHeader{Name: "member.txt", Method: zipBzip2})
	if err != nil {
		t.Fatal(err)
	}
	f.Write(plain)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

type fixedCompressor struct {
	out  io.Writer
	data []byte
}

func (c *fixedCompressor) Write(p []byte) (int, error) { return len(p), nil }
func (c *fixedCompressor) Close() error {
	_, err := c.out.Write(c.data)
	return err
}

func TestZIPBzip2Members(t *testing.T) {
	b := bzip2ZIP(t, []byte(bzip2Member), bzip2MemberData)
	if bytes.Contains(b, []byte(bzip2Member)) {
		t.Fatal("fixture leaks the member")
	}
	expectDetected(t, payloadState(t, []byte(bzip2Member)), b, "Payload.Hash0")
	// The reference does not check the CRC; a corrupt stream is incomplete.
	expectDetected(t, payloadState(t, []byte(bzip2Member)), bzip2ZIP(t, []byte("different text"), bzip2MemberData), "Payload.Hash0")
	corrupt := append([]byte(nil), bzip2MemberData...)
	corrupt[20] ^= 0xff
	expectIncomplete(t, payloadState(t), bzip2ZIP(t, []byte(bzip2Member), corrupt), ErrMalformed)
}

func TestUndecodedFormatsAreIncomplete(t *testing.T) {
	arj := func() []byte {
		basic := make([]byte, 30)
		basic[0] = 30
		basic = append(basic, "a.arj\x00\x00"...)
		h := binary.LittleEndian.AppendUint16([]byte{0x60, 0xea}, uint16(len(basic)))
		h = append(h, basic...)
		return binary.LittleEndian.AppendUint32(h, crc32.ChecksumIEEE(basic))
	}()
	sis := make([]byte, 16)
	binary.LittleEndian.PutUint32(sis[4:], 0x10003a12)
	binary.LittleEndian.PutUint32(sis[8:], 0x10000419)
	for name, b := range map[string][]byte{
		"Symbian 9 SIS": append(binary.LittleEndian.AppendUint32(nil, sisX), make([]byte, 100)...),
		"InstallShield": append([]byte("ISc("), make([]byte, 100)...),
	} {
		t.Run(name, func(t *testing.T) { expectIncomplete(t, payloadState(t), b, ErrUnsupported) })
	}
	// Near misses are ordinary data.
	badARJ := append([]byte(nil), arj...)
	badARJ[len(badARJ)-1] ^= 1
	for name, b := range map[string][]byte{
		"CHM version": append(append([]byte("ITSF"), 9, 0, 0, 0, 0x60, 0, 0, 0), make([]byte, 100)...),
		"ARJ CRC":     append(badARJ, make([]byte, 100)...),
		"SIS UIDs":    append(append([]byte(nil), sis[:8]...), make([]byte, 100)...),
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := payloadState(t).visit(b, 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

// FuzzExtractionFormats runs the decoders added for the reference sample
// gaps on mutated inputs: they must stay within the scan limits and never
// panic.
func FuzzExtractionFormats(f *testing.F) {
	gz := func(b []byte) []byte {
		var out bytes.Buffer
		w := gzipWriter(&out)
		w.Write(b)
		w.Close()
		return out.Bytes()
	}
	payload := gz([]byte(payloadMarker))
	items, plain := arjItems("fuzz seed ")
	pe := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	pe.overlay = installShieldOverlay(map[string][]byte{"a": payload}, []string{"a"})
	// Installer data followed by more overlay data.
	segments := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	segments.overlay = bytes.Join([][]byte{
		nsisOverlay(4, nsisBlock(payload, false)),
		installShieldOverlay(map[string][]byte{"b": []byte("record data")}, []string{"b"}),
		[]byte("\x01not a record\x00"), payload,
	}, nil)
	msi := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	msi.overlay = append(installShieldMSIOverlay(f, []string{"a.bin"}, [][]byte{payload}), payload...)
	// Signed: two signature entries, the first carrying a payload after its
	// signature.
	signed := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	signed.cert = append(authenticodeTable([]byte("signed data"), payload), authenticodeTable([]byte("second"), nil)...)
	for _, seed := range [][]byte{
		signed.build(),
		segments.build(),
		msi.build(),
		cpioArchive(cpioNewc, []string{"a"}, [][]byte{payload}, true),
		cpioArchive(cpioBinaryLE, []string{"a"}, [][]byte{payload}, true),
		cpioArchive(cpioODC, []string{"a"}, [][]byte{payload}, true),
		szddCompress([]any{byte('a'), byte('b'), [2]int{4080, 5}}),
		binhexEncode("n", payload, []byte("r"), false),
		rtfDocument(splitHex(append(ole1Header("Package"), ole10Native("a", payload)...))),
		[]byte(`<a href="data:;base64,` + base64.StdEncoding.EncodeToString(payload) + `">`),
		[]byte("From a@b  Thu Jul 31 13:49:50 2008\nMIME-Version: 1.0\nContent-Type: text/plain\n\nx\n"),
		arjArchive(arjMemberOf("a", 1, plain, arjCompress(items, 3)), arjMemberOf("b", 4, plain, arjCompressFast(items))),
		sisPackage(true, true, 1, []sisFixtureFile{{kind: 0, data: [][]byte{payload}}}),
		pe.build(),
	} {
		f.Add(seed)
	}
	db := payloadStateDB(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<16 {
			t.Skip()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s := &scanState{ctx: ctx, db: db, maxFile: 1 << 20, remaining: 4 << 20, maxFiles: 200, maxDepth: 8}
		_, _ = s.visit(data, 0)
		if s.files > s.maxFiles || s.remaining < 0 {
			t.Fatalf("limits exceeded: files=%d remaining=%d", s.files, s.remaining)
		}
	})
}
