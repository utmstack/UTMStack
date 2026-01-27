package client

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
)

type ClientDependencies struct {
	Backend     *BackendClient
	CM          *CustomersManagerClient
	ThreadWinds *ThreadWindsClient
	OpenSearch  *OpenSearchClient
}

func NewClientDependencies(cfg *config.TWConfig) (*ClientDependencies, error) {
	catcher.Info("initializing client dependencies", nil)

	opensearch, err := NewOpenSearchClient(cfg)
	if err != nil {
		return nil, catcher.Error("failed to initialize opensearch client", err, nil)
	}

	deps := &ClientDependencies{
		Backend:     NewBackendClient(cfg),
		CM:          &CustomersManagerClient{},
		ThreadWinds: NewThreadWindsClient(cfg),
		OpenSearch:  opensearch,
	}

	return deps, nil
}
