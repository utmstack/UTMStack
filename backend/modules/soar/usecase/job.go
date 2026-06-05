package usecase

import (
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type jobUsecase struct {
	repo connectors.JobRepository
}

func NewJobUsecase(repo connectors.JobRepository) connectors.JobUsecase {
	return &jobUsecase{repo: repo}
}

func (u *jobUsecase) Create(req dto.CreateJobRequest, user string) (*domain.UtmIncidentJob, error) {
	now := time.Now().UTC()

	j := &domain.UtmIncidentJob{
		ActionID:    req.ActionID,
		Params:      req.Params,
		Agent:       req.Agent,
		Status:      req.Status,
		OriginID:    req.OriginID,
		OriginType:  req.OriginType,
		CreatedDate: now,
		CreatedUser: user,
	}
	if err := u.repo.Save(j); err != nil {
		return nil, fmt.Errorf("jobUsecase.Create: %w", err)
	}
	return j, nil
}

func (u *jobUsecase) FindByID(id int64) (*domain.UtmIncidentJob, error) {
	return u.repo.FindByID(id)
}

func (u *jobUsecase) FindAll(f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error) {
	return u.repo.FindAll(f)
}

func (u *jobUsecase) Count(f dto.JobFilter) (int64, error) {
	return u.repo.Count(f)
}

func (u *jobUsecase) Delete(id int64) error {
	return u.repo.Delete(id)
}
