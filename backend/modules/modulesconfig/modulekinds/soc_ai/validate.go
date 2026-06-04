package soc_ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/soc_ai/providers"
)

type provider string

const (
	providerOpenai    provider = "openai"
	providerAnthropic provider = "anthropic"
	providerAzure     provider = "azure"
	providerGemini    provider = "gemini"
	providerOllama    provider = "ollama"
	providerMistral   provider = "mistral"
	providerDeepseek  provider = "deepseek"
	providerGroq      provider = "groq"
	providerCustom    provider = "custom"
)

type socAIConfig struct {
	AutoAnalyze       bool
	IncidentCreation  bool
	ChangeAlertStatus bool
	Provider          string
	URL               string
	Model             string
	AuthType          string
	MaxTokens         string
	CustomHeaders     map[string]string
}

func parseConfig(configs []domain.UtmModuleGroupConfiguration) socAIConfig {
	cfg := socAIConfig{AuthType: "none", CustomHeaders: map[string]string{}}
	for _, c := range configs {
		val := strings.TrimSpace(c.ConfValue)
		switch c.ConfKey {
		case "utmstack.socai.autoAnalyze":
			cfg.AutoAnalyze = val == "true"
		case "utmstack.socai.incidentCreation":
			cfg.IncidentCreation = val == "true"
		case "utmstack.socai.changeAlertStatus":
			cfg.ChangeAlertStatus = val == "true"
		case "utmstack.socai.provider":
			cfg.Provider = val
		case "utmstack.socai.url":
			cfg.URL = val
		case "utmstack.socai.model":
			cfg.Model = val
		case "utmstack.socai.authType":
			if val != "" {
				cfg.AuthType = val
			}
		case "utmstack.socai.maxTokens":
			cfg.MaxTokens = val
		case "utmstack.socai.customHeaders":
			if val != "" {
				_ = json.Unmarshal([]byte(val), &cfg.CustomHeaders)
			}
		}
	}
	return cfg
}

func (k *kind) ValidateConfiguration(ctx context.Context, _ *domain.UtmModule, configs []domain.UtmModuleGroupConfiguration) error {
	cfg := parseConfig(configs)

	var p providers.IProvider
	switch provider(cfg.Provider) {
	case providerOpenai:
		p = providers.NewOpenAIProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerAnthropic:
		p = providers.NewAnthropicProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders, cfg.MaxTokens)
	case providerAzure:
		p = providers.NewAzureProvider(cfg.URL, cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerGemini:
		p = providers.NewGeminiProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerOllama:
		p = providers.NewOllamaProvider(cfg.URL, cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerMistral:
		p = providers.NewMistralProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerDeepseek:
		p = providers.NewDeepSeekProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerGroq:
		p = providers.NewGroqProvider(cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	case providerCustom:
		p = providers.NewCustomProvider(cfg.URL, cfg.Model, cfg.AuthType, cfg.CustomHeaders)
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
	return p.Validate(ctx)
}
