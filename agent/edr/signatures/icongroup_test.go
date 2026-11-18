package signatures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// peWithResources is a PE32 program with 16 data directories and a resource
// section holding one resource of the given type (14 is a group icon, 16 a
// version record), "ABCD" at the entry point. characteristics are the COFF
// flags (0x2000 marks a DLL).
func peWithResources(resourceType uint32, characteristics uint16) []byte {
	le := binary.LittleEndian
	data := make([]byte, 2048)
	copy(data, "MZ")
	le.PutUint32(data[0x3c:], 128)
	copy(data[128:], "PE\x00\x00")
	le.PutUint16(data[132:], 0x14c)
	le.PutUint16(data[134:], 2)
	le.PutUint16(data[148:], 224)
	le.PutUint16(data[150:], characteristics)
	le.PutUint16(data[152:], 0x10b)
	le.PutUint32(data[168:], 0x1000) // entry point
	le.PutUint32(data[184:], 0x1000) // section alignment
	le.PutUint32(data[188:], 0x200)  // file alignment
	le.PutUint32(data[212:], 0x200)  // size of headers
	le.PutUint32(data[244:], 16)     // data directories
	le.PutUint32(data[264:], 0x2000) // resource directory address
	le.PutUint32(data[268:], 0x100)  // and size
	section := func(at int, name string, rva, raw uint32) {
		copy(data[at:], name)
		le.PutUint32(data[at+8:], 0x200)
		le.PutUint32(data[at+12:], rva)
		le.PutUint32(data[at+16:], 0x200)
		le.PutUint32(data[at+20:], raw)
	}
	section(376, ".text", 0x1000, 0x200)
	section(416, ".rsrc", 0x2000, 0x400)
	copy(data[0x200:], "ABCD")
	// Resource tree: type -> name -> language -> data entry.
	dir := func(at int, ids uint16) { le.PutUint16(data[at+14:], ids) }
	entry := func(at int, id, offset uint32) { le.PutUint32(data[at:], id); le.PutUint32(data[at+4:], offset) }
	dir(0x400, 1)
	entry(0x410, resourceType, 0x80000000|0x18)
	dir(0x418, 1)
	entry(0x428, 1, 0x80000000|0x30)
	dir(0x430, 1)
	entry(0x440, 0x409, 0x48)
	le.PutUint32(data[0x448:], 0x2100)
	le.PutUint32(data[0x44c:], 0x20)
	if resourceType == 14 {
		// one icon, id 1, that the resources do not hold: a group this
		// engine cannot decide
		le.PutUint16(data[0x502:], 1)
		le.PutUint16(data[0x504:], 1)
		data[0x506], data[0x507] = 16, 16
		le.PutUint16(data[0x50a:], 1)
		le.PutUint16(data[0x50c:], 32)
		le.PutUint16(data[0x512:], 1)
	}
	return data
}

// iconMetrics is a valid 16-pixel icon hash.
const iconMetrics = "10" + "000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000000000000000000000"

func iconDatabase(t *testing.T, idb string, rules ...string) *Database {
	t.Helper()
	files := map[string][]byte{"a.ldb": []byte(strings.Join(rules, "\n") + "\n")}
	if idb != "" {
		files["a.idb"] = []byte(idb)
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestIconGroupIsUnknownOnlyWhenTheReferenceCouldCompareIcons(t *testing.T) {
	idb := "Win.Test.IconA:MS:IE:" + iconMetrics + "\nWin.Test.IconB:DOC:PDF:" + iconMetrics + "\n"
	d := iconDatabase(t, idb, "Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;41424344")
	exe := MatchContext{Type: TypeMSEXE}
	malformed := peWithResources(14, 0x0102)
	binary.LittleEndian.PutUint32(malformed[0x44c:], 4) // no complete group entry
	expectOutcome(t, d, string(malformed), exe, "", "")
	expectOutcome(t, d, string(peWithResources(14, 0x0102)), exe, "", "")
	// No resource address: a DLL, or fewer than three data directories.
	expectOutcome(t, d, string(peWithResources(14, 0x2102)), exe, "", "")
	expectOutcome(t, d, string(testPE()), exe, "", "")
	// Resources, but no group icon.
	expectOutcome(t, d, string(peWithResources(16, 0x0102)), exe, "", "")
}

// ClamAV 1.5.4 scans complete icon-group entries and ignores a short tail;
// the adjacent audit records its verdict and parser diagnostics for each.
func TestIconGroupShortTailReferenceModes(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-28-v128-icon-group-tail", "oracle")
	d := iconDatabase(t, "Win.Test.IconA:MS:IE:"+iconMetrics+"\n",
		"Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;41424344")
	for _, name := range []string{"short-group", "partial-first", "partial-second", "declared-many", "complete-first"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, name+".exe"))
			if err != nil {
				t.Fatal(err)
			}
			expectOutcome(t, d, string(data), MatchContext{Type: TypeMSEXE}, "", "")
		})
	}
}

