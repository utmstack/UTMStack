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

// A found group data entry with an unreadable image RVA or declared group
// length is ignored by ClamAV's group scanner. The resource tree itself is
// readable, so the icon condition has a definite nonmatch.
func TestInvalidIconGroupDataIsClean(t *testing.T) {
	const flatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	const idb = "Win.Test.IconA:MS:IE:" + flatHash + "\n"
	const rule = "Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;4d5a\n"
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	for _, tc := range []struct {
		name   string
		change func([]byte, int)
		want   string
	}{
		{"baseline", nil, "Icon.Live"},
		{"unmapped-group-rva", func(pe []byte, at int) { binary.LittleEndian.PutUint32(pe[at:], 0xffffffff) }, ""},
		{"oversized-group-length", func(pe []byte, at int) { binary.LittleEndian.PutUint32(pe[at+4:], 0xffffffff) }, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flat := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
			pe := labProgram([]labIcon{flat})
			meta := parsePE(pe)
			if meta == nil {
				t.Fatal("PE resources missing")
			}
			entryAt := -1
			if !meta.resources(pe, 14, func(entry uint32) bool {
				if at, ok := meta.rawAddr(entry); ok {
					entryAt = at
				}
				return true
			}) || entryAt < 0 || entryAt+16 > len(pe) {
				t.Fatal("group data entry unavailable")
			}
			if tc.change != nil {
				tc.change(pe, entryAt)
			}
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				dir := t.TempDir()
				for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "object.exe": pe} {
					if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
				object := filepath.Join(dir, "object.exe")
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", object)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
					"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				status := object + ": OK"
				if tc.want != "" {
					status = object + ": " + tc.want + ".UNOFFICIAL FOUND"
				}
				validExit := err == nil
				if tc.want != "" {
					var exit *exec.ExitError
					validExit = errors.As(err, &exit) && exit.ExitCode() == 1
				}
				if !validExit || !strings.Contains(string(out), status) {
					t.Fatalf("ClamAV 1.5.4 group-data scan: %v: %s", err, out)
				}
				t.Logf("ClamAV 1.5.4: %s", strings.TrimSpace(string(out)))
			}
			expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, tc.want, "")
		})
	}
}
