package usecase

import (
	"context"
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

func (u *jobUsecase) Create(ctx context.Context, req dto.CreateJobRequest, user string) (*domain.UtmIncidentJob, error) {
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
	if err := u.repo.Save(ctx, j); err != nil {
		return nil, fmt.Errorf("jobUsecase.Create: %w", err)
	}
	return j, nil
}

func (u *jobUsecase) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentJob, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *jobUsecase) FindAll(ctx context.Context, f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error) {
	return u.repo.FindAll(ctx, f)
}

func (u *jobUsecase) Count(ctx context.Context, f dto.JobFilter) (int64, error) {
	return u.repo.Count(ctx, f)
}

func (u *jobUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
