package signatures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// labTwoLanguages puts a second image under the same icon ID in another
// language. The group still has only one icon entry, so resource traversal
// must visit both language data entries to find the second image.
func labTwoLanguages(first, second labIcon) []byte {
	const sectionRaw = 0x200
	const sectionHeader = 0x138
	le := binary.LittleEndian
	pe := labProgram([]labIcon{first})
	align := func(n, by int) int { return (n + by - 1) &^ (by - 1) }
	imageAt := len(pe)
	pe = append(pe, second.bitmap()...)
	pe = append(pe, make([]byte, align(len(pe), 4)-len(pe))...)
	dataEntryAt := len(pe)
	pe = append(pe, make([]byte, 16)...)
	langDirAt := len(pe)
	pe = append(pe, make([]byte, 32)...)
	// The original type-3, id-1 name entry points to its language directory.
	le.PutUint32(pe[sectionRaw+32+20:], 0x80000000|uint32(langDirAt-sectionRaw))
	le.PutUint16(pe[langDirAt+14:], 2)
	le.PutUint32(pe[langDirAt+16:], 0x409)
	le.PutUint32(pe[langDirAt+20:], 128)
	le.PutUint32(pe[langDirAt+24:], 0x40a)
	le.PutUint32(pe[langDirAt+28:], uint32(dataEntryAt-sectionRaw))
	le.PutUint32(pe[dataEntryAt:], uint32(0x1000+imageAt-sectionRaw))
	le.PutUint32(pe[dataEntryAt+4:], uint32(len(second.bitmap())))
	virtualSize := len(pe) - sectionRaw
	le.PutUint32(pe[sectionHeader+8:], uint32(virtualSize))
	le.PutUint32(pe[sectionHeader+16:], uint32(align(virtualSize, 0x200)))
	le.PutUint32(pe[0x58+56:], uint32(0x1000+align(virtualSize, 0x1000)))
	le.PutUint32(pe[0x58+116:], uint32(virtualSize))
	return pe
}

func TestIconResourcesVisitBothLanguages(t *testing.T) {
	first := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	second := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := y*16 + x
			second.rgba[i] = [4]byte{byte(x * 17), byte(y * 15), byte((x*13 + y*11) & 255), 255}
		}
	}
	pe := labTwoLanguages(first, second)
	got := labIcons(t, pe)
	if len(got) != 2 {
		t.Fatalf("want both language icons, got %d", len(got))
	}
	for i, ic := range []labIcon{first, second} {
		want, skip, sure := parseIconImage(ic.bitmap())
		if !sure || skip || got[i] != want {
			t.Fatalf("language %d icon mismatch: got=%+v want=%+v", i, got[i], want)
		}
	}
	if got[0] == got[1] {
		t.Fatal("the two language images should have distinct hashes")
	}
	// The pinned ClamAV 1.5.4 debug trace prints both hashes in this order,
	// and its IDB/LDB rule for only the second language detects the PE.
	const firstHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	const secondHex = "10c480a00bbd000bb830a0b3f301014ce04055e70609e60a07e0060bcf0a0b27000057000460040069070b4008073b06031400001700051b090022231f57"
	for i, raw := range []string{firstHex, secondHex} {
		reference, err := decodeIconHash(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got[i].hash != reference {
			t.Fatalf("language %d hash differs from ClamAV: got=%+v want=%+v", i, got[i].hash, reference)
		}
	}
	d := iconDatabase(t, "Probe.Language2:A:B:"+secondHex+"\n",
		"Icon.Language2;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Language2", "")
	if dir := os.Getenv("EDR_ICON_LANG_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "two-icon-languages.exe"), pe, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIconEarlierLanguageMatchSurvivesLaterUnreadableImage(t *testing.T) {
	first := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	second := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	pe := labTwoLanguages(first, second)
	// The second language's data entry points outside the file. The first
	// language is still a complete icon and can establish a detection.
	const sectionRaw = 0x200
	langDirAt := len(pe) - 32
	secondEntryAt := sectionRaw + int(binary.LittleEndian.Uint32(pe[langDirAt+28:]))
	binary.LittleEndian.PutUint32(pe[secondEntryAt:], 0xffffffff)
	const firstHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	d := iconDatabase(t, "Probe.Language1:A:B:"+firstHex+"\n",
		"Icon.Language1;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Language1", "")
	const otherHex = "10c480a00bbd000bb830a0b3f301014ce04055e70609e60a07e0060bcf0a0b27000057000460040069070b4008073b06031400001700051b090022231f57"
	other := iconDatabase(t, "Probe.Other:A:B:"+otherHex+"\n",
		"Icon.Other;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, other, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
	if dir := os.Getenv("EDR_ICON_LANG_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "second-language-unreadable.exe"), pe, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
