package client

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
)

type ClientDependencies struct {
	Backend     *BackendClient
	CM          *CustomersManagerClient
	ThreadWinds *ThreadWindsClient
	Alerts      *AlertClient
}

func NewClientDependencies(cfg *config.TWConfig) (*ClientDependencies, error) {
	catcher.Info("initializing client dependencies", nil)

	alerts, err := NewAlertClient(cfg)
	if err != nil {
		return nil, catcher.Error("failed to initialize the alert client", err, nil)
	}

	deps := &ClientDependencies{
		Backend:     NewBackendClient(cfg),
		CM:          &CustomersManagerClient{},
		ThreadWinds: NewThreadWindsClient(cfg),
		Alerts:      alerts,
	}

	return deps, nil
}
