package usecase

import (
	"context"
	"strconv"

	"strings"

	"github.com/google/uuid"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type adUserUsecase struct {
	repo connectors.ADUserRepository
}

func NewADUserUsecase(repo connectors.ADUserRepository) connectors.ADUserUsecase {
	return &adUserUsecase{repo: repo}
}

func (u *adUserUsecase) Ingest(ctx context.Context, req dto.IngestRequest) (int, error) {
	ctx = tenancy.WithAllTenants(ctx)

	users := make([]domain.ADUser, 0, len(req.Users))
	skipped := 0
	for _, in := range req.Users {
		if strings.TrimSpace(in.TenantID) == "" {
			skipped++
			continue
		}

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
				parseUID(in.UIDNumber) != nil
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

		tenantID, err := uuid.Parse(strings.TrimSpace(in.TenantID))
		if err != nil {
			continue
		}

		ad := domain.ADUser{
			TenantID:         tenantID,
			Source:           source,
			SamAccountName:   textOrNil(in.SamAccountName),
			Domain:           textOrNil(in.Domain),
			MachineID:        in.MachineID,
			UIDNumber:        parseUID(in.UIDNumber),
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

	if skipped > 0 {
		_ = catcher.Error("adaudit: dropped users that named no tenant", nil, map[string]any{"dropped": skipped})
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

func (u *adUserUsecase) Each(ctx context.Context, source string, fn func(domain.ADUser) error) error {
	return u.repo.Each(tenancy.WithAllTenants(ctx), source, fn)
}

func (u *adUserUsecase) Stats(ctx context.Context) (*dto.ADUserStats, error) {
	return u.repo.Stats(ctx)
}

func (u *adUserUsecase) ResolveLinuxIdentity(ctx context.Context, req dto.ResolveLinuxIdentityRequest) (int64, error) {
	return u.repo.ResolveLinuxIdentity(tenancy.WithAllTenants(ctx), req.TenantID, req.Hostname, req.MachineID)
}

func textOrNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func parseUID(raw *string) *uint32 {
	if raw == nil {
		return nil
	}
	n, err := strconv.ParseUint(strings.TrimSpace(*raw), 10, 32)
	if err != nil {
		return nil
	}
	uid := uint32(n)
	return &uid
}
