package verifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client/event_streams"
)

// verifyCrowdStrike — ported from modulekinds/crowdstrike/validate.go.
func verifyCrowdStrike(c map[string]string) error {
	if err := required(c, "CrowdStrike",
		"crowdstrike_client_id",
		"crowdstrike_client_secret",
		"crowdstrike_cloud_region_url",
		"crowdstrike_app_name",
	); err != nil {
		return err
	}

	clientID := c["crowdstrike_client_id"]
	clientSecret := c["crowdstrike_client_secret"]
	cloud := c["crowdstrike_cloud_region_url"]
	appName := c["crowdstrike_app_name"]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cloudType, err := extractCrowdStrikeCloudFromURL(cloud)
	if err != nil {
		return err
	}

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Cloud:        cloudType,
		Context:      ctx,
	})
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "401"), strings.Contains(errMsg, "unauthorized"):
			return fmt.Errorf("Invalid Client ID or Client Secret. CrowdStrike rejected the authentication. Please verify your API credentials.")
		case strings.Contains(errMsg, "403"), strings.Contains(errMsg, "forbidden"):
			return fmt.Errorf("Client ID does not have permission to authenticate. Please verify the API Client is enabled and has the required scopes.")
		case strings.Contains(errMsg, "timeout"), strings.Contains(errMsg, "deadline"):
			return fmt.Errorf("Connection to CrowdStrike timed out. Please verify the Cloud Region '%s' is correct and accessible.", cloud)
		case strings.Contains(errMsg, "no such host"), strings.Contains(errMsg, "lookup"):
			return fmt.Errorf("Cannot resolve CrowdStrike API host. Please verify the Cloud Region '%s' is correct.", cloud)
		}
		return fmt.Errorf("Cannot authenticate with CrowdStrike. Please verify your Client ID, Client Secret, and Cloud Region are correct.")
	}

	streamCtx, streamCancel := context.WithTimeout(ctx, 30*time.Second)
	defer streamCancel()

	jsonFormat := "json"
	response, err := client.EventStreams.ListAvailableStreamsOAuth2(&event_streams.ListAvailableStreamsOAuth2Params{
		AppID:   appName,
		Format:  &jsonFormat,
		Context: streamCtx,
	})
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "403"), strings.Contains(errMsg, "forbidden"):
			return fmt.Errorf("Client ID does not have permission to access Event Streams. Please add the 'Event streams: Read' scope to your API Client in the CrowdStrike console.")
		case strings.Contains(errMsg, "401"), strings.Contains(errMsg, "unauthorized"):
			return fmt.Errorf("Invalid Client ID or Client Secret. Please verify your CrowdStrike API credentials.")
		case strings.Contains(errMsg, "timeout"), strings.Contains(errMsg, "deadline"):
			return fmt.Errorf("CrowdStrike Event Stream request timed out. Please try again.")
		}
		return fmt.Errorf("Cannot access CrowdStrike Event Streams. Please verify your credentials have the 'Event streams: Read' permission.")
	}

	if err := falcon.AssertNoError(response.Payload.Errors); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "access") || strings.Contains(errMsg, "permission") {
			return fmt.Errorf("CrowdStrike API returned a permission error. Please verify your API Client has the 'Event streams: Read' scope enabled.")
		}
		return fmt.Errorf("CrowdStrike API returned an error: %s. Please verify your API Client permissions.", errMsg)
	}
	return nil
}

func extractCrowdStrikeCloudFromURL(cloudValue string) (falcon.CloudType, error) {
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
