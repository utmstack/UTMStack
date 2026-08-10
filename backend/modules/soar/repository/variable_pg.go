package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

var _ connectors.VariableRepository = (*variableRepository)(nil)

type variableRepository struct {
	db *gorm.DB
}

func NewVariableRepository(db *gorm.DB) *variableRepository {
	return &variableRepository{db: db}
}

func (r *variableRepository) Save(ctx context.Context, v *domain.SoarVariable) error {
	if err := r.db.WithContext(ctx).Save(v).Error; err != nil {
		return fmt.Errorf("variableRepository.Save: %w", err)
	}
	return nil
}

func (r *variableRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.SoarVariable, error) {
	var v domain.SoarVariable
	err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrVariableNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("variableRepository.FindByID: %w", err)
	}
	return &v, nil
}

func (r *variableRepository) FindAll(ctx context.Context, f dto.VariableFilter) ([]domain.SoarVariable, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.SoarVariable{})
	if f.Name != nil {
		q = q.Where("variable_name ILIKE ?", "%"+*f.Name+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("variableRepository.FindAll count: %w", err)
	}

	var items []domain.SoarVariable
	if err := q.Order("variable_name").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("variableRepository.FindAll: %w", err)
	}
	return items, total, nil
}

func (r *variableRepository) FindAllPlain(ctx context.Context) ([]domain.SoarVariable, error) {
	var items []domain.SoarVariable
	if err := r.db.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("variableRepository.FindAllPlain: %w", err)
	}
	return items, nil
}

func (r *variableRepository) FindByName(ctx context.Context, name string) (*domain.SoarVariable, error) {
	var v domain.SoarVariable
	err := r.db.WithContext(ctx).Where("variable_name = ?", name).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrVariableNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("variableRepository.FindByName: %w", err)
	}
	return &v, nil
}

func (r *variableRepository) FindByNames(ctx context.Context, names []string) ([]domain.SoarVariable, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var items []domain.SoarVariable
	if err := r.db.WithContext(ctx).Where("variable_name IN ?", names).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("variableRepository.FindByNames: %w", err)
	}
	return items, nil
}

func (r *variableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.SoarVariable{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("variableRepository.Delete: %w", err)
	}
	return nil
}
