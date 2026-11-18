package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/edr/connectors"
	"github.com/utmstack/utmstack/backend/modules/edr/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) connectors.Repository {
	return &pgRepository{db: db}
}

var ErrPolicyNotFound = errors.New("edr policy not found")

func (r *pgRepository) Create(ctx context.Context, p *domain.Policy) error {
	p.TenantID = authz.TenantIDFromContext(ctx)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *pgRepository) Update(ctx context.Context, p *domain.Policy) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *pgRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Policy{}, "id = ?", id).Error
}

func (r *pgRepository) Get(ctx context.Context, id uint) (*domain.Policy, error) {
	var p domain.Policy
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgRepository) List(ctx context.Context) ([]domain.Policy, error) {
	var ps []domain.Policy
	if err := r.db.WithContext(ctx).Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}
