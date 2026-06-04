package usecase

import (
	"context"
	"strconv"
	"time"

	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
	notiferrors "github.com/utmstack/utmstack/backend/modules/notifications/errors"
)

const evtNotificationMutation audit_domain.ApplicationEventType = "NOTIFICATION_MUTATION"

type notificationUsecase struct {
	repo  connectors.NotificationRepository
	audit audit_connectors.Logger
}

func NewNotificationUsecase(repo connectors.NotificationRepository, audit audit_connectors.Logger) *notificationUsecase {
	return &notificationUsecase{repo: repo, audit: audit}
}

func (u *notificationUsecase) Create(ctx context.Context, req dto.CreateNotificationRequest) (*domain.UtmNotification, error) {
	return u.create(ctx, req.Source, req.Type, req.Message)
}

func (u *notificationUsecase) Notify(ctx context.Context, source domain.NotificationSource, ntype domain.NotificationType, message string) error {
	_, err := u.create(ctx, source, ntype, message)
	return err
}

func (u *notificationUsecase) create(ctx context.Context, source domain.NotificationSource, ntype domain.NotificationType, message string) (*domain.UtmNotification, error) {
	n := &domain.UtmNotification{
		Source:    source,
		Type:      ntype,
		Message:   message,
		CreatedAt: time.Now().UTC(),
		Read:      false,
		Status:    domain.StatusActive,
	}
	if err := u.repo.Save(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (u *notificationUsecase) List(ctx context.Context, q dto.NotificationListQuery) ([]domain.UtmNotification, int64, error) {
	return u.repo.FindAll(ctx, q)
}

func (u *notificationUsecase) GetByID(ctx context.Context, id int64) (*domain.UtmNotification, error) {
	row, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notiferrors.ErrNotFound
	}
	return row, nil
}

func (u *notificationUsecase) MarkRead(ctx context.Context, id int64, read bool) (*domain.UtmNotification, error) {
	return u.repo.UpdateRead(ctx, id, read)
}

func (u *notificationUsecase) UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) (*domain.UtmNotification, error) {
	return u.repo.UpdateStatus(ctx, id, status)
}

func (u *notificationUsecase) MarkAllRead(ctx context.Context) error {
	affected, err := u.repo.MarkAllRead(ctx)
	if err != nil {
		u.audit.Log(ctx, audit_connectors.Event{
			Action:       "notification.read_all.fail",
			EventType:    evtNotificationMutation,
			Status:       audit_domain.StatusFailure,
			ErrorMessage: err.Error(),
		})
		return err
	}
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "notification.read_all.success",
		EventType: evtNotificationMutation,
		Status:    audit_domain.StatusSuccess,
		Metadata:  map[string]any{"rowsAffected": affected},
	})
	return nil
}

func (u *notificationUsecase) CountUnread(ctx context.Context) (int64, error) {
	return u.repo.CountUnread(ctx)
}

func (u *notificationUsecase) Delete(ctx context.Context, id int64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		if err == notiferrors.ErrNotFound {
			return err
		}
		u.audit.Log(ctx, audit_connectors.Event{
			Action:       "notification.delete.fail",
			EventType:    evtNotificationMutation,
			Status:       audit_domain.StatusFailure,
			ResourceID:   strconv.FormatInt(id, 10),
			ErrorMessage: err.Error(),
		})
		return err
	}
	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "notification.delete.success",
		EventType:  evtNotificationMutation,
		Status:     audit_domain.StatusSuccess,
		ResourceID: strconv.FormatInt(id, 10),
	})
	return nil
}
