package mail

import (
	app_cfg_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/internal/mail/repository"
	"github.com/utmstack/utmstack/backend/internal/mail/usecase"
)

type Module struct {
	service    connectors.MailService
	configRepo connectors.MailConfigurationRepository
}

func NewModule(store app_cfg_connectors.Store) *Module {
	repo := repository.NewMailConfigurationRepository(store)
	return &Module{
		service:    usecase.New(repo),
		configRepo: repo,
	}
}

func (m *Module) Service() connectors.MailService                       { return m.service }
func (m *Module) ConfigRepo() connectors.MailConfigurationRepository    { return m.configRepo }
