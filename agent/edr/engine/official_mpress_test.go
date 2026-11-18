package engine

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

// The authenticated MPRESS CBC reads this first-section header, rebuilds the
// LZMA property byte, and appends the decoded page to a copy of the parent.
// This is an inert PE fixture; neither scanner executes its contents.
func officialMPRESSControl(t *testing.T) (parent, child []byte) {
	t.Helper()
	plain := bytes.Repeat([]byte("OPENEDR-HARMLESS-MPRESS-CHILD-"), 147)[:4096]
	var packed bytes.Buffer
	cfg := lzma.WriterConfig{DictCap: 1 << 16, Size: int64(len(plain))}
	w, err := cfg.NewWriter(&packed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	encoded := packed.Bytes()
	if len(encoded) < 13 || encoded[0] != 0x5d || len(encoded)-13 > 0x100 {
		t.Fatalf("unexpected LZMA control encoding: %x", encoded)
	}
	section := make([]byte, 512)
	binary.LittleEndian.PutUint16(section, 1) // one 4096-byte output page
	binary.LittleEndian.PutUint32(section[2:], uint32(len(encoded)-13))
	section[6], section[7] = 0x20, 3 // pb=2, lp=0, lc=3
	copy(section[8:], encoded[13:])
	copy(section[0x180:], []byte{0x80, 0xe1, 0xf0, 0x24, 0x0f, 0xc1, 0xe1, 0x0c})
	copy(section[0x190:], []byte("\x00v1.23"))
	f := simplePE(section)
	f.sections[0].name = ".MPRESS1"
	parent = f.build()
	if bytes.Contains(parent, plain[:64]) {
		t.Fatal("plaintext marker is present in parent")
	}
	child = bytes.Clone(parent)
	at := bytes.Index(child, []byte("MPRESS1"))
	if at < 0 {
		t.Fatal("MPRESS section name missing")
	}
	copy(child[at:at+6], "ClamAV") // the official CBC's recursion guard
	child = append(child, plain...)
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != "64641e99e8dd9c2368e5f80af91e20a2f04193fc267701ca9823e6cd2c16c566" {
		t.Fatalf("inert MPRESS parent SHA-256 = %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "da90c9b7b1c3a00ee88f19723d993a00c4dee2582770be14cb4fb6c27f594c2c" {
		t.Fatalf("inert MPRESS child SHA-256 = %s", got)
	}
	return parent, child
}

func TestOfficialMPRESSExtractedChild(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	program, err := os.ReadFile(filepath.Join(dir, "bytecode-6300337.cbc"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != "8b2cc4d1f2878dc1bf974363899b555e520c4989d67e8e92818cb643b0ff8c9c" {
		t.Fatalf("official CBC SHA-256 = %s", got)
	}
	parent, child := officialMPRESSControl(t)
	rule := fmt.Sprintf("%x:%d:Child.Official.MPRESS\n", md5.Sum(child), len(child))
	s := pipelineState(t, map[string]string{"hook.cbc": string(program), "child.hdb": rule})
	if sig, err := s.visit(parent, 0); sig != "Child.Official.MPRESS" || err != nil {
		t.Fatalf("official CBC child: signature=%q error=%v", sig, err)
	}
	without := pipelineState(t, map[string]string{"child.hdb": rule})
	if sig, err := without.visit(parent, 0); sig != "" || err != nil {
		t.Fatalf("parent without CBC: signature=%q error=%v", sig, err)
	}
	broken := bytes.Clone(parent)
	binary.LittleEndian.PutUint32(broken[0x202:], 0) // header rejects an empty compressed stream
	if sig, err := s.visit(broken, 0); sig != "" || err != nil {
		t.Fatalf("broken MPRESS header: signature=%q error=%v", sig, err)
	}
	t.Logf("parent bytes=%d SHA256=%x; child bytes=%d SHA256=%x", len(parent), sha256.Sum256(parent), len(child), sha256.Sum256(child))
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned ClamAV 1.5.4 scanner: %v", err)
	}
	refDir := t.TempDir()
	for name, data := range map[string][]byte{"hook.cbc": program, "child.hdb": []byte(rule)} {
		if err := os.WriteFile(filepath.Join(refDir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		found bool
	}{
		{"positive.exe", parent, true}, {"broken.exe", broken, false},
	} {
		path := filepath.Join(refDir, tc.name)
		if err := os.WriteFile(path, tc.data, 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if tc.found {
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), tc.name+": Child.Official.MPRESS.UNOFFICIAL FOUND") ||
				!strings.Contains(string(out), fmt.Sprintf("previous tempfile had %d bytes", len(child))) ||
				!strings.Contains(string(out), fmt.Sprintf("fmap_get_hash: calculated sha2-256 hash: %x", sha256.Sum256(child))) {
				t.Fatalf("reference MPRESS output: %v: %s", err, out)
			}
		} else if err != nil || !strings.Contains(string(out), tc.name+": OK") {
			t.Fatalf("reference broken header: %v: %s", err, out)
		}
		t.Logf("ClamAV %s detected=%t", tc.name, tc.found)
	}
	cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode=no", "--no-summary", filepath.Join(refDir, "positive.exe"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "positive.exe: OK") {
		t.Fatalf("reference parent without bytecode: %v: %s", err, out)
	}
	t.Log("ClamAV bytecode-disabled control detected=false; Go controls detected=false")
}
