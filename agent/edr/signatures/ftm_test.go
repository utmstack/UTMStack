package signatures

import (
	"strings"
	"testing"
)

func compileFTM(t *testing.T, ftm string) *Database {
	t.Helper()
	d, err := Compile(map[string][]byte{"test.ftm": []byte(ftm), "test.ndb": []byte("Marker:0:*:6d61726b6572\n")})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestBuiltInOneNoteTypingWithoutFTM(t *testing.T) {
	d := compileTest(t, "hdb", "aa15bcf478d165efd2065190eb473bcb:544:Child\n")
	header := []byte("\xe4\x52\x5c\x7b\x8c\xd8\xa7\x4d\xae\xb1\x53\x78\xd0\x29\x96\xd3")
	if typ, complete := d.DetermineType(append(header, make([]byte, 24)...)); !complete || typ != TypeOneNote {
		t.Fatalf("OneNote section with hash-only database: %s %v", typ, complete)
	}
	header[0] ^= 1
	if typ, _ := d.DetermineType(append(header, make([]byte, 24)...)); typ == TypeOneNote {
		t.Fatal("wrong OneNote header GUID was classified as a section")
	}
}

func TestFileTypeMagicLinesAreTriedLastLineFirst(t *testing.T) {
	d := compileFTM(t, "0:0:4142:first:CL_TYPE_ANY:CL_TYPE_GIF\n0:0:4142:second:CL_TYPE_ANY:CL_TYPE_PNG\n")
	if typ, ok := d.compareMagic([]byte("ABCDEF")); !ok || typ != TypePNG {
		t.Fatalf("compareMagic = %s %v, want PNG from the later line", typ, ok)
	}
}

func TestFileTypeMagicNeedsItsBytesInTheBuffer(t *testing.T) {
	d := compileFTM(t, "0:4:4142:at four:CL_TYPE_ANY:CL_TYPE_GIF\n")
	if typ, ok := d.compareMagic([]byte("xxxxAB")); !ok || typ != TypeGIF {
		t.Fatalf("offset 4: %s %v", typ, ok)
	}
	if _, ok := d.compareMagic([]byte("xxxxA")); ok {
		t.Fatal("a magic running past the buffer matched")
	}
}

func TestFileTypeRequiredTypeIsIgnoredForMagicLines(t *testing.T) {
	d := compileFTM(t, "0:0:4142:x:CL_TYPE_PDF:CL_TYPE_GIF\n")
	if typ, ok := d.compareMagic([]byte("ABCDEF")); !ok || typ != TypeGIF {
		t.Fatalf("compareMagic = %s %v, want GIF", typ, ok)
	}
}

func TestFileTypePatternLinesCompileWithFullSyntax(t *testing.T) {
	d := compileFTM(t, "0:0:4d5a:exe:CL_TYPE_ANY:CL_TYPE_MSEXE\n1:0,1024:4d5a{60-300}50450000:PE in text:CL_TYPE_ANY:CL_TYPE_MSEXE\n1:*:3c68746d6c3e:html:CL_TYPE_TEXT_ASCII:CL_TYPE_HTML\n")
	p := d.ftypes.patterns
	if len(p) != 2 || p[0].typ != TypeMSEXE || p[1].typ != TypeHTML || p[1].rtype != TypeTextASCII {
		t.Fatalf("%d patterns loaded, want the two lines in file order", len(p))
	}
	if got := d.Report().FileTypes; got != 3 {
		t.Fatalf("Report().FileTypes = %d, want 3", got)
	}
}

func TestFileTypeLineSelectionAndErrors(t *testing.T) {
	for _, line := range []string{
		"0:0:4d5a:x:CL_TYPE_ANY:CL_TYPE_MSEXE:235",
		"0:0:4d5a:x:CL_TYPE_ANY:CL_TYPE_MSEXE:1:215",
		"0:0:4d5a:x:CL_TYPE_ANY:CL_TYPE_BOGUS:235",
		"3:0:4d5a:unsupported magic type:CL_TYPE_ANY:CL_TYPE_MSEXE",
	} {
		d := compileFTM(t, line+"\n")
		if n := d.Report().FileTypes; n != 0 {
			t.Errorf("%q loaded %d entries", line, n)
		}
	}
	for _, line := range []string{
		"0:0:4d5a:x:CL_TYPE_ANY",
		"0:0:4d5a:x:CL_TYPE_ANY:CL_TYPE_BOGUS",
		"0:0:4d5a:x:CL_TYPE_ANY:CL_TYPE_MSEXE:abc",
		"0:x:4d5a:x:CL_TYPE_ANY:CL_TYPE_MSEXE",
		"0:0:4d5:x:CL_TYPE_ANY:CL_TYPE_MSEXE",
		"0:0:zz:x:CL_TYPE_ANY:CL_TYPE_MSEXE",
		"1:0:4d5a(:x:CL_TYPE_ANY:CL_TYPE_MSEXE",
	} {
		_, err := Compile(map[string][]byte{"test.ftm": []byte(line + "\n"), "test.ndb": []byte("M:0:*:41414141\n")})
		if err == nil {
			t.Errorf("%q compiled", line)
		}
	}
	if _, err := Compile(map[string][]byte{"test.ftm": []byte("# only a comment\n"), "test.ndb": []byte("M:0:*:41414141\n")}); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("an empty file-type database compiled: %v", err)
	}
}

