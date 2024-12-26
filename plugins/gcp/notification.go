package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	gosdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Message struct {
	Cause      string `json:"cause"`
	DataType   string `json:"data_type"`
	DataSource string `json:"data_source"`
}

func processNotification() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(
		gosdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = gosdk.Error("failed to connect to engine server", err, map[string]any{})
		os.Exit(1)
	}

	client := gosdk.NewEngineClient(conn)

	notifyClient, err := client.Notify(context.Background())
	if err != nil {
		_ = gosdk.Error("failed to create notify client", err, map[string]any{})
		os.Exit(1)
	}

	for {
		msg := <-localNotificationsChannel

		err = notifyClient.Send(msg)
		if err != nil {
			_ = gosdk.Error("failed to send notification", err, map[string]any{})
			return
		}

		_, err := notifyClient.Recv()
		if err != nil {
			_ = gosdk.Error("failed to receive notification ack", err, map[string]any{})
			return
		}
	}
}

func notify(topic string, body Message) {
	mByte, err := json.Marshal(body)
	if err != nil {
		_ = gosdk.Error("failed to encode message", err, map[string]any{})
		return
	}

	msg := &gosdk.Message{
		Id:        uuid.NewString(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Topic:     topic,
		Message:   string(mByte),
	}

	localNotificationsChannel <- msg
}
