package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

type fakeRepo struct {
	connectors.ADUserRepository
	got      []domain.ADUser
	spanning bool
}

func (f *fakeRepo) Upsert(ctx context.Context, users []domain.ADUser) error {
	f.spanning = tenancy.SpansAllTenants(ctx)
	f.got = append(f.got, users...)
	return nil
}

func sid(s string) *string { return &s }

// One plugin watches every tenant's domain controllers, so a batch spans them
// and the write cannot be scoped to one — each entry says whose it is.
func TestIngestSpansTenantsAndKeepsEachEntrysOwn(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewADUserUsecase(repo)

	n, err := uc.Ingest(context.Background(), dto.IngestRequest{Users: []dto.IngestUser{
		{TenantID: customerTenant, Source: "windows", SID: "S-1-5-21-1"},
	}})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if n != 1 || len(repo.got) != 1 {
		t.Fatalf("wrote %d entries, want 1", len(repo.got))
	}
	if repo.got[0].TenantID != customerTenant {
		t.Errorf("tenant = %q, want the entry's own", repo.got[0].TenantID)
	}
	if !repo.spanning {
		t.Error("the write was scoped to one tenant")
	}
}

// An entry with no tenant would be written into none at all, and the write is
// unscoped so nothing would stop it. It is never seen again — worse than being
// dropped, since the plugin pushes it again next cycle.
func TestIngestDropsEntriesWithNoTenant(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewADUserUsecase(repo)

	n, err := uc.Ingest(context.Background(), dto.IngestRequest{Users: []dto.IngestUser{
		{TenantID: "", Source: "windows", SID: "S-1-5-21-1"},
		{TenantID: "   ", Source: "windows", SID: "S-1-5-21-2"},
		{TenantID: customerTenant, Source: "windows", SID: "S-1-5-21-3"},
	}})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if n != 1 {
		t.Errorf("wrote %d entries, want only the one that named a tenant", n)
	}
	for _, u := range repo.got {
		if u.TenantID == "" {
			t.Error("an entry with no tenant reached the database")
		}
	}
	_ = sid
}
