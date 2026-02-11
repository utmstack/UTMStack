package validations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client/event_streams"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateCrowdstrikeConfig(config *config.ModuleGroup) error {
	var clientID, clientSecret, cloud, appName string

	if config == nil {
		return fmt.Errorf("CROWDSTRIKE configuration is nil")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "crowdstrike_client_id":
			clientID = cnf.ConfValue
		case "crowdstrike_client_secret":
			clientSecret = cnf.ConfValue
		case "crowdstrike_cloud_region_url":
			cloud = cnf.ConfValue
		case "crowdstrike_app_name":
			appName = cnf.ConfValue
		}
	}

	if clientID == "" {
		return fmt.Errorf("Client ID is required in CROWDSTRIKE configuration")
	}
	if clientSecret == "" {
		return fmt.Errorf("Client Secret is required in CROWDSTRIKE configuration")
	}
	if cloud == "" {
		return fmt.Errorf("Cloud is required in CROWDSTRIKE configuration")
	}
	if appName == "" {
		return fmt.Errorf("App Name is required in CROWDSTRIKE configuration")
	}

	cloudType, err := extractCloudFromURL(cloud)
	if err != nil {
		return fmt.Errorf("invalid cloud region configuration: %w", err)
	}

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Cloud:        cloudType,
		Context:      context.Background(),
	})
	if err != nil {
		return fmt.Errorf("failed to create CrowdStrike client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	json := "json"
	response, err := client.EventStreams.ListAvailableStreamsOAuth2(
		&event_streams.ListAvailableStreamsOAuth2Params{
			AppID:   appName,
			Format:  &json,
			Context: ctx,
		},
	)
	if err != nil {
		return fmt.Errorf("CrowdStrike credentials validation failed: %w", err)
	}

	if err = falcon.AssertNoError(response.Payload.Errors); err != nil {
		return fmt.Errorf("CrowdStrike API error during validation: %w", err)
	}

	return nil
}

func extractCloudFromURL(cloudValue string) (falcon.CloudType, error) {
	trimmed := strings.TrimSpace(cloudValue)

	urlToRegion := map[string]string{
		"api.crowdstrike.com":            "us-1",
		"api.us-2.crowdstrike.com":       "us-2",
		"api.eu-1.crowdstrike.com":       "eu-1",
		"api.laggar.gcw.crowdstrike.com": "us-gov-1",
		"api.us-gov-2.crowdstrike.mil":   "us-gov-2",
	}

	if strings.Contains(trimmed, "://") || strings.Contains(trimmed, ".crowdstrike.") {
		for host, region := range urlToRegion {
			if strings.Contains(trimmed, host) {
				return falcon.CloudValidate(region)
			}
		}
		return 0, fmt.Errorf("unrecognized CrowdStrike URL: %s", trimmed)
	}

	return falcon.CloudValidate(trimmed)
}
