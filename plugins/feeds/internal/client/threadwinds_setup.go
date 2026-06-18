package client

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/utils"
)

func ConfigureThreadWindsCredentials(ctx context.Context, deps *ClientDependencies, twConfig *ThreadWindsConfig) error {
	if twConfig.APIKey == "" || twConfig.APISecret == "" {
		catcher.Info("ThreadWinds not configured, will attempt registration with retry...", nil)

		if err := deps.CM.LoadInstanceConfig(); err != nil {
			return catcher.Error("failed to load instance configuration", err, nil)
		}

		regResp, err := registerWithRetry(deps.CM)
		if err != nil {
			return catcher.Error("failed to register after all retry attempts", err, nil)
		}

		if err := deps.Backend.SaveThreadWindsCredentials(ctx,
			regResp.APIKey,
			regResp.APISecret); err != nil {
			return catcher.Error("failed to save ThreadWinds credentials", err, nil)
		}

		deps.ThreadWinds.UpdateCredentials(regResp.APIKey, regResp.APISecret)
		catcher.Info("ThreadWinds configured successfully with new credentials", nil)
	} else {
		catcher.Info("ThreadWinds already configured", nil)
		deps.ThreadWinds.UpdateCredentials(twConfig.APIKey, twConfig.APISecret)
	}

	return nil
}

func registerWithRetry(cm *CustomersManagerClient) (*RegistrationResponse, error) {
	var regResp *RegistrationResponse

	registerFunc := func() error {
		resp, err := cm.RegisterUserReporter()
		if err != nil {
			return err
		}
		regResp = resp
		return nil
	}

	if err := utils.Retry(registerFunc, "ThreadWinds registration", utils.DefaultRetryConfig()); err != nil {
		return nil, err
	}

	return regResp, nil
}
