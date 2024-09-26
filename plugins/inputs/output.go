package main

import (
	"context"
	"fmt"
	"os"
	"path"

	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func sendLog() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(
		go_sdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		l := <-localLogsChannel

		err := inputClient.Send(l)
		if err != nil {
			e := go_sdk.Logger().ErrorF("failed to send log: %v", err)
			notify(TOPIC_ENQUEUE_FAILURE, Message{Cause: go_sdk.PointerOf(e.Message), DataType: l.DataType, DataSource: l.DataSource})
			continue
		}

		// TODO: implement a logic to resend failed logs
		ack, err := inputClient.Recv()
		if err != nil {
			e := go_sdk.Logger().ErrorF("failed to receive ack: %v", err)
			notify(TOPIC_ENQUEUE_FAILURE, Message{Cause: go_sdk.PointerOf(e.Message), DataType: l.DataType, DataSource: l.DataSource})
			continue
		}

		go_sdk.Logger().LogF(100, "received ack: %v", ack)
		notify(TOPIC_ENQUEUE_SUCCESS, Message{DataType: l.DataType, DataSource: l.DataSource})
	}
}
