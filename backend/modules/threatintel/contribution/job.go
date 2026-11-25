package contribution

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/service"
)

const (
	// pollInterval is how often incidents are looked at, and cycleBudget the share of it a cycle may
	// use, so one that runs long does not overlap the next.
	pollInterval = 5 * time.Minute
	cycleBudget  = 0.9

	leaseName = "threatintel-contribution"
)

// Leases lets one replica at a time run the cycle.
type Leases interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

// Job is the periodic task. Whether it does anything is decided on every cycle by enabled, which is
// the instance's own switch.
type Job struct {
	incidents service.IncidentSource
	processor *service.IncidentProcessor
	enabled   func() bool
	leases    Leases
	interval  time.Duration
}

func NewJob(incidents service.IncidentSource, processor *service.IncidentProcessor, enabled func() bool, leases Leases) *Job {
	return &Job{incidents: incidents, processor: processor, enabled: enabled, leases: leases, interval: pollInterval}
}

// Start runs a cycle now and then every interval until ctx is done.
func (j *Job) Start(ctx context.Context) {
	if j == nil {
		return
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.runOnce(ctx)
		}
	}
}

func (j *Job) runOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("threat intelligence contribution: recovered from a panic", nil, map[string]any{"panic": r})
		}
	}()

	if !j.enabled() {
		return
	}

	if j.leases != nil {
		mine, err := j.leases.Acquire(ctx, leaseName, j.interval)
		if err != nil {
			_ = catcher.Error("threat intelligence contribution: cannot take the lease", err, nil)

			return
		}

		if !mine {
			return
		}
	}

	j.cycle(ctx)
}

func (j *Job) cycle(ctx context.Context) {
	start := time.Now()

	cycleCtx, cancel := context.WithTimeout(ctx, time.Duration(float64(j.interval)*cycleBudget))
	defer cancel()

	incidents, err := j.incidents.RecentIncidents(cycleCtx)
	if err != nil {
		_ = catcher.Error("threat intelligence contribution: failed to read the incidents", err, nil)

		return
	}

	if len(incidents) == 0 {
		return
	}

	entities, blocked := 0, false

	for i, incident := range incidents {
		if cycleCtx.Err() != nil {
			catcher.Info("threat intelligence contribution: the cycle ran out of time", map[string]any{
				"processed_incidents": i, "total_incidents": len(incidents),
			})

			return
		}

		n, isBlocked, err := j.processor.ProcessIncident(cycleCtx, incident)
		if err != nil {
			_ = catcher.Error("threat intelligence contribution: failed to process an incident", err, map[string]any{
				"incident_id": incident.ID, "incident_name": incident.Name,
			})

			continue
		}

		entities += n

		if isBlocked {
			blocked = true

			break
		}
	}

	catcher.Info("threat intelligence contribution: cycle completed", map[string]any{
		"duration_seconds": time.Since(start).Seconds(), "incidents": len(incidents),
		"entities": entities, "provisioning_blocked": blocked,
	})
}
