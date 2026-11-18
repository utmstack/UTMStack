package unpack

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

// lzmaRaw encodes data with an independent encoder and returns the raw
// stream (its 13-byte header removed), with no end marker.
func lzmaRaw(t *testing.T, data []byte, p LZMAProps) []byte {
	var buf bytes.Buffer
	w, err := lzma.WriterConfig{Properties: &lzma.Properties{LC: p.LC, LP: p.LP, PB: p.PB},
		DictCap: 1 << 20, Size: int64(len(data))}.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()[13:]
}

// lzmaSample is data with literals, short and long matches and repeats.
func lzmaSample(r *rand.Rand, n int) []byte {
	words := []string{"alpha ", "bravo ", "charlie ", "MZ\x90\x00", "\x00\x00\x00\x00", "PE\x00\x00L\x01"}
	var b []byte
	for len(b) < n {
		switch r.Intn(4) {
		case 0:
			b = append(b, byte(r.Intn(256)))
		case 1:
			b = append(b, words[r.Intn(len(words))]...)
		case 2:
			if len(b) > 64 {
				at := r.Intn(len(b) - 32)
				b = append(b, b[at:at+r.Intn(32)]...)
			}
		default:
			b = append(b, bytes.Repeat([]byte{byte(r.Intn(4))}, r.Intn(300))...)
		}
	}
	return b[:n]
}

func TestDecompressLZMARoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for i, p := range []LZMAProps{{3, 0, 2}, {0, 0, 0}, {8, 4, 4}, {1, 2, 3}, {4, 0, 2}} {
		for _, n := range []int{1, 100, 5000, 200000} {
			data := lzmaSample(r, n)
			raw := lzmaRaw(t, data, p)
			got, err := DecompressLZMA(raw, p, len(data))
			if err != nil || !bytes.Equal(got, data) {
				t.Fatalf("case %d, %d bytes: %v, equal %v", i, n, err, bytes.Equal(got, data))
			}
			// packers store the stream with other bytes after it: decoding
			// stops at the size
			got, err = DecompressLZMA(append(append([]byte(nil), raw...), lzmaSample(r, 64)...), p, len(data))
			if err != nil || !bytes.Equal(got, data) {
				t.Fatalf("case %d, %d bytes, trailing data: %v", i, n, err)
			}
			// a smaller size gives a prefix
			if n > 10 {
				got, err = DecompressLZMA(raw, p, n-7)
				if err != nil || !bytes.Equal(got, data[:n-7]) {
					t.Fatalf("case %d, %d bytes, shorter size: %v", i, n, err)
				}
			}
		}
	}
}

func TestDecompressLZMAFailures(t *testing.T) {
	p := LZMAProps{3, 0, 2}
	data := lzmaSample(rand.New(rand.NewSource(4)), 20000)
	raw := lzmaRaw(t, data, p)
	if _, err := DecompressLZMA(raw[:len(raw)/2], p, len(data)); !errors.Is(err, ErrLZMAInput) {
		t.Fatalf("cut stream: %v", err)
	}
	bad := append([]byte{1}, raw[1:]...)
	if _, err := DecompressLZMA(bad, p, len(data)); !errors.Is(err, ErrLZMAData) {
		t.Fatalf("first byte not zero: %v", err)
	}
	if _, err := DecompressLZMA(raw, LZMAProps{9, 0, 2}, len(data)); !errors.Is(err, ErrLZMAData) {
		t.Fatalf("literal context bits above 8: %v", err)
	}
	// Random streams never decode past an error into a panic.
	r := rand.New(rand.NewSource(5))
	for i := 0; i < 2000; i++ {
		b := make([]byte, 1+r.Intn(400))
		r.Read(b)
		b[0] = 0
		DecompressLZMA(b, p, 1+r.Intn(5000))
	}
}
