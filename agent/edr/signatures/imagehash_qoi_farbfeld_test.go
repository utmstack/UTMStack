package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func qoiProbe(width, height, seed int, ops bool) []byte {
	var b bytes.Buffer
	b.WriteString("qoif")
	_ = binary.Write(&b, binary.BigEndian, uint32(width))
	_ = binary.Write(&b, binary.BigEndian, uint32(height))
	b.Write([]byte{4, 0})
	if !ops {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				b.Write([]byte{0xff, byte(x*13 + y*7 + seed), byte(x*3 + y*19 + seed*17), byte(x*23 + y*5 + seed*23), byte(x*31 + y*17 + seed*13)})
			}
		}
	} else {
		// One of every chunk family, then bounded runs to complete the image.
		b.Write([]byte{0xfe, 37, 72, 111, 0xc2, 0x79, 0xa8, 0x4d, 0xff, 12, 210, 64, 90})
		index := (37*3 + 72*5 + 111*7 + 255*11) % 64
		b.WriteByte(byte(index))
		left := width*height - 8
		for left > 0 {
			n := min(left, 62)
			b.WriteByte(byte(0xc0 | (n - 1)))
			left -= n
		}
	}
	b.Write(qoiEnd)
	return b.Bytes()
}

func farbfeldProbe(width, height, seed int) []byte {
	var b bytes.Buffer
	b.WriteString("farbfeld")
	_ = binary.Write(&b, binary.BigEndian, uint32(width))
	_ = binary.Write(&b, binary.BigEndian, uint32(height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			for c := 0; c < 4; c++ {
				v := uint16((x*3187 + y*5123 + c*7297 + seed*1049) & 0xffff)
				_ = binary.Write(&b, binary.BigEndian, v)
			}
		}
	}
	return b.Bytes()
}

