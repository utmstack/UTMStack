package jpeg154

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTIFFComponentsIndependentOracle(t *testing.T) {
	root := filepath.Join("testdata", "components")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, Colorspace        string
		Accepted                bool
		Width, Height, Channels int
		JPEG, Raw               struct {
			Bytes  int
			SHA256 string
		}
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 273 {
		t.Fatalf("component manifest: %v rows=%d", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.JPEG.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.JPEG.SHA256 {
				t.Fatalf("JPEG identity: %v", err)
			}
			if !row.Accepted {
				if _, err := DecodeTIFFComponents(bytes.NewReader(data), nil); err == nil {
					t.Fatal("locked native JPEG decoder rejects this stream")
				}
				return
			}
			raw, err := os.ReadFile(filepath.Join(root, row.Name[:len(row.Name)-4]+".raw"))
			if err != nil || len(raw) != row.Raw.Bytes || fmt.Sprintf("%x", sha256.Sum256(raw)) != row.Raw.SHA256 {
				t.Fatalf("raw identity: %v", err)
			}
			got, err := DecodeTIFFComponents(bytes.NewReader(data), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got.Width != row.Width || got.Height != row.Height || got.Channels != row.Channels || !bytes.Equal(got.Pix, raw) {
				for i := range raw {
					if i >= len(got.Pix) || got.Pix[i] != raw[i] {
						t.Fatalf("raw %s byte %d: dimensions=%dx%dx%d want=%dx%dx%d; byte Go=%v native=%v", row.Colorspace, i, got.Width, got.Height, got.Channels, row.Width, row.Height, row.Channels, got.Pix[min(i, len(got.Pix)):min(i+1, len(got.Pix))], raw[i])
					}
				}
				t.Fatalf("component dimensions/length differ: %dx%dx%d bytes=%d want=%d", got.Width, got.Height, got.Channels, len(got.Pix), len(raw))
			}
		})
	}
}
