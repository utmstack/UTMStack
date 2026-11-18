package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestImageHashBeyond512ReferenceProbe measures whether the 512-pixel exact
// decision profile can safely be widened for ordinary GIFs.
func TestImageHashBeyond512ReferenceProbe(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	shapes := [][2]int{{513, 513}, {600, 400}, {768, 768}, {1024, 1024}, {1024, 32}, {32, 1024}}
	dir := t.TempDir()
	refRoot := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
	mismatches, uncertain, currentUnknown := 0, 0, 0
	for _, shape := range shapes {
		for seed := 0; seed < 12; seed++ {
			for _, transparent := range []bool{false, true} {
				encoded := largeGIFProbe(t, shape[0], shape[1], seed, transparent)
				path := filepath.Join(dir, fmt.Sprintf("%dx%d-%d-%t.gif", shape[0], shape[1], seed, transparent))
				if err := os.WriteFile(path, encoded, 0o600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(sigtool, "--fuzzy-img", path)
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("sigtool %s: %v: %s", path, err, out)
				}
				_, value, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
				if !ok {
					t.Fatalf("unexpected sigtool output: %s", out)
				}
				want, err := strconv.ParseUint(value, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				img, err := gif.Decode(bytes.NewReader(encoded))
				if err != nil {
					t.Fatal(err)
				}
				if transparent {
					img, ok = gifKeepTransparentColor(encoded, img)
					if !ok {
						t.Fatal("cannot restore GIF transparency color")
					}
				}
				got, u := perceptualHash(img, 0)
				if got != want {
					mismatches++
					t.Logf("candidate mismatch %dx%d seed=%d got=%016x ref=%016x uncertain=%016x", shape[0], shape[1], seed, got, want, u)
				}
				if u != 0 {
					uncertain++
				}
				actual := fuzzyImageHash(encoded, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if actual.uncertain != 0 {
					currentUnknown++
				}
				if (actual.hash^want)&^actual.uncertain != 0 {
					t.Errorf("current profile falsely certain: %+v ref=%016x", actual, want)
				}
			}
		}
	}
	t.Logf("beyond-512 candidate: %d/%d mismatches, %d candidate uncertain, %d current uncertain", mismatches, len(shapes)*24, uncertain, currentUnknown)
}

// These hashes are from the SHA-pinned ClamAV 1.5.4 sigtool. A zero-margin
// resampler computes different bits, so a matching rule must never be
// silently treated as false. This default test guards the uncertainty path
// even when the external reference binary is unavailable.
func TestLargeGIFRoundingBoundaryNeverReturnsFalseClean(t *testing.T) {
	for _, tc := range []struct {
		width, height, seed int
		referenceHash       uint64
	}{
		{513, 513, 2, 0x985119f3a60d1977},
		{768, 768, 6, 0x966678bc3cdcc303},
	} {
		data := largeGIFProbe(t, tc.width, tc.height, tc.seed)
		got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || (got.hash^tc.referenceHash)&^got.uncertain != 0 {
			t.Fatalf("%dx%d seed %d: false-certain fuzzy image hash %+v, reference %016x", tc.width, tc.height, tc.seed, got, tc.referenceHash)
		}
		rule := compileTest(t, "ldb", fmt.Sprintf("Large.GIF.Rounding;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.referenceHash))
		name, err := matchObject(t, rule, string(data), MatchContext{Type: TypeGIF})
		if name == "" && err == nil {
			t.Fatalf("%dx%d seed %d: matching reference rule was wrongly decided clean", tc.width, tc.height, tc.seed)
		}
	}
}
