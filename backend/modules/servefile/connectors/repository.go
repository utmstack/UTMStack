package connectors

import (
	"context"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/servefile/domain"
)

type ServedFileRepository interface {
	Save(ctx context.Context, f *domain.ServedFile) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ServedFile, error)
}
