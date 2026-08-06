package notifications

import (
	"context"

	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/handler"
	"github.com/utmstack/utmstack/backend/modules/notifications/repository"
	"github.com/utmstack/utmstack/backend/modules/notifications/usecase"
	"gorm.io/gorm"
)

type Module struct {
	purger              *usecase.Purger
	usecase             connectors.NotificationUsecase
	producer            connectors.Producer
	notificationHandler *handler.NotificationHandler
}

func NewModule(db *gorm.DB, audit audit_connectors.Logger, leases usecase.Leases, readDays, allDays int) *Module {
	repo := repository.NewNotificationRepository(db)
	uc := usecase.NewNotificationUsecase(repo, audit)

	return &Module{
		purger:              usecase.NewPurger(repo, leases, readDays, allDays),
		usecase:             uc,
		producer:            uc,
		notificationHandler: handler.NewNotificationHandler(uc),
	}
}

func (m *Module) GetNotificationHandler() *handler.NotificationHandler { return m.notificationHandler }

func (m *Module) GetNotificationUsecase() connectors.NotificationUsecase { return m.usecase }

func (m *Module) Producer() connectors.Producer { return m.producer }

func (m *Module) Start(ctx context.Context) { go m.purger.Start(ctx) }
