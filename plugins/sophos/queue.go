package main

import (
	"context"
	"fmt"
	"strings"
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
	queuePlugin = "sophos"
)

func logCoordinationReady() {
	catcher.Info(coordinationReadyMsg, map[string]any{
		"process":        processName,
		"plugin":         queuePlugin,
		"schedulerLease": coordination.SchedulerLeaseKey(queuePlugin),
	})
}

// WindowStart is a first-activation hint only: once a group has a cursor it is
// ignored, since only the cursor knows where the previous worker got to.
func jobsForGroups(groups []*ModuleGroup, windowEnd time.Time) []coordination.Job {
	jobs := make([]coordination.Job, 0, len(groups))
	for _, group := range groups {
		jobs = append(jobs, coordination.Job{
			TenantID:    group.TenantId,
			GroupName:   group.GroupName,
			WindowStart: windowEnd.Add(-tickInterval),
			WindowEnd:   windowEnd,
		})
	}
	return jobs
}

func publishTickJobs(ctx context.Context, scheduler coordination.Store, publisher coordination.JobPublisher, holder string, groups []*ModuleGroup, windowEnd time.Time) {
	coordination.PublishJobsIfElected(ctx, scheduler, publisher,
		coordination.SchedulerLeaseKey(queuePlugin), holder,
		jobsForGroups(groups, windowEnd),
		func(job coordination.Job, err error) {
			_ = catcher.Error("error publishing job", err, map[string]any{
				"process": processName,
				"group":   job.TenantID + "/" + job.GroupName,
			})
		})
}

func resolveGroup(tenantId, groupName string) (*ModuleGroup, bool) {
	activeGroupsMu.RLock()
	defer activeGroupsMu.RUnlock()
	group, ok := activeGroups[tenantId+"/"+groupName]
	return group, ok
}

func hasCompleteCredentials(group *ModuleGroup) bool {
	for _, c := range group.ModuleGroupConfigurations {
		if strings.TrimSpace(c.ConfValue) == "" {
			return false
		}
	}
	return true
}

// Load/work/Save/Ack ordering is owned by ConsumeAndAdvanceCursor: saving
// before the pull completes moves the position past data never ingested.
func consumeJob(
	ctx context.Context,
	cursors coordination.CursorStore,
	delivery coordination.JobDelivery,
	group *ModuleGroup,
	in ingestion,
	encryptionKey string,
) error {
	key := sophosCursorKey(group)

	// Captured by whichever callback ran; ConsumeWork cannot return them.
	var floorFrom, floorTo int64
	var ingested int

	pullFrom := func(from cursorSnapshot, windowEnd time.Time) ([]byte, error) {
		nextKey, records, err := pull(group, from.StartTime, from.NextKey, in)
		if err != nil {
			return nil, err
		}
		advanced := from.advanced(nextKey, windowEnd)
		floorFrom, floorTo, ingested = from.StartTime, advanced.StartTime, records
		return coordination.MarshalCursorPayload(advanced, encryptionKey)
	}

	work := coordination.ConsumeWork{
		Activate: func(ctx context.Context, job coordination.Job) ([]byte, error) {
			return pullFrom(seedFrom(job.WindowStart), job.WindowEnd)
		},

		// A persisted position is authoritative and inherited unchanged.
		// Re-seeding it from job.WindowStart would discard the backlog the
		// stored continuation key exists to fetch.
		Resume: func(ctx context.Context, job coordination.Job, cur coordination.Cursor) ([]byte, error) {
			persisted, err := coordination.UnmarshalCursorPayload[cursorSnapshot](cur.Data, encryptionKey)
			if err != nil {
				return nil, err
			}
			if !persisted.usable() {
				// Fail closed. Re-seeding from now would skip everything
				// since the last real position; adopting the zero floor
				// would request Sophos Central's whole retained history.
				return nil, fmt.Errorf("refusing to ingest %s from a cursor with no event-time floor", group.Key())
			}
			return pullFrom(persisted, job.WindowEnd)
		},
	}

	if err := coordination.ConsumeAndAdvanceCursor(ctx, cursors, delivery, key, work); err != nil {
		return err
	}

	// Only here, after the cursor is durable and the job acked. Logging from
	// inside pullFrom would report success for a cycle that then lost its CAS.
	catcher.Info(jobConsumedMsg, map[string]any{
		"process":     processName,
		"group":       group.Key(),
		"windowStart": floorFrom,
		"windowEnd":   floorTo,
		"records":     ingested,
	})
	return nil
}

func handleJob(ctx context.Context, cursors coordination.CursorStore, delivery coordination.JobDelivery, encryptionKeyFn func() string) error {
	group, ok := resolveGroup(delivery.Job.TenantID, delivery.Job.GroupName)
	if !ok {
		// Group not configured here: no Ack and no Nak, so AckWait redelivers
		// it to a worker that has it, and MaxDeliver drops it if none does.
		return nil
	}

	if !hasCompleteCredentials(group) {
		// Same treatment: a group that cannot authenticate must not have its
		// position advanced, nor be acked as though its window was ingested.
		return nil
	}

	return consumeJob(ctx, cursors, delivery, group, liveIngestion(group), encryptionKeyFn())
}

func runQueueConsumer(ctx context.Context, consumer coordination.JobConsumer, cursors coordination.CursorStore, encryptionKeyFn func() string) {
	coordination.RunJobConsumer(ctx, consumer,
		func(ctx context.Context, delivery coordination.JobDelivery) error {
			return handleJob(ctx, cursors, delivery, encryptionKeyFn)
		},
		func(err error) {
			_ = catcher.Error("error consuming job", err, map[string]any{"process": processName})
		})
}
