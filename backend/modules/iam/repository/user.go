package repository

import (
	"context"
	"errors"
	"time"

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

func (r *pgUserRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("login = ?", login).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) FindByLoginOrEmail(ctx context.Context, loginOrEmail string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).
		Where("login = ? OR email = ?", loginOrEmail, loginOrEmail).
		First(&u).Error
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
		q = q.Where(
			"login ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	if err := q.
		Order("id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *pgUserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *pgUserRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *pgUserRepository) UpdatePassword(ctx context.Context, userID uint64, hash string, defaultPassword bool) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash":    hash,
			"default_password": defaultPassword,
		}).Error
}

func (r *pgUserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&n).Error
	return n, err
}

func (r *pgUserRepository) ExistsByLoginOrEmail(ctx context.Context, login, email string, excludeID uint64) (bool, error) {
	var n int64
	q := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("(login = ? OR email = ?)", login, email)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *pgUserRepository) FindByResetKey(ctx context.Context, key string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("reset_key = ?", key).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) SetResetKey(ctx context.Context, userID uint64, key string, issuedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"reset_key":  key,
			"reset_date": issuedAt,
		}).Error
}

// ClearResetKey is called when a password reset is completed. Finishing a reset
// also activates the account, which is what turns an invited (created with
// activated=false) user into an active one once they set their password. For an
// already-active user resetting a forgotten password, activated=true is a no-op.
func (r *pgUserRepository) ClearResetKey(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"reset_key":  "",
			"reset_date": nil,
			"activated":  true,
		}).Error
}

func (r *pgUserRepository) SetTfaConfig(ctx context.Context, userID uint64, secret, method string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"tfa_secret": secret,
			"tfa_method": method,
		}).Error
}

func (r *pgUserRepository) ClearTfaConfig(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"tfa_secret": "",
			"tfa_method": "",
		}).Error
}

func (r *pgUserRepository) FindPermissionsByUserID(ctx context.Context, userID uint64) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN authority_permissions ap ON ap.permission_id = p.id
		JOIN jhi_user_authority ua    ON ua.authority_name = ap.authority_name
		WHERE ua.user_id = ?
	`, userID).Scan(&names).Error
	return names, err
}

func (r *pgUserRepository) FindRolesByUserID(ctx context.Context, userID uint64) ([]domain.Authority, error) {
	var roles []domain.Authority
	err := r.db.WithContext(ctx).Raw(`
		SELECT a.*
		FROM jhi_authority a
		JOIN jhi_user_authority ua ON ua.authority_name = a.name
		WHERE ua.user_id = ?
		ORDER BY a.name
	`, userID).Scan(&roles).Error
	return roles, err
}

// FindRolesByUserIDs loads the roles for many users in a single query, grouped by
// user id. Used to enrich the paginated user list without an N+1 per row.
func (r *pgUserRepository) FindRolesByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]domain.Authority, error) {
	out := make(map[uint64][]domain.Authority, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	type row struct {
		UserID      uint64
		Name        string
		NameShow    string
		Description string
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("jhi_user_authority ua").
		Select("ua.user_id AS user_id, a.name AS name, a.name_show AS name_show, a.description AS description").
		Joins("JOIN jhi_authority a ON a.name = ua.authority_name").
		Where("ua.user_id IN ?", userIDs).
		Order("ua.user_id, a.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		out[rw.UserID] = append(out[rw.UserID], domain.Authority{
			Name:        rw.Name,
			NameShow:    rw.NameShow,
			Description: rw.Description,
		})
	}
	return out, nil
}
