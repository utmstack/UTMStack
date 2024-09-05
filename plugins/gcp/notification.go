package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
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
		go_sdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		go_sdk.Logger().ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := go_sdk.NewEngineClient(conn)

	notifyClient, err := client.Notify(context.Background())
	if err != nil {
		go_sdk.Logger().ErrorF("failed to create notify client: %v", err)
		os.Exit(1)
	}

	for {
		msg := <-localNotificationsChannel

		err = notifyClient.Send(msg)
		if err != nil {
			go_sdk.Logger().ErrorF("failed to send notification: %v", err)
			return
		}

		ack, err := notifyClient.Recv()
		if err != nil {
			go_sdk.Logger().ErrorF("failed to receive notification ack: %v", err)
			return
		}

		go_sdk.Logger().LogF(100, "received notification ack: %v", ack)
	}
}

func notify(topic string, body Message) {
	mByte, err := json.Marshal(body)
	if err != nil {
		go_sdk.Logger().ErrorF("failed to marshal notification body: %v", err)
		return
	}

	msg := &go_sdk.Message{
		Id:        uuid.NewString(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Topic:     topic,
		Message:   string(mByte),
	}

	localNotificationsChannel <- msg
}
