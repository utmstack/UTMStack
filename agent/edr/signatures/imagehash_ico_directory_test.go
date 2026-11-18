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

// Optional native oracle; fixed reference values still run without the tools.
func referenceICODirectory(t *testing.T, data []byte, want uint64, negative ...uint64) {
	t.Helper()
	wrong := uint64(0x87852e6dc81eb5a5)
	if want == wrong {
		wrong = 0xf87ad19237e04a5a
	}
	if len(negative) != 0 {
		wrong = negative[0]
	}
	referenceImageRules(t, data, want, wrong, false)
}

func referenceImageRules(t *testing.T, data []byte, want, wrong uint64, partial bool) {
	t.Helper()
	fileTypes := "0:0:00000100:ICO:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	if bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}) {
		fileTypes += "0:0:ffd8ff:JPEG:CL_TYPE_ANY:CL_TYPE_JPEG\n"
	}
	if bytes.HasPrefix(data, []byte("#?RADIANCE")) {
		fileTypes += "0:0:233f52414449414e4345:HDR:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	}
	if bytes.HasPrefix(data, []byte("DDS ")) {
		fileTypes += "0:0:44445320:DDS:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	}
	if bytes.HasPrefix(data, []byte("P6\n")) {
		fileTypes += "0:0:50360a:PPM:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	}
	if bytes.HasPrefix(data, []byte{0x76, 0x2f, 0x31, 1}) {
		fileTypes += "0:0:762f3101:EXR:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	}
	if bytes.HasPrefix(data, []byte("II*\x00")) || bytes.HasPrefix(data, []byte("MM\x00*")) {
		fileTypes += "0:0:49492a00:TIFF:CL_TYPE_ANY:CL_TYPE_TIFF\n0:0:4d4d002a:TIFF:CL_TYPE_ANY:CL_TYPE_TIFF\n"
	}
	for _, tc := range []struct {
		hash uint64
		name string
	}{{want, "ICO.Selected"}, {wrong, ""}} {
		db, err := Compile(map[string][]byte{
			"ico.ftm": []byte(fileTypes),
			"ico.ldb": []byte(fmt.Sprintf("ICO.Selected;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash)),
		})
		if err != nil {
			t.Fatal(err)
		}
		typ, complete := db.DetermineType(data)
		if !complete || targetOf(typ) != 5 {
			t.Fatalf("ICO typing: %s %t", typ, complete)
		}
		name, reason := tc.name, ""
		if partial && tc.name != "" {
			name, reason = "", UnknownFuzzyImage
		}
		expectOutcome(t, db, string(data), MatchContext{Type: typ}, name, reason)
	}
	if dir := os.Getenv("EDR_ICO_DIRECTORY_FIXTURES"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%x.ico", sha256.Sum256(data))), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		return
	}
	tool := filepath.Join(root, "bin", "sigtool")
	b, err := os.ReadFile(tool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("unpinned sigtool: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "directory.ico")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"), "CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	cmd := exec.Command(tool, "--fuzzy-img", path)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native image hash: %v: %s", err, out)
	}
	_, text, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
	hash, err := strconv.ParseUint(text, 16, 64)
	if !ok || err != nil || hash != want {
		t.Fatalf("native hash: %s, want %016x", out, want)
	}
	t.Logf("ClamAV input SHA256=%x hash=%016x", sha256.Sum256(data), hash)
	tool = filepath.Join(root, "bin", "clamscan")
	b, err = os.ReadFile(tool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned clamscan: %v", err)
	}
	db := filepath.Join(dir, "db")
	if err := os.Mkdir(db, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(db, "ico.ftm"), []byte(fileTypes), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		hash uint64
		hit  bool
	}{{want, true}, {wrong, false}} {
		rule := fmt.Sprintf("ICO.Selected;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash)
		if err := os.WriteFile(filepath.Join(db, "ico.ldb"), []byte(rule), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(tool, "--database="+db, "--no-summary", path)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if tc.hit {
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "ICO.Selected.UNOFFICIAL FOUND") {
				t.Fatalf("native selected-image detection: %v: %s", err, out)
			}
		} else if err != nil || !strings.Contains(string(out), "directory.ico: OK") {
			t.Fatalf("native wrong-image exclusion: %v: %s", err, out)
		}
		t.Logf("ClamAV rule=%016x input SHA256=%x detected=%t", tc.hash, sha256.Sum256(data), tc.hit)
	}
}

func referenceICOInvalid(t *testing.T, data []byte) {
	t.Helper()
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		return
	}
	tool := filepath.Join(root, "bin", "sigtool")
	b, err := os.ReadFile(tool)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df" {
		t.Fatalf("unpinned sigtool: %v", err)
	}
	path := filepath.Join(t.TempDir(), "invalid.ico")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(tool, "--fuzzy-img", path)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"), "CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if status, ok := err.(*exec.ExitError); !ok || status.ExitCode() != 1 {
		t.Fatalf("invalid input produced a native hash or unexpected failure: %v: %s", err, out)
	}
	t.Logf("ClamAV input SHA256=%x rejected=true", sha256.Sum256(data))
}

func TestICODirectoryThreeFrameTie(t *testing.T) {
	// The reference starts with the last entry and then compares preceding
	// entries in forward order using strict greater-than. If the last entry
	// has a lower score, the first of earlier tied winners remains selected.
	data := multiICOProbe(t, dibIconProbe(t, 32, 32, false, 0), invertedDIBIconProbe(t, 32, 32), dibIconProbe(t, 64, 24, false, 0))
	const want = uint64(0x87852e6dc81eb5a5)
	referenceICODirectory(t, data, want)
	got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	if !got.ok || got.uncertain != 0 || got.hash != want {
		t.Fatalf("selected wrong tied image: %+v want %016x", got, want)
	}
}

func TestICODirectoryTieOrdering(t *testing.T) {
	a := dibIconProbe(t, 32, 32, false, 0)
	b := invertedDIBIconProbe(t, 32, 32)
	l := dibIconProbe(t, 64, 24, false, 0)
	for _, tc := range []struct {
		name   string
		frames [][]byte
		want   uint64
	}{
		{"last-tied-b", [][]byte{a, b}, 0xf87ad19237e04a5a},
		{"last-tied-a", [][]byte{b, a}, 0x87852e6dc81eb5a5},
		{"first-a-last-low", [][]byte{a, b, l}, 0x87852e6dc81eb5a5},
		{"first-b-last-low", [][]byte{b, a, l}, 0xf87ad19237e04a5a},
		{"low-a-last-b", [][]byte{l, a, b}, 0xf87ad19237e04a5a},
		{"low-b-last-a", [][]byte{l, b, a}, 0x87852e6dc81eb5a5},
		{"a-low-last-b", [][]byte{a, l, b}, 0xf87ad19237e04a5a},
		{"b-low-last-a", [][]byte{b, l, a}, 0x87852e6dc81eb5a5},
		{"four-first-a", [][]byte{l, a, b, l}, 0x87852e6dc81eb5a5},
		{"four-first-b", [][]byte{l, b, a, l}, 0xf87ad19237e04a5a},
		{"four-last-a", [][]byte{a, b, l, a}, 0x87852e6dc81eb5a5},
		{"four-last-b", [][]byte{b, a, l, b}, 0xf87ad19237e04a5a},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := multiICOProbe(t, tc.frames...)
			referenceICODirectory(t, data, tc.want)
		})
	}
}

func TestICODirectoryLargeAliasedDirectory(t *testing.T) {
	frame := dibIconProbe(t, 32, 32, false, 0)
	for _, count := range []int{17, 100, 1024, 65535} {
		t.Run(fmt.Sprintf("%d-entries", count), func(t *testing.T) {
			data := make([]byte, 6+16*count)
			data[2] = 1
			binary.LittleEndian.PutUint16(data[4:6], uint16(count))
			for i := 0; i < count; i++ {
				entry := data[6+16*i : 6+16*(i+1)]
				copy(entry, frame[6:22])
				binary.LittleEndian.PutUint32(entry[12:16], uint32(len(data)))
			}
			data = append(data, frame[22:]...)
			referenceICODirectory(t, data, 0x87852e6dc81eb5a5)
		})
	}
}

func TestICODirectoryUnselectedAndPadding(t *testing.T) {
	for _, name := range []string{"unselected-compressed", "unselected-offset", "unselected-short", "trailer", "gap", "aliases", "17-frames", "single-DIB-trailer", "single-PNG-trailer", "directory-overlap"} {
		t.Run(name, func(t *testing.T) {
			low := dibIconProbe(t, 16, 24, false, 0)
			high := dibIconProbe(t, 32, 32, false, 0)
			data := multiICOProbe(t, low, high)
			switch name {
			case "unselected-compressed":
				at := int(binary.LittleEndian.Uint32(data[18:22]))
				binary.LittleEndian.PutUint32(data[at+16:], 1)
			case "unselected-offset":
				binary.LittleEndian.PutUint32(data[18:22], 0xffffffff)
			case "unselected-short":
				binary.LittleEndian.PutUint32(data[14:18], 0)
			case "trailer":
				data = append(data, []byte("ignored physical trailer")...)
			case "gap":
				at := int(binary.LittleEndian.Uint32(data[34:38]))
				data = append(append(append([]byte(nil), data[:at]...), make([]byte, 13)...), data[at:]...)
				binary.LittleEndian.PutUint32(data[34:38], uint32(at+13))
			case "aliases":
				copy(data[6:22], data[22:38])
			case "17-frames":
				frames := make([][]byte, 17)
				for i := range frames {
					frames[i] = low
				}
				frames[8] = high
				data = multiICOProbe(t, frames...)
			case "single-DIB-trailer":
				data = append(high, 0)
			case "single-PNG-trailer":
				data = append(pngIconProbe(t, 32, false), 0)
			case "directory-overlap":
				data = high
				data[4] = 2 // The unused second entry overlaps the DIB header.
			}
			const want = uint64(0x87852e6dc81eb5a5)
			referenceICODirectory(t, data, want)
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.uncertain != 0 || got.hash != want {
				t.Fatalf("valid selected image hidden by %s: %+v", name, got)
			}
		})
	}
}

func TestICODirectoryAdvisoryMetadata(t *testing.T) {
	for _, format := range []string{"PNG", "DIB", "indexed-DIB"} {
		for _, name := range []string{"zero-planes", "maximum-planes", "zero-depth", "maximum-depth", "color-count", "reserved"} {
			t.Run(format+"/"+name, func(t *testing.T) {
				data := pngIconProbe(t, 32, false)
				want := uint64(0x87852e6dc81eb5a5)
				if format == "DIB" {
					data = dibIconProbe(t, 32, 32, false, 0)
				} else if format == "indexed-DIB" {
					data = indexedDIBIconProbe(t, 32, 8)
					want = 0xf8f2d7531351d510
				}
				switch name {
				case "zero-planes":
					binary.LittleEndian.PutUint16(data[10:12], 0)
				case "maximum-planes":
					binary.LittleEndian.PutUint16(data[10:12], 256)
				case "zero-depth":
					binary.LittleEndian.PutUint16(data[12:14], 0)
				case "maximum-depth":
					binary.LittleEndian.PutUint16(data[12:14], 256)
				case "color-count":
					data[8] = 53
				case "reserved":
					data[9] = 0xa5
				}
				referenceICODirectory(t, data, want)
			})
		}
	}
}

func TestICODirectorySelectorBounds(t *testing.T) {
	data := multiICOProbe(t, dibIconProbe(t, 16, 24, false, 0), dibIconProbe(t, 32, 32, false, 0))
	for _, name := range []string{"short-directory", "zero-count", "selected-offset", "invalid-unused-planes", "invalid-unused-depth"} {
		t.Run(name, func(t *testing.T) {
			b := bytes.Clone(data)
			switch name {
			case "short-directory":
				b = b[:37]
			case "zero-count":
				binary.LittleEndian.PutUint16(b[4:6], 0)
			case "selected-offset":
				binary.LittleEndian.PutUint32(b[34:38], 0xffffffff)
			case "invalid-unused-planes":
				binary.LittleEndian.PutUint16(b[10:12], 257)
			case "invalid-unused-depth":
				binary.LittleEndian.PutUint16(b[12:14], 257)
			}
			if got := fuzzyImageHash(b, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("invalid directory became a definite image: %+v", got)
			}
			referenceICOInvalid(t, b)
		})
	}
}
