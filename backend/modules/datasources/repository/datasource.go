package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var datasourceFilterFields = []string{
	"asset_name", "asset_ip", "source_kind", "labels", "group_id",
	"discovered_at", "modified_at", "last_ping_at",
}

type pgDatasourceRepository struct {
	database.AbstractRepository[domain.Datasource, uint64]
	db *database.DB
}

func NewDatasourceRepository(gdb *gorm.DB) connectors.DatasourceRepository {
	provider := database.New(gdb)
	return &pgDatasourceRepository{
		AbstractRepository: database.NewAbstractRepository[domain.Datasource, uint64](provider),
		db:                 provider,
	}
}

func (r *pgDatasourceRepository) FindByID(ctx context.Context, id uint64) (*domain.Datasource, error) {
	var d domain.Datasource
	err := r.db.FindOne(ctx, &d, database.Where("id = ?", id), database.Preload("Group"))
	if errors.Is(err, database.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *pgDatasourceRepository) FindByName(ctx context.Context, name string) (*domain.Datasource, error) {
	var d domain.Datasource
	err := r.db.FindOne(ctx, &d, database.Where("asset_name = ?", name), database.Preload("Group"))
	if errors.Is(err, database.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *pgDatasourceRepository) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error) {
	return r.GetAll(ctx, req, datasourceFilterFields, "id DESC", database.Preload("Group"))
}

func (r *pgDatasourceRepository) UpsertBatch(ctx context.Context, items []domain.Datasource) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "asset_name"}},
			// Update only ping-owned fields; group_id, labels and discovered_at are
			// preserved on existing rows. A fresh last_ping_at is what revives a source.
			DoUpdates: clause.AssignmentColumns([]string{"source_kind", "asset_ip", "metadata", "last_ping_at", "modified_at"}),
		}).
		Create(&items).Error
}

func (r *pgDatasourceRepository) UpdateGroup(ctx context.Context, ids []uint64, groupID *uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("id IN ?", ids).
		Update("group_id", groupID).Error
}

func (r *pgDatasourceRepository) UpdateLabels(ctx context.Context, id uint64, labels string) error {
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("id = ?", id).
		Update("labels", labels).Error
}

func (r *pgDatasourceRepository) Delete(ctx context.Context, id uint64) error {
	return r.Remove(ctx, id)
}
