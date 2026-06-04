// Package soc_ai implements the SOC_AI module kind (UTMStack's GPT-based
// alert analyst). The model select carries the legacy verbatim options JSON
// so the panel keeps rendering the same dropdown of GPT model names.
package soc_ai

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "SOC_AI"

const (
	confDataTypeBool   = "bool"
	confDataTypeSelect = "select"
	confDataTypeNumber = "number"
)

const providerOptions = `[` +
	`{"value": "openai", "label": "OpenAI"},` +
	`{"value": "anthropic", "label": "Anthropic"},` +
	`{"value": "azure", "label": "Azure OpenAI"},` +
	`{"value": "gemini", "label": "Google Gemini"},` +
	`{"value": "ollama", "label": "Ollama"},` +
	`{"value": "mistral", "label": "Mistral AI"},` +
	`{"value": "deepseek", "label": "DeepSeek"},` +
	`{"value": "groq", "label": "Groq"},` +
	`{"value": "custom", "label": "Custom"}` +
	`]`

const authTypeOptions = `[` +
	`{"value": "none", "label": "None"},` +
	`{"value": "custom-headers", "label": "Custom Headers"}` +
	`]`

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "utmstack.socai.autoAnalyze",
			ConfName:        "Automatic Analysis",
			ConfDescription: `If set to "true", SOC AI will automatically analyze incoming alerts.`,
			ConfDataType:    confDataTypeBool, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.incidentCreation",
			ConfName:        "Automatic Incident creation",
			ConfDescription: `If set to "true", the system will create incidents based on analysis of alerts.`,
			ConfDataType:    confDataTypeBool, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.changeAlertStatus",
			ConfName:        "Change Alert Status",
			ConfDescription: `If set to "true", SOC Ai will automatically change the status of alerts. Analysts should investigate those with the status "In Review".`,
			ConfDataType:    confDataTypeBool, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.provider",
			ConfName:        "Provider",
			ConfDescription: "Select the generative AI provider that SOC AI will use.",
			ConfDataType:    confDataTypeSelect, ConfRequired: true,
			ConfOptions: providerOptions,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.url",
			ConfName:        "Provider URL",
			ConfDescription: "API endpoint for the selected provider. Required for Azure OpenAI, Ollama, and Custom; ignored otherwise.",
			ConfDataType:    domain.ConfTypeText, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.model",
			ConfName:        "Model",
			ConfDescription: "Model name to use with the selected provider (e.g., gpt-4o, claude-3-5-sonnet-latest).",
			ConfDataType:    domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.authType",
			ConfName:        "Authentication Type",
			ConfDescription: `Choose "custom-headers" to send credentials in request headers, or "none" for unauthenticated providers (e.g., local Ollama).`,
			ConfDataType:    confDataTypeSelect, ConfRequired: false,
			ConfValue:   "none",
			ConfOptions: authTypeOptions,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.maxTokens",
			ConfName:        "Max Tokens",
			ConfDescription: "Maximum tokens to request per completion. Required by Anthropic; optional for others.",
			ConfDataType:    confDataTypeNumber, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.customHeaders",
			ConfName:        "Custom Headers (JSON)",
			ConfDescription: `JSON object of headers to send with each request, e.g., {"Authorization": "Bearer ...", "x-api-key": "..."}. Required when Authentication Type is "custom-headers".`,
			ConfDataType:    domain.ConfTypePassword, ConfRequired: false,
		},
	}
}
