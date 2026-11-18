package engine

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This independently assembled, never-executed PE has one configuration
// record and a 20-byte permuted blob. The blob expands to 4096 'A' bytes.
// Its inverse permutation was generated with Python Random(42).getrandbits(32)
// using the feed program's successive modulo-20 swaps. No malware is present.
func officialLizardNestControl(t *testing.T) (parent, child []byte) {
	t.Helper()
	code := make([]byte, 512)
	copy(code, []byte{0xe8, 0, 0, 0, 0, 0xe9, 0, 0, 0xff, 0xff})
	config := make([]byte, 100)
	le := binary.LittleEndian
	key := uint32(0x12345678)
	le.PutUint32(config, key)
	le.PutUint32(config[8:], 42)
	le.PutUint32(config[12:], 0x1100) // one file-backed block at offset 0x300
	le.PutUint32(config[16:], 20)
	for i := 4; i < len(config); i++ {
		config[i] ^= byte(key)
		key = bits.RotateLeft32(key, -3) - 0x21524111
	}
	copy(code[0x20:], config)
	blob, err := hex.DecodeString("0008100331b000fc414e00000f4c00005a000200")
	if err != nil {
		t.Fatal(err)
	}
	copy(code[0x100:], blob)
	parent = simplePE(code).build()
	parent = append(parent, make([]byte, 0x18000-len(parent))...)
	child = bytes.Repeat([]byte{'A'}, 4096)
	if bytes.Contains(parent, child[:64]) {
		t.Fatal("plaintext marker is present in parent")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != "93feaa363d1e9fd596333df5951066ab7288aee812619d9d326226a2af12741d" {
		t.Fatalf("inert LizardNest parent SHA-256 = %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "6896d9ea3f73a4434f5832bc65714e7d066f177373f36f34dc8a6f735daa41b1" {
		t.Fatalf("inert LizardNest child SHA-256 = %s", got)
	}
	return parent, child
}

func TestOfficialLizardNestExtractedChild(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	program, err := os.ReadFile(filepath.Join(dir, "bytecode-5588995.cbc"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != "c52182036629a6ea8cea53e52113a8e380b1a4c3763bc79072d0ebaff2a37b0f" {
		t.Fatalf("official CBC SHA-256 = %s", got)
	}
	parent, child := officialLizardNestControl(t)
	rule := fmt.Sprintf("%x:%d:Child.Official.LizardNest\n", md5.Sum(child), len(child))
	s := pipelineState(t, map[string]string{"hook.cbc": string(program), "child.hdb": rule})
	if sig, err := s.visit(parent, 0); sig != "Child.Official.LizardNest" || err != nil {
		t.Fatalf("official CBC child: signature=%q error=%v", sig, err)
	}
	without := pipelineState(t, map[string]string{"child.hdb": rule})
	if sig, err := without.visit(parent, 0); sig != "" || err != nil {
		t.Fatalf("parent without CBC: signature=%q error=%v", sig, err)
	}
	short := parent[:0x17fff] // below the official program's minimum input size
	if sig, err := s.visit(short, 0); sig != "" || err != nil {
		t.Fatalf("short parent: signature=%q error=%v", sig, err)
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
	for name, data := range map[string][]byte{"hook.cbc": program, "child.hdb": []byte(rule), "positive.exe": parent, "short.exe": short} {
		if err := os.WriteFile(filepath.Join(refDir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", filepath.Join(refDir, "positive.exe"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	status, ok := err.(*exec.ExitError)
	if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "positive.exe: Child.Official.LizardNest.UNOFFICIAL FOUND") ||
		!strings.Contains(string(out), fmt.Sprintf("previous tempfile had %d bytes", len(child))) ||
		!strings.Contains(string(out), fmt.Sprintf("fmap_get_hash: calculated sha2-256 hash: %x", sha256.Sum256(child))) {
		t.Fatalf("reference LizardNest output: %v: %s", err, out)
	}
	for _, tc := range []struct{ name, bytecode string }{{"short.exe", "yes"}, {"positive.exe", "no"}} {
		cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode-unsigned=yes", "--bytecode="+tc.bytecode, "--no-summary", filepath.Join(refDir, tc.name))
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), tc.name+": OK") {
			t.Fatalf("reference control %s bytecode=%s: %v: %s", tc.name, tc.bytecode, err, out)
		}
	}
	t.Log("ClamAV exact extracted child detected; short-input and disabled-hook controls detected=false in both scanners")
}
