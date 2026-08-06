package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
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
	if q.Message != nil {
		db = db.Where("message = ?", *q.Message)
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

	orderBy := defaultOrder
	if q.Sort != "" {
		orderBy = parseSortParam(q.Sort)
	}

	var rows []domain.UtmNotification
	if err := db.Order(orderBy).
		Offset(q.Offset()).
		Limit(q.Limit()).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgNotificationRepository) FindAllGrouped(ctx context.Context, q dto.NotificationListQuery) ([]domain.NotificationGroup, int64, error) {
	base := r.db.WithContext(ctx).Model(&domain.UtmNotification{})

	if q.Source != nil {
		base = base.Where("source = ?", *q.Source)
	}
	if q.Type != nil {
		base = base.Where("type = ?", *q.Type)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	if q.Read != nil {
		base = base.Where("read = ?", *q.Read)
	}
	if q.From != nil && q.To != nil {
		base = base.Where("created_at BETWEEN ? AND ?", *q.From, *q.To)
	} else if q.From != nil {
		base = base.Where("created_at >= ?", *q.From)
	} else if q.To != nil {
		base = base.Where("created_at <= ?", *q.To)
	}

	countQ := base.Session(&gorm.Session{}).
		Select("source, type, message").
		Group("source, type, message")

	var total int64
	if err := r.db.WithContext(ctx).
		Table("(?) as g", countQ).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.NotificationGroup
	if err := base.Session(&gorm.Session{}).
		Select(`source, type, message,
			COUNT(*) AS count,
			MAX(created_at) AS last_created,
			SUM(CASE WHEN read = false THEN 1 ELSE 0 END) AS unread_count`).
		Group("source, type, message").
		Order("last_created DESC").
		Offset(q.Offset()).
		Limit(q.Limit()).
		Scan(&rows).Error; err != nil {
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
		return nil, domain.ErrNotFound
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
		return nil, domain.ErrNotFound
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
	res := r.db.WithContext(ctx).Model(&domain.UtmNotification{}).
		Where("read = ?", false).
		Update("read", true)
	return res.RowsAffected, res.Error
}

func (r *pgNotificationRepository) CountUnread(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.UtmNotification{}).
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
		return domain.ErrNotFound
	}
	return nil
}

var sortable = map[string]bool{
	"id": true, "created_at": true, "updated_at": true,
	"source": true, "type": true, "status": true, "read": true,
}

const defaultOrder = "created_at DESC"

func parseSortParam(sort string) string {
	field, dirPart, _ := strings.Cut(sort, ",")

	field = strings.TrimSpace(field)
	if !sortable[field] {
		return defaultOrder
	}

	dir := "ASC"
	if strings.EqualFold(strings.TrimSpace(dirPart), "desc") {
		dir = "DESC"
	}
	return field + " " + dir
}

func (r *pgNotificationRepository) DeleteOlderThan(
	ctx context.Context, cutoff time.Time, onlyRead bool, limit int,
) (int64, error) {
	q := r.db.WithContext(ctx).Where("created_at < ?", cutoff)
	if onlyRead {
		q = q.Where("read = ?", true)
	}
	res := q.Limit(limit).Delete(&domain.UtmNotification{})
	return res.RowsAffected, res.Error
}
