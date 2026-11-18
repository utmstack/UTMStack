package signatures

import (
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The first group-language entry is unreadable, while the second contains
// a valid group referring to the icon. The reference keeps visiting the
// first group ID's languages after an unreadable data entry.
func peWithSecondGroupLanguage() []byte {
	const sectionRaw = 0x200
	const sectionHeader = 0x138
	le := binary.LittleEndian
	flat := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	pe := labProgram([]labIcon{flat})
	meta := parsePE(pe)
	var firstEntry uint32
	meta.resources(pe, 14, func(entry uint32) bool { firstEntry = entry; return true })
	group, _ := meta.resourceData(pe, firstEntry)
	align := func(n, by int) int { return (n + by - 1) &^ (by - 1) }
	secondAt := len(pe)
	pe = append(pe, group...)
	pe = append(pe, make([]byte, align(len(pe), 4)-len(pe))...)
	secondEntryAt := len(pe)
	pe = append(pe, make([]byte, 16)...)
	langsAt := len(pe)
	pe = append(pe, make([]byte, 32)...)
	// Replace the group name's single-language directory with two entries.
	le.PutUint32(pe[sectionRaw+0x38+20:], 0x80000000|uint32(langsAt-sectionRaw))
	le.PutUint16(pe[langsAt+14:], 2)
	le.PutUint32(pe[langsAt+16:], 0x409)
	le.PutUint32(pe[langsAt+20:], firstEntry-meta.resDir)
	le.PutUint32(pe[langsAt+24:], 0x40a)
	le.PutUint32(pe[langsAt+28:], uint32(secondEntryAt-sectionRaw))
	le.PutUint32(pe[secondEntryAt:], uint32(0x1000+secondAt-sectionRaw))
	le.PutUint32(pe[secondEntryAt+4:], uint32(len(group)))
	// The first group-data RVA is unmapped. The second remains valid.
	firstAt, _ := meta.rawAddr(firstEntry)
	le.PutUint32(pe[firstAt:], 0xffffffff)
	virtualSize := len(pe) - sectionRaw
	le.PutUint32(pe[sectionHeader+8:], uint32(virtualSize))
	le.PutUint32(pe[sectionHeader+16:], uint32(align(virtualSize, 0x200)))
	le.PutUint32(pe[0x58+56:], uint32(0x1000+align(virtualSize, 0x1000)))
	le.PutUint32(pe[0x58+116:], uint32(virtualSize))
	return pe
}

func peWithSecondGroupName() []byte {
	const sectionRaw = 0x200
	const sectionHeader = 0x138
	le := binary.LittleEndian
	pe := peWithSecondGroupLanguage()
	secondLangAt := len(pe) - 32
	newNameDirAt := len(pe)
	pe = append(pe, make([]byte, 32)...)
	// The first name gets the original, unreadable group language. The
	// second name gets the two-language directory, including the valid one.
	le.PutUint16(pe[newNameDirAt+14:], 2)
	le.PutUint32(pe[newNameDirAt+16:], 1)
	le.PutUint32(pe[newNameDirAt+20:], 0x80000068)
	le.PutUint32(pe[newNameDirAt+24:], 2)
	le.PutUint32(pe[newNameDirAt+28:], 0x80000000|uint32(secondLangAt-sectionRaw))
	le.PutUint32(pe[sectionRaw+28:], 0x80000000|uint32(newNameDirAt-sectionRaw))
	virtualSize := len(pe) - sectionRaw
	align := func(n, by int) int { return (n + by - 1) &^ (by - 1) }
	le.PutUint32(pe[sectionHeader+8:], uint32(virtualSize))
	le.PutUint32(pe[sectionHeader+16:], uint32(align(virtualSize, 0x200)))
	le.PutUint32(pe[0x58+56:], uint32(0x1000+align(virtualSize, 0x1000)))
	le.PutUint32(pe[0x58+116:], uint32(virtualSize))
	return pe
}

func peWithFullFirstGroupBudget() []byte {
	const sectionRaw = 0x200
	const sectionHeader = 0x138
	le := binary.LittleEndian
	pe := peWithSecondGroupLanguage()
	meta := parsePE(pe)
	var firstEntry uint32
	meta.resources(pe, 14, func(entry uint32) bool { firstEntry = entry; return true })
	firstAt, _ := meta.rawAddr(firstEntry)
	group := make([]byte, 6+14*maxGroupIcons)
	le.PutUint16(group[2:], 1)
	le.PutUint16(group[4:], maxGroupIcons)
	for i := 0; i < maxGroupIcons; i++ {
		le.PutUint16(group[6+14*i+12:], 2) // no such icon resource
	}
	groupAt := len(pe)
	pe = append(pe, group...)
	le.PutUint32(pe[firstAt:], uint32(0x1000+groupAt-sectionRaw))
	le.PutUint32(pe[firstAt+4:], uint32(len(group)))
	virtualSize := len(pe) - sectionRaw
	align := func(n, by int) int { return (n + by - 1) &^ (by - 1) }
	le.PutUint32(pe[sectionHeader+8:], uint32(virtualSize))
	le.PutUint32(pe[sectionHeader+16:], uint32(align(virtualSize, 0x200)))
	le.PutUint32(pe[0x58+56:], uint32(0x1000+align(virtualSize, 0x1000)))
	le.PutUint32(pe[0x58+116:], uint32(virtualSize))
	return pe
}

func TestSecondGroupLanguageDetectsAfterUnreadableFirst(t *testing.T) {
	const flatHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	idb := "Probe.Group:A:B:" + flatHex + "\n"
	rule := "Icon.SecondGroup;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
	pe := peWithSecondGroupLanguage()
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": pe} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		object := filepath.Join(dir, "object.exe")
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", object)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(out), object+": Icon.SecondGroup.UNOFFICIAL FOUND") ||
			!strings.Contains(string(out), "icons across 2 groups") {
			t.Fatalf("reference second-group scan: error=%v output=%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "groupicon_cb: scanning group") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": ") {
				t.Logf("ClamAV 1.5.4: %s", line)
			}
		}
	}
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.SecondGroup", "")
}

