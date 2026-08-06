package connectors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type ListUsersFilter struct {
	Search   string
	Page     int
	PageSize int
	IDs      []uuid.UUID
}

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, f ListUsersFilter) (users []domain.User, total int64, err error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error
	Count(ctx context.Context) (int64, error)
	ExistsByEmail(ctx context.Context, email string, excludeID uuid.UUID) (bool, error)

	FindPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)
	FindRolesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Role, error)
	FindRolesByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]domain.Role, error)
}

type ChallengeRepository interface {
	Get(ctx context.Context, purpose domain.ChallengePurpose, userID uuid.UUID) (*domain.UserChallenge, error)
	FindBySecret(ctx context.Context, purpose domain.ChallengePurpose, secretHash string) (*domain.UserChallenge, error)
	Put(ctx context.Context, c *domain.UserChallenge) error
	Delete(ctx context.Context, purpose domain.ChallengePurpose, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

type TfaFactorRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.TfaFactor, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.TfaFactor, error)
	FindByUserAndType(ctx context.Context, userID uuid.UUID, t domain.TfaFactorType) (*domain.TfaFactor, error)
	Create(ctx context.Context, f *domain.TfaFactor) error
	Confirm(ctx context.Context, id uuid.UUID, at time.Time) error
	MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	DeleteUnconfirmed(ctx context.Context, before time.Time) (int64, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t *domain.RefreshToken) error
	FindActiveByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	FindActiveByID(ctx context.Context, id uuid.UUID) (*domain.RefreshToken, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	RevokeAllForUserExcept(ctx context.Context, userID uuid.UUID, exceptID uuid.UUID) error

	// RevokeOldestBeyond keeps the newest `keep` sessions of a user and revokes
	// the rest, so the list stays short enough to be read.
	RevokeOldestBeyond(ctx context.Context, userID uuid.UUID, keep int) (int64, error)

	// DeleteSpent removes rows that can no longer authorise anything: revoked
	// before graceCutoff, or expired before expiredCutoff.
	DeleteSpent(ctx context.Context, graceCutoff, expiredCutoff time.Time, limit int) (int64, error)
}

type RBACRepository interface {
	FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindRoleByName(ctx context.Context, name string) (*domain.Role, error)
	FindRolesByNames(ctx context.Context, names []string) ([]domain.Role, error)
	ListRoles(ctx context.Context) ([]domain.Role, error)
	CreateRole(ctx context.Context, r *domain.Role, permissions []string) error
	UpdateRole(ctx context.Context, r *domain.Role, permissions []string) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error)
	PermissionsExist(ctx context.Context, names []string) (bool, error)

	ReplaceUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, k *domain.APIKey) error
	Update(ctx context.Context, k *domain.APIKey) error
	DeleteByIDAndUser(ctx context.Context, id, userID uuid.UUID) error
	FindByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*domain.APIKey, error)
	FindByNameAndUser(ctx context.Context, name string, userID uuid.UUID) (*domain.APIKey, error)
	FindByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.APIKey, int64, error)
}

type IdentityProviderRepository interface {
	ListMappings(ctx context.Context, providerID uuid.UUID) ([]domain.IdentityProviderGroupMapping, error)
	ReplaceMappings(ctx context.Context, providerID uuid.UUID, mappings []domain.IdentityProviderGroupMapping) error

	Save(ctx context.Context, c *domain.IdentityProviderConfig) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IdentityProviderConfig, error)
	FindByName(ctx context.Context, name string) (*domain.IdentityProviderConfig, error)
	List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error)
	ListActive(ctx context.Context) ([]domain.IdentityProviderConfig, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
