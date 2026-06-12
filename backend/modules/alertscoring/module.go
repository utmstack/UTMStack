package alertscoring

import (
	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/repository"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/usecase"
	dsconnectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

type Module struct {
	scoringUC connectors.ScoringUsecase
}

func NewModule(search connectors.AlertSearch, agents *agentmanager.AgentManagerClient, ds dsconnectors.DatasourceUsecase) *Module {
	assets := repository.NewAssetLookup(agents, ds)
	return &Module{
		scoringUC: usecase.NewScorer(search, assets),
	}
}

func (m *Module) GetScoringUsecase() connectors.ScoringUsecase { return m.scoringUC }
