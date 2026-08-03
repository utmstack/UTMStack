package usecase

import (
	"context"
	"net"
	"strings"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

const defaultPageSize = 25

type tenantUsecase struct{ repo connectors.TenantRepository }

func NewTenantUsecase(repo connectors.TenantRepository) connectors.TenantUsecase {
	return &tenantUsecase{repo: repo}
}

func (u *tenantUsecase) Create(ctx context.Context, req dto.CreateRequest) (*domain.Tenant, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, domain.ErrNameRequired
	}

	host, err := normalizeDomain(req.Domain)
	if err != nil {
		return nil, err
	}

	existing, err := u.repo.FindByDomain(ctx, host)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDomainTaken
	}

	t := &domain.Tenant{
		ID:     uuid.NewString(),
		Name:   name,
		Domain: host,
		Status: domain.StatusActive,
	}
	if err := u.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	// TODO(multi-tenant): create the tenant's first administrator here, in the
	// same transaction as the row above.
	//
	// A tenant with no user is worse than no tenant: provisioning reports
	// success and nobody can log in. So the two writes have to succeed or fail
	// together — see rbac.go:62 for the transaction shape already in use.
	//
	// Blocked on User.Login and User.Email being globally unique: today only
	// one tenant in the whole instance can own the login "admin". They need
	// (tenant_id, login) and (tenant_id, email) first.
	//
	// The account is created inactive with an activation key — the fields are
	// already on the model — and the invitation reuses the existing password
	// reset flow. Delivery is best effort and stays OUTSIDE the transaction:
	// a tenant must not be rolled back because SMTP hiccuped. That implies a
	// resend endpoint.

	return t, nil
}

func (u *tenantUsecase) Update(ctx context.Context, id string, req dto.UpdateRequest) (*domain.Tenant, error) {
	t, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.StatusTerminated {
		return nil, domain.ErrAlreadyTerminated
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		t.Name = name
	}

	// Changing the domain is allowed but never routine: it breaks bookmarks and
	// whatever the customer pointed at the instance.
	if req.Domain != "" {
		host, err := normalizeDomain(req.Domain)
		if err != nil {
			return nil, err
		}
		if host != t.Domain {
			taken, err := u.repo.FindByDomain(ctx, host)
			if err != nil {
				return nil, err
			}
			if taken != nil {
				return nil, domain.ErrDomainTaken
			}
			t.Domain = host
		}
	}

	if req.Status != "" {
		if !validStatus(req.Status) {
			return nil, domain.ErrStatusInvalid
		}
		t.Status = req.Status
	}

	if err := u.repo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (u *tenantUsecase) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (u *tenantUsecase) List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error) {
	if f.Size <= 0 {
		f.Size = defaultPageSize
	}
	if f.Page < 0 {
		f.Page = 0
	}
	return u.repo.List(ctx, f)
}

// Terminate is a status change, not a delete: the data outlives the
// subscription and an operator has to be able to see what was terminated.
func (u *tenantUsecase) Terminate(ctx context.Context, id string) error {
	t, err := u.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if t.Status == domain.StatusTerminated {
		return nil
	}

	t.Status = domain.StatusTerminated

	return u.repo.Update(ctx, t)
}

func (u *tenantUsecase) ResolveDomain(ctx context.Context, host string) (*domain.Tenant, error) {
	// Hosts arrive with the port attached often enough to strip it here rather
	// than in every caller.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	t, err := u.repo.FindByDomain(ctx, strings.ToLower(strings.TrimSpace(host)))
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.ErrNotFound
	}

	return t, nil
}

func validStatus(s domain.TenantStatus) bool {
	switch s {
	case domain.StatusActive, domain.StatusSuspended, domain.StatusTerminated:
		return true
	}
	return false
}

// normalizeDomain lower-cases and validates a hostname. It is deliberately
// strict: the value ends up in URLs and TLS certificates, so a typo here is
// expensive to undo once customers have bookmarked it.
func normalizeDomain(raw string) (string, error) {
	host := strings.ToLower(strings.TrimSpace(raw))
	host = strings.TrimSuffix(host, ".")

	if host == "" {
		return "", domain.ErrDomainRequired
	}
	if len(host) > 253 || strings.Contains(host, " ") {
		return "", domain.ErrDomainInvalid
	}

	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return "", domain.ErrDomainInvalid
	}

	for _, l := range labels {
		if !validLabel(l) {
			return "", domain.ErrDomainInvalid
		}
	}

	return host, nil
}

func validLabel(l string) bool {
	if l == "" || len(l) > 63 {
		return false
	}
	if strings.HasPrefix(l, "-") || strings.HasSuffix(l, "-") {
		return false
	}
	for _, r := range l {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return true
}
