package engine

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestBytecodeLZMAStreamingChildren(t *testing.T) {
	file, err := os.Open("testdata/lzma_streaming.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	reader := tar.NewReader(z)
	fixtures := map[string][]byte{}
	for {
		h, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag != tar.TypeReg || h.Size < 0 || h.Size > 1<<20 {
			t.Fatal("invalid fixture member")
		}
		if _, ok := fixtures[h.Name]; ok {
			t.Fatal("duplicate fixture")
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		fixtures[h.Name] = data
	}
	if len(fixtures) != 32 {
		t.Fatalf("fixture inventory %d", len(fixtures))
	}
	for _, window := range []string{"1", "5", "13", "64"} {
		t.Run(window, func(t *testing.T) {
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			engine := New(cfg)
			t.Cleanup(func() { _ = engine.Stop() })
			if _, err := engine.ActivateGeneration(map[string][]byte{"program.cbc": fixtures[window+"/program.cbc"], "child.ndb": fixtures[window+"/child.ndb"]}, 1, manifestHash("lzma-stream-"+window)); err != nil {
				t.Fatal(err)
			}
			for _, sample := range []string{"sample-AB-known-false.bin", "sample-AC-known-false.bin", "sample-AB-known-true.bin", "sample-AC-known-true.bin", "sample-AB-long.bin", "sample-AC-long.bin"} {
				data := fixtures[window+"/"+sample]
				clean, name, _, err := engine.ScanSnapshot(context.Background(), data)
				want := ""
				if strings.HasPrefix(sample, "sample-AB") {
					want = "Child.LZMA"
				}
				if err != nil || clean != (want == "") || name != want {
					t.Fatalf("%s: clean=%t name=%q err=%v want %q", sample, clean, name, err, want)
				}
			}
		})
	}
}
