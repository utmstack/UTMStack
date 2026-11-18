package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func cabQuantumControl(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/libmspack_mszip_lzx_qtm.cab")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "0ce0b55fe705b744d41bb361170c0467db30da0c7f9bdd386d5dade71a78e171" {
		t.Fatalf("unpinned independent CAB control: %v", err)
	}
	return b
}

func TestCABQuantumIndependentChild(t *testing.T) {
	parent := cabQuantumControl(t)
	child := []byte("If you can read this, the Quantum decompressor is working!\n")
	if fmt.Sprintf("%x", sha256.Sum256(child)) != "bdcfdaf09e54d61f950b165b201d4ad5f5acfdecff1fc5641e382aa382c74b45" || bytes.Contains(parent, child) {
		t.Fatal("invalid or exposed Quantum child")
	}
	rule := fmt.Sprintf("%x:%d:CAB.Quantum.Child\n", md5.Sum(child), len(child))
	if sig, err := pipelineState(t, map[string]string{"child.hdb": rule}).visit(parent, 0); sig != "CAB.Quantum.Child" || err != nil {
		t.Fatalf("Quantum CAB: signature=%q error=%v", sig, err)
	}
	referenceQuantumChild(t, parent, rule, "bdcfdaf09e54d61f950b165b201d4ad5f5acfdecff1fc5641e382aa382c74b45")
}

func referenceQuantumChild(t *testing.T, parent []byte, rule, childSHA256 string) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	tool, err := os.ReadFile(filepath.Join(ref, "bin", "clamscan"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned ClamAV 1.5.4 scanner: %v", err)
	}
	dir := t.TempDir()
	for name, data := range map[string][]byte{"child.hdb": []byte(rule), "control.cab": parent} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--debug", "--no-summary", filepath.Join(dir, "control.cab"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	status, ok := err.(*exec.ExitError)
	if childSHA256 == "" {
		if err != nil || !strings.Contains(string(out), "control.cab: OK") {
			t.Fatalf("reference Quantum negative control: %v: %s", err, out)
		}
	} else if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "control.cab: CAB.Quantum.Child.UNOFFICIAL FOUND") ||
		!strings.Contains(string(out), childSHA256) {
		t.Fatalf("reference Quantum extraction: %v: %s", err, out)
	}
	cmd = exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--scan-archive=no", "--no-summary", filepath.Join(dir, "control.cab"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "control.cab: OK") {
		t.Fatalf("reference disabled archive control: %v: %s", err, out)
	}
	if childSHA256 == "" {
		t.Log("ClamAV truncated-input control is clean after failed extraction; Go remains incomplete")
	} else {
		t.Logf("ClamAV and Go detect the exact Quantum child SHA256=%s; disabled archive control is clean", childSHA256)
	}
}

func TestCABQuantumGeneratedOracle(t *testing.T) {
	b, err := os.ReadFile("testdata/quantum/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		File        string `json:"file"`
		Size        int    `json:"size"`
		CABSHA256   string `json:"cab_sha256"`
		ChildSHA256 string `json:"child_sha256"`
		ChildMD5    string `json:"child_md5"`
	}
	if err := json.Unmarshal(b, &cases); err != nil || len(cases) != 18 {
		t.Fatalf("Quantum oracle manifest: cases=%d error=%v", len(cases), err)
	}
	for _, tc := range cases {
		t.Run(tc.File, func(t *testing.T) {
			parent, err := os.ReadFile(filepath.Join("testdata", "quantum", tc.File))
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(parent)) != tc.CABSHA256 {
				t.Fatalf("Quantum oracle input: %v", err)
			}
			if bytes.Contains(parent, []byte("QUANTUM-HIDDEN-CHILD")) {
				t.Fatal("plaintext marker in encoded parent")
			}
			rule := fmt.Sprintf("%s:%d:CAB.Quantum.Child\n", tc.ChildMD5, tc.Size)
			if sig, err := pipelineState(t, map[string]string{"child.hdb": rule}).visit(parent, 0); sig != "CAB.Quantum.Child" || err != nil {
				t.Fatalf("Quantum oracle: signature=%q error=%v", sig, err)
			}
			referenceQuantumChild(t, parent, rule, tc.ChildSHA256)
		})
	}
}

