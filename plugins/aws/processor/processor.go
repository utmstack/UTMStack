package processor

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/aws/schema"
	"github.com/utmstack/UTMStack/plugins/aws/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func PullLogs(queue chan *plugins.Log, stopChan chan struct{}, tenant string, cnf schema.AWSConfig) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cnf.Region),
		Credentials: credentials.NewStaticCredentials(cnf.AccessKeyID, cnf.SecretAccessKey, ""),
	})
	if err != nil {
		utils.Logger.ErrorF("Failed to create AWS session: %v", err)
		return
	}

	client := cloudwatchlogs.New(sess)

	for {
		select {
		case <-stopChan:
			return
		default:
			params := &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(cnf.LogGroupName),
				LogStreamName: aws.String(cnf.LogStreamName),
			}

			resp, err := client.GetLogEvents(params)
			if err != nil {
				utils.Logger.ErrorF("Failed to get log events: %v", err)
				continue
			}

			for _, event := range resp.Events {
				queue <- &plugins.Log{
					DataType:   "aws",
					DataSource: tenant,
					Raw:        *event.Message,
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func ProcessLogs(queue chan *plugins.Log) {

	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.ErrorF("Failed to connect to engine server: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		utils.Logger.ErrorF("Failed to create input client: %v", err)
		os.Exit(1)
	}

	go receiveAcks(inputClient)

	for log := range queue {
		err := inputClient.Send(log)
		if err != nil {
			utils.Logger.ErrorF("Failed to send log: %v", err)
		} else {
			utils.Logger.Info("Successfully sent log to processing engine: %v", log)
		}
	}
}

func receiveAcks(inputClient plugins.Engine_InputClient) {
	for {
		ack, err := inputClient.Recv()
		if err != nil {
			utils.Logger.ErrorF("Failed to receive ack: %v", err)
			os.Exit(1)
		}

		utils.Logger.Info("Received ack: %v", ack)
	}
}