func TestSecondGroupNameIsNotScanned(t *testing.T) {
	const flatHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	idb := "Probe.Group:A:B:" + flatHex + "\n"
	rule := "Icon.SecondGroup;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
	pe := peWithSecondGroupName()
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": pe} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		object := filepath.Join(dir, "object.exe")
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", object)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), object+": OK") ||
			!strings.Contains(string(out), "icons across 1 groups") {
			t.Fatalf("reference second-name scan: error=%v output=%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "groupicon_cb: scanning group") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": ") {
				t.Logf("ClamAV 1.5.4: %s", line)
			}
		}
	}
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
}

func TestIconEntryBudgetStopsBeforeSecondLanguage(t *testing.T) {
	const flatHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	idb := "Probe.Group:A:B:" + flatHex + "\n"
	rule := "Icon.SecondGroup;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
	pe := peWithFullFirstGroupBudget()
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": pe} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		object := filepath.Join(dir, "object.exe")
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", object)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), object+": OK") ||
			!strings.Contains(string(out), "max icon count reached") ||
			!strings.Contains(string(out), "icons across 1 groups") {
			t.Fatalf("reference shared-budget scan: error=%v output=%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "max icon count reached") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": ") {
				t.Logf("ClamAV 1.5.4: %s", line)
			}
		}
	}
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
}

func TestSignedIconIDFFFFSelectsAnyName(t *testing.T) {
	const flatHex = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	idb := "Probe.Wildcard:A:B:" + flatHex + "\n"
	rule := "Icon.Wildcard;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
	flat := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	pe := labProgram([]labIcon{flat})
	meta := parsePE(pe)
	var group []byte
	meta.resources(pe, 14, func(entry uint32) bool { group, _ = meta.resourceData(pe, entry); return true })
	binary.LittleEndian.PutUint16(group[18:], 0xffff)
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": pe} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		object := filepath.Join(dir, "object.exe")
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", object)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(out), object+": Icon.Wildcard.UNOFFICIAL FOUND") {
			t.Fatalf("reference signed-ID scan: error=%v output=%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "Icongrp @") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": ") {
				t.Logf("ClamAV 1.5.4: %s", line)
			}
		}
	}
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Wildcard", "")
}
