package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/image/webp"
)

// Optional pinned-reference differential for the admitted 5–16 pixel VP8 profile.
func TestWebPTinyLossyReferenceProbe(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned 1.5.4 sigtool")
	}
	tool, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("unpinned sigtool: %v", err)
	}
	encoder, err := os.ReadFile("/opt/homebrew/bin/cwebp")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(encoder)) != "46f37f040f9f248ab5225a0dea0d92b697d08bdb712c8677f5cea922bde3b8e7" {
		t.Skipf("pinned cwebp unavailable: %v", err)
	}
	root := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"), "CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	dir := t.TempDir()
	mismatches, falseCertain, uncertainCases, compared := 0, 0, 0, 0
	var shapes [][2]int
	for width := 5; width <= 16; width++ {
		for height := 5; height <= 16; height++ {
			shapes = append(shapes, [2]int{width, height})
		}
	}
	for _, shape := range shapes {
		for seed := 0; seed < 24; seed++ {
			if target := os.Getenv("EDR_TINY_TARGET"); target != "" && target != fmt.Sprintf("%dx%d:%d", shape[0], shape[1], seed) {
				continue
			}
			img := image.NewNRGBA(image.Rect(0, 0, shape[0], shape[1]))
			for y := 0; y < shape[1]; y++ {
				for x := 0; x < shape[0]; x++ {
					pixel := color.NRGBA{R: byte(x*13 + y*7 + seed*19), G: byte(x*3 + y*19 + seed*17), B: byte(x*23 + y*5 + seed*23), A: 255}
					if seed >= 8 && seed < 16 {
						v := uint32(x*73856093 ^ y*19349663 ^ seed*83492791)
						v ^= v << 13
						v ^= v >> 17
						v ^= v << 5
						pixel = color.NRGBA{R: byte(v), G: byte(v >> 8), B: byte(v >> 16), A: 255}
					} else if seed >= 16 {
						v := byte(seed*23 + ((x/2+y/2)&1)*103)
						pixel = color.NRGBA{R: v, G: 255 - v, B: v / 2, A: 255}
					}
					img.SetNRGBA(x, y, pixel)
				}
			}
			var src bytes.Buffer
			if err := png.Encode(&src, img); err != nil {
				t.Fatal(err)
			}
			input := filepath.Join(dir, "source.png")
			output := filepath.Join(dir, fmt.Sprintf("probe-%dx%d-%d.webp", shape[0], shape[1], seed))
			if err := os.WriteFile(input, src.Bytes(), 0o600); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("/opt/homebrew/bin/cwebp", "-quiet", "-q", strconv.Itoa([]int{15, 50, 87, 100}[seed%4]), input, "-o", output).CombinedOutput(); err != nil {
				t.Fatalf("cwebp: %v: %s", err, out)
			}
			cmd := exec.Command(sigtool, "--fuzzy-img", output)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("reference cannot hash %dx%d seed=%d: %v: %s", shape[0], shape[1], seed, err, out)
				continue
			}
			_, value, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
			if !ok {
				t.Fatalf("unexpected sigtool output: %s", out)
			}
			want, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			observed := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !observed.ok || (observed.hash^want)&^observed.uncertain != 0 {
				t.Fatalf("runtime certain bit mismatch %dx%d seed=%d: %+v want=%016x", shape[0], shape[1], seed, observed, want)
			}
			decoded, err := webp.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			yuv, ok := decoded.(*image.YCbCr)
			if !ok {
				t.Fatalf("unexpected decoded layout %T", decoded)
			}
			got, uncertain := perceptualHash(webpFancyRGB(yuv), 0)
			compared++
			if uncertain != 0 {
				uncertainCases++
			}
			if (got^want)&^uncertain != 0 {
				falseCertain++
			}
			if got != want {
				mismatches++
				if mismatches <= 20 {
					var decodedPNG bytes.Buffer
					if err := png.Encode(&decodedPNG, webpFancyRGB(yuv)); err != nil {
						t.Fatal(err)
					}
					decodedPath := filepath.Join(dir, "decoded.png")
					if err := os.WriteFile(decodedPath, decodedPNG.Bytes(), 0o600); err != nil {
						t.Fatal(err)
					}
					decodedCmd := exec.Command(sigtool, "--fuzzy-img", decodedPath)
					decodedCmd.Env = env
					decodedOut, decodedErr := decodedCmd.CombinedOutput()
					if decodedErr != nil || !strings.HasSuffix(strings.TrimSpace(string(decodedOut)), fmt.Sprintf(": %016x", want)) {
						t.Fatalf("Go-decoded PNG oracle differs from WebP oracle: %s (%v)", decodedOut, decodedErr)
					}
					t.Logf("mismatch %dx%d seed=%d got=%016x uncertain=%016x want=%016x decodedPNG=%s err=%v", shape[0], shape[1], seed, got, uncertain, want, decodedOut, decodedErr)
				}
			}
		}
	}
	t.Logf("tiny WebP: %d compared, %d mismatches, %d false-certain differences, %d uncertain", compared, mismatches, falseCertain, uncertainCases)
}
