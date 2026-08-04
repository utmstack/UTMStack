package publish

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

// Declared the same way on both sides so neither has to start first.
const (
	streamName    = "INGEST_LOGS"
	shardSubjects = "logs.shard.*"
	dlqSubject    = "logs.dlq"

	streamMaxAge = 24 * time.Hour
)

type Publisher struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	shards int
}

func New(ctx context.Context, url string, shards int) (*Publisher, error) {
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.Name("log-input"),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("creating the JetStream context: %w", err)
	}

	if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      streamName,
		Subjects:  []string{shardSubjects, dlqSubject},
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    streamMaxAge,
	}); err != nil {
		nc.Close()
		return nil, fmt.Errorf("creating the stream: %w", err)
	}

	return &Publisher{nc: nc, js: js, shards: shards}, nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}

// Publish waits for the server to acknowledge the message.
func (p *Publisher) Publish(ctx context.Context, l *plugins.Log) error {
	payload, err := utils.ProtoMessageToString(l)
	if err != nil {
		return fmt.Errorf("serialising the log: %w", err)
	}

	if _, err := p.js.Publish(ctx, p.subject(l), []byte(*payload)); err != nil {
		return fmt.Errorf("publishing to %s: %w", p.subject(l), err)
	}
	return nil
}

// subject spreads logs by source rather than at random, so one source's logs
// land on one subject and a shard can be given its own consumer later without
// resharding what is already published.
func (p *Publisher) subject(l *plugins.Log) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(l.TenantId))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(l.DataSource))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(l.DataType))

	return fmt.Sprintf("logs.shard.%d", int(h.Sum32())%p.shards)
}
