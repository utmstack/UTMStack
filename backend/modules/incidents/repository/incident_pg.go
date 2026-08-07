package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
	"gorm.io/gorm"
)

func tenantFromCtx(ctx context.Context) uuid.UUID {
	tid, _ := uuid.Parse(authz.TenantIDFromContext(ctx))
	return tid
}

func scopeTenant(ctx context.Context, q *gorm.DB) *gorm.DB {
	tid := tenantFromCtx(ctx)
	if tid != uuid.Nil {
		return q.Where("tenant_id = ?", tid)
	}
	// Fail closed. The gorm tenancy plugin already aborts an unscoped query on a
	// multi-tenant install, but this must not be the thing that reads every
	// tenant's incidents if that plugin is ever unregistered or bypassed. On a
	// single-tenant install the plugin fills in the default tenant instead.
	if tenancy.Enabled() {
		q.AddError(ErrNoTenant)
	}
	return q
}

// ErrNoTenant is returned when a query would otherwise run unscoped.
var ErrNoTenant = errors.New("incidents: no tenant in scope")

type pgIncidentRepository struct {
	db *gorm.DB
}

func NewIncidentRepository(db *gorm.DB) connectors.IncidentRepository {
	return &pgIncidentRepository{db: db}
}

// alertInsertBatch bounds how many rows go in one INSERT. Linking a selection
// of alerts used to be one round trip per alert; a few hundred of those is the
// difference between a click and a wait.
const alertInsertBatch = 200

func (r *pgIncidentRepository) Create(ctx context.Context, incident *domain.Incident, alerts []domain.IncidentAlert) error {
	tenant := tenantFromCtx(ctx)
	if incident.TenantID == uuid.Nil {
		incident.TenantID = tenant
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(incident).Error; err != nil {
			return err
		}
		return insertAlerts(tx, incident, alerts, tenant)
	})
}

func (r *pgIncidentRepository) LinkAlerts(ctx context.Context, incident *domain.Incident, alerts []domain.IncidentAlert) error {
	tenant := tenantFromCtx(ctx)
	return scopeTenant(ctx, r.db.WithContext(ctx)).Transaction(func(tx *gorm.DB) error {
		if err := insertAlerts(tx, incident, alerts, tenant); err != nil {
			return err
		}
		return tx.Save(incident).Error
	})
}

// insertAlerts stamps the parent and tenant on each row and writes them in
// batches. The ids are left to the database default rather than generated here.
func insertAlerts(tx *gorm.DB, incident *domain.Incident, alerts []domain.IncidentAlert, tenant uuid.UUID) error {
	if len(alerts) == 0 {
		return nil
	}
	for i := range alerts {
		alerts[i].IncidentID = incident.ID
		if alerts[i].TenantID == uuid.Nil {
			alerts[i].TenantID = tenant
		}
	}
	return tx.CreateInBatches(alerts, alertInsertBatch).Error
}

func (r *pgIncidentRepository) Update(ctx context.Context, incident *domain.Incident) error {
	return scopeTenant(ctx, r.db.WithContext(ctx)).Save(incident).Error
}

func (r *pgIncidentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	var row domain.Incident
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgIncidentRepository) FindAll(ctx context.Context, q dto.IncidentListQuery) ([]domain.Incident, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.Incident{}))

	if q.IncidentName != nil && *q.IncidentName != "" {
		db = db.Where("incident_name ILIKE ?", "%"+*q.IncidentName+"%")
	}
	if q.IncidentStatus != nil && *q.IncidentStatus != "" {
		db = db.Where("incident_status = ?", *q.IncidentStatus)
	}
	if q.IncidentAssignedTo != nil && *q.IncidentAssignedTo != "" {
		db = db.Where("incident_assigned_to ILIKE ?", "%"+*q.IncidentAssignedTo+"%")
	}
	if q.CreatedDateStart != nil {
		db = db.Where("incident_created_date >= ?", *q.CreatedDateStart)
	}
	if q.CreatedDateEnd != nil {
		db = db.Where("incident_created_date <= ?", *q.CreatedDateEnd)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.Incident
	if err := db.Order(orderBy(q.Sort, incidentSortable, "incident_created_date DESC")).
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if err := r.fillAlertCounts(ctx, rows); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// fillAlertCounts counts the alerts of a whole page in one grouped query rather
// than one per row.
func (r *pgIncidentRepository) fillAlertCounts(ctx context.Context, rows []domain.Incident) error {
	if len(rows) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}

	var counts []struct {
		IncidentID uuid.UUID
		N          int
	}
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentAlert{})).
		Select("incident_id, count(*) AS n").
		Where("incident_id IN ?", ids).
		Group("incident_id").
		Scan(&counts).Error
	if err != nil {
		return err
	}

	byID := make(map[uuid.UUID]int, len(counts))
	for _, c := range counts {
		byID[c.IncidentID] = c.N
	}
	for i := range rows {
		rows[i].AlertCount = byID[rows[i].ID]
	}
	return nil
}

// DistinctAssignees reads the column rather than walking every incident: the
// old version paged through the first thousand and split them on commas, so an
// assignee whose incidents had aged past that simply vanished from the filter.
func (r *pgIncidentRepository) DistinctAssignees(ctx context.Context) ([]string, error) {
	var out []string
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.Incident{})).
		Where("incident_assigned_to <> ''").
		Distinct().
		Order("incident_assigned_to ASC").
		Pluck("incident_assigned_to", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}
