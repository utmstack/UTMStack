package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
	"google.golang.org/api/option"
)

func (k *kind) ValidateConfiguration(ctx context.Context, _ *domain.UtmModule, configs []domain.UtmModuleGroupConfiguration) error {
	if err := baseline.RequireFields(configs, "GCP",
		"jsonKey", "JSON Key",
		"projectId", "Project ID",
		"subscription", "Subscription ID",
	); err != nil {
		return err
	}

	jsonKey := baseline.ConfigValue(configs, "jsonKey")
	projectID := baseline.ConfigValue(configs, "projectId")
	subscriptionID := baseline.ConfigValue(configs, "subscription")

	timed, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(timed, projectID, option.WithCredentialsJSON([]byte(jsonKey)))
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "invalid"), strings.Contains(errMsg, "parse"), strings.Contains(errMsg, "json"):
			return fmt.Errorf("Invalid JSON Key format. Please verify the service account key file content is correct and valid JSON.")
		case strings.Contains(errMsg, "project"):
			return fmt.Errorf("Invalid Project ID '%s'. Please verify the GCP Project ID is correct.", projectID)
		}
		return fmt.Errorf("Cannot connect to GCP Pub/Sub. Please verify your JSON Key and Project ID '%s' are correct.", projectID)
	}
	defer client.Close()

	exists, err := client.Subscription(subscriptionID).Exists(timed)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "unauthenticated"):
			return fmt.Errorf("Invalid JSON Key. GCP rejected the authentication. Please verify the service account key is correct and not expired.")
		case strings.Contains(errMsg, "permission_denied"), strings.Contains(errMsg, "permission denied"):
			return fmt.Errorf("The service account does not have permission to access Pub/Sub in project '%s'. Please add the 'Pub/Sub Subscriber' role to the service account.", projectID)
		case strings.Contains(errMsg, "not_found"), strings.Contains(errMsg, "not found"):
			return fmt.Errorf("GCP Project '%s' was not found. Please verify the Project ID is correct.", projectID)
		case strings.Contains(errMsg, "timeout"), strings.Contains(errMsg, "deadline"):
			return fmt.Errorf("Connection to GCP timed out. Please verify the service account has network access to Google Cloud APIs.")
		}
		return fmt.Errorf("Cannot verify the Pub/Sub Subscription in project '%s'. Please check that the JSON Key has permission to access Pub/Sub.", projectID)
	}
	if !exists {
		return fmt.Errorf("The Pub/Sub Subscription '%s' was not found in project '%s'. Please verify the Subscription ID and Project ID are correct.", subscriptionID, projectID)
	}
	return nil
}
