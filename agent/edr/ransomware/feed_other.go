//go:build !windows && !linux

package ransomware

import "context"

type nopFeed struct{}

// NewFeed returns a no-op feed off Windows and Linux so the module
// builds/tests on other hosts (e.g. darwin cross-builds). The real feeds are
// the Windows ETW feed and the Linux fanotify feed.
func NewFeed(vols ...string) FileActivityFeed { return nopFeed{} }

func (nopFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
