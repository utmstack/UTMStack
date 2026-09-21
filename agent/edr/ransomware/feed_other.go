//go:build !windows

package ransomware

import "context"

type nopFeed struct{}

// NewFeed returns a no-op feed off Windows so the module builds/tests on the
// host. The real ETW feed is Windows-only.
func NewFeed() FileActivityFeed { return nopFeed{} }

func (nopFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
