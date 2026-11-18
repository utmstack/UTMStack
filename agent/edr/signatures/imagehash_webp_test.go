package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
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
)

func TestWebPFuzzyImageCustomGraphicsType(t *testing.T) {
	cases := []struct {
		name, file, sha string
		hash            uint64
	}{
		{"lossless", "webp-lossless.webp", "a699f6093ace15241bcf4f95b016aa35b71b296fa8ee2d7718beb632c46082d2", 0x95cc150b9649973f},
		{"alpha-lossless", "webp-alpha.webp", "b2ce253e9ab3af2e4ea997cc468f8ee78479e1614130001659f2ce7d45b80a93", 0x95cc150b9649973f},
		{"extended-opaque", "webp-extended-opaque.webp", "205eda8bc35450404b9478701cb1b37e8d40ecd98f887d39ac48c32ac1e1d6db", 0x95cc150b9649973f},
		{"extended-alpha", "webp-extended-alpha.webp", "3e221b99a4e0f0ee7169098ab6b8be6c09942de0a548e9461f00958f9b920573", 0x95cc150b9649973f},
		{"lossy", "webp-lossy.webp", "8c6fea9f877c63213a3cf540d9bfb706c5c4e772f38a92de5c1dd2078e4f2ebe", 0x954c150b9e49973f},
		{"extended-lossy", "webp-extended-lossy.webp", "0644abb5f110ad366eec2c4878a9cc2a4c1226e3dfb33670bc13306adf366f06", 0x954c150b9e49973f},
		{"alpha-lossy", "webp-alpha-lossy.webp", "a77ee0ba511ae6f6ec0971bd6687a9d84b74938bbaa277709fdfcb888cffabbd", 0x95cc150b9649973f},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != tc.sha {
				t.Fatalf("fixture SHA-256 %s, want %s", got, tc.sha)
			}
			if tc.name != "lossy" && tc.name != "extended-lossy" && tc.name != "alpha-lossy" && !webpLosslessProfile(data) {
				t.Fatal("validated lossless profile not recognized")
			}
			if (tc.name == "lossy" || tc.name == "extended-lossy" || tc.name == "alpha-lossy") && !webpLossyProfile(data) {
				t.Fatal("validated lossy profile not recognized")
			}
			h := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !h.ok || h.hash != tc.hash || h.uncertain != 0 {
				t.Fatalf("hash %+v, reference %016x", h, tc.hash)
			}
			custom, err := Compile(map[string][]byte{
				"probe.ftm": []byte("0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
				"probe.ldb": []byte(fmt.Sprintf("Image.WebP;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash)),
			})
			if err != nil {
				t.Fatal(err)
			}
			if typ, complete := custom.DetermineType(data); !complete || typ != TypeGraphics {
				t.Fatalf("custom type %s, complete=%t", typ, complete)
			}
			name, matchErr := matchObject(t, custom, string(data), MatchContext{Type: TypeGraphics})
			if name != "Image.WebP" || matchErr != nil {
				t.Fatalf("custom match %q, %v", name, matchErr)
			}
			// The official file-type feed gives RIFF containers a non-graphics
			// type, which must not make an image rule fire on that object.
			official, err := Compile(map[string][]byte{
				"probe.ftm": []byte("0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n"),
				"probe.ldb": []byte(fmt.Sprintf("Image.WebP;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash)),
			})
			if err != nil {
				t.Fatal(err)
			}
			if typ, complete := official.DetermineType(data); !complete || typ != TypeRIFF {
				t.Fatalf("official type %s, complete=%t", typ, complete)
			}
			name, matchErr = matchObject(t, official, string(data), MatchContext{Type: TypeRIFF})
			if name != "" || matchErr != nil {
				t.Fatalf("official RIFF match %q, %v", name, matchErr)
			}
		})
	}
}

func TestWebPTinyLossyOwnedResizeHash(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "webp-lossy-tiny-10x11.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "341d2bd65f9f4a2ee941bb60ac19ddc3187703d21262e39e62979a19221bff9a" {
		t.Fatalf("fixture SHA-256 %s", got)
	}
	if !webpLossyProfile(data) {
		t.Fatal("bounded tiny VP8 profile rejected")
	}
	h := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	const reference = uint64(0x875717525346f8a9)
	if !h.ok || h.hash != reference || h.uncertain != (reference^nativeWindowsImageHash(t, data)) {
		t.Fatalf("tiny hash %+v does not safely cover pinned reference %016x", h, reference)
	}
	for _, size := range [][2]uint16{{0, 16}, {16, 0}, {1, 33}, {33, 1}, {15, 512}, {512, 15}, {513, 32}, {32, 513}} {
		body := []byte{0, 0, 0, 0x9d, 0x01, 0x2a, byte(size[0]), byte(size[0] >> 8), byte(size[1]), byte(size[1] >> 8)}
		if _, _, admitted := vp8StillDimensions(body); admitted {
			t.Fatalf("unproved shape %dx%d admitted", size[0], size[1])
		}
	}
}

func TestWebPLosslessContainerBounds(t *testing.T) {
	valid, err := os.ReadFile(filepath.Join("testdata", "webp-lossless.webp"))
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated":      func(b []byte) []byte { return b[:len(b)-1] },
		"extra-trailing": func(b []byte) []byte { return append(b, 0, 0) },
		"short-riff-size": func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[4:8], binary.LittleEndian.Uint32(b[4:8])-1)
			return b
		},
		"short-chunk-size": func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[16:20], binary.LittleEndian.Uint32(b[16:20])-1)
			return b
		},
		"nonzero-padding": func(b []byte) []byte { b[len(b)-1] = 1; return b },
	} {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			if simpleWebPLossless(data) {
				t.Fatal("malformed container accepted as exact profile")
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("malformed WebP got exact hash %+v", got)
			}
		})
	}
}

