package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"gorm.io/gorm"
)

type pgRepo struct {
	db *database.DB
}

func NewRepository(db *gorm.DB) connectors.Repository {
	return &pgRepo{db: database.New(db)}
}

func (r *pgRepo) List(ctx context.Context) ([]domain.Config, error) {
	var items []domain.Config
	if err := r.db.FindAll(ctx, &items, database.OrderBy("conf_param_short ASC")); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgRepo) GetByKey(ctx context.Context, key string) (*domain.Config, error) {
	var c domain.Config
	err := r.db.FindOne(ctx, &c, database.Where("conf_param_short = ?", key))
	if errors.Is(err, database.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgRepo) Update(ctx context.Context, c *domain.Config) error {
	return r.db.Update(ctx, c)
}
