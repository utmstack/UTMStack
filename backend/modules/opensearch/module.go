package opensearch

import (
	"github.com/utmstack/utmstack/backend/modules/opensearch/repository"
	"github.com/utmstack/utmstack/backend/modules/opensearch/usecase"
)

type Module struct {
	gateway *usecase.Gateway
}

func NewModule() *Module {
	repo := repository.NewOSGatewayRepository()
	gw := usecase.NewGateway(repo)
	return &Module{gateway: gw}
}
