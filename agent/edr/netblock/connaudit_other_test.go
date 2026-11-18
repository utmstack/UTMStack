// edr/netblock/connaudit_other_test.go
//go:build !windows

package netblock

import (
	"context"
	"testing"
	"time"
)

func TestNoopConnAuditStops(t *testing.T) {
	ca := NewConnAudit()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { ca.Run(ctx, make(chan BlockEvent)); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("conn audit did not stop on ctx cancel")
	}
}
