package indexpattern

import (
	"context"
	"os"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/connectors"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/handler"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/repository"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/usecase"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"
)

type Module struct {
	handler       *handler.IndexPatternHandler
	policyHandler *handler.PolicyHandler
	registry      *domain.IndexPatternRegistry
	bootstrap     *usecase.BootstrapService
	repo          connectors.IndexPatternRepository
	policyUC      connectors.PolicyUsecase // nil when no OpenSearch config
}

func NewModule(db *gorm.DB, osEnabled bool) *Module {
	reg := &domain.IndexPatternRegistry{}
	repo := repository.NewIndexPatternRepository(db)
	bootstrap := usecase.NewBootstrapService(repo, reg)

	m := &Module{
		registry:  reg,
		bootstrap: bootstrap,
		repo:      repo,
	}

	if osEnabled {
		ismClient := repository.NewISMClient()

		// Wire ISM repo into bootstrap so InitPolicy can run.
		bootstrap.SetISMRepo(ismClient)

		// Build the pattern usecase with ISM wired for /fields.
		patternUC := usecase.NewIndexPatternUsecaseWithISM(repo, ismClient)
		m.handler = handler.NewIndexPatternHandler(patternUC)

		// Build the policy usecase.
		policyUC := usecase.NewPolicyUsecase(ismClient, reg)
		m.policyUC = policyUC
		m.policyHandler = handler.NewPolicyHandler(policyUC)
	} else {
		// No OpenSearch config — serve pattern CRUD only.
		patternUC := usecase.NewIndexPatternUsecase(repo)
		m.handler = handler.NewIndexPatternHandler(patternUC)
	}

	return m
}

func (m *Module) GetHandler() *handler.IndexPatternHandler {
	return m.handler
}

func (m *Module) GetPolicyHandler() *handler.PolicyHandler {
	return m.policyHandler
}

func (m *Module) Registry() *domain.IndexPatternRegistry {
	return m.registry
}

func (m *Module) IsIndexRemovable(ctx context.Context, indexName string) bool {
	if m.policyUC == nil {
		return false
	}
	return m.policyUC.IsIndexRemovable(ctx, indexName)
}

func (m *Module) Start(ctx context.Context) error {
	if err := m.bootstrap.InitPatterns(ctx); err != nil {
		logger.Critical("indexpattern: InitPatterns failed — cannot start: " + err.Error())
		os.Exit(-1)
	}

	if err := m.bootstrap.InitPolicy(ctx); err != nil {
		logger.Critical("indexpattern: InitPolicy failed — cannot start: " + err.Error())
		os.Exit(-1)
	}

	logger.Info("indexpattern: module started — patterns loaded successfully")
	return nil
}
