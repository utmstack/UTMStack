package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
)

const (
	schedulerInterval     = 30 * time.Second
	schedulerInitialDelay = 10 * time.Second
)

// Scheduler periodically assigns asset groups to alerts in AUTOMATIC_REVIEW.
// Tag-rule evaluation and release-to-Open were moved to the alerts plugin and
// now happen synchronously at ingest time. The asset-group sweep stays here
// because its data source (network_scan) is not yet available to the plugin.
type Scheduler struct {
	alertRepo connectors.AlertRepository
}

func NewScheduler(alertRepo connectors.AlertRepository) *Scheduler {
	return &Scheduler{alertRepo: alertRepo}
}

func (s *Scheduler) Start(ctx context.Context) {
	catcher.Info("alerts scheduler: starting (initial delay 10s)", nil)

	select {
	case <-time.After(schedulerInitialDelay):
	case <-ctx.Done():
		catcher.Info("alerts scheduler: cancelled during initial delay — stopped", nil)
		return
	}

	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	catcher.Info("alerts scheduler: running", nil)

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			catcher.Info("alerts scheduler: context cancelled — stopped", nil)
			return
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	// TODO(module-33): wire real mapping when network_scan is ported.
	if err := s.alertRepo.AssignAssetGroups(ctx, nil); err != nil {
		_ = catcher.Error("alerts scheduler: AssignAssetGroups error", err, nil)
	}
}
