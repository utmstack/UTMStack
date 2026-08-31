package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// A job is only a hint; CursorStore holds the authoritative position.
	jobStreamMaxAge = 1 * time.Hour

	// Long enough that a real pull against a third-party API is not read as a
	// dead worker.
	jobConsumerAckWait = 4*time.Minute + 30*time.Second

	jobConsumerMaxDeliver = 3

	jobConsumerDurableSuffix = "-workers"

	jobSubjectPrefix = "jobs"
)

func jobStreamName(plugin string) string {
	return "INGEST_JOBS_" + strings.ToUpper(plugin)
}

func jobSubject(plugin string, job Job) string {
	return fmt.Sprintf("%s.%s.%s.%s", jobSubjectPrefix, plugin, job.TenantID, job.GroupName)
}

func jobSubjectFilter(plugin string) string {
	return fmt.Sprintf("%s.%s.>", jobSubjectPrefix, plugin)
}

// Called by both constructors, so a worker may start before any scheduler.
func ensureJobStream(ctx context.Context, js jetstream.JetStream, plugin string) error {
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      jobStreamName(plugin),
		Subjects:  []string{jobSubjectFilter(plugin)},
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    jobStreamMaxAge,
	})
	return err
}

type NATSJobPublisher struct {
	js     jetstream.JetStream
	plugin string
}

func NewNATSJobPublisher(ctx context.Context, js jetstream.JetStream, plugin string) (*NATSJobPublisher, error) {
	if err := ensureJobStream(ctx, js, plugin); err != nil {
		return nil, err
	}
	return &NATSJobPublisher{js: js, plugin: plugin}, nil
}

func (p *NATSJobPublisher) Publish(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = p.js.Publish(ctx, jobSubject(p.plugin, job), data)
	return err
}

type NATSJobConsumer struct {
	consumer jetstream.Consumer
}

func NewNATSJobConsumer(ctx context.Context, js jetstream.JetStream, plugin string) (*NATSJobConsumer, error) {
	return NewNATSJobConsumerWithConfig(ctx, js, plugin, jobConsumerAckWait, jobConsumerMaxDeliver)
}

func NewNATSJobConsumerWithConfig(ctx context.Context, js jetstream.JetStream, plugin string, ackWait time.Duration, maxDeliver int) (*NATSJobConsumer, error) {
	if err := ensureJobStream(ctx, js, plugin); err != nil {
		return nil, err
	}

	consumer, err := js.CreateOrUpdateConsumer(ctx, jobStreamName(plugin), jetstream.ConsumerConfig{
		Durable:       plugin + jobConsumerDurableSuffix,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       ackWait,
		MaxDeliver:    maxDeliver,
		FilterSubject: jobSubjectFilter(plugin),
	})
	if err != nil {
		return nil, err
	}

	return &NATSJobConsumer{consumer: consumer}, nil
}

// JetStream has no idle signal: Next returns nats.ErrTimeout for an empty
// fetch, the same error a broken connection produces.
func isEmptyFetchWindow(ctx context.Context, err error) bool {
	// Which spelling arrives depends on the server version and on whether the
	// request expired server-side or returned an empty batch.
	if errors.Is(err, nats.ErrTimeout) || errors.Is(err, jetstream.ErrNoMessages) {
		return true
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// A ctx already done on entry never reaches the server: FetchContext
	// rejects it with ErrInvalidOption, carrying neither context sentinel.
	return ctx.Err() != nil && errors.Is(err, jetstream.ErrInvalidOption)
}

func (c *NATSJobConsumer) Fetch(ctx context.Context) (JobDelivery, error) {
	msg, err := c.consumer.Next(jetstream.FetchContext(ctx))
	if err != nil {
		if isEmptyFetchWindow(ctx, err) {
			// Wrapped, so the transport's own cause stays unwrappable.
			return JobDelivery{}, fmt.Errorf("%w: %w", ErrNoJobs, err)
		}
		return JobDelivery{}, err
	}

	var job Job
	if err := json.Unmarshal(msg.Data(), &job); err != nil {
		return JobDelivery{}, err
	}

	meta, err := msg.Metadata()
	if err != nil {
		return JobDelivery{}, err
	}

	return JobDelivery{
		Job:          job,
		NumDelivered: meta.NumDelivered,
		Ack:          msg.Ack,
		Nak:          msg.Nak,
	}, nil
}
