package validations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
	"google.golang.org/api/option"
)

func ValidateGcpConfig(config *config.ModuleGroup) error {
	var jsonKey, projectID, subscriptionID string

	if config == nil {
		return fmt.Errorf("GCP configuration is not provided")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "jsonKey":
			jsonKey = cnf.ConfValue
		case "projectId":
			projectID = cnf.ConfValue
		case "subscription":
			subscriptionID = cnf.ConfValue
		}
	}

	if jsonKey == "" {
		return fmt.Errorf("JSON Key is required in GCP configuration")
	}
	if projectID == "" {
		return fmt.Errorf("Project ID is required in GCP configuration")
	}
	if subscriptionID == "" {
		return fmt.Errorf("Subscription ID is required in GCP configuration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(jsonKey)))
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "json") {
			return fmt.Errorf("Invalid JSON Key format. Please verify the service account key file content is correct and valid JSON.")
		}
		if strings.Contains(errMsg, "project") {
			return fmt.Errorf("Invalid Project ID '%s'. Please verify the GCP Project ID is correct.", projectID)
		}
		return fmt.Errorf("Cannot connect to GCP Pub/Sub. Please verify your JSON Key and Project ID '%s' are correct.", projectID)
	}
	defer client.Close()

	subscription := client.Subscription(subscriptionID)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "unauthenticated") {
			return fmt.Errorf("Invalid JSON Key. GCP rejected the authentication. Please verify the service account key is correct and not expired.")
		}
		if strings.Contains(errMsg, "permission_denied") || strings.Contains(errMsg, "permission denied") {
			return fmt.Errorf("The service account does not have permission to access Pub/Sub in project '%s'. Please add the 'Pub/Sub Subscriber' role to the service account.", projectID)
		}
		if strings.Contains(errMsg, "not_found") || strings.Contains(errMsg, "not found") {
			return fmt.Errorf("GCP Project '%s' was not found. Please verify the Project ID is correct.", projectID)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Connection to GCP timed out. Please verify the service account has network access to Google Cloud APIs.")
		}
		return fmt.Errorf("Cannot verify the Pub/Sub Subscription in project '%s'. Please check that the JSON Key has permission to access Pub/Sub.", projectID)
	}

	if !exists {
		return fmt.Errorf("The Pub/Sub Subscription '%s' was not found in project '%s'. Please verify the Subscription ID and Project ID are correct.", subscriptionID, projectID)
	}

	return nil
}
