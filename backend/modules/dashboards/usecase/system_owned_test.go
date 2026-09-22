package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"

	"github.com/google/uuid"
)

type fakeDashboardRepo struct {
	row   *domain.Dashboard
	saved bool
	del   bool
}

func (f *fakeDashboardRepo) Save(context.Context, *domain.Dashboard) error {
	f.saved = true
	return nil
}
func (f *fakeDashboardRepo) FindByID(context.Context, uuid.UUID) (*domain.Dashboard, error) {
	return f.row, nil
}
func (f *fakeDashboardRepo) List(context.Context, dto.DashboardFilter) ([]domain.Dashboard, int64, error) {
	return nil, 0, nil
}
func (f *fakeDashboardRepo) Delete(context.Context, uuid.UUID) error { f.del = true; return nil }

var _ connectors.DashboardRepository = (*fakeDashboardRepo)(nil)

// Every tenant reads the dashboards the product ships and none may write them.
// The tenancy callbacks already filter the write out — which is the problem:
// the update reported success and changed nothing.
func TestUpdatingASystemDashboardIsRefused(t *testing.T) {
	repo := &fakeDashboardRepo{row: &domain.Dashboard{ID: someID, SystemOwner: true, Name: "Overview"}}
	uc := NewDashboardUsecase(repo)

	_, err := uc.Update(context.Background(), &domain.Dashboard{ID: someID, Name: "Mine"}, "someone")
	if !errors.Is(err, domain.ErrSystemOwned) {
		t.Errorf("err = %v, want ErrSystemOwned", err)
	}
	if repo.saved {
		t.Error("the write reached the repository")
	}
}

// DashboardBootstrap reseeds system dashboards by name on every boot, so a
// hard delete would just come back on the next restart. Deleting one instead
// dismisses it (soft) — the bootstrap skips a dismissed row, so it stays gone
// for good. There's no undo; the dismissed row only sticks around so it can
// still be used as a copy source when creating a new dashboard.
func TestDeletingASystemDashboardDismissesItInstead(t *testing.T) {
	repo := &fakeDashboardRepo{row: &domain.Dashboard{ID: someID, SystemOwner: true}}
	uc := NewDashboardUsecase(repo)

	if err := uc.Delete(context.Background(), someID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if repo.del {
		t.Error("the hard delete reached the repository — a system dashboard must not be hard-deleted")
	}
	if !repo.saved {
		t.Fatal("the dismiss did not reach the repository")
	}
	if repo.row.DismissedAt == nil {
		t.Error("DismissedAt was not set")
	}
}

func TestDeletingATenantDashboardStillHardDeletes(t *testing.T) {
	repo := &fakeDashboardRepo{row: &domain.Dashboard{ID: someID, SystemOwner: false}}
	uc := NewDashboardUsecase(repo)

	if err := uc.Delete(context.Background(), someID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !repo.del {
		t.Error("the delete did not reach the repository")
	}
	if repo.saved {
		t.Error("a dismiss (Save) happened for a tenant-owned dashboard")
	}
}

// A tenant's own dashboard is still editable.
func TestUpdatingAnOwnDashboardWorks(t *testing.T) {
	repo := &fakeDashboardRepo{row: &domain.Dashboard{ID: someID, SystemOwner: false, Name: "Mine"}}
	uc := NewDashboardUsecase(repo)

	if _, err := uc.Update(context.Background(), &domain.Dashboard{ID: someID, Name: "Renamed"}, "someone"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !repo.saved {
		t.Error("the write did not reach the repository")
	}
}
