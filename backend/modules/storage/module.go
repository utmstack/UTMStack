package storage

import (
	"context"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/storage/connectors"
	"github.com/utmstack/utmstack/backend/modules/storage/handler"
	"github.com/utmstack/utmstack/backend/modules/storage/repository"
	"github.com/utmstack/utmstack/backend/modules/storage/usecase"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

type Module struct {
	handler   *handler.Handler
	uc        connectors.Usecase
	localCold bool
}

func NewModule(events *eventstore.Store, configDir string, enabled, localCold bool) *Module {
	if events == nil {
		return &Module{}
	}
	uc := usecase.New(
		repository.NewStoreRepository(events),
		repository.NewConfigRepository(configDir),
	)
	return &Module{uc: uc, handler: handler.New(uc, enabled), localCold: localCold}
}

func (m *Module) Handler() *handler.Handler { return m.handler }

func (m *Module) Start(ctx context.Context) {
	if m.uc == nil {
		return
	}
	go func() {
		m.EnsureLocalCold(ctx)
	}()
}

func (m *Module) EnsureLocalCold(ctx context.Context) {
	if m.uc == nil || !m.localCold {
		return
	}
	if err := m.uc.EnsureLocalCold(ctx); err != nil {
		catcher.Info(fmt.Sprintf("storage: local cold storage not active: %v", err), nil)
	}
}
