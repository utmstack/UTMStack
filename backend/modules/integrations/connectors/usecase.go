package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
)

type IntegrationUsecase interface {
	Create(ctx context.Context, req dto.CreateIntegrationRequest) (*dto.IntegrationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateIntegrationRequest) (*dto.IntegrationResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter IntegrationListFilter) ([]dto.IntegrationResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.IntegrationResponse, error)
	GetByName(ctx context.Context, name string) (*dto.IntegrationResponse, error)
	DataTypes(ctx context.Context) ([]dto.DataTypeOption, error)
}

type CollectorUsecase interface {
	ListForwarders(ctx context.Context) ([]dto.CollectorResponse, error)
	SetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string, req dto.SetDataTypeConfigRequest) (*dto.ConfigKnowledgeResponse, error)
	GetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string) (*dto.GetDataTypeConfigResponse, error)
	SetForwarderCertificates(ctx context.Context, collectorID uint32, req dto.SetForwarderCertificatesRequest) (*dto.ConfigKnowledgeResponse, error)
	GetTLSStatus(ctx context.Context, collectorID uint32) (*dto.TLSStatusResponse, error)
}
