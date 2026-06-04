package datainput

import (
	"context"

	correlation_connectors "github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/datainput/handler"
	"github.com/utmstack/utmstack/backend/modules/datainput/repository"
	"github.com/utmstack/utmstack/backend/modules/datainput/usecase"
	"gorm.io/gorm"
)

type Module struct {
	statusHandler *handler.DataInputStatusHandler
	repo          interface {
		FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error)
	}
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewDataInputStatusRepository(db)
	uc := usecase.NewDataInputStatusUsecase(repo)
	h := handler.NewDataInputStatusHandler(uc)

	return &Module{
		statusHandler: h,
		repo:          repo,
	}
}

func (m *Module) GetHandler() *handler.DataInputStatusHandler {
	return m.statusHandler
}

func (m *Module) GetReader() correlation_connectors.DataInputStatusReader {
	return m
}

func (m *Module) FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error) {
	return m.repo.FindDataSourcesToConfigure(ctx, excludeType)
}
