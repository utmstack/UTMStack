package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type pgScheduleRepo struct{ db *gorm.DB }

func NewScheduleRepository(db *gorm.DB) connectors.ScheduleRepository {
	return &pgScheduleRepo{db: db}
}

func (r *pgScheduleRepo) Create(ctx context.Context, s *domain.ReportSchedule) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *pgScheduleRepo) Update(ctx context.Context, s *domain.ReportSchedule) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *pgScheduleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportSchedule, error) {
	var s domain.ReportSchedule
	err := scopeTenant(ctx, r.db.WithContext(ctx)).First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrScheduleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgScheduleRepo) ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]domain.ReportSchedule, int64, error) {
	q := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.ReportSchedule{})).
		Where("user_id = ?", userID)
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
	var items []domain.ReportSchedule
	if err := q.Order("framework_key ASC").Offset(f.Page * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgScheduleRepo) ListDue(ctx context.Context, now time.Time) ([]domain.ReportSchedule, error) {
	var items []domain.ReportSchedule
	return items, r.db.WithContext(ctx).
		Where("next_execution_date <= ?", now.UTC()).
		Order("next_execution_date ASC").
		Find(&items).Error
}

func (r *pgScheduleRepo) ForFramework(ctx context.Context, frameworkKey string) (*domain.ReportSchedule, error) {
	var s domain.ReportSchedule
	err := scopeTenant(ctx, r.db.WithContext(ctx)).
		Order("id ASC").
		First(&s, "framework_key = ?", frameworkKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res := scopeTenant(ctx, r.db.WithContext(ctx)).Delete(&domain.ReportSchedule{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrScheduleNotFound
	}
	return nil
}

func (r *pgScheduleRepo) ClaimDue(ctx context.Context, id uuid.UUID, expectedNext, newLast, newNext time.Time) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.ReportSchedule{}).
		Where("id = ? AND next_execution_date = ?", id, expectedNext.UTC()).
		Updates(map[string]any{
			"last_execution_date": newLast.UTC(),
			"next_execution_date": newNext.UTC(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
