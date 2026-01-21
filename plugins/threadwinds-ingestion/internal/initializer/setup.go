package initializer

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
)

func (a *App) configureThreadWinds(ctx context.Context) error {
	catcher.Info("configuring ThreadWinds credentials", nil)

	twConfig, err := a.clients.Backend.GetThreadWindsConfig(ctx)
	if err != nil {
		return catcher.Error("failed to check ThreadWinds configuration", err, nil)
	}

	err = client.ConfigureThreadWindsCredentials(ctx, a.clients, twConfig)
	if err != nil {
		return catcher.Error("failed to configure ThreadWinds", err, nil)
	}

	catcher.Info("ThreadWinds configured successfully", nil)
	return nil
}
