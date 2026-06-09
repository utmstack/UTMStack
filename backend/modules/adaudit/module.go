package adaudit

import (
	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/handler"
	"github.com/utmstack/utmstack/backend/modules/adaudit/repository"
	"github.com/utmstack/utmstack/backend/modules/adaudit/usecase"
	"gorm.io/gorm"
)

type Module struct {
	handler *handler.ADUserHandler
	uc      connectors.ADUserUsecase
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewADUserRepository(db)
	uc := usecase.NewADUserUsecase(repo)
	return &Module{handler: handler.NewADUserHandler(uc), uc: uc}
}

func (m *Module) Handler() *handler.ADUserHandler         { return m.handler }
func (m *Module) GetUsecase() connectors.ADUserUsecase    { return m.uc }
