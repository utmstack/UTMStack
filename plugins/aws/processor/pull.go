package processor

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/aws/config"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogQueue = make(chan *go_sdk.Log)

func PullLogs(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
	go_sdk.Logger().Info("starting log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime)
	if err != nil {
		go_sdk.Logger().ErrorF("failed to get logs: %v", err)
		return
	}

	for _, log := range logs {
		LogQueue <- &go_sdk.Log{
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
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(go_sdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		log := <-LogQueue
		go_sdk.Logger().LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			go_sdk.Logger().ErrorF("failed to send log: %v", err)
		} else {
			go_sdk.Logger().LogF(100, "successfully sent log to processing engine: %v", log)
		}

		ack, err := inputClient.Recv()
		if err != nil {
			go_sdk.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		go_sdk.Logger().LogF(100, "received ack: %v", ack)
	}
}
