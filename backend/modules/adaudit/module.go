package adaudit

import (
	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/handler"
	"github.com/utmstack/utmstack/backend/modules/adaudit/repository"
	"github.com/utmstack/utmstack/backend/modules/adaudit/usecase"
	"gorm.io/gorm"
)

type Module struct {
	handler  *handler.ADUserHandler
	identity *handler.IdentityHandler
	uc       connectors.ADUserUsecase
}

func NewModule(db *gorm.DB) *Module {
	adRepo := repository.NewADUserRepository(db)
	identityRepo := repository.NewIdentityRepository(db)
	uc := usecase.NewADUserUsecase(adRepo)
	return &Module{
		handler:  handler.NewADUserHandler(uc),
		identity: handler.NewIdentityHandler(usecase.NewIdentityUsecase(identityRepo)),
		uc:       uc,
	}
}

func (m *Module) Handler() *handler.ADUserHandler { return m.handler }
func (m *Module) Identity() *handler.IdentityHandler {
	return m.identity
}
func (m *Module) GetUsecase() connectors.ADUserUsecase { return m.uc }
