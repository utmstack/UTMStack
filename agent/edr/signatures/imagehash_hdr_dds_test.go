package signatures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestHDRDDSNativeOracle(t *testing.T) {
	root := filepath.Join("testdata", "hdr_dds")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Hash         *string
		Bytes        int
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) != 178 {
		t.Fatalf("HDR/DDS oracle manifest: %v (%d cases)", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("HDR/DDS fixture identity: %v", err)
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if row.Hash == nil {
				if got.ok {
					t.Fatalf("native reader rejects this image, Go returned %+v", got)
				}
				referenceICOInvalid(t, data)
				return
			}
			want, err := strconv.ParseUint(*row.Hash, 16, 64)
			if err != nil || !got.ok || (got.hash^want)&^got.uncertain != 0 {
				t.Fatalf("HDR/DDS mismatch: Go %+v reference=%016x parse=%v", got, want, err)
			}
			wantUncertain := want ^ nativeWindowsImageHash(t, data)
			if got.uncertain != wantUncertain {
				t.Fatalf("native Windows disagreement must be covered exactly: %+v expected=%016x", got, wantUncertain)
			}
			if got.uncertain != 0 {
				t.Logf("HDR/DDS partial: hash=%016x uncertain=%016x reference=%016x", got.hash, got.uncertain, want)
			}
			if got.uncertain == ^uint64(0) {
				t.Fatal("all hash bits uncertain")
			}
			wrong := want ^ (uint64(1) << bits.TrailingZeros64(^got.uncertain))
			referenceImageRules(t, data, want, wrong, got.uncertain != 0)
		})
	}
}

func TestHDRDDSLimits(t *testing.T) {
	load := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", "hdr_dds", name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, name := range []string{"hdr-32x32-new-runs.hdr", "dds-dxt5-32x32-0.dds"} {
		t.Run(name, func(t *testing.T) {
			data := load(name)
			for _, work := range []*matchWork{
				{ctx: context.Background(), limit: 1},
				{ctx: ctx, limit: 1 << 40},
			} {
				if fuzzyImageHash(data, work).ok || work.steps > work.limit {
					t.Fatalf("limited/canceled image returned a hash or overspent: %+v", work)
				}
			}
		})
	}
	t.Run("zero-progress-hdr-streams", func(t *testing.T) {
		for _, modern := range []bool{false, true} {
			data := []byte{2, 3, 4, 128}
			if modern {
				data = append([]byte{2, 2, 0, 32}, make([]byte, 16384)...)
			} else {
				data = append(data, bytes.Repeat([]byte{1, 1, 1, 0}, 4096)...)
			}
			work := &matchWork{ctx: context.Background(), limit: 64}
			_, err := decodeHDR(data, hdrInfo{width: 32, height: 1}, work)
			if !errors.Is(err, ErrMatchLimit) || work.steps > work.limit {
				t.Fatalf("modern=%v: stream ignored budget: err=%v steps=%d", modern, err, work.steps)
			}
		}
	})
	t.Run("row-budgets", func(t *testing.T) {
		data := load("hdr-old-chain.hdr")
		info, err := readHDRHeader(data)
		if err != nil {
			t.Fatal(err)
		}
		work := &matchWork{ctx: context.Background(), limit: 100}
		if _, err := decodeHDR(data, info, work); !errors.Is(err, ErrMatchLimit) || work.steps > work.limit {
			t.Fatalf("compressed HDR bypassed row budget: %v steps=%d", err, work.steps)
		}
		data = load("dds-dxt5-32x32-0.dds")
		dds, err := readDDSHeader(data)
		if err != nil {
			t.Fatal(err)
		}
		work = &matchWork{ctx: context.Background(), limit: 10}
		if _, err := decodeDDS(data, dds, work); !errors.Is(err, ErrMatchLimit) || work.steps > work.limit {
			t.Fatalf("DDS bypassed row budget: %v steps=%d", err, work.steps)
		}
	})
	t.Run("dimension-bounds", func(t *testing.T) {
		data := load("hdr-32x32-raw.hdr")
		data = bytes.Replace(data, []byte("-Y 32 +X 32"), []byte("-Y 4294967295 +X 4294967295"), 1)
		if fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Fatal("oversized HDR returned a hash")
		}
		data = load("dds-dxt5-32x32-0.dds")
		binary.LittleEndian.PutUint32(data[12:16], 0xfffffffc)
		binary.LittleEndian.PutUint32(data[16:20], 0xfffffffc)
		if fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Fatal("oversized DDS returned a hash")
		}
		for _, bad := range []ddsInfo{{width: 1, height: 1}, {width: 4, height: 4}, {width: 4, height: 4, blockBytes: 7}} {
			if _, err := decodeDDS(nil, bad, nil); err == nil {
				t.Fatalf("invalid DDS layout accepted: %+v", bad)
			}
		}
	})
}

