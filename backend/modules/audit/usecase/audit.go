package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/audit/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type service struct {
	repo connectors.Repository
}

func New(repo connectors.Repository) *service {
	return &service{repo: repo}
}

func (s *service) Log(ctx context.Context, ev connectors.Event) {
	row := &domain.AuditLog{
		Timestamp:    time.Now().UTC(),
		Action:       ev.Action,
		Status:       ev.Status,
		ErrorMessage: ev.ErrorMessage,
		ResourceType: ev.ResourceType,
		ResourceID:   ev.ResourceID,
		IP:           ev.IP,
		UserAgent:    ev.UserAgent,
		UserLogin:    ev.UserLogin,
		UserID:       ev.UserID,
		SessionID:    ev.SessionID,
		EventType:    ev.EventType,
	}
	if ev.Status == "" {
		row.Status = domain.StatusSuccess
	}
	if len(ev.Metadata) > 0 {
		if b, err := json.Marshal(ev.Metadata); err == nil {
			row.Metadata = b
		}
	}

	emitToStdout(ev, row)

	if err := s.repo.Insert(ctx, row); err != nil {
		_ = catcher.Error("audit insert failed", err, nil)
	}
}

func emitToStdout(ev connectors.Event, row *domain.AuditLog) {
	args := map[string]any{
		"context":    "audit",
		"event_type": ev.EventType,
		"outcome":    row.Status,
	}
	if row.UserLogin != "" {
		args["username"] = row.UserLogin
	}
	if row.UserID != nil {
		args["user_id"] = *row.UserID
	}
	if row.IP != "" {
		args["remote_addr"] = row.IP
	}
	if row.UserAgent != "" {
		args["user_agent"] = row.UserAgent
	}
	if row.SessionID != nil {
		args["session_id"] = *row.SessionID
	}
	if row.ResourceType != "" {
		args["resource_type"] = row.ResourceType
	}
	if row.ResourceID != "" {
		args["resource_id"] = row.ResourceID
	}

	if row.Status == domain.StatusFailure {
		var cause error
		if ev.ErrorMessage != "" {
			cause = errors.New(ev.ErrorMessage)
		}
		_ = catcher.Error(ev.Action, cause, args)
		return
	}
	catcher.Info(ev.Action, args)
}

func (s *service) List(ctx context.Context, q dto.ListQuery) (*dto.ListResponse, error) {
	res, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}
	out := common_models.MapList(res, dto.ToResponse)
	return &out, nil
}

func (s *service) Get(ctx context.Context, id uint64) (*dto.LogResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	resp := dto.ToResponse(*row)
	return &resp, nil
}
