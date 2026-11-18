package bytecode

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBzip2StreamAcrossWindows(t *testing.T) {
	packed := bzip2TestPacked(t)
	for _, window := range [][2]int{{1, 1}, {2, 3}, {7, 1}, {3, 5}, {64, 13}, {64, 64}} {
		t.Run(fmt.Sprint(window), func(t *testing.T) {
			d := new(streamBzip2)
			var output []byte
			offset := 0
			for calls := 0; calls < 1000; calls++ {
				buf := make([]byte, window[1])
				n, m, status, err := d.process(packed[offset:min(len(packed), offset+window[0])], buf, nil)
				if err != nil || status < 0 {
					t.Fatalf("decode: n=%d m=%d status=%d err=%v", n, m, status, err)
				}
				offset += n
				output = append(output, buf[:m]...)
				if status == bzipEnd {
					if offset != len(packed) || string(output) != bzip2TestChild {
						t.Fatalf("decoded %q at %d", output, offset)
					}
					return
				}
				if n == 0 && m == 0 {
					t.Fatal("stalled")
				}
			}
			t.Fatal("did not finish")
		})
	}
}

func TestBzip2StreamNativeOracle(t *testing.T) {
	oracle := os.Getenv("EDR_BZIP2_ORACLE")
	corpus := os.Getenv("EDR_BZIP2_CORPUS")
	if oracle == "" || corpus == "" {
		t.Skip("set EDR_BZIP2_ORACLE and EDR_BZIP2_CORPUS for libbz2 differential")
	}
	entries, err := os.ReadDir(corpus)
	if err != nil {
		t.Fatal(err)
	}
	comparisons := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bz2") {
			continue
		}
		packed, err := os.ReadFile(corpus + "/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, windows := range [][2]int{{1, 1}, {3, 7}, {17, 23}, {512, 31}, {1024, 1024}, {8192, 128}, {1 << 20, 1 << 20}} {
			t.Run(fmt.Sprintf("%s/%d/%d", entry.Name(), windows[0], windows[1]), func(t *testing.T) {
				cmd := exec.Command(oracle, strconv.Itoa(windows[0]), strconv.Itoa(windows[1]))
				cmd.Stdin = bytes.NewReader(packed)
				raw, err := cmd.Output()
				if err != nil {
					t.Fatal(err)
				}
				if dir := os.Getenv("EDR_BZIP2_TRACE_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(dir, fmt.Sprintf("%s-%d-%d.txt.gz", entry.Name(), windows[0], windows[1]))
					file, err := os.Create(path)
					if err != nil {
						t.Fatal(err)
					}
					z := gzip.NewWriter(file)
					_, writeErr := z.Write(raw)
					closeErr := z.Close()
					fileErr := file.Close()
					if writeErr != nil || closeErr != nil || fileErr != nil {
						t.Fatalf("retain trace: %v %v %v", writeErr, closeErr, fileErr)
					}
				}
				comparisons += compareBzip2Trace(t, packed, windows[0], windows[1], bytes.NewReader(raw))
			})
		}
	}
	if comparisons == 0 {
		t.Fatal("empty corpus")
	}
	t.Logf("matched %d independent libbz2 calls", comparisons)
}

func FuzzBzip2Stream(f *testing.F) {
	seed, _ := hex.DecodeString(bzip2TestHex)
	f.Add(seed, uint8(1), uint8(1))
	f.Add([]byte("BZh9"), uint8(7), uint8(3))
	f.Fuzz(func(t *testing.T, packed []byte, iw, ow uint8) {
		if len(packed) > 65536 {
			t.Skip()
		}
		d := new(streamBzip2)
		offset := 0
		var output []byte
		for calls := 0; calls < 2*len(packed)+100; calls++ {
			buf := make([]byte, int(ow)+1)
			n, m, status, err := d.process(packed[offset:min(len(packed), offset+int(iw)+1)], buf, nil)
			if n < 0 || n > min(len(packed)-offset, int(iw)+1) || m < 0 || m > len(buf) {
				t.Fatal("invalid cursor")
			}
			offset += n
			output = append(output, buf[:m]...)
			if len(output) > 1<<20 {
				return
			}
			if status == bzipEnd {
				if d.randomSeen {
					return
				} // Go's standard decoder deliberately rejects this legacy wire profile.
				want, e := io.ReadAll(io.LimitReader(bzip2.NewReader(bytes.NewReader(packed[:offset])), 1<<20+1))
				if e != nil || !bytes.Equal(want, output) {
					t.Fatalf("completed output contradicts independent Go decoder: %v", e)
				}
				return
			}
			if err != nil || status < 0 || n == 0 && m == 0 {
				return
			}
		}
	})
}

func compareBzip2Trace(t *testing.T, packed []byte, inWindow, outWindow int, source io.Reader) int {
	t.Helper()
	scan := bufio.NewScanner(source)
	scan.Buffer(make([]byte, 4096), 4<<20)
	d := new(streamBzip2)
	offset, calls := 0, 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) < 3 {
			t.Fatal("short oracle row")
		}
		n, _ := strconv.Atoi(fields[0])
		m, _ := strconv.Atoi(fields[1])
		status, _ := strconv.Atoi(fields[2])
		buf := make([]byte, outWindow)
		var gn, gm, gs int
		var ge error
		if offset == len(packed) {
			gs = -1
		} else {
			gn, gm, gs, ge = d.process(packed[offset:min(len(packed), offset+inWindow)], buf, nil)
		}
		if ge != nil || gn != n || gm != m || gs != status {
			t.Fatalf("call %d offset %d: Go=(%d,%d,%d,%v,phase %d,bits %d) C=(%d,%d,%d)", calls, offset, gn, gm, gs, ge, d.phase, d.bitCount, n, m, status)
		}
		want, err := hex.DecodeString(strings.Join(fields[3:], ""))
		if err != nil || !bytes.Equal(buf[:gm], want) {
			t.Fatalf("call %d output differs", calls)
		}
		offset += gn
		calls++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if calls == 0 {
		t.Fatal("no oracle calls")
	}
	return calls
}
