package exr154

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHalfRGB8Independent(t *testing.T) {
	want, err := os.ReadFile("testdata/half_rgb8.bin")
	if err != nil || len(want) != 65536 {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(want)) != "e4941ba773dfb43adc14ae4f58275e276a23f860c8dbaa2ec23d0989d884af7c" {
		t.Fatal("half golden identity")
	}
	var b [2]byte
	for i, v := range want {
		binary.LittleEndian.PutUint16(b[:], uint16(i))
		if got := sample8(b[:], 1); got != v {
			t.Fatalf("half bits=%04x got=%d want=%d", i, got, v)
		}
	}
	t.Logf("verified all 65,536 binary16 patterns; golden SHA256=%x", sha256.Sum256(want))
}
func TestDecoderLimitsAndCancellation(t *testing.T) {
	limitError := errors.New("test EXR work limit")
	for codec := 0; codec < 8; codec++ {
		t.Run(fmt.Sprint(codec), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("codec-%d-type-0-layout-0.exr", codec)))
			if err != nil {
				t.Fatal(err)
			}
			used := int64(0)
			step := func(n int64) error {
				if n < 0 || n > 512-used {
					return limitError
				}
				used += n
				return nil
			}
			d, err := Read(data, Limits{Step: step})
			if err == nil {
				_, err = d.Decode()
			}
			if !errors.Is(err, limitError) || used > 512 {
				t.Fatalf("work was not bounded: err=%v steps=%d", err, used)
			}
			if _, err := Read(data, Limits{MaxDimension: 16}); !errors.Is(err, ErrLimit) {
				t.Fatalf("dimensions not bounded: %v", err)
			}
			if _, err := Read(data, Limits{MaxPixels: 32}); !errors.Is(err, ErrLimit) {
				t.Fatalf("pixels not bounded: %v", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			calls := 0
			d, err = Read(data, Limits{Step: func(n int64) error {
				calls++
				if calls > 20 {
					cancel()
				}
				return ctx.Err()
			}})
			if err == nil {
				_, err = d.Decode()
			}
			cancel()
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v calls=%d", err, calls)
			}
		})
	}
	// Oversized physical block declarations cannot create buffers or slices.
	data, err := os.ReadFile("testdata/codec-0-type-0-layout-0.exr")
	if err != nil {
		t.Fatal(err)
	}
	d, err := Read(data, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	h := d.headers[d.selected]
	at := int(binary.LittleEndian.Uint64(data[h.table:]))
	binary.LittleEndian.PutUint32(data[at+4:], 0x7fffffff)
	if _, err := d.Decode(); err == nil {
		t.Fatal("oversized block declaration decoded")
	}
	t.Run("physical-suffix-allocation", func(t *testing.T) {
		base, err := os.ReadFile("testdata/codec-0-type-0-layout-0.exr")
		if err != nil {
			t.Fatal(err)
		}
		base = append(base, make([]byte, 16<<20)...)
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		reader, err := Read(base, Limits{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := reader.Decode(); err != nil {
			t.Fatal(err)
		}
		runtime.ReadMemStats(&after)
		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8<<20 {
			t.Fatalf("physical suffix was copied: %d allocated bytes", allocated)
		}
	})
}
func FuzzReadDecoder(f *testing.F) {
	names := []string{"piz-wide-range.exr", "multipart-gray-rgb.exr", "linear-b44-7.exr", "pxr24-special-floats.exr", "multipart-deep-rgb.exr", "round-up-layout-3-type-0.exr", "meta-raw-fallback-11.exr"}
	for codec := 0; codec < 8; codec++ {
		names = append(names, fmt.Sprintf("codec-%d-type-0-layout-1.exr", codec))
	}
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	f.Add([]byte{0x76, 0x2f, 0x31, 1, 2, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		used := int64(0)
		d, err := Read(data, Limits{MaxDimension: 512, MaxPixels: 65536, Step: func(n int64) error {
			if n < 0 || n > (1<<20)-used {
				return ErrLimit
			}
			used += n
			return nil
		}})
		if err == nil {
			_, _ = d.Decode()
		}
		if used > 1<<20 {
			t.Fatal("work budget overspent")
		}
		d, err = Read(data, Limits{Step: func(int64) error { return context.Canceled }})
		if err == nil {
			if _, err := d.Decode(); err == nil {
				t.Fatal("canceled decoding returned an image")
			}
		}
	})
}

func TestLargeTypeAttributeAllocation(t *testing.T) {
	const size = 16 << 20
	prefix := []byte("type\x00string\x00")
	data := make([]byte, 8+len(prefix)+4+size+1)
	copy(data, []byte{0x76, 0x2f, 0x31, 1, 2, 0, 0, 0})
	copy(data[8:], prefix)
	at := 8 + len(prefix)
	binary.LittleEndian.PutUint32(data[at:], size)
	for i := at + 4; i < len(data)-1; i++ {
		data[i] = 'x'
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, err := Read(data, Limits{}); err == nil {
		t.Fatal("unknown huge part type accepted")
	}
	runtime.ReadMemStats(&after)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8<<20 {
		t.Fatalf("part type copied before rejection: %d bytes", allocated)
	}
}
