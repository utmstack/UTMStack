package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/lzx/lzxtest"
)

// The fixture writer emits independent LZX literal codes. CFDATA boundaries
// split frames while the folder keeps one continuous compression history.
func cabLZXFixture(t testing.TB, data []byte, window int, reserve bool) []byte {
	t.Helper()
	packed := lzxtest.Literals(data, window)
	var blocks [][]byte
	if len(data) > 32768 {
		first := len(lzxtest.Literals(data[:32768], window))
		blocks = append(blocks, packed[:first])
		packed = packed[first:]
		for len(packed) > 32768 {
			blocks = append(blocks, packed[:32768])
			packed = packed[32768:]
		}
	}
	blocks = append(blocks, packed)
	b := cabFixture(t, uint16(window<<8)|3, blocks, []cabFixtureFile{{"inert.txt", 0, uint32(len(data))}}, reserve)
	folder, blockReserve := 36, 0
	if reserve {
		folder, blockReserve = 42, 1
	}
	p := int(binary.LittleEndian.Uint32(b[folder:]))
	remaining := len(data)
	for _, block := range blocks {
		binary.LittleEndian.PutUint32(b[p:], 0) // optional checksum omitted
		n := min(remaining, 32768)
		binary.LittleEndian.PutUint16(b[p+6:], uint16(n))
		remaining -= n
		p += 8 + blockReserve + len(block)
	}
	return b
}

func TestCABLZXHiddenContentAndFrames(t *testing.T) {
	for window := 15; window <= 21; window++ {
		for _, n := range []int{73, 32768, 65573} {
			b := bytes.Repeat([]byte("ordinary text "), (n+13)/14)[:n]
			copy(b[len(b)-len(payloadMarker):], payloadMarker)
			fixture := cabLZXFixture(t, b, window, window%2 == 0)
			requireHidden(t, fixture)
			expectDetected(t, payloadState(t), fixture, "Payload.Test")
		}
	}
}

func TestCABLZXLimitsAndFailuresRemainIncomplete(t *testing.T) {
	b := cabLZXFixture(t, bytes.Repeat([]byte("ordinary harmless text "), 4000), 16, false)
	s := payloadState(t)
	s.remaining = int64(len(b) + 1)
	expectIncomplete(t, s, b, ErrLimit)
	broken := append([]byte(nil), b...)
	start := int(binary.LittleEndian.Uint32(broken[36:]))
	broken[start+8+1] |= 0x70 // invalid LZX block type
	expectIncomplete(t, payloadState(t), broken, ErrMalformed)
	unsupported := append([]byte(nil), b...)
	binary.LittleEndian.PutUint16(unsupported[42:], 22<<8|3)
	expectIncomplete(t, payloadState(t), unsupported, ErrUnsupported)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s = payloadState(t)
	s.ctx = ctx
	expectIncomplete(t, s, b, context.Canceled)
}
