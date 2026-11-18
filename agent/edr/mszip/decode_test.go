package mszip

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/mszip/msziptest"
)

func TestMSZIPStoredAndFixedFrames(t *testing.T) {
	for _, n := range []int{0, 1, 59, 32768} {
		child := bytes.Repeat([]byte("ordinary harmless MSZIP text "), (n+28)/29)[:n]
		for _, kind := range []string{"stored", "fixed"} {
			var data []byte
			if kind == "fixed" {
				data = msziptest.Encode(child, nil)
			} else {
				data = append([]byte{'C', 'K', 1, byte(n), byte(n >> 8), byte(^uint16(n)), byte(^uint16(n) >> 8)}, child...)
			}
			out, repaired, err := DecodeContext(context.Background(), bytes.NewReader(data), n)
			if err != nil || repaired != 0 || !bytes.Equal(out, child) {
				t.Fatalf("%s %d: bytes=%d repaired=%d error=%v", kind, n, len(out), repaired, err)
			}
			reader := flate.NewReader(bytes.NewReader(data[2:]))
			independent, err := io.ReadAll(reader)
			reader.Close()
			if err != nil || !bytes.Equal(independent, child) {
				t.Fatalf("invalid independent encoder: %s %d: %v", kind, n, err)
			}
		}
	}
}

func TestMSZIPRepairsIncompleteDistanceTree(t *testing.T) {
	child := []byte("If you can read this, the Quantum decompressor is working!\n")
	var encoded bytes.Buffer
	w, err := flate.NewWriterDict(&encoded, flate.BestCompression, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(child); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	data := append([]byte("CK"), encoded.Bytes()...)
	got, repaired, err := DecodeContext(context.Background(), bytes.NewReader(data), len(child))
	if err != nil || repaired != 1 || !bytes.Equal(got, make([]byte, len(child))) {
		t.Fatalf("repair: bytes=%x repaired=%d error=%v", got, repaired, err)
	}
}

func TestMSZIPHistoryAndZeroInitialWindow(t *testing.T) {
	first := bytes.Repeat([]byte("history!"), 4096)
	second := append(bytes.Clone(first[len(first)-1000:]), []byte("second frame")...)
	packed := append(msziptest.Encode(first, nil), msziptest.Encode(second, first)...)
	want := append(bytes.Clone(first), second...)
	got, repaired, err := DecodeContext(context.Background(), bytes.NewReader(packed), len(want))
	if err != nil || repaired != 0 || !bytes.Equal(got, want) {
		t.Fatalf("history: %d/%d repaired=%d error=%v", len(got), len(want), repaired, err)
	}
	got, repaired, err = DecodeContext(context.Background(), bytes.NewReader(msziptest.InitialMatch(258, 1)), 258)
	if err != nil || repaired != 0 || !bytes.Equal(got, make([]byte, 258)) {
		t.Fatalf("initial history: %d repaired=%d error=%v", len(got), repaired, err)
	}
}

type stopReader struct{}

func (stopReader) Read([]byte) (int, error) { return 0, ErrLimit }

func TestMSZIPBoundsCancellationAndInputFailure(t *testing.T) {
	if _, _, err := DecodeContext(context.Background(), bytes.NewReader(nil), -1); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, _, err := DecodeContext(ctx, bytes.NewReader([]byte("CK")), 100); len(got) != 0 || !errors.Is(err, context.Canceled) {
		t.Fatal(len(got), err)
	}
	if got, repaired, err := DecodeContext(context.Background(), stopReader{}, 100); len(got) != 0 || repaired != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(len(got), repaired, err)
	}
	if got, _, err := DecodeContext(context.Background(), bytes.NewReader(nil), 100); len(got) != 0 || !errors.Is(err, ErrInput) {
		t.Fatal(len(got), err)
	}
}

func FuzzMSZIP(f *testing.F) {
	f.Add([]byte("CK\x07"), uint16(100))
	f.Add(msziptest.Encode([]byte("ordinary safe compression test"), nil), uint16(30))
	for _, name := range []string{"text-32768-level6-valid.cab", "weighted-random-4096-level9-valid.cab", "all-bytes-4096-level6-cut-tail.cab"} {
		b, err := os.ReadFile(filepath.Join("../engine/testdata/mszip", name))
		if err != nil {
			f.Fatal(err)
		}
		at := int(binary.LittleEndian.Uint32(b[36:]))
		n := int(binary.LittleEndian.Uint16(b[at+4:]))
		f.Add(b[at+8:at+8+n], uint16(binary.LittleEndian.Uint32(b[44:])))
	}
	f.Fuzz(func(t *testing.T, data []byte, size uint16) {
		if len(data) > 65536 {
			t.Skip()
		}
		got, repaired, err := DecodeContext(context.Background(), bytes.NewReader(data), int(size))
		if len(got) > int(size) || repaired < 0 || err == nil && len(got) != int(size) {
			t.Fatalf("%d/%d repaired=%d error=%v", len(got), size, repaired, err)
		}
	})
}
