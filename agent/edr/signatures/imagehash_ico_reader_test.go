package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestICOReaderNativeOracle(t *testing.T) {
	root := filepath.Join("testdata", "ico_reader")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256 string
		Hash         *string
		Bytes        int
	}
	if err := json.Unmarshal(b, &rows); err != nil || len(rows) < 80 {
		t.Fatalf("ICO reader oracle manifest: %v (%d cases)", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("ICO reader input identity: %v", err)
			}
			got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if row.Hash == nil {
				if got.ok {
					t.Fatalf("native decoder rejects this image, Go returned %+v", got)
				}
				referenceICOInvalid(t, data)
				return
			}
			want, err := strconv.ParseUint(*row.Hash, 16, 64)
			if err != nil || !got.ok || got.uncertain != 0 || got.hash != want {
				t.Fatalf("reader mismatch: Go %+v reference=%016x parse=%v", got, want, err)
			}
			referenceICODirectory(t, data, want)
		})
	}
}

func TestICOReaderAllocationAndWorkLimits(t *testing.T) {
	for _, name := range []string{"length-png-4294967295.ico", "length-dib-4294967295.ico"} {
		base, err := os.ReadFile(filepath.Join("testdata", "ico_reader", name))
		if err != nil {
			t.Fatal(err)
		}
		// A large unrelated physical suffix and a maximum declaration must
		// neither hide the selected image nor be copied into a new frame.
		data := append(base, make([]byte, 16<<20)...)
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		runtime.ReadMemStats(&after)
		if !got.ok || got.uncertain != 0 || got.hash != 0x87852e6dc81eb5a5 || after.TotalAlloc-before.TotalAlloc > 8<<20 {
			t.Fatalf("%s: hash %+v allocation %d", name, got, after.TotalAlloc-before.TotalAlloc)
		}
	}
	base, err := os.ReadFile(filepath.Join("testdata", "ico_reader", "rle-8-encoded.ico"))
	if err != nil {
		t.Fatal(err)
	}
	data := append([]byte(nil), base[:22+40+256*4]...)
	for i := 0; i < 4000; i++ {
		data = append(data, 255, 0)
	}
	data = append(data, 0, 1)
	data = append(data, make([]byte, 128)...)
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(data)-22))
	frame, ok := selectedICOImage(data)
	if !ok {
		t.Fatal("RLE selector")
	}
	info, err := readICOImage(frame)
	if err != nil {
		t.Fatal(err)
	}
	work := &matchWork{ctx: context.Background(), limit: 2}
	if _, err := info.decode(work); !errors.Is(err, ErrMatchLimit) || work.steps > work.limit {
		t.Fatalf("RLE instructions escaped budget: %v steps=%d", err, work.steps)
	}
	work = &matchWork{ctx: context.Background(), limit: 1}
	if got := fuzzyImageHash(base, work); got.ok || work.steps > work.limit {
		t.Fatalf("image escaped work limit: %+v", got)
	}
}
