package processor

import (
	"context"
	"fmt"
	"os"
	"path"

	"cloud.google.com/go/pubsub"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/UTMStack/plugins/gcp/utils"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogsQueue = make(chan *plugins.Log)

func PullLogs(stopChan chan struct{}, tenant string, cnf schema.ModuleConfig) {
	for {
		select {
		case <-stopChan:
			utils.Logger.LogF(100, "stopping logs puller for %s", tenant)
			return
		default:
			utils.Logger.LogF(100, "pulling logs for %s", tenant)
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, cnf.ProjectID, option.WithCredentialsJSON([]byte(cnf.JsonKey)))
			if err != nil {
				utils.Logger.ErrorF("failed to create client: %v", err)
				return
			}

			sub := client.Subscription(cnf.SubscriptionID)

			err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				messageData := string(msg.Data)
				if err != nil {
					utils.Logger.ErrorF("failed to unmarshal message: %v", err)
					msg.Nack()
					return
				}

				LogsQueue <- &plugins.Log{
					DataType:   "google",
					DataSource: tenant,
					Raw:        messageData,
				}

				msg.Ack()
			})
			if err != nil {
				utils.Logger.ErrorF("failed to receive messages: %v", err)
				continue
			}
		}
	}
}

func ProcessLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		utils.Logger.ErrorF("failed to create input client: %v", err)
		os.Exit(1)
	}

	go receiveAcks(inputClient)

	for log := range LogsQueue {
		utils.Logger.LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			utils.Logger.ErrorF("failed to send log: %v", err)
		}
	}
}

func receiveAcks(inputClient plugins.Engine_InputClient) {
	for {
		ack, err := inputClient.Recv()
		if err != nil {
			utils.Logger.ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		utils.Logger.LogF(100, "received ack: %v", ack)
	}
}
