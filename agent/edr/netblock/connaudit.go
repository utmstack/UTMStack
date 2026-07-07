// edr/netblock/connaudit.go
package netblock

import (
	"context"
	"net/netip"
)

// BlockEvent describes one enforced WFP drop, for detection eventing.
type BlockEvent struct {
	RemoteIP    netip.Addr
	RemotePort  int
	PID         int
	ProcessPath string
	Direction   string
}

// ConnAudit surfaces WFP drop events. Real on windows; no-op elsewhere.
type ConnAudit interface {
	Run(ctx context.Context, out chan<- BlockEvent)
}
