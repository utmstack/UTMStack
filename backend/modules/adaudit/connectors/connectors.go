package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ADUserRepository interface {
	Upsert(ctx context.Context, users []domain.ADUser) error
	List(ctx context.Context, f dto.ADUserFilter) ([]domain.ADUser, int64, error)
	All(ctx context.Context) ([]domain.ADUser, error)
}

type ADUserUsecase interface {
	Ingest(ctx context.Context, req dto.IngestRequest) (int, error)
	List(ctx context.Context, f dto.ADUserFilter) (*database.List[domain.ADUser], error)
	All(ctx context.Context) ([]domain.ADUser, error)
}
