package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	syncInterval     = 60 * time.Second
	syncInitialDelay = 2 * time.Minute
)

// Scheduler runs AssetSync on a fixed interval after an initial delay.
// Follows the same pattern as modules/alerts/usecase/scheduler.go:42.
type Scheduler struct {
	sync *AssetSync
}

func NewScheduler(sync *AssetSync) *Scheduler {
	return &Scheduler{sync: sync}
}

func (s *Scheduler) Start(ctx context.Context) {
	catcher.Info("network_scan scheduler: starting (initial delay 2m)", nil)

	select {
	case <-time.After(syncInitialDelay):
	case <-ctx.Done():
		catcher.Info("network_scan scheduler: cancelled during initial delay — stopped", nil)
		return
	}

	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	catcher.Info("network_scan scheduler: running", nil)
	for {
		select {
		case <-ticker.C:
			if err := s.sync.RunOnce(ctx); err != nil {
				_ = catcher.Error("network_scan scheduler: RunOnce error", err, nil)
			}
		case <-ctx.Done():
			catcher.Info("network_scan scheduler: context cancelled — stopped", nil)
			return
		}
	}
}
