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

func extendedBitfieldDIBIconProbe(t *testing.T, size, depth, header int) []byte {
	t.Helper()
	if header != 52 && header != 56 && header != 108 && header != 124 {
		t.Fatal("unsupported probe header")
	}
	base := bitfieldDIBICOProbe(t, size, depth, true)
	old := base[22:]
	payload := make([]byte, len(old)+header-52)
	copy(payload[:40], old[:40])
	copy(payload[40:52], old[40:52])
	copy(payload[header:], old[52:])
	binary.LittleEndian.PutUint32(payload[:4], uint32(header))
	if depth == 32 && header >= 56 {
		binary.LittleEndian.PutUint32(payload[52:56], 0xff000000)
	}
	data := append(bytes.Clone(base[:22]), payload...)
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	return data
}

func TestExtendedBitfieldDIBICOReferenceProbe(t *testing.T) {
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
		for _, header := range []int{52, 56} {
			for _, size := range []int{16, 32, 64} {
				t.Run(fmt.Sprintf("%d-bit-%d-header-%d", depth, header, size), func(t *testing.T) {
					data := extendedBitfieldDIBIconProbe(t, size, depth, header)
					path := filepath.Join(t.TempDir(), "extended-bitfields.ico")
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
					if !got.ok || got.uncertain != 0 || got.hash != want {
						t.Fatalf("Go %+v, reference %016x", got, want)
					}
					t.Logf("Go %016x reference %016x", got.hash, want)
				})
			}
		}
	}
	for _, header := range []int{108, 124} {
		t.Run(fmt.Sprintf("unproved-%d-header", header), func(t *testing.T) {
			data := extendedBitfieldDIBIconProbe(t, 32, 16, header)
			path := filepath.Join(t.TempDir(), "extended-bitfields.ico")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(sigtool, "--fuzzy-img", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "ICO image data size did not match expected size") {
				t.Fatalf("unexpected reference result: %v: %s", err, out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}

func TestExtendedBitfieldDIBICOReferenceHashes(t *testing.T) {
	for _, depth := range []int{16, 32} {
		for _, header := range []int{52, 56} {
			data := extendedBitfieldDIBIconProbe(t, 32, depth, header)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.uncertain != 0 || got.hash != 0x88ec267b3e95c0b5 {
				t.Fatalf("%d-bit %d-byte header: %+v", depth, header, got)
			}
		}
	}
	rule := compileTest(t, "ldb", "Icon.ExtendedBitfields;Engine:150-255,Target:0;0;fuzzy_img#88ec267b3e95c0b5\n")
	expectOutcome(t, rule, string(extendedBitfieldDIBIconProbe(t, 32, 16, 52)),
		MatchContext{Type: TypeGraphics}, "Icon.ExtendedBitfields", "")
	for _, header := range []int{108, 124} {
		data := extendedBitfieldDIBIconProbe(t, 32, 16, header)
		expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	}
	bad := extendedBitfieldDIBIconProbe(t, 32, 16, 52)
	bad[22+40] ^= 1
	expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	bad = extendedBitfieldDIBIconProbe(t, 32, 16, 56)
	bad[22+52] = 5 // A noncontiguous alpha mask is invalid.
	expectOutcome(t, rule, string(bad), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
}

func TestExtendedBitfieldDIBICOClamAVTypingReference(t *testing.T) {
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
	for _, depth := range []int{16, 32} {
		for _, header := range []int{52, 56} {
			t.Run(fmt.Sprintf("%d-bit-%d-header", depth, header), func(t *testing.T) {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte("0:0:00000100:ICO:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "probe.ldb"), []byte("Icon.ExtendedBitfields;Engine:150-255,Target:0;0;fuzzy_img#88ec267b3e95c0b5\n"), 0600); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "probe.ico")
				if err := os.WriteFile(path, extendedBitfieldDIBIconProbe(t, 32, depth, header), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(clamscan, "--database="+dir, "--no-summary", path)
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err == nil || !strings.Contains(string(out), "Icon.ExtendedBitfields.UNOFFICIAL FOUND") {
					t.Fatalf("reference scan: %v: %s", err, out)
				}
				t.Log(strings.TrimSpace(string(out)))
			})
		}
	}
}
