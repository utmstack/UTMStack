package socai

import (
	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/socai/usecase"
	"github.com/utmstack/utmstack/backend/modules/socai/verifier"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type Module struct {
	client *SocAIClient
	config *usecase.ConfigService
	quota  *AIQuota
	store  *repository.ConfigStore
	cipher *secret.Cipher
}

func NewModule(
	baseURL, internalKey string, cipher *secret.Cipher, pipelineDir, updatesDir string,
	quota *AIQuota, leases usecase.Leases,
) *Module {
	instanceconfig.Init(updatesDir)

	store := repository.NewConfigStore(pipelineDir)
	config := usecase.NewConfigService(store, cipher, verifier.New()).
		WithLeases(leases)
	config.StartEnsureDefaultLoop()

	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
		config: config,
		quota:  quota,
		store:  store,
		cipher: cipher,
	}
}

// ConfigStore is where the module's configuration is kept.
func (m *Module) ConfigStore() *repository.ConfigStore { return m.store }

// Cipher is what encrypts the secrets in that configuration, and the copy of it that the SOC AI service reads.
func (m *Module) Cipher() *secret.Cipher { return m.cipher }

func (m *Module) Client() *SocAIClient { return m.client }

// Config returns the soc-ai configuration service.
func (m *Module) Config() *usecase.ConfigService { return m.config }
