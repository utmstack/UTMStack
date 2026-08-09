package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

var datasourceFilterFields = []string{
	"asset_name", "data_type", "asset_ip", "source_kind", "labels",
	"discovered_at", "modified_at", "last_ping_at",
}

var identityColumns = []clause.Column{
	{Name: "tenant_id"}, {Name: "data_type"}, {Name: "asset_name"},
}

type pgDatasourceRepository struct {
	database.AbstractRepository[domain.Datasource, uuid.UUID]
	db *database.DB
}

func NewDatasourceRepository(gdb *gorm.DB) connectors.DatasourceRepository {
	provider := database.New(gdb)
	return &pgDatasourceRepository{
		AbstractRepository: database.NewAbstractRepository[domain.Datasource, uuid.UUID](provider),
		db:                 provider,
	}
}

func (r *pgDatasourceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Datasource, error) {
	var d domain.Datasource
	err := r.db.FindOne(ctx, &d, database.Where("id = ?", id))
	if errors.Is(err, database.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *pgDatasourceRepository) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error) {
	return r.GetAll(ctx, req, datasourceFilterFields, "discovered_at DESC NULLS LAST, id DESC")
}

func (r *pgDatasourceRepository) Count(ctx context.Context) (int64, error) {
	return r.db.Count(ctx, new(domain.Datasource))
}

func dedupByIdentity(items []domain.Datasource) []domain.Datasource {
	type key struct {
		tenant   uuid.UUID
		dataType string
		name     string
	}
	seen := make(map[key]int, len(items))
	out := make([]domain.Datasource, 0, len(items))
	for _, item := range items {
		k := key{item.TenantID, item.DataType, item.Name}
		if idx, ok := seen[k]; ok {
			out[idx] = item
		} else {
			seen[k] = len(out)
			out = append(out, item)
		}
	}
	return out
}

func (r *pgDatasourceRepository) UpsertBatch(ctx context.Context, items []domain.Datasource) error {
	items = dedupByIdentity(items)
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: identityColumns,
			DoUpdates: clause.AssignmentColumns([]string{
				"source_kind", "asset_ip", "metadata", "last_ping_at", "modified_at",
			}),
		}).
		Create(&items).Error
}

func (r *pgDatasourceRepository) RegisterBatch(ctx context.Context, items []domain.Datasource) error {
	items = dedupByIdentity(items)
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: identityColumns,
			DoUpdates: clause.AssignmentColumns([]string{
				"source_kind", "asset_ip", "metadata", "modified_at",
			}),
		}).
		Create(&items).Error
}

func (r *pgDatasourceRepository) UpsertLivenessBatch(ctx context.Context, items []domain.Datasource) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   identityColumns,
			DoUpdates: clause.AssignmentColumns([]string{"last_ping_at", "modified_at"}),
			Where: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("datasources.source_kind <> ?", domain.SourceKindAgent),
			}},
		}).
		Create(&items).Error
}

func (r *pgDatasourceRepository) UpdateLabels(ctx context.Context, id uuid.UUID, labels string) error {
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("id = ?", id).
		Update("labels", labels).Error
}

func (r *pgDatasourceRepository) UpdateSensitivity(ctx context.Context, id uuid.UUID, conf, integ, avail int) error {
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"asset_confidentiality": conf,
			"asset_integrity":       integ,
			"asset_availability":    avail,
		}).Error
}

func (r *pgDatasourceRepository) ListSensitive(ctx context.Context) ([]domain.Datasource, error) {
	var rows []domain.Datasource
	err := r.db.GORM().WithContext(ctx).
		Where("asset_confidentiality > 0 OR asset_integrity > 0 OR asset_availability > 0").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgDatasourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.Remove(ctx, id)
}
