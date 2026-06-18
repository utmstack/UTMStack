package socai

import (
	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/socai/usecase"
	"github.com/utmstack/utmstack/backend/modules/socai/verifier"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type Module struct {
	client *SocAIClient
	config *usecase.ConfigService
}

func NewModule(baseURL, internalKey string, cipher *secret.Cipher, pipelineDir string) *Module {
	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
		config: usecase.NewConfigService(repository.NewConfigStore(pipelineDir), cipher, verifier.New()),
	}
}

func (m *Module) Client() *SocAIClient { return m.client }

// Config returns the soc-ai configuration service.
func (m *Module) Config() *usecase.ConfigService { return m.config }
