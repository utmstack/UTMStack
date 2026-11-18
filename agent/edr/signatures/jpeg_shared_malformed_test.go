package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// These same physical JPEG streams were checked by native ClamAV's standalone
// decoder as well as its distinct TIFF dependency. Missing DQT is a refusal;
// enough complete entropy data can remain hashable without a terminal EOI.
func TestJPEGSharedMalformedOracle(t *testing.T) {
	root := filepath.Join("..", "jpeg154", "testdata", "components")
	b, err := os.ReadFile(filepath.Join(root, "standalone_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Bytes        int
		Hash         *string
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 3 {
		t.Fatalf("standalone oracle: %v rows=%d", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("standalone JPEG identity: %v", err)
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if row.Hash == nil {
				if got.ok {
					t.Fatalf("native standalone rejects missing table, Go %+v", got)
				}
				return
			}
			want, err := strconv.ParseUint(*row.Hash, 16, 64)
			if err != nil || !got.ok || got.hash != want {
				t.Fatalf("standalone JPEG Go %+v native=%016x parse=%v", got, want, err)
			}
		})
	}
}
