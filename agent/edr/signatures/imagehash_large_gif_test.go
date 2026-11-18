package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This optional pinned-reference probe checks whether the exact-decision
// resampler profile can extend beyond 256 pixels for ordinary opaque GIFs.
func TestLargeGIFFuzzyReferenceProbe(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	refRoot := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
	dir := t.TempDir()
	shapes := [][2]int{{257, 257}, {300, 200}, {512, 512}, {511, 37}, {32, 512}, {512, 32}, {300, 512}}
	identical, uncertain := 0, 0
	for _, shape := range shapes {
		for seed := 0; seed < 8; seed++ {
			encoded := largeGIFProbe(t, shape[0], shape[1], seed)
			path := filepath.Join(dir, fmt.Sprintf("gif-%dx%d-%d.gif", shape[0], shape[1], seed))
			if err := os.WriteFile(path, encoded, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(sigtool, "--fuzzy-img", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("sigtool %s: %v: %s", path, err, out)
			}
			_, hashText, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
			if !ok {
				t.Fatalf("bad reference hash for %s: %q", path, out)
			}
			want, err := strconv.ParseUint(hashText, 16, 64)
			if err != nil {
				t.Fatalf("bad reference hash for %s: %q", path, out)
			}
			decoded, err := gif.Decode(bytes.NewReader(encoded))
			if err != nil {
				t.Fatal(err)
			}
			got, uncertainBits := perceptualHash(decoded, 0)
			actual := fuzzyImageHash(encoded, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !actual.ok || actual.hash != got || actual.uncertain != uncertainBits {
				t.Errorf("%s: matcher profile differs: actual %+v, candidate %016x/%016x", filepath.Base(path), actual, got, uncertainBits)
			}
			if (got^want)&^uncertainBits != 0 {
				t.Errorf("%s: certain hash bits differ: Go %016x, reference %016x, uncertain %016x", filepath.Base(path), got, want, uncertainBits)
			}
			if got == want {
				identical++
			}
			if uncertainBits != 0 {
				uncertain++
			}
		}
	}
	t.Logf("large opaque GIFs: %d/%d identical hashes, %d with uncertain bits", identical, len(shapes)*8, uncertain)
	if identical != len(shapes)*8 {
		t.Errorf("too many large GIF hash differences: %d/%d", identical, len(shapes)*8)
	}
}

func largeGIFProbe(t *testing.T, width, height, seed int, transparent ...bool) []byte {
	t.Helper()
	palette := make(color.Palette, 16)
	for i := range palette {
		palette[i] = color.NRGBA{R: uint8((17*i + 29*seed) & 255), G: uint8((53*i + 11*seed) & 255), B: uint8((97*i + 7*seed) & 255), A: 255}
	}
	if len(transparent) != 0 && transparent[0] {
		c := palette[0].(color.NRGBA)
		c.A = 0
		palette[0] = c
	}
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	state := uint32(seed+1)*0x9e3779b9 + uint32(width)*31 + uint32(height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			state = state*1664525 + 1013904223
			var index uint8
			switch seed % 4 {
			case 0:
				index = uint8(state >> 28)
			case 1:
				index = uint8((x/7 + y/11 + seed) & 15)
			case 2:
				index = uint8((x*y/17 + seed) & 15)
			default:
				index = uint8((x/23 ^ y/9 ^ int(state>>30)) & 15)
			}
			img.Pix[y*img.Stride+x] = index
		}
	}
	var encoded bytes.Buffer
	if err := gif.Encode(&encoded, img, nil); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func TestLargeTransparentGIFFuzzyReferenceProbe(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	refRoot := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
	dir := t.TempDir()
	shapes := [][2]int{{257, 257}, {300, 200}, {512, 512}, {511, 37}, {32, 512}, {512, 32}, {300, 512}}
	identical, uncertain := 0, 0
	for _, shape := range shapes {
		for seed := 0; seed < 8; seed++ {
			data := largeGIFProbe(t, shape[0], shape[1], seed, true)
			path := filepath.Join(dir, fmt.Sprintf("transparent-%dx%d-%d.gif", shape[0], shape[1], seed))
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(sigtool, "--fuzzy-img", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("sigtool %s: %v: %s", path, err, out)
			}
			_, hashText, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
			if !ok {
				t.Fatalf("bad reference hash for %s: %q", path, out)
			}
			want, err := strconv.ParseUint(hashText, 16, 64)
			if err != nil {
				t.Fatalf("bad reference hash for %s: %q", path, out)
			}
			decoded, err := gif.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			decoded, ok = gifKeepTransparentColor(data, decoded)
			if !ok {
				t.Fatal("cannot recover transparent palette")
			}
			got, uncertainBits := perceptualHash(decoded, 0)
			actual := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !actual.ok || actual.hash != got || actual.uncertain != uncertainBits {
				t.Errorf("%s: matcher profile differs: actual %+v, candidate %016x/%016x", filepath.Base(path), actual, got, uncertainBits)
			}
			if (got^want)&^uncertainBits != 0 {
				t.Errorf("%s: certain hash bits differ: Go %016x, reference %016x, uncertain %016x", filepath.Base(path), got, want, uncertainBits)
			}
			if got == want {
				identical++
			}
			if uncertainBits != 0 {
				uncertain++
			}
		}
	}
	t.Logf("large transparent GIFs: %d/%d identical hashes, %d with uncertain bits", identical, len(shapes)*8, uncertain)
	if identical != len(shapes)*8 {
		t.Errorf("too many large transparent GIF hash differences: %d/%d", identical, len(shapes)*8)
	}
}

func TestLargeGIFExactRuleDecisions(t *testing.T) {
	data := largeGIFProbe(t, 257, 257, 0)
	got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	const referenceHash uint64 = 0xd444b11ed4e4f0e7
	if !got.ok || got.hash != referenceHash || got.uncertain != 0 {
		t.Fatalf("large GIF reference hash: %+v, want %016x with no uncertain bits", got, referenceHash)
	}
	rule := func(hash uint64) *Database {
		return compileTest(t, "ldb", fmt.Sprintf("Large.GIF;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
	}
	ctx := MatchContext{Type: TypeGIF}
	if name, err := matchObject(t, rule(referenceHash), string(data), ctx); name != "Large.GIF" || err != nil {
		t.Fatalf("positive large-GIF rule: %q, %v", name, err)
	}
	if name, err := matchObject(t, rule(referenceHash^(1<<40)), string(data), ctx); name != "" || err != nil {
		t.Fatalf("one-bit negative large-GIF rule: %q, %v", name, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		path := filepath.Join(dir, "large.gif")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			hash   uint64
			want   string
			detect bool
		}{{referenceHash, "Large.GIF.UNOFFICIAL FOUND", true}, {referenceHash ^ (1 << 40), "OK", false}} {
			ruleText := fmt.Sprintf("Large.GIF;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash)
			if err := os.WriteFile(filepath.Join(dir, "large.ldb"), []byte(ruleText), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", path)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.detect || !strings.Contains(string(out), path+": "+tc.want) {
				t.Fatalf("native large-GIF rule %016x: %v %s", tc.hash, err, out)
			}
		}
	}
}
