package bytecode

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ulikunitz/xz/lzma"
)

func TestWriteLZMAStreamingPropertyCorpus(t *testing.T) {
	dir := os.Getenv("EDR_LZMA_STREAM_WRITE_PROPERTIES")
	if dir == "" {
		t.Skip("set corpus path to export owned property controls")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	body := []byte("HARMLESS-CHILD")
	for i := 0; i < 512; i++ {
		body = append(body, byte(i*73), byte(i>>2))
	}
	for prop := 0; prop < 225; prop++ {
		properties, err := lzma.PropertiesForCode(byte(prop))
		if err != nil {
			t.Fatal(err)
		}
		for _, known := range []bool{false, true} {
			cfg := lzma.WriterConfig{Properties: &properties, DictCap: 4096}
			if known {
				cfg.Size = int64(len(body))
			}
			var b bytes.Buffer
			w, err := cfg.NewWriter(&b)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(body); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, fmt.Sprintf("body-prop-%03d-known-%t.lzma", prop, known))
			if err := os.WriteFile(path, b.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestLZMAStreamNativeReference(t *testing.T) {
	oracle, corpus := os.Getenv("EDR_LZMA_STREAM_ORACLE"), os.Getenv("EDR_LZMA_STREAM_CORPUS")
	if oracle == "" || corpus == "" {
		t.Skip("set native SDK oracle and owned corpus paths")
	}
	entries, err := os.ReadDir(corpus)
	if err != nil {
		t.Fatal(err)
	}
	traces, calls := 0, 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lzma") {
			continue
		}
		input, err := os.ReadFile(filepath.Join(corpus, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, window := range [][2]int{{1, 1}, {3, 5}, {17, 23}, {512, 31}, {8192, 4096}, {1 << 20, 1 << 20}} {
			t.Run(fmt.Sprintf("%s/%d/%d", entry.Name(), window[0], window[1]), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, oracle, strconv.Itoa(window[0]), strconv.Itoa(window[1]))
				cmd.Stdin = bytes.NewReader(input)
				raw, err := cmd.Output()
				if err != nil {
					t.Fatal(err)
				}
				if dir := os.Getenv("EDR_LZMA_STREAM_TRACE_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(dir, fmt.Sprintf("%s-%d-%d.txt.gz", entry.Name(), window[0], window[1]))
					f, e := os.Create(path)
					if e != nil {
						t.Fatal(e)
					}
					z := gzip.NewWriter(f)
					_, e = z.Write(raw)
					ze, fe := z.Close(), f.Close()
					if e != nil || ze != nil || fe != nil {
						t.Fatalf("preserve trace: %v %v %v", e, ze, fe)
					}
				}
				calls += compareLZMATrace(t, input, window[0], window[1], bytes.NewReader(raw))
				traces++
			})
		}
	}
	if traces == 0 {
		t.Fatal("empty corpus")
	}
	t.Logf("matched %d native decisions across %d traces", calls, traces)
}

func compareLZMATrace(t *testing.T, input []byte, inWindow, outWindow int, source io.Reader) int {
	t.Helper()
	scan := bufio.NewScanner(source)
	scan.Buffer(make([]byte, 4096), 4<<20)
	var d *streamLZMA
	offset, call := 0, 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) < 3 {
			t.Fatal("short reference row")
		}
		values := [3]int{}
		for i := range values {
			n, e := strconv.Atoi(fields[i])
			if e != nil {
				t.Fatal(e)
			}
			values[i] = n
		}
		var n, m, status int
		var e error
		buf := make([]byte, outWindow)
		if call == 0 {
			if len(input) < 13 {
				status = -1
			} else {
				n = 13
				_, _, valid := lzmaAllocation(input[:13])
				if !valid {
					status = 1
				} else {
					d, e = newStreamLZMA(input[:13])
				}
			}
		} else if offset == len(input) {
			status = -1
		} else {
			n, m, status, e = d.process(input[offset:min(len(input), offset+inWindow)], buf, nil)
		}
		if e != nil || n != values[0] || m != values[1] || status != values[2] {
			var state any
			if d != nil {
				state = d.lzmaRegisters
			}
			t.Fatalf("call %d offset %d: Go=(%d,%d,%d,%v) reference=%v state=%+v", call, offset, n, m, status, e, values, state)
		}
		want, e := hex.DecodeString(strings.Join(fields[3:], ""))
		if e != nil || !bytes.Equal(buf[:m], want) {
			t.Fatalf("call %d output differs: %v", call, e)
		}
		offset += n
		call++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if call == 0 {
		t.Fatal("empty native trace")
	}
	return call
}

func FuzzLZMAStream(f *testing.F) {
	var b bytes.Buffer
	w, err := (lzma.WriterConfig{DictCap: 4096}).NewWriter(&b)
	if err != nil {
		f.Fatal(err)
	}
	if _, err := w.Write([]byte(lzmaTestChild)); err != nil {
		f.Fatal(err)
	}
	if err := w.Close(); err != nil {
		f.Fatal(err)
	}
	f.Add(b.Bytes(), uint8(1), uint8(1))
	f.Add([]byte("short"), uint8(3), uint8(5))
	f.Fuzz(func(t *testing.T, input []byte, iw, ow uint8) {
		if len(input) < 13 || len(input) > 65536 {
			return
		}
		dict, _, valid := lzmaAllocation(input[:13])
		if !valid || dict > 65536 {
			return
		}
		d, err := newStreamLZMA(input[:13])
		if err != nil {
			t.Fatal(err)
		}
		offset := 13
		var output []byte
		work := uint64(0)
		charge := func(n uint64) error {
			work += n
			if work > 1<<18 {
				return ErrLimit
			}
			return nil
		}
		for calls := 0; calls < 2*len(input)+16384 && offset < len(input); calls++ {
			buf := make([]byte, int(ow)+1)
			available := min(len(input)-offset, int(iw)+1)
			n, m, status, err := d.process(input[offset:offset+available], buf, charge)
			if n < 0 || n > available || m < 0 || m > len(buf) {
				t.Fatal("invalid streaming cursor")
			}
			offset += n
			output = append(output, buf[:m]...)
			if len(output) > 1<<20 {
				return
			}
			if err != nil || status == 1 {
				return
			}
			if status == 2 {
				size := binary.LittleEndian.Uint64(input[5:13])
				if size != ^uint64(0) {
					// The native interface also reports stream end on an EOS
					// marker before the declared size is exhausted. It still
					// never publishes more than the declared output ceiling.
					if uint64(len(output)) > size {
						t.Fatal("exceeded declared output ceiling")
					}
					return
				}
				r, e := (lzma.ReaderConfig{DictCap: 65536}).NewReader(bytes.NewReader(input[:offset]))
				if e != nil {
					t.Fatal(e)
				}
				want, e := io.ReadAll(io.LimitReader(r, 1<<20+1))
				if e != nil || !bytes.Equal(want, output) {
					t.Fatalf("completed output differs from independent Go decoder: %v", e)
				}
				return
			}
			if n == 0 && m == 0 {
				return
			}
		}
	})
}
