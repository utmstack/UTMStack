package audit

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/handler"
	"github.com/utmstack/utmstack/backend/modules/audit/repository"
	"github.com/utmstack/utmstack/backend/modules/audit/usecase"
	"gorm.io/gorm"
)

type Module struct {
	logger  connectors.Logger
	usecase connectors.Usecase
	handler *handler.Handler
	svc     *usecase.Service
}

func NewModule(db *gorm.DB, retainDays int) *Module {
	catcher.Configure(false, true, true)
	repo := repository.NewRepository(db)
	svc := usecase.New(repo, retainDays)
	initRecorder(svc)
	return &Module{
		logger:  svc,
		usecase: svc,
		handler: handler.NewHandler(svc),
		svc:     svc,
	}
}

func (m *Module) Start(ctx context.Context)   { m.svc.Start(ctx) }
func (m *Module) Stop()                       { m.svc.Stop() }
func (m *Module) Logger() connectors.Logger   { return m.logger }
func (m *Module) Handler() *handler.Handler   { return m.handler }
func (m *Module) Usecase() connectors.Usecase { return m.usecase }
