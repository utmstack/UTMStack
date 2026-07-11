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
}

func NewModule(baseURL, internalKey string, cipher *secret.Cipher, pipelineDir, updatesDir string) *Module {
	instanceconfig.Init(updatesDir)

	config := usecase.NewConfigService(repository.NewConfigStore(pipelineDir), cipher, verifier.New())
	config.StartEnsureDefaultLoop()

	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
		config: config,
	}
}

func (m *Module) Client() *SocAIClient { return m.client }

// Config returns the soc-ai configuration service.
func (m *Module) Config() *usecase.ConfigService { return m.config }
