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
		source := in.Source
		if source == "" {
			source = "windows"
		}

		switch source {
		case "windows":
			if strings.TrimSpace(in.SID) == "" {
				continue
			}
		case "linux":
			hasResolved := in.MachineID != nil && strings.TrimSpace(*in.MachineID) != "" &&
				in.UIDNumber != nil && strings.TrimSpace(*in.UIDNumber) != ""
			hasProvisional := in.Hostname != nil && strings.TrimSpace(*in.Hostname) != "" &&
				in.Username != nil && strings.TrimSpace(*in.Username) != ""
			if !hasResolved && !hasProvisional {
				continue
			}
		default:
			continue
		}

		active := true
		if in.Active != nil {
			active = *in.Active
		}

		ad := domain.ADUser{
			TenantID:         in.TenantID,
			Source:           source,
			SamAccountName:   in.SamAccountName,
			Domain:           in.Domain,
			MachineID:        in.MachineID,
			UIDNumber:        in.UIDNumber,
			Hostname:         in.Hostname,
			Username:         in.Username,
			Active:           active,
			AccountCreatedAt: in.AccountCreatedAt,
			LastLogon:        in.LastLogon,
			AccountDeletedAt: in.AccountDeletedAt,
			LastSeen:         in.LastSeen,
		}
		if source == "windows" {
			sid := strings.TrimSpace(in.SID)
			ad.SID = &sid
		}

		users = append(users, ad)
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

func (u *adUserUsecase) All(ctx context.Context, source string) ([]domain.ADUser, error) {
	return u.repo.All(ctx, source)
}

func (u *adUserUsecase) Stats(ctx context.Context, tenantID string) (*dto.ADUserStats, error) {
	return u.repo.Stats(ctx, tenantID)
}

func (u *adUserUsecase) ResolveLinuxIdentity(ctx context.Context, req dto.ResolveLinuxIdentityRequest) (int64, error) {
	return u.repo.ResolveLinuxIdentity(ctx, req.TenantID, req.Hostname, req.MachineID)
}