func TestIconGroupCompleteFirstEntryDetectsDespiteShortSecond(t *testing.T) {
	icon := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	pe := labProgram([]labIcon{icon})
	meta := parsePE(pe)
	if meta == nil {
		t.Fatal("PE resources missing")
	}
	var group []byte
	if !meta.resources(pe, 14, func(entry uint32) bool {
		group, _ = meta.resourceData(pe, entry)
		return true
	}) || len(group) != 20 {
		t.Fatalf("group size %d", len(group))
	}
	binary.LittleEndian.PutUint16(group[4:], 2) // second 14-byte entry is absent
	const flatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	d := iconDatabase(t, "Win.Test.IconA:MS:IE:"+flatHash+"\n",
		"Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Live", "")
	if dir := os.Getenv("EDR_ICON_TAIL_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "partial-second-live.exe"), pe, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGroupIconResourceCountWrapsToZero(t *testing.T) {
	base := peWithResources(14, 0x0102)
	for _, tc := range []struct {
		name string
		at   int
	}{
		{"name", 0x424},
		{"language", 0x43c},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pe := append([]byte(nil), base...)
			binary.LittleEndian.PutUint16(pe[tc.at:], 0xffff)
			meta := parsePE(pe)
			if meta == nil || !meta.valid || meta.hasGroupIcon(pe) {
				t.Fatal("overflowing resource entry counts exposed a group icon")
			}
		})
	}
}

func TestIconGroupNamesNoIconEntryCarriesNeverMatch(t *testing.T) {
	idb := "Win.Test.IconA:MS:IE:" + iconMetrics + "\nWin.Test.IconB:DOC:PDF:" + iconMetrics + "\n"
	pe := string(peWithResources(14, 0x0102))
	exe := MatchContext{Type: TypeMSEXE}
	for _, rule := range []string{
		"Icon.Dead1;Engine:51-255,Target:1,IconGroup1:BROWSER;0;41424344",           // no such group-1 name
		"Icon.Dead2;Engine:51-255,Target:1,IconGroup2:MS;0;41424344",                // MS is only a group-1 name
		"Icon.Dead3;Engine:51-255,Target:1,IconGroup1:MS,IconGroup2:PDF;0;41424344", // no entry has both
	} {
		d := iconDatabase(t, idb, rule)
		expectOutcome(t, d, pe, exe, "", "")
		if d.Report().Unknown[UnknownIconGroup] != 0 {
			t.Errorf("%s counted as unknown when reached: %v", rule, d.Report().Unknown)
		}
	}
	live := iconDatabase(t, idb, "Icon.Both;Engine:51-255,Target:1,IconGroup1:DOC,IconGroup2:PDF;0;41424344")
	expectOutcome(t, live, pe, exe, "", "")
	if live.Report().Unknown[UnknownIconGroup] != 0 {
		t.Errorf("implemented icon predicate counted as missing analysis: %v", live.Report().Unknown)
	}
	// Without any icon database no icon rule can match.
	none := iconDatabase(t, "", "Icon.NoDatabase;Engine:51-255,Target:1,IconGroup1:MS;0;41424344", "Other;Engine:51-255,Target:0;0;57575757")
	expectOutcome(t, none, pe, exe, "", "")
}

func TestExecutableWithoutPESignatureIsAFailedHeader(t *testing.T) {
	// A 16-bit program or font: MZ, then an NE header where PE would be.
	ne := make([]byte, 512)
	copy(ne, "MZ")
	binary.LittleEndian.PutUint32(ne[0x3c:], 128)
	copy(ne[128:], "NE ABCD")
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("Ep:1:EP+0:41424344\n"),
		"a.mdb": []byte("512:0123456789abcdef0123456789abcdef:Section.Hash\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	expectOutcome(t, d, string(ne), MatchContext{Type: TypeMSEXE}, "", "")
	short := []byte("MZ with too little header")
	expectOutcome(t, d, string(short), MatchContext{Type: TypeMSEXE}, "", "")
}
