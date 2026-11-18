package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func dibIconProbe(t *testing.T, size, depth int, alpha bool, mask byte) []byte {
	t.Helper()
	stride := (size*depth + 31) / 32 * 4
	maskBytes := (size + 31) / 32 * 4 * size
	payload := make([]byte, 40+stride*size+maskBytes)
	binary.LittleEndian.PutUint32(payload[:4], 40)
	binary.LittleEndian.PutUint32(payload[4:8], uint32(size))
	binary.LittleEndian.PutUint32(payload[8:12], uint32(2*size))
	binary.LittleEndian.PutUint16(payload[12:14], 1)
	binary.LittleEndian.PutUint16(payload[14:16], uint16(depth))
	binary.LittleEndian.PutUint32(payload[20:24], uint32(stride*size+maskBytes))
	for y := 0; y < size; y++ {
		row := payload[40+(size-1-y)*stride:]
		for x := 0; x < size; x++ {
			r, g, b := byte(x*17+y*3), byte(x*5+y*11), byte(x*19+y*7)
			if depth == 24 {
				copy(row[3*x:], []byte{b, g, r})
			} else {
				a := byte(255)
				if alpha {
					a = byte(x*13 + y*29)
				} else if x == 0 && y == 0 {
					a = 254
				}
				copy(row[4*x:], []byte{b, g, r, a})
			}
		}
	}
	for i := len(payload) - maskBytes; i < len(payload); i++ {
		payload[i] = mask
	}
	data := make([]byte, 22, 22+len(payload))
	data[2], data[4] = 1, 1
	data[6], data[7] = byte(size), byte(size) // zero denotes 256.
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], uint16(depth))
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	binary.LittleEndian.PutUint32(data[18:22], 22)
	return append(data, payload...)
}

func indexedDIBIconProbe(t *testing.T, size, depth int) []byte {
	t.Helper()
	colors := 1 << depth
	stride := (size*depth + 31) / 32 * 4
	maskBytes := (size + 31) / 32 * 4 * size
	payload := make([]byte, 40+4*colors+stride*size+maskBytes)
	binary.LittleEndian.PutUint32(payload[:4], 40)
	binary.LittleEndian.PutUint32(payload[4:8], uint32(size))
	binary.LittleEndian.PutUint32(payload[8:12], uint32(2*size))
	binary.LittleEndian.PutUint16(payload[12:14], 1)
	binary.LittleEndian.PutUint16(payload[14:16], uint16(depth))
	binary.LittleEndian.PutUint32(payload[20:24], uint32(stride*size+maskBytes))
	binary.LittleEndian.PutUint32(payload[32:36], uint32(colors))
	for i := 0; i < colors; i++ {
		copy(payload[40+4*i:], []byte{byte(i * 7), byte(255 - i), byte(i), 0})
	}
	for y := 0; y < size; y++ {
		row := payload[40+4*colors+(size-1-y)*stride:]
		for x := 0; x < size; x++ {
			index := (x*17 + y*3) & (colors - 1)
			bit := x * depth
			row[bit/8] |= byte(index << (8 - depth - bit%8))
		}
	}
	data := make([]byte, 22, 22+len(payload))
	data[2], data[4] = 1, 1
	data[6], data[7], data[8] = byte(size), byte(size), byte(colors)
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], uint16(depth))
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	binary.LittleEndian.PutUint32(data[18:22], 22)
	return append(data, payload...)
}

func extendedDIBIconProbe(t *testing.T, original []byte, header int) []byte {
	t.Helper()
	if header != 52 && header != 56 && header != 108 && header != 124 {
		t.Fatal("unsupported probe header")
	}
	old := original[22:]
	payload := make([]byte, len(old)+header-40)
	copy(payload[:40], old[:40])
	copy(payload[header:], old[40:])
	binary.LittleEndian.PutUint32(payload[:4], uint32(header))
	depth := binary.LittleEndian.Uint16(payload[14:16])
	if depth == 24 || depth == 32 {
		for i, mask := range []uint32{0x00ff0000, 0x0000ff00, 0x000000ff} {
			binary.LittleEndian.PutUint32(payload[40+4*i:], mask)
		}
		if header >= 56 {
			binary.LittleEndian.PutUint32(payload[52:56], 0xff000000)
		}
	}
	data := append(bytes.Clone(original[:22]), payload...)
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	return data
}

