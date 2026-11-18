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

func pngIconProbe(t *testing.T, size int, alpha bool) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			a := byte(255)
			if alpha {
				a = byte(x*13 + y*29)
			} else if x == 0 && y == 0 {
				// Force Go's encoder to retain RGBA instead of optimizing
				// an all-opaque image to RGB; the reference ICO decoder
				// requires RGBA and ignores alpha for the fuzzy hash.
				a = 254
			}
			img.SetNRGBA(x, y, color.NRGBA{byte(x*17 + y*3), byte(x*5 + y*11), byte(x*19 + y*7), a})
		}
	}
	var payload bytes.Buffer
	if err := png.Encode(&payload, img); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 22, 22+payload.Len())
	data[2], data[4] = 1, 1
	data[6], data[7] = byte(size), byte(size) // zero denotes 256.
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], 32)
	binary.LittleEndian.PutUint32(data[14:18], uint32(payload.Len()))
	binary.LittleEndian.PutUint32(data[18:22], 22)
	return append(data, payload.Bytes()...)
}

func TestSinglePNGIconPinnedReferenceHashes(t *testing.T) {
	// Pinned ClamAV 1.5.4 sigtool --fuzzy-img output for independently
	// generated RGBA PNG-backed ICOs with these exact pixel formulas.
	for _, tc := range []struct {
		size int
		hash uint64
	}{
		{16, 0x9535ae2d622aaf2a},
		{32, 0x87852e6dc81eb5a5},
		{64, 0xcb10ca1a421fb7e7},
		{128, 0xd3ff535053aa0255},
		{256, 0xfdaa77aafdaa0000},
	} {
		for _, alpha := range []bool{false, true} {
			if alpha && tc.size > 64 {
				continue // reference probe covered transparent 16–64.
			}
			t.Run(fmt.Sprintf("%d-alpha-%t", tc.size, alpha), func(t *testing.T) {
				data := pngIconProbe(t, tc.size, alpha)
				got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !got.ok || got.uncertain != 0 || got.hash != tc.hash {
					t.Fatalf("hash: %+v, want %016x", got, tc.hash)
				}
			})
		}
	}
	data := pngIconProbe(t, 32, true)
	db, err := Compile(map[string][]byte{
		"probe.ftm": []byte("0:0:00000100:ICO:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("Icon.Probe;Engine:150-255,Target:0;0;fuzzy_img#87852e6dc81eb5a5\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if typ, complete := db.DetermineType(data); !complete || typ != TypeGraphics {
		t.Fatalf("typed ICO: %s, complete=%t", typ, complete)
	}
	if name, err := matchObject(t, db, string(data), MatchContext{Type: TypeGraphics}); name != "Icon.Probe" || err != nil {
		t.Fatalf("matching ICO: %q, %v", name, err)
	}
}

func TestSinglePNGIconRejectsUnprovedLayouts(t *testing.T) {
	good := pngIconProbe(t, 32, false)
	rule := compileTest(t, "ldb", "Icon.Probe;Engine:150-255,Target:0;0;fuzzy_img#87852e6dc81eb5a5\n")
	oversized := bytes.Clone(good)
	binary.LittleEndian.PutUint32(oversized[14:18], uint32(len(good)))
	referenceICODirectory(t, oversized, 0x87852e6dc81eb5a5)
	expectOutcome(t, rule, string(oversized), MatchContext{Type: TypeGraphics}, "Icon.Probe", "")
	mutate := func(change func([]byte)) []byte {
		bad := bytes.Clone(good)
		change(bad)
		return bad
	}
	for name, data := range map[string][]byte{
		"second frame":   mutate(func(b []byte) { b[4] = 2 }),
		"wrong offset":   mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[18:22], 23) }),
		"wrong width":    mutate(func(b []byte) { b[6] = 31 }),
		"wrong bitdepth": mutate(func(b []byte) { b[22+24] = 16 }),
		"interlaced":     mutate(func(b []byte) { b[22+28] = 1 }),
		"DIB resource":   mutate(func(b []byte) { copy(b[22:], []byte("DIB!")) }),
		"truncated":      good[:len(good)-1],
	} {
		t.Run(name, func(t *testing.T) {
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unproved ICO layout hashed: %+v", got)
			}
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}

// Optional differential over the exact Go-generated ICO bytes used above.
func TestSinglePNGIconPinnedReferenceProbe(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned 1.5.4 sigtool")
	}
	tool, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("unpinned sigtool: %v", err)
	}
	root := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"), "CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	for _, size := range []int{16, 32, 64, 128, 256} {
		for _, alpha := range []bool{false, true} {
			data := pngIconProbe(t, size, alpha)
			path := filepath.Join(t.TempDir(), fmt.Sprintf("icon-%d-alpha-%t.ico", size, alpha))
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(sigtool, "--fuzzy-img", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("reference %s: %v: %s", path, err, out)
			}
			_, value, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
			if !ok {
				t.Fatalf("unexpected sigtool output: %s", out)
			}
			want, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || (got.hash^want)&^got.uncertain != 0 {
				t.Fatalf("%s: Go %+v, reference %016x", path, got, want)
			}
		}
	}
}
