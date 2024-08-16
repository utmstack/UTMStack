package processor

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/aws/config"
	"github.com/utmstack/UTMStack/plugins/aws/utils"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogQueue = make(chan *plugins.Log)

func PullLogs(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
	utils.Logger.Info("starting log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime)
	if err != nil {
		utils.Logger.ErrorF("failed to get logs: %v", err)
		return
	}

	for _, log := range logs {
		LogQueue <- &plugins.Log{
			Id:         uuid.New().String(),
			TenantId:   config.DefaultTenant,
			DataType:   "aws",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		}
	}
}

func ProcessLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		utils.Logger.ErrorF("failed to create input client: %v", err)
		os.Exit(1)
	}

	go receiveAcks(inputClient)

	for log := range LogQueue {
		utils.Logger.LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			utils.Logger.ErrorF("failed to send log: %v", err)
		} else {
			utils.Logger.LogF(100, "successfully sent log to processing engine: %v", log)
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
