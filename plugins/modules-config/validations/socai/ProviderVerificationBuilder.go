package socai

import (
	"fmt"

	"github.com/utmstack/UTMStack/plugins/modules-config/validations/socai/providers"
)

type Provider string

const (
	Openai    Provider = "openai"
	Anthropic Provider = "anthropic"
	Azure     Provider = "azure"
	Gemini    Provider = "gemini"
	Ollama    Provider = "ollama"
	Mistral   Provider = "mistral"
	Deepseek  Provider = "deepseek"
	Groq      Provider = "groq"
	Custom    Provider = "custom"
)

type SOCAIConfig struct {
	AutoAnalyze       bool
	IncidentCreation  bool
	ChangeAlertStatus bool
	Provider          string
	URL               string
	Model             string
	AuthType          string            // "custom-headers", "none"
	MaxTokens         string
	CustomHeaders     map[string]string // All headers including auth (from frontend)
}

type ProviderVerificationBuilder struct {
	config SOCAIConfig
}

func NewSocaiVerification(config SOCAIConfig) providers.IProvider {
	return &ProviderVerificationBuilder{
		config: config,
	}
}

func (p *ProviderVerificationBuilder) Validate() error {
	var provider providers.IProvider

	switch Provider(p.config.Provider) {
	case Openai:
		provider = providers.NewOpenAIProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Anthropic:
		provider = providers.NewAnthropicProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders, p.config.MaxTokens)
	case Azure:
		provider = providers.NewAzureProvider(p.config.URL, p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Gemini:
		provider = providers.NewGeminiProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Ollama:
		provider = providers.NewOllamaProvider(p.config.URL, p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Mistral:
		provider = providers.NewMistralProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Deepseek:
		provider = providers.NewDeepSeekProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Groq:
		provider = providers.NewGroqProvider(p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	case Custom:
		provider = providers.NewCustomProvider(p.config.URL, p.config.Model, p.config.AuthType, p.config.CustomHeaders)
	default:
		return fmt.Errorf("unsupported provider: %s", p.config.Provider)
	}

	return provider.Validate()
}
