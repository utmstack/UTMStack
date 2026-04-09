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
		return fmt.Errorf("CrowdStrike configuration is not provided")
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
		return fmt.Errorf("Client ID is required in CrowdStrike configuration")
	}
	if clientSecret == "" {
		return fmt.Errorf("Client Secret is required in CrowdStrike configuration")
	}
	if cloud == "" {
		return fmt.Errorf("Cloud Region is required in CrowdStrike configuration")
	}
	if appName == "" {
		return fmt.Errorf("App Name is required in CrowdStrike configuration")
	}

	cloudType, err := extractCloudFromURL(cloud)
	if err != nil {
		return err
	}

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Cloud:        cloudType,
		Context:      context.Background(),
	})
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
			return fmt.Errorf("Invalid Client ID or Client Secret. CrowdStrike rejected the authentication. Please verify your API credentials.")
		}
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
			return fmt.Errorf("Client ID does not have permission to authenticate. Please verify the API Client is enabled and has the required scopes.")
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Connection to CrowdStrike timed out. Please verify the Cloud Region '%s' is correct and accessible.", cloud)
		}
		if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
			return fmt.Errorf("Cannot resolve CrowdStrike API host. Please verify the Cloud Region '%s' is correct.", cloud)
		}
		return fmt.Errorf("Cannot authenticate with CrowdStrike. Please verify your Client ID, Client Secret, and Cloud Region are correct.")
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
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
			return fmt.Errorf("Client ID does not have permission to access Event Streams. Please add the 'Event streams: Read' scope to your API Client in the CrowdStrike console.")
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
			return fmt.Errorf("Invalid Client ID or Client Secret. Please verify your CrowdStrike API credentials.")
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("CrowdStrike Event Stream request timed out. Please try again.")
		}
		return fmt.Errorf("Cannot access CrowdStrike Event Streams. Please verify your credentials have the 'Event streams: Read' permission.")
	}

	if err = falcon.AssertNoError(response.Payload.Errors); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "access") || strings.Contains(errMsg, "permission") {
			return fmt.Errorf("CrowdStrike API returned a permission error. Please verify your API Client has the 'Event streams: Read' scope enabled.")
		}
		return fmt.Errorf("CrowdStrike API returned an error: %s. Please verify your API Client permissions.", errMsg)
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
		return 0, fmt.Errorf("Unrecognized CrowdStrike Cloud Region '%s'. Supported regions: api.crowdstrike.com (US-1), api.us-2.crowdstrike.com (US-2), api.eu-1.crowdstrike.com (EU-1).", trimmed)
	}

	ct, err := falcon.CloudValidate(trimmed)
	if err != nil {
		return 0, fmt.Errorf("Invalid CrowdStrike Cloud Region '%s'. Supported values: us-1, us-2, eu-1, us-gov-1.", trimmed)
	}
	return ct, nil
}
