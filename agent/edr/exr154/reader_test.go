package exr154

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenEXRIndependentRGB(t *testing.T) {
	var rows []struct {
		Name, SHA256         string
		Bytes, Width, Height int
		Hash                 *string
		RGB                  *struct {
			SHA256 string
			Bytes  int
		}
	}
	b, err := os.ReadFile("testdata/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 291 {
		t.Fatalf("manifest %v cases=%d", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("input identity %v", err)
			}
			d, err := Read(data, Limits{})
			if err != nil {
				if row.Hash != nil {
					t.Fatal(err)
				}
				return
			}
			cfg := d.Config()
			if row.RGB != nil && (cfg.Width != row.Width || cfg.Height != row.Height) {
				t.Fatalf("dimensions %+v", cfg)
			}
			img, err := d.Decode()
			if row.Hash == nil {
				if err == nil {
					t.Fatal("reference rejects this compressed image")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if row.RGB == nil {
				return
			}
			want, err := os.ReadFile(filepath.Join("testdata", row.Name[:len(row.Name)-4]+".rgb"))
			if err != nil || len(want) != row.RGB.Bytes || fmt.Sprintf("%x", sha256.Sum256(want)) != row.RGB.SHA256 {
				t.Fatalf("independent RGB identity %v", err)
			}
			rgba := img.(*image.NRGBA)
			got := make([]byte, 0, len(want))
			for y := 0; y < row.Height; y++ {
				for x := 0; x < row.Width; x++ {
					at := y*rgba.Stride + x*4
					got = append(got, rgba.Pix[at:at+3]...)
				}
			}
			if !bytes.Equal(got, want) {
				for i := range got {
					if got[i] != want[i] {
						t.Fatalf("RGB byte %d: got %d want %d", i, got[i], want[i])
					}
				}
			}
		})
	}
}
