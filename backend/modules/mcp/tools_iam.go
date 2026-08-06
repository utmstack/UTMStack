package mcp

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	iamconnectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	iamdomain "github.com/utmstack/utmstack/backend/modules/iam/domain"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerIAM(m *Module) {
	registerIAMUsers(m)
	registerIAMRoles(m)
	registerIAMAuth(m)
	registerIAMAPIKeys(m)
	registerIAMIDPs(m)
}

// ---- iam.user.* ------------------------------------------------------------

type iamUserListInput struct {
	Search   string `json:"search,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

type iamUserIDInput struct {
	ID uuid.UUID `json:"id"`
}

type iamUserCreateInput struct {
	Email     string   `json:"email"`
	LangKey   string   `json:"lang_key,omitempty"`
	RoleNames []string `json:"role_names,omitempty"`
}

type iamUserUpdateInput struct {
	ID      uuid.UUID             `json:"id"`
	Email   string                `json:"email,omitempty"`
	LangKey string                `json:"lang_key,omitempty"`
	Status  *iamdomain.UserStatus `json:"status,omitempty"`
}

type iamUserAssignRolesInput struct {
	ID        uuid.UUID `json:"id"`
	RoleNames []string  `json:"role_names"`
}

func registerIAMUsers(m *Module) {
	uc := m.deps.IAM.GetUserUsecase()

	Add(m, &mcp.Tool{
		Name: "iam.user.list", Title: "List users",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "users.read"},
		func(ctx context.Context, _ *authz.Actor, in iamUserListInput) (any, error) {
			return uc.List(ctx, dto.ListUsersQuery{
				Page: in.Page, PageSize: clampPageSize(in.PageSize), Search: in.Search,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.user.get", Title: "Get user",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "users.read"},
		func(ctx context.Context, _ *authz.Actor, in iamUserIDInput) (any, error) {
			return uc.Get(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "iam.user.create", Title: "Create user (invitation flow)",
	}, Gate{Permission: "users.write"},
		func(ctx context.Context, actor *authz.Actor, in iamUserCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateUserRequest{
				Email: in.Email, LangKey: in.LangKey, RoleNames: in.RoleNames,
			}, iamconnectors.CreateUserOptions{Invite: true})
		})

	Add(m, &mcp.Tool{
		Name: "iam.user.update", Title: "Update user",
	}, Gate{Permission: "users.write"},
		func(ctx context.Context, actor *authz.Actor, in iamUserUpdateInput) (any, error) {
			return uc.Update(ctx, in.ID, dto.UpdateUserRequest{
				Email: in.Email, LangKey: in.LangKey, Status: in.Status,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.user.deactivate", Title: "Deactivate user",
	}, Gate{Permission: "users.delete"},
		func(ctx context.Context, _ *authz.Actor, in iamUserIDInput) (any, error) {
			if err := uc.SetStatus(ctx, in.ID, iamdomain.UserStatusInactive); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deactivated": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "iam.user.assign_roles", Title: "Assign roles to user",
	}, Gate{Permission: "users.write"},
		func(ctx context.Context, actor *authz.Actor, in iamUserAssignRolesInput) (any, error) {
			if err := uc.AssignRoles(ctx, in.ID, dto.AssignRolesRequest{RoleNames: in.RoleNames}); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "role_names": in.RoleNames}, nil
		})
}

// ---- iam.role.* ------------------------------------------------------------

type iamRoleNameInput struct {
	ID uuid.UUID `json:"id"`
}

func registerIAMRoles(m *Module) {
	uc := m.deps.IAM.GetRoleUsecase()

	Add(m, &mcp.Tool{
		Name: "iam.role.list", Title: "List roles",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "roles.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.List(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "iam.role.get", Title: "Get role with permissions",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "roles.read"},
		func(ctx context.Context, _ *authz.Actor, in iamRoleNameInput) (any, error) {
			return uc.Get(ctx, in.ID)
		})
}

// ---- iam.auth.* ------------------------------------------------------------

type iamAuthUpdateMeInput struct {
	Email   string `json:"email,omitempty"`
	LangKey string `json:"lang_key,omitempty"`
}

type iamAuthChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type iamAuthRevokeSessionInput struct {
	ID uuid.UUID `json:"id"`
}

func registerIAMAuth(m *Module) {
	uc := m.deps.IAM.GetAuthUsecase()

	Add(m, &mcp.Tool{
		Name: "iam.auth.me", Title: "Get my profile",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, _ struct{}) (any, error) {
			return uc.Me(ctx, actor.UserID)
		})

	Add(m, &mcp.Tool{
		Name: "iam.auth.update_me", Title: "Update my profile",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamAuthUpdateMeInput) (any, error) {
			return uc.UpdateMe(ctx, actor.UserID, dto.UpdateMeRequest{
				Email: in.Email, LangKey: in.LangKey,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.auth.list_sessions", Title: "List my active sessions",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, _ struct{}) (any, error) {
			return uc.ListSessions(ctx, actor.UserID, actor.SessionID)
		})

	Add(m, &mcp.Tool{
		Name: "iam.auth.revoke_session", Title: "Revoke a specific session",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamAuthRevokeSessionInput) (any, error) {
			if err := uc.RevokeSession(ctx, actor.UserID, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "revoked": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "iam.auth.revoke_other_sessions", Title: "Revoke all other sessions",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, _ struct{}) (any, error) {
			if err := uc.RevokeOtherSessions(ctx, actor.UserID, actor.SessionID); err != nil {
				return nil, err
			}
			return map[string]any{"revoked": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "iam.auth.change_password", Title: "Change my password",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamAuthChangePasswordInput) (any, error) {
			if err := uc.ChangePassword(ctx, actor.UserID, dto.ChangePasswordRequest{
				CurrentPassword: in.CurrentPassword, NewPassword: in.NewPassword,
			}); err != nil {
				return nil, err
			}
			return map[string]any{"changed": true}, nil
		})
}

// ---- iam.apikey.* ----------------------------------------------------------

type iamAPIKeyUpsertInput struct {
	ID        uuid.UUID  `json:"id,omitempty"`
	Name      string     `json:"name"`
	AllowedIP []string   `json:"allowed_ip,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type iamAPIKeyListInput struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

func registerIAMAPIKeys(m *Module) {
	uc := m.deps.IAM.GetAPIKeyUsecase()

	Add(m, &mcp.Tool{
		Name: "iam.apikey.list", Title: "List my API keys",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamAPIKeyListInput) (any, error) {
			page := in.Page
			if page <= 0 {
				page = 1
			}
			return uc.List(ctx, actor.UserID, dto.ListAPIKeysQuery{
				Page: page, PageSize: clampPageSize(in.PageSize),
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.apikey.get", Title: "Get API key",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamUserIDInput) (any, error) {
			return uc.Get(ctx, actor.UserID, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "iam.apikey.create", Title: "Create API key (Enterprise)",
	}, Gate{Enterprise: true},
		func(ctx context.Context, actor *authz.Actor, in iamAPIKeyUpsertInput) (any, error) {
			return uc.Create(ctx, actor.UserID, dto.APIKeyUpsertRequest{
				Name: in.Name, AllowedIP: in.AllowedIP, ExpiresAt: in.ExpiresAt,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.apikey.update", Title: "Update API key",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamAPIKeyUpsertInput) (any, error) {
			return uc.Update(ctx, actor.UserID, in.ID, dto.APIKeyUpsertRequest{
				Name: in.Name, AllowedIP: in.AllowedIP, ExpiresAt: in.ExpiresAt,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.apikey.delete", Title: "Delete API key",
	}, Gate{},
		func(ctx context.Context, actor *authz.Actor, in iamUserIDInput) (any, error) {
			if err := uc.Delete(ctx, actor.UserID, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "iam.apikey.generate", Title: "Generate API key secret (Enterprise)",
		Description: "Returns the cleartext secret exactly once. Re-generating invalidates the previous secret.",
	}, Gate{Enterprise: true},
		func(ctx context.Context, actor *authz.Actor, in iamUserIDInput) (any, error) {
			s, err := uc.Generate(ctx, actor.UserID, in.ID)
			if err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "secret": s}, nil
		})
}

// ---- iam.idp.* -------------------------------------------------------------

type iamIDPUpsertInput struct {
	ID           uuid.UUID       `json:"id,omitempty"`
	Name         string          `json:"name"`
	ProviderType string          `json:"provider_type"`
	Settings     json.RawMessage `json:"settings"`
	Active       bool            `json:"active"`
}

type iamIDPListInput struct {
	Name         string `json:"name,omitempty"`
	ProviderType string `json:"provider_type,omitempty"`
	Active       *bool  `json:"active,omitempty"`
	Page         int    `json:"page,omitempty"`
	Size         int    `json:"size,omitempty"`
}

func registerIAMIDPs(m *Module) {
	uc := m.deps.IAM.GetIdentityProviderUsecase()

	Add(m, &mcp.Tool{
		Name: "iam.idp.list", Title: "List identity providers",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "idp.read"},
		func(ctx context.Context, _ *authz.Actor, in iamIDPListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.IdentityProviderFilter{
				Name: in.Name, ProviderType: in.ProviderType, Active: in.Active,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "iam.idp.get", Title: "Get identity provider",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "idp.read"},
		func(ctx context.Context, _ *authz.Actor, in iamUserIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "iam.idp.create", Title: "Create identity provider (Enterprise)",
	}, Gate{Permission: "idp.write", Enterprise: true},
		func(ctx context.Context, _ *authz.Actor, in iamIDPUpsertInput) (any, error) {
			return uc.Create(ctx, dto.IdentityProviderRequest{
				ID: in.ID, Name: in.Name, ProviderType: in.ProviderType,
				Settings: in.Settings, Active: in.Active,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.idp.update", Title: "Update identity provider (Enterprise)",
	}, Gate{Permission: "idp.write", Enterprise: true},
		func(ctx context.Context, _ *authz.Actor, in iamIDPUpsertInput) (any, error) {
			return uc.Update(ctx, dto.IdentityProviderRequest{
				ID: in.ID, Name: in.Name, ProviderType: in.ProviderType,
				Settings: in.Settings, Active: in.Active,
			})
		})

	Add(m, &mcp.Tool{
		Name: "iam.idp.delete", Title: "Delete identity provider",
	}, Gate{Permission: "idp.write"},
		func(ctx context.Context, _ *authz.Actor, in iamUserIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
