package processor

import (
	"context"
	"fmt"
	"os"
	"path"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/UTMStack/plugins/gcp/utils"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func PullLogs(queue chan *plugins.Log, stopChan chan struct{}, tenant string, cnf schema.ModuleConfig) {
	uuid := uuid.New()
	filePath := "./" + cnf.ProjectID + "-" + uuid.String() + ".json"
	file, err := os.Create(filePath)
	if err != nil {
		utils.Logger.ErrorF("Error creating config file: %v", err)
		return
	}
	_, err = file.WriteString(cnf.JsonKey)
	if err != nil {
		utils.Logger.ErrorF("Error writing to config file: %v", err)
		return
	}
	defer file.Close()

	for {
		select {
		case <-stopChan:
			err = os.RemoveAll(filePath)
			if err != nil {
				utils.Logger.ErrorF("Error removing config file: %v", err)
			}
			return
		default:
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, cnf.ProjectID, option.WithCredentialsFile(filePath))
			if err != nil {
				utils.Logger.ErrorF("Failed to create client: %v", err)
				return
			}

			sub := client.Subscription(cnf.SubscriptionID)

			messageData := ""
			err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				messageData = string(msg.Data)
				if err != nil {
					utils.Logger.ErrorF("Failed to unmarshal message: %v", err)
					msg.Nack()
					return
				}
				msg.Ack()
			})
			if err != nil {
				utils.Logger.ErrorF("Failed to receive messages: %v", err)
				continue
			}
			if messageData != "" {
				queue <- &plugins.Log{
					DataType:   "google",
					DataSource: tenant,
					Raw:        messageData,
				}
			}
		}
	}
}

func ProcessLogs(queue chan *plugins.Log) {
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

	go receiveAcks(inputClient)

	for log := range queue {
		err := inputClient.Send(log)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
		}
	}
}

func receiveAcks(inputClient plugins.Engine_InputClient) {
	for {
		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
	}
}
