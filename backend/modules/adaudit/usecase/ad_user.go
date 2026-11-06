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
		case "o365", "azure", "gcp", "aws":
			if in.Username == nil || strings.TrimSpace(*in.Username) == "" {
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
		if source == "o365" || source == "azure" || source == "gcp" || source == "aws" {
			if ad.Username != nil {
				un := strings.ToLower(strings.TrimSpace(*ad.Username))
				ad.Username = &un
			}
			if ad.SamAccountName == nil && ad.Username != nil {
				sa := *ad.Username
				ad.SamAccountName = &sa
			}
		}

		users = append(users, ad)
	}

	if skipped > 0 {
		_ = catcher.Error("adaudit: dropped users that named no tenant", nil, map[string]any{"dropped": skipped})
	}

	users = dedupByIdentity(users)

	if err := u.repo.Upsert(ctx, users); err != nil {
		return 0, err
	}
	return len(users), nil
}

func dedupByIdentity(users []domain.ADUser) []domain.ADUser {
	order := make([]string, 0, len(users))
	byKey := make(map[string]domain.ADUser, len(users))
	for _, u := range users {
		key := identityKey(u)
		if prev, seen := byKey[key]; seen {
			byKey[key] = mergeObservation(prev, u)
			continue
		}
		order = append(order, key)
		byKey[key] = u
	}
	mergeProvisionalLinux(&order, &byKey)
	deduped := make([]domain.ADUser, 0, len(order))
	for _, k := range order {
		if u, ok := byKey[k]; ok {
			deduped = append(deduped, u)
		}
	}
	return deduped
}

func mergeObservation(prev, later domain.ADUser) domain.ADUser {
	m := later
	if m.SamAccountName == nil {
		m.SamAccountName = prev.SamAccountName
	}
	if m.Domain == nil {
		m.Domain = prev.Domain
	}
	if m.SID == nil {
		m.SID = prev.SID
	}
	if m.MachineID == nil {
		m.MachineID = prev.MachineID
	}
	if m.UIDNumber == nil {
		m.UIDNumber = prev.UIDNumber
	}
	if m.Hostname == nil {
		m.Hostname = prev.Hostname
	}
	if m.Username == nil {
		m.Username = prev.Username
	}
	if m.AccountCreatedAt == nil {
		m.AccountCreatedAt = prev.AccountCreatedAt
	}
	if m.LastLogon == nil {
		m.LastLogon = prev.LastLogon
	}
	if m.AccountDeletedAt == nil {
		m.AccountDeletedAt = prev.AccountDeletedAt
	}
	if m.LastSeen == nil {
		m.LastSeen = prev.LastSeen
	}
	return m
}

func mergeProvisionalLinux(order *[]string, byKey *map[string]domain.ADUser) {
	for _, key := range *order {
		u := (*byKey)[key]
		if u.Source != "linux" || u.MachineID == nil || *u.MachineID == "" || u.UIDNumber == nil {
			continue
		}
		if u.Hostname == nil || u.Username == nil {
			continue
		}
		pkey := "l\x00" + u.TenantID.String() + "\x00" + *u.Hostname + "\x00" + *u.Username
		if pkey == key {
			continue
		}
		p, ok := (*byKey)[pkey]
		if !ok {
			continue
		}
		if u.AccountCreatedAt == nil || u.LastLogon == nil || u.LastSeen == nil || u.AccountDeletedAt == nil {
			(*byKey)[key] = mergeObservation(p, u)
		} else {
			(*byKey)[key] = u
		}
		delete(*byKey, pkey)
	}
}

func identityKey(u domain.ADUser) string {
	switch u.Source {
	case "windows":
		return "w\x00" + u.TenantID.String() + "\x00" + strPtr(u.SID)
	case "linux":
		if u.MachineID != nil && *u.MachineID != "" && u.UIDNumber != nil {
			return "l\x00" + u.TenantID.String() + "\x00" + *u.MachineID + "\x00" + strconv.FormatUint(uint64(*u.UIDNumber), 10)
		}
		return "l\x00" + u.TenantID.String() + "\x00" + strPtr(u.Hostname) + "\x00" + strPtr(u.Username)
	default:
		return "c\x00" + u.TenantID.String() + "\x00" + u.Source + "\x00" + strPtr(u.Username)
	}
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (u *adUserUsecase) List(ctx context.Context, f dto.ADUserFilter) (*database.List[dto.UserRow], error) {
	items, total, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return &database.List[dto.UserRow]{Items: items, Total: total}, nil
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
