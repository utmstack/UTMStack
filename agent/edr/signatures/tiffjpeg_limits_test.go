package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTIFFJPEGWorkCancellationAndInput(t *testing.T) {
	root := filepath.Join("testdata", "tiff_jpeg")
	names := []string{"jpeg-1-b-strips-t3-q31-s0.tiff", "jpeg-2-l-tiles-t0-q90-s0.tiff", "jpeg-5-b-tiles-t1-q90-s1.tiff", "pillow-RGB-p1-sampling-2.tiff", "ycbcr-as-rgb-strips-s0.tiff", "reinterpret-32-p3.tiff", "body-length-642.tiff"}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		before := sha256.Sum256(data)
		ctx, cancel := context.WithCancel(context.Background())
		work := &matchWork{ctx: ctx, limit: 1 << 40}
		p, err := readTIFFProfile(data, work)
		if err != nil || p.jpeg == nil {
			cancel()
			t.Fatalf("%s profile: %v", name, err)
		}
		work.ctx = &tiffFloatCancelContext{Context: ctx, cancel: cancel, calls: 5}
		if _, err := p.decode(); err == nil || ctx.Err() != context.Canceled {
			cancel()
			t.Fatalf("%s mid-decode cancellation: %v", name, err)
		}
		cancel()
		work.ctx = context.Background()
		work.limit = work.steps + 1
		if _, err := p.decode(); err == nil || work.steps > work.limit {
			t.Fatalf("%s budget: %v", name, err)
		}
		if sha256.Sum256(data) != before {
			t.Fatalf("%s changed input", name)
		}
	}
}

func TestTIFFJPEGBoundsAndSuffixAllocation(t *testing.T) {
	root := filepath.Join("testdata", "tiff_jpeg")
	base, err := os.ReadFile(filepath.Join(root, "jpeg-2-l-tiles-t3-q90-s1.tiff"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []uint16{256, 322} {
		data := append([]byte(nil), base...)
		at := int(binary.LittleEndian.Uint32(data[4:]))
		for i := 0; i < int(binary.LittleEndian.Uint16(data[at:])); i++ {
			e := data[at+2+12*i:]
			if binary.LittleEndian.Uint16(e) == tag {
				binary.LittleEndian.PutUint16(e[2:], 4)
				binary.LittleEndian.PutUint32(e[4:], 1)
				binary.LittleEndian.PutUint32(e[8:], 1<<28)
			}
		}
		if _, err := readTIFFProfile(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); err == nil {
			t.Fatalf("oversized image/tile tag %d accepted", tag)
		}
	}
	data := append(base, bytes.Repeat([]byte{0x61}, 16<<20)...)
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
		t.Fatalf("small JPEG image with large suffix: %v allocated=%d", err, after.TotalAlloc-before.TotalAlloc)
	}
}

func FuzzTIFFJPEGDecoder(f *testing.F) {
	root := filepath.Join("testdata", "tiff_jpeg")
	for _, name := range []string{"jpeg-2-l-tiles-t0-q90-s0.tiff", "jpeg-5-b-tiles-t1-q90-s1.tiff", "pillow-RGB-p1-sampling-2.tiff", "ycbcr-as-rgb-strips-s0.tiff", "reinterpret-32-p3.tiff", "body-length-642.tiff", "table-mode-1-removed.tiff", "table-short-1.tiff", "sampling-4x2-p0-r0-s0.tiff"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	for _, name := range []string{"planes-0-p1-r0-s0.tiff", "planes-6-p0-r2-s0.tiff", "planes-7-p1-r0-s0.tiff", "planes-9-p0-r0-s0.tiff", "planes-11-p0-r2-s1.tiff", "planes-12-p1-r0-s0.tiff", "planes-4-p0-r0-s0.tiff", "planes-17-p1-r0-s0.tiff"} {
		data, err := os.ReadFile(filepath.Join("testdata", "tiff_planes", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		work := &matchWork{ctx: context.Background(), limit: 1 << 16}
		p, err := readTIFFProfile(data, work)
		if err != nil || p.jpeg == nil {
			return
		}
		cfg, err := p.config()
		if err != nil || cfg.Width > 64 || cfg.Height > 64 {
			return
		}
		_, _ = p.decode()
		if work.steps > work.limit {
			t.Fatal("budget exceeded")
		}
	})
}
