package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTIFFFloat32IndependentConversion(t *testing.T) {
	root := filepath.Join("testdata", "tiff_float")
	samples, err := os.ReadFile(filepath.Join(root, "float32_samples.bin"))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join(root, "float32_rgb8.bin"))
	if err != nil || len(samples) != 4*34057 || len(expected) != 34057 ||
		fmt.Sprintf("%x", sha256.Sum256(samples)) != "77287264872fb2862b8428ee5b949c0c390783782b21afe397dbfe6e44e47167" ||
		fmt.Sprintf("%x", sha256.Sum256(expected)) != "003069e89bcfb6db9409913a52f85987a63047dfdeadcf0a7969ea802a97055e" {
		t.Fatalf("independent float table identity: %v", err)
	}
	for i, want := range expected {
		bits := binary.LittleEndian.Uint32(samples[4*i:])
		if got := tiffFloatRGB8(math.Float32frombits(bits)); got != want {
			t.Fatalf("float32 %08x: RGB8=%d want=%d", bits, got, want)
		}
	}
}

type tiffFloatCancelContext struct {
	context.Context
	cancel context.CancelFunc
	calls  int
}

func (c *tiffFloatCancelContext) Err() error {
	c.calls--
	if c.calls <= 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestTIFFFloatWorkCancellationAndInput(t *testing.T) {
	root := filepath.Join("testdata", "tiff_float")
	for _, channels := range []int{3, 4} {
		for _, pair := range [][2]int{{1, 1}, {5, 2}, {5, 3}, {8, 3}, {32773, 1}} {
			name := fmt.Sprintf("float-%d-b-tiles-%d-p%d-s0.tiff", channels, pair[0], pair[1])
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			before := sha256.Sum256(data)
			ctx, cancel := context.WithCancel(context.Background())
			work := &matchWork{ctx: ctx, limit: 1 << 40}
			p, err := readTIFFProfile(data, work)
			if err != nil || p.floating == nil {
				cancel()
				t.Fatalf("%s profile: %v", name, err)
			}
			// Cancel after a few checks inside actual decoding, not header parsing.
			work.ctx = &tiffFloatCancelContext{Context: ctx, cancel: cancel, calls: 5}
			if _, err := p.decode(); err == nil || ctx.Err() != context.Canceled {
				cancel()
				t.Fatalf("%s mid-decode cancellation: %v", name, err)
			}
			cancel()
			work.ctx = context.Background()
			work.limit = work.steps + 1
			if _, err := p.decode(); err == nil || work.steps > work.limit {
				t.Fatalf("%s budget: %v steps=%d limit=%d", name, err, work.steps, work.limit)
			}
			if sha256.Sum256(data) != before {
				t.Fatalf("%s changed original bytes", name)
			}
		}
	}
}

func TestTIFFFloatBoundsAndSuffixAllocation(t *testing.T) {
	root := filepath.Join("testdata", "tiff_float")
	base, err := os.ReadFile(filepath.Join(root, "float-3-l-tiles-1-p1-s1.tiff"))
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
		t.Fatalf("small float image with large suffix: %v allocated=%d", err, after.TotalAlloc-before.TotalAlloc)
	}
}

func FuzzTIFFFloatDecoder(f *testing.F) {
	root := filepath.Join("testdata", "tiff_float")
	for _, name := range []string{"float-3-l-strips-1-p1-s1.tiff", "float-4-b-tiles-5-p3-s0.tiff", "float-3-b-strips-8-p2-s2.tiff", "float-4-l-tiles-32773-p1-s1.tiff", "padding-p3-kept-192.tiff", "padding-p3-kept-60.tiff", "deflate-bad-checksum.tiff", "meta-bits-single.tiff"} {
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
		if err != nil || p.floating == nil {
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
