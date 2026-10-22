package threatintel

import (
	"context"

	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/threatintel/handler"
	"github.com/utmstack/utmstack/backend/modules/threatintel/repository"
	"github.com/utmstack/utmstack/backend/modules/threatintel/usecase"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type Module struct {
	updatesDir   string
	feedsHandler *handler.FeedsHandler
	analystHandler *handler.AnalystHandler
}

// NewModule wires the ThreatWinds surface: the proxy reads the instance
// credentials from updatesDir, and the feed contribution is configured through
// a file in configDir that the feeds plugin watches.
func NewModule(updatesDir, configDir string, cipher *secret.Cipher, db *gorm.DB) *Module {
	instanceconfig.Init(updatesDir)

	m := &Module{updatesDir: updatesDir}
	if cipher != nil {
		m.feedsHandler = handler.NewFeedsHandler(
			usecase.NewFeedsService(repository.NewConfigStore(configDir), cipher),
		)
	}
	if db != nil {
		m.analystHandler = handler.NewAnalystHandler(repository.NewAnalystStore(db))
	}
	return m
}

func (m *Module) Start(ctx context.Context) {}

// FeedsHandler is nil when the instance has no encryption key, in which case
// credentials could not be stored safely and the routes are not registered.
func (m *Module) FeedsHandler() *handler.FeedsHandler { return m.feedsHandler }

func (m *Module) AnalystHandler() *handler.AnalystHandler { return m.analystHandler }
