//go:build !windows

package netblock

import (
	"context"
	"testing"
	"time"
)

func TestNoopDNSFeedStops(t *testing.T) {
	f := NewDNSFeed()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = f.Run(ctx, func(DNSEvent) {}); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("nop dns feed did not stop on ctx cancel")
	}
}
