package lzx

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/lzx/lzxtest"
)

type cancelAfterFrame struct {
	context.Context
	calls int
}

func (c *cancelAfterFrame) Err() error {
	c.calls++
	if c.calls > 1 {
		return context.Canceled
	}
	return nil
}

func TestDecodeContextStopsBetweenFrames(t *testing.T) {
	data := make([]byte, 65536)
	ctx := &cancelAfterFrame{Context: context.Background()}
	got, err := DecodeContext(ctx, lzxtest.Literals(data, 16), 16, len(data), 0)
	if !errors.Is(err, context.Canceled) || len(got) != 32768 {
		t.Fatalf("cancellation: %d decoded, %v", len(got), err)
	}
}

func TestDecodeValidatedContextRetainsTranslatedFramesAndCountsWork(t *testing.T) {
	data := bytes.Repeat([]byte("ordinary LZX output "), 4000)[:65536]
	packed := lzxtest.Literals(data, 18)
	got, produced, err := DecodeValidatedContext(context.Background(), packed[:len(packed)-100], 18, len(data), 0)
	if err == nil || len(got) != 32768 || !bytes.Equal(got, data[:32768]) || produced <= len(got) || produced > len(data) {
		t.Fatalf("validated frames: %d/%d bytes, %v", len(got), produced, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, produced, err := DecodeValidatedContext(ctx, packed, 18, len(data), 0); len(got) != 0 || produced != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled validated decoder: %d/%d bytes, %v", len(got), produced, err)
	}
}
