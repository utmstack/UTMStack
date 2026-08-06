package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type roleUsecase struct {
	repo connectors.RBACRepository
}

func NewRoleUsecase(repo connectors.RBACRepository) connectors.RoleUsecase {
	return &roleUsecase{repo: repo}
}

func toRoleResponse(r domain.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		System:      r.SystemOwner,
	}
}

func toPermissionResponses(perms []domain.Permission) []dto.PermissionResponse {
	out := make([]dto.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		out = append(out, dto.PermissionResponse{Name: p.Name, Description: p.Description})
	}
	return out
}

func (u *roleUsecase) List(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := u.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RoleResponse, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleResponse(r))
	}
	return out, nil
}

func (u *roleUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.RoleDetailResponse, error) {
	role, err := u.repo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound
	}
	perms, err := u.repo.ListRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return &dto.RoleDetailResponse{
		RoleResponse: toRoleResponse(*role),
		Permissions:  toPermissionResponses(perms),
	}, nil
}

func (u *roleUsecase) Create(ctx context.Context, input dto.RoleUpsertRequest) (*dto.RoleDetailResponse, error) {
	if err := u.checkPermissions(ctx, input.Permissions); err != nil {
		return nil, err
	}
	existing, err := u.repo.FindRoleByName(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrRoleNameTaken
	}

	role := &domain.Role{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
	}
	if err := u.repo.CreateRole(ctx, role, input.Permissions); err != nil {
		return nil, err
	}
	return u.Get(ctx, role.ID)
}

// Update refuses a system role here rather than letting the write silently match
// no row, so the caller is told why instead of being told it worked.
func (u *roleUsecase) Update(ctx context.Context, id uuid.UUID, input dto.RoleUpsertRequest) (*dto.RoleDetailResponse, error) {
	role, err := u.repo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound
	}
	if role.SystemOwner {
		return nil, domain.ErrRoleImmutable
	}
	if err := u.checkPermissions(ctx, input.Permissions); err != nil {
		return nil, err
	}
	if input.Name != role.Name {
		clash, err := u.repo.FindRoleByName(ctx, input.Name)
		if err != nil {
			return nil, err
		}
		if clash != nil {
			return nil, domain.ErrRoleNameTaken
		}
	}

	role.Name = input.Name
	role.DisplayName = input.DisplayName
	role.Description = input.Description
	if err := u.repo.UpdateRole(ctx, role, input.Permissions); err != nil {
		return nil, err
	}
	return u.Get(ctx, role.ID)
}

func (u *roleUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := u.repo.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return domain.ErrRoleNotFound
	}
	if role.SystemOwner {
		return domain.ErrRoleImmutable
	}
	return u.repo.DeleteRole(ctx, id)
}

func (u *roleUsecase) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := u.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return toPermissionResponses(perms), nil
}

func (u *roleUsecase) checkPermissions(ctx context.Context, names []string) error {
	ok, err := u.repo.PermissionsExist(ctx, names)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrPermissionNotFound
	}
	return nil
}
