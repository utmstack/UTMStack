package usecase

import (
	"context"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type adUserUsecase struct {
	repo connectors.ADUserRepository
}

func NewADUserUsecase(repo connectors.ADUserRepository) connectors.ADUserUsecase {
	return &adUserUsecase{repo: repo}
}

func (u *adUserUsecase) Ingest(ctx context.Context, req dto.IngestRequest) (int, error) {
	users := make([]domain.ADUser, 0, len(req.Users))
	for _, in := range req.Users {
		if strings.TrimSpace(in.SID) == "" {
			continue
		}
		active := true
		if in.Active != nil {
			active = *in.Active
		}
		users = append(users, domain.ADUser{
			TenantID:         in.TenantID,
			SID:              in.SID,
			SamAccountName:   in.SamAccountName,
			Domain:           in.Domain,
			Active:           active,
			AccountCreatedAt: in.AccountCreatedAt,
			LastLogon:        in.LastLogon,
			AccountDeletedAt: in.AccountDeletedAt,
			LastSeen:         in.LastSeen,
		})
	}
	if err := u.repo.Upsert(ctx, users); err != nil {
		return 0, err
	}
	return len(users), nil
}

func (u *adUserUsecase) List(ctx context.Context, f dto.ADUserFilter) (*database.List[domain.ADUser], error) {
	items, total, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return &database.List[domain.ADUser]{Items: items, Total: total}, nil
}

func (u *adUserUsecase) All(ctx context.Context) ([]domain.ADUser, error) {
	return u.repo.All(ctx)
}

func (u *adUserUsecase) Stats(ctx context.Context, tenantID string) (*dto.ADUserStats, error) {
	return u.repo.Stats(ctx, tenantID)
}
