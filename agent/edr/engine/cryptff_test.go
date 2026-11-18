package engine

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func cryptFFFixture(body []byte) []byte {
	out := make([]byte, 16+len(body))
	copy(out, []byte{0xb6, 0xb9, 0xac, 0xae, 0xfe, 0xff, 0xff, 0xff})
	for i, b := range body {
		out[16+i] = b ^ 0xff
	}
	return out
}

func TestCryptFFDecodedChildAndControl(t *testing.T) {
	payload := []byte("UTMSTACK-CRYPTFF-HARMLESS-CHILD-CONTENT")
	sum := md5.Sum(payload)
	files := map[string]string{"probe.hdb": fmt.Sprintf("%x:%d:CryptFF.Child\n", sum, len(payload))}
	positive := cryptFFFixture(payload)
	expectDetected(t, pipelineState(t, files), positive, "CryptFF.Child")
	control := append([]byte(nil), positive...)
	control[25] ^= 1
	expectScan(t, pipelineState(t, files), control, "", "")
	// Decoded data is retyped and parsed, so a compressed child receives the
	// ordinary nested decompression path as well as direct byte matching.
	expectDetected(t, pipelineState(t, files), cryptFFFixture(gzipFixture(t, payload)), "CryptFF.Child")
	expectScan(t, pipelineState(t, files), positive[:8], "", "")
	s := pipelineState(t, files)
	s.remaining = int64(len(positive) + len(payload) - 1)
	if sig, err := s.visit(positive, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("decoded budget: sig=%q err=%v", sig, err)
	}
}

func TestCryptFFPortableClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("EDR_CLAMAV_PORTABLE not set")
	}
	payload := []byte("UTMSTACK-CRYPTFF-HARMLESS-CHILD-CONTENT")
	sum := md5.Sum(payload)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "probe.hdb"), []byte(fmt.Sprintf("%x:%d:CryptFF.Child\n", sum, len(payload))), 0o600); err != nil {
		t.Fatal(err)
	}
	positive := cryptFFFixture(payload)
	control := append([]byte(nil), positive...)
	control[25] ^= 1
	fixtures := map[string][]byte{
		"positive.cryptff": positive,
		"control.cryptff":  control,
		"nested.cryptff":   cryptFFFixture(gzipFixture(t, payload)),
	}
	var paths []string
	for name, data := range fixtures {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append([]string{"--database=" + dir, "--no-summary"}, paths...)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference exit: %v output=%s", err, out)
	}
	for _, name := range []string{"positive", "nested"} {
		if !strings.Contains(string(out), name+".cryptff: CryptFF.Child.UNOFFICIAL FOUND") {
			t.Fatalf("reference missed %s: %s", name, out)
		}
	}
	if !strings.Contains(string(out), "control.cryptff: OK") {
		t.Fatalf("reference misclassified control: %s", out)
	}
}
