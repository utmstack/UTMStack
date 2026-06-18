package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"gorm.io/gorm"
)

var assetGroupFilterFields = []string{"group_name", "group_description", "created_date"}

type pgAssetGroupRepository struct {
	database.AbstractRepository[domain.UtmAssetGroup, uint64]
	db *database.DB
}

func NewAssetGroupRepository(gdb *gorm.DB) connectors.AssetGroupRepository {
	provider := database.New(gdb)
	return &pgAssetGroupRepository{
		AbstractRepository: database.NewAbstractRepository[domain.UtmAssetGroup, uint64](provider),
		db:                 provider,
	}
}

func (r *pgAssetGroupRepository) Save(ctx context.Context, g *domain.UtmAssetGroup) error {
	return r.db.GORM().WithContext(ctx).Save(g).Error
}

func (r *pgAssetGroupRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error) {
	return r.GetByID(ctx, id)
}

func (r *pgAssetGroupRepository) Delete(ctx context.Context, id uint64) error {
	return r.Remove(ctx, id)
}

func (r *pgAssetGroupRepository) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[connectors.AssetGroupRow], error) {
	groups, err := r.GetAll(ctx, req, assetGroupFilterFields, "created_date DESC NULLS LAST, id DESC")
	if err != nil {
		return common_models.ListResponse[connectors.AssetGroupRow]{}, err
	}

	items := groups.GetItems()
	ids := make([]uint64, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}

	counts := make(map[uint64]int64, len(ids))
	if len(ids) > 0 {
		type row struct {
			GroupID uint64
			N       int64
		}
		var rows []row
		if err := r.db.GORM().WithContext(ctx).
			Model(&domain.Datasource{}).
			Select("group_id, count(*) AS n").
			Where("group_id IN ?", ids).
			Group("group_id").
			Scan(&rows).Error; err != nil {
			return common_models.ListResponse[connectors.AssetGroupRow]{}, err
		}
		for _, x := range rows {
			counts[x.GroupID] = x.N
		}
	}

	return common_models.MapList(groups, func(g domain.UtmAssetGroup) connectors.AssetGroupRow {
		return connectors.AssetGroupRow{Group: g, DatasourcesCount: counts[g.ID]}
	}), nil
}
