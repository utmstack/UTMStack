package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
)

type ModuleUsecase interface {
	ActivateDeactivate(ctx context.Context, req dto.ModuleActivationRequest) (*dto.ModuleResponse, error)
	List(ctx context.Context, filter ModuleListFilter) ([]dto.ModuleResponse, int64, error)
	GetByID(ctx context.Context, id int64) (*dto.ModuleResponse, error)
	Details(ctx context.Context, serverID int64, moduleName string, reveal bool) (*dto.ModuleResponse, error)
	CheckRequirements(ctx context.Context, serverID int64, moduleName string) (*dto.CheckRequirementsResponse, error)
	Categories(ctx context.Context, serverID *int64) ([]string, error)
	IsActive(ctx context.Context, moduleName string) (bool, error)
}

type GroupUsecase interface {
	Create(ctx context.Context, req dto.CreateModuleGroupRequest) (*dto.ModuleGroupResponse, error)
	Update(ctx context.Context, req dto.UpdateModuleGroupRequest) (*dto.ModuleGroupResponse, error)
	ListByModuleID(ctx context.Context, moduleID int64) ([]dto.ModuleGroupResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.ModuleGroupResponse, error)
	Delete(ctx context.Context, id int64) error
}

type ConfigUsecase interface {
	ListByGroupID(ctx context.Context, groupID int64) ([]dto.GroupConfigurationItem, error)
	GetByGroupAndKey(ctx context.Context, groupID int64, confKey string) (*dto.GroupConfigurationItem, error)
	Update(ctx context.Context, req dto.UpdateGroupConfigurationRequest) error
}

type ModuleFactory interface {
	Register(kind ModuleKind)
	Get(name string) (ModuleKind, bool)
	Has(name string) bool
}

type ModuleKind interface {
	Name() string
	ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey
	CheckRequirements(ctx context.Context, serverID int64) ([]domain.ModuleRequirement, error)
	ValidateConfiguration(ctx context.Context, module *domain.UtmModule, configs []domain.UtmModuleGroupConfiguration) error
	UpdateModule(ctx context.Context, nameShort string,  configs []domain.UtmModuleGroupConfiguration) error
}
