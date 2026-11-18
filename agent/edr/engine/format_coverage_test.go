package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// These fixtures carry inert text only. Their marker is compressed before it
// reaches any wrapper so that matching the outer bytes cannot pass the test.
const formatCoverageMarker = "UTMSTACK-FORMAT-COVERAGE-HARMLESS-MARKER"

func formatZipFixture(t *testing.T, method uint16, payload []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.CreateHeader(&zip.FileHeader{Name: "../../inert.fixture", Method: method})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func formatRawZipFixture(t testing.TB, method uint16, packed, payload []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.CreateRaw(&zip.FileHeader{Name: "inert.fixture", Method: method,
		CompressedSize64: uint64(len(packed)), UncompressedSize64: uint64(len(payload)), CRC32: crc32.ChecksumIEEE(payload)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(packed); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func formatZstd(t *testing.T, payload []byte) []byte {
	t.Helper()
	w, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	return w.EncodeAll(payload, nil)
}

func formatCoverageFixtures(t *testing.T, payload []byte) map[string][]byte {
	t.Helper()
	gz := gzipFixture(t, payload)
	pdf := pdfFixture("/Filter /FlateDecode", zlibFixture(t, gz))
	encoded := base64.StdEncoding.EncodeToString(gz)
	mail := func(contentType, encoding string, body []byte) []byte {
		return append([]byte("MIME-Version: 1.0\r\nContent-Type: "+contentType+"\r\nContent-Transfer-Encoding: "+encoding+"\r\n\r\n"), body...)
	}
	zipData := formatZipFixture(t, zip.Deflate, gz)
	return map[string][]byte{
		"zip-zstd-gzip":             formatRawZipFixture(t, zip93Zstd, formatZstd(t, gz), gz),
		"zip-zstd-legacy-gzip":      formatRawZipFixture(t, zip20ZstdLegacy, formatZstd(t, gz), gz),
		"zip-stored-gzip":           formatZipFixture(t, zip.Store, gz),
		"zip-deflate-gzip":          zipData,
		"zip-pdf-gzip":              formatZipFixture(t, zip.Deflate, pdf),
		"pdf-flate-gzip":            pdf,
		"pdf-hex-flate-gzip":        pdfFixture("/Filter [/ASCIIHexDecode /FlateDecode]", []byte(hex.EncodeToString(zlibFixture(t, gz))+">")),
		"pdf-hex-string-gzip":       []byte("%PDF-1.7\n1 0 obj\n<" + hex.EncodeToString(gz) + ">\nendobj\n%%EOF\n"),
		"mail-base64-gzip":          mail("application/octet-stream", "base64", []byte(encoded)),
		"mail-quoted-printable-zip": mail("application/zip", "quoted-printable", []byte("="+strings.ToUpper(strings.Join(formatHexBytes(zipData), "=")))),
		"mail-base64-pdf":           mail("application/pdf", "base64", []byte(base64.StdEncoding.EncodeToString(pdf))),
		"html-base64-gzip":          []byte(`<html><body><img src="data:application/octet-stream;base64,` + encoded + `"></body></html>`),
		"html-percent-gzip":         []byte(`<html><body><object data='data:application/octet-stream,` + urlEscape(gz) + `'></object></body></html>`),
		"html-default-type-gzip":    []byte(`<html><iframe src="data:;base64,` + encoded + `"></iframe></html>`),
		"html-data-uri-pdf":         []byte(`<html><object data="data:application/pdf;base64,` + base64.StdEncoding.EncodeToString(pdf) + `"></object></html>`),
		"html-style-gzip":           []byte(`<html><style>body{background:url(data:application/octet-stream;base64,` + encoded + `)}</style></html>`),
	}
}

func formatHexBytes(data []byte) []string {
	parts := make([]string, len(data))
	for i, b := range data {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return parts
}

func TestFormatCoverageMatrix(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, positive := range []bool{false, true} {
		payload, want := []byte("ordinary harmless format coverage control"), ""
		label := "negative"
		if positive {
			payload, want, label = []byte(formatCoverageMarker), "Format.Coverage", "positive"
		}
		for name, data := range formatCoverageFixtures(t, payload) {
			t.Run(label+"/"+name, func(t *testing.T) {
				if positive && bytes.Contains(data, payload) {
					t.Fatal("fixture exposes marker without extraction")
				}
				expectScan(t, pipelineState(t, rules), data, want, "")
				// Explicit opt-in exports the exact independently scanned oracle
				// inputs; the normal regression run writes nothing outside TempDir.
				if dir := os.Getenv("EDR_FORMAT_FIXTURE_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, label+"-"+name+".fixture"), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func TestFormatCoverageLimits(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for name, data := range formatCoverageFixtures(t, []byte("ordinary harmless format coverage control")) {
		for _, limit := range []string{"depth", "files", "file-size", "scan-size"} {
			t.Run(name+"/"+limit, func(t *testing.T) {
				s := pipelineState(t, rules)
				switch limit {
				case "depth":
					s.maxDepth = 0
				case "files":
					s.maxFiles = 1
				case "file-size":
					s.maxFile = int64(len(data) - 1)
				case "scan-size":
					s.remaining = int64(len(data) - 1)
				}
				if sig, err := s.visit(data, 0); sig != "" || !errors.Is(err, ErrLimit) {
					t.Fatalf("scan: %q, %v; want resource limit", sig, err)
				}
			})
		}
	}
}

func TestHTMLDataEnvelopeKeepsRawRulesAndLimits(t *testing.T) {
	page := []byte(`<html><body><img src="data:application/octet-stream;base64,b3JkaW5hcnkgaGFybWxlc3MgZml4dHVyZSB0ZXh0"></body></html>`)
	rules := map[string]string{"format.ndb": "Never:0:*:" + hx(formatCoverageMarker) + "\n"}
	expectScan(t, pipelineState(t, rules), page, "", "")
	for _, value := range []string{
		"data:;charset=UTF-8;base64,b3JkaW5hcnkgaGFybWxlc3MgdGV4dA==",
		"data:text/plain;base64=ordinary,ordinary%20harmless%20text",
		"data:;charset=UTF-8,ordinary%20harmless%20text",
	} {
		expectScan(t, pipelineState(t, rules), []byte(`<html><img src="`+value+`"></html>`), "", "")
	}
	// The exact normalized envelope remains available to both generic and
	// mail-specific rules, with the HTML/mail ancestry intact.
	for _, target := range []int{0, 4} {
		rules = map[string]string{"format.ndb": fmt.Sprintf("Envelope.Raw:%d:*:%s\n", target, hx("From html-normalise"))}
		expectScan(t, pipelineState(t, rules), page, "Envelope.Raw", "")
	}
	envelope := normalizeHTML(page).rfc2397[0].compatibility
	sum := md5.Sum(envelope)
	rules = map[string]string{"format.hdb": fmt.Sprintf("%x:%d:Envelope.Hash\n", sum, len(envelope))}
	expectScan(t, pipelineState(t, rules), page, "Envelope.Hash", "")
	rules = map[string]string{"format.ldb": "Envelope.Ancestry;Engine:81-255,Container:CL_TYPE_HTML,Target:4;0;" + hx("From html-normalise") + "\n"}
	expectScan(t, pipelineState(t, rules), page, "Envelope.Ancestry", "")
	// A real sender's invalid MIME parameter must not acquire the normalizer's
	// private exception.
	s := pipelineState(t, map[string]string{"format.ndb": "Never:0:*:" + hx(formatCoverageMarker) + "\n"})
	_, err := s.visit([]byte("MIME-Version: 1.0\nContent-Type: text/plain;base64\nContent-Transfer-Encoding: base64\n\nSGVsbG8gd29ybGQ="), 0)
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("sender MIME: %v; want malformed", err)
	}
}

func TestZIPNewCodecsAndUnsupportedMethods(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, method := range []uint16{zip93Zstd, zip20ZstdLegacy} {
		gz := gzipFixture(t, []byte(formatCoverageMarker))
		packed := formatZstd(t, gz)
		archive := formatRawZipFixture(t, method, packed, gz)
		expectScan(t, pipelineState(t, rules), archive, "Format.Coverage", "")
		// The same codec is supported in bounded local-header ZIP records
		// embedded in a different object (self-extractor recognition).
		sfxRules := map[string]string{"format.ndb": rules["format.ndb"], "format.ftm": sfxTypes}
		local := localHeader(0, method, uint32(len(packed)), uint32(len(gz)), "inert.fixture", packed)
		expectScan(t, pipelineState(t, sfxRules), append(append([]byte(nil), pngPrefix...), local...), "Format.Coverage", "")
		// Both the expanded file size and the aggregate scan size bound the
		// codec before the compressed output becomes a child object.
		big := bytes.Repeat([]byte("ordinary harmless expansion data "), 4096)
		bomb := formatRawZipFixture(t, method, formatZstd(t, big), big)
		for _, budget := range []string{"file", "scan"} {
			s := pipelineState(t, rules)
			if budget == "file" {
				s.maxFile = int64(len(bomb) + 1024)
			} else {
				s.remaining = int64(len(bomb) + 1024)
			}
			if _, err := s.visit(bomb, 0); !errors.Is(err, ErrLimit) {
				t.Fatalf("method %d %s budget: %v", method, budget, err)
			}
		}
		broken := formatRawZipFixture(t, method, []byte("not a zstandard stream"), []byte("ordinary harmless data"))
		if _, err := pipelineState(t, rules).visit(broken, 0); !errors.Is(err, ErrMalformed) {
			t.Fatalf("method %d malformed stream: %v", method, err)
		}
	}
	for _, method := range []uint16{1, 95, 99, 21589} {
		data := []byte("ordinary harmless content whose codec is unavailable")
		archive := formatRawZipFixture(t, method, data, data)
		if sig, err := pipelineState(t, rules).visit(archive, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("method %d: %q, %v", method, sig, err)
		}
		sfxRules := map[string]string{"format.ndb": rules["format.ndb"], "format.ftm": sfxTypes}
		local := localHeader(0, method, uint32(len(data)), uint32(len(data)), "inert.fixture", data)
		if sig, err := pipelineState(t, sfxRules).visit(append(append([]byte(nil), pngPrefix...), local...), 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("embedded method %d: %q, %v", method, sig, err)
		}
	}
}

func TestFormatEncryptedAndMalformedBoundaries(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	archive := formatZipFixture(t, zip.Deflate, []byte("ordinary harmless encrypted-content control"))
	for _, magic := range [][]byte{[]byte("PK\x03\x04"), []byte("PK\x01\x02")} {
		at := bytes.Index(archive, magic)
		flagOffset := 6
		if magic[2] == 1 {
			flagOffset = 8
		}
		binary.LittleEndian.PutUint16(archive[at+flagOffset:], binary.LittleEndian.Uint16(archive[at+flagOffset:])|1)
	}
	if sig, err := pipelineState(t, rules).visit(archive, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("encrypted ZIP: %q, %v", sig, err)
	}
	for name, data := range map[string][]byte{
		"pdf-broken-flate":   pdfFixture("/Filter /FlateDecode", []byte("broken compressed stream")),
		"pdf-no-eof":         []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n"),
		"mail-no-boundary":   []byte("MIME-Version: 1.0\nContent-Type: multipart/mixed\n\nordinary text"),
		"mail-broken-base64": []byte("MIME-Version: 1.0\nContent-Type: application/octet-stream\nContent-Transfer-Encoding: base64\n\n@@@"),
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := pipelineState(t, rules).visit(data, 0); sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("scan: %q, %v; want malformed", sig, err)
			}
		})
	}
	// Cancelled analysis returns promptly through each top-level parser path.
	for name, data := range formatCoverageFixtures(t, []byte("ordinary harmless cancellation control")) {
		t.Run("cancelled/"+name, func(t *testing.T) {
			s := pipelineState(t, rules)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			s.ctx = ctx
			if _, err := s.visit(data, 0); !errors.Is(err, context.Canceled) {
				t.Fatalf("scan: %v", err)
			}
		})
	}
}
