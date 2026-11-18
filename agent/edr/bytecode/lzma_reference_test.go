package bytecode

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Expected cursors, statuses and output are retained from the independent
// pinned native SDK, including corrupt inputs. No C tooling is needed in CI.
func TestLZMAStreamReferenceReplay(t *testing.T) {
	f, err := os.Open("testdata/lzma_stream_reference.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	r := tar.NewReader(z)
	corpus := map[string][]byte{}
	traces, calls := 0, 0
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag != tar.TypeReg || h.Size < 0 || h.Size > 8<<20 {
			t.Fatal("invalid reference member")
		}
		if name, ok := strings.CutPrefix(h.Name, "corpus/"); ok {
			if _, exists := corpus[name]; exists {
				t.Fatal("duplicate input")
			}
			data, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			corpus[name] = data
			continue
		}
		name, ok := strings.CutPrefix(h.Name, "traces/")
		if !ok {
			t.Fatal("unknown reference member")
		}
		parts := strings.Split(strings.TrimSuffix(name, ".txt.gz"), "-")
		if len(parts) < 3 {
			t.Fatal("invalid trace name")
		}
		iw, e1 := strconv.Atoi(parts[len(parts)-2])
		ow, e2 := strconv.Atoi(parts[len(parts)-1])
		if e1 != nil || e2 != nil || iw < 1 || ow < 1 || iw > 1<<20 || ow > 1<<20 {
			t.Fatal("invalid windows")
		}
		input, ok := corpus[strings.Join(parts[:len(parts)-2], "-")]
		if !ok {
			t.Fatal("missing trace input")
		}
		trace, err := gzip.NewReader(r)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) { calls += compareLZMATrace(t, input, iw, ow, trace) })
		if err := trace.Close(); err != nil {
			t.Fatal(err)
		}
		traces++
	}
	if len(corpus) != 857 || traces != 5142 || calls != 2842278 {
		t.Fatalf("reference inventory: inputs=%d traces=%d calls=%d", len(corpus), traces, calls)
	}
	t.Logf("replayed %d independent native decisions across %d traces", calls, traces)
}
