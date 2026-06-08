package opensearch

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/domain"
	"github.com/utmstack/utmstack/backend/modules/opensearch/handler"
	"github.com/utmstack/utmstack/backend/modules/opensearch/repository"
	"github.com/utmstack/utmstack/backend/modules/opensearch/usecase"
	"gorm.io/gorm"
)

// Module is the OpenSearch bounded context. It owns:
//   - the query gateway (search/SQL/index/cluster) proxied to OpenSearch,
//   - the utm_index_pattern catalog (Postgres CRUD) + the system-pattern registry
//     consumed by other modules to resolve index names,
//   - the Index State Management (ISM) lifecycle policy and snapshot repository.
type Module struct {
	gateway *usecase.Gateway

	patternHandler *handler.IndexPatternHandler
	policyHandler  *handler.PolicyHandler
	registry       *domain.IndexPatternRegistry
	bootstrap      *usecase.BootstrapService
	patternRepo    connectors.IndexPatternRepository
	policyUC       connectors.PolicyUsecase // nil when no OpenSearch config

	spaceGuardOn bool
	spaceNotify  usecase.SpaceNotifyFunc
	spaceOpts    usecase.SpaceGuardOptions
}

func (m *Module) SetSpaceGuard(notify func(ctx context.Context, critical bool, message string) error, warnPct, deletePct float64, interval time.Duration) {
	m.spaceGuardOn = true
	m.spaceNotify = notify
	m.spaceOpts = usecase.SpaceGuardOptions{
		WarnPercent:   warnPct,
		DeletePercent: deletePct,
		Interval:      interval,
	}
}

func NewModule(db *gorm.DB, osEnabled bool) *Module {
	gw := usecase.NewGateway(repository.NewOSGatewayRepository())

	reg := &domain.IndexPatternRegistry{}
	patternRepo := repository.NewIndexPatternRepository(db)
	bootstrap := usecase.NewBootstrapService(patternRepo, reg)

	m := &Module{
		gateway:     gw,
		registry:    reg,
		bootstrap:   bootstrap,
		patternRepo: patternRepo,
	}

	if osEnabled {
		ismClient := repository.NewISMClient()

		// Wire ISM repo into bootstrap so InitPolicy can run.
		bootstrap.SetISMRepo(ismClient)

		// Pattern usecase with ISM wired for /fields.
		patternUC := usecase.NewIndexPatternUsecaseWithISM(patternRepo, ismClient)
		m.patternHandler = handler.NewIndexPatternHandler(patternUC)

		policyUC := usecase.NewPolicyUsecase(ismClient, reg)
		m.policyUC = policyUC
		m.policyHandler = handler.NewPolicyHandler(policyUC)
	} else {
		// No OpenSearch config — serve pattern CRUD only.
		m.patternHandler = handler.NewIndexPatternHandler(usecase.NewIndexPatternUsecase(patternRepo))
	}

	return m
}

func (m *Module) GetPatternHandler() *handler.IndexPatternHandler { return m.patternHandler }
func (m *Module) GetPolicyHandler() *handler.PolicyHandler        { return m.policyHandler }
func (m *Module) Registry() *domain.IndexPatternRegistry          { return m.registry }

func (m *Module) IsIndexRemovable(ctx context.Context, indexName string) bool {
	if m.policyUC == nil {
		return false
	}
	return m.policyUC.IsIndexRemovable(ctx, indexName)
}

// Start loads the index-pattern registry (required) and bootstraps the ISM
// policy + snapshot repository (only when OpenSearch is configured). Returns an
// error so the caller can decide whether the failure is fatal.
func (m *Module) Start(ctx context.Context) error {
	if err := m.bootstrap.InitPatterns(ctx); err != nil {
		return catcher.Error("opensearch: InitPatterns failed", err, nil)
	}
	if err := m.bootstrap.InitPolicy(ctx); err != nil {
		return catcher.Error("opensearch: InitPolicy failed", err, nil)
	}

	if m.spaceGuardOn && m.policyUC != nil {
		m.spaceOpts.LogsPattern = m.registry.Get(domain.SysPatternLogs)
		guard := usecase.NewSpaceGuard(m.gateway, m.policyUC, m.spaceNotify, m.spaceOpts)
		go guard.Run(ctx)
	}

	catcher.Info("opensearch: module started — index patterns loaded", nil)
	return nil
}
