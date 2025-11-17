package validations

import (
	"context"
	"fmt"
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
		case "client_id":
			clientID = cnf.ConfValue
		case "client_secret":
			clientSecret = cnf.ConfValue
		case "cloud":
			cloud = cnf.ConfValue
		case "app_name":
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

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Cloud:        falcon.Cloud(cloud),
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
