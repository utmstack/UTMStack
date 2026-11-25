package socai

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// AlertsStream and AlertsSubject are where new alerts wait for the SOC AI service. The service
	// declares the same stream, with the same settings, so either side may start first.
	AlertsStream  = "SOCAI_ALERTS"
	AlertsSubject = "socai.alerts"
)

// JetStreamPublisher puts messages on a stream that keeps them until the service takes them, so alerts
// raised while the service is down or busy are not lost.
type JetStreamPublisher struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// NewJetStreamPublisher connects to NATS and makes sure the alerts stream exists. NATS may still be
// starting with the backend, so it waits up to wait.
func NewJetStreamPublisher(ctx context.Context, url string, wait time.Duration) (*JetStreamPublisher, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.RetryOnFailedConnect(true), nats.Name("backend-socai"))
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()

		return nil, fmt.Errorf("creating the JetStream context: %w", err)
	}

	deadline := time.Now().Add(wait)

	for {
		_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:      AlertsStream,
			Subjects:  []string{AlertsSubject},
			Retention: jetstream.LimitsPolicy,
			Storage:   jetstream.FileStorage,
			MaxAge:    24 * time.Hour,
		})
		if err == nil {
			return &JetStreamPublisher{nc: nc, js: js}, nil
		}

		if time.Now().After(deadline) || ctx.Err() != nil {
			nc.Close()

			return nil, fmt.Errorf("creating the %s stream: %w", AlertsStream, err)
		}

		time.Sleep(time.Second)
	}
}

func (p *JetStreamPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.js.Publish(ctx, subject, data)

	return err
}

func (p *JetStreamPublisher) Close() { p.nc.Close() }
