package bytecode

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// This optional native oracle compares each zlib call with the Go decoder.
// It uses the host C compiler only in tests; endpoint binaries remain pure Go.
func TestStreamInflaterZlibCallBoundaries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("C zlib oracle requires a Unix compiler")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	oracle := filepath.Join(t.TempDir(), "zlib-oracle")
	build := exec.Command(cc, "-O2", "-o", oracle, "testdata/zlib_oracle.c", "-lz")
	if out, err := build.CombinedOutput(); err != nil {
		t.Skipf("zlib C development files unavailable: %v: %s", err, out)
	}
	rng := rand.New(rand.NewSource(186))
	random := make([]byte, 4096)
	if _, err := rng.Read(random); err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{[]byte("small"), bytes.Repeat([]byte("matching history and literal bytes "), 100), random} {
		for _, level := range []int{flate.NoCompression, flate.DefaultCompression, flate.BestSpeed, flate.HuffmanOnly} {
			for _, mode := range []struct {
				name string
				bits int32
				pack func() []byte
			}{
				{"raw", -15, func() []byte { return packedForInflate(t, raw, level, false) }},
				{"zlib", 15, func() []byte { return packedForInflate(t, raw, level, true) }},
				{"auto-zlib", 47, func() []byte { return packedForInflate(t, raw, level, true) }},
				{"gzip", 31, func() []byte {
					var b bytes.Buffer
					w, err := gzip.NewWriterLevel(&b, level)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := w.Write(raw); err != nil {
						t.Fatal(err)
					}
					if err := w.Close(); err != nil {
						t.Fatal(err)
					}
					return b.Bytes()
				}},
			} {
				packed := mode.pack()
				for _, windows := range [][2]int{{1, 1}, {3, 7}, {17, 23}, {8192, 128}} {
					t.Run(fmt.Sprintf("%s/%d/%d/%d/%d", mode.name, len(raw), level, windows[0], windows[1]), func(t *testing.T) {
						cmd := exec.Command(oracle, strconv.Itoa(int(mode.bits)), strconv.Itoa(windows[0]), strconv.Itoa(windows[1]))
						cmd.Stdin = bytes.NewReader(packed)
						output, err := cmd.Output()
						if err != nil {
							t.Fatalf("C zlib oracle: %v", err)
						}
						d, _ := newStreamInflater(mode.bits)
						consumed, steps := 0, 0
						scan := bufio.NewScanner(bytes.NewReader(output))
						scan.Buffer(make([]byte, 4096), 1<<20)
						for scan.Scan() {
							fields := strings.Fields(scan.Text())
							if len(fields) < 3 {
								t.Fatalf("short oracle row %q", scan.Text())
							}
							wantRead, _ := strconv.Atoi(fields[0])
							wantWrite, _ := strconv.Atoi(fields[1])
							wantStatus, _ := strconv.Atoi(fields[2])
							if len(fields) != 3+wantWrite {
								t.Fatalf("oracle row has %d bytes, expected %d", len(fields)-3, wantWrite)
							}
							end := min(len(packed), consumed+windows[0])
							buf := make([]byte, windows[1])
							read, wrote, status, err := d.process(packed[consumed:end], buf)
							if err != nil || read != wantRead || wrote != wantWrite || status != wantStatus {
								t.Fatalf("call %d at input %d: Go=(%d,%d,%d,%v phase=%d bits=%d stored=%d copy=%d) C=(%d,%d,%d)", steps, consumed, read, wrote, status, err, d.phase, d.nbits, d.storedLeft, d.copyLeft, wantRead, wantWrite, wantStatus)
							}
							for i := 0; i < wrote; i++ {
								want, err := strconv.ParseUint(fields[3+i], 16, 8)
								if err != nil || buf[i] != byte(want) {
									t.Fatalf("call %d byte %d: Go=%02x C=%02x error=%v", steps, i, buf[i], want, err)
								}
							}
							consumed += read
							steps++
						}
						if err := scan.Err(); err != nil {
							t.Fatal(err)
						}
						if steps == 0 || d.phase != phaseComplete || consumed != len(packed) {
							t.Fatalf("incomplete comparison: steps=%d phase=%d consumed=%d/%d", steps, d.phase, consumed, len(packed))
						}
					})
				}
			}
		}
	}
}

func TestStreamInflaterZlibSyncRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("C zlib oracle requires a Unix compiler")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	oracle := filepath.Join(t.TempDir(), "zlib-oracle")
	build := exec.Command(cc, "-O2", "-o", oracle, "testdata/zlib_oracle.c", "-lz")
	if out, err := build.CombinedOutput(); err != nil {
		t.Skipf("zlib C development files unavailable: %v: %s", err, out)
	}
	child := []byte("RECOVERED-CONTENT")
	raw := packedForInflate(t, child, flate.DefaultCompression, false)
	corrupt := append([]byte{0x06, 0, 0, 0xff, 0xff}, raw...)
	twice := append([]byte{0x06, 0, 0, 0xff, 0xff, 0x06, 0, 0, 0xff, 0xff}, raw...)
	badChecksum := packedForInflate(t, child, flate.DefaultCompression, true)
	badChecksum[len(badChecksum)-1] ^= 1
	for _, mode := range []struct {
		name     string
		bits     int32
		packed   []byte
		complete bool
	}{
		{"raw", -15, corrupt, true},
		{"raw-twice", -15, twice, true},
		{"zlib", 15, append(append([]byte{0x78, 0x9c}, corrupt...), 0, 0, 0, 0), true},
		{"bad-zlib-header", 15, append([]byte{0, 0, 0, 0, 0xff, 0xff}, raw...), true},
		{"gzip", 31, append(append([]byte{0x1f, 0x8b, 8, 0, 0, 0, 0, 0, 0, 3}, corrupt...), 0, 0, 0, 0, 0, 0, 0, 0), true},
		{"bad-gzip-header", 31, append([]byte{0, 0, 0, 0, 0xff, 0xff}, raw...), true},
		{"bad-checksum", 15, badChecksum, false},
	} {
		for _, windows := range [][2]int{{1, 1}, {3, 7}, {17, 23}, {8192, 128}} {
			t.Run(fmt.Sprintf("%s/%d/%d", mode.name, windows[0], windows[1]), func(t *testing.T) {
				cmd := exec.Command(oracle, strconv.Itoa(int(mode.bits)), strconv.Itoa(windows[0]), strconv.Itoa(windows[1]), "sync")
				cmd.Stdin = bytes.NewReader(mode.packed)
				oracleRows, err := cmd.Output()
				if err != nil {
					t.Fatalf("C zlib oracle: %v", err)
				}
				d, _ := newStreamInflater(mode.bits)
				consumed, steps := 0, 0
				var decoded []byte
				scan := bufio.NewScanner(bytes.NewReader(oracleRows))
				scan.Buffer(make([]byte, 4096), 1<<20)
				for scan.Scan() {
					fields := strings.Fields(scan.Text())
					if len(fields) < 3 {
						t.Fatalf("short oracle row %q", scan.Text())
					}
					wantRead, _ := strconv.Atoi(fields[0])
					wantWrite, _ := strconv.Atoi(fields[1])
					wantStatus, _ := strconv.Atoi(fields[2])
					end := min(len(mode.packed), consumed+windows[0])
					buf := make([]byte, windows[1])
					read, wrote, status, err := d.processWithRecovery(mode.packed[consumed:end], buf)
					if err != nil || read != wantRead || wrote != wantWrite || status != wantStatus {
						t.Fatalf("call %d at %d: Go=(%d,%d,%d,%v phase=%d bits=%d sync=%t have=%d) C=(%d,%d,%d)", steps, consumed, read, wrote, status, err, d.phase, d.nbits, d.needSync, d.syncHave, wantRead, wantWrite, wantStatus)
					}
					if len(fields) != 3+wantWrite {
						t.Fatalf("oracle row has %d bytes, expected %d", len(fields)-3, wantWrite)
					}
					for i := 0; i < wrote; i++ {
						want, err := strconv.ParseUint(fields[3+i], 16, 8)
						if err != nil || buf[i] != byte(want) {
							t.Fatalf("call %d byte %d: Go=%02x C=%02x error=%v", steps, i, buf[i], want, err)
						}
					}
					decoded = append(decoded, buf[:wrote]...)
					consumed += read
					steps++
				}
				if err := scan.Err(); err != nil {
					t.Fatal(err)
				}
				if mode.complete && d.phase != phaseComplete || !mode.complete && (d.phase == phaseComplete || !d.needSync) || consumed != len(mode.packed) || !bytes.Equal(decoded, child) {
					t.Fatalf("incomplete recovery: steps=%d phase=%d consumed=%d/%d decoded=%q", steps, d.phase, consumed, len(mode.packed), decoded)
				}
			})
		}
	}
}

func TestStreamInflaterZlibDictionaryStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("C zlib oracle requires a Unix compiler")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	oracle := filepath.Join(t.TempDir(), "zlib-oracle")
	build := exec.Command(cc, "-O2", "-o", oracle, "testdata/zlib_oracle.c", "-lz")
	if out, err := build.CombinedOutput(); err != nil {
		t.Skipf("zlib C development files unavailable: %v: %s", err, out)
	}
	// 0x78bb is an RFC 1950 header with FDICT set and a valid FCHECK.
	packed := append([]byte{0x78, 0xbb, 1, 2, 3, 4}, packedForInflate(t, []byte("dictionary"), flate.DefaultCompression, false)...)
	for _, windows := range [][2]int{{1, 1}, {3, 7}, {8192, 128}} {
		t.Run(fmt.Sprintf("%d/%d", windows[0], windows[1]), func(t *testing.T) {
			cmd := exec.Command(oracle, "15", strconv.Itoa(windows[0]), strconv.Itoa(windows[1]))
			cmd.Stdin = bytes.NewReader(packed)
			rows, err := cmd.Output()
			if err != nil {
				t.Fatalf("C zlib oracle: %v", err)
			}
			d, _ := newStreamInflater(15)
			consumed, lastStatus := 0, 0
			scan := bufio.NewScanner(bytes.NewReader(rows))
			for scan.Scan() {
				fields := strings.Fields(scan.Text())
				if len(fields) != 3 {
					t.Fatalf("unexpected dictionary oracle row %q", scan.Text())
				}
				wantRead, _ := strconv.Atoi(fields[0])
				wantWrite, _ := strconv.Atoi(fields[1])
				wantStatus, _ := strconv.Atoi(fields[2])
				end := min(len(packed), consumed+windows[0])
				buf := make([]byte, windows[1])
				read, wrote, status, err := d.process(packed[consumed:end], buf)
				if err != nil || read != wantRead || wrote != wantWrite || status != wantStatus {
					t.Fatalf("at %d Go=(%d,%d,%d,%v) C=(%d,%d,%d)", consumed, read, wrote, status, err, wantRead, wantWrite, wantStatus)
				}
				consumed += read
				lastStatus = status
			}
			if err := scan.Err(); err != nil {
				t.Fatal(err)
			}
			if consumed != 6 || lastStatus != inflateNeedDict {
				t.Fatalf("dictionary cursor=%d status=%d", consumed, lastStatus)
			}
		})
	}
}
