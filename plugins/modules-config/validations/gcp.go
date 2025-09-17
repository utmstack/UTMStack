package validations

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
	"google.golang.org/api/option"
)

func ValidateGcpConfig(config *config.ModuleGroup) error {
	var jsonKey, projectID, subscriptionID string

	if config == nil {
		return fmt.Errorf("GCP configuration is nil")
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
		return fmt.Errorf("failed to create GCP PubSub client: %w", err)
	}
	defer client.Close()

	subscription := client.Subscription(subscriptionID)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify GCP subscription: %w", err)
	}

	if !exists {
		return fmt.Errorf("GCP subscription '%s' does not exist in project '%s'", subscriptionID, projectID)
	}

	return nil
}
