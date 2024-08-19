package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Message struct {
	Cause      string `json:"cause"`
	DataType   string `json:"data_type"`
	DataSource string `json:"data_source"`
}

func sendNotification() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(
		helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		helpers.Logger().ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	notifyClient, err := client.Notify(context.Background())
	if err != nil {
		helpers.Logger().ErrorF("failed to create notify client: %v", err)
		os.Exit(1)
	}

	for {
		msg := <-localNotificationsChannel

		err = notifyClient.Send(msg)
		if err != nil {
			helpers.Logger().ErrorF("failed to send notification: %v", err)
			return
		}

		// TODO: implement a logic to resend failed notifications
		ack, err := notifyClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive notification ack: %v", err)
			return
		}

		helpers.Logger().LogF(100, "received notification ack: %v", ack)
	}
}

func notify(topic string, body Message) {
	mByte, err := json.Marshal(body)
	if err != nil {
		helpers.Logger().ErrorF("failed to marshal notification body: %v", err)
		return
	}

	msg := &plugins.Message{
		Id:        uuid.NewString(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Topic:     topic,
		Message:   string(mByte),
	}

	localNotificationsChannel <- msg
}
