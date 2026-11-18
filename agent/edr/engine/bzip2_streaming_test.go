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

func TestBytecodeBzip2StreamingChildren(t *testing.T) {
	file, err := os.Open("testdata/bzip2_streaming.tar.gz")
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
	if len(fixtures) != 24 {
		t.Fatalf("fixture inventory %d", len(fixtures))
	}
	for _, window := range []string{"1", "5", "13", "64"} {
		t.Run(window, func(t *testing.T) {
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			engine := New(cfg)
			t.Cleanup(func() { _ = engine.Stop() })
			if _, err := engine.ActivateGeneration(map[string][]byte{"program.cbc": fixtures[window+"/program.cbc"], "child.ndb": fixtures[window+"/child.ndb"]}, 1, manifestHash("bzip2-stream-"+window)); err != nil {
				t.Fatal(err)
			}
			for _, sample := range []string{"sample-ab.bin", "sample-ac.bin", "sample-ab-long.bin", "sample-ac-long.bin"} {
				data := fixtures[window+"/"+sample]
				clean, name, _, err := engine.ScanSnapshot(context.Background(), data)
				want := ""
				if strings.HasPrefix(sample, "sample-ab") {
					want = "Child.Bzip2"
				}
				if err != nil || clean != (want == "") || name != want {
					t.Fatalf("%s: clean=%t name=%q err=%v want %q", sample, clean, name, err, want)
				}
			}
		})
	}
}
