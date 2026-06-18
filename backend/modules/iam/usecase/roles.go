package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type roleUsecase struct {
	rbacRepo connectors.RBACRepository
}

func NewRoleUsecase(rbacRepo connectors.RBACRepository) connectors.RoleUsecase {
	return &roleUsecase{rbacRepo: rbacRepo}
}

func (u *roleUsecase) List(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := u.rbacRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.RoleResponse, 0, len(roles))
	for _, r := range roles {
		resp = append(resp, toRoleResponse(r))
	}
	return resp, nil
}

func (u *roleUsecase) Get(ctx context.Context, name string) (*dto.RoleDetailResponse, error) {
	role, err := u.rbacRepo.FindRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound
	}
	return u.toDetail(ctx, role)
}

func (u *roleUsecase) toDetail(ctx context.Context, role *domain.Authority) (*dto.RoleDetailResponse, error) {
	perms, err := u.rbacRepo.ListRolePermissions(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	resp := &dto.RoleDetailResponse{
		RoleResponse: toRoleResponse(*role),
		Permissions:  make([]dto.PermissionResponse, 0, len(perms)),
	}
	for _, p := range perms {
		resp.Permissions = append(resp.Permissions, toPermissionResponse(p))
	}
	return resp, nil
}

func toRoleResponse(r domain.Authority) dto.RoleResponse {
	return dto.RoleResponse{
		Name:        r.Name,
		DisplayName: r.NameShow,
		Description: r.Description,
	}
}

func toPermissionResponse(p domain.Permission) dto.PermissionResponse {
	return dto.PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
	}
}