func bitfieldDIBICOProbe(t *testing.T, size, depth int, bitfields bool) []byte {
	t.Helper()
	if depth != 16 && depth != 32 {
		t.Fatal("unsupported probe depth")
	}
	stride := (size*depth + 31) / 32 * 4
	maskBytes := (size + 31) / 32 * 4 * size
	extra := 0
	if bitfields {
		extra = 12
	}
	payload := make([]byte, 40+extra+stride*size+maskBytes)
	binary.LittleEndian.PutUint32(payload[:4], 40)
	binary.LittleEndian.PutUint32(payload[4:8], uint32(size))
	binary.LittleEndian.PutUint32(payload[8:12], uint32(2*size))
	binary.LittleEndian.PutUint16(payload[12:14], 1)
	binary.LittleEndian.PutUint16(payload[14:16], uint16(depth))
	if bitfields {
		binary.LittleEndian.PutUint32(payload[16:20], 3)
		masks := []uint32{0xf800, 0x07e0, 0x001f}
		if depth == 32 {
			masks = []uint32{0x00ff0000, 0x0000ff00, 0x000000ff}
		}
		for i, mask := range masks {
			binary.LittleEndian.PutUint32(payload[40+4*i:], mask)
		}
	}
	binary.LittleEndian.PutUint32(payload[20:24], uint32(stride*size))
	for y := 0; y < size; y++ {
		row := payload[40+extra+y*stride:]
		for x := 0; x < size; x++ {
			r, g, b := byte(x*11+y*3), byte(x*5+y*17), byte(x*19+y*7)
			if depth == 32 {
				copy(row[4*x:], []byte{b, g, r, 255})
			} else {
				binary.LittleEndian.PutUint16(row[2*x:], uint16(r>>3)<<11|uint16(g>>2)<<5|uint16(b>>3))
			}
		}
	}
	data := make([]byte, 22, 22+len(payload))
	data[2], data[4] = 1, 1
	data[6], data[7] = byte(size), byte(size)
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], uint16(depth))
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	binary.LittleEndian.PutUint32(data[18:22], 22)
	return append(data, payload...)
}

func TestBitfieldDIBICOReferenceHashes(t *testing.T) {
	for _, tc := range []struct {
		depth, size int
		bitfields   bool
		want        uint64
	}{
		{16, 32, false, 0x98fc16b316b512b5},
		{16, 32, true, 0x88ec267b3e95c0b5},
		{32, 32, true, 0x88ec267b3e95c0b5},
	} {
		name := fmt.Sprintf("%d-bit-fields-%t", tc.depth, tc.bitfields)
		t.Run(name, func(t *testing.T) {
			data := bitfieldDIBICOProbe(t, tc.size, tc.depth, tc.bitfields)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || (got.hash^tc.want)&^got.uncertain != 0 {
				t.Fatalf("Go hash %+v, pinned reference %016x", got, tc.want)
			}
		})
	}
	rule := compileTest(t, "ldb", "Icon.Bitfields;Engine:150-255,Target:0;0;fuzzy_img#88ec267b3e95c0b5\n")
	expectOutcome(t, rule, string(bitfieldDIBICOProbe(t, 32, 16, true)),
		MatchContext{Type: TypeGraphics}, "Icon.Bitfields", "")
	bad := bitfieldDIBICOProbe(t, 32, 16, true)
	bad[22+40] ^= 1 // A noncanonical bit mask is outside the proved profile.
	expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
}

func TestBitfieldDIBICOPinnedReferenceProbe(t *testing.T) {
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
	for _, depth := range []int{16, 32} {
		for _, fields := range []bool{false, true} {
			for _, size := range []int{16, 32, 64} {
				t.Run(fmt.Sprintf("%d-bit-fields-%t-%d", depth, fields, size), func(t *testing.T) {
					data := bitfieldDIBICOProbe(t, size, depth, fields)
					path := filepath.Join(t.TempDir(), "bitfields.ico")
					if err := os.WriteFile(path, data, 0600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(sigtool, "--fuzzy-img", path)
					cmd.Env = env
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("reference: %v: %s", err, out)
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
						t.Fatalf("Go %+v, reference %016x", got, want)
					}
					t.Logf("Go %016x uncertain %016x reference %016x", got.hash, got.uncertain, want)
				})
			}
		}
	}
}

