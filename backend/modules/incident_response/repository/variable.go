package repository

import (
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	"github.com/utmstack/utmstack/backend/modules/incident_response/dto"
	irerrors "github.com/utmstack/utmstack/backend/modules/incident_response/errors"
	"gorm.io/gorm"
)

type variableRepository struct {
	db *gorm.DB
}

func NewVariableRepository(db *gorm.DB) *variableRepository {
	return &variableRepository{db: db}
}

func (r *variableRepository) Save(v *domain.UtmIncidentVariable) error {
	if err := r.db.Save(v).Error; err != nil {
		return fmt.Errorf("variableRepository.Save: %w", err)
	}
	return nil
}

func (r *variableRepository) FindByID(id int64) (*domain.UtmIncidentVariable, error) {
	var v domain.UtmIncidentVariable
	err := r.db.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, irerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("variableRepository.FindByID: %w", err)
	}
	return &v, nil
}

func (r *variableRepository) FindAll(f dto.VariableFilter) ([]domain.UtmIncidentVariable, int64, error) {
	q := r.db.Model(&domain.UtmIncidentVariable{})
	if f.VariableName != nil {
		q = q.Where("variable_name ILIKE ?", "%"+*f.VariableName+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("variableRepository.FindAll count: %w", err)
	}

	page, size := pageSize(f.Page, f.Size)
	var items []domain.UtmIncidentVariable
	if err := q.Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("variableRepository.FindAll: %w", err)
	}
	return items, total, nil
}

func (r *variableRepository) FindAllPlain() ([]domain.UtmIncidentVariable, error) {
	var items []domain.UtmIncidentVariable
	if err := r.db.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("variableRepository.FindAllPlain: %w", err)
	}
	return items, nil
}

func (r *variableRepository) FindByName(name string) (*domain.UtmIncidentVariable, error) {
	var v domain.UtmIncidentVariable
	err := r.db.Where("variable_name = ?", name).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, irerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("variableRepository.FindByName: %w", err)
	}
	return &v, nil
}

func (r *variableRepository) FindByNames(names []string) ([]domain.UtmIncidentVariable, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var items []domain.UtmIncidentVariable
	if err := r.db.Where("variable_name IN ?", names).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("variableRepository.FindByNames: %w", err)
	}
	return items, nil
}

func (r *variableRepository) Delete(id int64) error {
	if err := r.db.Delete(&domain.UtmIncidentVariable{}, id).Error; err != nil {
		return fmt.Errorf("variableRepository.Delete: %w", err)
	}
	return nil
}
