package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var pnmReferenceHashes = [...]uint64{
	0xb1ce71cac631c93c, 0xa40d5e518bfa1d72, 0xa08b1e912b7ed734,
	0xb1ce71cac631c93c, 0xa40d5e518bfa1d72, 0xa08b1e912b7ed734,
}

func pnmProbe(kind int) []byte {
	return pnmProbeSize(kind, 32, 32, 0)
}

func pnmProbeSize(kind, width, height, seed int) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "P%d\n# benign reference probe\n%d %d\n", kind, width, height)
	if kind != 1 && kind != 4 {
		b.WriteString("255\n")
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			switch kind {
			case 1:
				fmt.Fprintf(&b, "%d ", pnmBit(x+seed, y+seed))
			case 2:
				fmt.Fprintf(&b, "%d ", (x*13+y*7+seed*11)%256)
			case 3:
				fmt.Fprintf(&b, "%d %d %d ", (x*13+y*7+seed*11)%256, (x*3+y*19+seed*17)%256, (x*23+y*5+seed*23)%256)
			case 4:
				if x%8 == 0 {
					v := byte(0)
					for bit := 0; bit < 8; bit++ {
						v |= byte(pnmBit(x+bit+seed, y+seed)) << uint(7-bit)
					}
					b.WriteByte(v)
				}
			case 5:
				b.WriteByte(byte((x*13 + y*7 + seed*11) % 256))
			case 6:
				b.Write([]byte{byte((x*13 + y*7 + seed*11) % 256), byte((x*3 + y*19 + seed*17) % 256), byte((x*23 + y*5 + seed*23) % 256)})
			}
		}
	}
	if kind < 4 {
		b.WriteByte('\n')
	}
	return b.Bytes()
}

// The oracle is intentionally opt-in: the pinned binary is outside the Go
// module, and routine package tests must not require ClamAV to be installed.
func TestPNMFuzzySigtoolDifferential(t *testing.T) {
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
	shapes := [][2]int{{17, 37}, {64, 64}, {128, 73}}
	dir := t.TempDir()
	exact, uncertain := 0, 0
	for kind := 1; kind <= 6; kind++ {
		for _, shape := range shapes {
			for seed := 0; seed < 3; seed++ {
				data := pnmProbeSize(kind, shape[0], shape[1], seed)
				path := filepath.Join(dir, fmt.Sprintf("P%d-%dx%d-%d.pnm", kind, shape[0], shape[1], seed))
				if err := os.WriteFile(path, data, 0o600); err != nil {
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
				got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				if !got.ok || (got.hash^want)&^got.uncertain != 0 {
					t.Errorf("%s: Go %+v, reference %016x", filepath.Base(path), got, want)
				}
				if got.hash == want {
					exact++
				}
				if got.uncertain != 0 {
					uncertain++
				}
			}
		}
	}
	t.Logf("PNM differential: %d/%d exact hashes; %d uncertain masks", exact, 6*len(shapes)*3, uncertain)
}

func pnmBit(x, y int) int {
	if (x*13+y*7+x*y/3)%11 < 5 {
		return 1
	}
	return 0
}

func pnmDepthProbe(kind, maxValue, seed int) []byte {
	return pnmDepthProbeSize(kind, maxValue, seed, 32, 32)
}

func pnmDepthProbeSize(kind, maxValue, seed, width, height int) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "P%d\n%d %d\n%d\n", kind, width, height, maxValue)
	channels := 1
	if kind == 3 || kind == 6 {
		channels = 3
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			for c := 0; c < channels; c++ {
				value := ((x*37 + y*17 + c*71 + seed*29 + x*y/7) * maxValue / 255) % (maxValue + 1)
				if kind == 2 || kind == 3 {
					fmt.Fprintf(&b, "%d ", value)
				} else if maxValue < 256 {
					b.WriteByte(byte(value))
				} else {
					b.WriteByte(byte(value >> 8))
					b.WriteByte(byte(value))
				}
			}
		}
	}
	if kind == 2 || kind == 3 {
		b.WriteByte('\n')
	}
	return b.Bytes()
}

