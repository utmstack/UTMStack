package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogsQueue = make(chan *plugins.Log)

func pullLogs(stopChan chan struct{}, tenant string, cnf ModuleConfig) {
	for {
		select {
		case <-stopChan:
			helpers.Logger().LogF(100, "stopping logs puller for %s", tenant)
			return
		default:
			helpers.Logger().LogF(100, "pulling logs for %s", tenant)
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, cnf.ProjectID, option.WithCredentialsJSON([]byte(cnf.JsonKey)))
			if err != nil {
				helpers.Logger().ErrorF("failed to create client: %v", err)
				return
			}

			sub := client.Subscription(cnf.SubscriptionID)
			helpers.Logger().LogF(100, "subscribing to %s", cnf.SubscriptionID)

			err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				helpers.Logger().LogF(100, "received message: %v", string(msg.Data))
				LogsQueue <- &plugins.Log{
					Id:         uuid.New().String(),
					TenantId:   DefaultTenant,
					DataType:   "google",
					DataSource: tenant,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        string(msg.Data),
				}

				msg.Ack()
			})
			if err != nil {
				helpers.Logger().ErrorF("failed to receive messages: %v", err)
				continue
			}
		}
	}
}

func processLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		log := <-LogsQueue
		helpers.Logger().LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
		} else {
			helpers.Logger().LogF(100, "successfully sent log to processing engine: %v", log)
		}

		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
	}
}
