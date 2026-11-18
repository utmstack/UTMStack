package signatures

import (
	"context"
	"crypto/md5"
	"fmt"
	"testing"
)

func matchObject(t *testing.T, d *Database, data string, mc MatchContext) (string, error) {
	t.Helper()
	res, err := d.MatchObject(context.Background(), []byte(data), mc)
	return res.Name, err
}

func TestTypedRootsApplyOnlyToTheirType(t *testing.T) {
	d := compileTest(t, "ndb", "Elf.Rule:6:*:4d41524b4552\n")
	elf := "\x7fELF\x02\x01\x01 then MARKER"
	if name, err := matchObject(t, d, elf, MatchContext{Type: TypeELF}); name != "Elf.Rule" || err != nil {
		t.Fatalf("typed ELF: %q, %v", name, err)
	}
	for _, typ := range []FileType{TypeAny, TypeMSEXE, TypeTextASCII, TypeBinaryData} {
		if name, err := matchObject(t, d, elf, MatchContext{Type: typ}); name != "" || err != nil {
			t.Fatalf("typed %s: %q, %v; a target-6 rule ran outside ELF objects", typ, name, err)
		}
	}
}

func TestTypedRootRunsBeforeRootZero(t *testing.T) {
	d := compileTest(t, "ndb", "Generic:0:*:4d41524b4552\nTyped:6:*:4d41524b4552\n")
	if name, err := matchObject(t, d, "\x7fELF MARKER", MatchContext{Type: TypeELF}); name != "Typed" || err != nil {
		t.Fatalf("got %q, %v; want the typed root's rule first", name, err)
	}
	if name, err := matchObject(t, d, "\x7fELF MARKER", MatchContext{Type: TypeAny}); name != "Generic" || err != nil {
		t.Fatalf("root 0 only: got %q, %v", name, err)
	}
}

func TestExecutableTypedWithoutHeaderHasNoRelativeOffsets(t *testing.T) {
	// Typed MSEXE by a type pattern, with MZ not at offset 0: both engines
	// fail the header, so relative offsets never match, in root 1 and root 0.
	data := "0123456789MZ" + "......................" + "MZ"
	d := compileTest(t, "ndb", "Absolute:1:10:4d5a\n")
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeMSEXE}); name != "Absolute" || err != nil {
		t.Fatalf("absolute offset: %q, %v", name, err)
	}
	for _, rule := range []string{"Entry:1:EP+0:4d5a", "Section:1:S0+0:4d5a", "End:0:EOF-2:4d5a"} {
		d := compileTest(t, "ndb", rule+"\n")
		if name, err := matchObject(t, d, data, MatchContext{Type: TypeMSEXE}); name != "" || err != nil {
			t.Errorf("%s: %q, %v; want no match and no unknown", rule, name, err)
		}
	}
	d = compileTest(t, "ndb", "End:0:EOF-2:4d5a\n")
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeTextASCII}); name != "End" || err != nil {
		t.Fatalf("EOF offset on a text object: %q, %v", name, err)
	}
}

func TestBufferModeHasNoRelativeOffsetsOrHashes(t *testing.T) {
	data := "xxABCD"
	d := compileTest(t, "ndb", "End:0:EOF-4:41424344\nAbsolute:0:2:41424344\n")
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeAny}); name != "End" || err != nil {
		t.Fatalf("file mode: %q, %v", name, err)
	}
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeAny, Mode: BufferMode}); name != "Absolute" || err != nil {
		t.Fatalf("buffer mode: %q, %v; want only the absolute rule", name, err)
	}
	h := md5.Sum([]byte(data))
	d = compileTest(t, "hdb", fmt.Sprintf("%x:%d:Hash.Rule\n", h, len(data)))
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeAny, Mode: BufferMode}); name != "" || err != nil {
		t.Fatalf("hash rule in buffer mode: %q, %v", name, err)
	}
	if name, err := matchObject(t, d, data, MatchContext{Type: TypeAny}); name != "Hash.Rule" || err != nil {
		t.Fatalf("hash rule in file mode: %q, %v", name, err)
	}
	d = compileTest(t, "ldb", "Sections;Engine:51-255,Target:1,NumberOfSections:0-10;0;41424344\n")
	if name, err := matchObject(t, d, string(testPE())+data, MatchContext{Type: TypeMSEXE, Mode: BufferMode}); name != "" || err != nil {
		t.Fatalf("executable descriptor in buffer mode: %q, %v; want false, not unknown", name, err)
	}
}

func TestBodyRulePrecedesWholeFileHash(t *testing.T) {
	// ClamAV 1.5.4 reports Body.Marker when an .ndb and .hdb both
	// match this harmless file, independent of the database file order.
	data := []byte("PRE-harmless-marker-POST")
	h := md5.Sum(data)
	files := map[string][]byte{
		"a.hdb": []byte(fmt.Sprintf("%x:%d:Hash.Marker\n", h, len(data))),
		"z.ndb": []byte(fmt.Sprintf("Body.Marker:0:*:%x\n", data)),
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if name, err := matchObject(t, d, string(data), MatchContext{Type: TypeAny}); name != "Body.Marker" || err != nil {
		t.Fatalf("body and hash: %q, %v", name, err)
	}
	files["z.ndb"] = []byte("Other:0:*:4e4f4d41544348\n")
	d, err = Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if name, err := matchObject(t, d, string(data), MatchContext{Type: TypeAny}); name != "Hash.Marker" || err != nil {
		t.Fatalf("hash fallback: %q, %v", name, err)
	}
	files["allowed.fp"] = []byte(fmt.Sprintf("%x:%d:Allowed\n", h, len(data)))
	d, err = Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if name, err := matchObject(t, d, string(data), MatchContext{Type: TypeAny}); name != "" || err != nil {
		t.Fatalf("allowlist: %q, %v", name, err)
	}
}

func TestUnmappedPEEntryPointRejectsHeaderOffsets(t *testing.T) {
	d := compileTest(t, "ndb", "Entry:1:EP+0:57575757\n")
	if name, err := matchObject(t, d, string(brokenPE()), MatchContext{Type: TypeMSEXE}); name != "" || err != nil {
		t.Fatalf("absent pattern: %q, %v; want clean", name, err)
	}
	name, err := matchObject(t, d, string(brokenPE())+"WWWW", MatchContext{Type: TypeMSEXE})
	if name != "" || err != nil {
		t.Fatalf("pattern present: %q, %v; want clean", name, err)
	}
	d = compileTest(t, "ndb", "Entry:1:EP+0:5757??57\n")
	if name, err := matchObject(t, d, string(brokenPE())+"WWxZWW", MatchContext{Type: TypeMSEXE}); name != "" || err != nil {
		t.Fatalf("anchor present but no match anywhere: %q, %v; want clean", name, err)
	}
}

func TestMatchTypesTheObjectForItsRawScan(t *testing.T) {
	d := compileTest(t, "ndb", "Pe.Rule:1:*:41424344\nText.Rule:7:*:6d61726b6572\n")
	assertMatch(t, d, testPE(), "Pe.Rule")
	assertMatch(t, d, []byte("text holding ABCD and marker"), "")
}
