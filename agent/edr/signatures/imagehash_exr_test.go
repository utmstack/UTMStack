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

func TestEXRNativeOracle(t *testing.T) {
	root := filepath.Join("..", "exr154", "testdata")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Bytes        int
		Hash         *string
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 291 {
		t.Fatalf("EXR manifest: %v cases=%d", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("EXR identity: %v", err)
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if row.Hash == nil {
				if got.ok {
					t.Fatalf("native reader rejects, Go %+v", got)
				}
				referenceICOInvalid(t, data)
				return
			}
			want, err := strconv.ParseUint(*row.Hash, 16, 64)
			if err != nil || !got.ok || got.hash != want || got.uncertain != 0 {
				t.Fatalf("EXR hash Go %+v native=%016x parse=%v", got, want, err)
			}
			referenceICODirectory(t, data, want)
		})
	}
}

func TestEXRDefaultTypingAndLimits(t *testing.T) {
	db := compileTest(t, "ldb", "EXR.Default;Engine:150-255,Target:0;0;fuzzy_img#a1862e6fc81fe40f\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, name := range []string{"codec-4-type-0-layout-0.exr", "multipart-gray-rgb.exr"} {
		data, err := os.ReadFile(filepath.Join("..", "exr154", "testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		typ, complete := db.DetermineType(data)
		if !complete || typ != TypeBinaryData {
			t.Fatalf("builtin EXR typing: %s complete=%t", typ, complete)
		}
		expectOutcome(t, db, string(data), MatchContext{Type: typ}, "", "")
		for _, work := range []*matchWork{{ctx: ctx, limit: 1 << 40}, {ctx: context.Background(), limit: 1}} {
			if got := fuzzyImageHash(data, work); got.ok || work.steps > work.limit {
				t.Fatalf("canceled/limited EXR hash: %+v steps=%d", got, work.steps)
			}
		}
	}
}