func TestHDRDDSDefaultTyping(t *testing.T) {
	db := compileTest(t, "ldb", "Untyped.Image;Engine:150-255,Target:0;0;fuzzy_img#938784356b1eb275\n")
	for _, name := range []string{"hdr-32x32-new-literal.hdr", "dds-dxt1-32x32-0.dds"} {
		data, err := os.ReadFile(filepath.Join("testdata", "hdr_dds", name))
		if err != nil {
			t.Fatal(err)
		}
		typ, complete := db.DetermineType(data)
		if !complete || typ != TypeBinaryData {
			t.Fatalf("default type of %s: %s complete=%t", name, typ, complete)
		}
		expectOutcome(t, db, string(data), MatchContext{Type: typ}, "", "")
	}
}

func TestHDRDDSMixedAxisResize(t *testing.T) {
	// This independently converted PPM has the same native hash as the HDR.
	// One native sinf-dependent resized pixel rounds to 135 rather than 134.
	data, err := os.ReadFile(filepath.Join("testdata", "hdr_dds", "resize-independent.ppm"))
	if err != nil {
		t.Fatal(err)
	}
	const native = uint64(0xd0d9f1e4c4d8a49b)
	const previous = uint64(0xd0d9f3e4c4d8a499)
	got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	wantUncertain := native ^ nativeWindowsImageHash(t, data)
	if !got.ok || got.hash != native || got.uncertain != wantUncertain {
		t.Fatalf("owned resize must preserve the original native hash and Windows difference: %+v expected=%016x", got, wantUncertain)
	}
	for _, hash := range []uint64{native, previous} {
		db := compileTest(t, "ldb", fmt.Sprintf("Resize.Partial;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
		if (hash^native)&^wantUncertain == 0 && wantUncertain != 0 {
			expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		} else if hash == native {
			expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "Resize.Partial", "")
		} else {
			expectOutcome(t, db, string(data), MatchContext{Type: TypeGraphics}, "", "")
		}
	}
	referenceImageRules(t, data, native, native^(uint64(1)<<63), wantUncertain != 0)
}

func FuzzHDRDDS(f *testing.F) {
	for _, name := range []string{"hdr-32x32-new-runs.hdr", "hdr-old-chain.hdr", "dds-dxt1-4x4-0.dds", "dds-dxt5-32x32-0.dds", "dds-dx10-71.dds"} {
		data, err := os.ReadFile(filepath.Join("testdata", "hdr_dds", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Add([]byte("#?RADIANCE\n\n-Y 1 +X 1\n"))
	f.Add([]byte("DDS "))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		work := &matchWork{ctx: context.Background(), limit: 1 << 20}
		fuzzyImageHash(data, work)
		if work.steps > work.limit {
			t.Fatal("image decoding overspent work budget")
		}
		if fuzzyImageHash(data, &matchWork{ctx: ctx, limit: 1 << 20}).ok {
			t.Fatal("canceled image decoding returned a hash")
		}
	})
}