func TestWebPLossyContainerBounds(t *testing.T) {
	valid, err := os.ReadFile(filepath.Join("testdata", "webp-lossy.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if !simpleWebPLossy(valid) {
		t.Fatal("golden VP8 file rejected")
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated":   func(b []byte) []byte { return b[:len(b)-1] },
		"trailing":    func(b []byte) []byte { return append(b, 0) },
		"bad-sync":    func(b []byte) []byte { b[23] = 0; return b },
		"inter-frame": func(b []byte) []byte { b[20] |= 1; return b },
		"tiny-width":  func(b []byte) []byte { b[26], b[27] = 1, 0; return b },
		"large-width": func(b []byte) []byte { b[26], b[27] = 1, 2; return b },
		"short-chunk": func(b []byte) []byte { binary.LittleEndian.PutUint32(b[16:20], 9); return b },
		"huge-chunk":  func(b []byte) []byte { binary.LittleEndian.PutUint32(b[16:20], ^uint32(0)); return b },
		"extra-chunk": func(b []byte) []byte {
			b = append(b, []byte("XMP \x00\x00\x00\x00")...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
			return b
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			if webpLossyProfile(data) {
				t.Fatal("unsupported VP8 container entered exact profile")
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unsupported VP8 got exact hash %+v", got)
			}
		})
	}
}

func TestWebPExtendedLossyContainerBounds(t *testing.T) {
	valid, err := os.ReadFile(filepath.Join("testdata", "webp-extended-lossy.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if !extendedWebPLossy(valid) {
		t.Fatal("golden extended VP8 file rejected")
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated":      func(b []byte) []byte { return b[:len(b)-1] },
		"trailing":       func(b []byte) []byte { return append(b, 0) },
		"animated":       func(b []byte) []byte { b[20] |= 0x02; return b },
		"alpha-flag":     func(b []byte) []byte { b[20] |= 0x10; return b },
		"missing-xmp":    func(b []byte) []byte { b[20] &^= 0x04; return b },
		"wrong-canvas":   func(b []byte) []byte { b[24]++; return b },
		"huge-vp8-chunk": func(b []byte) []byte { binary.LittleEndian.PutUint32(b[34:38], ^uint32(0)); return b },
		"duplicate-xmp": func(b []byte) []byte {
			b = append(b, b[2024:]...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
			return b
		},
		"unknown-chunk": func(b []byte) []byte {
			b = append(b, []byte("FOO \x00\x00\x00\x00")...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
			return b
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			if webpLossyProfile(data) {
				t.Fatal("unsupported extended VP8 entered exact profile")
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unsupported extended VP8 got exact hash %+v", got)
			}
		})
	}
}

func TestWebPAlphaLossyContainerBounds(t *testing.T) {
	valid, err := os.ReadFile(filepath.Join("testdata", "webp-alpha-lossy.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if !extendedWebPAlphaLossy(valid) {
		t.Fatal("golden ALPH/VP8 file rejected")
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated":     func(b []byte) []byte { return b[:len(b)-1] },
		"trailing":      func(b []byte) []byte { return append(b, 0) },
		"alpha-not-set": func(b []byte) []byte { b[20] &^= 0x10; return b },
		"animated":      func(b []byte) []byte { b[20] |= 0x02; return b },
		"no-alph":       func(b []byte) []byte { b[30] = 'X'; return b },
		"bad-alph-mode": func(b []byte) []byte { b[38] = 0xe2; return b },
		"wrong-canvas":  func(b []byte) []byte { b[24]++; return b },
		"huge-alph":     func(b []byte) []byte { binary.LittleEndian.PutUint32(b[34:38], ^uint32(0)); return b },
	} {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			if webpLossyProfile(data) {
				t.Fatal("unsupported ALPH/VP8 entered exact profile")
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unsupported ALPH/VP8 got exact hash %+v", got)
			}
		})
	}
	// A structurally valid ALPH chunk with a truncated compressed payload
	// must still fail the full-image decode before VP8-only hashing.
	short := append(append([]byte(nil), valid[:39]...), valid[61:]...)
	binary.LittleEndian.PutUint32(short[4:8], uint32(len(short)-8))
	binary.LittleEndian.PutUint32(short[34:38], 1)
	if !extendedWebPAlphaLossy(short) {
		t.Fatal("short ALPH control is not structurally valid")
	}
	if got := fuzzyImageHash(short, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
		t.Fatalf("truncated ALPH payload got exact hash %+v", got)
	}
}

func TestWebPExtendedContainerBounds(t *testing.T) {
	valid, err := os.ReadFile(filepath.Join("testdata", "webp-extended-alpha.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if !extendedWebPLossless(valid) {
		t.Fatal("golden extended file rejected")
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated":        func(b []byte) []byte { return b[:len(b)-1] },
		"trailing":         func(b []byte) []byte { return append(b, 0) },
		"animated":         func(b []byte) []byte { b[20] |= 0x02; return b },
		"missing-icc-flag": func(b []byte) []byte { b[20] &^= 0x20; return b },
		"missing-xmp-flag": func(b []byte) []byte { b[20] &^= 0x04; return b },
		"wrong-canvas":     func(b []byte) []byte { b[24]++; return b },
		"wrong-alpha-flag": func(b []byte) []byte { b[20] &^= 0x10; return b },
		"bad-icc-padding":  func(b []byte) []byte { b[65] = 1; return b },
		"oversize-vp8l-chunk": func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[70:74], ^uint32(0))
			return b
		},
		"duplicate-xmp": func(b []byte) []byte {
			b = append(b, b[150:]...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
			return b
		},
		"unknown-chunk": func(b []byte) []byte {
			b = append(b, []byte("FOO \x00\x00\x00\x00")...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
			return b
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			if extendedWebPLossless(data) || webpLosslessProfile(data) {
				t.Fatal("unsupported container entered exact profile")
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unsupported container got exact hash %+v", got)
			}
		})
	}
}

// This optional differential checks the actual image decoder and hash path,
// including cases where lossy VP8's YUV conversion may differ by one pixel.
func TestWebPSigtoolDifferential(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	cwebp := "/opt/homebrew/bin/cwebp"
	encoder, err := os.ReadFile(cwebp)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(encoder)) != "46f37f040f9f248ab5225a0dea0d92b697d08bdb712c8677f5cea922bde3b8e7" {
		t.Skipf("pinned cwebp unavailable: %v", err)
	}
	webpmux := "/opt/homebrew/bin/webpmux"
	muxBytes, err := os.ReadFile(webpmux)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(muxBytes)) != "d61e710d9dbe6a4898d6012c1ac0847452a385390ac84c2d93064b920b77de89" {
		t.Skipf("pinned webpmux unavailable: %v", err)
	}
	root := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	dir := t.TempDir()
	xmp := filepath.Join(dir, "probe.xmp")
	if err := os.WriteFile(xmp, []byte("<x:xmpmeta xmlns:x=\"adobe:ns:meta/\"><test>harmless</test></x:xmpmeta>"), 0o600); err != nil {
		t.Fatal(err)
	}
	count, extendedCount, lossyCount, alphaSimple, alphaLossyCount := 0, 0, 0, 0, 0
	for _, shape := range [][2]int{{17, 17}, {31, 33}, {32, 32}, {63, 64}, {64, 47}, {128, 128}, {257, 33}, {257, 511}, {512, 289}, {37, 256}} {
		for seed := 0; seed < 4; seed++ {
			img := image.NewNRGBA(image.Rect(0, 0, shape[0], shape[1]))
			for y := 0; y < shape[1]; y++ {
				for x := 0; x < shape[0]; x++ {
					img.SetNRGBA(x, y, color.NRGBA{R: byte(x*13 + y*7 + seed*19), G: byte(x*3 + y*19 + seed*17), B: byte(x*23 + y*5 + seed*23), A: 255})
				}
			}
			var src bytes.Buffer
			if err := png.Encode(&src, img); err != nil {
				t.Fatal(err)
			}
			input := filepath.Join(dir, "source.png")
			if err := os.WriteFile(input, src.Bytes(), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"lossless", "lossy", "alpha-lossless", "alpha-exact-lossless", "alpha-lossy"} {
				if strings.HasPrefix(mode, "alpha-") {
					for y := 0; y < shape[1]; y++ {
						for x := 0; x < shape[0]; x++ {
							p := img.NRGBAAt(x, y)
							p.A = byte((x*7 + y*17 + seed*47) & 255)
							img.SetNRGBA(x, y, p)
						}
					}
					src.Reset()
					if err := png.Encode(&src, img); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(input, src.Bytes(), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				output := filepath.Join(dir, fmt.Sprintf("probe-%dx%d-%d-%s.webp", shape[0], shape[1], seed, mode))
				args := []string{"-quiet"}
				if strings.Contains(mode, "lossless") {
					args = append(args, "-lossless")
					if mode == "alpha-exact-lossless" {
						args = append(args, "-exact")
					}
				} else {
					args = append(args, "-q", strconv.Itoa([]int{15, 50, 87, 100}[seed]))
				}
				args = append(args, input, "-o", output)
				if out, err := exec.Command(cwebp, args...).CombinedOutput(); err != nil {
					t.Fatalf("cwebp: %v, %s", err, out)
				}
				cmd := exec.Command(sigtool, "--fuzzy-img", output)
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("sigtool: %v, %s", err, out)
				}
				_, value, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
				if !ok {
					t.Fatalf("sigtool output: %s", out)
				}
				want, err := strconv.ParseUint(value, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(output)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(mode, "alpha-") && len(data) >= 25 && bytes.Equal(data[12:16], []byte("VP8L")) && data[24]&0x10 != 0 {
					alphaSimple++
				}
				got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !got.ok || got.hash != want || got.uncertain != 0 {
					t.Errorf("%s: Go %+v, reference %016x", filepath.Base(output), got, want)
				}
				if mode == "alpha-lossy" {
					alphaLossyCount++
				}
				if mode == "lossy" {
					lossyCount++
				}
				extended := output + ".extended.webp"
				if out, err := exec.Command(webpmux, "-set", "xmp", xmp, output, "-o", extended).CombinedOutput(); err != nil {
					t.Fatalf("webpmux: %v, %s", err, out)
				}
				cmd = exec.Command(sigtool, "--fuzzy-img", extended)
				cmd.Env = env
				out, err = cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("sigtool extended: %v, %s", err, out)
				}
				_, value, ok = strings.Cut(strings.TrimSpace(string(out)), ": ")
				if !ok {
					t.Fatalf("sigtool extended output: %s", out)
				}
				extendedHash, err := strconv.ParseUint(value, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				b, err := os.ReadFile(extended)
				if err != nil {
					t.Fatal(err)
				}
				h := fuzzyImageHash(b, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !h.ok || h.hash != extendedHash || h.uncertain != 0 || h.hash != want {
					t.Errorf("%s: Go %+v, reference %016x, base %016x", filepath.Base(extended), h, extendedHash, want)
				}
				extendedCount++
				count++
			}
		}
	}
	t.Logf("WebP differential: %d base and %d extended exact hashes, including %d opaque lossy VP8, %d alpha VP8L and %d alpha VP8", count, extendedCount, lossyCount, alphaSimple, alphaLossyCount)
}

func TestWebPClamAVTypingReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	clamscan := filepath.Join(root, "bin", "clamscan")
	toolBytes, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("clamscan is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	for _, tc := range []struct {
		name, fixture, hash, typeRule, expected string
	}{
		{"custom-lossless", "webp-lossless.webp", "95cc150b9649973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"custom-alpha", "webp-alpha.webp", "95cc150b9649973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"custom-extended", "webp-extended-alpha.webp", "95cc150b9649973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"custom-lossy", "webp-lossy.webp", "954c150b9e49973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"custom-extended-lossy", "webp-extended-lossy.webp", "954c150b9e49973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"custom-alpha-lossy", "webp-alpha-lossy.webp", "95cc150b9649973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "FOUND"},
		{"official-lossless", "webp-lossless.webp", "95cc150b9649973f", "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n", "OK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte(tc.typeRule), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte("Image.WebP;Engine:150-255,Target:0;0;fuzzy_img#"+tc.hash+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			fixture, err := os.ReadFile(filepath.Join("testdata", tc.fixture))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, tc.fixture)
			if err := os.WriteFile(path, fixture, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(clamscan, "--database="+dir, "--no-summary", path)
			cmd.Env = env
			out, runErr := cmd.CombinedOutput()
			if !strings.Contains(string(out), path+": ") || !strings.Contains(string(out), tc.expected) || tc.expected == "FOUND" && runErr == nil || tc.expected == "OK" && runErr != nil {
				t.Fatalf("reference scan: %v, %s", runErr, out)
			}
		})
	}
}
