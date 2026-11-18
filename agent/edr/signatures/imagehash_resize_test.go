package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/dct32"
)

func resizeProbeImage(w, h, seed int) ([]byte, *image.Gray) {
	img := image.NewGray(image.Rect(0, 0, w, h))
	out := bytes.NewBufferString(fmt.Sprintf("P6\n%d %d\n255\n", w, h))
	x := uint32(seed + 1)
	for y := 0; y < h; y++ {
		for i := 0; i < w; i++ {
			x ^= x << 13
			x ^= x >> 17
			x ^= x << 5
			v := byte(x)
			switch seed % 4 {
			case 0:
				v = 127 + v%2
			case 1:
				v = byte(i*17 + y*31)
			case 2:
				v = byte((i + y) % 2 * 255)
			}
			img.Pix[y*img.Stride+i] = v
			out.Write([]byte{v, v, v})
		}
	}
	return out.Bytes(), img
}

// The alternate profile's hash is calculated from its actual rounded pixels,
// rather than reusing perceptualHash's uncertainty calculation.
func resizeLinuxHash(img image.Image) uint64 {
	b := img.Bounds()
	t := greyPlane(img)
	if b.Dx() != 32 || b.Dy() != 32 {
		t = resample32Profile(resample32Profile(t, b.Dy(), true, resizeSine32), b.Dx(), false, resizeSine32)
	}
	var pixels [32][32]uint8
	for y := range pixels {
		for x := range pixels[y] {
			pixels[y][x] = uint8(math.Round(math.Min(255, math.Max(0, float64(t[y][x])))))
		}
	}
	coeff := dct32.Coefficients(&pixels)
	ordered := coeff
	sort.Slice(ordered[:], func(i, j int) bool { return ordered[i] < ordered[j] })
	median := (ordered[31] + ordered[32]) / 2
	var hash uint64
	for i, value := range coeff {
		if value > median {
			hash |= uint64(1) << (63 - i)
		}
	}
	return hash
}

func TestResizeNativePixels(t *testing.T) {
	root := filepath.Join("testdata", "resize_math")
	data, err := os.ReadFile(filepath.Join(root, "native-disagreement.ppm"))
	if err != nil {
		t.Fatal(err)
	}
	img, err := decodePNM(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, sha string
		sine      func(float32) float32
	}{
		{"mac", "98a6abea6277ec51b582519af7ec8536326db59dc7136fbdf3b8b0fbd92fe72d", resizeMacSine32},
		{"linux", "ce18e0ae6561893a892c52f7f9a0c14349f077594188902fd9328caf33e5c913", resizeSine32},
		{"compiler-mac", "f5813e1ffa11174e2ec2a9ba4effa69109ddccf5b485ec3e0a79fa7b7f301d81", resizeMacSine32},
		{"compiler-linux", "0d6e015fa1d263d59f88377a0d7f3a4ae9c98345084c0578e08009d6829dedc9", resizeSine32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img := img
			if tc.name == "compiler-mac" || tc.name == "compiler-linux" {
				b, err := os.ReadFile(filepath.Join(root, "compiler-rounding.ppm"))
				if err != nil {
					t.Fatal(err)
				}
				img, err = decodePNM(b)
				if err != nil {
					t.Fatal(err)
				}
			}
			gold, err := os.ReadFile(filepath.Join(root, tc.name+".pgm"))
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(gold)) != tc.sha {
				t.Fatalf("independent C pixel oracle changed: %v", err)
			}
			expected, err := decodePNM(gold)
			if err != nil {
				t.Fatal(err)
			}
			plane := resample32Profile(resample32Profile(greyPlane(img), img.Bounds().Dy(), true, tc.sine), img.Bounds().Dx(), false, tc.sine)
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					v := uint8(math.Round(math.Min(255, math.Max(0, float64(plane[y][x])))))
					if r, _, _, _ := expected.At(x, y).RGBA(); v != uint8(r>>8) {
						t.Fatalf("native pixel %d,%d: Go=%d C=%d", x, y, v, r>>8)
					}
				}
			}
		})
	}
}

