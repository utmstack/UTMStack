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

type pamCase struct {
	tuple           string
	depth, maxValue int
	hash            uint64
}

var pamReferenceCases = []pamCase{
	{"BLACKANDWHITE", 1, 1, 0xd2f878f878078707},
	{"BLACKANDWHITE_ALPHA", 2, 1, 0xd2f878f878078707},
	{"GRAYSCALE", 1, 255, 0xa40d5e518bfa1d72},
	{"GRAYSCALE_ALPHA", 2, 255, 0xa40d5e518bfa1d72},
	{"RGB", 3, 255, 0xa08b1e912b7ed734},
	{"RGB_ALPHA", 4, 255, 0xa08b1e912b7ed734},
	{"", 3, 255, 0xa08b1e912b7ed734},
	{"RGB_ALPHA", 4, 65535, 0xa08b1e912b7ed734},
	{"", 4, 65535, 0x957f5042ba559e83},
}

func pamProbe(tc pamCase, width, height, seed int) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "P7\nWIDTH %d\nHEIGHT %d\nDEPTH %d\nMAXVAL %d\n", width, height, tc.depth, tc.maxValue)
	if tc.tuple != "" {
		fmt.Fprintf(&b, "TUPLTYPE %s\n", tc.tuple)
	}
	b.WriteString("ENDHDR\n")
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var values []int
			switch {
			case strings.HasPrefix(tc.tuple, "BLACKANDWHITE"):
				values = []int{(x*3 + y*5 + x*y/7 + seed) % 2}
			case strings.HasPrefix(tc.tuple, "RGB") || tc.depth >= 3:
				values = []int{(x*13 + y*7 + seed*11) % 256, (x*3 + y*19 + seed*17) % 256, (x*23 + y*5 + seed*23) % 256}
			default:
				values = []int{(x*13 + y*7 + seed*11) % 256}
			}
			if !strings.HasPrefix(tc.tuple, "BLACKANDWHITE") && tc.maxValue != 255 {
				for i := range values {
					values[i] = (values[i]*tc.maxValue + 127) / 255
				}
			}
			if strings.HasSuffix(tc.tuple, "_ALPHA") || tc.tuple == "" && tc.depth%2 == 0 {
				values = append(values, (x*31+y*17+seed*13)%(tc.maxValue+1))
			}
			for _, value := range values {
				if tc.maxValue < 256 {
					b.WriteByte(byte(value))
				} else {
					b.WriteByte(byte(value >> 8))
					b.WriteByte(byte(value))
				}
			}
		}
	}
	return b.Bytes()
}

func TestPAMFuzzyRuleDecisions(t *testing.T) {
	for _, tc := range pamReferenceCases {
		t.Run(fmt.Sprintf("%s-%d", tc.tuple, tc.depth), func(t *testing.T) {
			data := pamProbe(tc, 32, 32, 0)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.hash != tc.hash || got.uncertain != 0 {
				t.Fatalf("Go %+v, reference %016x", got, tc.hash)
			}
			for _, match := range []bool{true, false} {
				hash := tc.hash
				if !match {
					hash ^= 1 << 40
				}
				rule := compileTest(t, "ldb", fmt.Sprintf("PAM.Probe;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
				name, err := matchObject(t, rule, string(data), MatchContext{Type: TypeGraphics})
				if match && (name != "PAM.Probe" || err != nil) || !match && (name != "" || err != nil) {
					t.Fatalf("match %t: %q, %v", match, name, err)
				}
			}
		})
	}
	rule := compileTest(t, "ldb", "PAM.Invalid;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000\n")
	for _, bad := range [][]byte{
		[]byte("P7\nWIDTH 1\nHEIGHT 1\nDEPTH 3\nMAXVAL 255\nTUPLTYPE GRAYSCALE\nENDHDR\nabc"),
		[]byte("P7\nWIDTH 1\nHEIGHT 1\nDEPTH 3\nMAXVAL 255\nTUPLTYPE RGB\nENDHDR\nab"),
		[]byte("P7\nWIDTH 1\nHEIGHT 1\nDEPTH 3\nMAXVAL 256\nTUPLTYPE RGB\nENDHDR\n\x01\x01\x00\x00\x00\x00"),
		[]byte("P7\nWIDTH 1\nHEIGHT 1\nDEPTH 3\nMAXVAL 65536\nTUPLTYPE RGB\nENDHDR\nabc"),
		[]byte("P7\n\nWIDTH 1\nHEIGHT 1\nDEPTH 1\nMAXVAL 255\nTUPLTYPE GRAYSCALE\nENDHDR\n\xff"),
		[]byte("P7\n WIDTH 1\nHEIGHT 1\nDEPTH 1\nMAXVAL 255\nTUPLTYPE GRAYSCALE\nENDHDR\n\xff"),
	} {
		expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	}
}

func TestPAMFuzzySigtoolDifferential(t *testing.T) {
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
	cases := []pamCase{
		{"GRAYSCALE", 1, 255, 0}, {"GRAYSCALE_ALPHA", 2, 255, 0},
		{"RGB", 3, 255, 0}, {"RGB_ALPHA", 4, 255, 0},
		{"", 1, 255, 0}, {"", 2, 255, 0}, {"", 3, 255, 0}, {"", 4, 255, 0},
	}
	count, exact := 0, 0
	for _, tc := range cases {
		for _, maxValue := range []int{1, 15, 100, 255, 256, 1000, 65535} {
			tc.maxValue = maxValue
			for _, shape := range [][2]int{{32, 32}, {64, 47}} {
				for seed := 0; seed < 3; seed++ {
					data := pamProbe(tc, shape[0], shape[1], seed)
					path := filepath.Join(dir, fmt.Sprintf("%s-%d-%d-%dx%d-%d.pam", tc.tuple, tc.depth, maxValue, shape[0], shape[1], seed))
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
				}
			}
		}
	}
	t.Logf("PAM differential: %d/%d exact hashes", exact, count)
}

func TestPAMFuzzyClamAVReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte("0:0:5037:PAM:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var ldb strings.Builder
	for i, tc := range pamReferenceCases {
		fmt.Fprintf(&ldb, "PAM.Probe%d;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", i, tc.hash)
	}
	if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte(ldb.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	for i, tc := range pamReferenceCases {
		path := filepath.Join(dir, fmt.Sprintf("PAM-%d.pam", i))
		if err := os.WriteFile(path, pamProbe(tc, 32, 32, 0), 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(root, "bin", "clamscan"), "--database="+dir, "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), path+": PAM.Probe") || !strings.Contains(string(out), "FOUND") {
			t.Fatalf("PAM %d reference: %v, %s", i, err, out)
		}
	}
}
