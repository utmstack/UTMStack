package engine

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func hx(s string) string { return hex.EncodeToString([]byte(s)) }

// pipelineState compiles the given signature files and returns a scan with
// room for large objects.
func pipelineState(t *testing.T, files map[string]string) *scanState {
	t.Helper()
	b := map[string][]byte{}
	for name, text := range files {
		b[name] = []byte(text)
	}
	db, err := signatures.Compile(b)
	if err != nil {
		t.Fatal(err)
	}
	return &scanState{ctx: context.Background(), db: db, maxFile: 64 << 20, remaining: 256 << 20, maxFiles: 1000, maxDepth: 16}
}

func unknownReasonOf(err error) string {
	var u *signatures.UnknownError
	if errors.As(err, &u) {
		return u.Reason
	}
	return ""
}

func expectScan(t *testing.T, s *scanState, data []byte, wantSig, wantUnknown string) {
	t.Helper()
	sig, err := s.visit(data, 0)
	if sig != wantSig || unknownReasonOf(err) != wantUnknown || wantUnknown == "" && err != nil {
		t.Fatalf("scan: %q, %v; want %q with unknown reason %q", sig, err, wantSig, wantUnknown)
	}
}

func le16(s string) []byte {
	var b []byte
	for _, c := range []byte(s) {
		b = append(b, c, 0)
	}
	return b
}

// As in the reference, HTML is matched only through its normalized views
// (lower-cased page, text, scripts), never raw: neither its bytes nor its
// hash.
func TestPipelineMatchesNormalizedHTMLOnly(t *testing.T) {
	page := []byte("<html><body>EVILMARKER in a <a href=\"LinkMarker\">page</a></body></html>")
	s := pipelineState(t, map[string]string{"a.ndb": "Html.Raw:0:*:" + hx("EVILMARKER") + "\n"})
	expectScan(t, s, page, "", "")
	s = pipelineState(t, map[string]string{"a.ndb": "Html.Target:3:*:" + hx("evilmarker in a") + "\n"})
	expectScan(t, s, page, "Html.Target", "")
	// The text view keeps a link target's case.
	s = pipelineState(t, map[string]string{"a.ndb": "Html.Link:0:*:" + hx(" LinkMarker page") + "\n"})
	expectScan(t, s, page, "Html.Link", "")
	sum := md5.Sum(page)
	s = pipelineState(t, map[string]string{"a.hdb": fmt.Sprintf("%x:%d:Html.Hash\n", sum, len(page))})
	expectScan(t, s, page, "", "")
	// The page view is matched with its hash.
	view := normalizeHTML(page).noComment
	sum = md5.Sum(view)
	s = pipelineState(t, map[string]string{"a.hdb": fmt.Sprintf("%x:%d:Html.View\n", sum, len(view))})
	expectScan(t, s, page, "Html.View", "")
	// A script is matched as normalized JavaScript.
	s = pipelineState(t, map[string]string{"a.ndb": "Html.Script:7:*:" + hx("var n000=\"evil\";") + "\n"})
	expectScan(t, s, []byte("<html><script>var x = 'ev' + 'il';</script></html>"), "Html.Script", "")
}

func TestPipelineScansUTF16HTMLRaw(t *testing.T) {
	// The reference skips the raw scan only for HTML, not for UTF-16 HTML:
	// wide rules still see a UTF-16 page (AMSI submits scripts this way).
	page := le16("<html><body>EvilPayloadMarker</body></html>")
	s := pipelineState(t, map[string]string{"a.ldb": "Wide.Marker;Engine:81-255,Target:0;0;" + hx("EvilPayloadMarker") + "::w\n"})
	expectScan(t, s, page, "Wide.Marker", "")
}

func TestPipelineMatchesTextRawAndNormalized(t *testing.T) {
	s := pipelineState(t, map[string]string{"a.ndb": "Upper.Raw:0:*:" + hx("UPPERMARK") + "\nText.Rule:7:*:" + hx("lowered words") + "\n"})
	expectScan(t, s, []byte("some text UPPERMARK here"), "Upper.Raw", "")
	expectScan(t, s, []byte("LOWERED \t  Words appear"), "Text.Rule", "")
	// Root 7 runs on normalized text only; raw text never meets it.
	s = pipelineState(t, map[string]string{"a.ndb": "Spaced:7:*:" + hx("two  spaces") + "\n"})
	expectScan(t, s, []byte("two  spaces"), "", "")
}

