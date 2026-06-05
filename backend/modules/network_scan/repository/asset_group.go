package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"gorm.io/gorm"
)

type pgAssetGroupRepository struct {
	db *gorm.DB
}

func NewAssetGroupRepository(db *gorm.DB) connectors.AssetGroupRepository {
	return &pgAssetGroupRepository{db: db}
}

func (r *pgAssetGroupRepository) Save(ctx context.Context, g *domain.UtmAssetGroup) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *pgAssetGroupRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error) {
	var g domain.UtmAssetGroup
	if err := r.db.WithContext(ctx).First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (r *pgAssetGroupRepository) List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetGroup, int64, error) {
	page, size := normalizePage(p)
	q := r.db.WithContext(ctx).Model(&domain.UtmAssetGroup{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.UtmAssetGroup
	if err := q.Order("created_date DESC NULLS LAST, id DESC").
		Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgAssetGroupRepository) SearchByFilter(
	ctx context.Context,
	f domain.AssetGroupFilter,
	p domain.Pagination,
) ([]connectors.AssetGroupRow, int64, error) {
	page, size := normalizePage(p)
	whereParts := []string{"1=1"}
	args := []any{}

	if f.ID != nil {
		whereParts = append(whereParts, "g.id = ?")
		args = append(args, *f.ID)
	}
	if strings.TrimSpace(f.GroupName) != "" {
		whereParts = append(whereParts, "g.group_name ILIKE ?")
		args = append(args, "%"+f.GroupName+"%")
	}
	if f.InitDate != nil {
		whereParts = append(whereParts, "g.created_date >= ?")
		args = append(args, *f.InitDate)
	}
	if f.EndDate != nil {
		whereParts = append(whereParts, "g.created_date <= ?")
		args = append(args, *f.EndDate)
	}
	whereSQL := strings.Join(whereParts, " AND ")

	var total int64
	if err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM utm_asset_group g WHERE "+whereSQL, args...).
		Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ID               uint64 `gorm:"column:id"`
		GroupName        string `gorm:"column:group_name"`
		GroupDescription string `gorm:"column:group_description"`
		CreatedDate      *string `gorm:"column:created_date"` // scanned as text to avoid driver fussiness; remapped below
		AssetsCount      int64  `gorm:"column:assets_count"`
	}
	var rows []row
	query := `SELECT g.id, g.group_name, g.group_description,
                     to_char(g.created_date, 'YYYY-MM-DD"T"HH24:MI:SSOF') AS created_date,
                     (SELECT COUNT(*) FROM utm_network_scan n WHERE n.group_id = g.id) AS assets_count
              FROM utm_asset_group g
              WHERE ` + whereSQL + `
              ORDER BY g.created_date DESC NULLS LAST, g.id DESC
              OFFSET ? LIMIT ?`
	args2 := append(args, (page-1)*size, size)
	if err := r.db.WithContext(ctx).Raw(query, args2...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]connectors.AssetGroupRow, 0, len(rows))
	for _, x := range rows {
		ag := connectors.AssetGroupRow{
			Group: domain.UtmAssetGroup{
				ID:               x.ID,
				GroupName:        x.GroupName,
				GroupDescription: x.GroupDescription,
			},
			AssetsCount: x.AssetsCount,
		}
		if x.CreatedDate != nil {
			// Postgres-side formatting; parse loosely with a couple of layouts.
			ag.Group.CreatedDate = parseTimestamp(*x.CreatedDate)
		}
		out = append(out, ag)
	}
	return out, total, nil
}

func (r *pgAssetGroupRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmAssetGroup{}, id).Error
}
