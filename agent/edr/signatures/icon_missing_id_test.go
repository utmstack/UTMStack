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

// A group entry that names no type-3 icon resource contributes no icon to
// ClamAV's comparison. It must not make an otherwise nonmatching icon rule
// incomplete.
func TestMissingIconResourceIsClean(t *testing.T) {
	idb := "Win.Test.IconA:MS:IE:" + iconMetrics + "\n"
	rule := "Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;41424344\n"
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	for _, tc := range []struct {
		name string
		pe   []byte
	}{
		{"type-3-absent", peWithResources(14, 0x0102)},
		{"icon-id-absent", peWithMissingIconID()},
		{"icon-type-not-directory", peWithSkippedIconType()},
		{"icon-name-not-directory", peWithSkippedIconResource(false)},
		{"icon-language-is-directory", peWithSkippedIconResource(true)},
		{"icon-type-directory-unmapped", peWithMalformedIconDirectory()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if ref != "" {
				dir := t.TempDir()
				for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": tc.pe} {
					if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				object := filepath.Join(dir, "object.exe")
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", object)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
					"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				if err != nil || !strings.Contains(string(out), object+": OK") {
					t.Fatalf("reference missing-icon scan: error=%v output=%s", err, out)
				}
				for _, line := range strings.Split(string(out), "\n") {
					if strings.Contains(line, "invalid icon entry") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": OK") {
						t.Log(line)
					}
				}
			}
			checkMissingIconResult(t, d, tc.pe)
		})
	}
}

func peWithSkippedIconType() []byte {
	pe := peWithMissingIconID()
	binary.LittleEndian.PutUint32(pe[0x41c:], 0x200)
	return pe
}

// findres ignores a matching name that does not point to a directory, and
// ignores a language entry that points to another directory.
func peWithSkippedIconResource(nestedLanguage bool) []byte {
	pe := peWithMissingIconID()
	le := binary.LittleEndian
	le.PutUint32(pe[0x610:], 1) // match the icon ID requested by the group
	if nestedLanguage {
		le.PutUint32(pe[0x62c:], 0x80000230)
	} else {
		le.PutUint32(pe[0x614:], 0x230)
	}
	return pe
}

func TestDuplicateIconIDVisitsBothResources(t *testing.T) {
	first := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	second := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := y*16 + x
			second.rgba[i] = [4]byte{byte(x * 17), byte(y * 15), byte((x*13 + y*11) & 255), 255}
		}
	}
	pe := labProgram([]labIcon{first, second})
	// The second type-3 name entry has the same ID as the first. The group
	// requests ID 1, so only a resource walk across duplicates sees image 2.
	binary.LittleEndian.PutUint32(pe[0x238:], 1)
	const secondHex = "10c480a00bbd000bb830a0b3f301014ce04055e70609e60a07e0060bcf0a0b27000057000460040069070b4008073b06031400001700051b090022231f57"
	idb := "Probe.Duplicate:A:B:" + secondHex + "\n"
	rule := "Icon.Duplicate;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
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
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(out), object+": Icon.Duplicate.UNOFFICIAL FOUND") {
			t.Fatalf("reference duplicate-ID scan: error=%v output=%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "parseicon: Bitmap") || strings.Contains(line, "scanned a total") || strings.Contains(line, object+": ") {
				t.Logf("ClamAV 1.5.4: %s", line)
			}
		}
	}
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Duplicate", "")
}

func peWithMalformedIconDirectory() []byte {
	pe := peWithMissingIconID()
	binary.LittleEndian.PutUint32(pe[0x41c:], 0xffffffff)
	return pe
}

func TestMalformedIconResourceDirectoryEndsLookup(t *testing.T) {
	pe := peWithMalformedIconDirectory()
	d := iconDatabase(t, "Win.Test.IconA:MS:IE:"+iconMetrics+"\n",
		"Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;41424344")
	// findres stops at this failed directory read; it does not invent an icon.
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
}

// Keep a valid type-3 resource subtree, but give it id 2 while the group
// requests id 1. The type-14 tree is moved within the expanded .rsrc section
// to make room for a second root entry.
func peWithMissingIconID() []byte {
	pe := peWithResources(14, 0x0102)
	le := binary.LittleEndian
	copy(pe[0x480:0x4c0], pe[0x418:0x458])
	le.PutUint16(pe[0x40e:], 2)
	le.PutUint32(pe[0x414:], 0x80000080)
	le.PutUint32(pe[0x490+4:], 0x80000098)
	le.PutUint32(pe[0x4a8+4:], 0xb0)
	le.PutUint32(pe[0x418:], 3)
	le.PutUint32(pe[0x41c:], 0x80000200)
	le.PutUint32(pe[0x1a8:], 0x400) // .rsrc virtual size
	le.PutUint32(pe[0x1b0:], 0x400) // .rsrc raw size
	le.PutUint32(pe[0x10c:], 0x400) // resource directory size
	le.PutUint16(pe[0x60e:], 1)
	le.PutUint32(pe[0x610:], 2)
	le.PutUint32(pe[0x614:], 0x80000218)
	le.PutUint16(pe[0x626:], 1)
	le.PutUint32(pe[0x628:], 0x409)
	le.PutUint32(pe[0x62c:], 0x230)
	le.PutUint32(pe[0x630:], 0x2380)
	le.PutUint32(pe[0x634:], 8)
	return pe
}

func checkMissingIconResult(t *testing.T, d *Database, pe []byte) {
	t.Helper()
	result, err := d.MatchObject(nil2ctx(), pe, MatchContext{Type: TypeMSEXE})
	if err != nil || result.Name != "" {
		t.Fatalf("missing icon: result=%+v error=%v; want clean", result, err)
	}
}