func TestBuiltInFileTypesApplyWhenNoMagicLineLoads(t *testing.T) {
	d := compileTest(t, "ndb", "M:0:*:41414141\n")
	for _, c := range []struct {
		data string
		want FileType
	}{
		{"MZ\x90\x00", TypeMSEXE}, {"\x7fELF\x02\x01", TypeELF}, {"\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1", TypeMSOLE2},
		{"PK\x03\x04", TypeZIP}, {"%PDF-1.7", TypePDF}, {"\x89PNG\r\n\x1a\n", TypePNG},
		{"{\n  \"Magic\":\"CLAMJSONv0\"", TypeInternal},
	} {
		if typ, ok := d.compareMagic([]byte(c.data)); !ok || typ != c.want {
			t.Errorf("built-in: %q typed %s %v, want %s", c.data, typ, ok, c.want)
		}
	}
	d = compileFTM(t, "0:0:4142:x:CL_TYPE_ANY:CL_TYPE_GIF\n")
	if _, ok := d.compareMagic([]byte("MZ\x90\x00")); ok {
		t.Fatal("the built-in table was used although the generation loads magic lines")
	}
	// Like the reference, a file-type database that loads only type patterns
	// still gets the built-in magic table.
	d = compileFTM(t, "1:*:3c68746d6c3e:html:CL_TYPE_ANY:CL_TYPE_HTML\n")
	if typ, ok := d.compareMagic([]byte("MZ\x90\x00")); !ok || typ != TypeMSEXE {
		t.Fatalf("patterns only: MZ typed %s %v, want MSEXE from the built-in table", typ, ok)
	}
}

func TestBuiltInHTMLTypeMarkersFollowReference(t *testing.T) {
	d := compileTest(t, "ndb", "Harmless.Marker:0:*:414243444546\n")
	for _, tc := range []struct {
		input string
		want  FileType
	}{
		{"ordinary <header> text", TypeTextASCII},
		{"ordinary <html lang=en> text", TypeTextASCII},
		{"ordinary <head> text", TypeHTML},
		{"ordinary <HTML> text", TypeHTML},
		{"ordinary <img src=x> text", TypeHTML},
		{"ordinary <a HREF=x> text", TypeHTML},
		{"ordinary <A href=x> text", TypeHTML},
		{"ordinary <TABLE> text", TypeHTML},
		{"<%@ LANGUAGE = VBScript %>", TypeHTML},
	} {
		got, complete := d.DetermineType([]byte(tc.input))
		if !complete || got != tc.want {
			t.Errorf("%q typed %v complete=%v; want %v", tc.input, got, complete, tc.want)
		}
	}
}

func TestBuiltInCompiledPythonTypeMarker(t *testing.T) {
	d := compileTest(t, "ndb", "Harmless.Marker:0:*:414243444546\n")
	for _, tc := range []struct {
		name string
		data []byte
		want FileType
	}{
		{"modern pyc", []byte{0x2d, 0x0d, 0x0d, 0x0a, 0x80}, TypePythonComp},
		{"different pyc revision", []byte{0x3f, 0x0d, 0x0d, 0x0a, 0x80}, TypePythonComp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, complete := d.DetermineType(tc.data)
			if !complete || got != tc.want {
				t.Fatalf("type=%s complete=%v, want %s", got, complete, tc.want)
			}
		})
	}
	if got, complete := d.DetermineType([]byte{0x2d, 0x0d, 0x0d, 0x0b, 0x80}); !complete || got == TypePythonComp {
		t.Fatalf("near-miss header typed %s complete=%v", got, complete)
	}
}
