package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

const (
	jobConsumedMsg       = "job consumed, cursor advanced"
	coordinationReadyMsg = "coordination ready, consuming jobs"

	// queuePlugin names the NATS stream, subjects, durable consumer and
	// scheduler election key. Changing it orphans all of them on an existing
	// deployment.
	queuePlugin = "o365"
)

func logCoordinationReady() {
	catcher.Info(coordinationReadyMsg, map[string]any{
		"process":        processName,
		"plugin":         queuePlugin,
		"schedulerLease": coordination.SchedulerLeaseKey(queuePlugin),
	})
}

// One job per group, never per tenant: a per-tenant job would collapse all of a
// tenant's groups onto a single worker. Job.TenantID carries UtmTenantId, not
// the Microsoft tenant id.
func jobsForGroups(groups []*ModuleGroup, windowStart, windowEnd time.Time) []coordination.Job {
	jobs := make([]coordination.Job, 0, len(groups))
	for _, grp := range groups {
		jobs = append(jobs, coordination.Job{
			TenantID:    grp.UtmTenantId,
			GroupName:   grp.GroupName,
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
		})
	}
	return jobs
}

func publishTickJobs(ctx context.Context, scheduler coordination.Store, publisher coordination.JobPublisher, holder string, groups []*ModuleGroup, windowStart, windowEnd time.Time) {
	coordination.PublishJobsIfElected(ctx, scheduler, publisher,
		coordination.SchedulerLeaseKey(queuePlugin), holder,
		jobsForGroups(groups, windowStart, windowEnd),
		func(job coordination.Job, err error) {
			_ = catcher.Error("error publishing job", err, map[string]any{
				"process": processName,
				"group":   job.TenantID + "/" + job.GroupName,
			})
		})
}

// The key must match what group.Key() produces. Every worker loads the same
// pipeline config, so a configured group resolves here too, modulo propagation lag.
func resolveGroup(tenantId, groupName string) (*ModuleGroup, bool) {
	activeGroupsMu.RLock()
	defer activeGroupsMu.RUnlock()
	grp, ok := activeGroups[tenantId+"/"+groupName]
	return grp, ok
}

// Turns a delivered job into a pull() call and a persisted cursor. The
// load-work-save-ack ordering belongs to coordination.ConsumeAndAdvanceCursor
// and must not be hand-rolled here.
func consumeJob(
	ctx context.Context,
	cursors coordination.CursorStore,
	delivery coordination.JobDelivery,
	group *ModuleGroup,
	pullFn func(startTime, endTime time.Time, group *ModuleGroup) (int, error),
	encryptionKey string,
) error {
	key := o365CursorKey(group)

	// Captured by whichever callback runs, so the outcome can be logged only
	// after the cursor is durable. ConsumeWork's signatures cannot return it.
	var windowStart, windowEnd time.Time
	var ingested int

	work := coordination.ConsumeWork{
		// First activation, no persisted cursor. job.WindowStart is seeded from
		// "now minus one tick", never the epoch, so no full-history replay.
		Activate: func(ctx context.Context, job coordination.Job) ([]byte, error) {
			now := time.Now().UTC()
			n, err := pullFn(job.WindowStart, now, group)
			if err != nil {
				return nil, err
			}
			windowStart, windowEnd, ingested = job.WindowStart, now, n
			return coordination.MarshalCursorPayload(cursorPayload{WindowEnd: now}, encryptionKey)
		},
		// The starting point is cur.Data, never job.WindowStart: a different
		// worker may have handled the previous tick.
		Resume: func(ctx context.Context, job coordination.Job, cur coordination.Cursor) ([]byte, error) {
			prev, err := coordination.UnmarshalCursorPayload[cursorPayload](cur.Data, encryptionKey)
			if err != nil {
				return nil, err
			}
			now := time.Now().UTC()
			n, err := pullFn(prev.WindowEnd, now, group)
			if err != nil {
				return nil, err
			}
			windowStart, windowEnd, ingested = prev.WindowEnd, now, n
			return coordination.MarshalCursorPayload(cursorPayload{WindowEnd: now}, encryptionKey)
		},
	}

	if err := coordination.ConsumeAndAdvanceCursor(ctx, cursors, delivery, key, work); err != nil {
		return err
	}

	catcher.Info(jobConsumedMsg, map[string]any{
		"process":     processName,
		"group":       group.Key(),
		"windowStart": windowStart.Format(time.RFC3339Nano),
		"windowEnd":   windowEnd.Format(time.RFC3339Nano),
		"records":     ingested,
	})
	return nil
}

// Durable consumer goroutine; start once per process.
func runQueueConsumer(ctx context.Context, consumer coordination.JobConsumer, cursors coordination.CursorStore, encryptionKeyFn func() string) {
	coordination.RunJobConsumer(ctx, consumer,
		func(ctx context.Context, delivery coordination.JobDelivery) error {
			group, ok := resolveGroup(delivery.Job.TenantID, delivery.Job.GroupName)
			if !ok {
				// Group not configured on this worker. Deliberately neither
				// acked nor nacked: AckWait drives redelivery, and MaxDeliver
				// exhaustion drops the job if the group is gone everywhere.
				return nil
			}
			return consumeJob(ctx, cursors, delivery, group, pull, encryptionKeyFn())
		},
		func(err error) {
			_ = catcher.Error("error consuming job", err, map[string]any{"process": processName})
		})
}
