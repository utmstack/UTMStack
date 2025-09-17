package validations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateAzureConfig(config *config.ModuleGroup) error {
	var eventHubConnection, consumerGroup, storageContainer, storageConnection string

	if config == nil {
		return fmt.Errorf("AZURE configuration is nil")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "eventHubConnection":
			eventHubConnection = cnf.ConfValue
		case "consumerGroup":
			consumerGroup = cnf.ConfValue
		case "storageContainer":
			storageContainer = cnf.ConfValue
		case "storageConnection":
			storageConnection = cnf.ConfValue
		}
	}

	if eventHubConnection == "" {
		return fmt.Errorf("eventHubConnection is required in AZURE configuration")
	}
	if consumerGroup == "" {
		return fmt.Errorf("consumerGroup is required in AZURE configuration")
	}
	if storageContainer == "" {
		return fmt.Errorf("storageContainer is required in AZURE configuration")
	}
	if storageConnection == "" {
		return fmt.Errorf("storageConnection is required in AZURE configuration")
	}

	eventHubParts := strings.Split(eventHubConnection, ";EntityPath=")
	if len(eventHubParts) != 2 {
		return fmt.Errorf("invalid Event Hub connection string format: missing EntityPath")
	}
	eventHubConnectionBase := eventHubParts[0]
	eventHubName := eventHubParts[1]

	consumerClient, err := azeventhubs.NewConsumerClientFromConnectionString(eventHubConnectionBase, eventHubName, consumerGroup, nil)
	if err != nil {
		return fmt.Errorf("failed to create Event Hub consumer client: %w", err)
	}
	defer consumerClient.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = consumerClient.GetEventHubProperties(ctx, nil)
	if err != nil {
		return fmt.Errorf("Event Hub connection validation failed: %w", err)
	}

	blobClient, err := azblob.NewClientFromConnectionString(storageConnection, nil)
	if err != nil {
		return fmt.Errorf("failed to create Storage client: %w", err)
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(storageContainer)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	_, err = containerClient.GetProperties(ctx2, nil)
	if err != nil {
		return fmt.Errorf("Storage container validation failed: %w", err)
	}

	return nil
}
