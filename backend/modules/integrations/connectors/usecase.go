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
	IsActive(ctx context.Context, moduleName string) (bool, error)
}
