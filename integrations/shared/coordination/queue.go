package coordination

import (
	"context"
	"errors"
	"time"
)

// WindowStart and WindowEnd are a hint only: the consumer's authoritative
// position is CursorStore.Load, so they matter only on a first activation.
type Job struct {
	TenantID    string    `json:"tenantId"`
	GroupName   string    `json:"groupName"`
	WindowStart time.Time `json:"windowStart"`
	WindowEnd   time.Time `json:"windowEnd"`
}

type JobPublisher interface {
	Publish(ctx context.Context, job Job) error
}

// Exactly one of Ack or Nak must eventually be called.
type JobDelivery struct {
	Job Job

	// 1 on first delivery, incremented on every redelivery.
	NumDelivered uint64

	Ack func() error

	// Requests immediate redelivery rather than waiting out AckWait.
	Nak func() error
}

// A job exceeding MaxDeliver is dropped by JetStream with no signal and no
// dead-letter subject. No DLQ is needed: a failed job never advanced the
// cursor, so the next tick republishes the group and retries the same window.
type JobConsumer interface {
	Fetch(ctx context.Context) (JobDelivery, error)
}

// A normal idle result, never a failure. JetStream reports an expired pull with
// nats.ErrTimeout, the same error a network partition produces.
var ErrNoJobs = errors.New("queue: no job available within the fetch window")
