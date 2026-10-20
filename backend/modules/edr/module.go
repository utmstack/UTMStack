package edr

import (
	"github.com/utmstack/utmstack/backend/modules/edr/connectors"
	"github.com/utmstack/utmstack/backend/modules/edr/handler"
	"github.com/utmstack/utmstack/backend/modules/edr/repository"
	"github.com/utmstack/utmstack/backend/modules/edr/usecase"
	"gorm.io/gorm"
)

type Module struct {
	usecase *usecase.Service
	handler *handler.Handler
}

func NewModule(db *gorm.DB, gateway connectors.AgentGateway, counter connectors.DetectionCounter) *Module {
	svc := usecase.New(repository.NewRepository(db), gateway, counter)
	return &Module{
		usecase: svc,
		handler: handler.NewHandler(svc),
	}
}

func (m *Module) Usecase() *usecase.Service { return m.usecase }
func (m *Module) Handler() *handler.Handler { return m.handler }