func TestExtendedDIBICOReferenceHashes(t *testing.T) {
	for _, tc := range []struct {
		depth int
		hash  uint64
	}{
		{1, 0x8000000000000000}, {2, 0xfa02fa0aeaaaaaaa},
		{4, 0xbc86a887aa9735e2}, {8, 0xead215ca05e57655},
		{24, 0x9535ae2d622aaf2a}, {32, 0x9535ae2d622aaf2a},
	} {
		var base []byte
		if tc.depth <= 8 {
			base = indexedDIBIconProbe(t, 16, tc.depth)
		} else {
			base = dibIconProbe(t, 16, tc.depth, false, 0)
		}
		for _, header := range []int{52, 56, 108, 124} {
			t.Run(fmt.Sprintf("%d-bit-%d-header", tc.depth, header), func(t *testing.T) {
				data := extendedDIBIconProbe(t, base, header)
				got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !got.ok || (got.hash^tc.hash)&^got.uncertain != 0 {
					t.Fatalf("extended ICO hash %+v, pinned reference %016x", got, tc.hash)
				}
			})
		}
	}
	rule := compileTest(t, "ldb", "Icon.V5;Engine:150-255,Target:0;0;fuzzy_img#9535ae2d622aaf2a\n")
	expectOutcome(t, rule, string(extendedDIBIconProbe(t, dibIconProbe(t, 16, 32, false, 0), 124)),
		MatchContext{Type: TypeGraphics}, "Icon.V5", "")
	base := dibIconProbe(t, 32, 32, false, 0)
	data := multiICOProbe(t, base, extendedDIBIconProbe(t, base, 124))
	got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	if !got.ok || got.hash != 0x87852e6dc81eb5a5 || got.uncertain != 0 {
		t.Fatalf("multi-frame V5 ICO hash: %+v", got)
	}
}

func TestExtendedDIBICOPinnedReferenceProbe(t *testing.T) {
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
	for _, depth := range []int{1, 2, 4, 8, 24, 32} {
		var base []byte
		if depth <= 8 {
			base = indexedDIBIconProbe(t, 16, depth)
		} else {
			base = dibIconProbe(t, 16, depth, false, 0)
		}
		for _, header := range []int{52, 56, 108, 124} {
			t.Run(fmt.Sprintf("%d-bit-%d-header", depth, header), func(t *testing.T) {
				data := extendedDIBIconProbe(t, base, header)
				path := filepath.Join(t.TempDir(), "extended.ico")
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(sigtool, "--fuzzy-img", path)
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("reference: %v: %s", err, out)
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
					t.Fatalf("Go %+v, reference %016x", got, want)
				}
				t.Logf("Go %016x uncertain %016x reference %016x", got.hash, got.uncertain, want)
			})
		}
	}
}

func TestExtendedDIBICOIgnoresAdvisoryMetadata(t *testing.T) {
	base := extendedDIBIconProbe(t, dibIconProbe(t, 32, 32, false, 0), 108)
	rule := compileTest(t, "ldb", "Icon.V4;Engine:150-255,Target:0;0;fuzzy_img#87852e6dc81eb5a5\n")
	for name, at := range map[string]int{"noncanonical mask": 22 + 40, "color space": 22 + 56, "gamma": 22 + 96} {
		t.Run(name, func(t *testing.T) {
			bad := bytes.Clone(base)
			bad[at] ^= 1
			referenceICODirectory(t, bad, 0x87852e6dc81eb5a5)
			expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "Icon.V4", "")
		})
	}
}

func invertedDIBIconProbe(t *testing.T, size, depth int) []byte {
	t.Helper()
	data := dibIconProbe(t, size, depth, false, 0)
	stride := (size*depth + 31) / 32 * 4
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			at := 22 + 40 + y*stride + x*depth/8
			for c := 0; c < 3; c++ {
				data[at+c] ^= 255
			}
		}
	}
	return data
}

func multiICOProbe(t *testing.T, frames ...[]byte) []byte {
	t.Helper()
	data := make([]byte, 6+16*len(frames))
	data[2] = 1
	binary.LittleEndian.PutUint16(data[4:6], uint16(len(frames)))
	for i, frame := range frames {
		if len(frame) < 22 {
			t.Fatal("invalid frame probe")
		}
		entry := data[6+16*i : 6+16*(i+1)]
		copy(entry, frame[6:22])
		binary.LittleEndian.PutUint32(entry[8:12], uint32(len(frame)-22))
		binary.LittleEndian.PutUint32(entry[12:16], uint32(len(data)))
		data = append(data, frame[22:]...)
	}
	return data
}

