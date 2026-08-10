package integrations

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"

	ds_connectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/repository"
	"github.com/utmstack/utmstack/backend/modules/integrations/usecase"
	"github.com/utmstack/utmstack/backend/modules/integrations/verifier"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

type Module struct {
	db           *gorm.DB
	groups       *usecase.ConfigGroupUsecase
	integrations connectors.IntegrationUsecase
	collectors   connectors.CollectorUsecase
}

func NewModule(
	db *gorm.DB,
	cipher connectors.Cipher,
	configDir string,
	datasources ds_connectors.DatasourceUsecase,
	agentClient *agentmanager.AgentManagerClient,
) *Module {
	store := repository.NewConfigStore(configDir)
	schema := repository.NewCodeSchemaProvider() // field schema lives in code
	verif := verifier.NewBackendVerifier()

	integrationRepo := repository.NewIntegrationRepository(db)

	var collectorClient connectors.AgentManagerCollectorClient
	if agentClient != nil {
		collectorClient = agentClient
	}

	return &Module{
		db:           db,
		groups:       usecase.NewConfigGroupUsecase(store, schema, verif, cipher, integrationRepo, datasources),
		integrations: usecase.NewIntegrationUsecase(integrationRepo, store),
		collectors:   usecase.NewCollectorUsecase(collectorClient),
	}
}

func (m *Module) Start(ctx context.Context) {
	if err := m.groups.SyncDatasources(ctx); err != nil {
		_ = catcher.Error("integrations: datasource sync failed", err, nil)
	}
}

func (m *Module) Groups() *usecase.ConfigGroupUsecase { return m.groups }

func (m *Module) Integrations() connectors.IntegrationUsecase { return m.integrations }

func (m *Module) Collectors() connectors.CollectorUsecase { return m.collectors }
