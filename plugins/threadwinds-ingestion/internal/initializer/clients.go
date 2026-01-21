package initializer

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
)

func (a *App) loadConfiguration() error {
	catcher.Info("loading configuration", nil)

	cfg, err := config.GetTWConfig()
	if err != nil {
		return catcher.Error("failed to load configuration", err, nil)
	}

	a.config = cfg
	return nil
}

func (a *App) initializeClients() error {
	catcher.Info("initializing clients", nil)

	clients, err := client.NewClientDependencies(a.config)
	if err != nil {
		return catcher.Error("failed to initialize clients", err, nil)
	}

	a.clients = clients

	catcher.Info("all client dependencies initialized successfully", nil)
	return nil
}
