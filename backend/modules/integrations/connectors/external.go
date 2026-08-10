package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
)

type CredentialVerifier interface {
	Verify(integration string, config map[string]string) error
}

type Cipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type AgentManagerCollectorClient interface {
	ListCollectors(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListCollectorResponse, error)
	SetCollectorIntegration(ctx context.Context, collectorID uint32, dataType string, kv map[string]string) (*agent.ConfigKnowledge, error)
	SetCollectorCertificates(ctx context.Context, collectorID uint32, certPem, keyPem, caPem string) (*agent.ConfigKnowledge, error)
	GetCollectorCertificatesStatus(ctx context.Context, collectorID uint32) (*agent.ConfigKnowledge, error)
	GetCollectorIntegrationState(ctx context.Context, collectorID uint32, dataType string) (*agent.IntegrationStateResponse, error)
}
