package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/repository"
	"github.com/utmstack/utmstack/backend/pkg/env"
)

// natsStartWait is how long the backend waits for NATS when it starts.
const natsStartWait = 90 * time.Second

// integrationsConfigStore returns where the module groups of the integrations are kept. With NATS_URL
// it is NATS, encrypted, which is where the integrations' containers read them; the files an earlier
// version left in the pipeline folder are moved there. Without it, the folder is used, as before.
func integrationsConfigStore(cipher connectors.Cipher) connectors.ConfigRepository {
	configDir := env.String("INTEGRATIONS_CONFIG_DIR", "/workdir/pipeline", false)

	url := env.String("NATS_URL", "", false)
	if url == "" {
		catcher.Info("integrations: NATS_URL is not set, keeping the module configuration in files", nil)

		return repository.NewConfigStore(configDir)
	}

	store, err := repository.NewNATSConfigStore(context.Background(), url, cipher, natsStartWait)
	if err != nil {
		// Falling back to files would split the configuration in two places, which is worse than
		// not starting.
		_ = catcher.Error("integrations: cannot reach the module configuration store in NATS", err, nil)
		panic("NATS is required for the integrations configuration: " + err.Error())
	}

	imported, err := store.MigrateFiles(context.Background(), configDir)
	if err != nil {
		_ = catcher.Error("integrations: some configuration files could not be moved to NATS; they were left in place", err, nil)
	}

	if imported > 0 {
		catcher.Info("integrations: moved the module configuration files into NATS", map[string]any{"integrations": imported})
	}

	return store
}