func TestPipelineUTF16ScriptsMatchNormalizedAndWide(t *testing.T) {
	script := le16("Invoke-Evil -Now")
	s := pipelineState(t, map[string]string{"a.ndb": "Script.Lower:0:*:" + hx("invoke-evil") + "\n"})
	expectScan(t, s, script, "Script.Lower", "")
	s = pipelineState(t, map[string]string{"a.ndb": "Script.Wide:0:*:" + hex.EncodeToString(le16("Invoke-Evil")) + "\n"})
	expectScan(t, s, script, "Script.Wide", "")
	// No UTF-8 transcoding exists as a representation.
	s = pipelineState(t, map[string]string{"a.ndb": "Utf8.Only:0:*:" + hx("caf\xc3\xa9") + "\n"})
	expectScan(t, s, le16("un caf\xe9 noir"), "", "")
}

func TestPipelineSkipsIgnoredTypesEntirely(t *testing.T) {
	s := pipelineState(t, map[string]string{
		"a.ftm": "0:0:494433:MP3:CL_TYPE_ANY:CL_TYPE_IGNORED\n",
		"a.ndb": "Ignored.Mark:0:*:" + hx("IGNOREDMARK") + "\n",
	})
	expectScan(t, s, []byte("ID3\x03\x00\x00 IGNOREDMARK"), "", "")
	expectScan(t, s, []byte("plain IGNOREDMARK"), "Ignored.Mark", "")
}

func TestPipelineDoesNotNormalizeHugeText(t *testing.T) {
	s := pipelineState(t, map[string]string{"a.ndb": "Text.Rule:7:*:" + hx("needle") + "\nOther:0:*:" + hx("ZZZZZZZZ") + "\n"})
	huge := append([]byte("NEEDLE "), bytes.Repeat([]byte("a"), 21<<20)...)
	expectScan(t, s, huge, "", "")
	expectScan(t, s, []byte("NEEDLE here"), "Text.Rule", "")
}

func TestPipelineMatchesRawMailWithRootFour(t *testing.T) {
	body := base64.StdEncoding.EncodeToString([]byte("MAILMARKER inside the part"))
	message := []byte("Received: from relay\r\nFrom: a@example.test\r\nSubject: Invoice due\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: text/plain\r\nContent-Transfer-Encoding: base64\r\n\r\n" + body + "\r\n")
	expectScan(t, pipelineState(t, map[string]string{"a.ndb": "Mail.Raw:4:*:" + hx("Subject: Invoice") + "\n"}), message, "Mail.Raw", "")
	expectScan(t, pipelineState(t, map[string]string{"a.ndb": "Mail.Decoded:4:*:" + hx("MAILMARKER") + "\n"}), message, "", "")
	expectScan(t, pipelineState(t, map[string]string{"a.ndb": "Part.Generic:0:*:" + hx("MAILMARKER") + "\n"}), message, "Part.Generic", "")
}

func TestPipelineTypedRootsForImagesAndFlash(t *testing.T) {
	s := pipelineState(t, map[string]string{"a.ndb": "Png.Rule:5:*:" + hx("PNGMARK") + "\n"})
	expectScan(t, s, []byte("\x89PNG\r\n\x1a\n image data PNGMARK"), "Png.Rule", "")
	expectScan(t, s, []byte("text with PNGMARK"), "", "")
	var z bytes.Buffer
	w := zlib.NewWriter(&z)
	w.Write([]byte("\x78\x00\x05\x5f\x00\x00\x0f\xa0\x00\x00\x0c\x01\x00 flash body SWFMARK"))
	w.Close()
	cws := append([]byte("CWS\x0a"), binary.LittleEndian.AppendUint32(nil, uint32(8+29+len(" flash body SWFMARK")))...)
	cws = append(cws, z.Bytes()...)
	s = pipelineState(t, map[string]string{"a.ndb": "Swf.Rule:11:*:" + hx("SWFMARK") + "\n"})
	sig, _ := s.visit(cws, 0)
	if sig != "Swf.Rule" {
		t.Fatalf("decompressed Flash child: %q", sig)
	}
}

