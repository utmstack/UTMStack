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

// applyStatus narrows a query to a derived lifecycle bucket. These derivations
// (active vs disabled vs deleted vs stale) live here because the raw record only
// stores `active` + the timestamps — the buckets the UI filters by are computed.
func applyStatus(q *gorm.DB, status string) *gorm.DB {
	switch status {
	case "active":
		return q.Where("active = true AND account_deleted_at IS NULL")
	case "disabled":
		return q.Where("active = false AND account_deleted_at IS NULL")
	case "deleted":
		return q.Where("account_deleted_at IS NOT NULL")
	case "stale":
		return q.Where("active = true AND account_deleted_at IS NULL AND last_logon IS NOT NULL AND last_logon < NOW() - INTERVAL '30 days'")
	case "service":
		return q.Where("account_deleted_at IS NULL AND sam_account_name ILIKE 'svc%'")
	}
	return q
}

func (r *pgADUserRepository) List(ctx context.Context, f dto.ADUserFilter) ([]domain.ADUser, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.ADUser{})
	if f.TenantID != "" {
		q = q.Where("tenant_id = ?", f.TenantID)
	}
	if f.Status != "" {
		// Status is a richer, derived filter; when present it supersedes the raw
		// `active` flag so the two can't contradict each other.
		q = applyStatus(q, f.Status)
	} else if f.Active != nil {
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

	order := "sam_account_name ASC"
	if f.Sort == "recent" {
		order = "last_seen DESC NULLS LAST"
	}
	var items []domain.ADUser
	if err := q.Order(order).Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgADUserRepository) Stats(ctx context.Context, tenantID string) (*dto.ADUserStats, error) {
	// Each count is its own scoped query so the tenant filter applies uniformly.
	base := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&domain.ADUser{})
		if tenantID != "" {
			q = q.Where("tenant_id = ?", tenantID)
		}
		return q
	}

	s := dto.ADUserStats{ByDomain: []dto.DomainCount{}, Tenants: []string{}}
	if err := base().Count(&s.Total).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "active").Count(&s.Active).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "disabled").Count(&s.Disabled).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "deleted").Count(&s.Deleted).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "stale").Count(&s.Stale).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "service").Count(&s.Service).Error; err != nil {
		return nil, err
	}
	if err := base().Where("last_seen IS NOT NULL AND last_seen > NOW() - INTERVAL '24 hours'").Count(&s.Seen24h).Error; err != nil {
		return nil, err
	}
	if err := base().Select("domain, COUNT(*) AS count").Group("domain").Order("count DESC").Limit(8).Scan(&s.ByDomain).Error; err != nil {
		return nil, err
	}
	// Tenants is intentionally global (no tenant scope) so the picker is stable.
	if err := r.db.WithContext(ctx).Model(&domain.ADUser{}).Distinct().Order("tenant_id ASC").Pluck("tenant_id", &s.Tenants).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgADUserRepository) All(ctx context.Context) ([]domain.ADUser, error) {
	var items []domain.ADUser
	if err := r.db.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
