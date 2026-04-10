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
		return fmt.Errorf("Azure configuration is not provided")
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
		return fmt.Errorf("Event Hub Connection String is required in Azure configuration")
	}
	if consumerGroup == "" {
		return fmt.Errorf("Consumer Group is required in Azure configuration")
	}
	if storageContainer == "" {
		return fmt.Errorf("Storage Container is required in Azure configuration")
	}
	if storageConnection == "" {
		return fmt.Errorf("Storage Connection String is required in Azure configuration")
	}

	eventHubParts := strings.Split(eventHubConnection, ";EntityPath=")
	if len(eventHubParts) != 2 {
		return fmt.Errorf("Invalid Event Hub Connection String format. It must include an EntityPath (e.g., 'Endpoint=sb://...;SharedAccessKeyName=...;SharedAccessKey=...;EntityPath=your-hub-name').")
	}
	eventHubConnectionBase := eventHubParts[0]
	eventHubName := eventHubParts[1]

	consumerClient, err := azeventhubs.NewConsumerClientFromConnectionString(eventHubConnectionBase, eventHubName, consumerGroup, nil)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Endpoint") || strings.Contains(errMsg, "parse") {
			return fmt.Errorf("Invalid Event Hub Connection String format. Please verify the Endpoint URL and SharedAccessKey are correct.")
		}
		return fmt.Errorf("Cannot connect to the Azure Event Hub. Please verify the Event Hub Connection String is correct.")
	}
	defer consumerClient.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = consumerClient.GetEventHubProperties(ctx, nil)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
			return fmt.Errorf("Event Hub Connection String is unauthorized. Please verify the SharedAccessKey has the required permissions (Listen).")
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not-found") {
			return fmt.Errorf("Event Hub '%s' was not found. Please verify the EntityPath in the Connection String is correct.", eventHubName)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Connection to Event Hub timed out. Please verify the Endpoint URL is accessible from this server.")
		}
		if strings.Contains(errMsg, "consumer group") || strings.Contains(errMsg, "consumergroup") {
			return fmt.Errorf("Consumer Group '%s' was not found on Event Hub '%s'. Please verify the Consumer Group name.", consumerGroup, eventHubName)
		}
		return fmt.Errorf("Cannot connect to Event Hub '%s'. Please verify the Connection String and Consumer Group '%s' are correct.", eventHubName, consumerGroup)
	}

	blobClient, err := azblob.NewClientFromConnectionString(storageConnection, nil)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "invalid") {
			return fmt.Errorf("Invalid Storage Connection String format. Please verify the Storage Connection String is correct.")
		}
		return fmt.Errorf("Cannot connect to Azure Storage. Please verify the Storage Connection String is correct.")
	}

	containerClient := blobClient.ServiceClient().NewContainerClient(storageContainer)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	_, err = containerClient.GetProperties(ctx2, nil)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "containernotfound") || strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") {
			return fmt.Errorf("Storage Container '%s' was not found. Please verify the container name exists in the Storage Account.", storageContainer)
		}
		if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
			return fmt.Errorf("Access denied to Storage Container '%s'. Please verify the Storage Connection String has the required permissions.", storageContainer)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Connection to Azure Storage timed out. Please verify the Storage Connection String and network connectivity.")
		}
		return fmt.Errorf("Cannot access Storage Container '%s'. Please verify the container name and Storage Connection String.", storageContainer)
	}

	return nil
}
