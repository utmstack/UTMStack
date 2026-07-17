package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
)

type ModuleUsecase interface {
	ActivateDeactivate(ctx context.Context, req dto.ModuleActivationRequest) (*dto.ModuleResponse, error)
	Create(ctx context.Context, req dto.CreateModuleRequest) (*dto.ModuleResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateModuleRequest) (*dto.ModuleResponse, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter ModuleListFilter) ([]dto.ModuleResponse, int64, error)
	GetByID(ctx context.Context, id int64) (*dto.ModuleResponse, error)
	GetByName(ctx context.Context, moduleName string) (*dto.ModuleResponse, error)
	Categories(ctx context.Context) ([]string, error)
	DataTypes(ctx context.Context) ([]dto.DataTypeOption, error)
	IsActive(ctx context.Context, moduleName string) (bool, error)
}

type CollectorUsecase interface {
	ListForwarders(ctx context.Context) ([]dto.CollectorResponse, error)
	SetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string, req dto.SetDataTypeConfigRequest) (*dto.ConfigKnowledgeResponse, error)
	GetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string) (*dto.GetDataTypeConfigResponse, error)
	SetForwarderCertificates(ctx context.Context, collectorID uint32, req dto.SetForwarderCertificatesRequest) (*dto.ConfigKnowledgeResponse, error)
	GetTLSStatus(ctx context.Context, collectorID uint32) (*dto.TLSStatusResponse, error)
}
