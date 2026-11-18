package signatures

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestVersionInfoReferenceProbe records damaged-PE boundaries and optionally
// emits inert variants for a differential oracle run.
func TestVersionInfoReferenceProbe(t *testing.T) {
	dir := os.Getenv("EDR_VI_FIXTURE_DIR")
	key := hex.EncodeToString(utf16z("CompanyName")[:22])
	rule := "Vi.Key;Engine:51-255,Target:1;0;VI:" + key + "\n"
	base := peWithVersion(versionResource(versionString("CompanyName", "Evil Corp")))
	variants := map[string][]byte{"valid.exe": base}
	invalidEntry := append([]byte(nil), base...)
	binary.LittleEndian.PutUint32(invalidEntry[168:], 0x3000)
	variants["invalid-entry.exe"] = invalidEntry
	noSections := append([]byte(nil), base...)
	binary.LittleEndian.PutUint16(noSections[134:], 0)
	variants["zero-sections.exe"] = noSections
	shortOptional := append([]byte(nil), base...)
	binary.LittleEndian.PutUint16(shortOptional[148:], 80)
	variants["short-optional.exe"] = shortOptional
	if dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte(rule), 0o600); err != nil {
			t.Fatal(err)
		}
		for name, data := range variants {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	d := compileTest(t, "ldb", rule)
	exe := MatchContext{Type: TypeMSEXE}
	expectOutcome(t, d, string(base), exe, "Vi.Key", "")
	expectOutcome(t, d, string(invalidEntry), exe, "", "")
	expectOutcome(t, d, string(noSections), exe, "", "")
	expectOutcome(t, d, string(shortOptional), exe, "", "")
}
