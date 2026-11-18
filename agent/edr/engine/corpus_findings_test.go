package engine

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Regression tests for clean files that came back incomplete, or slowly,
// when real Debian 13 and Windows Server 2022 files were scanned with every
// accepted official rule loaded.

// pseudoRandom returns deterministic, incompressible bytes.
func pseudoRandom(n int, seed uint64) []byte {
	b := make([]byte, n)
	for i := range b {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		b[i] = byte(seed)
	}
	return b
}

// chanceGzip is a gzip member header that passes the self-extractor header
// checks, followed by bytes that are not deflate data, as in the COFF symbol
// table of a Wine DLL. chanceBzip2 is the same for a bzip2 block header.
func chanceGzip() []byte {
	return append([]byte{0x1f, 0x8b, 8, 0, 0xe0, 0xbc, 1, 0, 4, 0}, pseudoRandom(4096, 1)...)
}
func chanceBzip2() []byte {
	return append([]byte("BZh91AY&SY"), pseudoRandom(4096, 2)...)
}

func TestChanceArchiveFixturesDoNotDecode(t *testing.T) {
	r, err := gzip.NewReader(bytes.NewReader(chanceGzip()))
	if err == nil {
		_, err = io.Copy(io.Discard, r)
	}
	if err == nil || err == io.ErrUnexpectedEOF {
		t.Fatalf("gzip fixture decodes: %v", err)
	}
	if _, err := io.Copy(io.Discard, bzip2.NewReader(bytes.NewReader(chanceBzip2()))); err == nil || err == io.ErrUnexpectedEOF {
		t.Fatalf("bzip2 fixture decodes: %v", err)
	}
	for _, b := range [][]byte{chanceGzip(), chanceBzip2()} {
		if !archiveStart(b) {
			t.Fatalf("%x does not pass the header checks", b[:10])
		}
	}
}

