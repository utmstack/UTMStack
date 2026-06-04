package opensearch

import (
	"github.com/utmstack/utmstack/backend/modules/opensearch/repository"
	"github.com/utmstack/utmstack/backend/modules/opensearch/usecase"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

type Module struct {
	gateway *usecase.Gateway
}

func NewModule(client *ospkg.Client) *Module {
	repo := repository.NewOSGatewayRepository(client)
	gw := usecase.NewGateway(repo)
	return &Module{gateway: gw}
}
