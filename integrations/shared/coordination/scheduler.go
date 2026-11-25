package coordination

import "context"

// Only covers one tick's publish; it expires before the next, unrenewed.
const SchedulerLeaseTTL = schedulerBucketTTL

// The plugin namespace is load-bearing: every queue-path plugin shares one
// bucket, so an unnamespaced key would elect one publisher across all of them.
func SchedulerLeaseKey(plugin string) string {
	return "scheduler." + plugin
}

// Any acquire failure is a silent no-op; a second publisher would only
// duplicate jobs. A Publish failure must not abort the rest of the tick.
func PublishJobsIfElected(
	ctx context.Context,
	scheduler Store,
	publisher JobPublisher,
	leaseKey string,
	holder string,
	jobs []Job,
	onError func(job Job, err error),
) {
	if _, err := scheduler.Acquire(ctx, leaseKey, holder, SchedulerLeaseTTL); err != nil {
		return
	}

	for _, job := range jobs {
		if err := publisher.Publish(ctx, job); err != nil && onError != nil {
			onError(job, err)
		}
	}
}