// Authenticode data holds 1f 8b and BZh by chance. Neither the table at the
// end of the file nor one followed by more data is searched for archives.
func TestPECertificateTableIsNotOverlayData(t *testing.T) {
	cert := authenticodeTable(append(append(pseudoRandom(64, 3), chanceGzip()...), chanceBzip2()...), nil)
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.cert = cert
	signed := f.build()
	for name, b := range map[string][]byte{
		"at the end":           signed,
		"before trailing data": append(append([]byte(nil), signed...), []byte("data appended after signing")...),
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := payloadState(t).visit(b, 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	// Data after the table is still overlay data, and an archive there is analyzed.
	b := append(append([]byte(nil), signed...), gzipFixture(t, []byte(payloadMarker))...)
	if sig, err := payloadState(t).visit(b, 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("after the table: sig=%q err=%v", sig, err)
	}
	// So is data before a table that does not end the file.
	f.overlay = gzipFixture(t, []byte(payloadMarker))
	b = append(f.build(), []byte("data appended after signing")...)
	if sig, err := payloadState(t).visit(b, 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("before the table: sig=%q err=%v", sig, err)
	}
}

// A header that passes the self-extractor checks but starts no stream that
// decodes is not an archive; a real archive that was cut short stays
// incomplete.
func TestPEOverlayChanceArchiveHeaders(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	for name, overlay := range map[string][]byte{
		"gzip":  append(append([]byte("COFF symbols "), chanceGzip()...), pseudoRandom(512, 4)...),
		"bzip2": append(append([]byte("COFF symbols "), chanceBzip2()...), pseudoRandom(512, 5)...),
		"both":  append(chanceGzip(), chanceBzip2()...),
	} {
		t.Run(name, func(t *testing.T) {
			f.overlay = overlay
			if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	// A real archive after a chance header is still found.
	f.overlay = append(chanceGzip(), gzipFixture(t, []byte(payloadMarker))...)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("after a chance header: sig=%q err=%v", sig, err)
	}
	truncatedGzip := gzipFixture(t, pseudoRandom(64<<10, 6))
	for name, overlay := range map[string][]byte{
		"gzip cut short":        truncatedGzip[:len(truncatedGzip)/2],
		"gzip header cut short": gzipFixture(t, []byte(payloadMarker))[:9],
		"bzip2 cut short":       bzip2MemberData[:len(bzip2MemberData)-12],
	} {
		t.Run(name, func(t *testing.T) {
			f.overlay = append([]byte("setup data "), overlay...)
			if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || err == nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

// Every header is decoded a little before it is rejected; crafted data full
// of them stops at a bound instead of multiplying that work.
func TestPEOverlayChanceArchiveHeadersAreBounded(t *testing.T) {
	// BFINAL with the reserved block type fails at once.
	header := []byte{0x1f, 0x8b, 8, 0, 0, 0, 0, 0, 0, 3, 0x07, 'x'}
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = bytes.Repeat(header, 100)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || err != nil {
		t.Fatalf("few headers: sig=%q err=%v", sig, err)
	}
	f.overlay = bytes.Repeat(header, maxEmbeddedCandidates+1)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("many headers: sig=%q err=%v", sig, err)
	}
}

// A section whose virtual extent passes SizeOfImage, reported for a Linux
// kernel's EFI stub, is scanned as the reference scans it; only file-backed
// bytes are read.
func TestPEVirtualExtentPastImageIsLoadable(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.sections = append(f.sections, peFixtureSection{name: ".data", va: 0x2000, vsize: 0x67000, raw: []byte(payloadMarker)})
	f.imageSize = 0x3000
	b := f.build()
	img, err := parsePEImage(b)
	if err != nil || img.packed || img.overlay != len(b) {
		t.Fatalf("img=%+v err=%v", img, err)
	}
	if sig, err := payloadState(t).visit(b, 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
	b = bytes.Replace(b, []byte(payloadMarker), bytes.Repeat([]byte{'.'}, len(payloadMarker)), 1)
	if sig, err := payloadState(t).visit(b, 0); sig != "" || err != nil {
		t.Fatalf("clean: sig=%q err=%v", sig, err)
	}
}

// pdfEmbeddedFile builds a PDF with one attachment, as writers store them: a
// file specification whose /EF dictionary names the embedded file stream.
func pdfEmbeddedFile(t *testing.T, content []byte, streamDict string) []byte {
	t.Helper()
	body := zlibFixture(t, content)
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R /Names << /EmbeddedFiles << /Names [(a.html) 4 0 R] >> >> >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>",
		"<< /Type /Filespec /F (a.html) /UF (a.html) /EF << /F 5 0 R >> >>",
		fmt.Sprintf("<< %s /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream", streamDict, len(body), body),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	for i, o := range objects {
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	b.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\n%%EOF\n")
	return b.Bytes()
}

// Attachments are data, not page content: text files whose bytes are not PDF
// syntax are analyzed as the files they are.
func TestPDFEmbeddedTextFilesAreData(t *testing.T) {
	files := map[string]string{
		"html":   "<html><body><p>Totals (draft] > summary)) {see below}</p></body></html>\n",
		"rtf":    "{\\rtf1\\ansi{\\fonttbl{\\f0 Arial;}}\\f0 Quarterly report) \\par}\n",
		"binhex": string(binhexEncode("a.txt", []byte("notes (draft] > final"), nil, false)),
		"text":   "closing ) and ] and > without openers\n",
	}
	for _, dict := range []string{"/Type /EmbeddedFile", "/Type /EmbeddedFile /Subtype /text#2Fhtml /Params << /Size 70 >>", ""} {
		for name, content := range files {
			t.Run(name+" "+dict, func(t *testing.T) {
				marked := []byte(strings.Replace(content, "\n", payloadMarker+"\n", 1))
				if name == "html" {
					// HTML is matched through its normalized views; the
					// text view keeps a link target's case.
					if sig, err := payloadState(t).visit(pdfEmbeddedFile(t, []byte(content), dict), 0); sig != "" || err != nil {
						t.Fatalf("sig=%q err=%v", sig, err)
					}
					marked := []byte(strings.Replace(content, "</p>", "</p><a href=\""+payloadMarker+"\">x</a>", 1))
					if sig, err := payloadState(t).visit(pdfEmbeddedFile(t, marked, dict), 0); sig != "Payload.Test" || err != nil {
						t.Fatalf("sig=%q err=%v", sig, err)
					}
					return
				}
				if sig, err := payloadState(t).visit(pdfEmbeddedFile(t, []byte(content), dict), 0); sig != "" || err != nil {
					t.Fatalf("sig=%q err=%v", sig, err)
				}
				if sig, err := payloadState(t).visit(pdfEmbeddedFile(t, marked, dict), 0); sig != "Payload.Test" || err != nil {
					t.Fatalf("marked: sig=%q err=%v", sig, err)
				}
			})
		}
	}
	// A page content stream that is not PDF syntax is analyzed as data; the
	// string search, which the reference does not make, stops there without
	// making it incomplete.
	if sig, err := payloadState(t).visit(pdfFixture("", []byte("BT (unbalanced ] ET")), 0); sig != "" || err != nil {
		t.Fatalf("content: sig=%q err=%v", sig, err)
	}
}

// Strings without escapes are already covered by the raw bytes around them.
// Decoded strings of one object share one representation, so a text-heavy
// document does not pay a full rule pass per string.
func TestPDFStringsShareRepresentations(t *testing.T) {
	var content bytes.Buffer
	for i := 0; i < 2000; i++ {
		fmt.Fprintf(&content, "BT (plain text string %d) Tj (escaped\\040string %d) Tj <%x> Tj ET\n", i, i, fmt.Sprintf("hex string %d", i))
	}
	s := payloadState(t)
	if sig, err := s.visit(pdfFixture("/Filter /FlateDecode", zlibFixture(t, content.Bytes())), 0); sig != "" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
	if s.files > 20 {
		t.Fatalf("%d objects analyzed", s.files)
	}
	utf16 := []byte{0xfe, 0xff}
	for _, c := range payloadMarker {
		utf16 = append(utf16, 0, byte(c))
	}
	half := len(payloadMarker) / 2
	for name, object := range map[string]string{
		// Escapes and hex strings still reveal hidden bytes, whichever string
		// of the object holds them.
		"escaped": "<< /A (first\\040string) /B (" + payloadMarker[:half] + "\\" + fmt.Sprintf("%03o", payloadMarker[half]) + payloadMarker[half+1:] + ") >>",
		"hex":     fmt.Sprintf("[(first\\040string) <%x> (last\\040string)]", payloadMarker),
	} {
		t.Run(name, func(t *testing.T) {
			b := []byte("%PDF-1.7\n1 0 obj\n" + object + "\nendobj\n%%EOF\n")
			if sig, err := payloadState(t).visit(b, 0); sig != "Payload.Test" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	// UTF-16 strings are matched raw and as normalized (lowercased) text,
	// as the reference matches text; no UTF-8 transcoding is a
	// representation, so an upper-case ASCII rule does not match them.
	for name, object := range map[string]string{
		"utf-16":   fmt.Sprintf("<< /Title (first\\040string) /Author <%x> /Subject (last\\040string) >>", utf16),
		"only one": fmt.Sprintf("<%x>", utf16),
	} {
		t.Run(name, func(t *testing.T) {
			b := []byte("%PDF-1.7\n1 0 obj\n" + object + "\nendobj\n%%EOF\n")
			if sig, err := payloadState(t).visit(b, 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

func TestHTMLDocumentMarkers(t *testing.T) {
	for _, c := range []struct {
		data string
		html bool
	}{
		{"<html>", true}, {"<HTML>", true}, {"<Html>", true}, {"<hTML>", false}, {"<html lang=en>", false},
		{"<head>", true}, {"<HEAD>", true}, {"<Head>", true}, {"<header>", false},
		{"<script>", true}, {"<script src=x>", true}, {"<SCRIPT\n", true}, {"<Script/", true}, {"<scripts>", false}, {"<script", false},
		{"<img src=x>", true}, {"<IMG\t", true}, {"<Img>", true}, {"<iMG ", false}, {"<image>", false},
		{"<iframe ", true}, {"<IFRAME>", true}, {"<object>", true}, {"<Object ", true}, {"<OBJECT/", true},
		{"<table>", true}, {"<TABLE ", true}, {"<Table>", false},
		{`<a href="x">`, true}, {`<A HREF=x>`, true}, {`<a class=b hRef=x>`, true}, {`<a class=b HRef=x>`, true},
		{`<a Href`, true}, {`<a hrEF=x`, true}, {`<a HrEf=x`, false}, {`<a hre`, false}, {"<a\thref=x>", false}, {`href=x <a `, false},
		{"<%@ LANGUAGE = VBScript %>", true}, {" <%@ LANGUAGE = VBScript %>", false},
		{"x < y and y > z", false}, {"", false},
		{"\x00<html>", false}, {"\x7f<html>", false}, {"\x1b[0m<html>", true}, {"\xef\xbb\xbf<html>", true}, {"caf\xc3\xa9 <table>", true},
		{strings.Repeat("a", 1024) + "\x00<html>", true}, {strings.Repeat("a", 1023) + "\x00<html>", false},
	} {
		if got := htmlDocument([]byte(c.data)); got != c.html {
			t.Errorf("%q: got %v", c.data, got)
		}
	}
}

func phishingState(t *testing.T) *scanState {
	t.Helper()
	db, err := signatures.Compile(map[string][]byte{"test.pdb": []byte("H:amazon.com")})
	if err != nil {
		t.Fatal(err)
	}
	return &scanState{ctx: t.Context(), db: db, maxFile: 4 << 20, remaining: 8 << 20, maxFiles: 100, maxDepth: 8}
}

// Link analysis runs on HTML: text holding one of the reference engine's HTML
// file-type markers, or a mail part declared as HTML. Binary data and text
// without markers that merely contain '<' are not HTML.
func TestPhishingRunsOnHTMLOnly(t *testing.T) {
	link := `<a href="https://evil.test">amazon.com</a>`
	longText := `<a href="https://example.test">` + strings.Repeat("text ", 2000) + `</a>`
	binary := "GIF89a\x10\x00\x10\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04"
	for _, c := range []struct {
		name, data string
		detect     bool
		incomplete bool
	}{
		{"document", "<html>\n<body>" + link + "</body></html>", true, false},
		{"fragment", link, true, false},
		{"script marker", "<script>var x = 1;</script><p><A\thref='https://evil.test'>amazon.com</A>", true, false},
		{"marker after 1 KiB", strings.Repeat("prose ", 400) + link, true, false},
		{"long link text", longText, false, true},
		{"binary", binary + link, false, false},
		{"binary long link text", binary + longText, false, false},
		{"text without markers", "Plain notes.\n<a\thref=\"https://evil.test\">amazon.com</a>\n", false, false},
		{"text without markers, long link text", "Plain notes.\n<a\thref=\"https://example.test\">" + strings.Repeat("x ", 5000), false, false},
		{"NUL after 1 KiB", strings.Repeat("prose ", 200) + "\x00" + link, true, false},
		{"NUL in first 1 KiB", "prose \x00" + link, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			sig, err := phishingState(t).visit([]byte(c.data), 0)
			if (sig != "") != c.detect || (err != nil) != c.incomplete {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	// A mail part declared as HTML is HTML whatever markers it holds.
	body := "<p><a\nhref=\"https://evil.test\">amazon.com</a></p>\n"
	for _, c := range []struct {
		contentType string
		detect      bool
	}{{"text/html; charset=utf-8", true}, {"text/plain", false}} {
		single := "From: a@example.test\nTo: b@example.test\nSubject: s\nMIME-Version: 1.0\nContent-Type: " + c.contentType + "\n\n" + body
		multi := "From: a@example.test\nMIME-Version: 1.0\nContent-Type: multipart/alternative; boundary=b\n\n--b\nContent-Type: " + c.contentType + "\nContent-Transfer-Encoding: quoted-printable\n\n" + strings.ReplaceAll(body, "=", "=3D") + "--b--\n"
		for _, m := range []string{single, multi} {
			if sig, err := phishingState(t).visit([]byte(m), 0); (sig != "") != c.detect || err != nil {
				t.Fatalf("%s: sig=%q err=%v", c.contentType, sig, err)
			}
		}
	}
}

// authenticodeTable returns a certificate table with one WIN_CERTIFICATE
// entry: PKCS#7 SignedData as a DER SEQUENCE holding content, then extra
// bytes that the entry's length also covers, padded to eight bytes.
func authenticodeTable(content, extra []byte) []byte {
	entry := binary.LittleEndian.AppendUint32(nil, 0)
	entry = binary.LittleEndian.AppendUint16(entry, 0x0200)
	entry = binary.LittleEndian.AppendUint16(entry, 2)
	entry = append(entry, 0x30, 0x82, byte(len(content)>>8), byte(len(content)))
	entry = append(entry, content...)
	entry = append(entry, extra...)
	entry = append(entry, make([]byte, (8-len(entry)%8)%8)...)
	binary.LittleEndian.PutUint32(entry, uint32(len(entry)))
	return entry
}

// chanceZIPHeader passes the self-extractor ZIP header checks but starts no
// archive.
func chanceZIPHeader() []byte {
	h := []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00")
	h = append(h, pseudoRandom(16, 7)...)
	h = append(h, 5, 0, 0, 0)
	return append(h, pseudoRandom(48, 8)...)
}

// Only the signature itself is left unsearched. Bytes an entry carries after
// its signature, as in certificate-padding attacks, and a security directory
// that names data holding no signature are overlay data like any other.
func TestPECertificateTableHidesOnlyTheSignature(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.cert = authenticodeTable(append(append(pseudoRandom(64, 9), chanceZIPHeader()...), chanceGzip()...), nil)
	for name, b := range map[string][]byte{
		"signature":             f.build(),
		"signature before data": append(f.build(), []byte("data appended after signing")...),
		"signature with zero tail": func() []byte {
			g := f
			g.cert = authenticodeTable(chanceZIPHeader(), make([]byte, 7))
			return g.build()
		}(),
		// UEFI binaries such as shim carry one entry per signature.
		"two signatures": func() []byte {
			g := f
			g.cert = append(authenticodeTable(chanceZIPHeader(), make([]byte, 3)), authenticodeTable(append(pseudoRandom(40, 13), chanceZIPHeader()...), nil)...)
			return g.build()
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := payloadState(t).visit(b, 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	// The archives end the table, so that no padding follows them.
	payload := gzipFixture(t, []byte(payloadMarker))
	padded := authenticodeTable(pseudoRandom(64+(8-(12+64+len(payload))%8)%8, 10), payload)
	unsigned := binary.LittleEndian.AppendUint32(nil, uint32(8+len(payload)))
	unsigned = append(binary.LittleEndian.AppendUint16(binary.LittleEndian.AppendUint16(unsigned, 0x0200), 2), payload...)
	for name, cert := range map[string][]byte{
		"archive after the signature": padded,
		"archive as the table":        zipFixture(t, []byte(payloadMarker)),
		"entry without a signature":   unsigned,
	} {
		t.Run(name, func(t *testing.T) {
			g := f
			g.cert = cert
			if sig, err := payloadState(t).visit(g.build(), 0); sig != "Payload.Test" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

// Executables are recognized by a loadable header, not a short magic, and the
// reference engine finds them anywhere in a file: one stored inside the
// signature is extracted like one stored anywhere else.
func TestPEStoredInSignatureIsExtracted(t *testing.T) {
	inner := simplePE(bytes.Repeat([]byte{0xcc}, 0x200)).build()
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.cert = authenticodeTable(append(append(pseudoRandom(64, 11), inner...), pseudoRandom(64, 12)...), nil)
	b := f.build()
	start := bytes.Index(b[1:], inner[:0x40]) + 1
	if sig, err := payloadState(t, b[start:]).visit(b, 0); sig != "Payload.Hash0" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// Object streams, which PDF 1.5 writers use for most dictionaries, may hold
// the file specification that names an attachment. The attachment is data
// even when its stream precedes the object stream and has no /Type.
func TestPDFAttachmentNamedFromObjectStream(t *testing.T) {
	for name, content := range map[string]string{
		"clean":  "<html><body><p>Totals (draft] > summary)) {see below}</p></body></html>\n",
		"marked": "<html><body><p>Totals (draft] > <a href=\"" + payloadMarker + "\">x</a>)) {see below}</p></body></html>\n",
	} {
		t.Run(name, func(t *testing.T) {
			file := zlibFixture(t, []byte(content))
			member := "<< /Type /Filespec /F (a.html) /UF (a.html) /EF << /F 4 0 R >> >>"
			header := "5 0 "
			objstm := header + member
			var b bytes.Buffer
			b.WriteString("%PDF-1.7\n")
			objects := []string{
				"<< /Type /Catalog /Pages 2 0 R /Names << /EmbeddedFiles << /Names [(a.html) 5 0 R] >> >> >>",
				"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
				"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>",
				fmt.Sprintf("<< /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream", len(file), file),
				fmt.Sprintf("<< /Type /ObjStm /N 1 /First %d /Length %d >>\nstream\n%s\nendstream", len(header), len(objstm), objstm),
			}
			for i, o := range objects {
				n := i + 1
				if n == 5 {
					n = 6
				}
				fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", n, o)
			}
			b.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\n%%EOF\n")
			// The attachment is found through the object stream.
			sig, err := payloadState(t).visit(b.Bytes(), 0)
			if name == "marked" && (sig != "Payload.Test" || err != nil) || name == "clean" && (sig != "" || err != nil) {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

// Every stream is analyzed once, even an object stream whose /Type is stored
// in another object stream that is read after it.
func TestPDFObjectStreamTypedFromAnotherObjectStream(t *testing.T) {
	hidden := zlibFixture(t, []byte("8 0 ("+payloadMarker+")"))
	typeHeader := "7 0 "
	typeStream := typeHeader + "/ObjStm"
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n")
	fmt.Fprintf(&b, "1 0 obj\n<< /Type 7 0 R /N 1 /First 4 /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(hidden), hidden)
	fmt.Fprintf(&b, "2 0 obj\n<< /Type /ObjStm /N 1 /First %d /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(typeHeader), len(typeStream), typeStream)
	b.WriteString("%%EOF\n")
	if sig, err := payloadState(t).visit(b.Bytes(), 0); sig != "Payload.Test" {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// PostScript calculator functions and PostScript XObjects are text in
// another language, not content syntax: their braces are not PDF tokens.
func TestPDFPostScriptStreamsAreData(t *testing.T) {
	for name, c := range map[string]struct{ dict, body string }{
		"type 4 function":    {"/FunctionType 4 /Domain [0 1] /Range [0 1 0 1]", "{ 360 mul sin 2 div exch 360 mul sin 2 div add }"},
		"PostScript XObject": {"/Type /XObject /Subtype /PS", "/Helvetica findfont 12 scalefont setfont 72 72 moveto (Hi) show { pop } forall"},
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := payloadState(t).visit(pdfFixture(c.dict, []byte(c.body)), 0); sig != "" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
			marked := pdfFixture(c.dict, []byte(c.body+" % "+payloadMarker))
			if sig, err := payloadState(t).visit(marked, 0); sig != "Payload.Test" || err != nil {
				t.Fatalf("marked: sig=%q err=%v", sig, err)
			}
		})
	}
}
