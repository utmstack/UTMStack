package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type pgReportStore struct{ db *gorm.DB }

// NewReportStore persists compliance report snapshots in postgres, one row per
// generated report. Rows are tenant-scoped end-to-end: writes stamp the acting
// tenant, reads narrow by it. An empty ctx tenant (on-prem/global) is
// unrestricted, matching the rest of the module.
func NewReportStore(db *gorm.DB) connectors.ReportStore { return &pgReportStore{db: db} }

func (r *pgReportStore) Save(ctx context.Context, snap *domain.ReportSnapshot) error {
	body, err := json.Marshal(snap.Report)
	if err != nil {
		return err
	}
	ts, err := parseSnapshotTime(snap.Timestamp)
	if err != nil {
		return err
	}
	tid := snap.TenantID
	if tid == "" {
		tid = authz.TenantIDFromContext(ctx)
	}
	row := domain.UtmComplianceReportSnapshot{
		ID:            snap.ID,
		TenantID:      tid,
		FrameworkKey:  snap.FrameworkKey,
		FrameworkName: snap.FrameworkName,
		Timestamp:     ts,
		Score:         snap.Score,
		Report:        datatypes.JSON(body),
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *pgReportStore) List(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error) {
	if limit <= 0 {
		limit = 50
	}
	q := scopeSnapshotTenant(ctx, r.db.WithContext(ctx).Model(&domain.UtmComplianceReportSnapshot{}))
	if frameworkKey != "" {
		q = q.Where("framework_key = ?", frameworkKey)
	}
	var rows []domain.UtmComplianceReportSnapshot
	if err := q.
		Select("id", "framework_key", "framework_name", "generated_at", "score").
		Order("generated_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ReportSnapshotMeta, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.ReportSnapshotMeta{
			ID:            r.ID,
			FrameworkKey:  r.FrameworkKey,
			FrameworkName: r.FrameworkName,
			Timestamp:     r.Timestamp.UTC().Format(time.RFC3339),
			Score:         r.Score,
		})
	}
	return out, nil
}

func (r *pgReportStore) Get(ctx context.Context, id string) (*domain.ReportSnapshot, error) {
	var row domain.UtmComplianceReportSnapshot
	if err := scopeSnapshotTenant(ctx, r.db.WithContext(ctx)).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrReportNotFound
		}
		return nil, err
	}
	var report domain.Report
	if err := json.Unmarshal(row.Report, &report); err != nil {
		return nil, err
	}
	return &domain.ReportSnapshot{
		ID:            row.ID,
		TenantID:      row.TenantID,
		FrameworkKey:  row.FrameworkKey,
		FrameworkName: row.FrameworkName,
		Timestamp:     row.Timestamp.UTC().Format(time.RFC3339),
		Score:         row.Score,
		Report:        report,
	}, nil
}

func (r *pgReportStore) Delete(ctx context.Context, id string) error {
	res := scopeSnapshotTenant(ctx, r.db.WithContext(ctx)).Delete(&domain.UtmComplianceReportSnapshot{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrReportNotFound
	}
	return nil
}

func scopeSnapshotTenant(ctx context.Context, q *gorm.DB) *gorm.DB {
	if tid := authz.TenantIDFromContext(ctx); tid != "" {
		return q.Where("tenant_id = ?", tid)
	}
	return q
}

// parseSnapshotTime accepts either RFC3339 or an empty string (defaults to now).
// The evaluator formats snapshot.Timestamp with time.RFC3339 today, but a
// blank comes from callers that stamp only ID and framework — don't reject
// those, they're the reason snapshots exist.
func parseSnapshotTime(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}
