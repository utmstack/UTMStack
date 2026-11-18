package netblock

import (
	"context"
	"net/netip"
)

// DNSEvent is one completed DNS resolution: the queried name and the A/AAAA
// addresses it resolved to, with the querying process.
type DNSEvent struct {
	Name    string
	IPs     []netip.Addr
	PID     int
	Process string
}

// DNSFeed surfaces DNS resolutions. Real on windows (ETW); nop elsewhere.
type DNSFeed interface {
	Run(ctx context.Context, sink func(DNSEvent)) error
}
