package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo   connectors.UserRepository
	rbacRepo   connectors.RBACRepository
	invitation connectors.UserInvitationMailer
}

func NewUserUsecase(userRepo connectors.UserRepository, rbacRepo connectors.RBACRepository, invitation connectors.UserInvitationMailer) connectors.UserUsecase {
	return &userUsecase{userRepo: userRepo, rbacRepo: rbacRepo, invitation: invitation}
}

func (u *userUsecase) List(ctx context.Context, q dto.ListUsersQuery) (*dto.UserListResponse, error) {
	users, total, err := u.userRepo.List(ctx, connectors.ListUsersFilter{
		Search:   q.Search,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		return nil, err
	}

	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	rolesByUser, err := u.userRepo.FindRolesByUserIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UserListItem, 0, len(users))
	for _, user := range users {
		roles := rolesByUser[user.ID]
		digests := make([]dto.RoleDigest, 0, len(roles))
		for _, r := range roles {
			digests = append(digests, dto.RoleDigest{Name: r.Name, DisplayName: r.NameShow})
		}
		data = append(data, dto.UserListItem{
			UserResponse:    dto.ToUserResponse(user),
			DefaultPassword: user.DefaultPassword,
			Roles:           digests,
		})
	}
	return &dto.UserListResponse{
		Data: data,
		PageInfo: dto.PageInfo{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

func (u *userUsecase) Get(ctx context.Context, id uint64) (*dto.UserDetailResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return u.toDetail(ctx, user)
}

func (u *userUsecase) Create(ctx context.Context, actor string, input dto.CreateUserRequest) (*dto.UserDetailResponse, error) {
	exists, err := u.userRepo.ExistsByLoginOrEmail(ctx, input.Login, input.Email, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrLoginOrEmailTaken
	}

	if err := u.assertRolesExist(ctx, input.RoleNames); err != nil {
		return nil, err
	}

	// The admin does not set a password. We store an unusable random hash so the
	// account cannot be logged into until the user activates it through the
	// invitation email and sets their own password via the reset flow.
	randomPass, err := secret.GenerateOpaque()
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(randomPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	resetKey, err := secret.GenerateOpaque()
	if err != nil {
		return nil, err
	}
	if len(resetKey) > resetKeyLen {
		resetKey = resetKey[:resetKeyLen]
	}

	now := time.Now().UTC()
	user := &domain.User{
		Login:           input.Login,
		Email:           input.Email,
		PasswordHash:    string(hash),
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		LangKey:         input.LangKey,
		Activated:       false,
		DefaultPassword: true,
		ResetKey:        resetKey,
		ResetDate:       &now,
		CreatedBy:       actor,
		CreatedDate:     now,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if len(input.RoleNames) > 0 {
		if err := u.rbacRepo.ReplaceUserRoles(ctx, user.ID, input.RoleNames, actor); err != nil {
			return nil, err
		}
	}

	// Best-effort, like the legacy async send: the account already exists and can
	// be re-invited via password-reset if the mail fails, so we don't roll back.
	if u.invitation != nil {
		if err := u.invitation.SendInvitation(ctx, user.Email, user.FirstName, resetKey); err != nil {
			_ = catcher.Error("send user invitation failed for "+user.Email, err, nil)
		}
	}

	return u.toDetail(ctx, user)
}

// CreateTenantAdmin provisions a tenant's first administrator. The tenant is
// forced onto the context here rather than trusted from the caller's, so the
// account lands in the tenant being created and not in whoever asked for it.
func (u *userUsecase) CreateTenantAdmin(ctx context.Context, tenantID, login, email string) error {
	_, err := u.Create(authz.WithTenantID(ctx, tenantID), "provisioning", dto.CreateUserRequest{
		Login:     login,
		Email:     email,
		RoleNames: []string{authz.RoleAdmin},
	})
	return err
}

func (u *userUsecase) Update(ctx context.Context, actor string, id uint64, input dto.UpdateUserRequest) (*dto.UserDetailResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Email != "" && input.Email != user.Email {
		exists, err := u.userRepo.ExistsByLoginOrEmail(ctx, user.Login, input.Email, user.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrLoginOrEmailTaken
		}
		user.Email = input.Email
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.LangKey != "" {
		user.LangKey = input.LangKey
	}
	if input.Activated != nil {
		user.Activated = *input.Activated
	}

	now := time.Now().UTC()
	user.LastModifiedBy = actor
	user.LastModifiedDate = &now

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return u.toDetail(ctx, user)
}

func (u *userUsecase) Deactivate(ctx context.Context, id uint64) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	user.Activated = false
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}
	return nil
}

func (u *userUsecase) AssignRoles(ctx context.Context, actor string, id uint64, input dto.AssignRolesRequest) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	if err := u.assertRolesExist(ctx, input.RoleNames); err != nil {
		return err
	}
	if err := u.rbacRepo.ReplaceUserRoles(ctx, id, input.RoleNames, actor); err != nil {
		return err
	}
	return nil
}

func (u *userUsecase) ResetTfa(ctx context.Context, id uint64) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	// Idempotent on purpose: clearing an already-disabled config is a no-op, so an
	// admin can always reset a stuck account without first checking its 2FA state.
	if err := u.userRepo.ClearTfaConfig(ctx, id); err != nil {
		return err
	}
	return nil
}

func (u *userUsecase) assertRolesExist(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	roles, err := u.rbacRepo.FindRolesByNames(ctx, names)
	if err != nil {
		return err
	}
	if len(roles) != len(names) {
		return domain.ErrInvalidRoleSubset
	}
	return nil
}

func (u *userUsecase) toDetail(ctx context.Context, user *domain.User) (*dto.UserDetailResponse, error) {
	roles, err := u.userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	digests := make([]dto.RoleDigest, 0, len(roles))
	for _, r := range roles {
		digests = append(digests, dto.RoleDigest{Name: r.Name, DisplayName: r.NameShow})
	}
	resp := &dto.UserDetailResponse{
		UserResponse:     dto.ToUserResponse(*user),
		DefaultPassword:  user.DefaultPassword,
		CreatedBy:        user.CreatedBy,
		CreatedDate:      &user.CreatedDate,
		LastModifiedBy:   user.LastModifiedBy,
		LastModifiedDate: user.LastModifiedDate,
		Roles:            digests,
	}
	return resp, nil
}
