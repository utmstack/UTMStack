package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// tenancyFakeDSRepo captures how the usecase calls the repository so we can
// assert scoping without spinning up a database. It embeds the interface to
// keep the surface minimal — only what tests exercise is overridden.
type tenancyFakeDSRepo struct {
	connectors.DatasourceRepository

	updatedGroupIDs []uint64
	updatedGroup    *uint64
	updateCalled    bool

	clearedGroup uint64
	clearCalled  bool

	enrichmentAllRead bool
	sensitiveAllScope bool
}

func (f *tenancyFakeDSRepo) UpdateGroup(_ context.Context, ids []uint64, groupID *uint64) error {
	f.updateCalled = true
	f.updatedGroupIDs = ids
	f.updatedGroup = groupID
	return nil
}

func (f *tenancyFakeDSRepo) ClearGroup(_ context.Context, groupID uint64) error {
	f.clearCalled = true
	f.clearedGroup = groupID
	return nil
}

func (f *tenancyFakeDSRepo) EnrichmentRows(ctx context.Context) ([]domain.Datasource, error) {
	f.enrichmentAllRead = tenancy.ReadsAllTenants(ctx)
	return nil, nil
}

func (f *tenancyFakeDSRepo) ListSensitive(ctx context.Context) ([]domain.Datasource, error) {
	f.sensitiveAllScope = tenancy.SpansAllTenants(ctx)
	return nil, nil
}

type tenancyFakeGroupRepo struct {
	connectors.AssetGroupRepository

	// findResult keys on the id — a nil result models "belongs to another
	// tenant" (the tenancy callback would filter it out).
	findResult map[uint64]*domain.UtmAssetGroup
}

func (f *tenancyFakeGroupRepo) FindByID(_ context.Context, id uint64) (*domain.UtmAssetGroup, error) {
	return f.findResult[id], nil
}

// UpdateGroup must reject a groupID the caller cannot see. Under the tenancy
// callback the group lookup already scopes by tenant, so a nil result means
// the group either does not exist or belongs to someone else — either way,
// stamping datasources with it is wrong.
func TestUpdateGroupRejectsCrossTenantGroup(t *testing.T) {
	repo := &tenancyFakeDSRepo{}
	groups := &tenancyFakeGroupRepo{findResult: map[uint64]*domain.UtmAssetGroup{}} // no matches
	u := &datasourceUsecase{repo: repo, groups: groups}

	groupID := uint64(99)
	err := u.UpdateGroup(context.Background(), dto.UpdateGroupRequest{IDs: []uint64{1, 2}, GroupID: &groupID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateGroup with cross-tenant group returned %v, want ErrNotFound", err)
	}
	if repo.updateCalled {
		t.Fatal("UpdateGroup wrote datasources despite an invalid group")
	}
}

// A groupID the caller owns must pass through unchanged.
func TestUpdateGroupAcceptsOwnedGroup(t *testing.T) {
	groupID := uint64(7)
	repo := &tenancyFakeDSRepo{}
	groups := &tenancyFakeGroupRepo{findResult: map[uint64]*domain.UtmAssetGroup{
		groupID: {ID: groupID, TenantID: "tenantA"},
	}}
	u := &datasourceUsecase{repo: repo, groups: groups}

	if err := u.UpdateGroup(context.Background(), dto.UpdateGroupRequest{IDs: []uint64{1}, GroupID: &groupID}); err != nil {
		t.Fatalf("UpdateGroup returned %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("UpdateGroup did not delegate to the repo")
	}
	if repo.updatedGroup == nil || *repo.updatedGroup != groupID {
		t.Fatalf("repo got groupID = %v, want %d", repo.updatedGroup, groupID)
	}
}

// Clearing the group (groupID == nil) is always allowed — no lookup needed.
func TestUpdateGroupClearingSkipsValidation(t *testing.T) {
	repo := &tenancyFakeDSRepo{}
	// Deliberately no groups repo — the usecase must not need one to clear.
	u := &datasourceUsecase{repo: repo, groups: nil}

	if err := u.UpdateGroup(context.Background(), dto.UpdateGroupRequest{IDs: []uint64{1}, GroupID: nil}); err != nil {
		t.Fatalf("UpdateGroup(nil) returned %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("UpdateGroup(nil) did not delegate to the repo")
	}
}

// Enrichment is served to the alert plugin cache, which needs every tenant's
// rows. Scoped to the caller's tenant it would return only their slice.
func TestEnrichmentReadsAllTenants(t *testing.T) {
	repo := &tenancyFakeDSRepo{}
	u := &datasourceUsecase{repo: repo}

	if _, err := u.Enrichment(context.Background()); err != nil {
		t.Fatalf("Enrichment: %v", err)
	}
	if !repo.enrichmentAllRead {
		t.Fatal("Enrichment did not opt into WithAllTenantsRead")
	}
}

// ProjectAssets rewrites a shared tenants.yaml. Scoped to the caller's tenant
// an UpdateSensitivity or Delete would overwrite the file with only that
// tenant's assets and wipe every other tenant's.
func TestProjectAssetsSpansAllTenants(t *testing.T) {
	repo := &tenancyFakeDSRepo{}
	u := &datasourceUsecase{repo: repo, projector: nopProjector{}}

	if err := u.ProjectAssets(context.Background()); err != nil {
		t.Fatalf("ProjectAssets: %v", err)
	}
	if !repo.sensitiveAllScope {
		t.Fatal("ProjectAssets read sensitive rows scoped to one tenant")
	}
}

type nopProjector struct{}

func (nopProjector) ProjectAssets([]common_models.AssetSensitivity) error { return nil }

// Deleting a group must clear datasource.group_id — postgres has no
// ON DELETE SET NULL, so a raw delete leaves dangling references.
func TestAssetGroupDeleteClearsDangling(t *testing.T) {
	dsRepo := &tenancyFakeDSRepo{}
	groupRepo := &tenancyFakeGroupRepoWithDelete{}
	u := &assetGroupUsecase{repo: groupRepo, datasources: dsRepo}

	if err := u.Delete(context.Background(), 42); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !dsRepo.clearCalled {
		t.Fatal("Delete did not clear datasource.group_id")
	}
	if dsRepo.clearedGroup != 42 {
		t.Fatalf("cleared groupID = %d, want 42", dsRepo.clearedGroup)
	}
	if !groupRepo.deleteCalled {
		t.Fatal("Delete did not remove the group row")
	}
}

type tenancyFakeGroupRepoWithDelete struct {
	connectors.AssetGroupRepository

	deleteCalled bool
}

func (f *tenancyFakeGroupRepoWithDelete) Delete(_ context.Context, _ uint64) error {
	f.deleteCalled = true
	return nil
}
