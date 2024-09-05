package processor

import (
	"context"
	"fmt"
	"os"
	"path"

	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogsChan = make(chan *go_sdk.Log)

func ProcessLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(go_sdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		go_sdk.Logger().ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := go_sdk.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		go_sdk.Logger().ErrorF("failed to create input client: %v", err)
		os.Exit(1)
	}

	for {
		log := <-LogsChan
		go_sdk.Logger().LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			go_sdk.Logger().ErrorF("failed to send log: %v", err)
		} else {
			go_sdk.Logger().LogF(100, "successfully sent log to processing engine: %v", log)
		}

		ack, err := inputClient.Recv()
		if err != nil {
			go_sdk.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		go_sdk.Logger().LogF(100, "received ack: %v", ack)
	}
}
