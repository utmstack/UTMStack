package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type identityRow struct {
	domain.IdentityAlias
	Count int
}

var (
	ErrIdentityNameTaken = errors.New("an identity with this name already exists")
	ErrIdentityNotFound  = errors.New("identity not found")
	ErrIdentityInUse     = errors.New("identity still has assigned users")
	ErrUserNotFound      = errors.New("user not found")
)

// escapeLike neutralizes % and _ in user input so it cannot be used to
// pattern-enumerate names in ILIKE filters.
func escapeLike(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	return s
}

type pgIdentityRepository struct{ db *gorm.DB }

func NewIdentityRepository(db *gorm.DB) *pgIdentityRepository {
	return &pgIdentityRepository{db: db}
}

func (r *pgIdentityRepository) Create(ctx context.Context, name, actor string) (*domain.IdentityAlias, error) {
	alias := domain.IdentityAlias{Name: name, CreatedBy: actor}
	if err := r.db.WithContext(ctx).Create(&alias).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrIdentityNameTaken
		}
		return nil, err
	}
	return &alias, nil
}

func (r *pgIdentityRepository) List(ctx context.Context, search string) ([]dto.IdentityListItem, error) {
	q := r.db.WithContext(ctx).Model(&domain.IdentityAlias{}).
		Select("identity_alias.id, identity_alias.name, identity_alias.created_at, COUNT(identity_alias_assignment.alias_id) AS count").
		Joins("LEFT JOIN identity_alias_assignment ON identity_alias_assignment.alias_id = identity_alias.id")
	if search != "" {
		q = q.Where("identity_alias.name ILIKE ?", "%"+escapeLike(search)+"%")
	}

	var rows []identityRow
	if err := q.Group("identity_alias.id").Order("identity_alias.name ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]dto.IdentityListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.IdentityListItem{
			ID:        row.ID.String(),
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			Count:     row.Count,
		})
	}
	return items, nil
}

func (r *pgIdentityRepository) Get(ctx context.Context, id uuid.UUID) (*domain.IdentityAlias, error) {
	var alias domain.IdentityAlias
	if err := r.db.WithContext(ctx).First(&alias, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIdentityNotFound
		}
		return nil, err
	}
	return &alias, nil
}

func (r *pgIdentityRepository) Users(ctx context.Context, id uuid.UUID) ([]dto.UserRow, error) {
	alias, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	name := alias.Name
	var rows []dto.UserRow
	err = r.db.WithContext(ctx).Model(&domain.ADUser{}).
		Joins("JOIN identity_alias_assignment ON identity_alias_assignment.user_id = ad_user.id").
		Where("identity_alias_assignment.alias_id = ?", id).
		Order("ad_user.sam_account_name ASC NULLS LAST").
		Scan(&rows).Error
	for i := range rows {
		rows[i].Identity = &name
	}
	return rows, err
}

func (r *pgIdentityRepository) Delete(ctx context.Context, id uuid.UUID) (int, error) {
	if _, err := r.Get(ctx, id); err != nil {
		return 0, err
	}

	var deleted int64
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Count under a row lock: a concurrent Assign's insert blocks on it and
		// then fails the FK after the alias is gone, so the in-use check is atomic.
		var assigned int64
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Model(&domain.IdentityAliasAssignment{}).
			Where("alias_id = ?", id).Count(&assigned).Error; err != nil {
			return err
		}
		if assigned > 0 {
			return ErrIdentityInUse
		}

		res := tx.Model(&domain.IdentityAliasAssignment{}).Where("alias_id = ?", id).Delete(&domain.IdentityAliasAssignment{})
		if res.Error != nil {
			return res.Error
		}
		res = tx.Delete(&domain.IdentityAlias{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		deleted = res.RowsAffected
		return nil
	}); err != nil {
		return 0, err
	}
	return int(deleted), nil
}

func (r *pgIdentityRepository) Assign(ctx context.Context, userID, aliasID uuid.UUID, actor string) error {
	var user domain.ADUser
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	var alias domain.IdentityAlias
	if err := r.db.WithContext(ctx).First(&alias, "id = ?", aliasID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIdentityNotFound
		}
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.IdentityAliasAssignment{}).Where("user_id = ?", userID).Delete(&domain.IdentityAliasAssignment{})
		if res.Error != nil {
			return res.Error
		}
		assignment := domain.IdentityAliasAssignment{
			AliasID:    aliasID,
			UserID:     userID,
			AssignedBy: actor,
		}
		return tx.Create(&assignment).Error
	})
}

func (r *pgIdentityRepository) Unassign(ctx context.Context, userID uuid.UUID, actor string) error {
	var user domain.ADUser
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	res := r.db.WithContext(ctx).Model(&domain.IdentityAliasAssignment{}).
		Where("user_id = ?", userID).
		Delete(&domain.IdentityAliasAssignment{})
	return res.Error
}
