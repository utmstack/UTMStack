package signatures

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// typingFTM is a small file-type database written from public format facts.
const typingFTM = `0:0:4d5a:DOS executable:CL_TYPE_ANY:CL_TYPE_MSEXE
0:0:7f454c46:ELF:CL_TYPE_ANY:CL_TYPE_ELF
0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP
0:0:5b616c69617365735d:alias script:CL_TYPE_ANY:CL_TYPE_SCRIPT
0:0:5b616c69617365735d:alias archive:CL_TYPE_ANY:CL_TYPE_POSIX_TAR
0:0:4d4d:TIFF big-endian:CL_TYPE_ANY:CL_TYPE_TIFF
0:0:424d:BMP:CL_TYPE_ANY:CL_TYPE_GRAPHICS
0:0:494433:MP3:CL_TYPE_ANY:CL_TYPE_IGNORED
1:0,1024:255044462d3?2e:PDF in text:CL_TYPE_ANY:CL_TYPE_PDF
1:0,1024:4d5a{60-300}50450000:PE in text:CL_TYPE_ANY:CL_TYPE_MSEXE
1:*:3c68746d6c3e:HTML:CL_TYPE_ANY:CL_TYPE_HTML
1:*:504b0304:ZIP self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX
1:*:2f5247420a49440a:PDF image:CL_TYPE_ANY:CL_TYPE_IGNORED
1:0:cafebabe0000000?:universal binary:CL_TYPE_ANY:CL_TYPE_MACHO_UNIBIN
1:0:cafebabe0000001?:universal binary:CL_TYPE_ANY:CL_TYPE_MACHO_UNIBIN
1:0:cafebabe0000002?:Java class:CL_TYPE_ANY:CL_TYPE_JAVA
1:0:cafebabe0000003?:Java class:CL_TYPE_ANY:CL_TYPE_JAVA
1:*:46726f6d3a20:mail in text:CL_TYPE_TEXT_ASCII:CL_TYPE_MAIL
`

func typingDB(t *testing.T) *Database {
	t.Helper()
	return compileFTM(t, typingFTM)
}

