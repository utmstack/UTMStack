package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type fakeRBACRepo struct {
	roles map[string]domain.Role
	last  []uuid.UUID
}

func (f *fakeRBACRepo) FindRoleByID(context.Context, uuid.UUID) (*domain.Role, error) { return nil, nil }
func (f *fakeRBACRepo) FindRoleByName(context.Context, string) (*domain.Role, error)  { return nil, nil }
func (f *fakeRBACRepo) FindRolesByNames(_ context.Context, names []string) ([]domain.Role, error) {
	out := make([]domain.Role, 0, len(names))
	for _, n := range names {
		if r, ok := f.roles[n]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeRBACRepo) ListRoles(context.Context) ([]domain.Role, error) { return nil, nil }
func (f *fakeRBACRepo) CreateRole(context.Context, *domain.Role, []string) error { return nil }
func (f *fakeRBACRepo) UpdateRole(context.Context, *domain.Role, []string) error { return nil }
func (f *fakeRBACRepo) DeleteRole(context.Context, uuid.UUID) error             { return nil }
func (f *fakeRBACRepo) ListPermissions(context.Context) ([]domain.Permission, error) { return nil, nil }
func (f *fakeRBACRepo) ListRolePermissions(context.Context, uuid.UUID) ([]domain.Permission, error) {
	return nil, nil
}
func (f *fakeRBACRepo) PermissionsExist(context.Context, []string) (bool, error) { return true, nil }
func (f *fakeRBACRepo) ReplaceUserRoles(_ context.Context, _ uuid.UUID, roleIDs []uuid.UUID) error {
	f.last = roleIDs
	return nil
}

// ACC-F3: a caller with only "users.write" (not an admin) must not be able to
// grant the platform Administrator role to anyone — including themselves —
// through AssignRoles. An admin caller must still be able to.
func TestAssignRolesCannotGrantAdminUnlessCallerIsAdmin(t *testing.T) {
	target := &domain.User{ID: uuid.New(), Email: "attacker@utmstack.com"}
	adminRole := domain.Role{ID: uuid.New(), Name: AdminRoleName}
	rbac := &fakeRBACRepo{roles: map[string]domain.Role{AdminRoleName: adminRole}}
	u := NewUserUsecase(&fakeUserRepo{user: target}, rbac, nil, nil, nil)

	req := dto.AssignRolesRequest{RoleNames: []string{AdminRoleName}}

	err := u.AssignRoles(context.Background(), target.ID, req, false)
	if !errors.Is(err, domain.ErrCannotGrantAdmin) {
		t.Fatalf("non-admin caller granting ROLE_ADMIN: got %v, want ErrCannotGrantAdmin", err)
	}
	if rbac.last != nil {
		t.Fatalf("ReplaceUserRoles must not have been called, got %v", rbac.last)
	}

	if err := u.AssignRoles(context.Background(), target.ID, req, true); err != nil {
		t.Fatalf("admin caller granting ROLE_ADMIN: got %v, want nil", err)
	}
	if len(rbac.last) != 1 || rbac.last[0] != adminRole.ID {
		t.Fatalf("ReplaceUserRoles roleIDs = %v, want [%v]", rbac.last, adminRole.ID)
	}
}
