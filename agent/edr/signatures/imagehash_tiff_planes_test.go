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

func TestTIFFPlanesNativeOracle(t *testing.T) {
	root := filepath.Join("testdata", "tiff_planes")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Bytes        int
		Hash         *string
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 200 {
		t.Fatalf("TIFF manifest: %v cases=%d", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("TIFF identity: %v", err)
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
				t.Fatalf("TIFF hash Go %+v native=%016x parse=%v", got, want, err)
			}
			referenceICODirectory(t, data, want)
		})
	}
}

func TestTIFFPlanesIndependentRGB(t *testing.T) {
	root := filepath.Join("testdata", "tiff_planes")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name          string
		Hash          *string
		Width, Height int
		RGB           *struct {
			Bytes  int
			SHA256 string
		}
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, row := range rows {
		if row.Hash == nil || row.RGB == nil {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil {
				t.Fatal(err)
			}
			rgb, err := os.ReadFile(filepath.Join(root, row.Name[:len(row.Name)-5]+".rgb"))
			if err != nil || len(rgb) != row.RGB.Bytes || fmt.Sprintf("%x", sha256.Sum256(rgb)) != row.RGB.SHA256 {
				t.Fatalf("independent RGB identity: %v", err)
			}
			profile, err := readTIFFProfile(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if err != nil {
				t.Fatal(err)
			}
			img, err := profile.decode()
			if err != nil || img.Bounds().Dx() != row.Width || img.Bounds().Dy() != row.Height {
				t.Fatalf("TIFF decode: %v", err)
			}
			for y := 0; y < row.Height; y++ {
				for x := 0; x < row.Width; x++ {
					r, g, b := rgb8(img.At(x, y))
					i := (y*row.Width + x) * 3
					if byte(r) != rgb[i] || byte(g) != rgb[i+1] || byte(b) != rgb[i+2] {
						t.Fatalf("RGB(%d,%d)=%v,%v,%v want %v", x, y, r, g, b, rgb[i:i+3])
					}
				}
			}
		})
	}
	if count != 114 {
		t.Fatalf("independent RGB controls=%d", count)
	}
}
