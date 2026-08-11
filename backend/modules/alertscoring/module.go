package alertscoring

import (
	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/repository"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/usecase"
	dsconnectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

type Module struct {
	scoringUC connectors.ScoringUsecase
}

func NewModule(events *eventstore.Store, agents *agentmanager.AgentManagerClient, ds dsconnectors.DatasourceUsecase) *Module {
	assets := repository.NewAssetLookup(agents, ds)
	return &Module{
		scoringUC: usecase.NewScorer(repository.NewAlertSearch(events), assets),
	}
}

func (m *Module) GetScoringUsecase() connectors.ScoringUsecase { return m.scoringUC }
