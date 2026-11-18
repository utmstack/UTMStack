package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestICOGeneralDIBReference(t *testing.T) {
	root := filepath.Join("testdata", "ico_general")
	manifest, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, SHA256, Hash string
		Bytes              int
	}
	if err := json.Unmarshal(manifest, &rows); err != nil || len(rows) != 118 {
		t.Fatalf("oracle manifest: %v (%d cases)", err, len(rows))
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, row.Name))
			if err != nil || len(data) != row.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
				t.Fatalf("oracle input identity: %v", err)
			}
			want, err := strconv.ParseUint(row.Hash, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); !got.ok || got.uncertain != 0 || got.hash != want {
				t.Fatalf("general DIB hash: %+v reference=%016x", got, want)
			}
			referenceICODirectory(t, data, want)
		})
	}
}

func TestICOGeneralOverlappingAlpha(t *testing.T) {
	data := extendedBitfieldDIBIconProbe(t, 32, 16, 56)
	binary.LittleEndian.PutUint32(data[22+52:], 1)
	// The reference permits overlap and fuzzy hashing drops alpha, so this
	// valid channel must not suppress an otherwise matching RGB image.
	referenceICODirectory(t, data, 0x88ec267b3e95c0b5)
}

func TestICOGeneralReinterpretedDepth(t *testing.T) {
	data := dibIconProbe(t, 32, 32, true, 0)
	binary.LittleEndian.PutUint16(data[22+14:], 8)
	got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
	const want = uint64(0x9bd6a370429c57aa) // Pinned native oracle in reinterpreted-depth.txt.
	if !got.ok || got.uncertain != 0 || got.hash != want {
		t.Fatalf("supported indexed interpretation: %+v", got)
	}
	referenceICODirectory(t, data, want)
}

func TestICOGeneralOptionalMaskAndAdvisoryFields(t *testing.T) {
	root := filepath.Join("testdata", "ico_general")
	manifest, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Name, Hash string }
	if err := json.Unmarshal(manifest, &rows); err != nil {
		t.Fatal(err)
	}
	wants := make(map[string]string)
	for _, row := range rows {
		wants[row.Name] = row.Hash
	}
	for _, file := range []string{"rgb-16-108-extra-1-gradient.ico", "rgb-24-52-extra-0-gradient.ico", "rgb-32-52-extra-0-gradient.ico", "bitfields-argb4444-56-gradient.ico", "bitfields-rgb1010102-124-gradient.ico"} {
		base, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		want, err := strconv.ParseUint(wants[file], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"no-AND", "image-size-one", "image-size-max", "important-colors", "unused-colors-used", "extra-padding"} {
			t.Run(file+"/"+mode, func(t *testing.T) {
				data := append([]byte(nil), base...)
				switch mode {
				case "no-AND":
					data = data[:len(data)-128]
					binary.LittleEndian.PutUint32(data[14:18], uint32(len(data)-22))
				case "image-size-one":
					binary.LittleEndian.PutUint32(data[22+20:], 1)
				case "image-size-max":
					binary.LittleEndian.PutUint32(data[22+20:], 0xffffffff)
				case "important-colors":
					binary.LittleEndian.PutUint32(data[22+36:], 0xffffffff)
				case "unused-colors-used":
					binary.LittleEndian.PutUint32(data[22+32:], 1234)
				case "extra-padding":
					data = append(data, []byte("frame padding")...)
					binary.LittleEndian.PutUint32(data[14:18], uint32(len(data)-22))
				}
				referenceICODirectory(t, data, want)
			})
		}
		partial := append([]byte(nil), base[:len(base)-1]...)
		binary.LittleEndian.PutUint32(partial[14:18], uint32(len(partial)-22))
		if got := fuzzyImageHash(partial, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
			t.Fatalf("partial AND mask was accepted: %s", file)
		}
		referenceICOInvalid(t, partial)
	}
	data := indexedDIBIconProbe(t, 32, 8)
	binary.LittleEndian.PutUint32(data[22+32:], 0)
	referenceICODirectory(t, data, 0xf8f2d7531351d510)
}

func TestICOGeneralBitfieldBounds(t *testing.T) {
	path := filepath.Join("testdata", "ico_general", "bitfields-argb4444-56.ico")
	base, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"zero-red", "noncontiguous-red", "oversized-red", "noncontiguous-alpha", "oversized-alpha", "short-masks", "short-pixels", "short-AND"} {
		t.Run(name, func(t *testing.T) {
			data := append([]byte(nil), base...)
			switch name {
			case "zero-red":
				binary.LittleEndian.PutUint32(data[22+40:], 0)
			case "noncontiguous-red":
				binary.LittleEndian.PutUint32(data[22+40:], 0x1500)
			case "oversized-red":
				binary.LittleEndian.PutUint32(data[22+40:], 1<<20)
			case "noncontiguous-alpha":
				binary.LittleEndian.PutUint32(data[22+52:], 5)
			case "oversized-alpha":
				binary.LittleEndian.PutUint32(data[22+52:], 1<<16)
			case "short-masks":
				data = data[:22+51]
			case "short-pixels":
				data = data[:len(data)/2]
			case "short-AND":
				data = data[:len(data)-1]
			}
			if got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); got.ok {
				t.Fatalf("invalid bitfields became a definite image: %+v", got)
			}
			referenceICOInvalid(t, data)
		})
	}
}

func FuzzICODirectory(f *testing.F) {
	for _, name := range []string{"bitfields-argb4444-56.ico", "bitfields-rgb1010102-124.ico", "rgb-16-108-extra-1.ico"} {
		data, err := os.ReadFile(filepath.Join("testdata", "ico_general", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	for _, name := range []string{"header-12-depth-8.ico", "rle-8-overrun-noop.ico", "png-rgba-adam7-1.ico"} {
		data, err := os.ReadFile(filepath.Join("testdata", "ico_reader", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Add([]byte{0, 0, 1, 0, 255, 255})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 20})
		if got := fuzzyImageHash(data, &matchWork{ctx: ctx, limit: 1 << 20}); got.ok {
			t.Fatalf("canceled image decoding produced a definite hash: %+v", got)
		}
	})
}
