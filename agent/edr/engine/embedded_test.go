package engine

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const embeddedTypes = "0:0:89504e47:PNG:CL_TYPE_ANY:CL_TYPE_PNG\n0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP\n" +
	"0:0:4d5a:PE:CL_TYPE_ANY:CL_TYPE_MSEXE\n0:0:23407e5e:Script Encoder:CL_TYPE_ANY:CL_TYPE_SCRENC\n" +
	"0:0:78617221:XAR:CL_TYPE_ANY:CL_TYPE_XAR\n" +
	"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n" +
	"1:*:526172211a0700:rar self-extractor:CL_TYPE_ANY:CL_TYPE_RARSFX\n" +
	"1:*:0a(4d|6d)(49|69)(4d|6d)(45|65)2d(56|76)657273696f6e3a20{-1024}0a(43|63)6f6e74656e742d(54|74)7970653a20:MHTML file:CL_TYPE_ANY:CL_TYPE_MHTML:90\n"

func deflated(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w, _ := flate.NewWriter(&out, flate.BestCompression)
	w.Write(b)
	w.Close()
	return out.Bytes()
}

// hiddenMember is a ZIP local header with a deflated member holding the
// marker, which appears only once the member is decompressed.
func hiddenMember(t *testing.T) []byte {
	member := bytes.Repeat([]byte("member text EMBEDDEDMARKER "), 20)
	packed := deflated(t, member)
	return localHeader(0, 8, uint32(len(packed)), uint32(len(member)), "m.txt", packed)
}

func embeddedState(t *testing.T, rules map[string]string) *scanState {
	files := map[string]string{"a.ftm": embeddedTypes}
	for k, v := range rules {
		files[k] = v
	}
	return pipelineState(t, files)
}

func TestArchivesInsideProgramSectionsAreExtracted(t *testing.T) {
	s := embeddedState(t, map[string]string{"a.ndb": "Embedded.Mark:0:*:" + hx("EMBEDDEDMARKER") + "\n"})
	code := append(bytes.Repeat([]byte{0x90}, 64), hiddenMember(t)...)
	code = append(code, bytes.Repeat([]byte{0xcc}, 64)...)
	pe := simplePE(code).build()
	expectScan(t, s, pe, "Embedded.Mark", "")
}

func TestEmbeddedArchiveMarkersThatDoNotParseAreNothing(t *testing.T) {
	s := embeddedState(t, map[string]string{"a.ndb": "Other:0:*:57575757\n"})
	garbage := append([]byte("Rar!\x1a\x07\x00"), bytes.Repeat([]byte{0xa5}, 200)...)
	png := append(append([]byte(nil), pngPrefix...), garbage...)
	expectScan(t, s, png, "", "")
	code := append(bytes.Repeat([]byte{0x90}, 64), garbage...)
	expectScan(t, s, simplePE(code).build(), "", "")
}

func TestEmbeddedZipMembersSitInAZipLayer(t *testing.T) {
	polyglot := append(append([]byte(nil), pngPrefix...), hiddenMember(t)...)
	marker := hx("EMBEDDEDMARKER")
	s := embeddedState(t, map[string]string{"a.ldb": "In.Zip;Engine:81-255,Target:0,Container:CL_TYPE_ZIP;0;" + marker + "\n"})
	expectScan(t, s, polyglot, "In.Zip", "")
	s = embeddedState(t, map[string]string{"a.ldb": "In.Png;Engine:81-255,Target:0,Container:CL_TYPE_PNG;0;" + marker + "\n"})
	expectScan(t, s, polyglot, "", "")
	s = embeddedState(t, map[string]string{"a.cdb": "Zip.Member:CL_TYPE_ZIP:*:m\\.txt:*:*:*:*:*:*\n", "a.ndb": "Other:0:*:57575757\n"})
	expectScan(t, s, polyglot, "Zip.Member", "")
}

func TestMHTMLFoundInsideBinaryIsParsedAsMail(t *testing.T) {
	s := embeddedState(t, map[string]string{"a.ndb": "Embedded.Mark:0:*:" + hx("EMBEDDEDMARKER") + "\n"})
	header := "Project-Id-Version: x\nMIME-Version: 1.0\nContent-Type: text/plain; charset=UTF-8\nContent-Transfer-Encoding: 8bit\n\x00"
	// A message catalog: binary tables, then a header entry with MIME
	// fields, found by the raw scan past the typing window.
	tables := append([]byte("\xde\x12\x04\x95\x00\x00\x00\x00\x02\x00\x00\x00\x1c\x00\x00\x00\x00\x00"), bytes.Repeat([]byte{0, 1, 2, 3}, 600)...)
	expectScan(t, s, append(append([]byte(nil), tables...), header...), "", "")
	// The same entry inside the typing window types the whole object MHTML.
	expectScan(t, s, append(append([]byte(nil), tables[:40]...), header...), "", "")
	// A message typed MHTML is parsed as mail and its parts analyzed.
	message := "X-Note: hello\nMIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=b\n\n--b\nContent-Type: application/octet-stream\n" +
		"Content-Transfer-Encoding: base64\n\n" + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("part EMBEDDEDMARKER "), 10)) + "\n--b--\n"
	if typ, _ := s.db.DetermineType([]byte(message)); typ != signatures.TypeMHTML {
		t.Fatalf("message typed %v", typ)
	}
	expectScan(t, s, []byte(message), "Embedded.Mark", "")
}

func TestFormatsTheReferenceParsesAndThisEngineDoesNotAreIncomplete(t *testing.T) {
	s := embeddedState(t, map[string]string{"a.ndb": "Other:0:*:57575757\n"})
	for data, reason := range map[string]string{
		"#@~^encoded script body that this engine cannot decode": "format-screnc",
	} {
		sig, err := s.visit([]byte(data), 0)
		if sig != "" || !errors.Is(err, ErrUnsupported) || !strings.Contains(err.Error(), reason) {
			t.Errorf("%q: %q, %v; want incomplete %s", data, sig, err, reason)
		}
	}
	// XAR is parsed now; a truncated archive remains incomplete as malformed.
	expectIncomplete(t, embeddedState(t, map[string]string{"a.ndb": "Other:0:*:57575757\n"}),
		[]byte("xar!\x00\x1c\x00\x01 archive table of contents"), ErrMalformed)
}

func TestSixteenBitProgramsAreNotMalformed(t *testing.T) {
	// A 16-bit program or font: MZ, then an NE header. The reference does
	// not analyze it as a PE and reports nothing.
	ne := make([]byte, 4096)
	copy(ne, "MZ")
	binary.LittleEndian.PutUint32(ne[0x3c:], 0x80)
	copy(ne[0x80:], "NE font resources")
	s := embeddedState(t, map[string]string{"a.ndb": "Other:0:*:57575757\n"})
	expectScan(t, s, ne, "", "")
	// A header offset past the end of the file is the same.
	past := append([]byte(nil), ne[:0x100]...)
	binary.LittleEndian.PutUint32(past[0x3c:], 0x10000)
	expectScan(t, s, past, "", "")
}
