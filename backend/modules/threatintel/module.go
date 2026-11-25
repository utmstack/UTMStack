package threatintel

import (
	"context"
	"time"

	"gorm.io/gorm"

	incidents_connectors "github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/association"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/extractor"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/mapper"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/service"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/threadwinds"
	"github.com/utmstack/utmstack/backend/modules/threatintel/handler"
	"github.com/utmstack/utmstack/backend/modules/threatintel/repository"
	"github.com/utmstack/utmstack/backend/modules/threatintel/usecase"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type Module struct {
	updatesDir     string
	configStore    *repository.ConfigStore
	feedsHandler   *handler.FeedsHandler
	analystHandler *handler.AnalystHandler
	contribution   *contribution.Job
}

// sentTTL is how long an entity that ThreatWinds accepted is remembered, so it is not sent again
// while it has not changed.
const sentTTL = 7 * 24 * time.Hour

// NewModule wires the ThreatWinds surface: the proxy reads the instance
// credentials from updatesDir, and the feed contribution is configured through
// a file in configDir that the feeds plugin watches.
func NewModule(updatesDir, configDir string, cipher *secret.Cipher, db *gorm.DB) *Module {
	instanceconfig.Init(updatesDir)

	m := &Module{updatesDir: updatesDir, configStore: repository.NewConfigStore(configDir)}
	if cipher != nil {
		m.feedsHandler = handler.NewFeedsHandler(usecase.NewFeedsService(m.configStore, cipher))
	}
	if db != nil {
		m.analystHandler = handler.NewAnalystHandler(repository.NewAnalystStore(db))
	}
	return m
}

// EnableContribution sets up the task that sends the entities found in incidents to ThreatWinds. It
// reads incidents and their alerts through the incident use cases and the alerts from the event
// store, and remembers in cache what was already sent: a cache that is nil or down makes every cycle
// send everything, as it always did.
func (m *Module) EnableContribution(
	incidents incidents_connectors.IncidentUsecase,
	links incidents_connectors.IncidentAlertUsecase,
	alerts contribution.AlertFetcher,
	leases contribution.Leases,
	cache threadwinds.KV,
) {
	source := contribution.NewIncidentSource(incidents, links)

	ingestor := threadwinds.NewSentRegistry(cache, sentTTL).Wrap(
		threadwinds.New(func() *instanceconfig.InstanceConfig { return instanceconfig.Get() }),
	)

	alertProcessor := service.NewAlertProcessor(
		contribution.NewAlertSource(alerts),
		extractor.NewFieldExtractor(),
		service.NewEntityBuilder(mapper.NewEntityMapper()),
	)

	processor := service.NewIncidentProcessor(source, ingestor, alertProcessor, association.NewAssociationBuilder())

	m.contribution = contribution.NewJob(source, processor, m.contributionEnabled, leases)
}

// contributionEnabled is the instance's own switch, read from the feeds configuration on every cycle.
func (m *Module) contributionEnabled() bool {
	cfg, err := m.configStore.Load()

	return err == nil && cfg.FeedsContributionEnabled.Enabled
}

func (m *Module) Start(ctx context.Context) {
	if m.contribution != nil {
		go m.contribution.Start(ctx)
	}
}

// FeedsHandler is nil when the instance has no encryption key, in which case
// credentials could not be stored safely and the routes are not registered.
func (m *Module) FeedsHandler() *handler.FeedsHandler { return m.feedsHandler }

func (m *Module) AnalystHandler() *handler.AnalystHandler { return m.analystHandler }
