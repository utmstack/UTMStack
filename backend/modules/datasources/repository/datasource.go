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
	"asset_name", "data_type", "asset_ip", "source_kind", "labels", "group_id",
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

func (r *pgDatasourceRepository) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error) {
	return r.GetAll(ctx, req, datasourceFilterFields, "id DESC", database.Preload("Group"))
}

func (r *pgDatasourceRepository) Count(ctx context.Context) (int64, error) {
	return r.db.Count(ctx, new(domain.Datasource))
}

// dedupBySourceRef keeps the last entry per (tenant, source_ref), which is what
// the unique index is: source_ref is "<dataType>:<dataSource>" and carries no
// tenant, so two tenants running an agent with the same hostname produce the
// same one.
func dedupBySourceRef(items []domain.Datasource) []domain.Datasource {
	type key struct{ tenant, ref string }
	seen := make(map[key]int, len(items))
	out := make([]domain.Datasource, 0, len(items))
	for _, item := range items {
		k := key{item.TenantID, item.SourceRef}
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
	items = dedupBySourceRef(items)
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tenant_id"}, {Name: "source_ref"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"asset_name", "data_type", "source_kind", "asset_ip", "metadata", "last_ping_at", "modified_at",
			}),
		}).
		Create(&items).Error
}

func (r *pgDatasourceRepository) RegisterBatch(ctx context.Context, items []domain.Datasource) error {
	items = dedupBySourceRef(items)
	if len(items) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tenant_id"}, {Name: "source_ref"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"asset_name", "data_type", "source_kind", "asset_ip", "metadata", "modified_at",
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
			Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "source_ref"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_ping_at", "modified_at"}),
			Where: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("datasources.source_kind <> 'agent'"),
			}},
		}).
		Create(&items).Error
}

// EnrichmentRows returns the datasources that actually carry enrichment (a group
// or labels), with their Group preloaded — the feed the alerts plugin caches.
func (r *pgDatasourceRepository) EnrichmentRows(ctx context.Context) ([]domain.Datasource, error) {
	var rows []domain.Datasource
	err := r.db.GORM().WithContext(ctx).
		Preload("Group").
		Where("group_id IS NOT NULL OR (labels IS NOT NULL AND labels <> '')").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
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

func (r *pgDatasourceRepository) ClearGroup(ctx context.Context, groupID uint64) error {
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("group_id = ?", groupID).
		Update("group_id", nil).Error
}

func (r *pgDatasourceRepository) UpdateLabels(ctx context.Context, id uint64, labels string) error {
	return r.db.GORM().WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("id = ?", id).
		Update("labels", labels).Error
}

func (r *pgDatasourceRepository) UpdateSensitivity(ctx context.Context, id uint64, conf, integ, avail int) error {
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

func (r *pgDatasourceRepository) Delete(ctx context.Context, id uint64) error {
	return r.Remove(ctx, id)
}