func TestMultiICOReferenceFrameSelection(t *testing.T) {
	normal16_24 := dibIconProbe(t, 16, 24, false, 0)
	normal16_32 := dibIconProbe(t, 16, 32, false, 0)
	normal32_32 := dibIconProbe(t, 32, 32, false, 0)
	normal64_24 := dibIconProbe(t, 64, 24, false, 0)
	inverted32_24 := invertedDIBIconProbe(t, 32, 24)
	inverted32_32 := invertedDIBIconProbe(t, 32, 32)
	png16 := pngIconProbe(t, 16, false)
	for _, tc := range []struct {
		name   string
		frames [][]byte
		want   uint64
	}{
		{"area", [][]byte{normal16_24, inverted32_24}, 0xf87ad19237e04a5a},
		{"depth-before-area", [][]byte{normal64_24, normal16_32}, 0x9535ae2d622aaf2a},
		{"tie-last-inverted", [][]byte{normal32_32, inverted32_32}, 0xf87ad19237e04a5a},
		{"tie-last-normal", [][]byte{inverted32_32, normal32_32}, 0x87852e6dc81eb5a5},
		{"png-depth", [][]byte{normal64_24, png16}, 0x9535ae2d622aaf2a},
		{"three-frames", [][]byte{normal16_32, inverted32_32, normal64_24}, 0xf87ad19237e04a5a},
		{"indexed-low-depth", [][]byte{indexedDIBIconProbe(t, 256, 8), normal16_24}, 0x9535ae2d622aaf2a},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := multiICOProbe(t, tc.frames...)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.hash != tc.want || got.uncertain != 0 {
				t.Fatalf("multi-frame hash: %+v, reference %016x", got, tc.want)
			}
		})
	}
	db := compileTest(t, "ldb", "Icon.Multi;Engine:150-255,Target:0;0;fuzzy_img#f87ad19237e04a5a\n")
	data := multiICOProbe(t, normal16_24, inverted32_24)
	expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "Icon.Multi", "")
	bad := bytes.Clone(data)
	binary.LittleEndian.PutUint32(bad[6+16+12:6+16+16], 22) // Alias the directory.
	expectOutcome(t, db, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	for name, malformed := range map[string][]byte{
		"missing tail": data[:len(data)-1],
		"unsupported second": multiICOProbe(t, normal16_24, func() []byte {
			b := bytes.Clone(inverted32_24)
			binary.LittleEndian.PutUint32(b[22+16:22+20], 1) // Compressed DIB.
			return b
		}()),
		"too many frames": func() []byte {
			b := bytes.Clone(data)
			binary.LittleEndian.PutUint16(b[4:6], 17)
			return b
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			expectOutcome(t, db, string(malformed), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}

func TestMultiICOPinnedReferenceProbe(t *testing.T) {
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
	for _, tc := range []struct {
		name   string
		frames [][]byte
	}{
		{"larger", [][]byte{dibIconProbe(t, 16, 24, false, 0), invertedDIBIconProbe(t, 32, 24)}},
		{"deeper", [][]byte{dibIconProbe(t, 64, 24, false, 0), dibIconProbe(t, 16, 32, false, 0)}},
		{"later-tie", [][]byte{dibIconProbe(t, 32, 32, false, 0), invertedDIBIconProbe(t, 32, 32)}},
		{"png-over-larger-dib", [][]byte{dibIconProbe(t, 64, 24, false, 0), pngIconProbe(t, 16, false)}},
		{"indexed-over-larger", [][]byte{indexedDIBIconProbe(t, 256, 8), dibIconProbe(t, 16, 24, false, 0)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := multiICOProbe(t, tc.frames...)
			path := filepath.Join(t.TempDir(), "multi.ico")
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
				t.Fatalf("multi-frame Go %+v, reference %016x", got, want)
			}
		})
	}
}

func TestSingleIndexedDIBIconPinnedReferenceHashes(t *testing.T) {
	// Each golden value is the SHA-pinned 1.5.4 sigtool result for this
	// generator's exact palette, packed indices and dimensions.
	for _, tc := range []struct {
		size, depth int
		hash        uint64
	}{
		{16, 1, 0x8000000000000000}, {16, 2, 0xfa02fa0aeaaaaaaa},
		{16, 4, 0xbc86a887aa9735e2}, {16, 8, 0xead215ca05e57655},
		{32, 1, 0x8000000000000000}, {32, 2, 0x8000000000000000},
		{32, 4, 0x8084008400040004}, {32, 8, 0xf8f2d7531351d510},
		{64, 1, 0x8000000000000000}, {64, 2, 0x8000000000000000},
		{64, 4, 0xff02bf00ff00ff00}, {64, 8, 0xaaff55b855805598},
		{256, 1, 0x8000000000000000}, {256, 2, 0xaa80aa80aa80aa81},
		{256, 4, 0x8000000000000000}, {256, 8, 0xd555c8550055ffaa},
	} {
		t.Run(fmt.Sprintf("%d-%d", tc.size, tc.depth), func(t *testing.T) {
			data := indexedDIBIconProbe(t, tc.size, tc.depth)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || (got.hash^tc.hash)&^got.uncertain != 0 ||
				(tc.depth == 1 || tc.depth == 8) && (got.hash != tc.hash || got.uncertain != 0) {
				t.Fatalf("indexed DIB certain bits: %+v, reference %016x", got, tc.hash)
			}
		})
	}
}

func TestLowColorBMPReferenceUncertainty(t *testing.T) {
	ico := indexedDIBIconProbe(t, 256, 2)
	// Build a BMP independently of the ICO reader; the appended AND data is
	// irrelevant to the BMP's pixel decoder.
	bmp := append(make([]byte, 14), ico[22:]...)
	copy(bmp, "BM")
	binary.LittleEndian.PutUint32(bmp[2:6], uint32(len(bmp)))
	binary.LittleEndian.PutUint32(bmp[10:14], 54+4*4)
	binary.LittleEndian.PutUint32(bmp[22:26], 256)
	got := fuzzyImageHash(bmp, &matchWork{ctx: context.Background(), limit: 1 << 40})
	const reference = uint64(0xaa80aa80aa80aa81) // Pinned sigtool on both ICO and BMP.
	if !got.ok || (got.hash^reference)&^got.uncertain != 0 {
		t.Fatalf("low-color BMP certain bits: %+v, reference %016x", got, reference)
	}
}

func TestSingleDIBIconPinnedReferenceHashes(t *testing.T) {
	// Golden hashes came from the SHA-pinned ClamAV 1.5.4 sigtool on the
	// generated 24/32-bit DIB ICOs, including alpha and AND-mask variants.
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
		for _, depth := range []int{24, 32} {
			for _, alpha := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d-%d-alpha-%t", tc.size, depth, alpha), func(t *testing.T) {
					data := dibIconProbe(t, tc.size, depth, alpha, 0)
					got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
					if !got.ok || got.uncertain != 0 || got.hash != tc.hash {
						t.Fatalf("DIB hash: %+v, want %016x", got, tc.hash)
					}
				})
			}
		}
	}
	for _, depth := range []int{24, 32} {
		got := fuzzyImageHash(dibIconProbe(t, 32, depth, true, 0xff), &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || got.uncertain != 0 || got.hash != 0x87852e6dc81eb5a5 {
			t.Fatalf("%d-bit AND-mask hash: %+v", depth, got)
		}
	}
	data := dibIconProbe(t, 32, 32, true, 0)
	db, err := Compile(map[string][]byte{
		"probe.ftm": []byte("0:0:00000100:ICO:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("Icon.DIB;Engine:150-255,Target:0;0;fuzzy_img#87852e6dc81eb5a5\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if name, err := matchObject(t, db, string(data), MatchContext{Type: TypeGraphics}); name != "Icon.DIB" || err != nil {
		t.Fatalf("matching DIB ICO: %q, %v", name, err)
	}
}

func TestSingleDIBIconRejectsUnprovedLayouts(t *testing.T) {
	good := dibIconProbe(t, 32, 32, true, 0)
	rule := compileTest(t, "ldb", "Icon.DIB;Engine:150-255,Target:0;0;fuzzy_img#87852e6dc81eb5a5\n")
	for _, at := range []int{42, 54} {
		advisory := bytes.Clone(good)
		binary.LittleEndian.PutUint32(advisory[at:], 1)
		referenceICODirectory(t, advisory, 0x87852e6dc81eb5a5)
		expectOutcome(t, rule, string(advisory), MatchContext{Type: TypeGraphics}, "Icon.DIB", "")
	}
	mutate := func(change func([]byte)) []byte {
		bad := bytes.Clone(good)
		change(bad)
		return bad
	}
	for name, data := range map[string][]byte{
		"wrong offset":     mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[18:22], 23) }),
		"wrong width":      mutate(func(b []byte) { b[6] = 31 }),
		"wrong DIB height": mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[30:34], 32) }),
		"compressed":       mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[38:42], 1) }),
		"invalid depth":    mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[36:38], 3) }),
		"truncated":        good[:len(good)-1],
	} {
		t.Run(name, func(t *testing.T) {
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("unproved DIB layout hashed: %+v", got)
			}
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}

func TestSingleDIBIconPinnedReferenceProbe(t *testing.T) {
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
		for _, depth := range []int{24, 32} {
			for _, alpha := range []bool{false, true} {
				data := dibIconProbe(t, size, depth, alpha, 0)
				path := filepath.Join(t.TempDir(), fmt.Sprintf("dib-%d-%d-alpha-%t.ico", size, depth, alpha))
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
				if !got.ok || got.uncertain != 0 || got.hash != want {
					t.Fatalf("%s: Go %+v, reference %016x", path, got, want)
				}
			}
		}
	}
}

