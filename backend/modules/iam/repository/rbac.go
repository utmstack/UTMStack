package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
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

func (r *pgRBACRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return r.findRole(ctx, "id = ?", id)
}

func (r *pgRBACRepository) FindRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	return r.findRole(ctx, "name = ?", name)
}

func (r *pgRBACRepository) findRole(ctx context.Context, query string, args ...any) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where(query, args...).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *pgRBACRepository) FindRolesByNames(ctx context.Context, names []string) ([]domain.Role, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var roles []domain.Role
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles).Error
	return roles, err
}

func (r *pgRBACRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Order("system_owner DESC, name").Find(&roles).Error
	return roles, err
}

func (r *pgRBACRepository) CreateRole(ctx context.Context, role *domain.Role, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		return replacePermissions(tx, role.ID, permissions)
	})
}

func (r *pgRBACRepository) UpdateRole(ctx context.Context, role *domain.Role, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.Role{}).
			Where("id = ?", role.ID).
			Updates(map[string]any{
				"name":         role.Name,
				"display_name": role.DisplayName,
				"description":  role.Description,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrRoleNotFound
		}
		return replacePermissions(tx, role.ID, permissions)
	})
}

func (r *pgRBACRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ?", id).Delete(&domain.Role{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrRoleNotFound
		}
		if err := tx.Where("role_id = ?", id).Delete(&domain.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("role_id = ?", id).Delete(&domain.UserRole{}).Error
	})
}

func replacePermissions(tx *gorm.DB, roleID uuid.UUID, permissions []string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error; err != nil {
		return err
	}
	if len(permissions) == 0 {
		return nil
	}
	rows := make([]domain.RolePermission, 0, len(permissions))
	for _, name := range permissions {
		rows = append(rows, domain.RolePermission{RoleID: roleID, PermissionName: name})
	}
	return tx.Create(&rows).Error
}

func (r *pgRBACRepository) ReplaceUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.UserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]domain.UserRole, 0, len(roleIDs))
		for _, id := range roleIDs {
			rows = append(rows, domain.UserRole{UserID: userID, RoleID: id})
		}
		return tx.Create(&rows).Error
	})
}

func (r *pgRBACRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Order("name").Find(&perms).Error
	return perms, err
}

func (r *pgRBACRepository) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Model(&domain.Permission{}).
		Joins("JOIN role_permission rp ON rp.permission_name = permissions.name").
		Where("rp.role_id = ?", roleID).
		Order("permissions.name").
		Find(&perms).Error
	return perms, err
}

func (r *pgRBACRepository) PermissionsExist(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return true, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.Permission{}).
		Where("name IN ?", names).
		Count(&n).Error
	return n == int64(len(names)), err
}