func TestPipelineContainerRules(t *testing.T) {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	f, _ := zw.Create("member.txt")
	f.Write([]byte("a ZIPMARKER member"))
	zw.Close()
	s := pipelineState(t, map[string]string{"a.ldb": "InZip;Engine:81-255,Container:CL_TYPE_ZIP,Target:0;0;" + hx("ZIPMARKER") + "\n"})
	expectScan(t, s, b.Bytes(), "InZip", "")
	expectScan(t, s, []byte("a ZIPMARKER alone"), "", "")
	script := bytes.Repeat([]byte("install script "), 20)
	payload := append(bytes.Repeat([]byte("installed file "), 30), payloadMarker...)
	installer := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	installer.overlay = nsisOverlay(len(script), append(nsisBlock(nsisLZMA(t, script), true), nsisBlock(nsisLZMA(t, payload), true)...))
	s = pipelineState(t, map[string]string{"a.ldb": "InNsis;Engine:81-255,Container:CL_TYPE_NULSFT,Target:0;0;" + hx(payloadMarker) + "\n"})
	expectScan(t, s, installer.build(), "InNsis", "")
	if strings.Contains(string(installer.build()), payloadMarker) {
		t.Fatal("fixture leaks the marker")
	}
}

func TestPipelineExtractsArchivesFoundInsideOtherFiles(t *testing.T) {
	ftm := "0:0:89504e47:PNG:CL_TYPE_ANY:CL_TYPE_PNG\n0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP\n" +
		"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n1:*:377abcaf271c:7z self-extractor:CL_TYPE_ANY:CL_TYPE_7ZSFX\n"
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	f, _ := zw.Create("inside.txt")
	f.Write([]byte("an EMBEDDEDMARKER member"))
	zw.Close()
	polyglot := append([]byte("\x89PNG\r\n\x1a\n image bytes then "), b.Bytes()...)
	s := pipelineState(t, map[string]string{"a.ftm": ftm, "a.ndb": "Embedded.Mark:0:*:" + hx("EMBEDDEDMARKER") + "\n"})
	if bytes.Contains(polyglot, []byte("EMBEDDEDMARKER")) {
		// The member is stored; the raw bytes hold the marker. Compress it.
		b.Reset()
		zw = zip.NewWriter(&b)
		f, _ = zw.CreateHeader(&zip.FileHeader{Name: "inside.txt", Method: zip.Deflate})
		f.Write(bytes.Repeat([]byte("an EMBEDDEDMARKER member "), 20))
		zw.Close()
		polyglot = append([]byte("\x89PNG\r\n\x1a\n image bytes then "), b.Bytes()...)
	}
	expectScan(t, s, polyglot, "Embedded.Mark", "")
	// A 7-Zip archive after other bytes is extracted the same way.
	archive, err := os.ReadFile("../sevenzip/testdata/lzma2.7z")
	if err != nil {
		t.Fatal(err)
	}
	s = pipelineState(t, map[string]string{"a.ftm": ftm, "a.ndb": "Payload.Test:0:*:" + hx(payloadMarker) + "\n"})
	expectScan(t, s, append([]byte("\x89PNG\r\n\x1a\n image bytes then "), archive...), "Payload.Test", "")
}

// machoSlice is a minimal 64-bit Mach-O ending in tail.
func machoSlice(tail string) []byte {
	b := make([]byte, 256)
	binary.LittleEndian.PutUint32(b[0:], 0xfeedfacf)
	binary.LittleEndian.PutUint32(b[4:], 0x01000007)
	binary.LittleEndian.PutUint32(b[16:], 1)
	binary.LittleEndian.PutUint32(b[32:], 0x80000028) // LC_MAIN
	binary.LittleEndian.PutUint32(b[36:], 24)
	return append(b, tail...)
}

func TestPipelineScansUniversalBinarySlices(t *testing.T) {
	first, second := machoSlice("TAIL"), machoSlice("more")
	fat := make([]byte, 8+2*20)
	binary.BigEndian.PutUint32(fat[0:], 0xcafebabe)
	binary.BigEndian.PutUint32(fat[4:], 2)
	offset := uint32(len(fat))
	for i, slice := range [][]byte{first, second} {
		arch := fat[8+20*i:]
		binary.BigEndian.PutUint32(arch[8:], offset)
		binary.BigEndian.PutUint32(arch[12:], uint32(len(slice)))
		offset += uint32(len(slice))
	}
	fat = append(append(fat, first...), second...)
	s := pipelineState(t, map[string]string{"a.ndb": "End:0:EOF-4:" + hx("TAIL") + "\n"})
	expectScan(t, s, fat, "End", "")
	// The universal binary's own header always fails: a match at its end
	// does not count, only one at the end of a slice.
	fat = append(fat[:len(fat)-len(second)], machoSlice("mid!")...)
	fat = append(fat, "TAIL"...)
	binary.BigEndian.PutUint32(fat[8+20:12+20], 0) // drop the first slice from the table
	binary.BigEndian.PutUint32(fat[4:], 1)
	copy(fat[8:28], fat[28:48])
	expectScan(t, s, fat, "", "")
}
