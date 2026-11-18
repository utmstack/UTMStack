package signatures

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Frozen C hashes cover formerly rejected axes, the profile boundary, both
// extended container forms, encoded transparency, and all five observed tiny numeric disagreements.
func TestWebPSmallNativeOracle(t *testing.T) {
	root := filepath.Join("testdata", "webp-small-native")
	manifest, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(manifest)) != "8afac2276c50536f2238efca528eb5c1d26791b79dd7981b8e0d27cc99c1fd31" {
		t.Fatal("native WebP manifest changed")
	}
	var rows []struct {
		File        string `json:"file"`
		Bytes       int64  `json:"bytes"`
		SHA256      string `json:"sha256"`
		NativeHash  string `json:"native_hash"`
		PrimaryHash string `json:"primary_hash"`
	}
	if err = json.Unmarshal(manifest, &rows); err != nil || len(rows) != 192 {
		t.Fatalf("native manifest count%d: %v", len(rows), err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "corpus.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "13725ee596893f629d1fb569c1eecefdd3537933ef0f496351ce4d7db6de07c0" {
		t.Fatal("native WebP corpus changed")
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for _, row := range rows {
		h, err := tr.Next()
		if err != nil || h.Name != row.File || h.Typeflag != tar.TypeReg || h.Size != row.Bytes {
			t.Fatalf("corpus member %s: %v", row.File, err)
		}
		data, err := io.ReadAll(io.LimitReader(tr, row.Bytes+1))
		if err != nil || int64(len(data)) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
			t.Fatalf("input identity %s: %v", row.File, err)
		}
		t.Run(row.File, func(t *testing.T) {
			reference, err := strconv.ParseUint(row.NativeHash, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			primary, err := strconv.ParseUint(row.PrimaryHash, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			if !webpLossyProfile(data) {
				t.Fatal("qualified small VP8 container rejected")
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !got.ok || got.hash != primary || got.uncertain != (primary^reference) {
				t.Fatalf("hash %+v primary%016x native%016x", got, primary, reference)
			}
			db := compileTest(t, "ldb", fmt.Sprintf("WebP.Native;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", reference))
			if got.uncertain == 0 {
				expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "WebP.Native", "")
			} else {
				expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
				other := compileTest(t, "ldb", fmt.Sprintf("WebP.Primary;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", primary))
				expectOutcome(t, other, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
			}
			certain := ^got.uncertain
			bit := certain & -certain
			if bit == 0 {
				t.Fatal("no certain negative control")
			}
			negative := compileTest(t, "ldb", fmt.Sprintf("WebP.Negative;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", reference^bit))
			expectOutcome(t, negative, string(data), MatchContext{Type: TypeGraphics}, "", "")
		})
	}
	if _, err = tr.Next(); err != io.EOF {
		t.Fatal("unexpected trailing corpus member")
	}
}
