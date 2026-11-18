package engine

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/ascii85"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ulikunitz/xz/lzma"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func zlibFixture(t testing.TB, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	if _, err := w.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func pdfFixture(filter string, b []byte) []byte {
	return []byte(fmt.Sprintf("%%PDF-1.7\n1 0 obj\n<< /Length %d %s >>\nstream\n%s\nendstream\nendobj\n%%%%EOF\n", len(b), filter, b))
}
func TestPDFExtractsFiltersStringsAndObjects(t *testing.T) {
	e := nativeTestEngine(t)
	marker := []byte(testMarker)
	a85 := make([]byte, ascii85.MaxEncodedLen(len(marker)))
	n := ascii85.Encode(a85, marker)
	a85 = append(a85[:n], []byte("~>")...)
	rl := append([]byte{byte(len(marker) - 1)}, marker...)
	rl = append(rl, 128)
	hexString := []byte("<" + hex.EncodeToString(marker) + ">")
	for name, data := range map[string][]byte{
		"flate":           pdfFixture("/Filter /FlateDecode", zlibFixture(t, marker)),
		"hex":             pdfFixture("/Filter /ASCIIHexDecode", []byte(hex.EncodeToString(marker)+">")),
		"ascii85":         pdfFixture("/Filter /ASCII85Decode", a85),
		"runlength":       pdfFixture("/Filter /RunLengthDecode", rl),
		"chained":         pdfFixture("/Filter [/ASCIIHexDecode /FlateDecode]", []byte(hex.EncodeToString(zlibFixture(t, marker))+">")),
		"hex-string":      []byte("%PDF-1.7\n1 0 obj\n" + string(hexString) + "\nendobj\n%%EOF"),
		"literal-escapes": []byte("%PDF-1.7\n1 0 obj\n(UTMSTACK\\055NATIVE\\055HARMLESS\\055TEST)\nendobj\n%%EOF"),
		"content-string":  pdfFixture("/Filter /FlateDecode", zlibFixture(t, append(hexString, []byte(" Tj")...))),
		"object-stream":   pdfFixture("/Type /ObjStm /N 1 /First 4 /Filter /FlateDecode", zlibFixture(t, append([]byte("2 0 "), hexString...))),
		"forward-length":  []byte(fmt.Sprintf("%%PDF-1.7\n1 0 obj\n<< /Length 2 0 R /Filter /ASCIIHexDecode >>\nstream\n%s>\nendstream\nendobj\n2 0 obj\n%d\nendobj\n%%%%EOF", hex.EncodeToString(marker), 2*len(marker)+1)),
		// Ghostscript writes the data straight into endstream, with no end
		// of line, and gives the exact length in a later object.
		"forward-length-no-eol": []byte(fmt.Sprintf("%%PDF-1.4\n5 0 obj\n<</Length 6 0 R/Filter /FlateDecode>>\nstream\n%sendstream\nendobj\n6 0 obj\n%d\nendobj\n%%%%EOF\n", zlibFixture(t, marker), len(zlibFixture(t, marker)))),
	} {
		t.Run(name, func(t *testing.T) {
			if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
				t.Fatal(clean, sig, err)
			}
		})
	}
}
func TestPDFIncompleteNeverClean(t *testing.T) {
	e := nativeTestEngine(t)
	// (An unknown filter name is skipped, and a /Length that does not end at
	// endstream is replaced by the end found there, as in the reference:
	// TestPDFFilterChainsAsReference, TestPDFStructuresTheReferenceReads.)
	for _, b := range [][]byte{
		pdfFixture("/Filter /FlateDecode", []byte("broken")),
		pdfFixture("/Filter /ASCIIHexDecode", []byte("666f6f")),
		pdfFixture("/Encrypt 2 0 R", []byte("test")),
	} {
		if clean, sig, err := e.Scan(context.Background(), b); clean || sig != "" || err == nil {
			t.Fatal(clean, sig, err)
		}
	}
	if clean, _, err := e.Scan(context.Background(), pdfFixture("/Filter /FlateDecode", zlibFixture(t, []byte("benign")))); !clean || err != nil {
		t.Fatal(clean, err)
	}
}
func swfFixture(t testing.TB, body []byte, codec byte) []byte {
	t.Helper()
	f := append([]byte{'F', 'W', 'S', 13, 0, 0, 0, 0, 0, 0, 12, 1, 0}, body...)
	f = append(f, 0, 0)
	binary.LittleEndian.PutUint32(f[4:8], uint32(len(f)))
	switch codec {
	case 'C':
		out := append([]byte(nil), f[:8]...)
		out[0] = 'C'
		return append(out, zlibFixture(t, f[8:])...)
	case 'Z':
		var encoded bytes.Buffer
		w, err := (lzma.WriterConfig{DictCap: 1 << 16, SizeInHeader: true, Size: int64(len(f) - 8)}).NewWriter(&encoded)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write(f[8:]); err != nil {
			t.Fatal(err)
		}
		if err = w.Close(); err != nil {
			t.Fatal(err)
		}
		b := encoded.Bytes()
		out := append([]byte(nil), f[:8]...)
		out[0] = 'Z'
		out = binary.LittleEndian.AppendUint32(out, uint32(len(b)-13))
		out = append(out, b[:5]...)
		return append(out, b[13:]...)
	default:
		return f
	}
}
func TestFlashExtraction(t *testing.T) {
	e := nativeTestEngine(t)
	for _, codec := range []byte{'F', 'C', 'Z'} {
		// DefineBinaryData includes a nested gzip. Its uncompressed marker is
		// absent from both the outer SWF and its tag payload.
		payload := append(make([]byte, 6), gzipFixture(t, []byte(testMarker))...)
		tag := binary.LittleEndian.AppendUint16(nil, 87<<6|63)
		tag = binary.LittleEndian.AppendUint32(tag, uint32(len(payload)))
		tag = append(tag, payload...)
		t.Run(string(codec), func(t *testing.T) {
			if clean, sig, err := e.Scan(context.Background(), swfFixture(t, tag, codec)); clean || sig != "Native.Test" || err != nil {
				t.Fatal(clean, sig, err)
			}
			if clean, _, err := e.Scan(context.Background(), swfFixture(t, nil, codec)); !clean || err != nil {
				t.Fatal(clean, err)
			}
		})
	}
}
func TestPhishingInEnginePreservesHTMLSemantics(t *testing.T) {
	db, err := signatures.Compile(map[string][]byte{"test.pdb": []byte("H:amazon.com")})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		text   string
		detect bool
	}{
		{`<a href="https://evil.test">amazon.com</a>`, true},
		{`<a href="https://evil.test">amazon&amp;#46;com</a>`, false},
		{`<a href="https://evil.test" href="https://amazon.com">amazon.com</a>`, true},
	} {
		s := scanState{ctx: context.Background(), db: db, maxFile: 1 << 20, remaining: 2 << 20, maxFiles: 100, maxDepth: 8}
		sig, err := s.visit([]byte(tt.text), 0)
		if (sig != "") != tt.detect || err != nil {
			t.Fatal(tt.text, sig, err)
		}
	}
}
func FuzzDocuments(f *testing.F) {
	f.Add(pdfFixture("/Filter /FlateDecode", zlibFixture(f, []byte("benign"))))
	f.Add(pdfFixture("/Filter /LZWDecode /DecodeParms << /EarlyChange 1 >>", pdfLZWLiteralFixture(bytes.Repeat([]byte{'A'}, 300), 1)))
	f.Add(pdfFixture("/Filter /FlateDecode /DecodeParms << /Predictor 15 /Columns 4 >>", zlibFixture(f, []byte{1, 'A', 1, 1, 1})))
	f.Add(swfFixture(f, nil, 'Z'))
	f.Add([]byte("%PDF-1.7"))
	for _, name := range []string{"stream-rc4-40", "stream-aes-256", "string-aes-256-objstm"} {
		if b, err := os.ReadFile("testdata/pdfcrypt/" + name + ".pdf"); err == nil {
			f.Add(b)
		}
	}
	e := nativeTestEngineForFuzz(f)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		s := scanState{ctx: ctx, db: e.current.Load().db, maxFile: 1 << 20, remaining: 2 << 20, maxFiles: 100, maxDepth: 8}
		switch {
		case bytes.HasPrefix(b, []byte("%PDF-")):
			s.pdf(b, 0)
		case len(b) >= 3 && strings.Contains("FCZ", string(b[0])) && string(b[1:3]) == "WS":
			s.swf(b, 0)
		}
	})
}
