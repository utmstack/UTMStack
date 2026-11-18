package signatures

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"testing"
)

// testdata/image_oracle.tar.gz holds 369 small synthetic PNG, GIF and JPEG
// images (random shapes and noise in many format variants, generated for
// this test) and hashes.tsv: each image's fuzzy hash as the reference's own
// signature tool computes it (release 1.4.3, its image-hash command).
//
// EDR_IMAGE_ORACLE names a larger file of path<TAB>hash lines in the same
// format (ERR lines are skipped), such as the 4,446 images of a Windows
// Server installation; EDR_IMAGE_ORACLE_BASE replaces a leading "/tmp/" in
// its paths.
func TestImageHashAgreesWithReference(t *testing.T) {
	f, err := os.Open("testdata/image_oracle.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(tr)
		files[h.Name] = b
	}
	var cases []imageCase
	for _, line := range strings.Split(strings.TrimSpace(string(files["hashes.tsv"])), "\n") {
		name, hash, _ := strings.Cut(line, "\t")
		// This older manifest used ClamAV 1.4.3, whose JPEG decoder hashes
		// differ from the 1.5.4 target. The separate 1.5.4 oracle covers JPEG.
		if strings.HasSuffix(name, ".jpg") {
			continue
		}
		cases = append(cases, imageCase{name, files[name], hash})
	}
	checkImageCases(t, cases, 1)
	if list := os.Getenv("EDR_IMAGE_ORACLE"); list != "" {
		lf, err := os.Open(list)
		if err != nil {
			t.Fatal(err)
		}
		defer lf.Close()
		var more []imageCase
		sc := bufio.NewScanner(lf)
		for sc.Scan() {
			path, hash, _ := strings.Cut(sc.Text(), "\t")
			if strings.HasPrefix(hash, "ERR") {
				continue
			}
			if base := os.Getenv("EDR_IMAGE_ORACLE_BASE"); base != "" {
				path = base + strings.TrimPrefix(path, "/tmp/")
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			more = append(more, imageCase{path, b, hash})
		}
		checkImageCases(t, more, 0.9)
	}
}

// The original mixed-format fixture was hashed with ClamAV 1.4.3. Version
// 1.5.4 changes nine JPEG hashes, so the target release has its own manifest.
func TestImageHash154Oracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	if len(cases) != 369 {
		t.Fatalf("got %d ClamAV 1.5.4 fixtures, want 369", len(cases))
	}
	checkImageCases(t, cases, 1)
	for _, c := range cases {
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok {
			t.Errorf("%s: ClamAV-hashable image was not decoded", c.name)
		}
	}
	// These JPEGs differ under Go's native chroma/color path but match the
	// pinned reference after its interpolation and fixed-point conversion.
	wantExact := map[string]bool{
		"s0497.jpg": true, "s0445.jpg": true, "s0252.jpg": true,
		"s0134.jpg": true, "s0070.jpg": true, "s0026.jpg": true,
	}
	for _, c := range cases {
		if !wantExact[c.name] {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || got.hash != want {
			t.Errorf("%s: JPEG color/interpolation hash %016x, reference %016x", c.name, got.hash, want)
		}
		delete(wantExact, c.name)
	}
	if len(wantExact) != 0 {
		t.Errorf("missing JPEG reference cases: %v", wantExact)
	}
}

// EDR_JPEG_PROBE_ARCHIVE can point to a generated independent corpus with
// hashes.tsv. Exactness is logged; certain bits must always agree.
func TestJPEGExternalProbeOracle(t *testing.T) {
	path := os.Getenv("EDR_JPEG_PROBE_ARCHIVE")
	if path == "" {
		t.Skip("set EDR_JPEG_PROBE_ARCHIVE for independent JPEG reference cases")
	}
	cases, _ := loadImageOracleCases(t, path)
	checkImageCases(t, cases, 0)
	for _, c := range cases {
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if got.ok && got.hash != want {
			t.Logf("%s: got %016x, reference %016x, uncertain %016x", c.name, got.hash, want, got.uncertain)
		}
	}
}

// A JFIF case with no floating-point boundary ambiguity can decide both a
// matching fuzzy rule and a one-bit nonmatch. The hash is from ClamAV 1.5.4.
func TestJPEGFuzzyRuleDecisions(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	var chosen *imageCase
	var hash uint64
	for i := range cases {
		c := &cases[i]
		if !strings.HasSuffix(c.name, ".jpg") || bytes.Contains(c.data[:min(len(c.data), 64)], []byte("Adobe")) ||
			!referenceJPEGProfile(c.data) {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if got.ok && got.hash == want && got.uncertain == 0 {
			chosen, hash = c, want
			break
		}
	}
	if chosen == nil {
		t.Fatal("no exact-decision JPEG reference case")
	}
	t.Logf("exact-decision fixture: %s, hash %016x", chosen.name, hash)
	rule := func(h uint64) *Database {
		return compileTest(t, "ldb", fmt.Sprintf("JPEG.Img;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", h))
	}
	ctx := MatchContext{Type: TypeJPEG}
	if name, err := matchObject(t, rule(hash), string(chosen.data), ctx); name != "JPEG.Img" || err != nil {
		t.Fatalf("%s exact rule: %q, %v", chosen.name, name, err)
	}
	if name, err := matchObject(t, rule(hash^1<<40), string(chosen.data), ctx); name != "" || err != nil {
		t.Fatalf("%s one-bit nonmatch: %q, %v", chosen.name, name, err)
	}
}

// Adobe CMYK JPEGs that used to be undecodable can now decide exact fuzzy
// rules after interpolating all four components and using the reference's
// 8-bit color multiplication.
func TestCMYKFuzzyRuleDecisions(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	var chosen *imageCase
	var hash uint64
	for i := range cases {
		c := &cases[i]
		if !strings.HasSuffix(c.name, ".jpg") || !bytes.Contains(c.data[:min(len(c.data), 64)], []byte("Adobe")) ||
			!referenceJPEGProfile(c.data) || c.hash == "8000000000000000" {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if got.ok && got.hash == want && got.uncertain == 0 {
			chosen, hash = c, want
			break
		}
	}
	if chosen == nil {
		t.Fatal("no exact-decision Adobe CMYK JPEG reference case")
	}
	t.Logf("exact-decision CMYK fixture: %s, hash %016x", chosen.name, hash)
	rule := func(h uint64) *Database {
		return compileTest(t, "ldb", fmt.Sprintf("CMYK.Img;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", h))
	}
	ctx := MatchContext{Type: TypeJPEG}
	if name, err := matchObject(t, rule(hash), string(chosen.data), ctx); name != "CMYK.Img" || err != nil {
		t.Fatalf("%s exact rule: %q, %v", chosen.name, name, err)
	}
	if name, err := matchObject(t, rule(hash^1<<40), string(chosen.data), ctx); name != "" || err != nil {
		t.Fatalf("%s one-bit nonmatch: %q, %v", chosen.name, name, err)
	}
}

// These independent JPEGs include nine near-median cases that were
// undecidable with the old f64 transform. The fixed f32 DCT agrees on all
// hashes; the mixed-axis resize guard keeps one rounded-pixel case partial.
func TestJPEGBoundaryOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_boundary_oracle.tar.gz")
	if len(cases) != 24 {
		t.Fatalf("got %d JPEG boundary cases, want 24", len(cases))
	}
	checkImageCases(t, cases, 1)
	for _, c := range cases {
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || !referenceJPEGProfile(c.data) || got.hash != want || got.uncertain != 0 {
			t.Errorf("%s: exact DCT boundary decision: %+v", c.name, got)
		}
	}
}

// The independent Adobe CMYK archive spans baseline and progressive frames,
// 4:4:4/4:2:2/4:2:0 sampling, and multiple dimensions and quality settings.
func TestCMYKIndependentOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_cmyk_oracle.tar.gz")
	if len(cases) != 24 {
		t.Fatalf("got %d CMYK JPEG cases, want 24", len(cases))
	}
	checkImageCases(t, cases, 1)
	for _, c := range cases {
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || !referenceJPEGProfile(c.data) || got.uncertain != 0 {
			t.Errorf("%s: CMYK exact-decision profile is incomplete: %+v", c.name, got)
		}
	}
}

// jpegtran preserves decoded pixels while adding restart markers to baseline
// and progressive JFIF/CMYK files. The pinned reference hashes these files
// identically to their originals; the native decoder must keep all certain
// bits and permit exact rule decisions away from transform boundaries.
func TestJPEGRestartOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_restart_oracle.tar.gz")
	if len(cases) != 33 {
		t.Fatalf("got %d restart-marker JPEGs, want 33", len(cases))
	}
	checkImageCases(t, cases, 1)
	exact, uncertain := 0, 0
	for _, c := range cases {
		if !bytes.Contains(c.data, []byte{0xff, 0xdd}) || !bytes.Contains(c.data, []byte{0xff, 0xd0}) || !referenceJPEGProfile(c.data) {
			t.Errorf("%s: restart-marker profile missing", c.name)
			continue
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if got.uncertain == 0 {
			exact++
		} else {
			uncertain++
		}
	}
	if exact != 33 || uncertain != 0 {
		t.Fatalf("restart decisions: %d exact, %d uncertain; want 33, 0", exact, uncertain)
	}
}

func TestJPEGRestartRuleDecisions(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_restart_oracle.tar.gz")
	chosen := map[string]bool{"r007.jpg": true, "r029.jpg": true}
	for _, c := range cases {
		if !chosen[c.name] {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || got.hash != want || got.uncertain != 0 {
			t.Fatalf("%s: restart rule cannot be decided exactly: %+v", c.name, got)
		}
		rule := func(hash uint64) *Database {
			return compileTest(t, "ldb", fmt.Sprintf("Restart.Img;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
		}
		ctx := MatchContext{Type: TypeJPEG}
		if name, err := matchObject(t, rule(want), string(c.data), ctx); name != "Restart.Img" || err != nil {
			t.Errorf("%s positive rule: %q, %v", c.name, name, err)
		}
		if name, err := matchObject(t, rule(want^(1<<40)), string(c.data), ctx); name != "" || err != nil {
			t.Errorf("%s one-bit negative: %q, %v", c.name, name, err)
		}
		delete(chosen, c.name)
	}
	if len(chosen) != 0 {
		t.Errorf("missing restart fixtures: %v", chosen)
	}
}

// Independently generated 513–2048-pixel JPEGs include one resize-rounding
// counterexample. Its differing bits must stay uncertain while other images
// in the tested profile can make exact rule decisions.
func TestJPEGLargeOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_large_oracle.tar.gz")
	if len(cases) != 27 {
		t.Fatalf("got %d large JPEG cases, want 27", len(cases))
	}
	checkImageCases(t, cases, 1)
	counterexample := false
	exact := 0
	for _, c := range cases {
		if !referenceJPEGProfile(c.data) {
			t.Errorf("%s: tested large JPEG outside exact-decoder profile", c.name)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if c.name == "large299.jpg" {
			counterexample = true
			wantUncertain := want ^ nativeWindowsImageHash(t, c.data)
			if !got.ok || got.hash != want || got.uncertain != wantUncertain {
				t.Errorf("%s: owned resize regression: %+v", c.name, got)
			}
		}
		if got.ok && got.hash == want && got.uncertain == 0 {
			exact++
			t.Logf("exact large JPEG %s: %016x", c.name, got.hash)
		}
	}
	if !counterexample {
		t.Fatal("missing large JPEG resize-rounding counterexample")
	}
	if exact == 0 {
		t.Fatal("no exact-decision large JPEG case")
	}
}

// A larger JFIF and Adobe CMYK image each make positive and one-bit negative
// fuzzy-rule decisions, rather than merely reproducing the reference hash.
func TestJPEGLargeRuleDecisions(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/jpeg_large_oracle.tar.gz")
	wantCases := map[string]bool{"large047.jpg": false, "large003.jpg": false}
	for _, c := range cases {
		if _, selected := wantCases[c.name]; !selected {
			continue
		}
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !referenceJPEGProfile(c.data) || !got.ok || got.hash != want || got.uncertain != 0 {
			t.Fatalf("%s: expected exact hash %016x, got %+v", c.name, want, got)
		}
		rule := func(hash uint64) *Database {
			return compileTest(t, "ldb", fmt.Sprintf("Large.Img;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
		}
		ctx := MatchContext{Type: TypeJPEG}
		if name, err := matchObject(t, rule(want), string(c.data), ctx); name != "Large.Img" || err != nil {
			t.Errorf("%s positive rule: %q, %v", c.name, name, err)
		}
		if name, err := matchObject(t, rule(want^(1<<40)), string(c.data), ctx); name != "" || err != nil {
			t.Errorf("%s one-bit negative: %q, %v", c.name, name, err)
		}
		wantCases[c.name] = true
	}
	for name, found := range wantCases {
		if !found {
			t.Errorf("missing exact large JPEG case %s", name)
		}
	}
}

// A four-component frame with subsampled K must stay bounded even when its
// Adobe transform asks for YCbCrK rather than the tested CMYK color path.
func TestSubsampledYCCKDoesNotPanic(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	for _, c := range cases {
		if c.name != "s0212.jpg" {
			continue
		}
		data := append([]byte(nil), c.data...)
		i := bytes.Index(data, []byte("Adobe"))
		if i < 0 || i+11 >= len(data) || data[i+11] != 0 {
			t.Fatal("expected an Adobe CMYK marker")
		}
		data[i+11] = 2 // Request YCbCrK with a subsampled K plane.
		got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok {
			t.Fatal("subsampled YCbCrK was not decoded")
		}
		return
	}
	t.Fatal("missing CMYK fixture")
}

type imageCase struct {
	name string
	data []byte
	hash string
}

// checkImageCases requires that no certain bit differs from the reference
// and that most hashes are identical.
func checkImageCases(t *testing.T, cases []imageCase, minExact float64) {
	t.Helper()
	exact, decoded, uncertain := 0, 0, 0
	for _, c := range cases {
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatalf("%s: bad hash %q", c.name, c.hash)
		}
		ih := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !ih.ok {
			if os.Getenv("EDR_IMAGE_DEBUG") != "" {
				t.Logf("unsupported image oracle case: %s (%d bytes)", c.name, len(c.data))
			}
			continue // a format this engine does not decode: unknown when reached
		}
		decoded++
		if ih.hash == want {
			exact++
		}
		if ih.uncertain != 0 {
			uncertain++
		}
		if d := (ih.hash ^ want) &^ ih.uncertain; d != 0 {
			t.Errorf("%s: %d certain bits differ from the reference (ours %016x, uncertain %016x, reference %016x)", c.name, bits.OnesCount64(d), ih.hash, ih.uncertain, want)
		}
	}
	t.Logf("%d images: %d decoded, %d hashes identical to the reference, %d with uncertain bits", len(cases), decoded, exact, uncertain)
	if decoded == 0 || float64(exact) < minExact*float64(decoded) {
		t.Errorf("only %d of %d hashes identical", exact, decoded)
	}
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// An image flat after resizing has only its brightness bit (none when
// black), whatever its transparency: the reference ignores alpha.
func TestImageHashOfFlatImages(t *testing.T) {
	for _, tt := range []struct {
		c    color.Color
		want uint64
	}{{color.NRGBA{200, 10, 10, 255}, 1 << 63}, {color.NRGBA{255, 255, 255, 0}, 1 << 63}, {color.NRGBA{0, 0, 0, 255}, 0}} {
		img := image.NewNRGBA(image.Rect(0, 0, 57, 23))
		for i := 0; i < len(img.Pix); i += 4 {
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = tt.c.(color.NRGBA).R, tt.c.(color.NRGBA).G, tt.c.(color.NRGBA).B, tt.c.(color.NRGBA).A
		}
		ih := fuzzyImageHash(encodePNG(t, img), &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !ih.ok || ih.hash != tt.want || ih.uncertain != 0 {
			t.Errorf("%v: %+v, want %016x certain", tt.c, ih, tt.want)
		}
	}
}

// A narrow but extremely wide image must be rejected from the fuzzy path
// before decoding or allocating a 32-row resampling intermediate.
func TestFuzzyImageDimensionLimit(t *testing.T) {
	data := append([]byte(nil), []byte("\x89PNG\r\n\x1a\n")...)
	chunk := make([]byte, 4+4+13+4)
	binary.BigEndian.PutUint32(chunk[:4], 13)
	copy(chunk[4:8], "IHDR")
	binary.BigEndian.PutUint32(chunk[8:12], imageHashMaxDimension+1)
	binary.BigEndian.PutUint32(chunk[12:16], 1)
	chunk[16], chunk[17] = 8, 2 // RGB, 8-bit
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	data = append(data, chunk...)
	if cfg, err := png.DecodeConfig(bytes.NewReader(data)); err != nil || cfg.Width != imageHashMaxDimension+1 {
		t.Fatalf("PNG fixture config: %+v, %v", cfg, err)
	}
	w := &matchWork{ctx: context.Background(), limit: 1 << 40}
	if got := fuzzyImageHash(data, w); got.ok || w.steps != 0 {
		t.Fatalf("oversized image decoded or charged work: %+v, %d steps", got, w.steps)
	}
}

// A larger independent sigtool corpus exercises small lossless images close
// to resize and median thresholds. Every certain bit must agree with ClamAV.
func TestImagePrecisionOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_precision_oracle.tar.gz")
	if len(cases) != 1000 {
		t.Fatalf("got %d precision fixtures, want 1000", len(cases))
	}
	for _, c := range cases {
		if !fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Errorf("%s was not decoded", c.name)
		}
	}
	checkImageCases(t, cases, 1)
}

// A no-resize corpus catches changes in the reference's f32 luma rounding.
// With fused multiply-add, four images acquired certainly wrong hash bits.
func TestImageLumaOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_luma_oracle.tar.gz")
	if len(cases) != 1000 {
		t.Fatalf("got %d luma fixtures, want 1000", len(cases))
	}
	checkImageCases(t, cases, 1)
}

// Samples beyond the former 256-pixel precision limit must hash exactly and
// yield a decision for exact image rules, rather than an incomplete verdict.
func TestImageLargeLosslessOracle(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_large_oracle.tar.gz")
	if len(cases) != 24 {
		t.Fatalf("got %d large fixtures, want 24", len(cases))
	}
	checkImageCases(t, cases, 1)
	for _, c := range cases {
		got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || got.uncertain != 0 {
			t.Errorf("%s: undecided large image hash: %+v", c.name, got)
		}
	}
}

func TestJPEG2000FuzzyRuleIsKnownFalse(t *testing.T) {
	data, err := os.ReadFile("testdata/jp2_no_fuzzy_hash.jp2")
	if err != nil {
		t.Fatal(err)
	}
	d := compileTest(t, "ldb", "JP2.Fuzzy;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000#0\n")
	expectOutcome(t, d, string(data), MatchContext{Type: TypeGraphics}, "", "")
}

// A fuzzy image rule matches an image whose hash is certainly its own, is
// false on an image whose certain bits differ, and is unknown on an image
// this engine cannot decode or whose uncertain bits could make the hash.
func TestFuzzyImageRule(t *testing.T) {
	// The first of a few smooth off-center patterns whose hash is certain.
	var data []byte
	var ih imageHash
	for k := 1; k < 40 && (data == nil || ih.uncertain != 0); k++ {
		img := image.NewNRGBA(image.Rect(0, 0, 96, 80))
		for y := 0; y < 80; y++ {
			for x := 0; x < 96; x++ {
				dx, dy := float64(x-17*k%96), float64(y-11*k%80)
				v := uint8(int(math.Sqrt(dx*dx+2*dy*dy)*float64(k%5+2)) % 256)
				img.SetNRGBA(x, y, color.NRGBA{v, uint8(x * 2), uint8(y * 3), 255})
			}
		}
		data = encodePNG(t, img)
		ih = fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	}
	if !ih.ok || ih.uncertain != 0 {
		t.Fatalf("no test image with a certain hash: %+v", ih)
	}
	rule := func(h uint64) *Database {
		return compileTest(t, "ldb", fmt.Sprintf("Img;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", h))
	}
	png := MatchContext{Type: TypePNG}
	if name, err := matchObject(t, rule(ih.hash), string(data), png); name != "Img" || err != nil {
		t.Fatalf("own hash: %q, %v", name, err)
	}
	if name, err := matchObject(t, rule(ih.hash^1<<40), string(data), png); name != "" || err != nil {
		t.Fatalf("other hash: %q, %v", name, err)
	}
	expectOutcome(t, rule(ih.hash), "\x89PNG\r\n\x1a\nnot an image", png, "", UnknownFuzzyImage)
	// Only objects typed as images are hashed, as in the reference.
	if name, err := matchObject(t, rule(ih.hash), string(data), anyType); name != "" || err != nil {
		t.Fatalf("untyped object: %q, %v", name, err)
	}
}