func TestQOIFarbfeldBoundsAndRules(t *testing.T) {
	typed, err := Compile(map[string][]byte{
		"probe.ftm": []byte("0:0:716f6966:QOI:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n0:0:6661726266656c64:Farbfeld:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("Image.Typed;Engine:150-255,Target:0;0;fuzzy_img#a08b1e912b7ed734\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
		hash uint64
	}{
		{"qoi-gradient", qoiProbe(32, 32, 0, false), 0xa08b1e912b7ed734},
		{"qoi-ops", qoiProbe(32, 32, 0, true), 0xe0e0e0e1e1f1f1f1},
		{"farbfeld", farbfeldProbe(32, 32, 0), 0xa4dca5dca50e5a74},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if typ, complete := typed.DetermineType(tc.data); !complete || typ != TypeGraphics {
				t.Fatalf("file type = %s, complete=%t", typ, complete)
			}
			got := fuzzyImageHash(tc.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.uncertain != 0 || got.hash != tc.hash {
				t.Fatalf("hash: %+v", got)
			}
			rule := compileTest(t, "ldb", fmt.Sprintf("Image.Probe;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash))
			name, err := matchObject(t, rule, string(tc.data), MatchContext{Type: TypeGraphics})
			if name != "Image.Probe" || err != nil {
				t.Fatalf("match: %q, %v", name, err)
			}
			negative := compileTest(t, "ldb", fmt.Sprintf("Image.Other;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash^(1<<40)))
			name, err = matchObject(t, negative, string(tc.data), MatchContext{Type: TypeGraphics})
			if name != "" || err != nil {
				t.Fatalf("non-match: %q, %v", name, err)
			}
		})
	}
	badQOI := qoiProbe(32, 32, 0, false)
	badQOI = badQOI[:len(badQOI)-10]
	badQOITrailer := qoiProbe(32, 32, 0, false)
	badQOITrailer[len(badQOITrailer)-1] = 0
	badQOIChannels := qoiProbe(32, 32, 0, false)
	badQOIChannels[12] = 2
	badQOIRun := qoiProbe(1, 1, 0, true)
	badQOIRun[14] = 0xfd
	badFarbfeld := farbfeldProbe(32, 32, 0)
	badFarbfeld = badFarbfeld[:len(badFarbfeld)-1]
	oversizeQOI := qoiProbe(32, 32, 0, false)
	binary.BigEndian.PutUint32(oversizeQOI[4:8], imageHashMaxDimension+1)
	oversizeFarbfeld := farbfeldProbe(32, 32, 0)
	binary.BigEndian.PutUint32(oversizeFarbfeld[8:12], imageHashMaxDimension+1)
	for _, bad := range [][]byte{badQOI, badQOITrailer, badQOIChannels, badQOIRun, badFarbfeld, oversizeQOI, oversizeFarbfeld} {
		h := fuzzyImageHash(bad, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if h.ok {
			t.Fatalf("invalid image hashed: %+v", h)
		}
		rule := compileTest(t, "ldb", "Image.Invalid;Engine:150-255,Target:0;0;fuzzy_img#a08b1e912b7ed734\n")
		expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	}
}

func TestQOIFarbfeldClamAVReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	dir := t.TempDir()
	ftm := "0:0:716f6966:QOI:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n" +
		"0:0:6661726266656c64:Farbfeld:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte(ftm), 0o600); err != nil {
		t.Fatal(err)
	}
	ldb := "Image.QOIGradient;Engine:150-255,Target:0;0;fuzzy_img#a08b1e912b7ed734\n" +
		"Image.QOIOps;Engine:150-255,Target:0;0;fuzzy_img#e0e0e0e1e1f1f1f1\n" +
		"Image.Farbfeld;Engine:150-255,Target:0;0;fuzzy_img#a4dca5dca50e5a74\n"
	if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte(ldb), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, signature string
		data            []byte
	}{
		{"gradient.qoi", "Image.QOIGradient", qoiProbe(32, 32, 0, false)},
		{"ops.qoi", "Image.QOIOps", qoiProbe(32, 32, 0, true)},
		{"probe.ff", "Image.Farbfeld", farbfeldProbe(32, 32, 0)},
	} {
		path := filepath.Join(dir, tc.name)
		if err := os.WriteFile(path, tc.data, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(root, "bin", "clamscan"), "--database="+dir, "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), path+": "+tc.signature+".UNOFFICIAL FOUND") {
			t.Fatalf("%s reference: %v, %s", tc.name, err, out)
		}
	}
}

func TestQOIFarbfeldSigtoolDifferential(t *testing.T) {
	sigtool := os.Getenv("EDR_CLAMAV_SIGTOOL")
	if sigtool == "" {
		t.Skip("set EDR_CLAMAV_SIGTOOL to the pinned ClamAV 1.5.4 sigtool")
	}
	toolBytes, err := os.ReadFile(sigtool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("sigtool is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	root := filepath.Dir(filepath.Dir(sigtool))
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	dir := t.TempDir()
	count := 0
	for _, shape := range [][2]int{{32, 32}, {64, 47}, {257, 33}} {
		for seed := 0; seed < 4; seed++ {
			for _, tc := range []struct {
				name string
				data []byte
			}{
				{"qoi-gradient", qoiProbe(shape[0], shape[1], seed, false)},
				{"qoi-ops", qoiProbe(shape[0], shape[1], seed, true)},
				{"farbfeld", farbfeldProbe(shape[0], shape[1], seed)},
			} {
				path := filepath.Join(dir, fmt.Sprintf("%s-%dx%d-%d.img", tc.name, shape[0], shape[1], seed))
				if err := os.WriteFile(path, tc.data, 0o600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(sigtool, "--fuzzy-img", path)
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("sigtool %s: %v, %s", path, err, out)
				}
				_, value, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
				if !ok {
					t.Fatalf("unexpected sigtool hash: %s", out)
				}
				want, err := strconv.ParseUint(value, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				if shape == [2]int{32, 32} && seed == 0 {
					t.Logf("golden %s %016x", tc.name, want)
				}
				got := fuzzyImageHash(tc.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !got.ok || (got.hash^want)&^got.uncertain != 0 {
					t.Errorf("%s: Go %+v, reference %016x", filepath.Base(path), got, want)
				}
				count++
			}
		}
	}
	t.Logf("QOI/farbfeld differential: %d hashes", count)
}
