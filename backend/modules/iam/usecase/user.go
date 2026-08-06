package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo      connectors.UserRepository
	rbacRepo      connectors.RBACRepository
	challengeRepo connectors.ChallengeRepository
	factorRepo    connectors.TfaFactorRepository
	mailer        connectors.ChallengeMailer
}

func NewUserUsecase(
	userRepo connectors.UserRepository,
	rbacRepo connectors.RBACRepository,
	challengeRepo connectors.ChallengeRepository,
	factorRepo connectors.TfaFactorRepository,
	mailer connectors.ChallengeMailer,
) connectors.UserUsecase {
	return &userUsecase{
		userRepo:      userRepo,
		rbacRepo:      rbacRepo,
		challengeRepo: challengeRepo,
		factorRepo:    factorRepo,
		mailer:        mailer,
	}
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

	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	ids := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	rolesByUser, err := u.userRepo.FindRolesByUserIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UserListItem, 0, len(users))
	for _, user := range users {
		tfa, err := u.hasConfirmedFactor(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		data = append(data, dto.UserListItem{
			UserResponse: dto.ToUserResponse(user, tfa),
			Roles:        roleDigests(rolesByUser[user.ID]),
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

func (u *userUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.UserDetailResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return u.toDetail(ctx, user)
}

func (u *userUsecase) Create(ctx context.Context, input dto.CreateUserRequest, opts connectors.CreateUserOptions) (*dto.UserDetailResponse, error) {
	exists, err := u.userRepo.ExistsByEmail(ctx, input.Email, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	roles, err := u.resolveRoles(ctx, input.RoleNames)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:   input.Email,
		Name:    input.Name,
		LangKey: input.LangKey,
		Status:  domain.UserStatusPending,
	}
	if opts.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		stored := string(hash)
		user.PasswordHash = &stored
		user.Status = domain.UserStatusActive
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if len(roles) > 0 {
		if err := u.rbacRepo.ReplaceUserRoles(ctx, user.ID, roles); err != nil {
			return nil, err
		}
	}

	// Only when asked. An account created with a password of its own must not
	// also be left holding a live activation token.
	if opts.Invite {
		if err := issueLinkChallenge(ctx, u.challengeRepo, u.mailer, user, domain.ChallengeActivation, activationTTL); err != nil {
			_ = catcher.Error("send activation mail failed for "+user.Email, err, nil)
		}
	}

	return u.toDetail(ctx, user)
}

func (u *userUsecase) Update(ctx context.Context, id uuid.UUID, input dto.UpdateUserRequest) (*dto.UserDetailResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Email != "" && input.Email != user.Email {
		exists, err := u.userRepo.ExistsByEmail(ctx, input.Email, user.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrEmailTaken
		}
		user.Email = input.Email
	}
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.LangKey != "" {
		user.LangKey = input.LangKey
	}
	if input.Status != nil {
		user.Status = *input.Status
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return u.toDetail(ctx, user)
}

func (u *userUsecase) SetStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	return u.userRepo.UpdateStatus(ctx, id, status)
}

func (u *userUsecase) AssignRoles(ctx context.Context, id uuid.UUID, input dto.AssignRolesRequest) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	roles, err := u.resolveRoles(ctx, input.RoleNames)
	if err != nil {
		return err
	}
	return u.rbacRepo.ReplaceUserRoles(ctx, id, roles)
}

func (u *userUsecase) ResetTfa(ctx context.Context, id uuid.UUID) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	return u.factorRepo.DeleteByUser(ctx, id)
}

func (u *userUsecase) resolveRoles(ctx context.Context, names []string) ([]uuid.UUID, error) {
	if len(names) == 0 {
		return nil, nil
	}
	roles, err := u.rbacRepo.FindRolesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(names) {
		return nil, domain.ErrInvalidRoleSubset
	}
	ids := make([]uuid.UUID, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID)
	}
	return ids, nil
}

func (u *userUsecase) hasConfirmedFactor(ctx context.Context, userID uuid.UUID) (bool, error) {
	factors, err := u.factorRepo.ListByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, f := range factors {
		if f.ConfirmedAt != nil {
			return true, nil
		}
	}
	return false, nil
}

func roleDigests(roles []domain.Role) []dto.RoleDigest {
	out := make([]dto.RoleDigest, 0, len(roles))
	for _, r := range roles {
		out = append(out, dto.RoleDigest{Name: r.Name, DisplayName: r.DisplayName})
	}
	return out
}

func (u *userUsecase) toDetail(ctx context.Context, user *domain.User) (*dto.UserDetailResponse, error) {
	roles, err := u.userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	tfa, err := u.hasConfirmedFactor(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &dto.UserDetailResponse{
		UserResponse: dto.ToUserResponse(*user, tfa),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Roles:        roleDigests(roles),
	}, nil
}
