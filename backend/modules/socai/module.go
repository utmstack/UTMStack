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
}

func NewModule(
	baseURL, internalKey string, cipher *secret.Cipher, pipelineDir, updatesDir string,
	quota *AIQuota, leases usecase.Leases,
) *Module {
	instanceconfig.Init(updatesDir)

	config := usecase.NewConfigService(repository.NewConfigStore(pipelineDir), cipher, verifier.New()).
		WithLeases(leases)
	config.StartEnsureDefaultLoop()

	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
		config: config,
		quota:  quota,
	}
}

func (m *Module) Client() *SocAIClient { return m.client }

// Config returns the soc-ai configuration service.
func (m *Module) Config() *usecase.ConfigService { return m.config }
