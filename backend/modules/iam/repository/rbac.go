package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"gorm.io/gorm"
)

type pgRBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) connectors.RBACRepository {
	return &pgRBACRepository{db: db}
}

func (r *pgRBACRepository) FindRoleByName(ctx context.Context, name string) (*domain.Authority, error) {
	var role domain.Authority
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *pgRBACRepository) ListRoles(ctx context.Context) ([]domain.Authority, error) {
	var roles []domain.Authority
	err := r.db.WithContext(ctx).Order("name").Find(&roles).Error
	return roles, err
}

func (r *pgRBACRepository) ListRolePermissions(ctx context.Context, name string) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.*
		FROM permissions p
		JOIN authority_permissions ap ON ap.permission_id = p.id
		WHERE ap.authority_name = ?
		ORDER BY p.name
	`, name).Scan(&perms).Error
	return perms, err
}

func (r *pgRBACRepository) FindRolesByNames(ctx context.Context, names []string) ([]domain.Authority, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var roles []domain.Authority
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles).Error
	return roles, err
}

// ReplaceUserRoles sets a user's role membership in jhi_user_authority. Roles are
// referenced by name (= jhi_authority.name), so no id translation is needed.
func (r *pgRBACRepository) ReplaceUserRoles(ctx context.Context, userID uint64, roleNames []string, _ string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.UserAuthority{}).Error; err != nil {
			return err
		}
		if len(roleNames) == 0 {
			return nil
		}
		rows := make([]domain.UserAuthority, 0, len(roleNames))
		for _, name := range roleNames {
			rows = append(rows, domain.UserAuthority{UserID: userID, AuthorityName: name})
		}
		return tx.Create(&rows).Error
	})
}
