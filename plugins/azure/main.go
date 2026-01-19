package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/plugins/azure/config"
)

type AzureCloud string

const (
	defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	wait                 = 1 * time.Second

	AzurePublic     AzureCloud = "AzurePublic"
	AzureGovernment AzureCloud = "AzureGovernment"
	AzureChina      AzureCloud = "AzureChina"
)

type CloudEndpoints struct {
	Name           AzureCloud
	EventHubSuffix string
	StorageSuffix  string
	LoginAuthority string
	Description    string
}

var SupportedClouds = []CloudEndpoints{
	{
		Name:           AzureGovernment,
		EventHubSuffix: ".servicebus.usgovcloudapi.net",
		StorageSuffix:  ".core.usgovcloudapi.net",
		LoginAuthority: "https://login.microsoftonline.us/",
		Description:    "Azure Government (US)",
	},
	{
		Name:           AzureChina,
		EventHubSuffix: ".servicebus.chinacloudapi.cn",
		StorageSuffix:  ".core.chinacloudapi.cn",
		LoginAuthority: "https://login.chinacloudapi.cn/",
		Description:    "Azure China (21Vianet)",
	},
	{
		Name:           AzurePublic,
		EventHubSuffix: ".servicebus.windows.net",
		StorageSuffix:  ".core.windows.net",
		LoginAuthority: "https://login.microsoftonline.com/",
		Description:    "Azure Public Cloud",
	},
}

func main() {
	if os.Getenv("PLAYGROUND") == "true" {
		return
	}

	mode := plugins.GetCfg("plugin_com.utmstack.azure").Env.Mode
	if mode != "worker" {
		return
	}

	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.azure")
		}()
	}

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for range ticker.C {
		moduleConfig := config.GetConfig()
		if moduleConfig != nil && moduleConfig.ModuleActive {
			cloudsInUse := detectCloudsInUse(moduleConfig)
			for cloudName, loginAuthority := range cloudsInUse {
				if err := connectionChecker(loginAuthority); err != nil {
					catcher.Info("Airgap or limited connectivity detected", map[string]any{
						"cloud":          cloudName,
						"loginAuthority": loginAuthority,
						"process":        "plugin_com.utmstack.azure",
					})
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ModuleGroups))
			for _, grp := range moduleConfig.ModuleGroups {
				go func(group *config.ModuleGroup) {
					defer wg.Done()
					var invalid bool
					for _, cnf := range group.ModuleGroupConfigurations {
						if strings.TrimSpace(cnf.ConfValue) == "" {
							invalid = true
							break
						}
					}
					if !invalid {
						pull(group)
					}
				}(grp)
			}

			wg.Wait()
		}

	}
}

func detectCloudsInUse(moduleConfig *config.ConfigurationSection) map[string]string {
	cloudsMap := make(map[string]string)

	for _, group := range moduleConfig.ModuleGroups {
		for _, cnf := range group.ModuleGroupConfigurations {
			if cnf.ConfKey == "eventHubConnection" || cnf.ConfKey == "storageConnection" {
				if cloud, err := detectCloudFromConnectionString(cnf.ConfValue); err == nil {
					cloudsMap[string(cloud.Name)] = cloud.LoginAuthority
				}
			}
		}
	}

	return cloudsMap
}

func detectCloudFromConnectionString(connectionString string) (CloudEndpoints, error) {
	if connectionString == "" {
		return CloudEndpoints{}, fmt.Errorf("connection string is empty")
	}

	for _, cloud := range SupportedClouds {
		if strings.Contains(connectionString, cloud.EventHubSuffix+"/") {
			return cloud, nil
		}

		if strings.Contains(connectionString, "EndpointSuffix="+cloud.StorageSuffix) {
			return cloud, nil
		}
	}

	return CloudEndpoints{}, fmt.Errorf("unable to detect Azure cloud from connection string")
}

