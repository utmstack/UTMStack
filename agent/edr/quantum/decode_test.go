package quantum

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"
)

// Static Quantum payload from libmspack's mixed CAB regression fixture.
// Its expected text and full-child hash are independently checked by ClamAV.
const knownBlockHex = "d606690bcb47f02c2a3a8f2cabbb3cb933018bd8584b7b01ba6f6d516e3ac367424beb023643d66656ca9e72cc300000"

func TestKnownQuantumBlock(t *testing.T) {
	b, err := hex.DecodeString(knownBlockHex)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("If you can read this, the Quantum decompressor is working!\n")
	got, err := DecodeContext(context.Background(), [][]byte{b}, 18, len(want))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("Quantum decoded %q: %v", got, err)
	}
	for _, n := range []int{0, 1, 3, 20, len(want)} {
		got, err := DecodeContext(context.Background(), [][]byte{b}, 18, n)
		if err != nil || !bytes.Equal(got, want[:n]) {
			t.Fatalf("Quantum prefix %d decoded %q: %v", n, got, err)
		}
	}
}

func TestQuantumBoundsAndCancellation(t *testing.T) {
	for _, window := range []int{-1, 0, 9, 22, 64} {
		if _, err := DecodeContext(context.Background(), nil, window, 1); !errors.Is(err, ErrFormat) {
			t.Fatalf("window %d: %v", window, err)
		}
	}
	if _, err := DecodeContext(context.Background(), nil, 10, -1); !errors.Is(err, ErrFormat) {
		t.Fatalf("negative size: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DecodeContext(ctx, nil, 18, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled decoder: %v", err)
	}
	if _, err := DecodeContext(context.Background(), [][]byte{{}}, 18, 100); err == nil {
		t.Fatal("empty block manufactured output from unlimited padding")
	}
}

func TestQuantumValidatedPrefixBeforeLaterFailure(t *testing.T) {
	b, _ := hex.DecodeString(knownBlockHex)
	got, produced, err := DecodeValidatedContext(context.Background(), [][]byte{b}, 18, 16<<20)
	want := []byte("If you can read this, the Quantum decompressor is working!\n")
	if !errors.Is(err, ErrFormat) || len(got) < len(want) || !bytes.Equal(got[:len(want)], want) || produced < len(got) {
		t.Fatalf("validated prefix: %d/%d bytes, %v", len(got), produced, err)
	}
	if cap(got) > 8192 {
		t.Fatalf("short retained prefix holds capacity for an untrusted large request: %d", cap(got))
	}
	for _, window := range []int{9, 22} {
		if got, produced, err := DecodeValidatedContext(context.Background(), nil, window, 100); len(got) != 0 || produced != 0 || !errors.Is(err, ErrFormat) {
			t.Fatalf("invalid validated decoder: %d/%d bytes, %v", len(got), produced, err)
		}
	}
}

func FuzzQuantum(f *testing.F) {
	b, _ := hex.DecodeString(knownBlockHex)
	f.Add(b, uint8(18), uint16(59))
	f.Add([]byte{}, uint8(10), uint16(100))
	f.Fuzz(func(t *testing.T, data []byte, window uint8, size uint16) {
		if len(data) > 65536 {
			t.Skip()
		}
		got, produced, err := DecodeValidatedContext(context.Background(), [][]byte{data}, int(window%12)+10, int(size))
		if len(got) > produced || produced > int(size) || err == nil && len(got) != int(size) {
			t.Fatalf("decoded length %d/%d, requested %d, error %v", len(got), produced, size, err)
		}
	})
}
