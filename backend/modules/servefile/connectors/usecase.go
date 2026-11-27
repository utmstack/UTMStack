package connectors

import (
	"context"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/servefile/domain"
)

type ServedFileUsecase interface {
	// Store encrypts data to disk under the acting tenant and registers the
	// row; it returns the row with TenantID stamped.
	Store(ctx context.Context, filename, contentType string, data []byte) (*domain.ServedFile, error)
	// Fetch returns the decrypted bytes and the row for fileID, verifying the
	// acting tenant (tenantID) owns it.
	Fetch(ctx context.Context, fileID uuid.UUID, tenantID string) ([]byte, *domain.ServedFile, error)
}
