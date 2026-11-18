package engine

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func damagedGzipVariants(t *testing.T, payload []byte) map[string][]byte {
	t.Helper()
	base := gzipFixture(t, payload)
	badCRC := append([]byte(nil), base...)
	badCRC[len(badCRC)-8] ^= 1
	return map[string][]byte{
		"valid":            base,
		"bad-crc":          badCRC,
		"missing-footer":   append([]byte(nil), base[:len(base)-8]...),
		"truncated-middle": append([]byte(nil), base[:len(base)/2]...),
	}
}

func TestGzipScansDecodedChildDespiteLateError(t *testing.T) {
	variants := damagedGzipVariants(t, []byte(payloadMarker))
	for _, name := range []string{"valid", "bad-crc", "missing-footer"} {
		t.Run(name, func(t *testing.T) {
			compressed := variants[name]
			requireHidden(t, compressed)
			expectDetected(t, payloadState(t), compressed, "Payload.Test")
		})
	}
	expectIncomplete(t, payloadState(t), variants["truncated-middle"], ErrMalformed)
	for _, name := range []string{"bad-crc", "missing-footer"} {
		t.Run("benign-"+name, func(t *testing.T) {
			benign := damagedGzipVariants(t, bytes.Repeat([]byte("harmless data "), 10))[name]
			expectIncomplete(t, payloadState(t), benign, ErrMalformed)
		})
	}
	limited := payloadState(t)
	limited.remaining = int64(len(variants["bad-crc"])) + 8
	if sig, err := limited.visit(variants["bad-crc"], 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("bounded damaged gzip: sig=%q err=%v, want limit", sig, err)
	}
}

// Optional independent comparison against the pinned local ClamAV binary.
func TestGzipLateErrorClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV portable root")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "local.ndb"),
		[]byte("Payload.Test:0:*:"+hex.EncodeToString([]byte(payloadMarker))+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	variants := damagedGzipVariants(t, []byte(payloadMarker))
	paths := make([]string, 0, 4)
	for _, name := range []string{"valid", "bad-crc", "missing-footer", "truncated-middle"} {
		path := filepath.Join(root, name+".gz")
		if err := os.WriteFile(path, variants[name], 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	args := append([]string{"--database=" + root, "--no-summary"}, paths...)
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), args...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference scan exit: %v output=%s", err, out)
	}
	for _, name := range []string{"valid", "bad-crc", "missing-footer"} {
		want := filepath.Join(root, name+".gz") + ": Payload.Test.UNOFFICIAL FOUND"
		if !strings.Contains(string(out), want) {
			t.Fatalf("reference %s: missing %q in %s", name, want, out)
		}
	}
	if want := filepath.Join(root, "truncated-middle.gz") + ": OK"; !strings.Contains(string(out), want) {
		t.Fatalf("reference truncated middle: missing %q in %s", want, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		t.Log(line)
	}
}
