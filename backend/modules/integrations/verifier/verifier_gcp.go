package verifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// verifyGCP — ported from modulekinds/gcp/validate.go.
func verifyGCP(c map[string]string) error {
	if err := required(c, "GCP", "jsonKey", "projectId", "subscription"); err != nil {
		return err
	}
	jsonKey, projectID, subscriptionID := c["jsonKey"], c["projectId"], c["subscription"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(jsonKey)))
	if err != nil {
		msg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(msg, "invalid"), strings.Contains(msg, "parse"), strings.Contains(msg, "json"):
			return fmt.Errorf("Invalid JSON Key format. Verify the service account key content is valid JSON.")
		case strings.Contains(msg, "project"):
			return fmt.Errorf("Invalid Project ID %q.", projectID)
		}
		return fmt.Errorf("Cannot connect to GCP Pub/Sub. Verify your JSON Key and Project ID %q.", projectID)
	}
	defer client.Close()

	exists, err := client.Subscription(subscriptionID).Exists(ctx)
	if err != nil {
		msg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(msg, "unauthenticated"):
			return fmt.Errorf("Invalid JSON Key. GCP rejected authentication.")
		case strings.Contains(msg, "permission"):
			return fmt.Errorf("The service account lacks Pub/Sub Subscriber permission in project %q.", projectID)
		case strings.Contains(msg, "not_found"), strings.Contains(msg, "not found"):
			return fmt.Errorf("GCP Project %q was not found.", projectID)
		}
		return fmt.Errorf("Cannot verify the Pub/Sub subscription in project %q.", projectID)
	}
	if !exists {
		return fmt.Errorf("Subscription %q not found in project %q.", subscriptionID, projectID)
	}
	return nil
}