func zipOf(t *testing.T, names ...string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, n := range names {
		f, err := w.CreateHeader(&zip.FileHeader{Name: n, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(strings.Repeat("content of "+n+"\n", 4)))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func utf16le(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		b.WriteByte(c)
		b.WriteByte(0)
	}
	return b.String()
}

func utf16be(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		b.WriteByte(0)
		b.WriteByte(c)
	}
	return b.String()
}

func TestDetermineTypeFollowsReferenceTyping(t *testing.T) {
	d := typingDB(t)
	pe := "0123456789MZ" + strings.Repeat("x", 78) + "PE\x00\x00" + "trailer"
	cases := []struct {
		name string
		data string
		want FileType
	}{
		{"later magic line wins", "[aliases]\nname = value\n", TypePosixTar},
		{"anything starting MM is TIFF", "MMabcdef", TypeTIFF},
		{"anything starting BM is BMP graphics", "BMabcdef", TypeGraphics},
		{"short executable", "MZ\x90\x00\x03\x00", TypeMSEXE},
		{"ignored by magic", "ID3\x03\x00\x00\x00\x00\x00", TypeIgnored},
		{"PDF marker inside text", strings.Repeat("a", 100) + "%PDF-1.4\n1 0 obj\n", TypePDF},
		{"PE header inside binary", pe, TypeMSEXE},
		{"HTML then a zip marker then NUL bytes", "<html>PK\x03\x04\x00\x00", TypeBinaryData},
		{"MSEXE match overrides an earlier HTML match", "<html>" + pe, TypeMSEXE},
		{"HTML is higher than an earlier MSEXE match", pe + "<html>", TypeHTML},
		{"an ignored type pattern returns at once", pe + "/RGB\nID\n<html>", TypeIgnored},
		{"Java class", "\xca\xfe\xba\xbe\x00\x00\x00\x2a\x00\x01", TypeJava},
		{"universal binary", "\xca\xfe\xba\xbe\x00\x00\x00\x02\x00\x01", TypeMachOUnibin},
		{"class file of Java 20 is not typed", "\xca\xfe\xba\xbe\x00\x00\x00\x40\x00\x01", TypeBinaryData},
		{"required type text: mail marker in text", "hello\nFrom: someone\n", TypeMail},
		{"required type text: not in UTF-16", "\x00\x01hello From: someone\n", TypeTextUTF16BE},
		{"UTF-16LE HTML without a mark", utf16le("<html><body>hi</body></html>"), TypeHTMLUTF16},
		{"UTF-16LE HTML with a mark", "\xff\xfe" + utf16le("<html><body>hi</body></html>"), TypeHTMLUTF16},
		{"UTF-16BE HTML with a mark (converted)", "\xfe\xff" + utf16be("<html><body>hi</body></html>"), TypeHTML},
		{"UTF-16BE HTML by its leading pattern", utf16be("<?xml?><html><body>hi</body></html>"), TypeHTML},
		{"UTF-16LE text without HTML", utf16le("plain words only"), TypeTextUTF16BE},
	}
	for _, c := range cases {
		got, complete := d.DetermineType([]byte(c.data))
		if got != c.want || !complete {
			t.Errorf("%s: DetermineType = %s (complete %v), want %s", c.name, got, complete, c.want)
		}
	}
}

func TestDetermineTypeRecognizesOfficeZips(t *testing.T) {
	d := typingDB(t)
	cases := []struct {
		name  string
		names []string
		want  FileType
	}{
		{"typed by its first member", []string{"word/document.xml", "word/styles.xml"}, TypeOOXMLWord},
		{"typed from the central directory", []string{"[Content_Types].xml", "_rels/.rels", "docProps/app.xml", "xl/workbook.xml"}, TypeOOXMLXL},
		{"presentation", []string{"ppt/presentation.xml"}, TypeOOXMLPPT},
		{"likely names but no document part", []string{"[Content_Types].xml", "_rels/.rels", "docProps/app.xml", "other/x.xml"}, TypeZIP},
		{"plain zip", []string{"a.txt", "b.txt", "c.txt", "d.txt"}, TypeZIP},
		{"word part after three plain members", []string{"a.txt", "b.txt", "c.txt", "word/document.xml"}, TypeZIP},
	}
	for _, c := range cases {
		if got, _ := d.DetermineType(zipOf(t, c.names...)); got != c.want {
			t.Errorf("%s: DetermineType = %s, want %s", c.name, got, c.want)
		}
	}
}

func TestDetermineTypeOfficeZipWithMaximumComment(t *testing.T) {
	d := typingDB(t)
	data := zipOf(t, "[Content_Types].xml", "_rels/.rels", "docProps/app.xml", "xl/workbook.xml")
	end := len(data) - 22
	if string(data[end:end+4]) != "PK\x05\x06" {
		t.Fatal("fixture lacks end-of-central-directory record")
	}
	binary.LittleEndian.PutUint16(data[end+20:], 65535)
	data = append(data, bytes.Repeat([]byte{'x'}, 65535)...)
	if got, complete := d.DetermineType(data); got != TypeOOXMLXL || !complete {
		t.Fatalf("maximum-comment OOXML ZIP typed as %s (complete %v)", got, complete)
	}
}

// UCS-4 in the 2143 and 3412 byte orders is converted as big-endian UCS-4,
// as the reference converts it (checked on its debug output: "00 00 FF FE"
// then big-endian markup is HTML, markup in the marked order is not, and a
// "FE FF 00 00" mark stops the conversion at once).
func TestDetermineTypeConvertsUnusualUCS4AsBigEndian(t *testing.T) {
	d := typingDB(t)
	ucs4 := func(s string, order [4]int) string {
		var out []byte
		for _, r := range s {
			b := []byte{byte(r >> 24), byte(r >> 16), byte(r >> 8), byte(r)}
			out = append(out, b[order[0]], b[order[1]], b[order[2]], b[order[3]])
		}
		return string(out)
	}
	html := "<html><body><p>marker</p></body></html>"
	be, o2143, o3412 := [4]int{0, 1, 2, 3}, [4]int{1, 0, 3, 2}, [4]int{2, 3, 0, 1}
	cases := []struct {
		name string
		data string
		html bool
	}{
		{"2143 mark, big-endian markup", "\x00\x00\xff\xfe" + ucs4(html, be), true},
		{"2143 mark, 2143 markup", "\x00\x00\xff\xfe" + ucs4(html, o2143), false},
		{"3412 mark, big-endian markup", "\xfe\xff\x00\x00" + ucs4(html, be), false},
		{"3412 mark, 3412 markup", "\xfe\xff\x00\x00" + ucs4(html, o3412), false},
		{"2143 markup without a mark", ucs4(html, o2143), false},
		{"3412 markup without a mark", ucs4(html, o3412), false},
	}
	for _, c := range cases {
		got, complete := d.DetermineType([]byte(c.data))
		if !complete || (got == TypeHTML) != c.html {
			t.Errorf("%s: %s complete=%v, want HTML=%v", c.name, got, complete, c.html)
		}
	}
}

func TestDetermineTypeOnlyTypePatternsInsideFirstKiB(t *testing.T) {
	d := typingDB(t)
	late := strings.Repeat("a", 2000) + "%PDF-1.4\n"
	if got, _ := d.DetermineType([]byte(late)); got != TypeTextASCII {
		t.Fatalf("a marker after the first KiB typed the object %s", got)
	}
	tarLike := make([]byte, 600)
	copy(tarLike, tarHeader("ustar  \x00", false))
	if got, _ := d.DetermineType(tarLike); got != TypePosixTar {
		t.Fatalf("binary tar header typed %s, want POSIX_TAR", got)
	}
}
