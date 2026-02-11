package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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

	receiveTimeout    = 30 * time.Second
	maxEventsPerBatch = 100

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

type ProcessorManager struct {
	processors sync.Map
}

type ActiveProcessor struct {
	cancel context.CancelFunc
	config AzureConfig
	done   chan struct{}
}

var processorManager = &ProcessorManager{}

func main() {
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

	catcher.Info("Azure plugin started", map[string]any{
		"process": "plugin_com.utmstack.azure",
	})

	processorManager.syncProcessors()

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for range ticker.C {
		processorManager.syncProcessors()
	}
}

func (pm *ProcessorManager) syncProcessors() {
	moduleConfig := config.GetConfig()
	if moduleConfig == nil || !moduleConfig.ModuleActive {
		pm.stopAll()
		return
	}

	cloudsInUse := detectCloudsInUse(moduleConfig)
	for cloudName, loginAuthority := range cloudsInUse {
		if err := connectionChecker(loginAuthority); err != nil {
			catcher.Info("airgap or limited connectivity detected", map[string]any{
				"cloud":          cloudName,
				"loginAuthority": loginAuthority,
				"process":        "plugin_com.utmstack.azure",
			})
		}
	}

	currentGroups := make(map[string]*config.ModuleGroup)
	for _, grp := range moduleConfig.ModuleGroups {
		valid := true
		for _, cnf := range grp.ModuleGroupConfigurations {
			if strings.TrimSpace(cnf.ConfValue) == "" {
				valid = false
				break
			}
		}
		if valid {
			currentGroups[grp.GroupName] = grp
		}
	}

	pm.processors.Range(func(key, value any) bool {
		groupName := key.(string)
		if _, exists := currentGroups[groupName]; !exists {
			pm.stop(groupName)
		}
		return true
	})

	for groupName, grp := range currentGroups {
		newConfig := getAzureProcessor(grp)

		if existing, ok := pm.processors.Load(groupName); ok {
			activeProc := existing.(*ActiveProcessor)

			if configChanged(activeProc.config, newConfig) {
				pm.stop(groupName)
				pm.start(groupName, newConfig)
			} else {
				select {
				case <-activeProc.done:
					pm.start(groupName, newConfig)
				default:
				}
			}
		} else {
			pm.start(groupName, newConfig)
		}
	}
}

func (pm *ProcessorManager) start(groupName string, config AzureConfig) {
	ctx, cancel := context.WithCancel(context.Background())

	activeProc := &ActiveProcessor{
		cancel: cancel,
		config: config,
		done:   make(chan struct{}),
	}

	pm.processors.Store(groupName, activeProc)

	go func() {
		defer close(activeProc.done)
		runProcessor(ctx, config)
	}()
}

func (pm *ProcessorManager) stop(groupName string) {
	if value, ok := pm.processors.LoadAndDelete(groupName); ok {
		activeProc := value.(*ActiveProcessor)
		activeProc.cancel()

		select {
		case <-activeProc.done:
		case <-time.After(30 * time.Second):
		}
	}
}

func (pm *ProcessorManager) stopAll() {
	pm.processors.Range(func(key, value any) bool {
		groupName := key.(string)
		pm.stop(groupName)
		return true
	})
}

func configChanged(old, new AzureConfig) bool {
	return old.EventHubConnection != new.EventHubConnection ||
		old.ConsumerGroup != new.ConsumerGroup ||
		old.StorageContainer != new.StorageContainer ||
		old.StorageConnection != new.StorageConnection
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

func runProcessor(ctx context.Context, agent AzureConfig) {
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
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
				retryDelay *= 2
			}
		}
	}

	if err != nil {
		_ = catcher.Error("all retries failed when creating Event Hub consumer client", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			_ = catcher.Error("error closing consumer client", err, map[string]any{
				"group":   agent.GroupName,
				"process": "plugin_com.utmstack.azure",
			})
		}
	}()

	processor, err := azeventhubs.NewProcessor(client, checkpointStore, &azeventhubs.ProcessorOptions{
		StartPositions: azeventhubs.StartPositions{
			Default: azeventhubs.StartPosition{
				Earliest: to.Ptr(true),
			},
		},
	})
	if err != nil {
		_ = catcher.Error("cannot create Event Hub processor", err, map[string]any{
			"group":   agent.GroupName,
			"process": "plugin_com.utmstack.azure",
		})
		return
	}

	var partitionsWg sync.WaitGroup

	go func() {
		for {
			pc := processor.NextPartitionClient(ctx)
			if pc == nil {
				return
			}

			partitionsWg.Add(1)
			go processPartition(ctx, pc, agent.GroupName, &partitionsWg)
		}
	}()

	processorDone := make(chan error, 1)
	go func() {
		processorDone <- processor.Run(ctx)
	}()

	select {
	case <-ctx.Done():
	case err := <-processorDone:
		if err != nil && !errors.Is(err, context.Canceled) {
			_ = catcher.Error("processor stopped with error", err, map[string]any{
				"group":   agent.GroupName,
				"process": "plugin_com.utmstack.azure",
			})
		}
	}

	partitionsWg.Wait()
}

func processPartition(ctx context.Context, pc *azeventhubs.ProcessorPartitionClient, groupName string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := pc.Close(context.Background()); err != nil {
			_ = catcher.Error("error closing partition client", err, map[string]any{
				"group":       groupName,
				"partitionID": pc.PartitionID(),
				"process":     "plugin_com.utmstack.azure",
			})
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		recvCtx, cancel := context.WithTimeout(ctx, receiveTimeout)
		events, err := pc.ReceiveEvents(recvCtx, maxEventsPerBatch, nil)
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
			_ = plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.New().String(),
				TenantId:   defaultTenant,
				DataType:   "azure",
				DataSource: groupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(event.Body),
			}, "com.utmstack.azure")
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
