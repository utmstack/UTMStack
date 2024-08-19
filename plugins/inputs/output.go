package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func sendLog() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(
		helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		helpers.Logger().ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		helpers.Logger().ErrorF("failed to create input client: %v", err)
		os.Exit(1)
	}

	for {
		l := <-localLogsChannel

		err := inputClient.Send(l)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
			notify("enqueue_failure", Message{Cause: err.Error(), DataType: l.DataType, DataSource: l.DataSource})
			continue
		}

		// TODO: implement a logic to resend failed logs
		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			notify("ack_failure", Message{Cause: err.Error(), DataType: l.DataType, DataSource: l.DataSource})
			continue
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
		notify("enqueue_success", Message{DataType: l.DataType, DataSource: l.DataSource})
	}
}
