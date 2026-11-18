package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestTIFFCMYKNativeOracle(t *testing.T) {
	root := filepath.Join("testdata", "tiff_cmyk")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Bytes        int
		Hash         *string
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 159 {
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

func TestTIFFCMYKIndependentRGB(t *testing.T) {
	root := filepath.Join("testdata", "tiff_cmyk")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name          string
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
		if row.RGB == nil {
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
	if count != 146 {
		t.Fatalf("independent RGB controls=%d", count)
	}
}

func TestTIFFProfileLimitsAndOriginalBytes(t *testing.T) {
	root := filepath.Join("testdata", "tiff_cmyk")
	for _, depth := range []int{8, 16} {
		for _, codec := range []int{1, 5, 8, 32773} {
			name := fmt.Sprintf("cmyk-%d-b-tiles-%d-p1-s0.tiff", depth, codec)
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			before := sha256.Sum256(data)
			ctx, cancel := context.WithCancel(context.Background())
			work := &matchWork{ctx: ctx, limit: 1 << 40}
			p, err := readTIFFProfile(data, work)
			if err != nil {
				t.Fatal(err)
			}
			cancel()
			if _, err := p.decode(); err == nil {
				t.Fatal("canceled decoder accepted")
			}
			work.ctx = context.Background()
			work.limit = work.steps + 1
			if _, err := p.decode(); err == nil || work.steps > work.limit {
				t.Fatal("limited decoder accepted or exceeded budget")
			}
			if sha256.Sum256(data) != before {
				t.Fatal("private TIFF view changed original bytes")
			}
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "cmyk-8-l-strips-1-p1-s1.tiff"))
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, bytes.Repeat([]byte{0x61}, 16<<20)...)
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	p, err := readTIFFProfile(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.decode()
	runtime.ReadMemStats(&after)
	if err != nil || after.TotalAlloc-before.TotalAlloc > 2<<20 {
		t.Fatalf("small image with large physical suffix: %v allocated=%d", err, after.TotalAlloc-before.TotalAlloc)
	}
}

func FuzzTIFFProfile(f *testing.F) {
	root := filepath.Join("testdata", "tiff_cmyk")
	for _, name := range []string{"cmyk-8-l-strips-1-p1-s1.tiff", "cmyk-16-b-tiles-5-p2-s0.tiff", "cmyk-8-b-strips-8-p2-s2.tiff", "cmyk-8-l-tiles-32773-p1-s1.tiff", "pixel-offset-into-shadow.tiff", "palette-8.tiff", "planar-separated.tiff"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		w := &matchWork{ctx: context.Background(), limit: 1 << 16}
		p, err := readTIFFProfile(data, w)
		if err != nil {
			return
		}
		cfg, err := p.config()
		if err != nil || cfg.Width > 64 || cfg.Height > 64 {
			return
		}
		_, _ = p.decode()
		if w.steps > w.limit {
			t.Fatal("budget exceeded")
		}
	})
}
