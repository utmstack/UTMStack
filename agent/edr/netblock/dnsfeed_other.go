//go:build !windows

package netblock

import "context"

type nopDNSFeed struct{}

func NewDNSFeed() DNSFeed { return nopDNSFeed{} }

func (nopDNSFeed) Run(ctx context.Context, sink func(DNSEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
