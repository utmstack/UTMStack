package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgADUserRepository struct{ db *gorm.DB }

func NewADUserRepository(db *gorm.DB) connectors.ADUserRepository {
	return &pgADUserRepository{db: db}
}

func (r *pgADUserRepository) Upsert(ctx context.Context, users []domain.ADUser) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "sid"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"sam_account_name", "domain", "active",
			"account_created_at", "last_logon", "account_deleted_at", "last_seen",
		}),
	}).Create(&users).Error
}

func (r *pgADUserRepository) List(ctx context.Context, f dto.ADUserFilter) ([]domain.ADUser, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.ADUser{})
	if f.TenantID != "" {
		q = q.Where("tenant_id = ?", f.TenantID)
	}
	if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("sam_account_name ILIKE ? OR sid ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.ADUser
	if err := q.Order("sam_account_name ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgADUserRepository) All(ctx context.Context) ([]domain.ADUser, error) {
	var items []domain.ADUser
	if err := r.db.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
