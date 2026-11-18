package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures (testdata/pdfcrypt) are qpdf's encryptions of one page whose
// content stream ("stream-*") or info title string ("string-*") holds the test
// marker: RC4 40 and 128 bits (V2, and V4 crypt filters), AES 128 bits,
// AES 256 bits (revisions 5 and 6, and with object streams), metadata left
// in clear text, and a user password ("-pw"). The reference (ClamAV 1.4.3,
// with a rule for the marker) decrypts the streams of every file with an
// empty user password, but not strings: a string's marker is found only in
// clear text or inside a decrypted object stream. A file with a user password
// cannot be decrypted (the reference raises an encryption alert when asked to).
func TestPDFDecryptsStreamsAsReference(t *testing.T) {
	e := nativeTestEngine(t)
	files, err := filepath.Glob("testdata/pdfcrypt/*.pdf")
	if err != nil || len(files) != 20 {
		t.Fatalf("%d fixtures, %v", len(files), err)
	}
	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".pdf")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		clean, sig, err := e.Scan(context.Background(), data)
		switch {
		case strings.HasSuffix(name, "-pw"):
			if clean || sig != "" || !errors.Is(err, ErrUnsupported) {
				t.Errorf("%s: (%v, %q, %v), want incomplete as encrypted", name, clean, sig, err)
			}
		case strings.HasPrefix(name, "stream-") || name == "string-plain" || name == "string-aes-256-objstm":
			if sig != "Native.Test" || err != nil {
				t.Errorf("%s: (%v, %q, %v), want the marker", name, clean, sig, err)
			}
		default:
			if !clean || sig != "" || err != nil {
				t.Errorf("%s: (%v, %q, %v), want clean", name, clean, sig, err)
			}
		}
	}
}

// Filters, as the reference applies them (checked on chains around a Flate
// stream holding the marker): Flate, LZW, ASCIIHex, ASCII85 and RunLength
// are decoded; at an image filter (DCT, JPX, JBIG2, CCITT fax) or Crypt it
// stops and the data decoded so far is analyzed; an unknown name is skipped.
// Parameters of filters that are not decoded are not read.
func TestPDFFilterChainsAsReference(t *testing.T) {
	e := nativeTestEngine(t)
	flate := zlibFixture(t, []byte("some text "+testMarker+" more text"))
	cases := map[string]bool{
		"/FlateDecode":              true,
		"[/FlateDecode /DCTDecode]": true, "[/DCTDecode /FlateDecode]": false, "/DCTDecode": false,
		"[/FlateDecode /JPXDecode]": true, "[/JBIG2Decode /FlateDecode]": false,
		"[/Fl /CCF]": true, "[/CCITTFaxDecode /FlateDecode]": false, "[/FlateDecode /Crypt]": true,
		"[/Foo /FlateDecode]": true, "[/FlateDecode /Foo]": true, "/Foo": false,
		"[/FlateDecode /CCITTFaxDecode] /DecodeParms [null << /K -1 /Columns 8 /Rows 2 >>]": true,
		"/DCTDecode /DecodeParms << /ColorTransform 1 >>":                                   false,
	}
	for filter, found := range cases {
		clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter "+filter, flate))
		if err != nil || found != (sig == "Native.Test") || !found && !clean {
			t.Errorf("%s: (%v, %q, %v), want found=%v", filter, clean, sig, err, found)
		}
	}
}

// Structures the reference reads that this engine rejected (groff's PDF
// manuals): an empty object ("N G obj endobj"); a stream length defined in
// an object stream, with the data running into endstream; a length object
// that does not exist (the reference then calculates the length from
// endstream); and image data that the content-string search cannot parse,
// which the reference does not search at all.
func TestPDFStructuresTheReferenceReads(t *testing.T) {
	e := nativeTestEngine(t)
	marked := zlibFixture(t, []byte(testMarker))
	stream := func(obj, length string, body []byte) string {
		return fmt.Sprintf("%s obj\n<< /Length %s /Filter /FlateDecode >>\nstream\n%sendstream\nendobj\n", obj, length, body)
	}
	head := "%PDF-1.5\n1 0 obj\n<< /Type /Catalog >>\nendobj\n"
	cases := map[string]string{
		"empty object":          head + "2 0 obj\nendobj\n" + stream("3 0", fmt.Sprint(len(marked)), marked) + "%%EOF\n",
		"missing length object": head + stream("3 0", "9 0 R", marked) + "%%EOF\n",
	}
	objstm := zlibFixture(t, []byte(fmt.Sprintf("9 0 %d", len(marked))))
	cases["length too short"] = head + stream("3 0", fmt.Sprint(len(marked)/2), marked) + "%%EOF\n"
	cases["length too long"] = head + stream("3 0", fmt.Sprint(len(marked)+40), marked) + "%%EOF\n"
	cases["repeated keys, the first one used"] = head + fmt.Sprintf("3 0 obj\n<< /Length %d /Filter /FlateDecode /Filter /DCTDecode >>\nstream\n%sendstream\nendobj\n", len(marked), marked) + "%%EOF\n"
	cases["length in an object stream"] = head + fmt.Sprintf("4 0 obj\n<< /Type /ObjStm /N 1 /First 4 /Length %d /Filter /FlateDecode >>\nstream\n%s\nendstream\nendobj\n", len(objstm), objstm) + stream("3 0", "9 0 R", marked) + stream("5 0", "7", []byte("1234567")) + "%%EOF\n"
	for name, doc := range cases {
		if clean, sig, err := e.Scan(context.Background(), []byte(doc)); sig != "Native.Test" || err != nil {
			t.Errorf("%s: (%v, %q, %v), want the marker", name, clean, sig, err)
		}
	}
	gray := bytes.Repeat([]byte{0x85, 0x84, 0x86, 0x87}, 4096)
	rle := []byte{}
	for i := 0; i < len(gray); i += 128 {
		rle = append(rle, 127)
		rle = append(rle, gray[i:i+128]...)
	}
	rle = append(rle, 128)
	image := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<< /Filter [ /RunLengthDecode ] /Width 128 /Height 128 /BitsPerComponent 8 /Length %d >>\nstream\n%s\nendstream\nendobj\n%%%%EOF\n", len(rle), rle)
	benign := zlibFixture(t, []byte("benign"))
	for name, doc := range map[string]string{
		"image data":   image,
		"empty object": head + "2 0 obj\nendobj\n" + stream("3 0", fmt.Sprint(len(benign)), benign) + "%%EOF\n",
	} {
		if clean, sig, err := e.Scan(context.Background(), []byte(doc)); !clean || sig != "" || err != nil {
			t.Errorf("%s: (%v, %q, %v), want clean", name, clean, sig, err)
		}
	}
}
