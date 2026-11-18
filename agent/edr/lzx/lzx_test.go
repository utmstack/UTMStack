package lzx

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/lzx/lzxtest"
)

func sample(n int) []byte {
	r := rand.New(rand.NewSource(int64(n)))
	b := make([]byte, n)
	r.Read(b)
	return b
}

func TestLiteralsAcrossFrames(t *testing.T) {
	for _, n := range []int{1, 100, 32768, 32769, 100000} {
		data := sample(n)
		for _, bits := range []int{15, 16, 21} {
			got, err := Decode(lzxtest.Literals(data, bits), bits, n, 0)
			if err != nil || !bytes.Equal(got, data) {
				t.Fatalf("%d bytes, window %d: %d bytes, %v", n, bits, len(got), err)
			}
		}
	}
}

func TestUncompressedBlock(t *testing.T) {
	for _, n := range []int{7, 8, 5000} {
		data := sample(n)
		got, err := Decode(lzxtest.Uncompressed(data), 16, n, 0)
		if err != nil || !bytes.Equal(got, data) {
			t.Fatalf("%d bytes: %d, %v", n, len(got), err)
		}
	}
}

func TestDamage(t *testing.T) {
	data := sample(50000)
	stream := lzxtest.Literals(data, 16)
	// Cut: what decoded is returned; its last two bytes come from the zero
	// word the reader supplies once past the end.
	got, err := Decode(stream[:len(stream)/2], 16, len(data), 0)
	if !errors.Is(err, ErrTruncated) || len(got) < 2 || !bytes.Equal(got[:len(got)-2], data[:len(got)-2]) {
		t.Errorf("cut: %d bytes, %v", len(got), err)
	}
	// A bad block type.
	bad := append([]byte(nil), stream...)
	bad[1] |= 0x70 // the first word's top bits: E8 flag, then the block type
	if _, err := Decode(bad, 16, len(data), 0); err == nil {
		t.Error("block type 7 decoded")
	}
	if _, err := NewDecoder(stream, 22, 0); !errors.Is(err, ErrData) {
		t.Error("window of 22 bits accepted")
	}
}

func TestTranslateE8(t *testing.T) {
	// An absolute address 0x100 at position 10 becomes relative 0x100-10.
	b := make([]byte, 40)
	b[10] = 0xe8
	b[11] = 0x00
	b[12] = 0x01
	translateE8(b, 0, 1<<20)
	if got := int(b[11]) | int(b[12])<<8; got != 0x100-10 {
		t.Errorf("relative %#x", got)
	}
}

func FuzzDecode(f *testing.F) {
	f.Add(lzxtest.Literals(sample(3000), 16), 3000)
	f.Add(lzxtest.Uncompressed(sample(301)), 301)
	f.Fuzz(func(t *testing.T, b []byte, n int) {
		if n < 0 || n > 1<<18 {
			return
		}
		Decode(b, 16, n, 1)
	})
}
