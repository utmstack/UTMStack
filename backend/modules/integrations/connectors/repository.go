package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ConfigRepository interface {
	Load(ctx context.Context, integration string) ([]domain.ConfigGroup, error)
	LoadAllTenants(ctx context.Context, integration string) ([]domain.TenantConfig, error)
	Save(ctx context.Context, integration string, groups []domain.ConfigGroup) error
	Upsert(ctx context.Context, integration string, group domain.ConfigGroup) error
	Delete(ctx context.Context, integration, groupName string) error
}

type IntegrationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error)
	GetByName(ctx context.Context, name string) (*domain.Integration, error)
	List(ctx context.Context, filter IntegrationListFilter) ([]domain.Integration, int64, error)
	Save(ctx context.Context, integration *domain.Integration) error
	Delete(ctx context.Context, id uuid.UUID) error
	DataTypes(ctx context.Context) ([]domain.Integration, error)
}

type IntegrationListFilter struct {
	NameContains *string
	IngestType   *domain.IngestType
	database.Params
}

type SchemaProvider interface {
	Schema(integration string) (map[string]string, error)
}
