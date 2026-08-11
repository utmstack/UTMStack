package storage

import (
	"github.com/utmstack/utmstack/backend/modules/storage/handler"
	"github.com/utmstack/utmstack/backend/modules/storage/repository"
	"github.com/utmstack/utmstack/backend/modules/storage/usecase"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

type Module struct {
	handler *handler.Handler
}

func NewModule(events *eventstore.Store, configDir string) *Module {
	if events == nil {
		return &Module{}
	}
	uc := usecase.New(
		repository.NewStoreRepository(events),
		repository.NewConfigRepository(configDir),
	)
	return &Module{handler: handler.New(uc)}
}

func (m *Module) Handler() *handler.Handler { return m.handler }
