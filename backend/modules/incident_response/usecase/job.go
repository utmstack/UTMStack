package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/incident_response/connectors"
	"github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	"github.com/utmstack/utmstack/backend/modules/incident_response/dto"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type jobUsecase struct {
	repo    connectors.JobRepository
	history connectors.IncidentHistoryWriter
}

func NewJobUsecase(repo connectors.JobRepository, history connectors.IncidentHistoryWriter) connectors.JobUsecase {
	return &jobUsecase{repo: repo, history: history}
}

func (u *jobUsecase) Create(req dto.CreateJobRequest, user string) (*domain.UtmIncidentJob, error) {
	now := time.Now().UTC()

	// TODO(module#33): Look up agent OS platform from network_scan (UtmNetworkScanService)
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

	if strings.EqualFold(j.OriginType, "INCIDENT") && j.OriginID != 0 {
		var actionID int64
		if j.ActionID != nil {
			actionID = *j.ActionID
		}
		detail := fmt.Sprintf("Job created for action %d", actionID)
		if err := u.history.WriteHistory(
			context.Background(),
			int64(j.OriginID),
			"INCIDENT_COMMAND_EXECUTED",
			"Incident command executed",
			detail,
			user,
		); err != nil {
			logger.Error(fmt.Sprintf("jobUsecase.Create: write history: %s", err.Error()))
		}
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
