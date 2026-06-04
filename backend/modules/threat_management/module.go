package threat_management

import (
	"github.com/utmstack/utmstack/backend/modules/threat_management/handler"
	"github.com/utmstack/utmstack/backend/modules/threat_management/usecase"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

type Module struct {
	adversaryHandler *handler.AdversaryHandler
}

func NewModule(osClient *ospkg.Client) *Module {
	uc := usecase.NewAdversaryUsecase(osClient)
	return &Module{
		adversaryHandler: handler.NewAdversaryHandler(uc),
	}
}
