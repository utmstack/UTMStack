package signatures

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"
	"unicode/utf16"
)

func utf16z(s string) []byte {
	var b []byte
	for _, u := range utf16.Encode([]rune(s)) {
		b = binary.LittleEndian.AppendUint16(b, u)
	}
	return append(b, 0, 0)
}

func pad4(b []byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}

// versionBlock is one version-resource node: a 6-byte header (length,
// value length, type), a key, then a value and children, each 32-bit aligned.
func versionBlock(key string, valueLength uint16, value []byte, children ...[]byte) []byte {
	b := pad4(append(make([]byte, 6), utf16z(key)...))
	b = append(b, value...)
	for _, c := range children {
		b = append(pad4(b), c...)
	}
	binary.LittleEndian.PutUint16(b, uint16(len(b)))
	binary.LittleEndian.PutUint16(b[2:], valueLength)
	return b
}

func versionString(key, value string) []byte {
	return versionBlock(key, uint16(len(utf16z(value))/2), utf16z(value))
}

func versionResource(strings ...[]byte) []byte {
	fixed := make([]byte, 0x34)
	binary.LittleEndian.PutUint32(fixed, 0xfeef04bd)
	table := versionBlock("040904b0", 0, nil, strings...)
	return versionBlock("VS_VERSION_INFO", 0x34, fixed, versionBlock("StringFileInfo", 0, nil, table))
}

// peWithVersion is peWithResources with the version resource as its data.
func peWithVersion(resource []byte) []byte {
	pe := peWithResources(16, 0x0102)
	copy(pe[0x500:], resource)
	binary.LittleEndian.PutUint32(pe[0x44c:], uint32(len(resource)))
	return pe
}

func TestVersionInfoOffsetsAreTheVersionStringKeys(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	value := hex.EncodeToString(utf16z("Evil Corp")[:18])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\nVi.Value;Engine:51-255,Target:1;0;VI:"+value+"\n")
	exe := MatchContext{Type: TypeMSEXE}
	pe := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp"), versionString("ProductName", "Tool")))
	expectOutcome(t, d, string(pe), exe, "Vi.Key", "")
	// Only keys start version strings: the value alone does not match.
	d = compileTest(t, "ldb", "Vi.Value;Engine:51-255,Target:1;0;VI:"+value+"\n")
	expectOutcome(t, d, string(pe), exe, "", "")
	// A string without a value is not recorded.
	d = compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	expectOutcome(t, d, string(peWithVersion(versionResource(versionBlock("CompanyName", 0, nil)))), exe, "", "")
	// Without a version resource the pattern elsewhere decides nothing.
	other := peWithResources(14, 0x0102)
	copy(other[0x300:], utf16z("CompanyName"))
	expectOutcome(t, d, string(other), exe, "", "")
}

func TestVersionInfoZeroResourceDirectorySizeDoesNotMatch(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	pe := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
	binary.LittleEndian.PutUint32(pe[268:], 0)
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
}

func TestVersionInfoResourceCountWrapsLikeReference(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	base := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
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
			// 0xffff named entries plus the one ID entry wraps to zero.
			expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
		})
	}
}

func TestVersionInfoTruncatedSectionTableIsKnownFalse(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	pe := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
	// The COFF header declares two sections; fewer than two complete
	// section records are present, but the signature bytes still occur.
	broken := append(append([]byte(nil), pe[:420]...), utf16z("CompanyName")...)
	expectOutcome(t, d, string(broken), MatchContext{Type: TypeMSEXE}, "", "")
}

func TestVersionInfoSignedSectionFieldsAreKnownFalse(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	base := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
	for _, at := range []int{376 + 12, 416 + 12} {
		pe := append([]byte(nil), base...)
		binary.LittleEndian.PutUint32(pe[at:], 0xffffffff)
		expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE, PEAlertBroken: true}, "Heuristics.Broken.Executable", "")
		expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
	}
}

func TestVersionInfoSupportsFullCOFFSectionCount(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	base := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
	for _, count := range []int{4097, 8192, 65535} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			// Place the original code and resource sections after the enlarged
			// section table; all additional section records are empty.
			raw := (376 + count*40 + 511) &^ 511
			pe := make([]byte, raw+0x400)
			copy(pe[:456], base[:456])
			copy(pe[raw:], base[0x200:0x400])
			copy(pe[raw+0x200:], base[0x400:0x600])
			binary.LittleEndian.PutUint16(pe[134:], uint16(count))
			binary.LittleEndian.PutUint32(pe[376+20:], uint32(raw))
			binary.LittleEndian.PutUint32(pe[416+20:], uint32(raw+0x200))
			expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Vi.Key", "")
		})
	}
}

func TestVersionInfoOffsetsFollowTheReferenceOnDamagedResources(t *testing.T) {
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	exe := MatchContext{Type: TypeMSEXE}
	d := compileTest(t, "ldb", "Vi.Key;Engine:51-255,Target:1;0;VI:"+key+"\n")
	resource := versionResource(versionString("CompanyName", "Evil Corp"))
	binary.LittleEndian.PutUint16(resource[0x5c:], 0xffff) // a StringFileInfo longer than its parent: nothing read
	expectOutcome(t, d, string(peWithVersion(resource)), exe, "", "")
	// A last string whose rounded length passes its table stops the reading
	// there: the keys before it are recorded, its own is not.
	pe := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp"), versionString("ProductName", "Tool")))
	expectOutcome(t, d, string(pe), exe, "Vi.Key", "")
	last := hex.EncodeToString(utf16z("ProductName")[:22])
	d = compileTest(t, "ldb", "Vi.Last;Engine:51-255,Target:1;0;VI:"+last+"\n")
	expectOutcome(t, d, string(pe), exe, "", "")
}
