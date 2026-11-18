package engine

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ulikunitz/xz"
)

func xzSingleBlockFixture(t testing.TB, data []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w, err := (xz.WriterConfig{DictCap: 1 << 16, BlockSize: 1 << 20}).NewWriter(&out)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestXZScansDecodedChildWithoutFooter(t *testing.T) {
	base := xzSingleBlockFixture(t, []byte(strings.Repeat(payloadMarker, 16)))
	requireHidden(t, base)
	for _, kept := range []int{0, 1, 6, 11} {
		t.Run(fmt.Sprintf("footer-prefix-%d", kept), func(t *testing.T) {
			partial := append([]byte(nil), base[:len(base)-12+kept]...)
			expectDetected(t, payloadState(t), partial, "Payload.Test")
		})
	}
	missing := base[:len(base)-12]
	multi := xzFixture(t, []byte(strings.Repeat(payloadMarker, 16)))
	requireHidden(t, multi)
	expectDetected(t, payloadState(t), multi[:len(multi)-12], "Payload.Test")
	benign := xzFixture(t, []byte("benign data"))
	expectIncomplete(t, payloadState(t), benign[:len(benign)-12], ErrMalformed)
	first := xzFixture(t, []byte("benign"))
	joined := append(append([]byte(nil), first...), missing...)
	expectDetected(t, payloadState(t), joined, "Payload.Test")
	limited := payloadState(t)
	limited.remaining = int64(len(missing)) + 8
	if sig, err := limited.visit(missing, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("bounded incomplete XZ: sig=%q err=%v, want limit", sig, err)
	}
	oversized := append([]byte(nil), missing...)
	oversized[16] = 40 // 4 GiB dictionary; the missing footer cannot bypass preflight.
	if _, err := payloadState(t).xz(oversized, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("oversized dictionary with missing footer: %v, want limit", err)
	}
	many := xzEmptyBlocks(5, 28) // Five empty blocks each request a 64 MiB dictionary.
	if _, err := payloadState(t).xz(many[:len(many)-12], 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate dictionary overhead with missing footer: %v, want limit", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	s := payloadState(t)
	s.ctx = cancelled
	if _, err := s.xz(missing, 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled incomplete XZ: %v", err)
	}
	corruptPrefix := append(append([]byte(nil), missing...), 0xff)
	if sig, err := payloadState(t).visit(corruptPrefix, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("corrupt footer prefix: sig=%q err=%v", sig, err)
	}
}

func TestXZMissingFooterClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV portable root")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "local.ndb"),
		[]byte("Payload.Test:0:*:"+hex.EncodeToString([]byte(payloadMarker))+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	base := xzSingleBlockFixture(t, []byte(strings.Repeat(payloadMarker, 16)))
	requireHidden(t, base)
	paths := make([]string, 0, 2)
	for name, b := range map[string][]byte{"valid": base, "missing-footer": base[:len(base)-12]} {
		path := filepath.Join(root, name+".xz")
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append([]string{"--database=" + root, "--no-summary"}, paths...)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference scan exit: %v output=%s", err, out)
	}
	if want := filepath.Join(root, "valid.xz") + ": Payload.Test.UNOFFICIAL FOUND"; !strings.Contains(string(out), want) {
		t.Fatalf("reference missing %q in %s", want, out)
	}
	if want := filepath.Join(root, "missing-footer.xz") + ": OK"; !strings.Contains(string(out), want) {
		t.Fatalf("reference missing %q in %s", want, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		t.Log(line)
	}
}

func FuzzXZMissingFooterPreflight(f *testing.F) {
	valid := xzSingleBlockFixture(f, []byte(strings.Repeat(payloadMarker, 16)))
	f.Add(valid[:len(valid)-12])
	largeDicts := xzEmptyBlocks(5, 28)
	f.Add(largeDicts[:len(largeDicts)-12])
	f.Add([]byte{0xfd, '7', 'z', 'X', 'Z', 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		s := scanState{ctx: context.Background()}
		prefix, footer, err := s.xzMissingFooter(data)
		if err != nil {
			return
		}
		if len(footer) != 12 || len(prefix) > len(data) {
			t.Fatalf("invalid recovered tail: prefix=%d footer=%d data=%d", len(prefix), len(footer), len(data))
		}
		_ = preflightXZWithTail(prefix, footer, 1<<20)
	})
}
