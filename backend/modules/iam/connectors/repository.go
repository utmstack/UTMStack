package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type ListUsersFilter struct {
	Search   string
	Page     int
	PageSize int
	IDs      []int64
}

type UserRepository interface {
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	FindByLoginOrEmail(ctx context.Context, loginOrEmail string) (*domain.User, error)
	List(ctx context.Context, f ListUsersFilter) (users []domain.User, total int64, err error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	UpdatePassword(ctx context.Context, userID uint64, hash string, defaultPassword bool) error
	Count(ctx context.Context) (int64, error)
	ExistsByLoginOrEmail(ctx context.Context, login, email string, excludeID uint64) (bool, error)

	FindByResetKey(ctx context.Context, key string) (*domain.User, error)
	SetResetKey(ctx context.Context, userID uint64, key string, issuedAt time.Time) error
	ClearResetKey(ctx context.Context, userID uint64) error

	FindPermissionsByUserID(ctx context.Context, userID uint64) ([]string, error)
	FindRolesByUserID(ctx context.Context, userID uint64) ([]domain.Authority, error)

	SetTfaConfig(ctx context.Context, userID uint64, secret, method string) error
	ClearTfaConfig(ctx context.Context, userID uint64) error
}

type TfaStateRepository interface {
	Get(ctx context.Context, purpose string, userID uint64, method string) (*domain.TfaSetupState, error)
	Put(ctx context.Context, state *domain.TfaSetupState) error
	Delete(ctx context.Context, purpose string, userID uint64, method string) error
	MarkVerified(ctx context.Context, purpose string, userID uint64, method string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t *domain.RefreshToken) error
	FindActiveByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	FindActiveByID(ctx context.Context, id uint64) (*domain.RefreshToken, error)
	ListActiveByUser(ctx context.Context, userID uint64) ([]domain.RefreshToken, error)
	Revoke(ctx context.Context, id uint64) error
	RevokeAllForUser(ctx context.Context, userID uint64) error
	RevokeAllForUserExcept(ctx context.Context, userID uint64, exceptID uint64) error
}

type RBACRepository interface {
	FindRoleByName(ctx context.Context, name string) (*domain.Authority, error)
	ListRoles(ctx context.Context) ([]domain.Authority, error)
	ListRolePermissions(ctx context.Context, name string) ([]domain.Permission, error)
	FindRolesByNames(ctx context.Context, names []string) ([]domain.Authority, error)

	ReplaceUserRoles(ctx context.Context, userID uint64, roleNames []string, createdBy string) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, k *domain.APIKey) error
	Update(ctx context.Context, k *domain.APIKey) error
	DeleteByIDAndUser(ctx context.Context, id, userID uint64) error
	FindByIDAndUser(ctx context.Context, id, userID uint64) (*domain.APIKey, error)
	FindByNameAndUser(ctx context.Context, name string, userID uint64) (*domain.APIKey, error)
	FindByKey(ctx context.Context, apiKey string) (*domain.APIKey, error)
	ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]domain.APIKey, int64, error)
}

type IdentityProviderRepository interface {
	Save(ctx context.Context, c *domain.IdentityProviderConfig) error
	FindByID(ctx context.Context, id uint64) (*domain.IdentityProviderConfig, error)
	FindByName(ctx context.Context, name string) (*domain.IdentityProviderConfig, error)
	List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error)
	ListActive(ctx context.Context) ([]domain.IdentityProviderConfig, error)
	Delete(ctx context.Context, id uint64) error
}
