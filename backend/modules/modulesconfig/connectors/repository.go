package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
)

type ModuleRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.UtmModule, error)
	GetByServerAndName(ctx context.Context, serverID int64, moduleName string) (*domain.UtmModule, error)
	List(ctx context.Context, filter ModuleListFilter) ([]domain.UtmModule, int64, error)
	Save(ctx context.Context, module *domain.UtmModule) error
	Categories(ctx context.Context, serverID *int64) ([]string, error)
	CountActiveByName(ctx context.Context, moduleName string) (int64, error)
}

type ModuleListFilter struct {
	ServerID       *int64
	ModuleCategory *string
	ModuleActive   *bool
	NameContains   *string
	Page           int
	Size           int
}

type GroupRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.UtmModuleGroup, error)
	ListByModuleID(ctx context.Context, moduleID int64) ([]domain.UtmModuleGroup, error)
	Save(ctx context.Context, group *domain.UtmModuleGroup) error
	Delete(ctx context.Context, id int64) error
	DeleteAllByModuleID(ctx context.Context, moduleID int64) error
}

type ConfigRepository interface {
	ListByGroupID(ctx context.Context, groupID int64) ([]domain.UtmModuleGroupConfiguration, error)
	GetByGroupAndKey(ctx context.Context, groupID int64, confKey string) (*domain.UtmModuleGroupConfiguration, error)
	SaveAll(ctx context.Context, configs []domain.UtmModuleGroupConfiguration) error
}
