package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ADUserRepository interface {
	Upsert(ctx context.Context, users []domain.ADUser) error
	List(ctx context.Context, f dto.ADUserFilter) ([]domain.ADUser, int64, error)
	ListWithIdentity(ctx context.Context, f dto.ADUserFilter) ([]dto.UserRow, int64, error)
	Each(ctx context.Context, source string, fn func(domain.ADUser) error) error
	Stats(ctx context.Context) (*dto.ADUserStats, error)
	ResolveLinuxIdentity(ctx context.Context, tenantID, hostname, machineID string) (int64, error)
}

type ADUserUsecase interface {
	Ingest(ctx context.Context, req dto.IngestRequest) (int, error)
	List(ctx context.Context, f dto.ADUserFilter) (*database.List[dto.UserRow], error)
	Each(ctx context.Context, source string, fn func(domain.ADUser) error) error
	Stats(ctx context.Context) (*dto.ADUserStats, error)
	ResolveLinuxIdentity(ctx context.Context, req dto.ResolveLinuxIdentityRequest) (int64, error)
}

type IdentityRepository interface {
	Create(ctx context.Context, name, actor string) (*domain.IdentityAlias, error)
	List(ctx context.Context, search string) ([]dto.IdentityListItem, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.IdentityAlias, error)
	Users(ctx context.Context, id uuid.UUID) ([]dto.UserRow, error)
	Delete(ctx context.Context, id uuid.UUID) (int, error)
	Assign(ctx context.Context, userID, aliasID uuid.UUID, actor string) error
	Unassign(ctx context.Context, userID uuid.UUID, actor string) error
}

type IdentityUsecase interface {
	Create(ctx context.Context, name, actor string) (*dto.IdentityResponse, error)
	List(ctx context.Context, search string) ([]dto.IdentityListItem, error)
	Detail(ctx context.Context, id uuid.UUID) (*dto.IdentityDetail, error)
	Delete(ctx context.Context, id uuid.UUID) (int, error)
	Assign(ctx context.Context, userID, aliasID uuid.UUID, actor string) error
	Unassign(ctx context.Context, userID uuid.UUID, actor string) error
}
