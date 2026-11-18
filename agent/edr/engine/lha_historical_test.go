package engine

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

// These historical compressed streams exercise multi-symbol Huffman tables
// and long back-references beyond the synthetic method fixtures.
func TestLHAHistoricalStaticHuffmanChildHashes(t *testing.T) {
	for _, tc := range []struct {
		file, archiveSHA, childMD5 string
		childSize                  int
	}{
		{"lh5-lha213.lzh", "7f10d0be69536733217e984e4ff81d8980f2df06822581d7c006257bfad34851", "b234ee4d69f5fce4486a80fdaf4a4263", 18092},
		{"lh5-lha213-long.lzh", "8bc9c2c20bf12d0fa50a8ffbfae5e2fc7af88f63fc3be894fa2aff633db37877", "a6264a9c4ba8e0128f7d17b3f8c5f5c8", 1241658},
		{"lh6-amiga.lzh", "1910a6d4dfe5db3b1ceaf0fb6f2ec5f245d084cc5262c04777c1889e6791afb0", "b234ee4d69f5fce4486a80fdaf4a4263", 18092},
		{"lh4-amiga-long.lzh", "c48af76b5b2e9498fbf5cf26b8c3013637ca7bd10dea9a39df1a2be52714eef5", "a6264a9c4ba8e0128f7d17b3f8c5f5c8", 1241658},
		{"lh6-unix-long.lzh", "8f00d1a3197a193ed014054f0b2756930f60163c88b6f5cae8dbc3ecb26e77ae", "a6264a9c4ba8e0128f7d17b3f8c5f5c8", 1241658},
		{"lh7-unix-long.lzh", "37ffaebe08c9d8dcf8e9add232187bcdaa1313b13f7c22efe810b8bf883eaae5", "a6264a9c4ba8e0128f7d17b3f8c5f5c8", 1241658},
	} {
		t.Run(tc.file, func(t *testing.T) {
			archive, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.archiveSHA {
				t.Fatalf("historical archive changed: %s", got)
			}
			rules := map[string]string{"test.hdb": fmt.Sprintf("%s:%d:Reference.LHA.Historical\n", tc.childMD5, tc.childSize)}
			expectDetected(t, pipelineState(t, rules), archive, "Reference.LHA.Historical")
			expectScan(t, pipelineState(t, map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}), archive, "", "")
		})
	}
}
