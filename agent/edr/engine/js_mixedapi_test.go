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

// The independent serialized program and samples were executed by pinned
// ClamAV, including file-pipe boundaries and disabled logical triggers.
func TestMixedDecoderBytecodeEngineReference(t *testing.T) {
	f, err := os.Open("testdata/mixed_jsnorm_api.tar.gz")
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
	files := map[string][]byte{}
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag != tar.TypeReg || h.Size < 0 || h.Size > 1<<20 {
			t.Fatal("invalid native fixture")
		}
		if _, exists := files[h.Name]; exists {
			t.Fatal("duplicate native fixture")
		}
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		files[h.Name] = data
	}
	if len(files) != 122 {
		t.Fatalf("native fixture inventory %d", len(files))
	}
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	engine := New(cfg)
	t.Cleanup(func() { _ = engine.Stop() })
	if _, err := engine.ActivateGeneration(map[string][]byte{"program.cbc": files["program.cbc"], "child.ndb": files["child.ndb"]}, 1, manifestHash("mixed-jsnorm-api")); err != nil {
		t.Fatal(err)
	}
	cases := 0
	for name, data := range files {
		if !strings.HasSuffix(name, ".bin") {
			continue
		}
		cases++
		t.Run(name, func(t *testing.T) {
			clean, signature, _, err := engine.ScanSnapshot(context.Background(), data)
			want := ""
			if strings.HasSuffix(name, "-positive.bin") {
				want = "Decoded.Mixed.API"
			}
			if err != nil || clean != (want == "") || signature != want {
				t.Fatalf("clean=%t signature=%q err=%v want=%q", clean, signature, err, want)
			}
		})
	}
	if cases != 120 {
		t.Fatalf("native sample inventory %d", cases)
	}
}
