package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/client"
)

func waitForBackend(bc *client.BackendClient) error {
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		err := bc.HealthCheck(context.Background())
		if err == nil {
			catcher.Info("Connected to Backend", map[string]any{
				"process": "compliance-orchestrator",
			})
			return nil
		}

		catcher.Error("Cannot connect to Backend, retrying", err, map[string]any{
			"retry": retry + 1,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		} else {
			return err
		}
	}

	return nil
}

func waitForOpenSearch() error {
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		err := client.ConnectOpenSearch()
		if err == nil {
			return nil
		}

		catcher.Error("Cannot connect to OpenSearch, retrying", err, map[string]any{
			"retry": retry + 1,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		} else {
			return err
		}
	}

	return nil
}

func bootstrap() (*client.BackendClient, error) {
	backend := client.NewBackendClient()

	if err := waitForBackend(backend); err != nil {
		return nil, err
	}

	if err := waitForOpenSearch(); err != nil {
		return nil, err
	}

	return backend, nil
}
