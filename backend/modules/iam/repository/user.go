package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"gorm.io/gorm"
)

type pgUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) connectors.UserRepository {
	return &pgUserRepository{db: db}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *pgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("email = ?", normalizeEmail(email)).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) List(ctx context.Context, f connectors.ListUsersFilter) ([]domain.User, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.User{})
	if len(f.IDs) > 0 {
		q = q.Where("id IN ?", f.IDs)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("email ILIKE ? OR name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	if err := q.
		Order("created_at ASC, id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *pgUserRepository) Create(ctx context.Context, u *domain.User) error {
	u.Email = normalizeEmail(u.Email)
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *pgUserRepository) Update(ctx context.Context, u *domain.User) error {
	u.Email = normalizeEmail(u.Email)
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *pgUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password_hash", hash).Error
}

func (r *pgUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

func (r *pgUserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&n).Error
	return n, err
}

func (r *pgUserRepository) ExistsByEmail(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var n int64
	q := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("email = ?", normalizeEmail(email))
	if excludeID != uuid.Nil {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *pgUserRepository) FindPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Model(&domain.UserRole{}).
		Select("DISTINCT rp.permission_name").
		Joins("JOIN role_permission rp ON rp.role_id = user_role.role_id").
		Where("user_role.user_id = ?", userID).
		Scan(&names).Error
	return names, err
}

func (r *pgUserRepository) FindRolesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Model(&domain.UserRole{}).
		Select("r.*").
		Joins("JOIN role r ON r.id = user_role.role_id").
		Where("user_role.user_id = ?", userID).
		Order("r.name").
		Scan(&roles).Error
	return roles, err
}

func (r *pgUserRepository) FindRolesByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]domain.Role, error) {
	out := make(map[uuid.UUID][]domain.Role, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}

	type row struct {
		UserID uuid.UUID
		domain.Role
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&domain.UserRole{}).
		Select("user_role.user_id AS user_id, r.*").
		Joins("JOIN role r ON r.id = user_role.role_id").
		Where("user_role.user_id IN ?", userIDs).
		Order("user_role.user_id, r.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		out[rw.UserID] = append(out[rw.UserID], rw.Role)
	}
	return out, nil
}
