package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/google/uuid"
	gosdk "github.com/threatwinds/go-sdk"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const delayCheck = 300
const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var logQueue = make(chan *gosdk.Log, 100*runtime.NumCPU())

func main() {
	mode := gosdk.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for t := 0; t < runtime.NumCPU(); t++ {
		go processLogs()
	}

	for {
		utmConfig := gosdk.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		moduleConfig, err := client.GetUTMConfig(enum.AZURE)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = gosdk.Error("cannot obtain module configuration", err, nil)
			}

			time.Sleep(time.Second * delayCheck)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							skip = true
							break
						}
					}

					if !skip {
						azureProcessor := getAzureProcessor(group)
						azureProcessor.pull()
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)
	}
}

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

func (ap *AzureConfig) pull() {
	cred, err := azidentity.NewClientSecretCredential(ap.TenantID, ap.ClientID, ap.ClientSecretValue, nil)
	if err != nil {
		_ = gosdk.Error("cannot obtain Azure credentials", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}

	client, err := azquery.NewLogsClient(cred, nil)
	if err != nil {
		_ = gosdk.Error("cannot create Logs client", err, map[string]any{
			"group": ap.GroupName,
		})
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
		_ = gosdk.Error("cannot query Logs", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}
	if res.Error != nil {
		_ = gosdk.Error("cannot query Logs", err, map[string]any{
			"group": ap.GroupName,
		})
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
				_ = gosdk.Error("cannot encode log to JSON", err, map[string]any{
					"group": ap.GroupName,
				})
				continue
			}
			logQueue <- &gosdk.Log{
				Id:         uuid.New().String(),
				TenantId:   defaultTenant,
				DataType:   "azure",
				DataSource: ap.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(jsonLog),
			}
		}
	}
}

func processLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = gosdk.Error("cannot connect to engine server", err, map[string]any{})
		os.Exit(1)
	}

	client := gosdk.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		_ = gosdk.Error("cannot create input client", err, map[string]any{})
		// TODO: notify engine about this error before exit
		os.Exit(1)
	}

	for {
		log := <-logQueue
		err := inputClient.Send(log)
		if err != nil {
			_ = gosdk.Error("cannot send log", err, map[string]any{})
			// TODO: notify engine about this error before exit
			os.Exit(1)
		}

		_, err = inputClient.Recv()
		if err != nil {
			_ = gosdk.Error("cannot receive Ack", err, map[string]any{})
			// TODO: notify engine about this error before exit
			os.Exit(1)
		}
	}
}