func TestOwnedResizeNativeRules(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "resize_math", "native-rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gold map[string]struct {
		SHA256, Hash string
		Partial      bool
	}
	if err := json.Unmarshal(raw, &gold); err != nil || len(gold) != 256 {
		t.Fatalf("native resize oracle: %v (%d cases)", err, len(gold))
	}
	shapes := [][2]int{{1, 1}, {1, 31}, {31, 1}, {4, 512}, {512, 4}, {8, 128}, {128, 8}, {10, 11},
		{11, 10}, {16, 16}, {31, 33}, {33, 31}, {32, 32}, {32, 513}, {513, 32}, {37, 511},
		{511, 37}, {127, 128}, {128, 127}, {257, 257}, {513, 513}, {768, 768}, {1023, 7}, {7, 1023},
		{1, 16384}, {16384, 1}, {37, 16384}, {16384, 37}, {2048, 37}, {37, 2048}, {1024, 1}, {1, 1024}}
	disagreements := 0
	for _, shape := range shapes {
		for seed := 0; seed < 8; seed++ {
			t.Run(fmt.Sprintf("%dx%d-%d", shape[0], shape[1], seed), func(t *testing.T) {
				data, img := resizeProbeImage(shape[0], shape[1], seed)
				got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				want, exists := gold[t.Name()[len("TestOwnedResizeNativeRules/"):]]
				if !exists || fmt.Sprintf("%x", sha256.Sum256(data)) != want.SHA256 || fmt.Sprintf("%016x", got.hash) != want.Hash {
					t.Fatalf("pinned native resize oracle mismatch: %+v, want %+v", got, want)
				}
				alternate := resizeLinuxHash(img)
				legacyUncertain := got.hash ^ alternate
				windows := nativeWindowsImageHash(t, data)
				wantUncertain := legacyUncertain | (got.hash ^ windows)
				if (legacyUncertain != 0) != want.Partial {
					t.Fatalf("original Mac/Linux partial oracle changed: mask=%016x want=%+v", legacyUncertain, want)
				}
				if !got.ok || wantUncertain != got.uncertain {
					t.Fatalf("native profiles escaped uncertainty: %+v Linux=%016x", got, alternate)
				}
				if got.hash != alternate {
					disagreements++
					db := compileTest(t, "ldb", fmt.Sprintf("Alternate.Resize;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", alternate))
					expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
					t.Logf("native math disagreement: Mac=%016x Linux=%016x uncertain=%016x", got.hash, alternate, got.uncertain)
				}
				// The independent reference obtains its hash from the original
				// encoded object, then tests both positive and negative rules.
				wrong := ^got.hash
				if got.uncertain == ^uint64(0) {
					t.Fatal("resize uncertainty covers every bit")
				}
				referenceImageRules(t, data, got.hash, wrong, got.uncertain != 0)
			})
		}
	}
	if disagreements != 2 {
		t.Fatalf("native profile disagreement controls changed: %d", disagreements)
	}
	t.Logf("owned resize controls=256 differing native hashes=%d", disagreements)
}

func FuzzResizeProfiles(f *testing.F) {
	f.Add(uint16(8), uint16(127), uint8(0))
	f.Add(uint16(4), uint16(256), uint8(7))
	f.Add(uint16(10), uint16(11), uint8(3))
	f.Fuzz(func(t *testing.T, width, height uint16, seed uint8) {
		_, img := resizeProbeImage(1+int(width%257), 1+int(height%257), int(seed))
		before := append([]byte(nil), img.Pix...)
		h, uncertain := perceptualHash(img, 0)
		if alternate := resizeLinuxHash(img); (h^alternate)&^uncertain != 0 {
			t.Fatalf("false-certain Linux bits: Mac=%016x Linux=%016x uncertain=%016x", h, alternate, uncertain)
		}
		if !bytes.Equal(before, img.Pix) {
			t.Fatal("resizing changed decoded pixels")
		}
	})
}
