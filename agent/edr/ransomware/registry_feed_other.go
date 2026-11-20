//go:build !windows

package ransomware

import (
	"context"

	"github.com/utmstack/UTMStack/shared/logger"
)

// noopRegistryFeed is the registry feed on non-Windows platforms. Registry
// tampering is Windows-only, so the feed blocks until ctx is cancelled and
// delivers no events. It reports as running (a healthy no-op) so the service
// can fold it into the ransomware sensor-health aggregate on every platform.
type noopRegistryFeed struct{}

// NewRegistryFeed returns the no-op registry feed for non-Windows platforms.
func NewRegistryFeed() RegistryFeed { return &noopRegistryFeed{} }

// Health reports no error: a no-op registry feed is healthy by design.
func (f *noopRegistryFeed) Health() error { return nil }

// Run blocks until ctx is cancelled.
func (f *noopRegistryFeed) Run(ctx context.Context, _ func(RegistryEvent)) error {
	logger.Info("UTMStack EDR: registry feed inactive (non-Windows)")
	<-ctx.Done()
	return ctx.Err()
}