func TestSingleIndexedDIBIconPinnedReferenceProbe(t *testing.T) {
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
	for _, size := range []int{16, 32, 64, 256} {
		for _, depth := range []int{1, 2, 4, 8} {
			data := indexedDIBIconProbe(t, size, depth)
			path := filepath.Join(t.TempDir(), fmt.Sprintf("indexed-%d-%d.ico", size, depth))
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
				t.Fatalf("%s: Go certain bits %+v, reference %016x", path, got, want)
			}
		}
	}
}

func TestIndexedDIBIconRandomizedReference(t *testing.T) {
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
	rng := rand.New(rand.NewSource(20261101))
	count, exact, uncertain := 0, 0, 0
	for _, size := range []int{16, 32, 64, 128, 256} {
		for _, depth := range []int{1, 2, 4, 8} {
			for sample := 0; sample < 3; sample++ {
				data := indexedDIBIconProbe(t, size, depth)
				palette := 22 + 40
				for i := 0; i < 1<<depth; i++ {
					_, _ = rng.Read(data[palette+4*i : palette+4*i+3])
				}
				xorStart := palette + 4*(1<<depth)
				_, _ = rng.Read(data[xorStart:]) // XOR bytes, padding and AND mask.
				path := filepath.Join(t.TempDir(), fmt.Sprintf("indexed-%d-%d-%d.ico", size, depth, sample))
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
					t.Fatalf("%d/%d/%d: Go certain bits %+v, reference %016x", size, depth, sample, got, want)
				}
				count++
				if got.uncertain != 0 {
					uncertain++
				} else if got.hash == want {
					exact++
				}
			}
		}
	}
	t.Logf("indexed ICO differential: %d images, %d exact, %d with uncertain bits", count, exact, uncertain)
}

func TestSingleDIBIconClamAVTypingReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	clamscan := filepath.Join(root, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned clamscan: %v", err)
	}
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	for _, depth := range []int{8, 24, 32} {
		t.Run(fmt.Sprintf("%d-bit", depth), func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte("0:0:00000100:ICO:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"), 0600); err != nil {
				t.Fatal(err)
			}
			hash := "87852e6dc81eb5a5"
			if depth == 8 {
				hash = "f8f2d7531351d510"
			}
			if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte("Icon.DIB;Engine:150-255,Target:0;0;fuzzy_img#"+hash+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "probe.ico")
			data := dibIconProbe(t, 32, depth, true, 0)
			if depth == 8 {
				data = indexedDIBIconProbe(t, 32, depth)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(clamscan, "--database="+dir, "--no-summary", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "Icon.DIB") || !strings.Contains(string(out), "FOUND") {
				t.Fatalf("reference scan: %v: %s", err, out)
			}
		})
	}
}