func TestCABQuantumHistoricalCrashControls(t *testing.T) {
	for _, tc := range []struct{ file, sha256 string }{
		{"cve-2014-9556-qtm-infinite-loop.cab", "2e6f57d26e12f3f1c0f1f968d7ebb92e77ea7720ea6f9e18f11eb00ef0e845ad"},
		{"cve-2018-18584-qtm-max-size-block.cab", "3ba0c719d8fb12dc531e2a7790c2d15c12d8840de8e4ddc60572c4145181fbf4"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata", "quantum", tc.file))
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != tc.sha256 {
				t.Fatalf("unpinned crash regression: %v", err)
			}
			s := pipelineState(t, map[string]string{"control.ndb": "Control:0:*:554e52454c415445442d4d41524b4552\n"})
			if sig, err := s.visit(b, 0); sig != "" || err == nil {
				t.Fatalf("crash regression completed as clean: signature=%q error=%v", sig, err)
			}
		})
	}
}

func TestCABQuantumIndependentFrames(t *testing.T) {
	// This historical window-flush regression is a valid 16-frame CAB. A
	// patched ClamAV decodes it to exactly 524159 zero bytes; it is not a
	// malformed-input negative control.
	parent, err := os.ReadFile("testdata/quantum/cve-2010-2801-qtm-flush.cab")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(parent)) != "78bb4c9540e36d3d4228e22cdff0a5677f9c2dddc4239c322277d0b9125d8364" {
		t.Fatalf("unpinned independent frame control: %v", err)
	}
	child := make([]byte, 524159)
	const hash = "0e005e3d2c88e64a5bf04c0a8cecc7ddeb2b820c28d70af04f0aa11d66be4045"
	if fmt.Sprintf("%x", sha256.Sum256(child)) != hash {
		t.Fatal("independent frame child hash")
	}
	rule := fmt.Sprintf("%x:%d:CAB.Quantum.Child\n", md5.Sum(child), len(child))
	if sig, err := pipelineState(t, map[string]string{"child.hdb": rule}).visit(parent, 0); sig != "CAB.Quantum.Child" || err != nil {
		t.Fatalf("Quantum frame control: signature=%q error=%v", sig, err)
	}
	referenceQuantumChild(t, parent, rule, hash)
}

func TestCABQuantumShortInputIsIncomplete(t *testing.T) {
	// Word-sized lookahead must exhaust this stream before a child can be
	// accepted. ClamAV leaves it clean after failing extraction; Go preserves
	// a malformed/incomplete result rather than trusting fabricated output.
	child := make([]byte, 100)
	parent := cabFixture(t, 18<<8|2, [][]byte{{0}}, []cabFixtureFile{{"zeros", 0, 100}}, false)
	at := int(binary.LittleEndian.Uint32(parent[36:]))
	binary.LittleEndian.PutUint32(parent[at:], 0) // omit the optional checksum
	binary.LittleEndian.PutUint16(parent[at+6:], 100)
	rule := fmt.Sprintf("%x:%d:CAB.Quantum.Child\n", md5.Sum(child), len(child))
	if sig, err := pipelineState(t, map[string]string{"child.hdb": rule}).visit(parent, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("short history stream: signature=%q error=%v", sig, err)
	}
	referenceQuantumChild(t, parent, rule, "")
}

func TestCABQuantumLimitsAndBadParameters(t *testing.T) {
	parent := cabQuantumControl(t)
	rules := map[string]string{"control.ndb": "Control:0:*:554e52454c415445442d4d41524b4552\n"}
	if sig, err := pipelineState(t, rules).visit(parent, 0); sig != "" || err != nil {
		t.Fatalf("clean mixed CAB: signature=%q error=%v", sig, err)
	}
	s := pipelineState(t, rules)
	s.remaining = 1
	input := len(parent)
	if _, err := s.cabFolder(parent, cabFolder{start: 323, blocks: 1, compression: 0x1222}, 0, &input, 59, nil); !errors.Is(err, ErrLimit) {
		t.Fatalf("Quantum output limit: %v", err)
	}
	for _, window := range []uint16{0, 9, 22, 31} {
		bad := bytes.Clone(parent)
		binary.LittleEndian.PutUint16(bad[58:], window<<8|2)
		if sig, err := pipelineState(t, rules).visit(bad, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("invalid window %d: signature=%q error=%v", window, sig, err)
		}
	}
	bad := bytes.Clone(parent)
	binary.LittleEndian.PutUint32(bad[323:], 1)
	if sig, err := pipelineState(t, rules).visit(bad, 0); sig != "" || err != nil {
		t.Fatalf("salvaged Quantum checksum: signature=%q error=%v", sig, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s = pipelineState(t, rules)
	s.ctx = ctx
	if sig, err := s.visit(parent, 0); sig != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled CAB: signature=%q error=%v", sig, err)
	}
}