func pull(group *config.ModuleGroup) {
	agent := getAzureProcessor(group)

	if agent.EventHubConnection == "" || agent.ConsumerGroup == "" ||
		agent.StorageContainer == "" || agent.StorageConnection == "" {
		_ = catcher.Error("missing required configuration for Event Hub", nil, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	eventHubParts := strings.Split(agent.EventHubConnection, ";EntityPath=")
	if len(eventHubParts) != 2 {
		_ = catcher.Error("invalid Event Hub connection string format", nil, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	eventHubConnection := eventHubParts[0]
	eventHubName := eventHubParts[1]

	blobClient, err := azblob.NewClientFromConnectionString(agent.StorageConnection, nil)
	if err != nil {
		_ = catcher.Error("cannot create blob client", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	checkpointStore, err := checkpoints.NewBlobStore(
		blobClient.ServiceClient().NewContainerClient(agent.StorageContainer), nil)
	if err != nil {
		_ = catcher.Error("cannot create checkpoint store", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	maxRetries := 3
	retryDelay := 2 * time.Second
	var client *azeventhubs.ConsumerClient

	for retry := 0; retry < maxRetries; retry++ {
		client, err = azeventhubs.NewConsumerClientFromConnectionString(
			eventHubConnection, eventHubName, agent.ConsumerGroup, nil)
		if err == nil {
			break
		}

		_ = catcher.Error("cannot create Event Hub consumer client, retrying", err, map[string]any{
			"group":      agent.GroupName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    "plugin_com.utmstack.azure",
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		_ = catcher.Error("all retries failed when creating Event Hub consumer client", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}
	defer client.Close(context.Background())

	processor, err := azeventhubs.NewProcessor(client, checkpointStore, nil)
	if err != nil {
		_ = catcher.Error("cannot create Event Hub processor", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	go func() {
		for {
			pc := processor.NextPartitionClient(ctx)
			if pc == nil {
				return
			}
			go processPartition(pc, agent.GroupName)
		}
	}()

	if err := processor.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_ = catcher.Error("error running Event Hub processor", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
	}
}

func processPartition(pc *azeventhubs.ProcessorPartitionClient, groupName string) {
	defer pc.Close(context.Background())

	for {
		recvCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		events, err := pc.ReceiveEvents(recvCtx, 100, nil)
		cancel()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			_ = catcher.Error("error receiving events", err, map[string]any{
				"group":       groupName,
				"partitionID": pc.PartitionID(),
				"process":     "plugin_com.utmstack.azure",
			})
			return
		}

		if len(events) == 0 {
			continue
		}

		for _, event := range events {
			var logData map[string]any
			if err := json.Unmarshal(event.Body, &logData); err != nil {
				_ = catcher.Error("cannot parse event body", err, map[string]any{
					"group":       groupName,
					"partitionID": pc.PartitionID(),
					"process":     "plugin_com.utmstack.azure",
				})
				continue
			}

			if records, ok := logData["records"].([]any); ok && len(records) > 0 {
				for _, record := range records {
					recordMap, ok := record.(map[string]any)
					if !ok {
						_ = catcher.Error("invalid record format in records array", nil, map[string]any{
							"group":       groupName,
							"partitionID": pc.PartitionID(),
							"process":     "plugin_com.utmstack.azure",
						})
						continue
					}

					jsonLog, err := json.Marshal(recordMap)
					if err != nil {
						_ = catcher.Error("cannot encode record to JSON", err, map[string]any{
							"group":       groupName,
							"partitionID": pc.PartitionID(),
							"process":     "plugin_com.utmstack.azure",
						})
						continue
					}

					plugins.EnqueueLog(&plugins.Log{
						Id:         uuid.New().String(),
						TenantId:   defaultTenant,
						DataType:   "azure",
						DataSource: groupName,
						Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
						Raw:        string(jsonLog),
					}, "com.utmstack.azure")
				}
			} else {
				jsonLog, err := json.Marshal(logData)
				if err != nil {
					_ = catcher.Error("cannot encode log to JSON", err, map[string]any{
						"group":       groupName,
						"partitionID": pc.PartitionID(),
						"process":     "plugin_com.utmstack.azure",
					})
					continue
				}

				plugins.EnqueueLog(&plugins.Log{
					Id:         uuid.New().String(),
					TenantId:   defaultTenant,
					DataType:   "azure",
					DataSource: groupName,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        string(jsonLog),
				}, "com.utmstack.azure")
			}
		}

		if err := pc.UpdateCheckpoint(context.Background(), events[len(events)-1], nil); err != nil {
			_ = catcher.Error("checkpoint error", err, map[string]any{
				"group":       groupName,
				"partitionID": pc.PartitionID(),
				"process":     "plugin_com.utmstack.azure",
			})
		}
	}
}

type AzureConfig struct {
	GroupName          string
	EventHubConnection string
	ConsumerGroup      string
	StorageContainer   string
	StorageConnection  string
}

func getAzureProcessor(group *config.ModuleGroup) AzureConfig {
	azurePro := AzureConfig{}
	azurePro.GroupName = group.GroupName
	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "eventHubConnection":
			azurePro.EventHubConnection = cnf.ConfValue
		case "consumerGroup":
			azurePro.ConsumerGroup = cnf.ConfValue
		case "storageContainer":
			azurePro.StorageContainer = cnf.ConfValue
		case "storageConnection":
			azurePro.StorageConnection = cnf.ConfValue
		}
	}
	return azurePro
}

func connectionChecker(url string) error {
	checkConn := func() error {
		if err := checkConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_ = catcher.Error("cannot close response body", err, map[string]any{"process": "plugin_com.utmstack.azure"})
		}
	}()

	return nil
}

func infiniteRetryIfXError(f func() error, exception string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exception) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, map[string]any{"process": "azure-plugin"})
				xErrorWasLogged = true
			}
			time.Sleep(wait)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
