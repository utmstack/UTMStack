package datainput

import (
	"github.com/utmstack/utmstack/backend/modules/datainput/handler"
	"github.com/utmstack/utmstack/backend/modules/datainput/repository"
	"github.com/utmstack/utmstack/backend/modules/datainput/usecase"
	"gorm.io/gorm"
)

type Module struct {
	statusHandler *handler.DataInputStatusHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewDataInputStatusRepository(db)
	uc := usecase.NewDataInputStatusUsecase(repo)
	h := handler.NewDataInputStatusHandler(uc)

	return &Module{
		statusHandler: h,
	}
}

func (m *Module) GetHandler() *handler.DataInputStatusHandler {
	return m.statusHandler
}
