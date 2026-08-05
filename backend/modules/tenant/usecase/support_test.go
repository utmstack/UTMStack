package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const customerID = "8f1c1b8e-0000-4000-8000-000000000001"

type fakeRepo struct {
	rows    map[string]*domain.Tenant
	updated *domain.Tenant
}

func (f *fakeRepo) Create(context.Context, *domain.Tenant) error { return nil }
func (f *fakeRepo) Delete(context.Context, string) error         { return nil }

func (f *fakeRepo) Update(_ context.Context, t *domain.Tenant) error {
	f.updated = t
	f.rows[t.ID] = t
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id string) (*domain.Tenant, error) {
	return f.rows[id], nil
}

func (f *fakeRepo) FindByDomain(context.Context, string) (*domain.Tenant, error) { return nil, nil }

func (f *fakeRepo) List(context.Context, dto.Filter) ([]domain.Tenant, int64, error) {
	return nil, 0, nil
}

func newUsecase(rows ...*domain.Tenant) (connectors.TenantUsecase, *fakeRepo) {
	repo := &fakeRepo{rows: map[string]*domain.Tenant{}}
	for _, t := range rows {
		repo.rows[t.ID] = t
	}
	return NewTenantUsecase(repo, nil), repo
}

func customer() *domain.Tenant {
	return &domain.Tenant{
		ID:            customerID,
		Name:          "Customer",
		Domain:        "customer.example",
		Status:        domain.StatusActive,
		SupportAccess: domain.SupportNone,
	}
}

func TestSetSupportAccess(t *testing.T) {
	uc, repo := newUsecase(customer())

	got, err := uc.SetSupportAccess(context.Background(), customerID, domain.SupportRead)
	if err != nil {
		t.Fatalf("SetSupportAccess: %v", err)
	}
	if got.SupportAccess != domain.SupportRead {
		t.Errorf("support access = %q, want READ", got.SupportAccess)
	}
	if repo.updated == nil || repo.updated.SupportAccess != domain.SupportRead {
		t.Error("the grant was not persisted")
	}
}

func TestSetSupportAccessRejectsUnknownLevels(t *testing.T) {
	uc, repo := newUsecase(customer())

	for _, level := range []domain.SupportAccess{"", "read", "ALL", "TRUE"} {
		if _, err := uc.SetSupportAccess(context.Background(), customerID, level); !errors.Is(err, domain.ErrSupportInvalid) {
			t.Errorf("level %q: err = %v, want ErrSupportInvalid", level, err)
		}
	}
	if repo.updated != nil {
		t.Error("an invalid level reached the repository")
	}
}

// The default tenant is the one being granted access, so a grant to itself
// would make the mechanism decorative.
func TestSetSupportAccessRefusesTheDefaultTenant(t *testing.T) {
	uc, repo := newUsecase(&domain.Tenant{
		ID:     authz.DefaultTenantID,
		Name:   "Default",
		Status: domain.StatusActive,
	})

	_, err := uc.SetSupportAccess(context.Background(), authz.DefaultTenantID, domain.SupportFull)
	if !errors.Is(err, domain.ErrDefaultTenant) {
		t.Errorf("err = %v, want ErrDefaultTenant", err)
	}
	if repo.updated != nil {
		t.Error("the default tenant was written to")
	}
}

// The platform may edit a tenant. If that path could reach support access it
// would be granting itself the access the field exists to withhold.
func TestUpdateCannotChangeSupportAccess(t *testing.T) {
	c := customer()
	c.SupportAccess = domain.SupportNone
	uc, _ := newUsecase(c)

	got, err := uc.Update(context.Background(), customerID, dto.UpdateRequest{
		Name:   "Renamed",
		Status: domain.StatusSuspended,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.SupportAccess != domain.SupportNone {
		t.Errorf("support access = %q after an ordinary update, want NONE", got.SupportAccess)
	}
}

func TestTerminateRefusesTheDefaultTenant(t *testing.T) {
	uc, repo := newUsecase(&domain.Tenant{
		ID:     authz.DefaultTenantID,
		Status: domain.StatusActive,
	})

	if err := uc.Terminate(context.Background(), authz.DefaultTenantID); !errors.Is(err, domain.ErrDefaultTenant) {
		t.Errorf("err = %v, want ErrDefaultTenant", err)
	}
	if repo.updated != nil {
		t.Error("the default tenant was terminated")
	}
}

func intp(v int) *int { return &v }

func TestUpdateSetsLimits(t *testing.T) {
	uc, repo := newUsecase(customer())

	got, err := uc.Update(context.Background(), customerID, dto.UpdateRequest{MaxAIRequests: intp(500)})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Limits.MaxAIRequests != 500 {
		t.Errorf("limits = %+v, want 500 AI requests", got.Limits)
	}
	if repo.updated == nil || repo.updated.Limits.MaxAIRequests != 500 {
		t.Error("the limits were not persisted")
	}
}

// An update that carries no limit is the ordinary case — a rename must not
// clear what the tenant was sold.
func TestUpdateLeavesLimitsAlone(t *testing.T) {
	c := customer()
	c.Limits = domain.Limits{MaxAIRequests: 500}
	uc, _ := newUsecase(c)

	got, err := uc.Update(context.Background(), customerID, dto.UpdateRequest{Name: "Renamed"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Limits.MaxAIRequests != 500 {
		t.Errorf("max AI requests = %d, want it untouched", got.Limits.MaxAIRequests)
	}
}

// Zero is how a limit is lifted, so it has to be distinguishable from a field
// that was not sent.
func TestUpdateZeroLiftsALimit(t *testing.T) {
	c := customer()
	c.Limits = domain.Limits{MaxAIRequests: 500}
	uc, _ := newUsecase(c)

	got, err := uc.Update(context.Background(), customerID, dto.UpdateRequest{MaxAIRequests: intp(0)})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Limits.MaxAIRequests != 0 {
		t.Errorf("max AI requests = %d, want 0", got.Limits.MaxAIRequests)
	}
}

func TestUpdateRejectsNegativeLimits(t *testing.T) {
	uc, repo := newUsecase(customer())

	_, err := uc.Update(context.Background(), customerID, dto.UpdateRequest{MaxAIRequests: intp(-1)})
	if !errors.Is(err, domain.ErrLimitNegative) {
		t.Errorf("err = %v, want ErrLimitNegative", err)
	}
	if repo.updated != nil {
		t.Error("a negative limit reached the repository")
	}
}
