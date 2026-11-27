package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/servefile/connectors"
	"github.com/utmstack/utmstack/backend/modules/servefile/domain"
	"gorm.io/gorm"
)

type pgServedFileRepository struct{ db *gorm.DB }

func NewServedFileRepository(db *gorm.DB) connectors.ServedFileRepository {
	return &pgServedFileRepository{db: db}
}

func (r *pgServedFileRepository) Save(ctx context.Context, f *domain.ServedFile) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *pgServedFileRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ServedFile, error) {
	var f domain.ServedFile
	err := r.db.WithContext(ctx).First(&f, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}
