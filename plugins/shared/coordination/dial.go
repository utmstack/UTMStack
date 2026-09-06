package coordination

import (
	"context"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ConnectJetStream dials NATS_URL. The caller owns the returned *nats.Conn and
// must Close it. connName identifies this plugin to the server.
func ConnectJetStream(ctx context.Context, connName string) (jetstream.JetStream, *nats.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	url := os.Getenv("NATS_URL")
	if url == "" {
		return nil, nil, fmt.Errorf("NATS_URL is required")
	}

	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.Name(connName))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("creating the JetStream context: %w", err)
	}

	return js, nc, nil
}
