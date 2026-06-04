package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
	notiferrors "github.com/utmstack/utmstack/backend/modules/notifications/errors"
	"gorm.io/gorm"
)

type pgNotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) connectors.NotificationRepository {
	return &pgNotificationRepository{db: db}
}

func (r *pgNotificationRepository) Save(ctx context.Context, n *domain.UtmNotification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *pgNotificationRepository) FindByID(ctx context.Context, id int64) (*domain.UtmNotification, error) {
	var row domain.UtmNotification
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgNotificationRepository) FindAll(ctx context.Context, q dto.NotificationListQuery) ([]domain.UtmNotification, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := r.db.WithContext(ctx).Model(&domain.UtmNotification{})

	if q.Source != nil {
		db = db.Where("source = ?", *q.Source)
	}
	if q.Type != nil {
		db = db.Where("type = ?", *q.Type)
	}
	if q.Status != nil {
		db = db.Where("status = ?", *q.Status)
	}
	if q.Read != nil {
		db = db.Where("read = ?", *q.Read)
	}
	if q.From != nil && q.To != nil {
		db = db.Where("created_at BETWEEN ? AND ?", *q.From, *q.To)
	} else if q.From != nil {
		db = db.Where("created_at >= ?", *q.From)
	} else if q.To != nil {
		db = db.Where("created_at <= ?", *q.To)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := "created_at DESC"
	if q.Sort != "" {
		orderBy = parseSortParam(q.Sort)
	}

	var rows []domain.UtmNotification
	if err := db.Order(orderBy).
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgNotificationRepository) UpdateRead(ctx context.Context, id int64, read bool) (*domain.UtmNotification, error) {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notiferrors.ErrNotFound
	}
	row.Read = read
	if err := r.db.WithContext(ctx).Save(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (r *pgNotificationRepository) UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) (*domain.UtmNotification, error) {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notiferrors.ErrNotFound
	}
	// Legacy side-effect: changing status also marks the notification as read.
	row.Status = status
	row.Read = true
	if err := r.db.WithContext(ctx).Save(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (r *pgNotificationRepository) MarkAllRead(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.UtmNotification{}).
		Where("read = ?", false).
		Update("read", true)
	return res.RowsAffected, res.Error
}

func (r *pgNotificationRepository) CountUnread(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&domain.UtmNotification{}).
		Where("read = ? AND status = ?", false, domain.StatusActive).
		Count(&n).Error
	return n, err
}

func (r *pgNotificationRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Unscoped().Delete(&domain.UtmNotification{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return notiferrors.ErrNotFound
	}
	return nil
}

func parseSortParam(sort string) string {
	parts := strings.SplitN(sort, ",", 2)
	if len(parts) == 2 {
		field := strings.TrimSpace(parts[0])
		dir := strings.ToUpper(strings.TrimSpace(parts[1]))
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		return fmt.Sprintf("%s %s", field, dir)
	}
	return sort
}