func TestPNMDeclaredMaxValueSigtoolDifferential(t *testing.T) {
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
	count, exact, uncertain := 0, 0, 0
	for _, kind := range []int{2, 3, 5, 6} {
		for _, maxValue := range []int{1, 15, 100, 254, 256, 1000, 65535} {
			for seed := 0; seed < 4; seed++ {
				for _, shape := range [][2]int{{32, 32}, {64, 47}} {
					data := pnmDepthProbeSize(kind, maxValue, seed, shape[0], shape[1])
					path := filepath.Join(dir, fmt.Sprintf("P%d-%d-%d-%dx%d.pnm", kind, maxValue, seed, shape[0], shape[1]))
					if err := os.WriteFile(path, data, 0o600); err != nil {
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
					got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
					if !got.ok || (got.hash^want)&^got.uncertain != 0 {
						t.Errorf("%s: Go %+v, reference %016x", filepath.Base(path), got, want)
					}
					count++
					if got.hash == want {
						exact++
					}
					if got.uncertain != 0 {
						uncertain++
					}
				}
			}
		}
	}
	t.Logf("PNM declared max values: %d/%d exact hashes, %d uncertain masks", exact, count, uncertain)
}

func TestPNMFuzzyRuleDecisions(t *testing.T) {
	for kind := 1; kind <= 6; kind++ {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			data := pnmProbe(kind)
			h := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			want := pnmReferenceHashes[kind-1]
			if !h.ok || h.hash != want || h.uncertain != 0 {
				t.Fatalf("P%d hash %+v, reference %016x", kind, h, want)
			}
			makeRule := func(hash uint64) *Database {
				return compileTest(t, "ldb", fmt.Sprintf("PNM.Probe;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
			}
			if name, err := matchObject(t, makeRule(want), string(data), MatchContext{Type: TypeGraphics}); name != "PNM.Probe" || err != nil {
				t.Fatalf("matching PNM rule: %q, %v", name, err)
			}
			if name, err := matchObject(t, makeRule(want^(1<<40)), string(data), MatchContext{Type: TypeGraphics}); name != "" || err != nil {
				t.Fatalf("nonmatching PNM rule: %q, %v", name, err)
			}
		})
	}
	unsupported := []byte("P5\n4 4\n65536\n0123456789012345")
	rule := compileTest(t, "ldb", "PNM.Depth;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000\n")
	expectOutcome(t, rule, string(unsupported), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	for _, tc := range []struct {
		kind, maxValue, seed int
		hash                 uint64
	}{
		{3, 100, 0, 0xa2e76effa8005592},
		{6, 65535, 0, 0xb8fbbaff82005505},
	} {
		data := pnmDepthProbe(tc.kind, tc.maxValue, tc.seed)
		for _, match := range []bool{true, false} {
			hash := tc.hash
			if !match {
				hash ^= 1 << 40
			}
			db := compileTest(t, "ldb", fmt.Sprintf("PNM.Depth;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
			name, err := matchObject(t, db, string(data), MatchContext{Type: TypeGraphics})
			if match && (name != "PNM.Depth" || err != nil) || !match && (name != "" || err != nil) {
				t.Fatalf("P%d max %d match %t: %q, %v", tc.kind, tc.maxValue, match, name, err)
			}
		}
	}
	for _, data := range [][]byte{
		[]byte("P6\n2 2\n255\n\x01\x02\x03"),
		[]byte("P2\n2 2\n255\n1 2 999 4"),
		[]byte("P4\n9 2\n\xff"),
		[]byte("P6\n16385 1\n255\n"),
		[]byte("P5\n2 2\n1000\n\x00\x01\x00\x02"),
		[]byte("P5\n1 1\n15\n\x10"),
	} {
		expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	}
}

func TestPNMFuzzyClamAVReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	dir := t.TempDir()
	var ftm, ldb strings.Builder
	var paths []string
	for kind := 1; kind <= 6; kind++ {
		fmt.Fprintf(&ftm, "0:0:503%d:PNM%d:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", kind, kind)
		fmt.Fprintf(&ldb, "PNM.Probe%d;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", kind, pnmReferenceHashes[kind-1])
		path := filepath.Join(dir, fmt.Sprintf("P%d.pnm", kind))
		if err := os.WriteFile(path, pnmProbe(kind), 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	for _, tc := range []struct {
		name                 string
		kind, maxValue, seed int
		hash                 uint64
	}{
		{"P3-max100.pnm", 3, 100, 0, 0xa2e76effa8005592},
		{"P6-max65535.pnm", 6, 65535, 0, 0xb8fbbaff82005505},
	} {
		fmt.Fprintf(&ldb, "PNM.Depth%d;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.maxValue, tc.hash)
		path := filepath.Join(dir, tc.name)
		if err := os.WriteFile(path, pnmDepthProbe(tc.kind, tc.maxValue, tc.seed), 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	for name, content := range map[string]string{"probe.ftm": ftm.String(), "probe.ldb": ldb.String()} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range paths {
		cmd := exec.Command(filepath.Join(root, "bin", "clamscan"), "--database="+dir, "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), path+": PNM.") || !strings.Contains(string(out), "FOUND") {
			t.Fatalf("%s reference: %v, %s", filepath.Base(path), err, out)
		}
	}
}
