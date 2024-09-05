package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogQueue = make(chan *go_sdk.Log)

type AzureConfig struct {
	GroupName         string
	TenantID          string
	ClientID          string
	ClientSecretValue string
	WorkspaceID       string
}

// Debug Change to real config names
func getAzureProcessor(group types.ModuleGroup) AzureConfig {
	azurePro := AzureConfig{}
	azurePro.GroupName = group.GroupName
	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "Event Hub Shared access policies - Connection string":
			azurePro.TenantID = cnf.ConfValue
		case "Consumer Group Name":
			azurePro.ClientID = cnf.ConfValue
		case "Storage Container Name":
			azurePro.ClientSecretValue = cnf.ConfValue
		case "Storage account connection string with key":
			azurePro.WorkspaceID = cnf.ConfValue
		}
	}
	return azurePro
}

func (ap *AzureConfig) pullLogs() {
	go_sdk.Logger().Info("starting log sync for : %s", ap.GroupName)

	cred, err := azidentity.NewClientSecretCredential(ap.TenantID, ap.ClientID, ap.ClientSecretValue, nil)
	if err != nil {
		go_sdk.Logger().ErrorF("error creating credentials: %v", err)
		return
	}

	client, err := azquery.NewLogsClient(cred, nil)
	if err != nil {
		go_sdk.Logger().ErrorF("error creating logs client: %v", err)
		return
	}

	res, err := client.QueryWorkspace(
		context.TODO(),
		ap.WorkspaceID,
		azquery.Body{
			Query: to.Ptr("union * | where TimeGenerated >= ago(5m)| order by TimeGenerated desc"),
		},
		nil,
	)
	if err != nil {
		go_sdk.Logger().ErrorF("error querying workspace: %v", err)
		return
	}
	if res.Error != nil {
		go_sdk.Logger().ErrorF("error in response: %v", res.Error)
		return
	}

	var logs []map[string]any
	for _, table := range res.Tables {
		for _, row := range table.Rows {
			rowMap := make(map[string]any)
			for i, column := range table.Columns {
				if row[i] != nil {
					if str, ok := row[i].(string); ok && str == "" {
						continue
					}
					rowMap[*column.Name] = row[i]
				}
			}
			logs = append(logs, rowMap)
		}
	}

	if len(logs) > 0 {
		for _, log := range logs {
			jsonLog, err := json.Marshal(log)
			if err != nil {
				go_sdk.Logger().ErrorF("error marshalling log: %v", err)
				continue
			}
			LogQueue <- &go_sdk.Log{
				Id:         uuid.New().String(),
				TenantId:   DefaultTenant,
				DataType:   "azure",
				DataSource: ap.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(jsonLog),
			}
		}
	} else {
		go_sdk.Logger().LogF(100, "no new azure logs found for %s", ap.GroupName)
	}
}

func processLogs() {
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
