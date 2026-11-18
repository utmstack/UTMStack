package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/jpeg154"
	"golang.org/x/image/bmp"
)

// This optional historical comparison verifies why the native hash path uses
// a project-owned JPEG decoder. It compares the standard Go decoder's hash
// with ClamAV's hash of a BMP containing those exact Go-decoded pixels.
func TestJPEGDecoderDifferential(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	const pinnedSHA256 = "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df"
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != pinnedSHA256 {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	refRoot := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
	refHash := func(path string) string {
		t.Helper()
		cmd := exec.Command(sigtool, "--fuzzy-img", path)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("sigtool %s: %v: %s", path, err, out)
		}
		_, hash, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
		if !ok || len(hash) != 16 {
			t.Fatalf("unexpected sigtool output for %s: %q", path, out)
		}
		return hash
	}
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	dir := t.TempDir()
	mismatches := 0
	for _, c := range cases {
		if !strings.HasSuffix(c.name, ".jpg") {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		img, err := jpeg.Decode(bytes.NewReader(c.data))
		if err != nil {
			continue
		}
		hash, uncertain := perceptualHash(img, imageHashJPEGError)
		got := imageHash{hash: hash, uncertain: uncertain, ok: true}
		if got.hash == want {
			continue
		}
		mismatches++
		t.Logf("%s: Go %016x, reference %s, uncertain %016x", c.name, got.hash, c.hash, got.uncertain)
		jpegPath := filepath.Join(dir, c.name)
		if err := os.WriteFile(jpegPath, c.data, 0o600); err != nil {
			t.Fatal(err)
		}
		if hash := refHash(jpegPath); hash != c.hash {
			t.Errorf("%s: reference hash %s, fixture %s", c.name, hash, c.hash)
		}
		bmpPath := strings.TrimSuffix(jpegPath, ".jpg") + ".bmp"
		f, err := os.Create(bmpPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := bmp.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if hash := refHash(bmpPath); hash != fmt.Sprintf("%016x", got.hash) {
			t.Errorf("%s: ClamAV hash of Go pixels %s, Go JPEG hash %016x", c.name, hash, got.hash)
		}
	}
	t.Logf("%d standard-library JPEG differences isolated before image hashing", mismatches)
}

// ClamAV hashes the JPEG and a BMP of our actual reference-compatible RGB
// output alike. The large299 hash gap therefore occurs after decoding.
// This optional test needs the pinned ClamAV 1.5.4 executable.
func TestJPEGLargeDecodedPixelDifferential(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	const pinnedSHA256 = "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df"
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != pinnedSHA256 {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_large_oracle.tar.gz")
	for _, c := range cases {
		if c.name != "large299.jpg" {
			continue
		}
		decoded, err := jpeg154.Decode(bytes.NewReader(c.data))
		if err != nil {
			t.Fatal(err)
		}
		ycbcr, ok := decoded.(*image.YCbCr)
		if !ok {
			t.Fatalf("expected YCbCr, got %T", decoded)
		}
		rgb := image.NewNRGBA(ycbcr.Bounds())
		for y := rgb.Bounds().Min.Y; y < rgb.Bounds().Max.Y; y++ {
			for x := rgb.Bounds().Min.X; x < rgb.Bounds().Max.X; x++ {
				yy := ycbcr.Y[ycbcr.YOffset(x, y)]
				cb := referenceJPEGChroma(ycbcr, x, y, true)
				cr := referenceJPEGChroma(ycbcr, x, y, false)
				r, g, b := referenceJPEGToRGB(yy, cb, cr)
				rgb.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
			}
		}
		dir := t.TempDir()
		path := filepath.Join(dir, "reference-rgb.bmp")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := bmp.Encode(f, rgb); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		refRoot := filepath.Dir(filepath.Dir(sigtool))
		cmd := exec.Command(sigtool, "--fuzzy-img", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("sigtool: %v: %s", err, out)
		}
		if !strings.HasSuffix(strings.TrimSpace(string(out)), ": "+c.hash) {
			t.Errorf("ClamAV BMP hash %q, JPEG oracle hash %s", out, c.hash)
		}
		grey := greyPlane(rgb)
		resized := resample32(resample32(grey, rgb.Bounds().Dy(), true), rgb.Bounds().Dx(), false)
		small := image.NewGray(image.Rect(0, 0, 32, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				v := math.Min(math.Max(float64(resized[y][x]), 0), 255)
				small.SetGray(x, y, color.Gray{Y: uint8(math.Round(v))})
			}
		}
		resizedPath := filepath.Join(dir, "native-resized.bmp")
		f, err = os.Create(resizedPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := bmp.Encode(f, small); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		cmd = exec.Command(sigtool, "--fuzzy-img", resizedPath)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(refRoot, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(refRoot, "etc", "certs"))
		out, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("sigtool resized BMP: %v: %s", err, out)
		}
		native := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !strings.HasSuffix(strings.TrimSpace(string(out)), fmt.Sprintf(": %016x", native.hash)) {
			t.Errorf("ClamAV resized BMP hash %q, Go JPEG hash %016x", out, native.hash)
		}
		return
	}
	t.Fatal("missing large299.jpg")
}
