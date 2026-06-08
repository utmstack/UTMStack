package adaudit

import (
	"github.com/utmstack/utmstack/backend/modules/adaudit/handler"
	"github.com/utmstack/utmstack/backend/modules/adaudit/repository"
	"github.com/utmstack/utmstack/backend/modules/adaudit/usecase"
	"gorm.io/gorm"
)

type Module struct {
	handler *handler.ADUserHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewADUserRepository(db)
	uc := usecase.NewADUserUsecase(repo)
	return &Module{handler: handler.NewADUserHandler(uc)}
}

func (m *Module) Handler() *handler.ADUserHandler { return m.handler }
