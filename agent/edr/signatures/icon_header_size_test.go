package signatures

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func iconHashWire(h iconHash) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%02x", h.width)
	for _, areas := range [][3]iconArea{h.color, h.gray} {
		for _, a := range areas {
			fmt.Fprintf(&b, "%03x%02x%02x", a.v, a.x, a.y)
		}
	}
	for _, areas := range [][3]iconArea{h.bright, h.dark, h.edge, h.noedge} {
		for _, a := range areas {
			fmt.Fprintf(&b, "%02x%02x%02x", a.v, a.x, a.y)
		}
	}
	for _, v := range h.spread {
		fmt.Fprintf(&b, "%02x", v)
	}
	return b.String()
}

func TestOversizedIconHeaderIsSkippedLikeReference(t *testing.T) {
	icon := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	pe := labProgram([]labIcon{icon})
	valid := append([]byte(nil), pe...)
	hash := iconHashWire(labIcons(t, valid)[0].hash)
	meta := parsePE(pe)
	if meta == nil || !meta.valid {
		t.Fatal("generated PE is invalid")
	}
	entries, state := meta.resourcesByID(pe, 3, 1)
	if state != iconFound || len(entries) != 1 {
		t.Fatalf("icon resource: %v %v", entries, state)
	}
	image, ok := meta.iconFileData(pe, entries[0])
	if !ok {
		t.Fatal("icon image resource unreadable")
	}
	binary.LittleEndian.PutUint32(pe[len(pe)-len(image):], uint32(len(image)+1))
	idb := "Probe.Header:A:B:" + hash + "\n"
	rule := "Icon.Header;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"
	d := iconDatabase(t, idb, strings.TrimSpace(rule))
	expectOutcome(t, d, string(valid), MatchContext{Type: TypeMSEXE}, "Icon.Header", "")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
	// ClamAV advances a uint32 offset. A wrapped declaration can reach
	// earlier file bytes, so do not infer a definite miss in that case.
	wrapping := append([]byte(nil), pe...)
	binary.LittleEndian.PutUint32(wrapping[len(pe)-len(image):], ^uint32(0))
	expectOutcome(t, d, string(wrapping), MatchContext{Type: TypeMSEXE}, "Icon.Header", "")
	if _, skip, sure := parseIconImageAt(wrapping[len(pe)-len(image):], uint64(len(pe)-len(image))); sure || skip {
		t.Fatalf("wrapping header must remain undecided: skip=%v sure=%v", skip, sure)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"icon.idb": []byte(idb), "icon.ldb": []byte(rule), "valid.exe": valid, "object.exe": pe, "wrapping.exe": wrapping} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		for _, tc := range []struct {
			name, want string
			detect     bool
		}{{"valid.exe", "Icon.Header.UNOFFICIAL FOUND", true}, {"object.exe", "OK", false}, {"wrapping.exe", "Icon.Header.UNOFFICIAL FOUND", true}} {
			path := filepath.Join(dir, tc.name)
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--debug", path)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.detect || !strings.Contains(string(out), path+": "+tc.want) {
				t.Fatalf("native %s: %v %s", tc.name, err, out)
			}
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "parseicon") || strings.Contains(line, path+": ") {
					t.Log(line)
				}
			}
		}
	}
}
