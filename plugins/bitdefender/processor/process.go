package processor

import (
	"context"
	"fmt"
	"os"
	"path"

	gosdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogsChan = make(chan *gosdk.Log)

func ProcessLogs() {
	conn, err := grpc.NewClient(
		fmt.Sprintf("unix://%s", path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = gosdk.Error("failed to connect to engine server", err, map[string]any{})
		os.Exit(1)
	}

	client := gosdk.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		_ = gosdk.Error("failed to create input client", err, map[string]any{})
		os.Exit(1)
	}

	for {
		log := <-LogsChan
		err := inputClient.Send(log)
		if err != nil {
			_ = gosdk.Error("failed to send log", err, map[string]any{})
		}

		_, err = inputClient.Recv()
		if err != nil {
			_ = gosdk.Error("failed to receive ack", err, map[string]any{})
			os.Exit(1)
		}
	}
}
