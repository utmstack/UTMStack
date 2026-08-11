package initializer

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/client"
)

func (a *App) configureThreadWinds(ctx context.Context) error {
	catcher.Info("configuring ThreadWinds credentials", nil)

	// What the backend wrote for this plugin, already decrypted.
	twConfig := config.GetPluginConfig()

	err := client.ConfigureThreadWindsCredentials(ctx, a.clients, twConfig)
	if err != nil {
		return catcher.Error("failed to configure ThreadWinds", err, nil)
	}

	catcher.Info("ThreadWinds configured successfully", nil)
	return nil
}
