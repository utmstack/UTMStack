package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
	"gorm.io/gorm"
)

type pgScheduleRepo struct{ db *gorm.DB }

func NewScheduleRepository(db *gorm.DB) connectors.ScheduleRepository {
	return &pgScheduleRepo{db: db}
}

func (r *pgScheduleRepo) Create(ctx context.Context, s *domain.UtmComplianceReportSchedule) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *pgScheduleRepo) Update(ctx context.Context, s *domain.UtmComplianceReportSchedule) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *pgScheduleRepo) GetByID(ctx context.Context, id int64) (*domain.UtmComplianceReportSchedule, error) {
	var s domain.UtmComplianceReportSchedule
	err := r.db.WithContext(ctx).First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *pgScheduleRepo) ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]domain.UtmComplianceReportSchedule, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmComplianceReportSchedule{}).Where("user_id = ?", userID)
	if f.FrameworkKey != "" {
		q = q.Where("framework_key = ?", f.FrameworkKey)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	size := f.Size
	if size <= 0 {
		size = 20
	}
	var items []domain.UtmComplianceReportSchedule
	if err := q.Offset(f.Page * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgScheduleRepo) ListAll(ctx context.Context) ([]domain.UtmComplianceReportSchedule, error) {
	var items []domain.UtmComplianceReportSchedule
	return items, r.db.WithContext(ctx).Find(&items).Error
}

func (r *pgScheduleRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmComplianceReportSchedule{}, id).Error
}

func (r *pgScheduleRepo) ClaimDue(ctx context.Context, id int64, expectedLast, newLast time.Time) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.UtmComplianceReportSchedule{}).
		Where("id = ? AND last_execution_date = ?", id, expectedLast).
		Update("last_execution_date", newLast)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
